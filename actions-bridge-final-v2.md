# GitHub Actions Bridge - Final Specification v2

*Author: Brian Grant*  
*Incorporating feedback from Jesper Joergensen*  
*Date: January 2024*

## Executive Summary

A ConfigHub bridge that enables local execution of GitHub Actions workflows using `act`, with proper workspace isolation, secret handling, and full bridge interface implementation. Designed for incremental delivery with explicit handling of act limitations.

## Core Architecture Insights

1. **Act is Quirky**: We document and handle its limitations explicitly
2. **Workspaces are Critical**: Every execution gets isolated workspace with cleanup
3. **Secrets via Files**: Never environment variables
4. **Full Bridge Pattern**: Implement all methods, not just Execute

## Revised Architecture

```
ConfigHub API
     │
     ▼
┌─────────────────────────────────┐
│   GitHub Actions Bridge Worker   │
│  ┌────────────────────────────┐  │
│  │   Bridge Interface         │  │
│  │   - Info()                 │  │
│  │   - Apply() → Execute      │  │
│  │   - Refresh() → Status     │  │
│  │   - Destroy() → Cleanup    │  │
│  │   - Import() → Discover    │  │
│  │   - Finalize() → Archive   │  │
│  └────────────────────────────┘  │
│  ┌────────────────────────────┐  │
│  │   Workspace Manager        │  │
│  │   - Isolation per exec     │  │
│  │   - Secure cleanup         │  │
│  │   - Audit trail           │  │
│  └────────────────────────────┘  │
│  ┌────────────────────────────┐  │
│  │   Act Wrapper              │  │
│  │   - Compatibility layer    │  │
│  │   - Secret file handling   │  │
│  │   - Output capture         │  │
│  └────────────────────────────┘  │
└─────────────────────────────────┘
```

## Phase 0: Act Validation (3 days)

### Goal
Prove act wrapper works before any ConfigHub integration.

### Implementation

```go
// cmd/act-test/main.go
package main

import (
    "github.com/nektos/act/pkg/runner"
)

func main() {
    // Test basic act functionality
    runner := &runner.Runner{
        EventName:   "push",
        EventPath:   "event.json",
        WorkflowPath: ".github/workflows/test.yml",
        Platforms: map[string]string{
            "ubuntu-latest": "catthehacker/ubuntu:act-latest",
        },
    }
    
    if err := runner.Run(); err != nil {
        log.Fatalf("Act failed: %v", err)
    }
}
```

### Validation Checklist
- [ ] Act runs simple workflow
- [ ] Secret file injection works  
- [ ] Output capture works
- [ ] Resource cleanup works
- [ ] Document all limitations found

## Phase 1: Bridge Foundation (1 week)

### Goal
Implement full ConfigHub bridge interface with workspace isolation.

### Core Components

