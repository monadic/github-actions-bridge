# GitHub Actions Bridge Configuration

## Project Overview

GitHub Actions Bridge enables running GitHub Actions workflows through ConfigHub's configuration management platform. It bridges `nektos/act` (local GitHub Actions runner) with ConfigHub's worker protocol, allowing workflows to be managed as configuration units with centralized versioning and deployment.

## Build & Commands

### Development
```bash
make build          # Build binaries
make test           # Run tests  
make docker         # Build Docker image
go mod tidy         # Update dependencies
```

### Testing
```bash
go test ./...                                    # Run all tests
SKIP_ACT_TESTS=1 go test -v ./pkg/bridge/      # Skip act integration tests
./bin/act-test                                   # Test act integration
```

### Running
```bash
# Bridge Worker
export CONFIGHUB_WORKER_ID=xxx
export CONFIGHUB_WORKER_SECRET=xxx
export CONFIGHUB_URL=https://hub.confighub.com
./bin/actions-bridge

# CLI Tool
./bin/cub-local-actions run examples/hello-world.yml
./bin/cub-local-actions validate examples/build.yml
```

## Code Style

### Go Standards
- Follow standard Go formatting (`gofmt`)
- Use Go modules for dependency management
- Keep functions focused and testable
- Handle errors explicitly, don't panic
- Use context.Context for cancellation

### Documentation
- Document exported functions and types
- Include examples in complex functions
- Keep comments concise and meaningful
- Update README.md when adding features

### Testing
- Write table-driven tests where applicable
- Mock external dependencies (Docker, ConfigHub)
- Test error conditions thoroughly
- Use testify/assert for assertions

## Testing Approach

### Unit Tests
- Test individual components in isolation
- Mock Docker client and ConfigHub SDK
- Focus on business logic validation

### Integration Tests
- Test act runner integration
- Verify workflow execution
- Check workspace management

### Skip Flags
- `SKIP_ACT_TESTS=1` - Skip tests requiring act/Docker
- Useful for CI environments without Docker

## Architecture

### Technology Stack
- **Language**: Go 1.23
- **CLI Framework**: spf13/cobra v1.9.1
- **Metrics**: prometheus/client_golang v1.22.0
- **ConfigHub SDK**: v0.0.0-20250804044729-f1517379cea0
- **GitHub Actions Runner**: nektos/act v0.2.80

### Core Components
- **Bridge Worker**: Implements ConfigHub worker protocol
- **CLI Tool**: Local workflow execution and validation
- **Act Integration**: Wraps nektos/act for workflow execution
- **Workspace Manager**: Isolated temporary directories per run

### Key Interfaces
```go
// Worker interface (from ConfigHub SDK)
type Worker interface {
    Info() (*WorkerInfo, error)
    Apply(unit *Unit) (*ApplyResult, error)
    Refresh(unit *Unit) (*RefreshResult, error)
    Destroy(unit *Unit) (*DestroyResult, error)
}
```

## Security Considerations

### Docker Socket Access
- Bridge requires Docker socket access
- Use docker-compose.secure.yml for production
- Consider rootless Docker for enhanced security

### Secrets Management
- Never log secrets or sensitive data
- Use environment variables for credentials
- Support .env files for local development
- Sanitize workflow outputs

### Workspace Isolation
- Each workflow runs in isolated /tmp directory
- Clean up workspaces after execution
- Prevent cross-workflow contamination

## Git Workflow

### Branching
- Main branch: `main`
- Feature branches: `feature/description`
- Fix branches: `fix/description`

### Commits
- Use conventional commit format
- Keep commits focused and atomic
- Reference issues when applicable

### Testing Requirements
- All PRs must pass CI tests
- New features require tests
- Update documentation with changes

## Configuration Management

### Environment Variables
```bash
CONFIGHUB_WORKER_ID       # Worker authentication
CONFIGHUB_WORKER_SECRET   # Worker secret
CONFIGHUB_URL            # ConfigHub API URL
CONFIGHUB_SPACE          # Target space
```

### Workflow Format
```yaml
# Required ConfigHub header
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: workflow-name

# Standard GitHub Actions workflow
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
```

### Common Issues
- Wrong ConfigHub URL (use hub.confighub.com, not api)
- Missing `--target docker-desktop` on unit create
- cub CLI not in PATH (~/.confighub/bin/cub)
- Docker not running or accessible

## Development Workflow

1. **Setup Environment**
   - Install Go 1.23+
   - Install Docker
   - Install cub CLI
   - Clone repository

2. **Make Changes**
   - Create feature branch
   - Write tests first
   - Implement feature
   - Run tests locally

3. **Validate**
   - Run `make test`
   - Test with example workflows
   - Update documentation
   - Create PR

## Dependencies

### Direct Dependencies
- github.com/confighub/sdk (ConfigHub integration)
- github.com/nektos/act (GitHub Actions runner)
- github.com/spf13/cobra (CLI framework)
- github.com/prometheus/client_golang (metrics)

### Version Management
- Use go.mod for all dependencies
- Pin specific versions, avoid latest
- Test thoroughly when updating

## Debugging

### Common Commands
```bash
# Check worker status
curl http://localhost:8080/metrics

# Validate workflow syntax
./bin/cub-local-actions validate workflow.yml

# Test with verbose output
./bin/cub-local-actions run workflow.yml -v

# Check Docker connectivity
docker ps
```

### Logging
- Use structured logging
- Include context in error messages
- Log workflow execution stages
- Avoid logging sensitive data