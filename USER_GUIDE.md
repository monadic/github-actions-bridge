# GitHub Actions Bridge - User Guide

[← Back to README](README.md) | [Examples →](examples/) | [CLI Reference →](CLI_REFERENCE.md)

---

Welcome! This guide will walk you through using the GitHub Actions Bridge step by step.

## Table of Contents
1. [Understanding the Problem](#understanding-the-problem)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [🚀 Quick Start - Choose Your Path](#-quick-start---choose-your-path)
5. [Your First Workflow](#your-first-workflow)
6. [Working with Secrets](#working-with-secrets)
7. [Local Development Workflow](#local-development-workflow)
8. [ConfigHub Integration](#confighub-integration)
9. [Common Use Cases](#common-use-cases)
10. [Troubleshooting](#troubleshooting)
11. [Next Steps](#next-steps)

## Understanding the Problem

Managing modern applications involves juggling multiple systems:
- ❌ Workflows in GitHub Actions
- ❌ Configurations in various config files
- ❌ Secrets in different secret managers
- ❌ No unified way to manage them together
- ❌ Configuration changes don't automatically update workflows

The GitHub Actions Bridge solves these problems by:
- ✅ Unifying workflows and configurations through ConfigHub
- ✅ Running GitHub Actions anywhere (not just on GitHub)
- ✅ Managing secrets in one secure place
- ✅ Making workflows configuration-driven

## Prerequisites

Before you start, you'll need:

1. **Docker Desktop** installed and running
   - Mac: Download from [docker.com](https://www.docker.com/products/docker-desktop)
   - Linux: Install Docker Engine
   - Windows: Use WSL2 with Docker Desktop
   - Verify: `docker --version`

2. **act** - The GitHub Actions local runner (REQUIRED)
   - Mac: `brew install act`
   - Linux: `curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash`
   - Windows: `choco install act-cli`
   - Verify: `act --version`
   - More info: [nektos/act](https://github.com/nektos/act)

3. **Go 1.21+** (only if building from source)
   - Check with: `go version`

4. **Basic command line knowledge**
   - How to open a terminal
   - How to navigate directories (`cd`)
   - How to run commands

**Important:** The GitHub Actions Bridge uses `act` to run workflows locally. Without act installed, the bridge cannot execute any workflows.

## Installation

### Option 1: Install ConfigHub CLI (Easiest)

```bash
# First, install act (if not already installed)
brew install act  # macOS
# or for Linux: curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Verify act is installed
act --version

# Install ConfigHub CLI (includes the actions-bridge worker)
curl -fsSL https://hub.confighub.com/cub/install.sh | bash

# IMPORTANT: Add cub to your PATH
# The installer places cub at ~/.confighub/bin/cub
export PATH="$HOME/.confighub/bin:$PATH"

# Make it permanent:
echo 'export PATH="$HOME/.confighub/bin:$PATH"' >> ~/.zshrc  # For macOS/zsh
# OR
echo 'export PATH="$HOME/.confighub/bin:$PATH"' >> ~/.bashrc  # For Linux/bash

# Reload your shell:
source ~/.zshrc  # or source ~/.bashrc

# Test it works
cub version

# Login to ConfigHub
cub auth login
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/confighub/actions-bridge
cd actions-bridge

# Build everything
make build

# The binaries will be in ./bin/
./bin/actions-bridge --version
```

## 🚀 Quick Start - Choose Your Path

This guide will help you run your first GitHub Actions workflow. We'll start with ConfigHub (if you have an account) or jump straight to local execution.

### Decision Point: Do You Have a ConfigHub Account?

**Option A: Yes, I have a ConfigHub account** → Continue to [ConfigHub Workflow](#confighub-workflow-setup)  
**Option B: No, I want to test locally first** → Skip to [Local Testing](#local-testing-quickstart)

---

## ConfigHub Workflow Setup

**📖 For a complete step-by-step guide with troubleshooting, see [CONFIGHUB_SETUP_GUIDE.md](CONFIGHUB_SETUP_GUIDE.md)**

### Step 1: Verify Your Tools

First, let's make sure you have everything needed:

```bash
# 1. Check if Docker is running
docker ps
```

**What to expect:**
- ✅ **Success**: You see either an empty table with headers (CONTAINER ID, IMAGE, etc.) or a list of running containers
- ❌ **Error**: "Cannot connect to the Docker daemon" or similar error message

**Did it work?**
- ✅ **Yes** → Docker is ready! Continue checking other tools
- ❌ **No** → Start Docker Desktop:
  - Mac: Open Docker Desktop from Applications
  - Linux: `sudo systemctl start docker`
  - Wait for Docker to fully start (icon shows "Docker Desktop is running")

**Important:** You do NOT need to:
- Create any containers manually in Docker Desktop
- Pull any images beforehand
- Configure anything in Docker Desktop

The bridge handles all Docker operations automatically when you run workflows.

**What `docker ps` verifies:**
- ✅ Docker daemon is running
- ✅ Docker CLI can communicate with the daemon
- ✅ You have permissions to use Docker
- That's all you need!

```bash
# 2. Check if act is installed
act --version
```

**Did it work?**
- ✅ **Yes** → Continue to check cub
- ❌ **No** → Install act (see Prerequisites section above)

```bash
# 3. Check if cub is installed
cub version
```

**Did it work?**
- ✅ **Yes** → Continue to Step 2
- ❌ **No** → Install cub first:
  ```bash
  curl -fsSL https://hub.confighub.com/cub/install.sh | bash
  # Add to PATH if needed
  sudo ln -sf ~/.confighub/bin/cub /usr/local/bin/cub
  ```

### Step 2: Log In to ConfigHub

```bash
# This opens your browser for authentication
cub auth login
```

**Verification**: Check you're logged in:
```bash
cub context get
```

You should see something like:
```
User Email             your-email@example.com
IDP User ID            user_xxxxxxxxxxxxxxxxxxxx
IDP Organization ID    org_xxxxxxxxxxxxxxxxxxxxx
ConfigHub URL          https://hub.confighub.com
Space                  default (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
Organization           xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

**Troubleshooting**:
- No organization shown? → Contact your ConfigHub admin
- Authentication failed? → Check your browser completed login
- Command not found? → Ensure cub is in your PATH

### Step 3: Choose Your Setup Parameters

**Quick Setup (Recommended for First Time)**

I'll use these defaults:
- Space name: `actions-demo`
- Worker name: `bridge-worker-1`

Want to use these? Type **Y** (or continue reading for custom setup)

**Custom Setup**

Choose your own:
- Space name: _________________ (letters, numbers, hyphens)
- Worker name: _________________ (letters, numbers, hyphens)

### Step 4: Create Your Space (if needed)

```bash
# For default setup
cub space create actions-demo

# For custom (replace YOUR-SPACE-NAME)
cub space create YOUR-SPACE-NAME
```

Set it as your working space:
```bash
# For default
cub context set --space actions-demo

# For custom
cub context set --space YOUR-SPACE-NAME
```

**Verify**: 
```bash
cub context get
# Should show Space: actions-demo (or your custom name)
```

### Step 5: Create and Start the Worker

```bash
# Create worker (default names)
cub worker create bridge-worker-1

# Get credentials
eval "$(cub worker get-envs bridge-worker-1)"

# Verify environment is set
echo $CONFIGHUB_WORKER_ID
# Should show a UUID
```

Now start the bridge in a **new terminal**:
```bash
cd github-actions-bridge

# IMPORTANT: Set the worker credentials in this new terminal
# Option 1: Run the eval command again AND set the URL
eval "$(cub worker get-envs bridge-worker-1)"
export CONFIGHUB_URL=https://hub.confighub.com

# Option 2: Or export ALL variables manually
# export CONFIGHUB_WORKER_ID=your-worker-id
# export CONFIGHUB_WORKER_SECRET=your-worker-secret
# export CONFIGHUB_URL=https://hub.confighub.com  # CRITICAL: Must be hub, not api!

# Now start the bridge
./bin/actions-bridge
```

You should see:
```
2025/08/07 10:30:45 Starting GitHub Actions Bridge worker...
2025/08/07 10:30:45 Worker ID: [your-id]
2025/08/07 10:30:45 Connected to ConfigHub
```

**Keep this terminal running!**

### Step 6: Run Your First ConfigHub Workflow

**Note:** This is when Docker is actually used - the bridge will:
- Automatically pull the `ubuntu-latest` image (or any other runner image specified)
- Create temporary containers for each job
- Clean up containers after execution
- You'll see this activity in Docker Desktop, but no manual intervention is needed

Back in your original terminal:

```bash
# Create a simple test workflow
cat > hello-confighub.yml << 'EOF'
apiVersion: actions.confighub.com/v1alpha1
kind: Actions  
metadata:
  name: hello-confighub
name: Hello from ConfigHub
on: push
jobs:
  greet:
    runs-on: ubuntu-latest
    steps:
      - name: Say Hello
        run: |
          echo "🎉 Hello from ConfigHub!"
          echo "Running via: ConfigHub + GitHub Actions Bridge"
          echo "Time: $(date)"
EOF

# Create the unit (MUST specify target!)
cub unit create hello hello-confighub.yml --target docker-desktop

# Run it!
cub unit apply hello
```

**Success looks like**:
```
Unit "hello" applied successfully
Status: Completed
Exit Code: 0
```

### Step 7: Troubleshooting ConfigHub Issues

**Problem: "dial tcp: lookup api.confighub.com: no such host"**
- The bridge is using the wrong ConfigHub URL
- You MUST set: `export CONFIGHUB_URL=https://hub.confighub.com`
- NOT `api.confighub.com` - this URL doesn't exist!
- Make sure to set this in the terminal where you run `./bin/actions-bridge`

**Problem: "Worker not found"**
- Is the bridge still running in the other terminal?
- Check: `cub worker list`

**Problem: "Unit creation failed"**
- Check YAML syntax: `cat hello-confighub.yml`
- Verify space context: `cub context get`

**Problem: "Apply failed" or "Apply didn't complete on unit"**
- First, check if the bridge worker is still running in the other terminal
- Look for error messages in the bridge terminal output
- Common causes:
  - Bridge worker crashed or was stopped
  - Workflow YAML has syntax errors
  - Docker image pull failed (check internet connection)
  - Insufficient disk space for Docker images
- Debug steps:
  ```bash
  # 1. Check if worker is registered
  cub worker list
  
  # 2. Check unit details
  cub unit get hello --verbose
  
  # 3. Verify Docker is working
  docker run hello-world
  
  # 4. Restart the bridge worker if needed
  # In the bridge terminal, Ctrl+C then start again:
  eval "$(cub worker get-envs bridge-worker-1)"
  export CONFIGHUB_URL=https://hub.confighub.com
  ./bin/actions-bridge
  ```

---

## Local Testing Quickstart

Don't have a ConfigHub account? No problem! Let's test locally:

### Step 1: Verify Local Tools

```bash
# Check Docker
docker --version

# Check act
act --version

# Build the local CLI
make build
ls ./bin/cub-local-actions
```

### Step 2: Run Your First Local Workflow

```bash
# Create a test workflow
cat > hello-local.yml << 'EOF'
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: hello-local
name: Hello Local
on: push
jobs:
  greet:
    runs-on: ubuntu-latest
    steps:
      - run: echo "🚀 Hello from Local Development!"
      - run: echo "No ConfigHub needed!"
EOF

# Run it locally
./bin/cub-local-actions run hello-local.yml
```

**Success looks like**:
```
[Hello Local/greet] 🏁 Job succeeded
```

---

## Next Steps After Quick Start

### If You Used ConfigHub

Try more examples:
```bash
# See which examples work with ConfigHub
cat EXAMPLES_COMPATIBILITY.md

# Run a more complex example
cub unit create build examples/build-test-deploy.yml --target docker-desktop
cub unit apply build
```

### If You Used Local Testing

Explore more:
```bash
# Run with verbose output
./bin/cub-local-actions run hello-local.yml -v

# Try other examples
./bin/cub-local-actions run examples/hello-world.yml
./bin/cub-local-actions run examples/multi-job.yml
```

---

## Your First Workflow

Now that you've verified your setup works, let's understand what just happened and dive deeper!

### Step 1: Create a test workflow

Create a file called `hello.yml`:

```yaml
name: Hello World
on: push

jobs:
  greet:
    runs-on: ubuntu-latest
    steps:
      - name: Say Hello
        run: echo "Hello from GitHub Actions Bridge!"
      
      - name: Show date
        run: date
      
      - name: List files
        run: ls -la
```

### Step 2: Run the workflow

```bash
# Create a ConfigHub unit for the workflow (MUST specify target!)
cub unit create --space default hello hello.yml --target docker-desktop

# Apply (run) the workflow
cub unit apply --space default hello
```

You should see output like:
```
Running workflow: hello.yml
[OK] Basic workflow execution successful
Execution completed in 5.2s
Exit code: 0
```

### Step 3: Run with more details

```bash
# See workflow details without running
cub unit get --space default hello --verbose

# Apply with verbose output
cub unit apply --space default hello --debug
```

**Congratulations!** You've just run your first GitHub Actions workflow locally. Let's explore more features.

## Working with Secrets

One of the biggest challenges with GitHub Actions is testing workflows that need secrets. The bridge solves this elegantly.

### The Problem with Secrets

In GitHub, you'd set secrets in your repository settings:
```yaml
# This won't work locally without the bridge
env:
  API_KEY: ${{ secrets.API_KEY }}
  DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

### The Solution

1. **Create a secrets file** (never commit this!):
```bash
# secrets.env
API_KEY=sk_live_abcd1234
DATABASE_URL=postgresql://user:pass@localhost/mydb
GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

2. **Use secrets in your workflow**:
```yaml
# deploy.yml
name: Deploy Application
on: push

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Connect to Database
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          echo "Connecting to database..."
          # Your database operations here
          
      - name: Deploy to API
        env:
          API_KEY: ${{ secrets.API_KEY }}
        run: |
          echo "Deploying with API key (length: ${#API_KEY})"
          # Your deployment code here
```

3. **Run with secrets**:
```bash
# Create unit and ConfigHub manages secrets
cub unit create --space production deploy deploy.yml --target docker-desktop
cub unit apply --space production deploy
```

**Security Notes:**
- Secrets are never logged or displayed
- Files are created with restricted permissions (0600)
- Secrets are cleaned up after execution
- Add `secrets.env` to `.gitignore`

## Local Development Workflow

Sometimes you want to test workflows quickly without ConfigHub. The `cub-local-actions` CLI lets you run workflows directly on your machine.

### When to Use Local Development

Use `cub-local-actions` when you:
- Need to test workflows during development
- Want quick iteration without ConfigHub setup
- Are debugging workflow issues
- Don't need centralized configuration

### Getting Started with Local Development

1. **Build the local CLI**:
```bash
make build
# Creates ./bin/cub-local-actions
```

2. **Run a workflow directly**:
```bash
# Basic execution
./bin/cub-local-actions run examples/hello-world.yml

# With secrets file
./bin/cub-local-actions run examples/deploy.yml --secrets-file secrets.env

# Validate without running
./bin/cub-local-actions validate examples/complex-workflow.yml
```

3. **Use watch mode for development**:
```bash
# Re-runs workflow when file changes
./bin/cub-local-actions run workflow.yml --watch
```

### Local CLI Commands

| Command | Description | Example |
|---------|-------------|---------|
| `run` | Execute a workflow | `cub-local-actions run workflow.yml` |
| `validate` | Check workflow syntax | `cub-local-actions validate workflow.yml` |
| `clean` | Remove temporary files | `cub-local-actions clean` |
| `version` | Show version info | `cub-local-actions version` |

### Example: Local Testing Workflow

```bash
# 1. Create a test workflow
cat > test-workflow.yml << 'EOF'
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: test
name: Test Workflow
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Running tests..."
      - run: npm test
EOF

# 2. Validate it
./bin/cub-local-actions validate test-workflow.yml

# 3. Run it
./bin/cub-local-actions run test-workflow.yml

# 4. Run with environment variables
./bin/cub-local-actions run test-workflow.yml --env-file .env

# 5. Debug mode
./bin/cub-local-actions run test-workflow.yml -v
```

### Local vs ConfigHub Workflows

| Feature | Local (`cub-local-actions`) | ConfigHub (`cub`) |
|---------|----------------------------|-------------------|
| Setup | No setup needed | Requires ConfigHub account |
| Secrets | Local files | Centralized management |
| Collaboration | Single user | Team access |
| Versioning | Git only | Configuration history |
| Environments | Manual | Built-in (dev/staging/prod) |
| Triggers | Manual only | Automated triggers |

### Tips for Local Development

1. **Use the compatibility guide**: Check [Examples Compatibility](EXAMPLES_COMPATIBILITY.md) to see which examples work locally
2. **Start simple**: Test with `hello-world.yml` first
3. **Use verbose mode**: Add `-v` flag for debugging
4. **Clean regularly**: Run `cub-local-actions clean` to remove temp files

See the [CLI Reference](CLI_REFERENCE.md) for complete `cub-local-actions` documentation.

## ConfigHub Integration

While the bridge works great standalone, its real power comes from ConfigHub integration.

### What is ConfigHub?

ConfigHub is a configuration management platform that:
- Stores configurations and secrets securely
- Manages different environments (dev, staging, production)
- Tracks configuration history
- Integrates with your workflows

### Setting Up ConfigHub Integration

1. **Get ConfigHub credentials** (from your ConfigHub admin)
2. **Set environment variables**:
```bash
export CONFIGHUB_WORKER_ID=your-worker-id
export CONFIGHUB_WORKER_SECRET=your-secret
export CONFIGHUB_URL=https://hub.confighub.com  # IMPORTANT: Use hub, not api!
```

3. **Run workflows with ConfigHub**:
```bash
# Use configurations from ConfigHub
cub unit create --space production webapp deploy.yml --target docker-desktop
cub unit apply --space production webapp

# Test with different environments
cub unit create --space staging webapp deploy.yml --target docker-desktop
cub unit apply --space staging webapp
cub unit create --space development webapp deploy.yml --target docker-desktop
cub unit apply --space development webapp
```

### Advanced ConfigHub Features

**Time Travel Testing:**
```bash
# Test with a previous revision
cub unit apply --space prod webapp --restore 1
```

**Configuration-Driven Workflows:**
```bash
# All values come from ConfigHub
cub unit create --space production config-deploy examples/config-driven-deployment.yml --target docker-desktop
cub unit apply --space production config-deploy
```

See the [ConfigHub examples](examples/) for more advanced use cases.

## Common Use Cases

### 1. Running with Environment Variables

Create an environment file (`dev.env`):
```bash
ENVIRONMENT=development
API_URL=https://api-dev.example.com
DEBUG=true
```

Run with environment:
```bash
# Environment variables are managed through ConfigHub
cub unit create --space dev myworkflow workflow.yml --target docker-desktop
cub unit apply --space dev myworkflow
```

### 3. Collecting Artifacts

Some workflows create files you want to keep:

```yaml
name: Build Project
on: push

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Create artifact
        run: |
          mkdir output
          echo "Build complete!" > output/result.txt
          date > output/build-time.txt
```

Run and save artifacts:
```bash
# Create unit and run with ConfigHub
cub unit create --space default build build.yml
cub unit apply --space default build
# Artifacts are managed through ConfigHub workers
```

### 4. Validating Workflows

Before running, check if a workflow will work:

```bash
# Validate by creating a unit without applying
cub unit create --space default test-workflow workflow.yml --dry-run
```

This will tell you about:
- Syntax errors
- Unsupported features
- Compatibility issues

## Understanding Limitations

Some GitHub Actions features don't work locally:

### What doesn't work:
- `actions/cache` - No caching support
- Creating pull requests
- Pushing to GitHub
- GitHub API calls (unless you provide a token)
- Cross-workflow artifacts

### What does work:
- Running commands
- Using Docker containers
- Environment variables
- Secrets (via files)
- Creating local artifacts
- Most popular actions

To see all limitations:
```bash
# Check ConfigHub documentation for limitations
cub help
# Or refer to act documentation for execution limitations
```

## Advanced Patterns

### GitOps Workflow

ConfigHub enables GitOps patterns for workflow management:

```bash
# 1. Make changes locally
edit workflow.yml

# 2. Preview changes
cub unit plan my-workflow

# 3. Apply to staging
cub context set --space staging
cub unit apply my-workflow

# 4. Promote to production
cub context set --space production
cub unit apply my-workflow
```

### Multi-Environment Setup

Manage different configurations per environment:

```yaml
# base-workflow.yml
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: deploy-app
  
# Override per environment in ConfigHub:
# staging:
#   environment: staging
#   replicas: 1
#   
# production:
#   environment: production
#   replicas: 3
```

### Dependency Management

Express dependencies between workflows:

```yaml
# ConfigHub tracks dependencies
metadata:
  name: frontend-deploy
  depends_on:
    - api-deploy
    - database-migration
```

### Composing Applications

ConfigHub can orchestrate multiple workers for complex applications:

```yaml
# App composed of multiple configurations
my-app:
  - database:     # Managed by Terraform worker
      type: rds
      size: db.t3.medium
  
  - api-service:  # Managed by Kubernetes worker
      image: myapp/api:v2.0
      replicas: 3
  
  - deployment:   # Managed by Actions bridge
      workflow: deploy-api.yml
      triggers: [database, api-service]
```

### Configuration Best Practices

1. **Separate Concerns**: Different configs for different aspects
2. **Use Spaces**: Dev/staging/prod in separate spaces
3. **Version Everything**: ConfigHub tracks all changes
4. **Link Dependencies**: Express relationships between configs

### Security Considerations

- **Secrets**: Store in ConfigHub, injected at runtime
- **Access Control**: Space-level permissions
- **Audit Trail**: All changes tracked
- **Encryption**: Transit and at-rest

### Testing Configurations

1. **Dry Run**: Use `cub unit plan` to preview changes
2. **Staging Spaces**: Test configs before production
3. **Rollback**: ConfigHub maintains history - use `cub unit rollback`
4. **Validation**: Workers validate before applying

## Troubleshooting

### Unit apply hangs forever

**Problem**: When running `cub unit apply`, the command hangs indefinitely
**Cause**: Worker-target mismatch - your worker isn't the one mapped to the target
**Solution**:
```bash
# 1. Check which worker the target uses
cub target list | grep docker-desktop
# Note the WORKER-SLUG column (e.g., actions-bridge-1)

# 2. Use THAT worker's credentials
cub worker get-envs <worker-from-target-list>
eval "$(cub worker get-envs <worker-from-target-list>)"

# 3. Restart the bridge with correct credentials
pkill actions-bridge
export CONFIGHUB_URL=https://hub.confighub.com
./bin/actions-bridge 2>&1 | tee bridge.log &
```

### "act not found" or workflow execution fails

**Problem**:
```
Error: execution failed: create workflow planner: stat .../workflow.yml: no such file or directory
```
or
```
act: command not found
```

**Solution**:
1. Install act:
   ```bash
   brew install act  # macOS
   # or see other installation methods above
   ```
2. Verify installation:
   ```bash
   act --version
   ```
3. Try running your workflow again

### "Docker daemon not running"

**Problem**: 
```
ERROR: Docker not accessible: Cannot connect to Docker daemon
```

**Solution**:
1. Start Docker Desktop
2. Wait for it to fully start (icon shows "running")
3. Try again

### "Workflow not supported"

**Problem**:
```
ERROR: workflow not supported: Container jobs are not fully supported
```

**Solution**:
Some features aren't supported locally. Check the workflow compatibility:
```bash
# Try creating the unit to see if it's valid
cub unit create --space default test workflow.yml --dry-run
```

### "Permission denied"

**Problem**:
```
ERROR: Permission denied
```

**Solution**:
```bash
# Make the binary executable
chmod +x ~/.confighub/bin/cub

# Or reinstall ConfigHub CLI
curl -fsSL https://hub.confighub.com/cub/install.sh | bash
```

### Workflow runs but does nothing

**Problem**: Workflow completes instantly with no output

**Solution**: Check the workflow trigger. Change:
```yaml
on: pull_request  # Won't trigger locally
```

To:
```yaml
on: push  # Will trigger
# or
on: workflow_dispatch  # Manual trigger
```

## Advanced Tips

### 1. Custom Docker Images

Use specific runner images:
```bash
# Platform is configured in the worker settings
cub worker create --platform linux/arm64 my-arm-worker
cub unit apply --space default myworkflow --worker my-arm-worker
```

### 3. Debugging Workflows

Add debug steps to your workflow:
```yaml
- name: Debug info
  run: |
    echo "Current directory: $(pwd)"
    echo "User: $(whoami)"
    echo "Environment variables:"
    env | sort
```

### 4. Running Specific Jobs

If your workflow has multiple jobs, you can focus on one:
```yaml
name: Multi-job workflow
on: push

jobs:
  test:  # This job will run
    runs-on: ubuntu-latest
    steps:
      - run: echo "Testing..."
  
  deploy:  # This job won't run locally
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - run: echo "Deploying..."
```

## FAQ

### Q: Why is my workflow slow the first time?
**A:** Docker needs to download the runner image. It's cached after the first run.

### Q: Can I use this in CI/CD?
**A:** Yes! The CLI is designed to work in automated environments.

### Q: How do I update?
**A:** Download the latest release or run `git pull && make build` if building from source.

### Q: Can I use custom actions from the marketplace?
**A:** Most actions work, but some that depend on GitHub-specific features won't.

### Q: Is this secure?
**A:** Yes! Secrets are handled securely, workspaces are isolated, and everything is cleaned up after execution.

## Getting Help

If you're stuck:

1. Check the built-in help:
   ```bash
   cub --help
   cub unit --help
   cub worker --help
   ```

2. Validate your workflow:
   ```bash
   cub unit create --space default test workflow.yml --dry-run
   ```

3. Run with verbose logging:
   ```bash
   cub unit apply --space default myworkflow --debug
   ```

4. Check the [README](README.md) for more technical details

5. Report issues at: https://github.com/confighub/actions-bridge/issues

## Next Steps

Now that you understand the basics, explore these resources:

### 1. Browse the Examples
Check out our [17 example workflows](examples/) that demonstrate:
- [Basic workflows](examples/hello-world.yml) - Start here
- [Secret handling](examples/with-secrets.yml) - Secure credential management
- [CI/CD pipelines](examples/build-test-deploy.yml) - Complete deployment flows
- [ConfigHub integration](examples/config-driven-deployment.yml) - Advanced features

### 2. Try Advanced Features
- **Time Travel Testing**: Test workflows with historical configurations
- **Workflow Comparison**: See what changes between versions
- **GitOps Preview**: Preview configuration changes before applying

### 3. Integrate with Your Projects
1. Copy your `.github/workflows` files locally
2. Create a `secrets.env` for your project
3. Test your workflows before pushing
4. Set up ConfigHub for production-grade configuration management

### 4. Learn More
- Read about [ConfigHub integration examples](examples/README.md#confighub-integration-examples)
- Explore the [API Reference](README.md#cli-reference)
- Join the community and contribute!

### Quick Reference Card

```bash
# Essential commands you'll use daily
cub unit create --space default myworkflow workflow.yml --target docker-desktop   # Create workflow unit
cub unit apply --space default myworkflow                                        # Run a workflow
cub unit get --space default myworkflow                                         # Check workflow details
cub unit apply --space default myworkflow --debug                              # Debug with verbose output
cub help                                                                       # See ConfigHub help

# With different spaces
cub unit create --space production webapp deploy.yml --target docker-desktop
cub unit apply --space production webapp
cub unit apply --space production webapp --restore 1      # Use previous revision
```

**Remember:** The goal is to make your CI/CD development faster and more reliable. Start simple, then explore the advanced features as you need them.

Happy workflow testing! 🚀