#### 1.1 Bridge Interface Implementation
```go
package actbridge

import (
    "github.com/confighub/sdk/bridge-worker/api"
    "github.com/confighub/sdk/workerapi"
)

type ActionsBridge struct {
    workspaceManager *WorkspaceManager
    actRunner        *ActRunner
    compatChecker    *CompatibilityChecker
}

// Full bridge interface
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
                            "socket": "/var/run/docker.sock",
                            "platform": "linux/amd64",
                        },
                    },
                },
            },
        },
        Capabilities: map[string]interface{}{
            "version": "1.0.0",
            "act_version": "0.2.65",
            "limitations": b.compatChecker.KnownLimitations(),
        },
    }
}

func (b *ActionsBridge) Apply(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
    // Create isolated workspace
    ws, err := b.workspaceManager.CreateWorkspace(payload.QueuedOperationID.String())
    if err != nil {
        return b.sendError(ctx, payload, "workspace creation failed", err)
    }
    defer ws.SecureCleanup() // Jesper's insight: secure cleanup critical

    // Validate workflow compatibility
    warnings := b.compatChecker.CheckWorkflow(payload.Data)
    if len(warnings) > 0 {
        b.sendWarnings(ctx, payload, warnings)
    }

    // Prepare execution context
    execCtx := &ExecutionContext{
        Workspace:  ws,
        ConfigData: payload.Data,
        Metadata:   b.extractMetadata(payload),
    }

    // Execute workflow
    result, err := b.actRunner.Execute(execCtx)
    if err != nil {
        return b.sendError(ctx, payload, "execution failed", err)
    }

    // Send success with execution details
    return ctx.SendStatus(&api.ActionResult{
        UnitID:            payload.UnitID,
        SpaceID:           payload.SpaceID,
        QueuedOperationID: payload.QueuedOperationID,
        ActionResultBaseMeta: api.ActionResultMeta{
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

func (b *ActionsBridge) Refresh(ctx api.BridgeWorkerContext, payload api.BridgeWorkerPayload) error {
    // Get last execution state
    lastExec, err := b.actRunner.GetLastExecution(payload.UnitID.String())
    if err != nil {
        return ctx.SendStatus(&api.ActionResult{
            UnitID:  payload.UnitID,
            SpaceID: payload.SpaceID,
            ActionResultBaseMeta: api.ActionResultMeta{
                Action:  api.ActionRefresh,
                Result:  api.ActionResultRefreshCompleted,
                Status:  api.ActionStatusCompleted,
                Message: "No previous execution found",
            },
            DriftDetected: false,
        })
    }

    // Compare with current
    drift := !bytes.Equal(lastExec.ConfigData, payload.Data)
    
    return ctx.SendStatus(&api.ActionResult{
        UnitID:  payload.UnitID,
        SpaceID: payload.SpaceID,
        ActionResultBaseMeta: api.ActionResultMeta{
            Action:  api.ActionRefresh,
            Result:  api.ActionResultRefreshCompleted,
            Status:  api.ActionStatusCompleted,
            Message: fmt.Sprintf("Last execution: %s", lastExec.Timestamp),
        },
        DriftDetected: drift,
        LiveState:     lastExec.OutputData,
    })
}
```

#### 1.2 Workspace Manager (Jesper's Key Insight)
```go
type WorkspaceManager struct {
    baseDir string
    mu      sync.Mutex
    active  map[string]*Workspace
}

type Workspace struct {
    ID         string
    Root       string
    WorkflowDir string
    ConfigDir  string
    SecretDir  string
    OutputDir  string
    created    time.Time
}

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

    // Create with proper permissions
    dirs := []struct {
        path string
        perm os.FileMode
    }{
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
            ws.SecureCleanup()
            delete(wm.active, execID)
        }
        wm.mu.Unlock()
    }()

    return ws, nil
}

func (ws *Workspace) SecureCleanup() error {
    // Jesper's pattern: overwrite secrets before deletion
    secretFiles, _ := filepath.Glob(filepath.Join(ws.SecretDir, "*"))
    for _, f := range secretFiles {
        if err := secureDelete(f); err != nil {
            log.Printf("Warning: failed to secure delete %s: %v", f, err)
        }
    }

    return os.RemoveAll(ws.Root)
}

func secureDelete(path string) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
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

    return os.Remove(path)
}
```

#### 1.3 Act Compatibility Layer
```go
type CompatibilityChecker struct {
    unsupportedActions map[string]string
}

func NewCompatibilityChecker() *CompatibilityChecker {
    return &CompatibilityChecker{
        unsupportedActions: map[string]string{
            "actions/cache@":              "Caching not supported locally",
            "actions/upload-artifact@":    "Artifacts saved to workspace only",
            "actions/download-artifact@":  "Cross-workflow artifacts not supported",
            "docker/build-push-action@":   "Registry push disabled locally",
        },
    }
}

func (cc *CompatibilityChecker) CheckWorkflow(workflowData []byte) []Warning {
    warnings := []Warning{}
    content := string(workflowData)

    for action, message := range cc.unsupportedActions {
        if strings.Contains(content, action) {
            warnings = append(warnings, Warning{
                Level:   "info",
                Message: fmt.Sprintf("%s: %s", action, message),
                Action:  action,
            })
        }
    }

    // Check for GitHub-specific contexts
    if strings.Contains(content, "${{ secrets.GITHUB_TOKEN }}") {
        warnings = append(warnings, Warning{
            Level:   "warning",
            Message: "GITHUB_TOKEN will be simulated locally",
        })
    }

    return warnings
}
```

## Phase 2: ConfigHub Integration (1 week)

### Goal
Add config and secret injection with proper security.

### Implementation

