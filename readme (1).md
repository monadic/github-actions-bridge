# GitHub Actions Bridge for ConfigHub

A ConfigHub bridge that enables local execution of GitHub Actions workflows using `act`, with proper workspace isolation, secret handling, and full integration with ConfigHub's configuration management.

## 🚀 Quick Start

If you're familiar with `act` or ConfigHub, here's the fastest way to get started:

```bash
# Install the bridge
go install github.com/confighub/actions-bridge/cmd/actions-bridge@latest
go install github.com/confighub/actions-bridge/cmd/actions-cli@latest

# Set up your worker credentials
export CONFIGHUB_WORKER_ID=your-worker-id
export CONFIGHUB_WORKER_SECRET=your-worker-secret

# Run the bridge
actions-bridge

# In another terminal, run a workflow
cub-actions run .github/workflows/deploy.yml \
  --space staging \
  --unit webapp \
  --input version=1.2.3
```

## 📋 Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [CLI Reference](#cli-reference)
- [Workflow Examples](#workflow-examples)
- [Act Compatibility](#act-compatibility)
- [Architecture](#architecture)
- [Troubleshooting](#troubleshooting)

## Overview

The GitHub Actions Bridge brings the power of GitHub Actions to ConfigHub, allowing you to:

- **Test workflows locally** with production configurations before committing
- **Use ConfigHub as your secret provider** instead of GitHub Secrets
- **Trigger workflows based on configuration changes**
- **Time-travel test** workflows with historical configurations

### How It Works

```
ConfigHub (configs/secrets) → Actions Bridge → act → Docker → Your Workflow
```

The bridge acts as an adapter between ConfigHub's configuration management and `act`'s local GitHub Actions runner, providing:

1. **Configuration Injection**: Automatically injects ConfigHub configurations into workflows
2. **Secret Management**: Securely provides secrets from ConfigHub without exposing them
3. **Workspace Isolation**: Each execution runs in an isolated environment
4. **Compatibility Layer**: Handles act limitations transparently

## Key Features

### 🔐 Security First
- Secrets are never exposed in environment variables
- Secure workspace cleanup with file overwriting
- Leak detection prevents secrets in logs
- Isolated execution environments

### 🎯 ConfigHub Integration
- Direct access to ConfigHub spaces and units
- Configuration-driven workflow execution
- Automatic config and secret injection
- Full audit trail of executions

### 🛠️ Developer Experience
- Enhanced CLI with helpful commands
- Compatibility warnings before execution
- Dry-run mode for testing
- Health monitoring and metrics

### 📊 Production Ready
- Prometheus metrics integration
- Health check endpoints
- Graceful shutdown handling
- Resource cleanup automation

## Installation

### Prerequisites

- Go 1.21 or later
- Docker (for act)
- ConfigHub CLI (`cub`)
- A ConfigHub worker ID and secret

### From Source

```bash
# Clone the repository
git clone https://github.com/confighub/actions-bridge.git
cd actions-bridge

# Build everything
make build

# Run tests (including act validation)
make test
make act-test

# Install binaries
make install
```

### Using Docker

```bash
# Pull the image
docker pull confighub/actions-bridge:latest

# Run with your credentials
docker run -d \
  --name actions-bridge \
  -e CONFIGHUB_WORKER_ID=your-worker-id \
  -e CONFIGHUB_WORKER_SECRET=your-worker-secret \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -p 8080:8080 \
  confighub/actions-bridge
```

### Docker Compose

```yaml
version: '3.8'
services:
  actions-bridge:
    image: confighub/actions-bridge:latest
    environment:
      - CONFIGHUB_WORKER_ID=${CONFIGHUB_WORKER_ID}
      - CONFIGHUB_WORKER_SECRET=${CONFIGHUB_WORKER_SECRET}
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./workspace:/var/lib/actions-bridge
    ports:
      - "8080:8080"
    restart: unless-stopped
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIGHUB_WORKER_ID` | Your ConfigHub worker ID | Required |
| `CONFIGHUB_WORKER_SECRET` | Your ConfigHub worker secret | Required |
| `CONFIGHUB_URL` | ConfigHub API URL | `https://api.confighub.com` |
| `ACTIONS_BRIDGE_BASE_DIR` | Base directory for workspaces | `/var/lib/actions-bridge` |
| `ACT_DEFAULT_IMAGE` | Default Docker image for act | `catthehacker/ubuntu:act-latest` |
| `MAX_CONCURRENT_WORKFLOWS` | Maximum concurrent executions | `5` |
| `DEBUG` | Enable debug logging | `false` |

### Configuration File

Create `/etc/actions-bridge/config.yaml`:

```yaml
worker:
  id: ${CONFIGHUB_WORKER_ID}
  secret: ${CONFIGHUB_WORKER_SECRET}
  url: https://api.confighub.com

bridge:
  baseDir: /var/lib/actions-bridge
  maxConcurrent: 5
  workspaceTimeout: 1h
  
act:
  defaultImage: catthehacker/ubuntu:act-latest
  platforms:
    ubuntu-latest: catthehacker/ubuntu:act-latest
    ubuntu-22.04: catthehacker/ubuntu:act-22.04
    ubuntu-20.04: catthehacker/ubuntu:act-20.04
  cacheDir: /tmp/act-cache
  verbose: true

security:
  secretsAsFiles: true
  leakDetection: true
  secureCleanup: true
```

## Usage

### Basic Workflow Execution

1. **Create a workflow** (`.github/workflows/deploy.yml`):

```yaml
name: Deploy Application
on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Target environment'
        required: true
        type: choice
        options:
          - staging
          - production

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Use ConfigHub configs
        run: |
          echo "Deploying ${{ env.CONFIG_APP_NAME }}"
          echo "Version: ${{ env.CONFIG_VERSION }}"
          echo "Replicas: ${{ env.CONFIG_REPLICAS }}"
      
      - name: Use ConfigHub secrets
        env:
          API_KEY: ${{ secrets.API_KEY }}
          DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
        run: |
          echo "Connecting to database..."
          # Your deployment logic here
```

2. **Run with ConfigHub**:

```bash
# Dry run first
cub-actions run .github/workflows/deploy.yml \
  --space staging \
  --unit webapp \
  --dry-run

# Execute for real
cub-actions run .github/workflows/deploy.yml \
  --space staging \
  --unit webapp \
  --input environment=staging
```

### Using ConfigHub Configurations

The bridge automatically injects configurations in multiple formats:

```yaml
steps:
  - name: Access configs as environment variables
    run: |
      echo "App: $CONFIG_APP_NAME"
      echo "Port: $CONFIG_PORT"
      echo "Replicas: $CONFIG_REPLICAS"
  
  - name: Access configs as JSON files
    run: |
      cat ${{ github.workspace }}/configs/config.json
      jq '.database.host' ${{ github.workspace }}/configs/database.json
  
  - name: Access nested configs
    run: |
      echo "DB Host: $CONFIG_DATABASE_HOST"
      echo "DB Port: $CONFIG_DATABASE_PORT"
```

### Secret Management

Secrets are injected securely:

```yaml
steps:
  - name: Use secrets safely
    env:
      API_KEY: ${{ secrets.API_KEY }}
    run: |
      # Secret is available but never logged
      curl -H "Authorization: Bearer $API_KEY" https://api.example.com
```

### Advanced Usage

#### Time-Travel Testing

Test how your workflow would behave with past configurations:

```bash
# Test with last week's configuration
cub-actions run deploy.yml \
  --space production \
  --unit webapp \
  --revision "@{1 week ago}" \
  --dry-run
```

#### Configuration-Triggered Workflows

Set up workflows that run when configurations change:

```yaml
# In ConfigHub unit configuration
apiVersion: v1
kind: WorkflowTrigger
metadata:
  name: auto-deploy
spec:
  watch:
    - space: staging
      units: ["webapp", "api"]
  on:
    - event: config.changed
      run: ".github/workflows/test.yml"
    - event: config.approved
      run: ".github/workflows/deploy.yml"
```

## CLI Reference

### Global Flags

- `--space` - ConfigHub space to use
- `--unit` - ConfigHub unit to use
- `--dry-run` - Show what would happen without executing
- `--verbose, -v` - Enable verbose output
- `--config` - Path to configuration file

### Commands

#### `run` - Execute a workflow

```bash
cub-actions run WORKFLOW [flags]

Flags:
  --event string        GitHub event type (default "workflow_dispatch")
  -i, --input strings   Workflow inputs (key=value)
  --platform string     Execution platform (default "linux/amd64")
  --artifact-dir string Directory for artifacts (default "./artifacts")
  --env-file string     Additional environment file
  --secrets-file string Secrets file (KEY=value format)
  --validate           Validate workflow without running
```

#### `validate` - Validate a workflow

```bash
cub-actions validate WORKFLOW

# Example
cub-actions validate .github/workflows/deploy.yml
```

#### `compat` - Show compatibility information

```bash
cub-actions compat

# Shows known act limitations and workarounds
```

#### `list` - List available workflows

```bash
cub-actions list [DIRECTORY]

# Example output
WORKFLOW                          TRIGGERS         LAST MODIFIED
.github/workflows/ci.yml         push, pr         2024-01-15 10:30
.github/workflows/deploy.yml     push (main)      2024-01-14 15:45
```

## Workflow Examples

### Simple Test Workflow

```yaml
name: Run Tests
on: [push, workflow_dispatch]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Run unit tests
        run: |
          echo "Running tests for ${{ env.CONFIG_APP_NAME }}"
          # Your test commands here
```

### Deployment with Approval

```yaml
name: Production Deploy
on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to deploy'
        required: true

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: production
    steps:
      - name: Deploy application
        env:
          KUBECONFIG: ${{ secrets.KUBECONFIG }}
        run: |
          kubectl set image deployment/webapp \
            app=webapp:${{ github.event.inputs.version }}
```

### Multi-Environment Workflow

```yaml
name: Multi-Environment Deploy
on:
  workflow_dispatch:

jobs:
  deploy:
    strategy:
      matrix:
        environment: [dev, staging, prod]
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to ${{ matrix.environment }}
        run: |
          echo "Deploying to ${{ matrix.environment }}"
          # ConfigHub will inject the right configs per environment
```

## Act Compatibility

The bridge handles known act limitations transparently:

### Fully Supported
- ✅ Basic workflow execution
- ✅ Docker container actions
- ✅ Composite actions
- ✅ Most JavaScript actions
- ✅ Workflow inputs and outputs
- ✅ Matrix strategies
- ✅ Job dependencies

### Limited Support
- ⚠️ Caching (local only)
- ⚠️ Artifacts (local directory only)
- ⚠️ GitHub API calls (mocked)
- ⚠️ Services (requires Docker)

### Not Supported
- ❌ Self-hosted runners
- ❌ GitHub Environments
- ❌ GitHub Packages
- ❌ Cross-workflow artifacts
- ❌ Concurrency groups

### Handling Limitations

The bridge provides warnings for compatibility issues:

```bash
$ cub-actions run workflow.yml --space prod --unit webapp

Compatibility warnings:
  [INFO] Line 15: Caching not supported locally - workflow will run without cache
  [WARNING] Line 23: GITHUB_TOKEN will be simulated locally with limited permissions
  [ERROR] Line 30: Self-hosted runners not supported in local execution

Suggestions:
  • Replace 'self-hosted' runners with 'ubuntu-latest' for local execution
  • Artifacts will be saved to the workspace output directory
```

## Architecture

### Component Overview

```
┌─────────────────────────────────┐
│   GitHub Actions Bridge Worker   │
│  ┌────────────────────────────┐  │
│  │   Bridge Interface         │  │ - Info, Apply, Refresh, Destroy, Import, Finalize
│  ├────────────────────────────┤  │
│  │   Workspace Manager        │  │ - Isolated execution environments
│  │                            │  │ - Secure cleanup with file overwriting
│  ├────────────────────────────┤  │
│  │   Act Wrapper              │  │ - act integration
│  │                            │  │ - Compatibility handling
│  ├────────────────────────────┤  │
│  │   Config Injector          │  │ - Multiple format support
│  │                            │  │ - Nested config handling
│  ├────────────────────────────┤  │
│  │   Secret Handler           │  │ - File-based secrets
│  │                            │  │ - Leak detection
│  └────────────────────────────┘  │
└─────────────────────────────────┘
```

### Execution Flow

1. **Request received** from ConfigHub
2. **Workspace created** with isolation
3. **Compatibility check** on workflow
4. **Configs injected** as files and env vars
5. **Secrets prepared** in secure files
6. **act executed** with proper context
7. **Results collected** and reported
8. **Workspace cleaned** securely

### Security Model

- **Workspace Isolation**: Each execution gets a unique workspace
- **Secret Files**: Secrets never in environment variables
- **Secure Cleanup**: Files overwritten before deletion
- **Leak Detection**: Automatic secret masking in logs
- **Audit Trail**: All operations logged

## Troubleshooting

### Common Issues

#### Docker not accessible
```bash
# Check Docker is running
docker info

# Ensure Docker socket is accessible
ls -la /var/run/docker.sock

# Add user to docker group
sudo usermod -aG docker $USER
```

#### Workflow fails with "command not found"
```yaml
# Ensure you're using a full Ubuntu image
runs-on: ubuntu-latest  # Uses catthehacker/ubuntu:act-latest
```

#### Secrets not available
```bash
# Check secrets are defined in ConfigHub
cub secret list --space staging

# Verify secret injection
cub-actions run workflow.yml --space staging --verbose
```

#### Disk space issues
```bash
# Check available space
df -h /var/lib/actions-bridge

# Clean up old workspaces
docker system prune -a
```

### Debug Mode

Enable debug logging:

```bash
# Via environment variable
export DEBUG=true
actions-bridge

# Via CLI flag
cub-actions run workflow.yml --verbose
```

### Health Checks

Monitor bridge health:

```bash
# Check overall health
curl http://localhost:8080/health

# Check readiness
curl http://localhost:8080/ready

# View metrics
curl http://localhost:8080/metrics
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Install development tools
make dev-setup

# Run tests
make test

# Lint code
make lint

# Format code
make fmt
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- The [act](https://github.com/nektos/act) project for local GitHub Actions execution
- ConfigHub team for the configuration management platform
- Contributors who helped identify and fix act compatibility issues

---

*Built with ❤️ by the ConfigHub team*