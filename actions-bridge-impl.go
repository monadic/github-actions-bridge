package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/confighub/actions-bridge/pkg/leakdetector"
	"github.com/confighub/sdk/bridge-worker/api"
	"github.com/confighub/sdk/workerapi"
	"github.com/google/uuid"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ActionsBridge implements the ConfigHub bridge interface for GitHub Actions
type ActionsBridge struct {
	workspaceManager *WorkspaceManager
	actRunner        *ActRunner
	compatChecker    *CompatibilityChecker
	leakDetector     *leakdetector.Detector
}

// NewActionsBridge creates a new GitHub Actions bridge instance
func NewActionsBridge(baseDir string) (*ActionsBridge, error) {
	workspaceManager, err := NewWorkspaceManager(baseDir)
	if err != nil {
		return nil, fmt.Errorf("create workspace manager: %w", err)
	}

	return &ActionsBridge{
		workspaceManager: workspaceManager,
		actRunner:        NewActRunner(),
		compatChecker:    NewCompatibilityChecker(),
		leakDetector:     leakdetector.New(),
	}, nil
}

// Info returns bridge capabilities and supported configurations
func (b *ActionsBridge) Info(opts api.InfoOptions) api.BridgeWorkerInfo {
	return api.BridgeWorkerInfo{
		SupportedConfigTypes: []*api.ConfigType{
			{
				ToolchainType: workerapi.ToolchainType("github-actions"),
				ProviderType:  api.ProviderType("act-local"),
				AvailableTargets: []api.Target{
					{
						Name: "docker-desktop",
						Params: map[string]interface{}{
							"socket":   "/var/run/docker.sock",
							"platform": "linux/amd64",
						},
					},
				},
			},
		},
		Capabilities: map[string]interface{}{
			"version":      "1.0.0",
			"act_version":  "0.2.65",
			"limitations":  b.compatChecker.KnownLimitations(),
		},
	}
}

// Apply executes a GitHub Actions workflow
func (b *ActionsBridge) Apply(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	logger := log.FromContext(ctx.Context())
	logger.Info("Starting Apply operation", "unit", payload.UnitSlug)

	// Create isolated workspace
	ws, err := b.workspaceManager.CreateWorkspace(payload.QueuedOperationID.String())
	if err != nil {
		return b.sendError(ctx, payload, "workspace creation failed", err)
	}
	defer ws.SecureCleanup()

	// Validate workflow compatibility
	warnings := b.compatChecker.CheckWorkflow(payload.Data)
	if len(warnings) > 0 {
		b.sendWarnings(ctx, payload, warnings)
	}

	// Extract metadata and prepare execution context
	metadata := b.extractMetadata(payload)
	execCtx := &ExecutionContext{
		Workspace:  ws,
		ConfigData: payload.Data,
		Metadata:   metadata,
		Space:      payload.SpaceID.String(),
		Unit:       payload.UnitSlug,
		Revision:   int(payload.RevisionNum),
		Actor:      "confighub-bridge",
		Repository: fmt.Sprintf("confighub/%s/%s", payload.SpaceID, payload.UnitSlug),
		Ref:        "refs/heads/main",
		SHA:        fmt.Sprintf("%x", payload.RevisionNum),
	}

	// Inject configurations
	if err := b.injectConfigs(ws, metadata.Configs); err != nil {
		return b.sendError(ctx, payload, "config injection failed", err)
	}

	// Prepare secrets
	secretsPath, err := b.prepareSecrets(ws, metadata.Secrets)
	if err != nil {
		return b.sendError(ctx, payload, "secret preparation failed", err)
	}

	// Execute workflow
	result, err := b.actRunner.Execute(execCtx, secretsPath)
	if err != nil {
		return b.sendError(ctx, payload, "execution failed", err)
	}

	// Send success result
	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:       api.ActionApply,
			Result:       api.ActionResultApplyCompleted,
			Status:       api.ActionStatusCompleted,
			Message:      "Workflow executed successfully",
			StartedAt:    result.StartTime,
			TerminatedAt: &result.EndTime,
		},
		Data:      payload.Data,
		LiveState: result.OutputData,
		Details: map[string]interface{}{
			"execution_id": result.ID,
			"duration":     result.Duration,
			"exit_code":    result.ExitCode,
			"artifacts":    result.Artifacts,
		},
	})
}