#### 2.1 Config Injection
```go
func (b *ActionsBridge) injectConfigs(ws *Workspace, configs map[string]interface{}) error {
    // Write configs as files (more flexible than env vars)
    for key, value := range configs {
        path := filepath.Join(ws.ConfigDir, key+".json")
        data, err := json.MarshalIndent(value, "", "  ")
        if err != nil {
            return fmt.Errorf("marshal %s: %w", key, err)
        }
        
        if err := os.WriteFile(path, data, 0644); err != nil {
            return fmt.Errorf("write %s: %w", path, err)
        }
    }

    // Also create env file for simple values
    envPath := filepath.Join(ws.ConfigDir, ".env")
    envFile, err := os.Create(envPath)
    if err != nil {
        return err
    }
    defer envFile.Close()

    for key, value := range configs {
        if str, ok := value.(string); ok {
            fmt.Fprintf(envFile, "CONFIG_%s=%s\n", 
                strings.ToUpper(key), str)
        }
    }

    return nil
}
```

#### 2.2 Secret Handling (Jesper's File Approach)
```go
func (b *ActionsBridge) prepareSecrets(ws *Workspace, secrets map[string]string) (string, error) {
    // Act-compatible secrets file
    secretsPath := filepath.Join(ws.SecretDir, ".secrets")
    
    // Create with restrictive permissions
    file, err := os.OpenFile(secretsPath, 
        os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
    if err != nil {
        return "", fmt.Errorf("create secrets file: %w", err)
    }
    defer file.Close()

    // Write in act format
    for key, value := range secrets {
        // Track for leak detection
        b.leakDetector.Track(key, value)
        
        // Write to file
        fmt.Fprintf(file, "%s=%s\n", key, value)
    }

    return secretsPath, nil
}
```

## Phase 3: Advanced Features (2 weeks)

### Goal
Production hardening with monitoring and advanced workflows.

### Features

#### 3.1 Execution Context Bridge
```go
// Jesper's insight: Bridge ConfigHub context to GitHub context
type ExecutionContext struct {
    // ConfigHub context
    Space      string
    Unit       string
    Revision   int
    
    // GitHub simulation
    Event      string
    EventPath  string
    Actor      string
    Repository string
    Ref        string
    SHA        string
    
    // Act configuration
    Platform   string
    Runner     string
}

func (ec *ExecutionContext) PrepareEvent() (string, error) {
    // Create GitHub event JSON
    event := map[string]interface{}{
        "action": "workflow_dispatch",
        "inputs": map[string]string{
            "space":    ec.Space,
            "unit":     ec.Unit,
            "revision": fmt.Sprintf("%d", ec.Revision),
        },
        "repository": map[string]interface{}{
            "name":       ec.Unit,
            "full_name":  fmt.Sprintf("confighub/%s/%s", ec.Space, ec.Unit),
        },
        "sender": map[string]interface{}{
            "login": ec.Actor,
        },
    }

    data, err := json.Marshal(event)
    if err != nil {
        return "", err
    }

    eventPath := filepath.Join(ec.Workspace.Root, "event.json")
    return eventPath, os.WriteFile(eventPath, data, 0644)
}
```

#### 3.2 Enhanced CLI (Jesper's Suggestions)
```go
var runCmd = &cobra.Command{
    Use:   "run WORKFLOW",
    Short: "Run a GitHub Actions workflow locally",
    RunE: func(cmd *cobra.Command, args []string) error {
        // All the flags Jesper identified as needed
        space, _ := cmd.Flags().GetString("space")
        unit, _ := cmd.Flags().GetString("unit")
        dryRun, _ := cmd.Flags().GetBool("dry-run")
        event, _ := cmd.Flags().GetString("event")
        inputs, _ := cmd.Flags().GetStringSlice("input")
        platform, _ := cmd.Flags().GetString("platform")
        artifactDir, _ := cmd.Flags().GetString("artifact-dir")
        envFile, _ := cmd.Flags().GetString("env-file")
        
        // Validation mode
        if validateOnly, _ := cmd.Flags().GetBool("validate"); validateOnly {
            return validateWorkflow(args[0])
        }
        
        // ... execution logic
    },
}

func init() {
    runCmd.Flags().String("space", "", "ConfigHub space")
    runCmd.Flags().String("unit", "", "ConfigHub unit")
    runCmd.Flags().Bool("dry-run", false, "Show what would be executed")
    runCmd.Flags().String("event", "workflow_dispatch", "GitHub event type")
    runCmd.Flags().StringSlice("input", nil, "Workflow inputs (key=value)")
    runCmd.Flags().String("platform", "linux/amd64", "Execution platform")
    runCmd.Flags().String("artifact-dir", "./artifacts", "Artifact output directory")
    runCmd.Flags().String("env-file", "", "Additional environment file")
    runCmd.Flags().Bool("validate", false, "Validate workflow only")
}
```

