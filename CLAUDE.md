# Claude Code Session Initialization Guide

This document helps Claude Code sessions quickly understand the GitHub Actions Bridge project and get into the correct context for development work.

## Project Overview

**GitHub Actions Bridge** is a tool that bridges GitHub Actions workflows with ConfigHub's configuration management platform. It allows users to:
- Run GitHub Actions workflows locally using `nektos/act`
- Manage workflows as ConfigHub units with centralized configuration
- Execute workflows with configurations from ConfigHub spaces
- Implement GitOps patterns without complex git branching

**Key Insight**: This project turns GitHub Actions workflows into configuration units that can be managed, versioned, and deployed through ConfigHub.

## Essential Context Commands

Run these commands first to understand the environment:

```bash
# 1. Check current git status and branch
git status
git branch --show-current

# 2. Verify cub CLI is installed and get overview
cub --version
cub --help-overview

# 3. Check if act is installed
act --version

# 4. List main project structure
ls -la
ls -la examples/
ls -la pkg/bridge/
```

## IMPORTANT: Version and Date Verification

**ALWAYS verify the current date and version information at the start of each session:**

1. **Check Today's Date**: Look for "Today's date" in the environment context
2. **Verify README.md Version Info**: 
   - Check the "Documentation Updated" date in README.md
   - Verify Go version context is current
   - Update if the date is stale (more than a few days old)

3. **Update Version Context if Needed**:
   ```bash
   # Search for current Go version online if README is outdated
   # Update the Version Context section in README.md with:
   # - Current date
   # - Latest Go stable version
   # - Confirmation that project versions are valid
   ```

4. **Check for Version Discrepancies**:
   - If go.mod shows a Go version that seems too high (e.g., 1.30+), flag it
   - If timestamps show future dates beyond today, investigate
   - Verify ConfigHub SDK pseudo-version timestamp is not in the future

## Critical Files to Read (in order)

1. **Project Documentation**
   - `/Users/alexisrichardson/github-actions-bridge/README.md` - Project overview and setup
   - `/Users/alexisrichardson/github-actions-bridge/go.mod` - Dependencies (IMPORTANT: Note the ConfigHub SDK pseudo-version)
   - `/Users/alexisrichardson/github-actions-bridge/USER_GUIDE.md` - Detailed usage guide
   - `/Users/alexisrichardson/github-actions-bridge/CONFIGHUB_SETUP_GUIDE.md` - Step-by-step ConfigHub setup (NEW!)
   - `/Users/alexisrichardson/github-actions-bridge/CONFIGHUB_TEST_RESULTS.md` - Real test results with learnings
   - `/Users/alexisrichardson/github-actions-bridge/SECURITY.md` - Security considerations

2. **Core Implementation**
   - `/Users/alexisrichardson/github-actions-bridge/cmd/actions-bridge/main.go` - Bridge worker entry point
   - `/Users/alexisrichardson/github-actions-bridge/cmd/actions-cli/main.go` - CLI tool entry point
   - `/Users/alexisrichardson/github-actions-bridge/pkg/bridge/bridge.go` - Core bridge implementation
   - `/Users/alexisrichardson/github-actions-bridge/pkg/bridge/act.go` - Act integration

3. **Example Workflows**
   - `/Users/alexisrichardson/github-actions-bridge/examples/README.md` - Examples overview
   - `/Users/alexisrichardson/github-actions-bridge/examples/hello-world.yml` - Simplest example
   - `/Users/alexisrichardson/github-actions-bridge/examples/config-driven-deployment.yml` - ConfigHub integration example

## GitHub Repositories to Review

1. **This Project**
   - URL: `https://github.com/confighub/actions-bridge`
   - Module Path: `github.com/confighub/actions-bridge`

2. **Core Dependencies**
   - **ConfigHub SDK**: `https://github.com/confighub/sdk`
     - Version: `v0.0.0-20250804044729-f1517379cea0` (pseudo-version - pinned to specific commit)
     - IMPORTANT: This is the ONLY ConfigHub dependency
   - **nektos/act**: `https://github.com/nektos/act` (v0.2.80)
     - The GitHub Actions local runner

3. **Supporting Libraries**
   - **spf13/cobra**: `https://github.com/spf13/cobra` (v1.9.1) - CLI framework
   - **prometheus/client_golang**: `https://github.com/prometheus/client_golang` (v1.22.0) - Metrics

## Understanding the cub CLI

The `cub` CLI is ConfigHub's command-line interface. To understand it:

```bash
# Get comprehensive documentation
cub --help-overview

# Key commands for this project
cub auth login                                    # Login to ConfigHub
cub context get                                   # Show current context
cub context set --space <space-name>             # Set working space
cub worker create <worker-name>                   # Create a worker
cub worker get-envs <worker-name>                # Get worker credentials
cub unit create <unit-name> <workflow.yml>       # Create workflow unit
cub unit apply <unit-name>                       # Execute workflow
```

**CLI Pattern**: `cub <entity> <verb> [flags] [arguments]`

## Key Project Concepts

1. **ConfigHub Resource Header**: All workflows need this header:
   ```yaml
   apiVersion: actions.confighub.com/v1alpha1
   kind: Actions
   metadata:
     name: your-workflow-name
   # Standard GitHub Actions workflow follows
   ```

2. **Bridge Worker Protocol**: Implements ConfigHub's worker interface:
   - `Info()` - Returns worker capabilities
   - `Apply()` - Executes the workflow
   - `Refresh()` - Gets workflow status
   - `Destroy()` - Cleanup

