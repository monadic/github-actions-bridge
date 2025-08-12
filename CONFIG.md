# ConfigHub Actions Bridge - Setup & Verification Guide

## Quick Status Check

Run this one-liner to see your setup status:
```bash
docker ps >/dev/null 2>&1 && echo "✅ Docker" || echo "❌ Docker not running"
act --version >/dev/null 2>&1 && echo "✅ act" || echo "❌ act not installed"
which cub >/dev/null && echo "✅ cub CLI" || echo "❌ cub CLI not found" 
cub context get >/dev/null 2>&1 && echo "✅ Authenticated" || echo "❌ Not authenticated/expired"
[ -n "$CONFIGHUB_WORKER_ID" ] && echo "✅ Worker configured" || echo "❌ Worker not configured"
```

## Pre-Flight Checklist

### 1. System Requirements
```bash
# Check Docker is running
docker --version && docker ps
# Expected: Docker version 20.10+ and no errors

# Check act is installed
act --version
# Expected: act version 0.2.50+

# Check cub CLI is accessible
which cub
# Expected: /Users/<you>/.confighub/bin/cub
# If not found: export PATH="$HOME/.confighub/bin:$PATH"

# Verify cub works
cub version
# Expected: ConfigHub CLI version output
```

### 2. ConfigHub Authentication
```bash
# Login to ConfigHub
cub auth login
# Expected: Opens browser for auth

# Verify authentication
cub context get
# Expected: Shows user, URL, and current space
# If "token expired": run cub auth login again

# Check if token is valid
cub worker list >/dev/null 2>&1 && echo "Token valid" || echo "Token expired/invalid"

# Set workspace
cub context set --space my-test-space
# Expected: Context updated successfully
```

## Project-Specific Requirements

This project requires:
- **act**: GitHub Actions local runner (`brew install act`)
- **Docker**: Required by act to run containers
- **Go 1.23+**: For building from source

## Project Setup

### 3. Clone and Build
```bash
# Clone the project
git clone https://github.com/confighub/actions-bridge.git
cd actions-bridge

# Build binaries
make build
# Expected: Binary created at ./bin/actions-bridge

# Verify build
./bin/actions-bridge --version
# Expected: Shows version info
```

### 4. Create Worker or Use Existing
```bash
# Check if worker already exists with target mapping
cub target list | grep -E "(WORKER-SLUG|docker-desktop)"
# If shows "actions-bridge-1" or similar, use that worker

# Option A: Use existing worker
cub worker get-envs actions-bridge-1  # Use name from target list
# Expected: Shows CONFIGHUB_WORKER_ID and CONFIGHUB_WORKER_SECRET

# Option B: Create new worker (only if needed)
cub worker create actions-worker
# Expected: Worker created successfully
# Then create target: cub target create ... (complex - see docs)

# Set credentials (from either option)
export CONFIGHUB_WORKER_ID=<shown-id>
export CONFIGHUB_WORKER_SECRET=<shown-secret>
export CONFIGHUB_URL=https://hub.confighub.com
```

## Verification Tests

### 5. Test Local CLI
```bash
# Test with simplest workflow
./bin/cub-local-actions run examples/hello-world.yml
# Expected: "Hello from ConfigHub!" output

# Validate workflow syntax
./bin/cub-local-actions validate examples/hello-world.yml
# Expected: "Workflow is valid"
```

### 6. Test Bridge Worker
```bash
# Start the bridge worker with visible logs
./bin/actions-bridge 2>&1 | tee bridge.log &
# Expected: Shows connection to ConfigHub event stream

# Wait for worker to be ready
sleep 5 && cub worker get <worker-name> | grep Condition
# Expected: "Condition       Ready"

# Check health
curl http://localhost:8080/health
# Expected: {"status":"healthy","service":"github-actions-bridge"}

# Monitor logs during apply
tail -f bridge.log | grep -v heartbeat
```

### 7. Test ConfigHub Integration
```bash
# Create a unit from example
cub unit create hello-test examples/hello-world.yml --target docker-desktop
# Expected: Unit created successfully

# Apply the unit
cub unit apply hello-test
# Expected: Shows execution output with "Hello from ConfigHub!"

# Check status
cub unit get hello-test
# Expected: Shows status as "completed"
```

