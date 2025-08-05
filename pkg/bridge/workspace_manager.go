package bridge

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// WorkspaceManager manages isolated workspaces for workflow executions
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
	mu          sync.Mutex
}

// NewWorkspaceManager creates a new workspace manager
func NewWorkspaceManager(baseDir string) (*WorkspaceManager, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create base dir: %w", err)
	}
	
	return &WorkspaceManager{
		baseDir: baseDir,
		active:  make(map[string]*Workspace),
	}, nil
}

// CreateWorkspace creates a new isolated workspace
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
		{ws.Root, 0755},
		{ws.WorkflowDir, 0755},
		{ws.ConfigDir, 0755},
		{ws.SecretDir, 0700}, // Restrictive for secrets
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
				log.Printf("Failed to auto-cleanup workspace %s: %v", execID, err)
			}
			delete(wm.active, execID)
		}
		wm.mu.Unlock()
	}()
	
	return ws, nil
}

// GetWorkspace retrieves an active workspace
func (wm *WorkspaceManager) GetWorkspace(execID string) (*Workspace, bool) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	ws, exists := wm.active[execID]
	return ws, exists
}

// RemoveWorkspace removes a workspace from active tracking
func (wm *WorkspaceManager) RemoveWorkspace(execID string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	delete(wm.active, execID)
}

// Cleanup removes all workspace files
func (ws *Workspace) Cleanup() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	return os.RemoveAll(ws.Root)
}

// SecureCleanup overwrites secrets before deletion (Jesper's pattern)
func (ws *Workspace) SecureCleanup() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	// First, securely delete secrets
	secretFiles, _ := filepath.Glob(filepath.Join(ws.SecretDir, "*"))
	for _, f := range secretFiles {
		if err := secureDelete(f); err != nil {
			log.Printf("Warning: failed to secure delete %s: %v", f, err)
		}
	}
	
	// Then remove everything
	return os.RemoveAll(ws.Root)
}

// WriteWorkflow writes a workflow file to the workspace
func (ws *Workspace) WriteWorkflow(name string, content []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	// Validate filename - prevent directory traversal
	if err := validateFilename(name); err != nil {
		return fmt.Errorf("invalid workflow name: %w", err)
	}
	
	path := filepath.Join(ws.WorkflowDir, name)
	return os.WriteFile(path, content, 0644)
}

// WriteSecret writes a secret file with restrictive permissions
func (ws *Workspace) WriteSecret(name string, content []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	// Validate filename - prevent directory traversal
	if err := validateFilename(name); err != nil {
		return fmt.Errorf("invalid secret name: %w", err)
	}
	
	path := filepath.Join(ws.SecretDir, name)
	return os.WriteFile(path, content, 0600)
}

// WriteConfig writes a configuration file
func (ws *Workspace) WriteConfig(name string, content []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	// Validate filename - prevent directory traversal
	if err := validateFilename(name); err != nil {
		return fmt.Errorf("invalid config name: %w", err)
	}
	
	path := filepath.Join(ws.ConfigDir, name)
	return os.WriteFile(path, content, 0644)
}

// validateFilename ensures the filename is safe from directory traversal attacks
func validateFilename(name string) error {
	if name == "" {
		return fmt.Errorf("empty filename")
	}
	
	// Check for directory traversal attempts
	if strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("filename contains invalid characters: %s", name)
	}
	
	// Check for absolute paths
	if filepath.IsAbs(name) {
		return fmt.Errorf("absolute paths not allowed: %s", name)
	}
	
	return nil
}

// GetArtifacts returns paths to all artifacts in the output directory
func (ws *Workspace) GetArtifacts() ([]string, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	var artifacts []string
	
	err := filepath.Walk(ws.OutputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(ws.OutputDir, path)
			artifacts = append(artifacts, rel)
		}
		return nil
	})
	
	return artifacts, err
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
	
	// Open file for writing
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	
	// Overwrite with random data
	_, err = io.CopyN(f, rand.Reader, info.Size())
	if err != nil {
		return err
	}
	
	// Sync to ensure write
	if err := f.Sync(); err != nil {
		return err
	}
	
	// Finally remove
	return os.Remove(path)
}

// CleanupOldWorkspaces removes workspaces older than the specified duration
func (wm *WorkspaceManager) CleanupOldWorkspaces(maxAge time.Duration) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	execDir := filepath.Join(wm.baseDir, "exec")
	entries, err := os.ReadDir(execDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	
	now := time.Now()
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		if now.Sub(info.ModTime()) > maxAge {
			wsPath := filepath.Join(execDir, entry.Name())
			// Use secure cleanup for old workspaces
			ws := &Workspace{Root: wsPath, SecretDir: filepath.Join(wsPath, ".secrets")}
			if err := ws.SecureCleanup(); err != nil {
				log.Printf("Failed to cleanup old workspace %s: %v", entry.Name(), err)
			}
		}
	}
	
	return nil
}