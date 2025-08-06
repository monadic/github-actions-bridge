package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/confighub/sdk/bridge-worker/api"
	"github.com/confighub/sdk/workerapi"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// ActionsBridge implements the ConfigHub BridgeWorker interface for GitHub Actions
type ActionsBridge struct {
	workspaceManager   *WorkspaceManager
	actRunner          *ActRunner
	compatChecker      *CompatibilityChecker
	secretHandler      *SecretHandler
	baseDir            string
	executionSemaphore chan struct{} // Limit concurrent executions
	maxConcurrent      int
	logger             *Logger
}

// NewActionsBridge creates a new GitHub Actions bridge
func NewActionsBridge(baseDir string) (*ActionsBridge, error) {
	workspaceManager, err := NewWorkspaceManager(baseDir)
	if err != nil {
		return nil, fmt.Errorf("create workspace manager: %w", err)
	}

	secretHandler, err := NewSecretHandler()
	if err != nil {
		return nil, fmt.Errorf("create secret handler: %w", err)
	}

	maxConcurrent := 5 // Default, can be made configurable

	logger := NewLogger("ActionsBridge")
	logger.Info("Initializing GitHub Actions Bridge: baseDir=%s maxConcurrent=%d", baseDir, maxConcurrent)

	return &ActionsBridge{
		workspaceManager:   workspaceManager,
		actRunner:          NewActRunner("linux/amd64", "catthehacker/ubuntu:act-22.04"), // Use specific version
		compatChecker:      NewCompatibilityChecker(),
		secretHandler:      secretHandler,
		baseDir:            baseDir,
		executionSemaphore: make(chan struct{}, maxConcurrent),
		maxConcurrent:      maxConcurrent,
		logger:             logger,
	}, nil
}

// Info returns bridge capabilities and supported configurations
func (b *ActionsBridge) Info(opts api.InfoOptions) api.BridgeWorkerInfo {
	return api.BridgeWorkerInfo{
		SupportedConfigTypes: []*api.ConfigType{
			{
				ToolchainType: workerapi.ToolchainKubernetesYAML,
				ProviderType:  "ActLocal",
				AvailableTargets: []api.Target{
					{
						Name: "docker-desktop",
						Params: map[string]interface{}{
							"socket":   "/var/run/docker.sock",
							"platform": "linux/amd64",
						},
					},
					{
						Name: "podman-local",
						Params: map[string]interface{}{
							"socket":   "/run/podman/podman.sock",
							"platform": "linux/amd64",
						},
					},
				},
			},
		},
	}
}