// Refresh checks the state of the last execution
func (b *ActionsBridge) Refresh(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	logger := log.FromContext(ctx.Context())
	logger.Info("Starting Refresh operation", "unit", payload.UnitSlug)

	// Get last execution state
	lastExec, err := b.actRunner.GetLastExecution(payload.UnitID.String())
	if err != nil {
		return ctx.SendStatus(&api.ActionResult{
			UnitID:  payload.UnitID,
			SpaceID: payload.SpaceID,
			ActionResultBaseMeta: api.ActionResultBaseMeta{
				Action:  api.ActionRefresh,
				Result:  api.ActionResultRefreshCompleted,
				Status:  api.ActionStatusCompleted,
				Message: "No previous execution found",
			},
			DriftDetected: false,
		})
	}

	// Compare with current configuration
	drift := !bytes.Equal(lastExec.ConfigData, payload.Data)
	
	return ctx.SendStatus(&api.ActionResult{
		UnitID:  payload.UnitID,
		SpaceID: payload.SpaceID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:  api.ActionRefresh,
			Result:  api.ActionResultRefreshCompleted,
			Status:  api.ActionStatusCompleted,
			Message: fmt.Sprintf("Last execution: %s", lastExec.Timestamp),
		},
		DriftDetected: drift,
		LiveState:     lastExec.OutputData,
	})
}

// Destroy cleans up resources
func (b *ActionsBridge) Destroy(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	logger := log.FromContext(ctx.Context())
	logger.Info("Starting Destroy operation", "unit", payload.UnitSlug)

	// Clean up any stored execution data
	if err := b.actRunner.CleanupExecution(payload.UnitID.String()); err != nil {
		logger.Error(err, "Failed to cleanup execution data")
	}

	return ctx.SendStatus(&api.ActionResult{
		UnitID:  payload.UnitID,
		SpaceID: payload.SpaceID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:  api.ActionDestroy,
			Result:  api.ActionResultDestroyCompleted,
			Status:  api.ActionStatusCompleted,
			Message: "Resources cleaned up successfully",
		},
		LiveState: []byte{},
	})
}

// Import discovers existing workflows
func (b *ActionsBridge) Import(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	// Not implemented for local act execution
	return ctx.SendStatus(&api.ActionResult{
		UnitID:  payload.UnitID,
		SpaceID: payload.SpaceID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:  api.ActionImport,
			Result:  api.ActionResultImportCompleted,
			Status:  api.ActionStatusCompleted,
			Message: "Import not supported for local execution",
		},
	})
}

// Finalize performs cleanup after operations
func (b *ActionsBridge) Finalize(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	// Ensure workspace cleanup
	b.workspaceManager.CleanupStale(1 * time.Hour)
	
	return ctx.SendStatus(&api.ActionResult{
		UnitID:  payload.UnitID,
		SpaceID: payload.SpaceID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:  api.ActionFinalize,
			Result:  api.ActionResultNone,
			Status:  api.ActionStatusCompleted,
			Message: "Finalization completed",
		},
	})
}

// Helper methods

func (b *ActionsBridge) sendError(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload, message string, err error) error {
	fullMessage := fmt.Sprintf("%s: %v", message, err)
	now := time.Now()
	
	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:       api.ActionApply,
			Result:       api.ActionResultApplyFailed,
			Status:       api.ActionStatusFailed,
			Message:      fullMessage,
			StartedAt:    now,
			TerminatedAt: &now,
		},
	})
}

func (b *ActionsBridge) sendWarnings(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload, warnings []Warning) {
	logger := log.FromContext(ctx.Context())
	for _, w := range warnings {
		logger.Info("Compatibility warning", "level", w.Level, "action", w.Action, "message", w.Message)
	}
}

func (b *ActionsBridge) extractMetadata(payload api.BridgeWorkerPayload) *WorkflowMetadata {
	metadata := &WorkflowMetadata{
		Configs: make(map[string]interface{}),
		Secrets: make(map[string]string),
	}

	// Extract from ExtraParams if provided
	if len(payload.ExtraParams) > 0 {
		var params map[string]interface{}
		if err := json.Unmarshal(payload.ExtraParams, &params); err == nil {
			if configs, ok := params["configs"].(map[string]interface{}); ok {
				metadata.Configs = configs
			}
			if secrets, ok := params["secrets"].(map[string]interface{}); ok {
				for k, v := range secrets {
					if str, ok := v.(string); ok {
						metadata.Secrets[k] = str
					}
				}
			}
		}
	}

	return metadata
}

// WorkflowMetadata contains extracted workflow configuration
type WorkflowMetadata struct {
	Configs map[string]interface{}
	Secrets map[string]string
}

// Ensure ActionsBridge implements the bridge interfaces
var _ api.BridgeWorker = (*ActionsBridge)(nil)
var _ api.WatchableWorker = (*ActionsBridge)(nil)

// WatchForApply monitors workflow execution progress
func (b *ActionsBridge) WatchForApply(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	// For act, execution is synchronous, so we don't need to watch
	return nil
}

// WatchForDestroy monitors cleanup progress
func (b *ActionsBridge) WatchForDestroy(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	// Cleanup is immediate for local execution
	return nil
}