## Common Issues Checklist

### ❌ If "token expired" error:
```bash
# Re-authenticate
cub auth login
# This opens browser for fresh login
```

### ❌ If `cub` command not found:
```bash
# Add to PATH
export PATH="$HOME/.confighub/bin:$PATH"
# Add to shell profile permanently
echo 'export PATH="$HOME/.confighub/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### ❌ If "no such host" error:
```bash
# Check URL is correct
echo $CONFIGHUB_URL
# MUST be: https://hub.confighub.com
# NOT: https://api.confighub.com
```

### ❌ If "cannot invoke action without target":
```bash
# Always include --target
cub unit create test workflow.yml --target docker-desktop
```

### ❌ If unit apply hangs forever:
```bash
# Check target-to-worker mapping
cub target list | grep docker-desktop
# Note the WORKER-SLUG (e.g., actions-bridge-1)

# Use the worker that matches the target
cub worker get-envs <WORKER-SLUG-FROM-TARGET>
# Re-export credentials and restart bridge
```

### ❌ If Docker errors:
```bash
# Ensure Docker Desktop is running
open -a Docker  # macOS
# Wait for Docker to start fully
docker ps  # Should not error
```

### ❌ If act not found:
```bash
# Install act first
brew install act  # macOS
# or
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

## Quick Validation Script

Save this as `check-setup.sh`:
```bash
#!/bin/bash
echo "=== ConfigHub Actions Bridge Setup Check ==="

# Check commands
for cmd in docker act cub make go; do
    if command -v $cmd &> /dev/null; then
        echo "✅ $cmd found"
    else
        echo "❌ $cmd NOT FOUND"
    fi
done

# Check cub in specific location if not in PATH
if ! command -v cub &> /dev/null; then
    if [ -f "$HOME/.confighub/bin/cub" ]; then
        echo "⚠️  cub found at ~/.confighub/bin/cub but not in PATH"
    fi
fi

# Check Docker
if docker ps &> /dev/null; then
    echo "✅ Docker is running"
else
    echo "❌ Docker not running"
fi

# Check ConfigHub auth
if cub context get &> /dev/null; then
    echo "✅ ConfigHub authenticated"
else
    echo "❌ ConfigHub not authenticated"
fi

# Check environment
for var in CONFIGHUB_WORKER_ID CONFIGHUB_WORKER_SECRET CONFIGHUB_URL; do
    if [ -n "${!var}" ]; then
        echo "✅ $var is set"
    else
        echo "❌ $var NOT SET"
    fi
done

echo "=== Check Complete ==="
```

## Next Steps

Once all checks pass:
1. Run example workflows: `ls examples/`
2. Create your own workflows
3. Set up production deployment
4. Read security guidelines

## Essential Commands Reference

### Setup
```bash
cub auth login                              # Authenticate to ConfigHub
cub context get                             # View current context
cub context set --space my-test-space      # Set working space
```

### Working with Units (Workflows)
```bash
cub unit create name file.yml --target docker-desktop
cub unit apply name                         # Execute workflow
cub unit get name                           # Check status
cub unit list                               # List all units
cub unit delete name                        # Remove unit
cub unit history name                       # View execution history
```

### Worker Management
```bash
cub worker create my-worker                 # Create new worker
cub worker list                             # List all workers  
cub worker get my-worker                    # Get worker details
cub worker get-envs my-worker              # Get credentials for export
cub worker delete my-worker                 # Remove worker
```

### Target Management
```bash
cub target list                             # List available targets
cub target get docker-desktop               # Get target details
```

### Key Environment Variables
```bash
CONFIGHUB_URL=https://hub.confighub.com    # Hub API endpoint (NOT api.confighub.com)
CONFIGHUB_WORKER_ID=<from-get-envs>        # Worker authentication
CONFIGHUB_WORKER_SECRET=<from-get-envs>    # Worker secret
CONFIGHUB_SPACE=<space-name>               # Default space (optional)
```

