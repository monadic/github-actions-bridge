package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nektos/act/pkg/model"
	"github.com/nektos/act/pkg/runner"
)

// ActRunner wraps the act library for workflow execution
type ActRunner struct {
	platform        string
	containerImage  string
	reuseContainers bool
	executions      sync.Map // map[string]*ExecutionRecord
}

// ExecutionRecord tracks a workflow execution
type ExecutionRecord struct {
	ID         string
	UnitID     string
	StartTime  time.Time
	EndTime    time.Time
	ExitCode   int
	ConfigData []byte
	OutputData []byte
	Logs       []string
	Artifacts  []string
	Timestamp  time.Time
}

// NewActRunner creates a new act runner
func NewActRunner(platform, containerImage string) *ActRunner {
	return &ActRunner{
		platform:        platform,
		containerImage:  containerImage,
		reuseContainers: false,
	}
}

// getContainerOptions returns container options including volume mounts
func (ar *ActRunner) getContainerOptions() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home dir, return empty options
		return ""
	}

	// Mount ~/.confighub to container's root and runner home directories
	// This ensures it's available regardless of which user runs the workflow
	// Act containers typically run as root, but we mount to both locations for compatibility
	confighubPath := filepath.Join(homeDir, ".confighub")

	// Check if .confighub directory exists
	if _, err := os.Stat(confighubPath); os.IsNotExist(err) {
		// No .confighub directory, no need to mount
		return ""
	}

	// Mount to both /root/.confighub and /home/runner/.confighub for compatibility
	return fmt.Sprintf("-v %s:/root/.confighub:ro -v %s:/home/runner/.confighub:ro", confighubPath, confighubPath)
}

// Validate checks if a GitHub Actions workflow is valid
func (ar *ActRunner) Validate(workflowData []byte) error {
	// Parse workflow
	tempFile, err := os.CreateTemp("", "workflow-*.yml")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if err := os.WriteFile(tempFile.Name(), workflowData, 0644); err != nil {
		return fmt.Errorf("write workflow: %w", err)
	}

	// Read and parse workflow
	planner, err := model.NewWorkflowPlanner(tempFile.Name(), false, false)
	if err != nil {
		return fmt.Errorf("parse workflow: %w", err)
	}

	// Try to get plan for workflow_dispatch event
	_, err = planner.PlanEvent("workflow_dispatch")
	if err != nil {
		return fmt.Errorf("plan workflow: %w", err)
	}

	return nil
}

// ExecutionResult contains the results of a workflow execution
type ExecutionResult struct {
	ID         string
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	ExitCode   int
	Logs       []string
	Artifacts  []string
	OutputData []byte
}

// Execute runs a GitHub Actions workflow
func (ar *ActRunner) Execute(ctx *ExecutionContext) (*ExecutionResult, error) {
	execID := uuid.New().String()
	result := &ExecutionResult{
		ID:        execID,
		StartTime: time.Now(),
		Logs:      []string{},
		Artifacts: []string{},
	}

	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			result.ExitCode = -1
			result.Logs = append(result.Logs, fmt.Sprintf("PANIC: %v", r))
			log.Printf("PANIC in Execute: %v", r)
		}
	}()

	// Prepare event file
	eventPath, err := ar.prepareEvent(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare event: %w", err)
	}

	// Create act runner config
	config := &runner.Config{
		EventPath: eventPath,
		EventName: "workflow_dispatch", // Default event
		Platforms: map[string]string{
			"ubuntu-latest": ar.containerImage,
			"ubuntu-22.04":  ar.containerImage,
			"ubuntu-20.04":  ar.containerImage,
			"ubuntu-18.04":  ar.containerImage,
		},
		Secrets:            ctx.Secrets,
		Env:                ctx.Environment,
		Privileged:         false,
		UsernsMode:         "auto",
		ReuseContainers:    ar.reuseContainers,
		BindWorkdir:        false,
		Workdir:            ctx.Workspace.Root,
		ArtifactServerPath: ctx.Workspace.OutputDir,
		Actor:              "confighub",
		InsecureSecrets:    false,
		LogOutput:          true,
		ContainerOptions:   ar.getContainerOptions(),
	}

	// Create runner
	actRunner, err := runner.New(config)
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	// Run the workflow
	runnerCtx := context.Background()

	// Read workflow file
	workflowPath := filepath.Join(ctx.Workspace.WorkflowDir, "workflow.yml")
	planner, err := model.NewWorkflowPlanner(workflowPath, false, false)
	if err != nil {
		return nil, fmt.Errorf("create workflow planner: %w", err)
	}

	// Get the plan
	plan, err := planner.PlanEvent(config.EventName)
	if err != nil {
		return nil, fmt.Errorf("plan event: %w", err)
	}

	// Create executor
	executor := actRunner.NewPlanExecutor(plan).Finally(func(_ context.Context) error {
		return nil
	})

	// Execute the plan
	err = executor(runnerCtx)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.ExitCode = 1
		// Don't return error, capture it in result
		result.Logs = append(result.Logs, fmt.Sprintf("ERROR: %v", err))
	}

	// Collect artifacts
	artifacts, _ := ctx.Workspace.GetArtifacts()
	result.Artifacts = artifacts

	// Store execution record
	record := &ExecutionRecord{
		ID:         execID,
		UnitID:     ctx.Metadata.Unit,
		StartTime:  result.StartTime,
		EndTime:    result.EndTime,
		ExitCode:   result.ExitCode,
		ConfigData: ctx.ConfigData,
		OutputData: result.OutputData,
		Logs:       result.Logs,
		Artifacts:  result.Artifacts,
		Timestamp:  time.Now(),
	}
	ar.executions.Store(ctx.Metadata.Unit, record)

	return result, nil
}

