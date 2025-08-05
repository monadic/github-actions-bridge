// Act test runner - validates nektos/act integration
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nektos/act/pkg/model"
	"github.com/nektos/act/pkg/runner"
)

func main() {
	log.Println("Act Test Runner - Validating nektos/act integration")
	
	// Create temporary directory
	baseDir, err := os.MkdirTemp("", "act-test-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(baseDir)
	
	log.Printf("Working directory: %s", baseDir)
	
	// Run tests
	tests := []struct {
		name string
		fn   func(string) error
	}{
		{"Basic workflow execution", testBasicWorkflow},
		{"Secret file injection", testSecretInjection},
		{"Platform detection", testPlatformDetection},
		{"Artifact generation", testArtifactGeneration},
	}
	
	failures := 0
	for _, test := range tests {
		log.Printf("\n=== Running: %s ===", test.name)
		if err := test.fn(baseDir); err != nil {
			log.Printf("❌ FAILED: %v", err)
			failures++
		} else {
			log.Printf("✅ PASSED")
		}
	}
	
	if failures > 0 {
		log.Fatalf("\n%d test(s) failed", failures)
	}
	
	log.Println("\nAll tests passed! ✅")
}

func testBasicWorkflow(baseDir string) error {
	log.Println("Test 1: Basic workflow execution")
	
	// Create workflow
	workflowDir := filepath.Join(baseDir, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("create workflow dir: %w", err)
	}
	
	workflow := `name: Test Workflow
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Echo test
        run: echo "Hello from act!"
      - name: Multi-line command
        run: |
          echo "Line 1"
          echo "Line 2"
`
	
	workflowPath := filepath.Join(workflowDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflow), 0644); err != nil {
		return fmt.Errorf("write workflow: %w", err)
	}
	
	// Create event
	event := map[string]interface{}{
		"repository": map[string]interface{}{
			"name": "test-repo",
		},
		"pusher": map[string]interface{}{
			"name": "test-user",
		},
		"ref": "refs/heads/main",
	}
	
	eventData, _ := json.Marshal(event)
	eventPath := filepath.Join(baseDir, "event.json")
	if err := os.WriteFile(eventPath, eventData, 0644); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	
	// Run workflow
	config := &runner.Config{
		EventPath:     eventPath,
		EventName:     "push",
		Platforms: map[string]string{
			"ubuntu-latest": "catthehacker/ubuntu:act-latest",
		},
		LogOutput:      true,
		ReuseContainers: false,
		Workdir:        baseDir,
	}
	
	runner, err := runner.New(config)
	if err != nil {
		return fmt.Errorf("create runner: %w", err)
	}
	
	// Get plan
	planner, err := model.NewWorkflowPlanner(workflowPath, false, false)
	if err != nil {
		return fmt.Errorf("create planner: %w", err)
	}
	
	plan, err := planner.PlanEvent(config.EventName)
	if err != nil {
		return fmt.Errorf("plan event: %w", err)
	}
	
	// Execute
	executor := runner.NewPlanExecutor(plan).Finally(func(_ context.Context) error {
		return nil
	})
	
	ctx := context.Background()
	if err := executor(ctx); err != nil {
		return fmt.Errorf("execute plan: %w", err)
	}
	
	log.Println("✓ Basic workflow execution successful")
	return nil
}

func testSecretInjection(baseDir string) error {
	log.Println("Test 2: Secret file injection")
	
	// Create workflow with secrets
	workflowDir := filepath.Join(baseDir, ".github", "workflows")
	workflow := `name: Secret Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Use secret
        run: |
          if [ -z "${{ secrets.TEST_SECRET }}" ]; then
            echo "Secret not found!"
            exit 1
          fi
          echo "Secret length: ${#TEST_SECRET}"
        env:
          TEST_SECRET: ${{ secrets.TEST_SECRET }}
`
	
	workflowPath := filepath.Join(workflowDir, "secret-test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflow), 0644); err != nil {
		return fmt.Errorf("write workflow: %w", err)
	}
	
	// Create event
	event := map[string]interface{}{
		"repository": map[string]interface{}{
			"name": "test-repo",
		},
	}
	
	eventData, _ := json.Marshal(event)
	eventPath := filepath.Join(baseDir, "event2.json")
	if err := os.WriteFile(eventPath, eventData, 0644); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	
	// Run with secrets
	config := &runner.Config{
		EventPath:     eventPath,
		EventName:     "push",
		Secrets: map[string]string{
			"TEST_SECRET": "supersecretvalue123",
		},
		Platforms: map[string]string{
			"ubuntu-latest": "catthehacker/ubuntu:act-latest",
		},
		LogOutput:      true,
		ReuseContainers: false,
		Workdir:        baseDir,
	}
	
	runner, err := runner.New(config)
	if err != nil {
		return fmt.Errorf("create runner: %w", err)
	}
	
	// Get plan
	planner, err := model.NewWorkflowPlanner(workflowPath, false, false)
	if err != nil {
		return fmt.Errorf("create planner: %w", err)
	}
	
	plan, err := planner.PlanEvent(config.EventName)
	if err != nil {
		return fmt.Errorf("plan event: %w", err)
	}
	
	// Execute
	executor := runner.NewPlanExecutor(plan).Finally(func(_ context.Context) error {
		return nil
	})
	
	ctx := context.Background()
	if err := executor(ctx); err != nil {
		return fmt.Errorf("execute plan: %w", err)
	}
	
	log.Println("✓ Secret injection successful")
	return nil
}

