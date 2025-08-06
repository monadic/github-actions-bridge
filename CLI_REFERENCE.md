# CLI Reference - cub-worker-actions

Complete command-line interface documentation for the GitHub Actions Bridge worker CLI.

## Installation

```bash
# Build from source
make build-worker

# Binary will be available at
./bin/cub-worker-actions
```

## Global Options

These options work with all commands:

- `-v, --verbose` - Enable verbose output for debugging
- `-h, --help` - Show help for any command
- `--version` - Show version information

## Commands

### `run` - Execute a workflow

Run a GitHub Actions workflow locally using act.

```bash
cub-worker-actions run WORKFLOW [flags]
```

**Arguments:**
- `WORKFLOW` - Path to the workflow YAML file

**Flags:**
- `--artifact-dir string` - Directory to save artifacts
- `--dry-run` - Show what would be executed without running
- `--env-file string` - Environment file to load (.env format)
- `--event string` - GitHub event type to simulate (default: "workflow_dispatch")
- `-i, --input strings` - Workflow inputs (key=value format, can be specified multiple times)
- `--platform string` - Execution platform (default: "linux/amd64")
- `--secrets-file string` - Secrets file to load (.env format)
- `--space string` - ConfigHub space
- `--timeout int` - Execution timeout in seconds (default: 3600)
- `--unit string` - ConfigHub unit
- `--validate` - Validate workflow without running
- `--watch` - Watch workflow file for changes and re-run

**Examples:**

```bash
# Basic execution
cub-worker-actions run examples/hello-world.yml

# With secrets
cub-worker-actions run examples/with-secrets.yml --secrets-file secrets.env

# With inputs
cub-worker-actions run examples/build.yml -i version=1.2.3 -i environment=prod

# Dry run to see what would execute
cub-worker-actions run examples/deploy.yml --dry-run

# Watch mode for development
cub-worker-actions run examples/test.yml --watch

# With custom timeout
cub-worker-actions run examples/long-running.yml --timeout 7200
```

### `validate` - Validate a workflow

Check if a workflow is valid and can be executed locally with act.

```bash
cub-worker-actions validate WORKFLOW [flags]
```

**Arguments:**
- `WORKFLOW` - Path to the workflow YAML file to validate

**Examples:**

```bash
# Validate a single workflow
cub-worker-actions validate examples/hello-world.yml

# Validate with verbose output
cub-worker-actions validate examples/complex-workflow.yml -v
```

**Output:**
- ✓ Workflow is valid
- Compatibility warnings (if any)
- Syntax errors (if any)

### `list-limitations` - Show known limitations

Display all known limitations when running GitHub Actions locally with act.

```bash
cub-worker-actions list-limitations
```

**Output:**
Lists limitations such as:
- Unsupported GitHub Actions
- Limited GitHub API access
- Container limitations
- Network restrictions

### `clean` - Clean up resources

Clean up temporary files and Docker resources created by the bridge.

```bash
cub-worker-actions clean [flags]
```

**Flags:**
- `--all` - Clean all resources including Docker containers
- `--dry-run` - Show what would be cleaned without doing it

**Examples:**

```bash
# Clean temporary files
cub-worker-actions clean

# Clean everything including Docker containers
cub-worker-actions clean --all

# Preview what would be cleaned
cub-worker-actions clean --dry-run
```

### `version` - Show version

Display version information for the CLI and its components.

```bash
cub-worker-actions version
```

**Output:**
```
GitHub Actions Bridge Worker v0.1.0
Built with:
  - act v0.2.80
  - Go 1.24.3
```

### `completion` - Generate shell completions

Generate shell completion scripts for various shells.

```bash
cub-worker-actions completion [bash|zsh|fish|powershell]
```

**Examples:**

```bash
# Bash
cub-worker-actions completion bash > ~/.bash_completion.d/cub-worker-actions

# Zsh
cub-worker-actions completion zsh > ~/.zsh/completions/_cub-worker-actions

# Fish
cub-worker-actions completion fish > ~/.config/fish/completions/cub-worker-actions.fish
```

## Environment Variables

The CLI respects these environment variables:

- `LOG_LEVEL` - Set logging level (debug, info, warn, error)
- `NO_COLOR` - Disable colored output
- `DOCKER_HOST` - Docker daemon socket
- `CONFIGHUB_WORKER_ID` - ConfigHub worker ID
- `CONFIGHUB_WORKER_SECRET` - ConfigHub worker secret
- `CONFIGHUB_URL` - ConfigHub API URL

## Configuration Files

### Secrets File Format (.env)

```bash
# secrets.env
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=sk_test_1234567890
GITHUB_TOKEN=ghp_1234567890
```

### Environment File Format (.env)

```bash
# environment.env
NODE_ENV=development
DEBUG=true
PORT=3000
```

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Invalid arguments
- `3` - Workflow validation failed
- `4` - Execution failed
- `5` - Timeout exceeded

## ConfigHub Integration

When running as a ConfigHub worker:

```bash
# Set worker credentials
export CONFIGHUB_WORKER_ID=worker-123
export CONFIGHUB_WORKER_SECRET=secret-key
export CONFIGHUB_URL=https://api.confighub.com

# Run as worker
./bin/actions-bridge
```

## Workflow Format

Workflows must use the ConfigHub Actions format:

```yaml
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: my-workflow
# Standard GitHub Actions workflow follows
name: My Workflow
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Hello"
```

The CLI automatically strips the first 4 lines of metadata.

## Common Use Cases

### Development Workflow

```bash
# Validate your workflow
cub-worker-actions validate my-workflow.yml

# Run with watch mode for rapid iteration
cub-worker-actions run my-workflow.yml --watch

# Clean up when done
cub-worker-actions clean
```

### CI/CD Pipeline Testing

```bash
# Test with production-like secrets
cub-worker-actions run deploy.yml \
  --secrets-file prod-secrets.env \
  --env-file prod-env.env \
  -i environment=production \
  -i version=$VERSION
```

### Debugging

```bash
# Verbose output for troubleshooting
cub-worker-actions run problematic-workflow.yml -v

# Dry run to see execution plan
cub-worker-actions run complex-workflow.yml --dry-run
```

## Related Documentation

- [User Guide](USER_GUIDE.md) - Step-by-step tutorials
- [Examples](examples/README.md) - Working workflow examples
- [YAML Formats](YAML_FORMATS.md) - Workflow format specifications
- [Enterprise Features](ENTERPRISE_FEATURES.md) - ConfigHub integration

## Getting Help

```bash
# General help
cub-worker-actions --help

# Command-specific help
cub-worker-actions run --help

# List known limitations
cub-worker-actions list-limitations
```

For bugs or feature requests, visit: https://github.com/confighub/actions-bridge/issues