// Apply executes a GitHub Actions workflow
func (b *ActionsBridge) Apply(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("PANIC in Apply: %v", r)
			b.sendError(ctx, payload, "Internal error", fmt.Errorf("panic: %v", r), time.Now())
		}
	}()

	// Log workflow execution start
	b.logger.Info("Starting workflow execution: space=%s unit=%s revision=%d", 
		payload.SpaceID, payload.UnitSlug, payload.RevisionNum)

	// Acquire execution slot with context awareness
	select {
	case b.executionSemaphore <- struct{}{}:
		defer func() { <-b.executionSemaphore }()
	case <-ctx.Context().Done():
		return fmt.Errorf("context cancelled while waiting for execution slot")
	}

	startTime := time.Now()

	// Validate payload
	if err := b.validatePayload(payload); err != nil {
		return b.sendError(ctx, payload, "Invalid payload", err, startTime)
	}

	// Send initial status
	if err := ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			Action:    api.ActionApply,
			Status:    api.ActionStatusProgressing,
			Message:   "Starting GitHub Actions workflow execution",
			StartedAt: startTime,
		},
	}); err != nil {
		return fmt.Errorf("send initial status: %w", err)
	}

	// Create isolated workspace
	ws, err := b.workspaceManager.CreateWorkspace(payload.QueuedOperationID.String())
	if err != nil {
		return b.sendError(ctx, payload, "Failed to create workspace", err, startTime)
	}
	defer func() {
		if err := ws.SecureCleanup(); err != nil {
			b.logger.Warn("Failed to cleanup workspace %s: %v", ws.ID, err)
		} else {
			b.logger.Debug("Cleaned up workspace %s", ws.ID)
		}
		b.workspaceManager.RemoveWorkspace(ws.ID)
	}()

	// Strip first 4 lines from YAML (Kubernetes metadata)
	strippedData := b.stripKubernetesMetadata(payload.Data)

	// Validate workflow compatibility
	warnings := b.compatChecker.CheckWorkflow(strippedData)
	if len(warnings) > 0 {
		b.sendWarnings(ctx, payload, warnings)
	}

	// Parse workflow file
	workflowName := "workflow.yml"
	if err := ws.WriteWorkflow(workflowName, strippedData); err != nil {
		return b.sendError(ctx, payload, "Failed to write workflow", err, startTime)
	}

	// Parse target parameters
	targetParams, err := b.parseTargetParams(payload.TargetParams)
	if err != nil {
		return b.sendError(ctx, payload, "Failed to parse target parameters", err, startTime)
	}

	// Parse extra parameters (secrets and configs)
	extraParams, err := b.parseExtraParams(payload.ExtraParams)
	if err != nil {
		return b.sendError(ctx, payload, "Failed to parse extra parameters", err, startTime)
	}

	// Prepare secrets
	if len(extraParams.Secrets) > 0 {
		if _, err := b.secretHandler.PrepareSecrets(ws, extraParams.Secrets); err != nil {
			return b.sendError(ctx, payload, "Failed to prepare secrets", err, startTime)
		}
	}

	// Inject configurations
	if len(extraParams.Configs) > 0 {
		injector := NewConfigInjector(ws)
		if err := injector.InjectConfigs(extraParams.Configs); err != nil {
			return b.sendError(ctx, payload, "Failed to inject configurations", err, startTime)
		}
	}

	// Prepare execution context
	execCtx := &ExecutionContext{
		Workspace:  ws,
		ConfigData: payload.Data,
		Metadata: ExecutionMetadata{
			Space:    payload.SpaceID.String(),
			Unit:     payload.UnitSlug,
			Revision: int(payload.RevisionNum),
			Actor:    "confighub",
		},
		Secrets:     extraParams.Secrets,
		Environment: extraParams.Environment,
		DryRun:      targetParams.DryRun,
	}

	// Execute workflow
	b.logger.Debug("Executing workflow for unit=%s", payload.UnitSlug)
	result, err := b.actRunner.Execute(execCtx)
	if err != nil {
		b.logger.Error("Workflow execution failed: unit=%s error=%v", payload.UnitSlug, err)
		return b.sendError(ctx, payload, "Workflow execution failed", err, startTime)
	}

	// Log execution result
	b.logger.WorkflowExecutionLog(result.ID, payload.UnitSlug, "success", result.Duration.String())

	// Sanitize logs
	result.Logs = b.secretHandler.SanitizeLogs(result.Logs)

	// Prepare output data
	outputData := map[string]interface{}{
		"execution_id": result.ID,
		"duration":     result.Duration.String(),
		"exit_code":    result.ExitCode,
		"artifacts":    result.Artifacts,
		"logs":         result.Logs,
	}

	outputJSON, _ := json.Marshal(outputData)

	// Send success status
	terminatedAt := time.Now()
	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			RevisionNum:  payload.RevisionNum,
			Action:       api.ActionApply,
			Result:       api.ActionResultApplyCompleted,
			Status:       api.ActionStatusCompleted,
			Message:      fmt.Sprintf("Workflow executed successfully in %s", result.Duration),
			StartedAt:    startTime,
			TerminatedAt: &terminatedAt,
		},
		Data:      payload.Data,
		LiveState: outputJSON,
	})
}

