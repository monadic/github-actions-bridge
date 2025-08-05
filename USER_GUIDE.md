# GitHub Actions Bridge - User Guide

[← Back to README](README.md) | [Examples →](examples/) | [API Reference](README.md#cli-reference)

---

Welcome! This guide will walk you through using the GitHub Actions Bridge step by step.

## Table of Contents
1. [Understanding the Problem](#understanding-the-problem)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Your First Workflow](#your-first-workflow)
5. [Working with Secrets](#working-with-secrets)
6. [ConfigHub Integration](#confighub-integration)
7. [Common Use Cases](#common-use-cases)
8. [Troubleshooting](#troubleshooting)
9. [Next Steps](#next-steps)

## Understanding the Problem

If you've used GitHub Actions, you've probably experienced:
- ❌ Pushing code just to test if your workflow works
- ❌ Waiting 5-10 minutes to see if your change broke CI
- ❌ Debugging through dozens of commit messages like "fix CI attempt #23"
- ❌ Being unable to test workflows that need production secrets

The GitHub Actions Bridge solves these problems by letting you:
- ✅ Test workflows locally before pushing
- ✅ Use real configurations from ConfigHub
- ✅ Debug with instant feedback
- ✅ Test with actual secrets safely

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
cub-actions run deploy.yml --secrets-file secrets.env
```

**Security Notes:**
- Secrets are never logged or displayed
- Files are created with restricted permissions (0600)
- Secrets are cleaned up after execution
- Add `secrets.env` to `.gitignore`

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
export CONFIGHUB_URL=https://api.confighub.com
```

3. **Run workflows with ConfigHub**:
```bash
# Use configurations from ConfigHub
cub-actions run deploy.yml --space production --unit webapp

# Test with different environments
cub-actions run deploy.yml --space staging --unit webapp
cub-actions run deploy.yml --space development --unit webapp
```

### Advanced ConfigHub Features

**Time Travel Testing:**
```bash
# Test with last week's configuration
cub-actions run deploy.yml --space prod --as-of "2024-01-01"
```

**Configuration-Driven Workflows:**
```bash
# All values come from ConfigHub
cub-actions run examples/config-driven-deployment.yml \
  --space production \
  --unit webapp
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

### 1. Custom Docker Images

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

Now that you understand the basics, explore these resources:

### 1. Browse the Examples
Check out our [15+ example workflows](examples/) that demonstrate:
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
cub-actions run workflow.yml              # Run a workflow
cub-actions validate workflow.yml         # Check compatibility
cub-actions run workflow.yml --dry-run    # Preview execution
cub-actions run workflow.yml -v           # Debug with verbose output
cub-actions list-limitations              # See what's not supported

# With configurations
cub-actions run workflow.yml --secrets-file secrets.env
cub-actions run workflow.yml --space production --unit webapp
cub-actions run workflow.yml --as-of "2024-01-01"
```

**Remember:** The goal is to make your CI/CD development faster and more reliable. Start simple, then explore the advanced features as you need them.

Happy workflow testing! 🚀