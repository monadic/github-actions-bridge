package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/confighub/actions-bridge/pkg/bridge"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	// Version is set at build time
	Version = "dev"

	// Global flag
	verbose bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cub-local-actions",
		Short: "Worker for GitHub Actions Bridge",
		Long: `A ConfigHub worker that runs GitHub Actions workflows locally 
using the Actions Bridge. This worker integrates with ConfigHub to execute 
workflows based on configuration units.`,
		Version: Version,
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add commands
	rootCmd.AddCommand(
		runCommand(),
		validateCommand(),
		listCommand(),
		cleanCommand(),
		versionCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runCommand creates the run command
func runCommand() *cobra.Command {
	var (
		space        string
		unit         string
		dryRun       bool
		event        string
		inputs       []string
		platform     string
		artifactDir  string
		envFile      string
		secretsFile  string
		validateOnly bool
		watch        bool
		timeout      int
	)

	cmd := &cobra.Command{
		Use:   "run WORKFLOW",
		Short: "Run a GitHub Actions workflow locally",
		Long: `Run a GitHub Actions workflow using act. This command provides
full control over the execution environment and allows testing workflows
before deploying them.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowPath := args[0]

			// Validate workflow exists
			if _, err := os.Stat(workflowPath); err != nil {
				return fmt.Errorf("workflow file not found: %s", workflowPath)
			}

			// Read workflow
			workflowData, err := os.ReadFile(workflowPath)
			if err != nil {
				return fmt.Errorf("read workflow: %w", err)
			}

			// Strip ConfigHub metadata if present
			workflowData = stripConfigHubMetadata(workflowData)

			// Create temporary workspace
			tempDir, err := os.MkdirTemp("", "actions-cli-*")
			if err != nil {
				return fmt.Errorf("create temp dir: %w", err)
			}
			defer os.RemoveAll(tempDir)

			// Initialize bridge components
			workspaceManager, err := bridge.NewWorkspaceManager(tempDir)
			if err != nil {
				return fmt.Errorf("create workspace manager: %w", err)
			}

			ws, err := workspaceManager.CreateWorkspace(uuid.New().String())
			if err != nil {
				return fmt.Errorf("create workspace: %w", err)
			}
			defer ws.SecureCleanup()

			// Check compatibility
			checker := bridge.NewCompatibilityChecker()
			warnings := checker.CheckWorkflow(workflowData)

			if len(warnings) > 0 {
				fmt.Println("Compatibility warnings:")
				for _, w := range warnings {
					fmt.Printf("  [%s] Line %d: %s\n", w.Level, w.Line, w.Message)
				}
				fmt.Println()
			}

			// Validation mode
			if validateOnly {
				supported, reason := checker.IsWorkflowSupported(workflowData)
				if !supported {
					return fmt.Errorf("workflow not supported: %s", reason)
				}
				fmt.Println("✓ Workflow is valid and supported")
				return nil
			}

			// Write workflow to workspace
			// Act expects the workflow to be named "workflow.yml"
			if err := ws.WriteWorkflow("workflow.yml", workflowData); err != nil {
				return fmt.Errorf("write workflow: %w", err)
			}

			// Parse inputs
			inputMap := make(map[string]interface{})
			for _, input := range inputs {
				parts := strings.SplitN(input, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid input format: %s (expected key=value)", input)
				}
				inputMap[parts[0]] = parts[1]
			}

			// Load secrets if provided
			secrets := make(map[string]string)
			if secretsFile != "" {
				secrets, err = bridge.ParseSecretsFile(secretsFile)
				if err != nil {
					return fmt.Errorf("parse secrets: %w", err)
				}
			}

			// Load environment if provided
			environment := make(map[string]string)
			if envFile != "" {
				environment, err = parseEnvFile(envFile)
				if err != nil {
					return fmt.Errorf("parse env file: %w", err)
				}
			}

			// Prepare execution context
			execCtx := &bridge.ExecutionContext{
				Workspace:  ws,
				ConfigData: workflowData,
				Metadata: bridge.ExecutionMetadata{
					Space:    space,
					Unit:     unit,
					Revision: 1,
					Actor:    os.Getenv("USER"),
				},
				Secrets:     secrets,
				Environment: environment,
				EventPayload: map[string]interface{}{
					"action": event,
					"inputs": inputMap,
				},
				DryRun: dryRun,
			}
			
			// Create runner with default container image
			containerImage := "catthehacker/ubuntu:act-latest"
			runner := bridge.NewActRunner(platform, containerImage)
			
			// Execute workflow
			fmt.Printf("Running workflow: %s\n", workflowPath)
			if dryRun {
				fmt.Println("DRY RUN - No actual execution")
			}

			result, err := runner.Execute(execCtx)
			if err != nil {
				return fmt.Errorf("execution failed: %w", err)
			}

			// Display results
			fmt.Printf("\nExecution completed in %s\n", result.Duration)
			fmt.Printf("Exit code: %d\n", result.ExitCode)

			if len(result.Artifacts) > 0 {
				fmt.Printf("\nArtifacts:\n")
				for _, artifact := range result.Artifacts {
					fmt.Printf("  - %s\n", artifact)
				}

				// Copy artifacts if directory specified
				if artifactDir != "" {
					if err := copyArtifacts(ws.OutputDir, artifactDir); err != nil {
						log.Printf("Failed to copy artifacts: %v", err)
					} else {
						fmt.Printf("\nArtifacts copied to: %s\n", artifactDir)
					}
				}
			}

			// Show logs if verbose
			if verbose && len(result.Logs) > 0 {
				fmt.Printf("\nExecution logs:\n")
				for _, log := range result.Logs {
					fmt.Println(log)
				}
			}

			if result.ExitCode != 0 {
				return fmt.Errorf("workflow failed with exit code %d", result.ExitCode)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&space, "space", "", "ConfigHub space")
	cmd.Flags().StringVar(&unit, "unit", "", "ConfigHub unit")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without running")
	cmd.Flags().StringVar(&event, "event", "workflow_dispatch", "GitHub event type to simulate")
	cmd.Flags().StringSliceVarP(&inputs, "input", "i", nil, "Workflow inputs (key=value)")
	cmd.Flags().StringVar(&platform, "platform", "linux/amd64", "Execution platform")
	cmd.Flags().StringVar(&artifactDir, "artifact-dir", "", "Directory to save artifacts")
	cmd.Flags().StringVar(&envFile, "env-file", "", "Environment file to load")
	cmd.Flags().StringVar(&secretsFile, "secrets-file", "", "Secrets file to load")
	cmd.Flags().BoolVar(&validateOnly, "validate", false, "Validate workflow without running")
	cmd.Flags().BoolVar(&watch, "watch", false, "Watch workflow file for changes")
	cmd.Flags().IntVar(&timeout, "timeout", 3600, "Execution timeout in seconds")

	return cmd
}

// validateCommand creates the validate command
func validateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate WORKFLOW",
		Short: "Validate a GitHub Actions workflow",
		Long:  "Check if a workflow is valid and can be executed locally with act.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowPath := args[0]

			// Read workflow
			workflowData, err := os.ReadFile(workflowPath)
			if err != nil {
				return fmt.Errorf("read workflow: %w", err)
			}

			// Parse as YAML to check syntax
			var workflow map[string]interface{}
			if err := yaml.Unmarshal(workflowData, &workflow); err != nil {
				return fmt.Errorf("invalid YAML syntax: %w", err)
			}

			// Check with compatibility checker
			checker := bridge.NewCompatibilityChecker()
			warnings := checker.CheckWorkflow(workflowData)

			supported, reason := checker.IsWorkflowSupported(workflowData)
			if !supported {
				return fmt.Errorf("workflow not supported: %s", reason)
			}

			fmt.Printf("✓ Workflow is valid: %s\n", workflowPath)

			if len(warnings) > 0 {
				fmt.Printf("\nCompatibility notes:\n")
				for _, w := range warnings {
					fmt.Printf("  [%s] Line %d: %s\n", w.Level, w.Line, w.Message)
				}

				// Show suggestions
				suggestions := checker.SuggestFixes(warnings)
				if len(suggestions) > 0 {
					fmt.Printf("\nSuggestions:\n")
					for _, s := range suggestions {
						fmt.Printf("  - %s\n", s)
					}
				}
			}

			return nil
		},
	}
}

// listCommand lists known limitations
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-limitations",
		Short: "List known act limitations",
		Long:  "Display all known limitations when running GitHub Actions locally.",
		Run: func(cmd *cobra.Command, args []string) {
			checker := bridge.NewCompatibilityChecker()
			limitations := checker.KnownLimitations()

			fmt.Println("Known limitations when running GitHub Actions locally:")
			fmt.Println()
			for i, limitation := range limitations {
				fmt.Printf("%2d. %s\n", i+1, limitation)
			}
			fmt.Println()
			fmt.Println("For more information, see: https://github.com/nektos/act#known-issues")
		},
	}
}

// cleanCommand creates the clean command
func cleanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Clean up temporary files and Docker resources",
		Long:  "Remove temporary workspaces and optionally clean Docker containers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Clean temporary directories
			tempDir := os.TempDir()
			pattern := filepath.Join(tempDir, "actions-cli-*")
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return err
			}

			removed := 0
			for _, match := range matches {
				if err := os.RemoveAll(match); err != nil {
					log.Printf("Failed to remove %s: %v", match, err)
				} else {
					removed++
				}
			}

			fmt.Printf("Cleaned up %d temporary directories\n", removed)

			// TODO: Add Docker cleanup
			fmt.Println("\nTo clean Docker resources, run:")
			fmt.Println("  docker system prune -f")

			return nil
		},
	}
}

// versionCommand shows version information
func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cub-local-actions version %s\n", Version)
			fmt.Println("GitHub Actions Bridge Worker")
			fmt.Println("https://github.com/confighub/actions-bridge")
		},
	}
}

// Helper functions

func parseEnvFile(path string) (map[string]string, error) {
	env := make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			env[key] = value
		}
	}

	return env, nil
}

func copyArtifacts(srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dst := filepath.Join(dstDir, rel)

		if info.IsDir() {
			return os.MkdirAll(dst, info.Mode())
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dstFile, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, src)
		return err
	})
}

// stripConfigHubMetadata removes the first 4 lines if they contain ConfigHub metadata
func stripConfigHubMetadata(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	
	// Check if the first line contains apiVersion: actions.confighub.com
	if len(lines) > 0 && bytes.Contains(lines[0], []byte("apiVersion:")) && bytes.Contains(lines[0], []byte("actions.confighub.com")) {
		// If we have more than 4 lines, skip the first 4
		if len(lines) > 4 {
			return bytes.Join(lines[4:], []byte("\n"))
		}
	}
	
	// Otherwise, return the original data
	return data
}
