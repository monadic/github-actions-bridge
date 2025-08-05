package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/confighub/actions-bridge/pkg/bridge"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceIsolation(t *testing.T) {
	// Create temporary base directory
	baseDir := t.TempDir()
	
	manager, err := bridge.NewWorkspaceManager(baseDir)
	require.NoError(t, err)
	
	// Create multiple workspaces
	ws1, err := manager.CreateWorkspace("exec-1")
	require.NoError(t, err)
	
	ws2, err := manager.CreateWorkspace("exec-2")
	require.NoError(t, err)
	
	// Verify isolation
	assert.NotEqual(t, ws1.Root, ws2.Root)
	assert.DirExists(t, ws1.Root)
	assert.DirExists(t, ws2.Root)
	assert.DirExists(t, ws1.SecretDir)
	assert.DirExists(t, ws2.SecretDir)
	
	// Verify permissions on secret directories
	info1, err := os.Stat(ws1.SecretDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), info1.Mode().Perm())
	
	// Write to workspaces
	err = ws1.WriteWorkflow("test.yml", []byte("workflow1"))
	require.NoError(t, err)
	
	err = ws2.WriteWorkflow("test.yml", []byte("workflow2"))
	require.NoError(t, err)
	
	// Verify content isolation
	content1, err := os.ReadFile(filepath.Join(ws1.WorkflowDir, "test.yml"))
	require.NoError(t, err)
	assert.Equal(t, "workflow1", string(content1))
	
	content2, err := os.ReadFile(filepath.Join(ws2.WorkflowDir, "test.yml"))
	require.NoError(t, err)
	assert.Equal(t, "workflow2", string(content2))
	
	// Test secure cleanup
	err = ws1.SecureCleanup()
	require.NoError(t, err)
	assert.NoDirExists(t, ws1.Root)
	assert.DirExists(t, ws2.Root) // Other workspace unaffected
	
	// Cleanup
	err = ws2.SecureCleanup()
	require.NoError(t, err)
}