func testPlatformDetection(baseDir string) error {
	log.Println("Test 3: Platform detection")
	
	// Create workflow with different platforms
	workflowDir := filepath.Join(baseDir, ".github", "workflows")
	workflow := `name: Platform Test
on: push
jobs:
  ubuntu-test:
    runs-on: ubuntu-latest
    steps:
      - run: uname -a
  ubuntu-20:
    runs-on: ubuntu-20.04
    steps:
      - run: lsb_release -a
`
	
	workflowPath := filepath.Join(workflowDir, "platform-test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflow), 0644); err != nil {
		return fmt.Errorf("write workflow: %w", err)
	}
	
	// Create event
	eventPath := filepath.Join(baseDir, "event3.json")
	eventData, _ := json.Marshal(map[string]interface{}{})
	if err := os.WriteFile(eventPath, eventData, 0644); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	
	// Run workflow
	config := &runner.Config{
		EventPath: eventPath,
		EventName: "push",
		Platforms: map[string]string{
			"ubuntu-latest": "catthehacker/ubuntu:act-latest",
			"ubuntu-20.04":  "catthehacker/ubuntu:act-20.04",
		},
		LogOutput:      true,
		ReuseContainers: false,
		Workdir:        baseDir,
	}
	
	runner, err := runner.New(config)
	if err != nil {
		return fmt.Errorf("create runner: %w", err)
	}
	
	// Get plan
	planner, err := model.NewWorkflowPlanner(workflowPath, false, false)
	if err != nil {
		return fmt.Errorf("create planner: %w", err)
	}
	
	plan, err := planner.PlanEvent(config.EventName)
	if err != nil {
		return fmt.Errorf("plan event: %w", err)
	}
	
	// Execute
	executor := runner.NewPlanExecutor(plan).Finally(func(_ context.Context) error {
		return nil
	})
	
	ctx := context.Background()
	if err := executor(ctx); err != nil {
		return fmt.Errorf("execute plan: %w", err)
	}
	
	log.Println("✓ Platform detection successful")
	return nil
}

func testArtifactGeneration(baseDir string) error {
	log.Println("Test 4: Artifact generation")
	
	// Create workflow that generates artifacts
	workflowDir := filepath.Join(baseDir, ".github", "workflows")
	workflow := `name: Artifact Test
on: push
jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - name: Create artifact
        run: |
          mkdir -p output
          echo "Test artifact content" > output/artifact.txt
          echo "Another file" > output/file2.txt
      - name: Upload artifact
        uses: actions/upload-artifact@v2
        with:
          name: test-artifacts
          path: output/
`
	
	workflowPath := filepath.Join(workflowDir, "artifact-test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflow), 0644); err != nil {
		return fmt.Errorf("write workflow: %w", err)
	}
	
	// Create event
	eventPath := filepath.Join(baseDir, "event4.json")
	eventData, _ := json.Marshal(map[string]interface{}{})
	if err := os.WriteFile(eventPath, eventData, 0644); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	
	// Create artifact directory
	artifactDir := filepath.Join(baseDir, "artifacts")
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return fmt.Errorf("create artifact dir: %w", err)
	}
	
	// Run workflow
	config := &runner.Config{
		EventPath: eventPath,
		EventName: "push",
		Platforms: map[string]string{
			"ubuntu-latest": "catthehacker/ubuntu:act-latest",
		},
		ArtifactServerPath: artifactDir,
		LogOutput:         true,
		ReuseContainers:   false,
		Workdir:           baseDir,
	}
	
	runner, err := runner.New(config)
	if err != nil {
		return fmt.Errorf("create runner: %w", err)
	}
	
	// Get plan
	planner, err := model.NewWorkflowPlanner(workflowPath, false, false)
	if err != nil {
		return fmt.Errorf("create planner: %w", err)
	}
	
	plan, err := planner.PlanEvent(config.EventName)
	if err != nil {
		return fmt.Errorf("plan event: %w", err)
	}
	
	// Execute
	executor := runner.NewPlanExecutor(plan).Finally(func(_ context.Context) error {
		return nil
	})
	
	ctx := context.Background()
	if err := executor(ctx); err != nil {
		return fmt.Errorf("execute plan: %w", err)
	}
	
	// Check artifacts
	entries, err := os.ReadDir(artifactDir)
	if err != nil {
		return fmt.Errorf("read artifact dir: %w", err)
	}
	
	if len(entries) == 0 {
		return fmt.Errorf("no artifacts generated")
	}
	
	log.Printf("✓ Generated %d artifact(s)", len(entries))
	return nil
}