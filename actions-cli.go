package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/confighub/actions-bridge/pkg/bridge"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	space       string
	unit        string
	dryRun      bool
	verbose     bool
	configFile  string
	
	// Run command flags
	event       string
	inputs      []string
	platform    string
	artifactDir string
	envFile     string
	secretsFile string
	validateOnly bool
	
	// Root command
	rootCmd = &cobra.Command{
		Use:   "cub-actions",
		Short: "ConfigHub GitHub Actions Bridge CLI",
		Long:  "Run GitHub Actions workflows locally with ConfigHub configurations",
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&space, "space", "", "ConfigHub space")
	rootCmd.PersistentFlags().StringVar(&unit, "unit", "", "ConfigHub unit")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without running")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Configuration file")
	
	// Add commands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(compatCmd)
	rootCmd.AddCommand(listCmd)
}

// Run command - execute a workflow
var runCmd = &cobra.Command{
	Use:   "run WORKFLOW",
	Short: "Run a GitHub Actions workflow locally",
	Long: `Run a GitHub Actions workflow locally with ConfigHub configurations.
	
Examples:
  # Run a workflow with staging configs
  cub-actions run deploy.yml --space staging --unit webapp
  
  # Dry run to see what would happen
  cub-actions run deploy.yml --space prod --unit webapp --dry-run
  
  # Run with custom inputs
  cub-actions run deploy.yml --space dev --input version=1.2.3 --input environment=test`,
	Args: cobra.ExactArgs(1),
	RunE: runWorkflow,
}

func init() {
	runCmd.Flags().StringVar(&event, "event", "workflow_dispatch", "GitHub event type")
	runCmd.Flags().StringSliceVarP(&inputs, "input", "i", nil, "Workflow inputs (key=value)")
	runCmd.Flags().StringVar(&platform, "platform", "linux/amd64", "Execution platform")
	runCmd.Flags().StringVar(&artifactDir, "artifact-dir", "./artifacts", "Directory for artifacts")
	runCmd.Flags().StringVar(&envFile, "env-file", "", "Additional environment file")
	runCmd.Flags().StringVar(&secretsFile, "secrets-file", "", "Secrets file (KEY=value format)")
	runCmd.Flags().BoolVar(&validateOnly, "validate", false, "Validate workflow without running")
}