func TestSecretHandling(t *testing.T) {
	// Create workspace
	baseDir := t.TempDir()
	manager, err := bridge.NewWorkspaceManager(baseDir)
	require.NoError(t, err)
	
	ws, err := manager.CreateWorkspace(uuid.New().String())
	require.NoError(t, err)
	defer ws.SecureCleanup()
	
	// Create secret handler
	handler, err := bridge.NewSecretHandler()
	require.NoError(t, err)
	
	// Prepare secrets
	secrets := map[string]string{
		"API_KEY":      "super-secret-key-123",
		"DATABASE_URL": "postgres://user:pass@localhost/db",
	}
	
	secretsFile, err := handler.PrepareSecrets(ws, secrets)
	require.NoError(t, err)
	assert.FileExists(t, secretsFile)
	
	// Verify file permissions
	info, err := os.Stat(secretsFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	
	// Test leak detection
	logs := []string{
		"Starting deployment",
		"Using API key: super-secret-key-123",
		"Connected to postgres://user:pass@localhost/db",
		"Deployment complete",
	}
	
	sanitized := handler.SanitizeLogs(logs)
	assert.Len(t, sanitized, 4)
	assert.Contains(t, sanitized[1], "***API_KEY***")
	assert.Contains(t, sanitized[2], "***DATABASE_URL***")
	assert.NotContains(t, sanitized[1], "super-secret-key-123")
	assert.NotContains(t, sanitized[2], "postgres://user:pass@localhost/db")
}

func TestCompatibilityChecker(t *testing.T) {
	checker := bridge.NewCompatibilityChecker()
	
	// Test workflow with known limitations
	workflow := `
name: Test Workflow
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/cache@v3
        with:
          path: ~/.npm
          key: cache-key
      - uses: actions/upload-artifact@v3
        with:
          name: test-artifact
          path: dist/
      - name: Use GitHub token
        env:
          TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: echo "Using token"
`
	
	warnings := checker.CheckWorkflow([]byte(workflow))
	assert.NotEmpty(t, warnings)
	
	// Check specific warnings
	var foundCache, foundArtifact, foundToken bool
	for _, w := range warnings {
		if w.Action != "" && w.Action[:13] == "actions/cache" {
			foundCache = true
			assert.Equal(t, "info", w.Level)
		}
		if w.Action != "" && len(w.Action) > 22 && w.Action[:22] == "actions/upload-artifact" {
			foundArtifact = true
		}
		if w.Message == "GITHUB_TOKEN will be simulated locally" {
			foundToken = true
		}
	}
	
	assert.True(t, foundCache, "Should warn about cache action")
	assert.True(t, foundArtifact, "Should warn about artifact action")
	assert.True(t, foundToken, "Should warn about GITHUB_TOKEN")
}

func TestConfigInjection(t *testing.T) {
	// Create workspace
	baseDir := t.TempDir()
	manager, err := bridge.NewWorkspaceManager(baseDir)
	require.NoError(t, err)
	
	ws, err := manager.CreateWorkspace(uuid.New().String())
	require.NoError(t, err)
	defer ws.SecureCleanup()
	
	// Create injector
	injector := bridge.NewConfigInjector(ws)
	
	// Inject configs
	configs := map[string]interface{}{
		"database": map[string]interface{}{
			"host": "localhost",
			"port": 5432,
			"name": "testdb",
		},
		"api_url": "https://api.example.com",
		"debug":   true,
	}
	
	err = injector.InjectConfigs(configs)
	require.NoError(t, err)
	
	// Verify JSON file
	jsonPath := filepath.Join(ws.ConfigDir, "config.json")
	assert.FileExists(t, jsonPath)
	
	// Verify YAML file
	yamlPath := filepath.Join(ws.ConfigDir, "config.yaml")
	assert.FileExists(t, yamlPath)
	
	// Verify env file
	envPath := filepath.Join(ws.ConfigDir, ".env")
	assert.FileExists(t, envPath)
	
	envContent, err := os.ReadFile(envPath)
	require.NoError(t, err)
	assert.Contains(t, string(envContent), "CONFIG_API_URL=https://api.example.com")
	assert.Contains(t, string(envContent), "CONFIG_DEBUG=true")
	assert.Contains(t, string(envContent), "CONFIG_DATABASE_HOST=localhost")
}

func TestActRunnerExecution(t *testing.T) {
	if os.Getenv("SKIP_ACT_TESTS") == "1" {
		t.Skip("Skipping act tests (requires Docker)")
	}
	
	// Create workspace
	baseDir := t.TempDir()
	manager, err := bridge.NewWorkspaceManager(baseDir)
	require.NoError(t, err)
	
	ws, err := manager.CreateWorkspace(uuid.New().String())
	require.NoError(t, err)
	defer ws.SecureCleanup()
	
	// Write simple workflow
	workflow := `
name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Echo test
        run: echo "Test successful"
`
	
	err = ws.WriteWorkflow("test.yml", []byte(workflow))
	require.NoError(t, err)
	
	// Create runner
	runner := bridge.NewActRunner("linux/amd64", "catthehacker/ubuntu:act-latest")
	
	// Create execution context
	ctx := &bridge.ExecutionContext{
		Workspace:  ws,
		ConfigData: []byte(workflow),
		Metadata: bridge.ExecutionMetadata{
			Space:    "test",
			Unit:     "test-unit",
			Revision: 1,
			Actor:    "test-user",
		},
		DryRun: true, // Dry run for faster tests
	}
	
	// Execute
	result, err := runner.Execute(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.NotEmpty(t, result.Logs)
}

func TestWorkspaceCleanup(t *testing.T) {
	baseDir := t.TempDir()
	manager, err := bridge.NewWorkspaceManager(baseDir)
	require.NoError(t, err)
	
	// Create workspace with secret
	ws, err := manager.CreateWorkspace(uuid.New().String())
	require.NoError(t, err)
	
	secretPath := filepath.Join(ws.SecretDir, "test-secret")
	err = os.WriteFile(secretPath, []byte("secret-value"), 0600)
	require.NoError(t, err)
	
	// Secure cleanup
	err = ws.SecureCleanup()
	require.NoError(t, err)
	
	// Verify complete removal
	assert.NoDirExists(t, ws.Root)
	assert.NoFileExists(t, secretPath)
}

func TestHealthMonitor(t *testing.T) {
	baseDir := t.TempDir()
	actionsBridge, err := bridge.NewActionsBridge(baseDir)
	require.NoError(t, err)
	
	monitor := bridge.NewHealthMonitor(actionsBridge, "test-version")
	status := monitor.HealthCheck()
	
	// Basic health should be true if Docker is available
	if _, exists := status.Checks["docker"]; exists {
		assert.NotNil(t, status.Checks["docker"])
	}
	
	// Workspace check should always pass
	assert.True(t, status.Checks["workspaces"].Healthy)
	
	// Disk check should pass in test environment
	assert.True(t, status.Checks["disk"].Healthy)
	assert.Greater(t, status.Checks["disk"].FreePercent, float64(0))
}