3. **Workspace Management**: Each workflow execution gets an isolated workspace in `/tmp`

## Common Development Tasks

### Running the Bridge Worker
```bash
# Set worker credentials (from cub worker get-envs)
export CONFIGHUB_WORKER_ID=xxx
export CONFIGHUB_WORKER_SECRET=xxx
export CONFIGHUB_URL=https://hub.confighub.com

# Run the bridge
./bin/actions-bridge
```

### Testing Workflows Locally
```bash
# Using the CLI tool
./bin/cub-local-actions run examples/hello-world.yml

# With validation
./bin/cub-local-actions validate examples/build.yml

# With secrets
./bin/cub-local-actions run examples/deploy.yml --secrets-file secrets.env
```

### Building the Project
```bash
make build     # Build binaries
make test      # Run tests
make docker    # Build Docker image
```

## Important Gotchas and Notes

1. **Dependency Versions**
   - The ConfigHub SDK uses a pseudo-version (timestamp-based), not a tagged release
   - This means it's pinned to a specific commit, not a stable version
   - Be careful when updating dependencies

2. **Docker Socket Access**
   - The bridge requires Docker socket access (security risk)
   - See SECURITY.md for production deployment guidance
   - Use docker-compose.secure.yml for production

3. **Act Limitations**
   - Some GitHub Actions features don't work locally (caching, artifacts to GitHub)
   - See examples/README.md for workarounds

4. **Workflow Format**
   - MUST include the ConfigHub resource header
   - The bridge strips these headers before passing to act

5. **File Paths**
   - All file paths in tool calls should be absolute (start with /)
   - The working directory is `/Users/alexisrichardson/github-actions-bridge`

6. **PATH Issue for cub CLI**
   - The cub installer places the binary at `~/.confighub/bin/cub`
   - This is NOT in PATH by default
   - Users must add: `export PATH="$HOME/.confighub/bin:$PATH"`
   - See CONFIGHUB_SETUP_GUIDE.md for details

7. **ConfigHub URL**
   - MUST use `https://hub.confighub.com` (NOT api.confighub.com)
   - Wrong URL causes "no such host" errors
   - This is a common mistake in documentation

8. **Target Requirement**
   - ALL `cub unit create` commands MUST include `--target docker-desktop`
   - Without target: "cannot invoke action on a unit without a target" error
   - This is not well documented in ConfigHub's main docs

## Quick Verification Checklist

- [ ] Can run `cub --version` successfully
- [ ] Can run `act --version` successfully
- [ ] Docker is running (`docker ps`)
- [ ] Understand the project is at `github.com/confighub/actions-bridge`
- [ ] Know that ConfigHub SDK is the only ConfigHub dependency
- [ ] Aware of the pseudo-version for ConfigHub SDK

## AI Hallucination Detection and Remediation

### Recommended Prompt for Hallucination Detection

When checking for AI hallucinations in this project, use this prompt:

```
Please perform an AI Hallucination Detection and Remediation analysis on the GitHub Actions Bridge documentation.

Context:
- Today's date: [INSERT TODAY'S DATE]
- Current Go stable version: [CHECK AND INSERT]
- Focus on detecting overpromised features, simulated functionality presented as real, and version/date discrepancies

Check for:
1. **Version/Date Issues**: Future dates, non-existent versions, timestamps beyond today
2. **Simulated Features**: Examples that mock functionality but claim it's real (especially AI integrations)
3. **Overpromised Capabilities**: Documentation that exceeds actual implementation
4. **Unverifiable Claims**: Features, integrations, or platforms that might not exist
5. **Mock vs Real**: Examples using simulation/mocking but not clearly marked as such

Review these files:
- README.md (check version context and feature claims)
- EXAMPLES_COMPATIBILITY.md (verify accuracy of compatibility claims)
- examples/*.yml (especially claude-*.yml and time-travel-*.yml)
- Any files claiming AI integration or advanced features

For each issue found:
1. Identify the specific file and line
2. Explain why it's likely a hallucination
3. Suggest a correction that maintains honesty about current capabilities
4. Distinguish between "planned features" vs "current features"

Remember: Some examples are intentionally conceptual/simulated to show possibilities. These should be clearly marked as such, not presented as working features.
```

### Key Hallucination Patterns in This Project

Based on previous analysis, watch for:

1. **Claude AI Integration** - Often simulated with shell scripts
2. **Time Travel Testing** - Usually just date comparisons, not real versioning
3. **Advanced ConfigHub Features** - May be mocked in examples
4. **Version Numbers** - Go versions above 1.24-1.25 in 2025 are suspicious
5. **Future Timestamps** - Dates beyond today in version strings

### When You Find Hallucinations

1. **Document them** in AI_HALLUCINATION_FINDINGS.md
2. **Update documentation** to clarify what's real vs simulated
3. **Add disclaimers** to conceptual examples
4. **Update EXAMPLES_COMPATIBILITY.md** with accurate status

## Need More Context?

1. Run `cub --help-overview` for comprehensive cub CLI documentation
2. Read `/Users/alexisrichardson/github-actions-bridge/docs/INDEX.md` for full documentation map
3. Check `/Users/alexisrichardson/github-actions-bridge/examples/` for practical examples
4. Review recent commits with `git log --oneline -10`
5. Check AI_HALLUCINATION_FINDINGS.md for known issues

---
*This initialization guide ensures Claude Code sessions have the correct context for working on the GitHub Actions Bridge project.*