## Emergency Commands

```bash
# Stop all containers
docker stop $(docker ps -aq)

# Kill bridge worker
pkill actions-bridge

# Reset ConfigHub context
cub auth logout
cub auth login

# Clean build artifacts
make clean
```

---
*This CONFIG.md provides concrete steps to verify your GitHub Actions Bridge setup is working correctly.*

## To Check Status

Ask Claude (or any AI tool) to: "Run CONFIG.md and report project status"

The AI will execute the verification steps and provide a status report like:

```
Project Status Report

Quick Status Summary:
- ✅/❌ Docker running/not running
- ✅/❌ act installed/not installed  
- ✅/❌ cub CLI installed/not found
- ✅/⚠️/❌ Authenticated/expired/not authenticated
- ✅/❌ Worker configured/not configured
- ✅/❌ Binaries built/not built
- ✅/❌ Bridge running/not running

Required Actions:
[List of specific steps needed to fix any ❌ items]

Current State:
- Build: [status]
- Dependencies: [status]
- Authentication: [status]
- Worker: [status]
- Runtime: [status]
```

This provides a quick health check of your entire setup.

<!--
## CONFIG.md Pattern for ConfigHub SDK Projects

This file demonstrates a reusable pattern:
1. **Quick Status Check** - One-liner health check
2. **Core Dependencies** - cub CLI and ConfigHub auth  
3. **Project-Specific Requirements** - What makes this project unique
4. **Worker Setup** - Standard ConfigHub worker pattern
5. **Verification Tests** - Prove everything works
6. **Common Issues** - Known gotchas with fixes

For other ConfigHub SDK projects, adapt sections 3 and 5 to your specific needs.

## Lessons Learned from GitHub Actions Bridge

### Project-Specific Insights:
1. **Docker Dependency Chain** - Docker must be running BEFORE attempting any act operations
2. **Environment Variable Scope** - Worker credentials set in subshell won't persist; need explicit export
3. **Background Process Management** - Use `&` to background the bridge, but note it ties to terminal session
4. **Health Checks Are Essential** - The /health endpoint confirms worker is truly running vs just started
5. **Token Expiry is Common** - ConfigHub tokens expire; always check with `cub worker list` as canary

### Status Check Best Practices:
1. **Layer Your Checks** - Quick one-liner first, then detailed verification
2. **Check Process AND Endpoint** - Process running doesn't mean service is healthy
3. **Verify Worker Registration** - Check both local process and ConfigHub's view of worker
4. **Include PID in Status** - Helps identify specific instances when debugging
5. **Test a Real Operation** - Consider adding "create and apply test unit" to verification

### Generalizing to Other ConfigHub Projects:

#### For Infrastructure Workers (Terraform, Pulumi):
- Replace "Docker/act" checks with terraform/pulumi CLI checks
- Add state file verification steps
- Include plan/apply dry-run tests

#### For Container Workers (Docker, K8s):
- Check kubectl/docker context
- Verify namespace/registry access
- Add deployment rollout status checks

#### For Custom Workers:
- Define worker-specific health endpoints
- Create minimal test configurations
- Document any special network requirements

### Universal ConfigHub Patterns:
1. **Always Check Token First** - Use `cub worker list` as auth canary
2. **Worker ID Not in Environment** - Common issue; provide explicit export commands
3. **URL Must Be Exact** - hub.confighub.com, never api.confighub.com
4. **Target is Required** - Every unit create needs --target specified
5. **Space Context Matters** - Always show current space in status

### AI Tool Integration Tips:
1. **Executable Documentation** - CONFIG.md should be runnable, not just readable
2. **Clear Expected Outputs** - Show exact strings/patterns to match
3. **Graceful Degradation** - Check each dependency independently
4. **Actionable Errors** - Every ❌ should have a specific fix command
5. **State Verification** - Don't assume; always verify current state

### Future CONFIG.md Enhancements:
- Add automated fix attempts for common issues
- Include rollback procedures for failed setups
- Add performance baselines (startup time, memory usage)
- Create CONFIG.schema.json for validation
- Support for multiple deployment targets
-->