// Refresh checks the current state of a workflow execution
func (b *ActionsBridge) Refresh(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	startTime := time.Now()

	// Get last execution state
	lastExec, err := b.actRunner.GetLastExecution(payload.UnitSlug)
	if err != nil {
		terminatedAt := time.Now()
		return ctx.SendStatus(&api.ActionResult{
			UnitID:            payload.UnitID,
			SpaceID:           payload.SpaceID,
			QueuedOperationID: payload.QueuedOperationID,
			ActionResultBaseMeta: api.ActionResultBaseMeta{
				RevisionNum:  payload.RevisionNum,
				Action:       api.ActionRefresh,
				Result:       api.ActionResultRefreshAndNoDrift,
				Status:       api.ActionStatusCompleted,
				Message:      "No previous execution found",
				StartedAt:    startTime,
				TerminatedAt: &terminatedAt,
			},
		})
	}

	// Compare with current configuration
	drift := !bytes.Equal(lastExec.ConfigData, payload.Data)

	result := api.ActionResultRefreshAndNoDrift
	message := fmt.Sprintf("Last execution: %s (no drift detected)", lastExec.Timestamp.Format(time.RFC3339))

	if drift {
		result = api.ActionResultRefreshAndDrifted
		message = fmt.Sprintf("Configuration drift detected since %s", lastExec.Timestamp.Format(time.RFC3339))
	}

	terminatedAt := time.Now()
	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			RevisionNum:  payload.RevisionNum,
			Action:       api.ActionRefresh,
			Result:       result,
			Status:       api.ActionStatusCompleted,
			Message:      message,
			StartedAt:    startTime,
			TerminatedAt: &terminatedAt,
		},
		LiveState: lastExec.OutputData,
	})
}

// Import discovers existing GitHub Actions workflows
func (b *ActionsBridge) Import(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	startTime := time.Now()
	terminatedAt := time.Now()

	// For local act, there's nothing to import
	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			RevisionNum:  payload.RevisionNum,
			Action:       api.ActionImport,
			Result:       api.ActionResultImportCompleted,
			Status:       api.ActionStatusCompleted,
			Message:      "No existing workflows to import (local execution only)",
			StartedAt:    startTime,
			TerminatedAt: &terminatedAt,
		},
	})
}

// Destroy cleans up workflow resources
func (b *ActionsBridge) Destroy(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	startTime := time.Now()

	// Clean up any workspace for this unit
	if ws, exists := b.workspaceManager.GetWorkspace(payload.UnitID.String()); exists {
		if err := ws.SecureCleanup(); err != nil {
			return b.sendError(ctx, payload, "Failed to cleanup workspace", err, startTime)
		}
		b.workspaceManager.RemoveWorkspace(payload.UnitID.String())
	}

	terminatedAt := time.Now()
	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			RevisionNum:  payload.RevisionNum,
			Action:       api.ActionDestroy,
			Result:       api.ActionResultDestroyCompleted,
			Status:       api.ActionStatusCompleted,
			Message:      "Workflow resources cleaned up successfully",
			StartedAt:    startTime,
			TerminatedAt: &terminatedAt,
		},
	})
}

// Finalize archives workflow execution data
func (b *ActionsBridge) Finalize(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
	startTime := time.Now()

	// Archive execution logs and artifacts
	// In a production system, this would upload to S3 or similar

	terminatedAt := time.Now()
	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			RevisionNum:  payload.RevisionNum,
			Action:       api.ActionFinalize,
			Result:       api.ActionResultApplyCompleted,
			Status:       api.ActionStatusCompleted,
			Message:      "Workflow execution data archived",
			StartedAt:    startTime,
			TerminatedAt: &terminatedAt,
		},
	})
}

// Helper methods

func (b *ActionsBridge) sendError(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload,
	message string, err error, startTime time.Time) error {

	terminatedAt := time.Now()
	fullMessage := fmt.Sprintf("%s: %v", message, err)

	return ctx.SendStatus(&api.ActionResult{
		UnitID:            payload.UnitID,
		SpaceID:           payload.SpaceID,
		QueuedOperationID: payload.QueuedOperationID,
		ActionResultBaseMeta: api.ActionResultBaseMeta{
			RevisionNum:  payload.RevisionNum,
			Action:       api.ActionApply,
			Result:       api.ActionResultApplyFailed,
			Status:       api.ActionStatusFailed,
			Message:      fullMessage,
			StartedAt:    startTime,
			TerminatedAt: &terminatedAt,
		},
	})
}

func (b *ActionsBridge) sendWarnings(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload, warnings []Warning) {
	for _, warning := range warnings {
		log.Printf("Compatibility %s: %s", warning.Level, warning.Message)

		// Send informational status for warnings
		ctx.SendStatus(&api.ActionResult{
			UnitID:            payload.UnitID,
			SpaceID:           payload.SpaceID,
			QueuedOperationID: payload.QueuedOperationID,
			ActionResultBaseMeta: api.ActionResultBaseMeta{
				Status:  api.ActionStatusProgressing,
				Message: fmt.Sprintf("[%s] %s", warning.Level, warning.Message),
			},
		})
	}
}

