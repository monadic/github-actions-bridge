# GitHub Actions Bridge - User Guide for Beginners

Welcome! This guide will help you get started with the GitHub Actions Bridge, even if you're new to GitHub Actions or ConfigHub.

## Table of Contents
1. [What is this?](#what-is-this)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Your First Workflow](#your-first-workflow)
5. [Common Use Cases](#common-use-cases)
6. [Troubleshooting](#troubleshooting)
7. [FAQ](#faq)

## What is this?

The GitHub Actions Bridge lets you run GitHub Actions workflows on your local machine without pushing to GitHub. This is useful for:
- Testing workflows before committing
- Running workflows in environments without GitHub access
- Integrating GitHub Actions with ConfigHub

## Prerequisites

Before you start, you'll need:

1. **Docker Desktop** installed and running
   - Mac: Download from [docker.com](https://www.docker.com/products/docker-desktop)
   - Linux: Install Docker Engine
   - Windows: Use WSL2 with Docker Desktop

2. **Go 1.21+** (only if building from source)
   - Check with: `go version`

3. **Basic command line knowledge**
   - How to open a terminal
   - How to navigate directories (`cd`)
   - How to run commands

## Installation

### Option 1: Download Pre-built Binaries (Easiest)

```bash
# Download the latest release (example for Mac)
curl -L https://github.com/confighub/actions-bridge/releases/latest/download/cub-actions-darwin-arm64 -o cub-actions
chmod +x cub-actions
sudo mv cub-actions /usr/local/bin/

# Test it works
cub-actions version
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/confighub/actions-bridge
cd actions-bridge

# Build everything
make build

# The binaries will be in ./bin/
./bin/cub-actions version
```

## Your First Workflow

Let's run a simple workflow to make sure everything works!

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
cub-actions run hello.yml
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
# See what would happen without running
cub-actions run hello.yml --dry-run

# See detailed logs
cub-actions run hello.yml --verbose
```

## Common Use Cases

### 1. Running with Secrets

Many workflows need secrets (like API keys). Here's how to provide them:

Create a `secrets.env` file:
```bash
API_KEY=your-api-key-here
DATABASE_PASSWORD=super-secret
```

Create a workflow that uses secrets (`with-secrets.yml`):
```yaml
name: Secret Test
on: push

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Use API Key
        env:
          API_KEY: ${{ secrets.API_KEY }}
        run: |
          echo "API Key length: ${#API_KEY}"
          # Never echo the actual secret!
```

Run it:
```bash
cub-actions run with-secrets.yml --secrets-file secrets.env
```

### 2. Running with Environment Variables

Create an environment file (`dev.env`):
```bash
ENVIRONMENT=development
API_URL=https://api-dev.example.com
DEBUG=true
```

Run with environment:
```bash
cub-actions run workflow.yml --env-file dev.env
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
cub-actions run build.yml --artifact-dir ./my-artifacts
```

### 4. Validating Workflows

Before running, check if a workflow will work:

```bash
cub-actions validate workflow.yml
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
cub-actions list-limitations
```

## Troubleshooting

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
Some features aren't supported locally. Check the limitations with:
```bash
cub-actions validate workflow.yml
```

### "Permission denied"

**Problem**:
```
ERROR: Permission denied
```

**Solution**:
```bash
# Make the binary executable
chmod +x cub-actions

# Or use sudo if needed
sudo cub-actions run workflow.yml
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

### 1. Using ConfigHub Integration

If you have ConfigHub access:

```bash
# Set up worker credentials
export CONFIGHUB_WORKER_ID=your-worker-id
export CONFIGHUB_WORKER_SECRET=your-secret

# Run the bridge worker
./bin/actions-bridge
```

### 2. Custom Docker Images

Use specific runner images:
```bash
cub-actions run workflow.yml --platform linux/arm64
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
   cub-actions --help
   cub-actions run --help
   ```

2. Validate your workflow:
   ```bash
   cub-actions validate workflow.yml
   ```

3. Run with verbose logging:
   ```bash
   cub-actions run workflow.yml --verbose
   ```

4. Check the [README](README.md) for more technical details

5. Report issues at: https://github.com/confighub/actions-bridge/issues

## Next Steps

Now that you've got the basics:

1. Try running your own workflows
2. Experiment with secrets and environment variables
3. Set up the bridge worker for ConfigHub integration
4. Contribute improvements to the project!

Happy workflow running!