// GetLastExecution retrieves the last execution for a unit
func (ar *ActRunner) GetLastExecution(unitID string) (*ExecutionRecord, error) {
	if val, ok := ar.executions.Load(unitID); ok {
		return val.(*ExecutionRecord), nil
	}
	return nil, fmt.Errorf("no execution found for unit %s", unitID)
}

// prepareEvent creates the GitHub event JSON file
func (ar *ActRunner) prepareEvent(ctx *ExecutionContext) (string, error) {
	// Default event payload
	event := map[string]interface{}{
		"action": "workflow_dispatch",
		"inputs": map[string]interface{}{
			"space":    ctx.Metadata.Space,
			"unit":     ctx.Metadata.Unit,
			"revision": fmt.Sprintf("%d", ctx.Metadata.Revision),
		},
		"repository": map[string]interface{}{
			"name":      ctx.Metadata.Unit,
			"full_name": fmt.Sprintf("confighub/%s/%s", ctx.Metadata.Space, ctx.Metadata.Unit),
			"owner": map[string]interface{}{
				"login": "confighub",
			},
		},
		"sender": map[string]interface{}{
			"login": ctx.Metadata.Actor,
			"type":  "User",
		},
		"ref": "refs/heads/main",
		"sha": fmt.Sprintf("%040d", ctx.Metadata.Revision),
	}

	// Merge with custom event payload
	for k, v := range ctx.EventPayload {
		event[k] = v
	}

	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return "", err
	}

	// Act expects event files in .github directory
	githubDir := filepath.Join(ctx.Workspace.Root, ".github")
	if err := os.MkdirAll(githubDir, 0755); err != nil {
		return "", fmt.Errorf("create .github dir: %w", err)
	}

	eventPath := filepath.Join(githubDir, "event.json")
	return eventPath, os.WriteFile(eventPath, data, 0644)
}

// prepareSecrets creates the secrets file
func (ar *ActRunner) prepareSecrets(ctx *ExecutionContext) (string, error) {
	secretsPath := filepath.Join(ctx.Workspace.SecretDir, ".secrets")

	file, err := os.Create(secretsPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	for k, v := range ctx.Secrets {
		fmt.Fprintf(file, "%s=%s\n", k, v)
	}

	return secretsPath, nil
}

// prepareEnvironment creates the environment file
func (ar *ActRunner) prepareEnvironment(ctx *ExecutionContext) (string, error) {
	envPath := filepath.Join(ctx.Workspace.Root, ".env")

	file, err := os.Create(envPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	for k, v := range ctx.Environment {
		fmt.Fprintf(file, "%s=%s\n", k, v)
	}

	return envPath, nil
}

// formatSecrets converts secrets to act format
func (ar *ActRunner) formatSecrets(secrets map[string]string) []string {
	var result []string
	for k, v := range secrets {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// formatEnvs converts environment to act format
func (ar *ActRunner) formatEnvs(env map[string]string) []string {
	var result []string
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}