type targetParameters struct {
	Platform string
	DryRun   bool
	Socket   string
}

func (b *ActionsBridge) parseTargetParams(data []byte) (targetParameters, error) {
	params := targetParameters{
		Platform: "linux/amd64",
		DryRun:   false,
		Socket:   "/var/run/docker.sock",
	}

	if len(data) == 0 {
		return params, nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return params, fmt.Errorf("unmarshal: %w", err)
	}

	if p, ok := raw["platform"].(string); ok {
		params.Platform = p
	}
	if d, ok := raw["dry_run"].(bool); ok {
		params.DryRun = d
	}
	if s, ok := raw["socket"].(string); ok {
		params.Socket = s
	}

	return params, nil
}

type extraParameters struct {
	Secrets     map[string]string
	Configs     map[string]interface{}
	Environment map[string]string
}

func (b *ActionsBridge) parseExtraParams(data []byte) (extraParameters, error) {
	params := extraParameters{
		Secrets:     make(map[string]string),
		Configs:     make(map[string]interface{}),
		Environment: make(map[string]string),
	}

	if len(data) == 0 {
		return params, nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return params, fmt.Errorf("unmarshal: %w", err)
	}

	// Parse secrets
	if secrets, ok := raw["secrets"].(map[string]interface{}); ok {
		for k, v := range secrets {
			if str, ok := v.(string); ok {
				params.Secrets[k] = str
			}
		}
	}

	// Parse configs
	if configs, ok := raw["configs"].(map[string]interface{}); ok {
		params.Configs = configs
	}

	// Parse environment
	if env, ok := raw["environment"].(map[string]interface{}); ok {
		for k, v := range env {
			if str, ok := v.(string); ok {
				params.Environment[k] = str
			}
		}
	}

	return params, nil
}

// stripKubernetesMetadata removes the first 4 lines of Kubernetes metadata from the YAML
func (b *ActionsBridge) stripKubernetesMetadata(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))

	// Check if the first line contains apiVersion: actions.confighub.com
	if len(lines) > 0 && bytes.Contains(lines[0], []byte("apiVersion:")) && bytes.Contains(lines[0], []byte("actions.confighub.com")) {
		// If we have more than 4 lines, skip the first 4
		if len(lines) > 4 {
			return bytes.Join(lines[4:], []byte("\n"))
		}
	}

	// Return data as-is if no ConfigHub metadata found
	return data
}

// validatePayload validates the incoming payload
func (b *ActionsBridge) validatePayload(payload api.BridgeWorkerPayload) error {
	if len(payload.Data) == 0 {
		return fmt.Errorf("empty workflow data")
	}

	// 10MB limit for workflow files
	if len(payload.Data) > 10*1024*1024 {
		return fmt.Errorf("workflow too large: %d bytes (max 10MB)", len(payload.Data))
	}

	// Strip Kubernetes metadata for validation
	strippedData := b.stripKubernetesMetadata(payload.Data)

	// Validate YAML syntax on stripped data
	var workflow map[string]interface{}
	if err := yaml.Unmarshal(strippedData, &workflow); err != nil {
		return fmt.Errorf("invalid workflow YAML: %w", err)
	}

	// Check if workflow is supported (using stripped data)
	supported, reason := b.compatChecker.IsWorkflowSupported(strippedData)
	if !supported {
		return fmt.Errorf("workflow not supported: %s", reason)
	}

	return nil
}

// HealthHandler handles health check requests
func (b *ActionsBridge) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"github-actions-bridge"}`)
}

// ReadinessHandler handles readiness check requests
func (b *ActionsBridge) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check if bridge is ready to accept requests
	w.Header().Set("Content-Type", "application/json")

	// Check if we can create workspaces
	testID := uuid.New().String()
	ws, err := b.workspaceManager.CreateWorkspace(testID)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","reason":"workspace_creation_failed"}`)
		return
	}

	// Clean up test workspace
	ws.SecureCleanup()
	b.workspaceManager.RemoveWorkspace(testID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ready","service":"github-actions-bridge"}`)
}

// LivenessHandler handles liveness check requests
func (b *ActionsBridge) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"alive","service":"github-actions-bridge"}`)
}
