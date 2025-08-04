// +build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/confighub/actions-bridge/pkg/bridge"
	"github.com/confighub/sdk/bridge-worker/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBridgeIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests (set RUN_INTEGRATION_TESTS=true to run)")
	}

	// Create test bridge
	baseDir := t.TempDir()
	b, err := bridge.NewActionsBridge(baseDir)
	require.NoError(t, err)

	// Test Info method
	t.Run("Info", func(t *testing.T) {
		info := b.Info(api.InfoOptions{})
		
		assert.NotEmpty(t, info.SupportedConfigTypes)
		assert.Equal(t, "github-actions", string(info.SupportedConfigTypes[0].ToolchainType))
		assert.Equal(t, "act-local", string(info.SupportedConfigTypes[0].ProviderType))
		assert.NotEmpty(t, info.Capabilities)
	})

	// Test Apply with simple workflow
	t.Run("Apply", func(t *testing.T) {
		ctx := &testBridgeContext{
			ctx: context.Background(),
		}

		workflowContent := `
name: Test Workflow
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Test step
        run: echo "Hello from test"
      - name: Create output
        run: |
          mkdir -p output
          echo "test complete" > output/result.txt
`

		payload := api.BridgeWorkerPayload{
			QueuedOperationID: uuid.New(),
			UnitID:            uuid.New(),
			SpaceID:           uuid.New(),
			UnitSlug:          "test-unit",
			Data:              []byte(workflowContent),
			RevisionNum:       1,
			ExtraParams: mustMarshal(map[string]interface{}{
				"configs": map[string]interface{}{
					"app_name": "test-app",
					"replicas": 3,
				},
				"secrets": map[string]interface{}{
					"API_KEY": "test-secret-123",
				},
			}),
		}

		err := b.Apply(ctx, payload)
		require.NoError(t, err)

		// Verify status was sent
		assert.NotNil(t, ctx.lastStatus)
		assert.Equal(t, api.ActionApply, ctx.lastStatus.Action)
		assert.Equal(t, api.ActionResultApplyCompleted, ctx.lastStatus.Result)
		assert.Equal(t, api.ActionStatusCompleted, ctx.lastStatus.Status)
	})

	// Test workspace isolation
	t.Run("WorkspaceIsolation", func(t *testing.T) {
		manager := b.workspaceManager
		
		// Create multiple workspaces
		ws1, err := manager.CreateWorkspace("test-1")
		require.NoError(t, err)
		
		ws2, err := manager.CreateWorkspace("test-2")
		require.NoError(t, err)
		
		// Verify they're isolated
		assert.NotEqual(t, ws1.Root, ws2.Root)
		assert.DirExists(t, ws1.Root)
		assert.DirExists(t, ws2.Root)
		
		// Test secure cleanup
		testSecret := "super-secret-value"
		secretFile := filepath.Join(ws1.SecretDir, "test-secret")
		err = os.WriteFile(secretFile, []byte(testSecret), 0600)
		require.NoError(t, err)
		
		// Cleanup
		err = ws1.SecureCleanup()
		require.NoError(t, err)
		
		// Verify workspace is gone
		assert.NoDirExists(t, ws1.Root)
		
		// ws2 should still exist
		assert.DirExists(t, ws2.Root)
		
		// Cleanup ws2
		ws2.SecureCleanup()
	})

	// Test compatibility checker
	t.Run("CompatibilityChecker", func(t *testing.T) {
		checker := bridge.NewCompatibilityChecker()
		
		tests := []struct {
			name     string
			workflow string
			expected []string
		}{
			{
				name: "workflow with cache",
				workflow: `
name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/cache@v3
        with:
          path: ~/.cache
          key: cache-key
`,
				expected: []string{"Caching not supported locally"},
			},
			{
				name: "self-hosted runner",
				workflow: `
name: Test
on: push
jobs:
  test:
    runs-on: self-hosted
    steps:
      - run: echo "test"
`,
				expected: []string{"self-hosted"},
			},
			{
				name: "github token",
				workflow: `
name: Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Use token
        env:
          TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: echo "Using token"
`,
				expected: []string{"GITHUB_TOKEN will be simulated"},
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				warnings := checker.CheckWorkflow([]byte(tt.workflow))
				
				found := false
				for _, w := range warnings {
					for _, exp := range tt.expected {
						if strings.Contains(w.Message, exp) {
							found = true
							break
						}
					}
				}
				
				assert.True(t, found, "Expected warning not found in %v", warnings)
			})
		}
	})

	// Test secret leak detection
	t.Run("LeakDetection", func(t *testing.T) {
		detector := b.leakDetector
		
		// Track secrets
		detector.Track("API_KEY", "super-secret-123")
		detector.Track("DB_PASS", "password123")
		
		// Test detection
		testContent := "This contains super-secret-123 in the output"
		hasLeak, leakedKeys := detector.CheckForLeaks(testContent)
		
		assert.True(t, hasLeak)
		assert.Contains(t, leakedKeys, "API_KEY")
		
		// Test masking
		masked := detector.ScanAndMask(testContent)
		assert.NotContains(t, masked, "super-secret-123")
		assert.Contains(t, masked, "***API_KEY***")
	})

	// Test health checks
	t.Run("HealthCheck", func(t *testing.T) {
		health := b.HealthCheck()
		
		assert.NotZero(t, health.Timestamp)
		assert.NotEmpty(t, health.Version)
		assert.NotEmpty(t, health.Checks)
		
		// Should have key health checks
		assert.Contains(t, health.Checks, "docker")
		assert.Contains(t, health.Checks, "workspaces")
		assert.Contains(t, health.Checks, "disk")
		assert.Contains(t, health.Checks, "act")
	})
}

// Test helpers

type testBridgeContext struct {
	ctx        context.Context
	lastStatus *api.ActionResult
}

func (t *testBridgeContext) Context() context.Context {
	return t.ctx
}

func (t *testBridgeContext) GetServerURL() string {
	return "https://test.confighub.com"
}

func (t *testBridgeContext) GetWorkerID() string {
	return "test-worker-123"
}

func (t *testBridgeContext) SendStatus(status *api.ActionResult) error {
	t.lastStatus = status
	return nil
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
