package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nektos/act/pkg/runner"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ActRunner wraps the act tool for workflow execution
type ActRunner struct {
	mu         sync.RWMutex
	executions map[string]*ExecutionRecord
	config     *ActConfig
}

// ActConfig contains act configuration options
type ActConfig struct {
	DefaultImage   string
	Platforms      map[string]string
	MaxConcurrent  int
	CacheDir       string
	ArtifactServer string
	NoOutput       bool
	Verbose        bool
}

// ExecutionRecord stores information about a workflow execution
type ExecutionRecord struct {
	ID         string
	UnitID     string
	ConfigData []byte
	OutputData []byte
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	ExitCode   int
	Artifacts  []string
	Logs       []byte
	Timestamp  time.Time
}

// ExecutionContext contains all information needed to execute a workflow
type ExecutionContext struct {
	Workspace  *Workspace
	ConfigData []byte
	Metadata   *WorkflowMetadata
	
	// GitHub context simulation
	Space      string
	Unit       string
	Revision   int
	Event      string
	EventPath  string
	Actor      string
	Repository string
	Ref        string
	SHA        string
	Platform   string
	Runner     string
}

// ExecutionResult contains the results of a workflow execution
type ExecutionResult struct {
	ID         string
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	ExitCode   int
	OutputData []byte
	Artifacts  []string
	Logs       []byte
}

// NewActRunner creates a new act runner instance
func NewActRunner() *ActRunner {
	return &ActRunner{
		executions: make(map[string]*ExecutionRecord),
		config: &ActConfig{
			DefaultImage: "catthehacker/ubuntu:act-latest",
			Platforms: map[string]string{
				"ubuntu-latest": "catthehacker/ubuntu:act-latest",
				"ubuntu-22.04":  "catthehacker/ubuntu:act-22.04",
				"ubuntu-20.04":  "catthehacker/ubuntu:act-20.04",
			},
			MaxConcurrent: 5,
			CacheDir:      "/tmp/act-cache",
			NoOutput:      false,
			Verbose:       true,
		},
	}
}

// Execute runs a GitHub Actions workflow using act
func (ar *ActRunner) Execute(ctx *ExecutionContext, secretsPath string) (*ExecutionResult, error) {
	logger := log.Log
	logger.Info("Starting workflow execution", "unit", ctx.Unit, "revision", ctx.Revision)

	// Prepare the event JSON
	eventPath, err := ctx.PrepareEvent()
	if err != nil {
		return nil, fmt.Errorf("prepare event: %w", err)
	}

	// Write the workflow file
	if err := ctx.Workspace.WriteWorkflow("workflow.yml", ctx.ConfigData); err != nil {
		return nil, fmt.Errorf("write workflow: %w", err)
	}

	// Prepare act configuration
	runnerConfig := &runner.Config{
		EventName:    ctx.Event,
		EventPath:    eventPath,
		DefaultBranch: "main",
		Workdir:      ctx.Workspace.Root,
		BindWorkdir:  true,
		Platforms:    ar.config.Platforms,
		Secrets:      secretsPath,
		Env:          ar.prepareEnvironment(ctx),
		NoOutput:     ar.config.NoOutput,
		Verbose:      ar.config.Verbose,
		UseNewActionCache: true,
		ActionCacheDir: ar.config.CacheDir,
	}

	// Create log buffer
	logBuffer := &bytes.Buffer{}
	
	// Create runner
	r, err := runner.New(runnerConfig)
	if err != nil {
		return nil, fmt.Errorf("create runner: %w", err)
	}

	// Set up log capture
	r.Logger = logger
	
	// Record start time
	startTime := time.Now()
	execID := fmt.Sprintf("%s-%d-%d", ctx.Unit, ctx.Revision, startTime.Unix())

	// Run the workflow
	exitCode := 0
	if err := r.Run(); err != nil {
		logger.Error(err, "Workflow execution failed")
		exitCode = 1
	}

	// Record end time
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Collect artifacts
	artifacts, err := ctx.Workspace.ListOutputs()
	if err != nil {
		logger.Error(err, "Failed to list outputs")
		artifacts = []string{}
	}

	// Create execution result
	result := &ExecutionResult{
		ID:         execID,
		StartTime:  startTime,
		EndTime:    endTime,
		Duration:   duration,
		ExitCode:   exitCode,
		OutputData: ctx.ConfigData, // For now, return the same data
		Artifacts:  artifacts,
		Logs:       logBuffer.Bytes(),
	}

	// Store execution record
	ar.storeExecution(ctx.Unit, result, ctx.ConfigData)

	return result, nil
}

