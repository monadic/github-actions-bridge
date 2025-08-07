# YAML Toolchain Formats Documentation

This document explains the YAML formats used in the GitHub Actions Bridge project.

## Format Types

### 1. ConfigHub Actions Format

All workflow examples in the `examples/` directory use the ConfigHub Actions format:

```yaml
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: workflow-name
# Standard GitHub Actions workflow follows
name: Workflow Name
on: [push, workflow_dispatch]
jobs:
  job-name:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Hello"
```

**Key Points:**
- First 4 lines are ConfigHub metadata (Kubernetes-style)
- Bridge automatically strips these lines before passing to act
- After line 4, it's standard GitHub Actions YAML
- All examples include this header for ConfigHub compatibility

### 2. GitHub Actions Format

After metadata stripping, workflows follow standard GitHub Actions format:

```yaml
name: Workflow Name
on: [push, workflow_dispatch]
jobs:
  job-name:
    runs-on: ubuntu-latest
    steps:
      - name: Step name
        run: command
```

**Validation**: All 17 examples validate correctly as GitHub Actions workflows

### 3. Docker Compose Format

Two docker-compose files with different security profiles:

**docker-compose.yml** (Development):
```yaml
version: '3.8'  # Note: version field is now optional
services:
  actions-bridge:
    image: confighub/actions-bridge:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock  # Security warning added
```

**docker-compose.secure.yml** (Production):
```yaml
version: '3.8'
services:
  actions-bridge:
    # Uses Docker-in-Docker instead of socket mounting
  docker:
    image: docker:24-dind
    privileged: true  # Required for DinD
```

Both files validate correctly with docker-compose.

### 4. Prometheus Configuration

Standard Prometheus format:

```yaml
global:
  scrape_interval: 15s
scrape_configs:
  - job_name: 'actions-bridge'
    static_configs:
      - targets: ['actions-bridge:8080']
    metrics_path: '/metrics'
```

## Validation Status

✅ **All YAML files are valid**

| File Type | Count | Status | Notes |
|-----------|-------|--------|-------|
| Workflow Examples | 21 | ✅ Valid | All include ConfigHub headers |
| Docker Compose | 2 | ✅ Valid | Warnings about version field (cosmetic) |
| Prometheus | 1 | ✅ Valid | Standard Prometheus config |
| Test Fixtures | 3 | ✅ Valid | Simple workflows for testing |

## Common Patterns

### Workflow Triggers
```yaml
on: [push, workflow_dispatch]  # Most common
on: push                       # Simple
on:                           # Detailed
  push:
    branches: [main]
  workflow_dispatch:
```

### Runner Specifications
```yaml
runs-on: ubuntu-latest    # Default for all examples
runs-on: ubuntu-22.04     # Specific version
runs-on: ${{ matrix.os }} # Matrix builds
```

### Secret Usage
```yaml
env:
  API_KEY: ${{ secrets.API_KEY }}
with:
  token: ${{ secrets.GITHUB_TOKEN }}  # Simulated locally
```

## Toolchain Compatibility

1. **ConfigHub CLI** (`cub`): Expects ConfigHub metadata
2. **Bridge CLI** (`cub-local-actions`): Strips metadata automatically
3. **act**: Requires standard GitHub Actions format (after stripping)
4. **Docker Compose**: Both files compatible with v2/v3

## Best Practices

1. **Always include ConfigHub headers** in examples
2. **Validate with** `cub-local-actions validate <file>`
3. **Test locally with** `cub-local-actions run <file>`
4. **Use docker-compose.secure.yml** for production

## Related Documentation

- [Examples README](examples/README.md) - Detailed example descriptions
- [User Guide](USER_GUIDE.md) - How to write workflows
- [SDK Validation](SDK_VALIDATION.md) - ConfigHub SDK details

All YAML formats in this project are valid and properly structured for their respective toolchains.