## Monitoring & Operations

### Health Monitoring
```go
func (b *ActionsBridge) HealthCheck() HealthStatus {
    status := HealthStatus{
        Healthy: true,
        Checks:  make(map[string]CheckResult),
    }

    // Docker connectivity
    dockerCheck := b.checkDocker()
    status.Checks["docker"] = dockerCheck
    if !dockerCheck.Healthy {
        status.Healthy = false
    }

    // Workspace cleanup working
    wsCheck := b.checkWorkspaceManager()
    status.Checks["workspaces"] = wsCheck

    // Disk space
    diskCheck := b.checkDiskSpace()
    status.Checks["disk"] = diskCheck
    if diskCheck.FreePercent < 10 {
        status.Healthy = false
    }

    return status
}
```

### Metrics
```yaml
metrics:
  - name: actions_executions_total
    type: counter
    labels: [workflow, space, status, platform]
    
  - name: actions_execution_duration_seconds
    type: histogram
    labels: [workflow, space]
    buckets: [10, 30, 60, 120, 300, 600]
    
  - name: actions_workspace_cleanup_duration_seconds
    type: histogram
    
  - name: actions_compatibility_warnings_total
    type: counter
    labels: [action, level]
```

## Testing Strategy

### Phase 0 Tests
```go
func TestActBasics(t *testing.T) {
    // Run without ConfigHub
    runner := NewActRunner()
    result, err := runner.Execute(&ExecutionContext{
        WorkflowPath: "testdata/simple.yml",
        EventName:    "push",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 0, result.ExitCode)
}
```

### Integration Tests
```go
func TestWorkspaceIsolation(t *testing.T) {
    manager := NewWorkspaceManager(t.TempDir())
    
    // Create multiple workspaces
    ws1, _ := manager.CreateWorkspace("exec-1")
    ws2, _ := manager.CreateWorkspace("exec-2")
    
    // Verify isolation
    assert.NotEqual(t, ws1.Root, ws2.Root)
    assert.DirExists(t, ws1.SecretDir)
    assert.DirExists(t, ws2.SecretDir)
    
    // Verify cleanup
    ws1.SecureCleanup()
    assert.NoDirExists(t, ws1.Root)
    assert.DirExists(t, ws2.Root) // Other workspace unaffected
}
```

## Success Metrics

### Phase 1 (Foundation)
- Successfully execute 10 different workflows
- Zero workspace leaks after 100 executions
- All bridge methods implemented and tested

### Phase 2 (Integration) 
- 50 workflows using ConfigHub configs
- Zero secret exposures in logs
- 100% audit coverage of secret access

### Phase 3 (Advanced)
- Support 95% of common GitHub Actions
- P99 execution start time < 5s
- Automated compatibility warnings for all known issues

## Key Improvements from v1

1. **Phase 0 Added**: Validate act before ConfigHub integration
2. **Full Bridge Pattern**: All methods implemented properly
3. **Workspace Isolation**: Jesper's secure cleanup patterns
4. **Secret Files**: Never use environment variables
5. **Compatibility Layer**: Explicit handling of act limitations
6. **Enhanced CLI**: All flags users actually need

## Conclusion

This specification incorporates hard-won production insights while maintaining architectural clarity. The phased approach now includes a "Phase 0" to validate act integration separately, reducing risk. Workspace isolation and secure cleanup are first-class concerns, not afterthoughts.

Most importantly, we acknowledge act's limitations upfront and handle them explicitly rather than discovering them in production.

---
*Brian Grant*  
*January 2024*

*P.S. Jesper - Thanks for the reality check. Your production scars saved us months of pain.*
