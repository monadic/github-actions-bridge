package bridge

import (
	"fmt"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"
)

// CompatibilityChecker validates workflows for act compatibility
type CompatibilityChecker struct {
	unsupportedActions map[string]string
	warningPatterns    []WarningPattern
}

// Warning represents a compatibility warning
type Warning struct {
	Level   string // "info", "warning", "error"
	Action  string
	Message string
	Line    int
}

// WarningPattern defines patterns to check for compatibility issues
type WarningPattern struct {
	Pattern *regexp.Regexp
	Level   string
	Message string
}

// NewCompatibilityChecker creates a new compatibility checker
func NewCompatibilityChecker() *CompatibilityChecker {
	return &CompatibilityChecker{
		unsupportedActions: map[string]string{
			"actions/cache@":              "Caching not supported locally - workflow will run without cache",
			"actions/upload-artifact@":    "Artifacts saved to workspace output directory only",
			"actions/download-artifact@":  "Cross-workflow artifacts not supported locally",
			"docker/build-push-action@":   "Registry push disabled in local execution",
			"actions/create-release@":     "Release creation skipped in local execution",
			"actions/upload-release-asset@": "Release asset upload skipped in local execution",
			"azure/":                      "Azure-specific actions may not work locally",
			"aws-actions/":                "AWS-specific actions require local credentials",
			"google-github-actions/":      "Google Cloud actions require local credentials",
		},
		warningPatterns: []WarningPattern{
			{
				Pattern: regexp.MustCompile(`\$\{\{\s*secrets\.GITHUB_TOKEN\s*\}\}`),
				Level:   "warning",
				Message: "GITHUB_TOKEN will be simulated locally with limited permissions",
			},
			{
				Pattern: regexp.MustCompile(`\$\{\{\s*github\.event\.pull_request\.\w+\s*\}\}`),
				Level:   "info",
				Message: "Pull request context will be simulated in local execution",
			},
			{
				Pattern: regexp.MustCompile(`\$\{\{\s*runner\.os\s*==\s*'Windows'\s*\}\}`),
				Level:   "warning",
				Message: "Windows runners not supported - will use Linux",
			},
			{
				Pattern: regexp.MustCompile(`runs-on:\s*\[\s*self-hosted`),
				Level:   "error",
				Message: "Self-hosted runners not supported in local execution",
			},
			{
				Pattern: regexp.MustCompile(`\$\{\{\s*github\.event\.deployment\.\w+\s*\}\}`),
				Level:   "warning", 
				Message: "Deployment events will be simulated locally",
			},
		},
	}
}

// CheckWorkflow analyzes a workflow for compatibility issues
func (cc *CompatibilityChecker) CheckWorkflow(workflowData []byte) []Warning {
	warnings := []Warning{}
	content := string(workflowData)
	lines := strings.Split(content, "\n")

	// Parse workflow to check structure
	var workflow map[string]interface{}
	if err := yaml.Unmarshal(workflowData, &workflow); err == nil {
		warnings = append(warnings, cc.checkWorkflowStructure(workflow)...)
	}

	// Check for unsupported actions
	for action, message := range cc.unsupportedActions {
		if idx := strings.Index(content, action); idx >= 0 {
			line := cc.getLineNumber(content, idx)
			warnings = append(warnings, Warning{
				Level:   "info",
				Message: message,
				Action:  action,
				Line:    line,
			})
		}
	}

	// Check patterns
	for _, pattern := range cc.warningPatterns {
		matches := pattern.Pattern.FindAllStringIndex(content, -1)
		for _, match := range matches {
			line := cc.getLineNumber(content, match[0])
			warnings = append(warnings, Warning{
				Level:   pattern.Level,
				Message: pattern.Message,
				Line:    line,
			})
		}
	}

	// Check for specific GitHub features
	warnings = append(warnings, cc.checkGitHubSpecificFeatures(content, lines)...)

	return warnings
}

