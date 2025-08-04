package bridge

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WorkspaceManager handles isolated workspace creation and cleanup
type WorkspaceManager struct {
	baseDir string
	mu      sync.Mutex
	active  map[string]*Workspace
}

// Workspace represents an isolated execution environment
type Workspace struct {
	ID          string
	Root        string
	WorkflowDir string
	ConfigDir   string
	SecretDir   string
	OutputDir   string
	created     time.Time
}

// NewWorkspaceManager creates a new workspace manager
func NewWorkspaceManager(baseDir string) (*WorkspaceManager, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create base directory: %w", err)
	}

	return &WorkspaceManager{
		baseDir: baseDir,
		active:  make(map[string]*Workspace),
	}, nil
}

// CreateWorkspace creates a new isolated workspace for execution
func (wm *WorkspaceManager) CreateWorkspace(execID string) (*Workspace, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	ws := &Workspace{
		ID:          execID,
		Root:        filepath.Join(wm.baseDir, "exec", execID),
		WorkflowDir: filepath.Join(wm.baseDir, "exec", execID, ".github", "workflows"),
		ConfigDir:   filepath.Join(wm.baseDir, "exec", execID, "configs"),
		SecretDir:   filepath.Join(wm.baseDir, "exec", execID, ".secrets"),
		OutputDir:   filepath.Join(wm.baseDir, "exec", execID, "output"),
		created:     time.Now(),
	}

	// Create directories with proper permissions
	dirs := []struct {
		path string
		perm os.FileMode
	}{
		{ws.WorkflowDir, 0755},
		{ws.ConfigDir, 0755},
		{ws.SecretDir, 0700}, // Restrictive permissions for secrets
		{ws.OutputDir, 0755},
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d.path, d.perm); err != nil {
			ws.Cleanup() // Cleanup on partial creation
			return nil, fmt.Errorf("create %s: %w", d.path, err)
		}
	}

	wm.active[execID] = ws
	
	// Auto-cleanup after timeout
	go func() {
		time.Sleep(1 * time.Hour)
		wm.mu.Lock()
		if ws, exists := wm.active[execID]; exists {
			if err := ws.SecureCleanup(); err != nil {
				logger := log.Log
				logger.Error(err, "Failed to auto-cleanup workspace", "id", execID)
			}
			delete(wm.active, execID)
		}
		wm.mu.Unlock()
	}()

	return ws, nil
}

// CleanupStale removes workspaces older than the specified duration
func (wm *WorkspaceManager) CleanupStale(maxAge time.Duration) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	
	for id, ws := range wm.active {
		if ws.created.Before(cutoff) {
			if err := ws.SecureCleanup(); err != nil {
				logger := log.Log
				logger.Error(err, "Failed to cleanup stale workspace", "id", id)
			}
			delete(wm.active, id)
		}
	}
}

// Cleanup removes the workspace directory
func (ws *Workspace) Cleanup() error {
	return os.RemoveAll(ws.Root)
}

// SecureCleanup overwrites secrets before deletion (Jesper's pattern)
func (ws *Workspace) SecureCleanup() error {
	// Overwrite secrets before deletion
	secretFiles, _ := filepath.Glob(filepath.Join(ws.SecretDir, "*"))
	for _, f := range secretFiles {
		if err := secureDelete(f); err != nil {
			// Log but don't fail the cleanup
			logger := log.Log
			logger.Info("Warning: failed to secure delete", "file", f, "error", err)
		}
	}

	// Remove the entire workspace
	return os.RemoveAll(ws.Root)
}

// secureDelete overwrites a file with random data before deletion
func secureDelete(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Skip directories
	if info.IsDir() {
		return nil
	}

	// Overwrite with random data
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.CopyN(f, rand.Reader, info.Size())
	if err != nil {
		return err
	}

	// Sync to ensure data is written
	if err := f.Sync(); err != nil {
		return err
	}

	// Now remove the file
	return os.Remove(path)
}

// WriteWorkflow writes a workflow file to the workspace
func (ws *Workspace) WriteWorkflow(name string, content []byte) error {
	workflowPath := filepath.Join(ws.WorkflowDir, name)
	return os.WriteFile(workflowPath, content, 0644)
}

// WriteConfig writes a configuration file to the workspace
func (ws *Workspace) WriteConfig(name string, content []byte) error {
	configPath := filepath.Join(ws.ConfigDir, name)
	return os.WriteFile(configPath, content, 0644)
}

// WriteSecret writes a secret file with restrictive permissions
func (ws *Workspace) WriteSecret(name string, content []byte) error {
	secretPath := filepath.Join(ws.SecretDir, name)
	return os.WriteFile(secretPath, content, 0600)
}

// GetOutputPath returns the path for output files
func (ws *Workspace) GetOutputPath(name string) string {
	return filepath.Join(ws.OutputDir, name)
}

// ListOutputs returns all files in the output directory
func (ws *Workspace) ListOutputs() ([]string, error) {
	entries, err := os.ReadDir(ws.OutputDir)
	if err != nil {
		return nil, err
	}

	var outputs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			outputs = append(outputs, entry.Name())
		}
	}
	
	return outputs, nil
}