func runWorkflow(cmd *cobra.Command, args []string) error {
	workflowPath := args[0]
	
	// Read workflow file
	workflowData, err := os.ReadFile(workflowPath)
	if err != nil {
		return fmt.Errorf("read workflow: %w", err)
	}
	
	// Create bridge instance
	baseDir := "/tmp/actions-cli"
	actionsBridge, err := bridge.NewActionsBridge(baseDir)
	if err != nil {
		return fmt.Errorf("create bridge: %w", err)
	}
	
	// Validate workflow
	checker := bridge.NewCompatibilityChecker()
	if err := checker.ValidateWorkflowFile(workflowData); err != nil {
		return fmt.Errorf("invalid workflow: %w", err)
	}
	
	// Check compatibility
	warnings := checker.CheckWorkflow(workflowData)
	if len(warnings) > 0 {
		fmt.Println("Compatibility warnings:")
		for _, w := range warnings {
			fmt.Printf("  [%s] Line %d: %s\n", strings.ToUpper(w.Level), w.Line, w.Message)
		}
		
		// Show suggestions
		suggestions := checker.SuggestFixes(warnings)
		if len(suggestions) > 0 {
			fmt.Println("\nSuggestions:")
			for _, s := range suggestions {
				fmt.Printf("  • %s\n", s)
			}
		}
	}
	
	// Stop here if validate only
	if validateOnly {
		fmt.Println("\n✓ Workflow validation passed")
		return nil
	}
	
	// Check for errors in warnings
	hasErrors := false
	for _, w := range warnings {
		if w.Level == "error" {
			hasErrors = true
		}
	}
	
	if hasErrors && !dryRun {
		return fmt.Errorf("workflow has compatibility errors, use --dry-run to see what would happen")
	}
	
	// Prepare execution context
	if space == "" || unit == "" {
		return fmt.Errorf("--space and --unit are required")
	}
	
	// Parse inputs
	inputMap := make(map[string]string)
	for _, input := range inputs {
		parts := strings.SplitN(input, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid input format: %s (expected key=value)", input)
		}
		inputMap[parts[0]] = parts[1]
	}
	
	// Load secrets
	secrets := make(map[string]string)
	if secretsFile != "" {
		secrets, err = loadSecretsFile(secretsFile)
		if err != nil {
			return fmt.Errorf("load secrets: %w", err)
		}
	}
	
	// Dry run output
	if dryRun {
		fmt.Println("\n=== DRY RUN MODE ===")
		fmt.Printf("Would execute workflow: %s\n", workflowPath)
		fmt.Printf("Space: %s\n", space)
		fmt.Printf("Unit: %s\n", unit)
		fmt.Printf("Event: %s\n", event)
		fmt.Printf("Platform: %s\n", platform)
		
		if len(inputMap) > 0 {
			fmt.Println("Inputs:")
			for k, v := range inputMap {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
		
		if len(secrets) > 0 {
			fmt.Printf("Secrets: %d loaded\n", len(secrets))
		}
		
		fmt.Println("\nNo changes made - dry run complete")
		return nil
	}
	
	// Execute workflow
	fmt.Printf("Executing workflow %s...\n", workflowPath)
	
	// Create workspace
	ws, err := actionsBridge.workspaceManager.CreateWorkspace(fmt.Sprintf("cli-%d", time.Now().Unix()))
	if err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}
	defer ws.SecureCleanup()
	
	// Write workflow
	if err := ws.WriteWorkflow("workflow.yml", workflowData); err != nil {
		return fmt.Errorf("write workflow: %w", err)
	}
	
	// Prepare execution context
	ctx := &bridge.ExecutionContext{
		Workspace:  ws,
		ConfigData: workflowData,
		Metadata: &bridge.WorkflowMetadata{
			Configs: map[string]interface{}{
				"inputs": inputMap,
			},
			Secrets: secrets,
		},
		Space:    space,
		Unit:     unit,
		Revision: 1,
		Event:    event,
		Platform: platform,
		Actor:    "cli-user",
		Repository: fmt.Sprintf("confighub/%s/%s", space, unit),
		Ref:       "refs/heads/main",
		SHA:       "0000000",
	}
	
	// Prepare secrets
	secretsPath := ""
	if len(secrets) > 0 {
		secretsPath, err = actionsBridge.prepareSecrets(ws, secrets)
		if err != nil {
			return fmt.Errorf("prepare secrets: %w", err)
		}
	}
	
	// Execute
	result, err := actionsBridge.actRunner.Execute(ctx, secretsPath)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}
	
	// Show results
	fmt.Printf("\n✓ Workflow completed in %v\n", result.Duration)
	fmt.Printf("Exit code: %d\n", result.ExitCode)
	
	if len(result.Artifacts) > 0 {
		fmt.Printf("Artifacts:\n")
		for _, a := range result.Artifacts {
			fmt.Printf("  - %s\n", a)
		}
	}
	
	return nil
}