// checkWorkflowStructure validates the workflow structure
func (cc *CompatibilityChecker) checkWorkflowStructure(workflow map[string]interface{}) []Warning {
	warnings := []Warning{}

	// Check jobs
	if jobs, ok := workflow["jobs"].(map[string]interface{}); ok {
		for jobName, jobData := range jobs {
			if job, ok := jobData.(map[string]interface{}); ok {
				// Check runs-on
				if runsOn, ok := job["runs-on"].(string); ok {
					if strings.Contains(runsOn, "self-hosted") {
						warnings = append(warnings, Warning{
							Level:   "error",
							Message: fmt.Sprintf("Job '%s' uses self-hosted runner which is not supported", jobName),
						})
					}
					if strings.Contains(runsOn, "windows") || strings.Contains(runsOn, "macos") {
						warnings = append(warnings, Warning{
							Level:   "warning",
							Message: fmt.Sprintf("Job '%s' uses %s runner - will run on Linux instead", jobName, runsOn),
						})
					}
				}

				// Check services
				if services, ok := job["services"].(map[string]interface{}); ok && len(services) > 0 {
					warnings = append(warnings, Warning{
						Level:   "info",
						Message: fmt.Sprintf("Job '%s' uses services - ensure Docker is available", jobName),
					})
				}

				// Check strategy matrix
				if strategy, ok := job["strategy"].(map[string]interface{}); ok {
					if matrix, ok := strategy["matrix"].(map[string]interface{}); ok && len(matrix) > 0 {
						warnings = append(warnings, Warning{
							Level:   "info",
							Message: fmt.Sprintf("Job '%s' uses matrix strategy - all combinations will run sequentially", jobName),
						})
					}
				}
			}
		}
	}

	return warnings
}

// checkGitHubSpecificFeatures checks for GitHub-specific features
func (cc *CompatibilityChecker) checkGitHubSpecificFeatures(content string, lines []string) []Warning {
	warnings := []Warning{}

	// Check for GitHub environments
	if strings.Contains(content, "environment:") {
		warnings = append(warnings, Warning{
			Level:   "info",
			Message: "GitHub environments not supported - environment protection rules will be skipped",
		})
	}

	// Check for concurrency groups
	if strings.Contains(content, "concurrency:") {
		warnings = append(warnings, Warning{
			Level:   "info",
			Message: "Concurrency groups not enforced in local execution",
		})
	}

	// Check for specific webhook events
	unsupportedEvents := []string{
		"pull_request_review",
		"pull_request_review_comment", 
		"check_suite",
		"check_run",
		"deployment",
		"deployment_status",
		"page_build",
		"project_card",
	}

	for _, event := range unsupportedEvents {
		if strings.Contains(content, event+":") {
			warnings = append(warnings, Warning{
				Level:   "warning",
				Message: fmt.Sprintf("Event '%s' cannot be fully simulated locally", event),
			})
		}
	}

	return warnings
}

// KnownLimitations returns a list of known act limitations
func (cc *CompatibilityChecker) KnownLimitations() []string {
	return []string{
		"No support for GitHub-hosted runner hardware specs",
		"Limited GitHub API access (no real GITHUB_TOKEN)",
		"No support for GitHub Packages or Container Registry",
		"Artifacts are local only (no cross-workflow sharing)",
		"No support for GitHub Environments and protection rules",
		"Limited support for complex webhook event payloads",
		"Services run in Docker containers (no cloud services)",
		"No support for GitHub Advanced Security features",
		"Caching is local only (not shared between runs)",
		"Windows and macOS runners run as Linux containers",
	}
}

// ValidateWorkflowFile performs basic YAML validation
func (cc *CompatibilityChecker) ValidateWorkflowFile(workflowData []byte) error {
	var workflow map[string]interface{}
	if err := yaml.Unmarshal(workflowData, &workflow); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Check required fields
	if _, ok := workflow["name"]; !ok {
		return fmt.Errorf("workflow missing 'name' field")
	}

	if _, ok := workflow["on"]; !ok {
		return fmt.Errorf("workflow missing 'on' field")
	}

	if jobs, ok := workflow["jobs"].(map[string]interface{}); !ok || len(jobs) == 0 {
		return fmt.Errorf("workflow must have at least one job")
	}

	return nil
}

// getLineNumber finds the line number for a given position in the content
func (cc *CompatibilityChecker) getLineNumber(content string, position int) int {
	line := 1
	for i := 0; i < position && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}
	return line
}

// SuggestFixes provides suggestions for common compatibility issues
func (cc *CompatibilityChecker) SuggestFixes(warnings []Warning) []string {
	suggestions := []string{}
	
	hasCache := false
	hasArtifacts := false
	hasSelfHosted := false
	
	for _, w := range warnings {
		if strings.Contains(w.Action, "cache@") && !hasCache {
			hasCache = true
			suggestions = append(suggestions, "Consider removing cache actions for local testing or using local directory mounting")
		}
		if strings.Contains(w.Action, "artifact@") && !hasArtifacts {
			hasArtifacts = true
			suggestions = append(suggestions, "Artifacts will be saved to the workspace output directory")
		}
		if w.Level == "error" && strings.Contains(w.Message, "self-hosted") && !hasSelfHosted {
			hasSelfHosted = true
			suggestions = append(suggestions, "Replace 'self-hosted' runners with 'ubuntu-latest' for local execution")
		}
	}
	
	return suggestions
}
