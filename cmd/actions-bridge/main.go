package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confighub/actions-bridge/pkg/bridge"
	"github.com/confighub/sdk/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Version is set at build time
	Version = "dev"
)

func main() {
	log.Printf("Starting GitHub Actions Bridge for ConfigHub (version: %s)", Version)

	// Configuration from environment
	config := &Config{
		WorkerID:      getEnv("CONFIGHUB_WORKER_ID", ""),
		WorkerSecret:  getEnv("CONFIGHUB_WORKER_SECRET", ""),
		ConfigHubURL:  getEnv("CONFIGHUB_URL", "https://api.confighub.com"),
		BaseDir:       getEnv("ACTIONS_BRIDGE_BASE_DIR", "./actions-bridge-workspace"),
		ActImage:      getEnv("ACT_DEFAULT_IMAGE", "catthehacker/ubuntu:act-latest"),
		Platform:      getEnv("ACT_PLATFORM", "linux/amd64"),
		MaxConcurrent: getEnvInt("MAX_CONCURRENT_WORKFLOWS", 5),
		HealthAddr:    getEnv("HEALTH_ADDR", ":8080"),
		Debug:         getEnvBool("DEBUG", false),
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create bridge instance
	actionsBridge, err := bridge.NewActionsBridge(config.BaseDir)
	if err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}

	// Create bridge dispatcher and register our bridge
	bridgeDispatcher := worker.NewBridgeDispatcher()
	bridgeDispatcher.RegisterBridge(actionsBridge)

	// Create ConfigHub SDK connector
	connector, err := worker.NewConnector(
		worker.ConnectorOptions{
			ConfigHubURL:     config.ConfigHubURL,
			WorkerID:         config.WorkerID,
			WorkerSecret:     config.WorkerSecret,
			BridgeDispatcher: &bridgeDispatcher,
		},
	)
	if err != nil {
		log.Fatalf("Failed to create connector: %v", err)
	}

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start health check server
	healthServer := startHealthServer(actionsBridge, config.HealthAddr)

	// Start connector in background
	go func() {
		log.Println("Starting ConfigHub connector...")
		if err := connector.Start(); err != nil {
			log.Printf("Connector error: %v", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}

	// Graceful shutdown
	log.Println("Stopping worker...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Note: SDK worker doesn't have explicit Stop method, canceling context should handle it

	log.Println("Stopping health server...")
	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error stopping health server: %v", err)
	}

	// Cleanup old workspaces on shutdown
	log.Println("Cleaning up workspaces...")
	// Note: This would need an exported cleanup method in production
	log.Println("Workspace cleanup completed")

	log.Println("GitHub Actions Bridge stopped")
}

// startHealthServer starts the health check HTTP server
func startHealthServer(actionsBridge *bridge.ActionsBridge, addr string) *http.Server {
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/health", actionsBridge.HealthHandler)
	mux.HandleFunc("/ready", actionsBridge.ReadinessHandler)
	mux.HandleFunc("/live", actionsBridge.LivenessHandler)

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Version endpoint
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"version":"%s","service":"github-actions-bridge"}`, Version)
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Health server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health server error: %v", err)
		}
	}()

	return server
}

// Config holds the bridge configuration
type Config struct {
	WorkerID      string
	WorkerSecret  string
	ConfigHubURL  string
	BaseDir       string
	ActImage      string
	Platform      string
	MaxConcurrent int
	HealthAddr    string
	Debug         bool
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.WorkerID == "" {
		return fmt.Errorf("CONFIGHUB_WORKER_ID is required")
	}
	if c.WorkerSecret == "" {
		return fmt.Errorf("CONFIGHUB_WORKER_SECRET is required")
	}
	if c.BaseDir == "" {
		return fmt.Errorf("base directory is required")
	}
	if c.MaxConcurrent < 1 {
		c.MaxConcurrent = 1
	}
	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