// Validate command - validate workflows
var validateCmd = &cobra.Command{
	Use:   "validate WORKFLOW",
	Short: "Validate a GitHub Actions workflow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowPath := args[0]
		
		// Read workflow
		workflowData, err := os.ReadFile(workflowPath)
		if err != nil {
			return fmt.Errorf("read workflow: %w", err)
		}
		
		// Validate
		checker := bridge.NewCompatibilityChecker()
		if err := checker.ValidateWorkflowFile(workflowData); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		
		// Check compatibility
		warnings := checker.CheckWorkflow(workflowData)
		
		if len(warnings) == 0 {
			fmt.Printf("✓ Workflow %s is valid and compatible\n", workflowPath)
			return nil
		}
		
		// Show warnings
		errorCount := 0
		warningCount := 0
		infoCount := 0
		
		for _, w := range warnings {
			switch w.Level {
			case "error":
				errorCount++
			case "warning":
				warningCount++
			case "info":
				infoCount++
			}
		}
		
		fmt.Printf("Validation results for %s:\n", workflowPath)
		fmt.Printf("  Errors: %d\n", errorCount)
		fmt.Printf("  Warnings: %d\n", warningCount)
		fmt.Printf("  Info: %d\n", infoCount)
		
		// Show details
		fmt.Println("\nDetails:")
		for _, w := range warnings {
			icon := "ℹ"
			if w.Level == "warning" {
				icon = "⚠"
			} else if w.Level == "error" {
				icon = "✗"
			}
			fmt.Printf("  %s [%s] Line %d: %s\n", icon, strings.ToUpper(w.Level), w.Line, w.Message)
		}
		
		// Show suggestions
		suggestions := checker.SuggestFixes(warnings)
		if len(suggestions) > 0 {
			fmt.Println("\nSuggestions:")
			for _, s := range suggestions {
				fmt.Printf("  • %s\n", s)
			}
		}
		
		if errorCount > 0 {
			return fmt.Errorf("workflow has %d errors", errorCount)
		}
		
		return nil
	},
}

// Compat command - check compatibility
var compatCmd = &cobra.Command{
	Use:   "compat",
	Short: "Show act compatibility information",
	RunE: func(cmd *cobra.Command, args []string) error {
		checker := bridge.NewCompatibilityChecker()
		limitations := checker.KnownLimitations()
		
		fmt.Println("Known act limitations:")
		fmt.Println()
		
		for _, limitation := range limitations {
			fmt.Printf("• %s\n", limitation)
		}
		
		fmt.Println()
		fmt.Println("For more information, see: https://github.com/nektos/act#known-issues")
		
		return nil
	},
}

// List command - list available workflows
var listCmd = &cobra.Command{
	Use:   "list [DIRECTORY]",
	Short: "List available workflows",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}
		
		// Find workflow files
		workflows := []string{}
		
		// Check .github/workflows
		workflowDir := filepath.Join(dir, ".github", "workflows")
		if entries, err := os.ReadDir(workflowDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yml") || strings.HasSuffix(entry.Name(), ".yaml") {
					workflows = append(workflows, filepath.Join(".github", "workflows", entry.Name()))
				}
			}
		}
		
		// Check root directory
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				name := entry.Name()
				if !entry.IsDir() && (strings.HasSuffix(name, ".workflow.yml") || strings.HasSuffix(name, ".workflow.yaml")) {
					workflows = append(workflows, name)
				}
			}
		}
		
		if len(workflows) == 0 {
			fmt.Println("No workflows found")
			return nil
		}
		
		fmt.Printf("Found %d workflow(s):\n\n", len(workflows))
		fmt.Printf("%-40s %-20s %s\n", "WORKFLOW", "TRIGGERS", "LAST MODIFIED")
		fmt.Println(strings.Repeat("-", 80))
		
		for _, workflow := range workflows {
			// Read workflow to get triggers
			data, err := os.ReadFile(filepath.Join(dir, workflow))
			if err != nil {
				continue
			}
			
			// Parse triggers (simplified)
			triggers := "unknown"
			if strings.Contains(string(data), "on:") {
				// Extract triggers (very simplified parsing)
				lines := strings.Split(string(data), "\n")
				for i, line := range lines {
					if strings.TrimSpace(line) == "on:" && i+1 < len(lines) {
						triggers = strings.TrimSpace(lines[i+1])
						triggers = strings.TrimPrefix(triggers, "- ")
						break
					}
				}
			}
			
			// Get file info
			info, _ := os.Stat(filepath.Join(dir, workflow))
			modified := "unknown"
			if info != nil {
				modified = info.ModTime().Format("2006-01-02 15:04")
			}
			
			fmt.Printf("%-40s %-20s %s\n", workflow, triggers, modified)
		}
		
		return nil
	},
}

// Helper functions

func loadSecretsFile(path string) (map[string]string, error) {
	secrets := make(map[string]string)
	
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
			secrets[parts[0]] = parts[1]
		}
	}
	
	return secrets, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
