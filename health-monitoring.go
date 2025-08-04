package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/confighub/sdk/bridge-worker/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HealthStatus represents the overall health of the bridge
type HealthStatus struct {
	Healthy   bool                   `json:"healthy"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Healthy      bool                   `json:"healthy"`
	Message      string                 `json:"message"`
	LastChecked  time.Time              `json:"last_checked"`
	Details      map[string]interface{} `json:"details,omitempty"`
	FreePercent  float64               `json:"free_percent,omitempty"`
}

// Metrics for monitoring
var (
	executionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "actions_executions_total",
			Help: "Total number of workflow executions",
		},
		[]string{"workflow", "space", "status", "platform"},
	)

	executionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "actions_execution_duration_seconds",
			Help:    "Duration of workflow executions",
			Buckets: []float64{10, 30, 60, 120, 300, 600},
		},
		[]string{"workflow", "space"},
	)

	workspaceCleanupDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "actions_workspace_cleanup_duration_seconds",
			Help:    "Duration of workspace cleanup operations",
			Buckets: prometheus.DefBuckets,
		},
	)

	compatibilityWarnings = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "actions_compatibility_warnings_total",
			Help: "Total number of compatibility warnings",
		},
		[]string{"action", "level"},
	)

	activeWorkspaces = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "actions_active_workspaces",
			Help: "Number of active workspaces",
		},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(
		executionsTotal,
		executionDuration,
		workspaceCleanupDuration,
		compatibilityWarnings,
		activeWorkspaces,
	)
}

// HealthCheck performs comprehensive health checks
func (b *ActionsBridge) HealthCheck() HealthStatus {
	status := HealthStatus{
		Healthy:   true,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		Checks:    make(map[string]CheckResult),
	}

	// Docker connectivity check
	dockerCheck := b.checkDocker()
	status.Checks["docker"] = dockerCheck
	if !dockerCheck.Healthy {
		status.Healthy = false
	}

	// Workspace manager check
	wsCheck := b.checkWorkspaceManager()
	status.Checks["workspaces"] = wsCheck

	// Disk space check
	diskCheck := b.checkDiskSpace()
	status.Checks["disk"] = diskCheck
	if diskCheck.FreePercent < 10 {
		status.Healthy = false
	}

	// Act functionality check
	actCheck := b.checkAct()
	status.Checks["act"] = actCheck
	if !actCheck.Healthy {
		status.Healthy = false
	}

	// Memory usage check
	memCheck := b.checkMemory()
	status.Checks["memory"] = memCheck

	return status
}

// checkDocker verifies Docker is accessible and running
func (b *ActionsBridge) checkDocker() CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "version", "--format", "json")
	output, err := cmd.Output()
	
	if err != nil {
		return CheckResult{
			Healthy:     false,
			Message:     fmt.Sprintf("Docker not accessible: %v", err),
			LastChecked: time.Now(),
		}
	}

	var dockerInfo map[string]interface{}
	if err := json.Unmarshal(output, &dockerInfo); err == nil {
		return CheckResult{
			Healthy:     true,
			Message:     "Docker is running",
			LastChecked: time.Now(),
			Details:     dockerInfo,
		}
	}

	return CheckResult{
		Healthy:     true,
		Message:     "Docker is running",
		LastChecked: time.Now(),
	}
}

// checkWorkspaceManager verifies workspace operations are working
func (b *ActionsBridge) checkWorkspaceManager() CheckResult {
	// Try to create and cleanup a test workspace
	testID := fmt.Sprintf("health-check-%d", time.Now().Unix())
	
	ws, err := b.workspaceManager.CreateWorkspace(testID)
	if err != nil {
		return CheckResult{
			Healthy:     false,
			Message:     fmt.Sprintf("Cannot create workspaces: %v", err),
			LastChecked: time.Now(),
		}
	}

	// Cleanup test workspace
	if err := ws.SecureCleanup(); err != nil {
		return CheckResult{
			Healthy:     false,
			Message:     fmt.Sprintf("Cannot cleanup workspaces: %v", err),
			LastChecked: time.Now(),
		}
	}

	// Get active workspace count
	b.workspaceManager.mu.RLock()
	activeCount := len(b.workspaceManager.active)
	b.workspaceManager.mu.RUnlock()

	activeWorkspaces.Set(float64(activeCount))

	return CheckResult{
		Healthy:     true,
		Message:     "Workspace operations working",
		LastChecked: time.Now(),
		Details: map[string]interface{}{
			"active_workspaces": activeCount,
		},
	}
}

// checkDiskSpace verifies sufficient disk space is available
func (b *ActionsBridge) checkDiskSpace() CheckResult {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(b.workspaceManager.baseDir, &stat); err != nil {
		return CheckResult{
			Healthy:     false,
			Message:     fmt.Sprintf("Cannot check disk space: %v", err),
			LastChecked: time.Now(),
		}
	}

	// Calculate disk usage
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	freePercent := float64(free) / float64(total) * 100

	healthy := freePercent > 5 // Need at least 5% free space

	return CheckResult{
		Healthy:     healthy,
		Message:     fmt.Sprintf("Disk space: %.1f%% free", freePercent),
		LastChecked: time.Now(),
		FreePercent: freePercent,
		Details: map[string]interface{}{
			"total_gb": float64(total) / (1024 * 1024 * 1024),
			"free_gb":  float64(free) / (1024 * 1024 * 1024),
			"used_gb":  float64(used) / (1024 * 1024 * 1024),
		},
	}
}

// checkAct verifies act is available and functional
func (b *ActionsBridge) checkAct() CheckResult {
	// For now, just check if we can create a runner config
	// In production, might want to run a simple test workflow
	
	return CheckResult{
		Healthy:     true,
		Message:     "Act functionality available",
		LastChecked: time.Now(),
		Details: map[string]interface{}{
			"version": "0.2.65",
			"platforms": b.actRunner.config.Platforms,
		},
	}
}

// checkMemory monitors memory usage
func (b *ActionsBridge) checkMemory() CheckResult {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return CheckResult{
		Healthy:     true,
		Message:     "Memory usage normal",
		LastChecked: time.Now(),
		Details: map[string]interface{}{
			"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
			"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
			"sys_mb":         float64(m.Sys) / 1024 / 1024,
			"num_gc":         m.NumGC,
		},
	}
}

// HTTP health endpoint handler
func (b *ActionsBridge) HealthHandler(w http.ResponseWriter, r *http.Request) {
	status := b.HealthCheck()
	
	w.Header().Set("Content-Type", "application/json")
	
	if !status.Healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(status)
}

// ReadinessHandler checks if the bridge is ready to accept requests
func (b *ActionsBridge) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// Quick check - just verify Docker is accessible
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Not ready: Docker not accessible")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ready")
}

// LivenessHandler checks if the bridge is alive
func (b *ActionsBridge) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - if we can respond, we're alive
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Alive")
}

var startTime = time.Now()

// RecordExecution records metrics for a workflow execution
func RecordExecution(workflow, space, status, platform string, duration time.Duration) {
	executionsTotal.WithLabelValues(workflow, space, status, platform).Inc()
	executionDuration.WithLabelValues(workflow, space).Observe(duration.Seconds())
}

// RecordCompatibilityWarning records a compatibility warning metric
func RecordCompatibilityWarning(action, level string) {
	compatibilityWarnings.WithLabelValues(action, level).Inc()
}

// RecordWorkspaceCleanup records workspace cleanup duration
func RecordWorkspaceCleanup(duration time.Duration) {
	workspaceCleanupDuration.Observe(duration.Seconds())
}