// PrepareEvent creates the GitHub event JSON for the workflow
func (ec *ExecutionContext) PrepareEvent() (string, error) {
	if ec.Event == "" {
		ec.Event = "workflow_dispatch"
	}

	// Create GitHub-compatible event
	event := map[string]interface{}{
		"action": ec.Event,
		"inputs": map[string]string{
			"space":    ec.Space,
			"unit":     ec.Unit,
			"revision": fmt.Sprintf("%d", ec.Revision),
		},
		"repository": map[string]interface{}{
			"name":       ec.Unit,
			"full_name":  fmt.Sprintf("confighub/%s/%s", ec.Space, ec.Unit),
			"owner": map[string]interface{}{
				"login": "confighub",
			},
		},
		"sender": map[string]interface{}{
			"login": ec.Actor,
		},
		"ref": ec.Ref,
		"sha": ec.SHA,
	}

	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return "", err
	}

	eventPath := filepath.Join(ec.Workspace.Root, "event.json")
	return eventPath, os.WriteFile(eventPath, data, 0644)
}

// prepareEnvironment creates environment variables for act
func (ar *ActRunner) prepareEnvironment(ctx *ExecutionContext) map[string]string {
	env := map[string]string{
		// GitHub environment variables
		"GITHUB_WORKFLOW":    "ConfigHub Workflow",
		"GITHUB_RUN_ID":      fmt.Sprintf("%d", time.Now().Unix()),
		"GITHUB_RUN_NUMBER":  fmt.Sprintf("%d", ctx.Revision),
		"GITHUB_ACTION":      "__run",
		"GITHUB_ACTIONS":     "true",
		"GITHUB_ACTOR":       ctx.Actor,
		"GITHUB_REPOSITORY":  ctx.Repository,
		"GITHUB_EVENT_NAME":  ctx.Event,
		"GITHUB_EVENT_PATH":  filepath.Join(ctx.Workspace.Root, "event.json"),
		"GITHUB_WORKSPACE":   ctx.Workspace.Root,
		"GITHUB_SHA":         ctx.SHA,
		"GITHUB_REF":         ctx.Ref,
		"RUNNER_OS":          "Linux",
		"RUNNER_TEMP":        filepath.Join(ctx.Workspace.Root, "runner", "temp"),
		"RUNNER_TOOL_CACHE":  filepath.Join(ctx.Workspace.Root, "runner", "tool_cache"),
		
		// ConfigHub-specific variables
		"CONFIGHUB_SPACE":    ctx.Space,
		"CONFIGHUB_UNIT":     ctx.Unit,
		"CONFIGHUB_REVISION": fmt.Sprintf("%d", ctx.Revision),
	}

	// Add config values as environment variables
	for key, value := range ctx.Metadata.Configs {
		if str, ok := value.(string); ok {
			env[fmt.Sprintf("CONFIG_%s", key)] = str
		}
	}

	return env
}

// GetLastExecution retrieves the last execution record for a unit
func (ar *ActRunner) GetLastExecution(unitID string) (*ExecutionRecord, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	if exec, ok := ar.executions[unitID]; ok {
		return exec, nil
	}

	return nil, fmt.Errorf("no execution found for unit %s", unitID)
}

// CleanupExecution removes execution records for a unit
func (ar *ActRunner) CleanupExecution(unitID string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	delete(ar.executions, unitID)
	return nil
}

// storeExecution saves an execution record
func (ar *ActRunner) storeExecution(unitID string, result *ExecutionResult, configData []byte) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	ar.executions[unitID] = &ExecutionRecord{
		ID:         result.ID,
		UnitID:     unitID,
		ConfigData: configData,
		OutputData: result.OutputData,
		StartTime:  result.StartTime,
		EndTime:    result.EndTime,
		Duration:   result.Duration,
		ExitCode:   result.ExitCode,
		Artifacts:  result.Artifacts,
		Logs:       result.Logs,
		Timestamp:  time.Now(),
	}
}

// SetConfig updates the act runner configuration
func (ar *ActRunner) SetConfig(config *ActConfig) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.config = config
}

// StreamLogs implements log streaming for real-time output
type LogStreamer struct {
	writer io.Writer
	buffer *bytes.Buffer
}

func NewLogStreamer(writer io.Writer) *LogStreamer {
	return &LogStreamer{
		writer: writer,
		buffer: &bytes.Buffer{},
	}
}

func (ls *LogStreamer) Write(p []byte) (n int, err error) {
	// Write to both the writer and buffer
	if ls.writer != nil {
		ls.writer.Write(p)
	}
	return ls.buffer.Write(p)
}

func (ls *LogStreamer) String() string {
	return ls.buffer.String()
}
