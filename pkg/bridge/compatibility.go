package bridge

import (
	"fmt"
	"regexp"
	"strings"
)

// Warning represents a compatibility warning
type Warning struct {
	Level   string // "info", "warning", "error"
	Message string
	Action  string
	Line    int
}

// CompatibilityChecker checks workflows for act limitations
type CompatibilityChecker struct {
	unsupportedActions map[string]string
	warningPatterns    []patternCheck
}

type patternCheck struct {
	pattern *regexp.Regexp
	message string
	level   string
}

// NewCompatibilityChecker creates a new compatibility checker
func NewCompatibilityChecker() *CompatibilityChecker {
	return &CompatibilityChecker{
		unsupportedActions: map[string]string{
			"actions/cache@":              "Caching not supported locally",
			"actions/upload-artifact@":    "Artifacts saved to workspace only",
			"actions/download-artifact@":  "Cross-workflow artifacts not supported",
			"docker/build-push-action@":   "Registry push disabled locally",
			"actions/create-release@":     "GitHub releases not supported locally",
			"actions/upload-release-asset@": "Release assets not supported locally",
			"peter-evans/create-pull-request@": "Pull requests not supported locally",
			"github/super-linter@":        "May timeout locally due to resource constraints",
		},
		warningPatterns: []patternCheck{
			{
				pattern: regexp.MustCompile(`\$\{\{\s*secrets\.GITHUB_TOKEN\s*\}\}`),
				message: "GITHUB_TOKEN will be simulated locally",
				level:   "warning",
			},
			{
				pattern: regexp.MustCompile(`\$\{\{\s*github\.event\.pull_request\.\w+\s*\}\}`),
				message: "Pull request events are simulated locally",
				level:   "info",
			},
			{
				pattern: regexp.MustCompile(`\$\{\{\s*github\.repository_owner\s*\}\}`),
				message: "Repository owner will be 'nektos' locally",
				level:   "info",
			},
			{
				pattern: regexp.MustCompile(`if:\s*github\.event_name\s*==\s*'schedule'`),
				message: "Scheduled workflows must be triggered manually locally",
				level:   "warning",
			},
			{
				pattern: regexp.MustCompile(`runs-on:\s*\[\s*self-hosted`),
				message: "Self-hosted runners not supported, will use docker",
				level:   "warning",
			},
			{
				pattern: regexp.MustCompile(`services:`),
				message: "Service containers require Docker networking configuration",
				level:   "info",
			},
			{
				pattern: regexp.MustCompile(`strategy:\s*matrix:`),
				message: "Matrix builds may impact performance locally",
				level:   "info",
			},
		},
	}
}

// CheckWorkflow analyzes a workflow for compatibility issues
func (cc *CompatibilityChecker) CheckWorkflow(workflowData []byte) []Warning {
	warnings := []Warning{}
	content := string(workflowData)
	lines := strings.Split(content, "\n")
	
	// Check for unsupported actions
	for lineNum, line := range lines {
		for action, message := range cc.unsupportedActions {
			if strings.Contains(line, action) {
				// Extract the full action reference
				actionMatch := regexp.MustCompile(`uses:\s*([^\s]+)`).FindStringSubmatch(line)
				actionRef := action
				if len(actionMatch) > 1 {
					actionRef = actionMatch[1]
				}
				
				warnings = append(warnings, Warning{
					Level:   "info",
					Message: fmt.Sprintf("%s: %s", actionRef, message),
					Action:  actionRef,
					Line:    lineNum + 1,
				})
			}
		}
	}
	
	// Check warning patterns
	for _, check := range cc.warningPatterns {
		matches := check.pattern.FindAllStringIndex(content, -1)
		for _, match := range matches {
			// Find line number
			lineNum := strings.Count(content[:match[0]], "\n") + 1
			
			warnings = append(warnings, Warning{
				Level:   check.level,
				Message: check.message,
				Line:    lineNum,
			})
		}
	}
	
	// Check for common issues
	warnings = append(warnings, cc.checkCommonIssues(content)...)
	
	return warnings
}

// checkCommonIssues checks for common workflow issues when running locally
func (cc *CompatibilityChecker) checkCommonIssues(content string) []Warning {
	warnings := []Warning{}
	
	// Check for hardcoded paths
	if strings.Contains(content, "/home/runner/") {
		warnings = append(warnings, Warning{
			Level:   "warning",
			Message: "Hardcoded runner paths may not work locally",
		})
	}
	
	// Check for network-dependent steps without proper handling
	if strings.Contains(content, "curl ") || strings.Contains(content, "wget ") {
		if !strings.Contains(content, "|| true") && !strings.Contains(content, "|| exit") {
			warnings = append(warnings, Warning{
				Level:   "info",
				Message: "Network operations may fail locally without internet access",
			})
		}
	}
	
	// Check for large resource requirements
	if regexp.MustCompile(`timeout-minutes:\s*(\d+)`).MatchString(content) {
		matches := regexp.MustCompile(`timeout-minutes:\s*(\d+)`).FindStringSubmatch(content)
		if len(matches) > 1 && matches[1] > "60" {
			warnings = append(warnings, Warning{
				Level:   "info",
				Message: "Long-running workflows may timeout on local resources",
			})
		}
	}
	
	// Check for GitHub API usage
	if strings.Contains(content, "api.github.com") {
		warnings = append(warnings, Warning{
			Level:   "warning",
			Message: "GitHub API calls require authentication and may be rate-limited",
		})
	}
	
	return warnings
}

// KnownLimitations returns a list of all known act limitations
func (cc *CompatibilityChecker) KnownLimitations() []string {
	limitations := []string{
		"Caching (actions/cache) not supported",
		"Artifacts limited to local workspace",
		"No cross-workflow artifact sharing",
		"Docker registry push disabled",
		"GitHub releases not supported",
		"Pull request creation not supported",
		"GITHUB_TOKEN is simulated",
		"Self-hosted runners use Docker",
		"Service containers need Docker setup",
		"Matrix builds may be slow",
		"GitHub API calls may fail",
		"Scheduled workflows need manual trigger",
	}
	
	return limitations
}

// IsWorkflowSupported does a quick check if a workflow can run at all
func (cc *CompatibilityChecker) IsWorkflowSupported(workflowData []byte) (bool, string) {
	content := string(workflowData)
	
	// Check for completely unsupported features
	if strings.Contains(content, "container-job:") {
		return false, "Container jobs are not fully supported"
	}
	
	if strings.Contains(content, "concurrency:") {
		return false, "Concurrency controls are not supported locally"
	}
	
	// Workflows are generally supported with limitations
	return true, ""
}

// SuggestFixes provides suggestions for common issues
func (cc *CompatibilityChecker) SuggestFixes(warnings []Warning) []string {
	suggestions := []string{}
	
	hasCache := false
	hasArtifacts := false
	hasDocker := false
	
	for _, w := range warnings {
		if strings.Contains(w.Action, "actions/cache") {
			hasCache = true
		}
		if strings.Contains(w.Action, "artifact") {
			hasArtifacts = true
		}
		if strings.Contains(w.Action, "docker") {
			hasDocker = true
		}
	}
	
	if hasCache {
		suggestions = append(suggestions, 
			"Consider using volume mounts for caching dependencies locally")
	}
	
	if hasArtifacts {
		suggestions = append(suggestions,
			"Artifacts will be saved to the workspace output directory")
	}
	
	if hasDocker {
		suggestions = append(suggestions,
			"Ensure Docker daemon is running and accessible")
	}
	
	return suggestions
}