# Examples Compatibility Guide

This table shows which tool to use for each example workflow in the `examples/` directory.

## Tool Overview

- **`cub`** - ConfigHub CLI for production workflows managed through ConfigHub
- **`cub-local-actions`** - Local CLI for development and testing without ConfigHub

## Compatibility Table

| Example File | Works with `cub` | Works with `cub-local-actions` | Notes |
|--------------|------------------|--------------------------------|-------|
| `hello-world.yml` | ✅ | ✅ | Simple example works with both |
| `environment-variables.yml` | ✅ | ✅ | Both tools support environment variables |
| `with-secrets.yml` | ✅ | ✅ | ConfigHub manages secrets centrally, local uses file |
| `build-test-deploy.yml` | ✅ | ✅ | Standard CI/CD workflow |
| `multi-job.yml` | ✅ | ✅ | Job dependencies work in both |
| `conditional-execution.yml` | ✅ | ✅ | Conditions evaluated locally |
| `matrix-builds.yml` | ✅ | ✅ | Matrix strategy supported |
| `docker-compose-improved.yml` | ✅ | ✅ | Docker commands work locally |
| `config-driven-deployment.yml` | ⚠️ | ❌ | **Partially simulated** - Basic ConfigHub works, advanced features mocked |
| `config-triggered-workflow.yml` | ✅ | ❌ | **Requires ConfigHub** triggers |
| `time-travel-testing.yml` | 🚧 | ❌ | **Simulated** - Uses date comparisons, not real versioning |
| `claude-orchestrated-ops.yml` | 🚧 | ❌ | **Simulated** - Claude responses are mocked |
| `worker-calls-claude.yml` | 🚧 | ❌ | **Simulated** - Claude API calls are mocked |
| `workflow-diff-testing.yml` | ✅ | ❌ | **Requires ConfigHub** for comparisons |
| `artifact-handling-improved.yml` | ⚠️ | ✅ | Limited in ConfigHub, full support locally |
| `file-persistence-improved.yml` | ⚠️ | ✅ | Better local file handling |
| `gitops-preview-improved.yml` | ⚠️ | ✅ | Git operations work better locally |

## Legend

- ✅ **Full Support** - Example works as intended
- ❌ **Not Supported** - Example requires features only available in that tool
- ⚠️ **Limited Support** - Example works but with limitations
- 🚧 **Simulated** - Example demonstrates concept but uses mocked functionality

## Usage Guidelines

### Use `cub` (ConfigHub) when:
- Managing production workflows
- Need centralized configuration management
- Working with teams
- Using ConfigHub-specific features (triggers, time travel, config-driven deployment)

### Use `cub-local-actions` when:
- Developing and testing workflows locally
- Quick iteration during development
- Working with local files and secrets
- Don't need ConfigHub features

## Example Commands

### ConfigHub Workflow (`cub`)
```bash
# Create and apply a workflow through ConfigHub
cub unit create --space production hello hello-world.yml
cub unit apply --space production hello
```

### Local Development (`cub-local-actions`)
```bash
# Run a workflow locally
./bin/cub-local-actions run examples/hello-world.yml

# With secrets file
./bin/cub-local-actions run examples/with-secrets.yml --secrets-file secrets.env
```

## Important Notes About Examples

### Simulated Examples
The following examples demonstrate **concepts** but use mocked functionality:
- **claude-orchestrated-ops.yml** - Simulates Claude AI responses with shell scripts
- **worker-calls-claude.yml** - Mocks Claude API calls with conditional logic
- **time-travel-testing.yml** - Uses date comparisons instead of real revision history

### ConfigHub-Only Examples
These examples require ConfigHub but have varying levels of implementation:
- **config-driven-deployment.yml** - Basic ConfigHub integration works, advanced features are simulated
- **config-triggered-workflow.yml** - Requires ConfigHub triggers
- **workflow-diff-testing.yml** - Requires ConfigHub for comparisons

### Fully Functional Examples
Most other examples (hello-world, build-test-deploy, multi-job, etc.) work as documented with both tools.