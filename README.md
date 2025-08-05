# GitHub Actions Bridge for ConfigHub

A ConfigHub bridge that enables local execution of GitHub Actions workflows using `act`, with proper workspace isolation, secret handling, and full bridge interface implementation.

## Features

- **Local GitHub Actions Execution** - Run workflows locally using `act`
- **Secure Secret Management** - File-based secrets with encryption and leak detection
- **Workspace Isolation** - Each execution gets an isolated workspace with cleanup
- **ConfigHub Integration** - Full bridge interface implementation
- **Compatibility Checking** - Warns about act limitations before execution
- **Health Monitoring** - Built-in health checks and metrics

## Quick Start

### Prerequisites

- Go 1.21 or later
- Docker (for running workflows)
- ConfigHub worker credentials

### Installation

```bash
# Clone the repository
git clone https://github.com/confighub/actions-bridge
cd actions-bridge

# Build the project
make build

# Run tests
make test
```

### Running with Docker (Recommended)

```bash
# Copy environment file
cp .env.example .env

# Edit .env with your ConfigHub credentials
vim .env

# Start with Docker Compose
docker-compose up -d

# Check logs
docker-compose logs -f actions-bridge

# Stop the bridge
docker-compose down
```

### Running Manually

```bash
# Set required environment variables
export CONFIGHUB_WORKER_ID=your-worker-id
export CONFIGHUB_WORKER_SECRET=your-worker-secret
export CONFIGHUB_URL=https://api.confighub.com
export ACTIONS_BRIDGE_BASE_DIR=/var/lib/actions-bridge

# Run the bridge
./bin/actions-bridge
```

### Using the CLI

```bash
# Run a workflow
./bin/cub-actions run test/fixtures/workflows/simple.yml

# Run with secrets
./bin/cub-actions run workflow.yml --secrets-file secrets.env

# Validate a workflow
./bin/cub-actions validate workflow.yml

# List known limitations
./bin/cub-actions list-limitations
```

## Architecture

```
ConfigHub API
     |
     v
+----------------------------------+
|  GitHub Actions Bridge Worker    |
+----------------------------------+
|    Bridge Interface              |
|    - Info()                      |
|    - Apply() -> Execute          |
|    - Refresh() -> Status         |
|    - Destroy() -> Cleanup        |
|    - Import() -> Discover        |
|    - Finalize() -> Archive       |
+----------------------------------+
|                                  |
|    Workspace Manager             |
|    - Isolation per exec          |
|    - Secure cleanup              |
|    - Audit trail                 |
|                                  |
+----------------------------------+
|                                  |
|    Act Wrapper                   |
|    - Compatibility layer         |
|    - Secret file handling        |
|    - Output capture              |
|                                  |
+----------------------------------+
```

## CLI Usage

### Run Command

```bash
cub-actions run [workflow-file] [flags]

Flags:
  --space string         ConfigHub space
  --unit string          ConfigHub unit
  --dry-run              Show what would be executed
  --event string         GitHub event type (default "workflow_dispatch")
  -i, --input strings    Workflow inputs (key=value)
  --platform string      Execution platform (default "linux/amd64")
  --artifact-dir string  Directory to save artifacts
  --env-file string      Environment file to load
  --secrets-file string  Secrets file to load
  --validate             Validate workflow without running
  --timeout int          Execution timeout in seconds (default 3600)
```

### Validate Command

```bash
cub-actions validate [workflow-file]
```

Checks if a workflow is valid and can be executed locally with act.

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIGHUB_WORKER_ID` | ConfigHub worker ID | Required |
| `CONFIGHUB_WORKER_SECRET` | ConfigHub worker secret | Required |
| `CONFIGHUB_URL` | ConfigHub API URL | `https://api.confighub.com` |
| `ACTIONS_BRIDGE_BASE_DIR` | Base directory for workspaces | `/var/lib/actions-bridge` |
| `ACT_DEFAULT_IMAGE` | Default Docker image for act | `catthehacker/ubuntu:act-latest` |
| `ACT_PLATFORM` | Default platform | `linux/amd64` |
| `MAX_CONCURRENT_WORKFLOWS` | Max concurrent executions | `5` |
| `HEALTH_ADDR` | Health check server address | `:8080` |

### Secrets File Format

```bash
# secrets.env
API_KEY=your-api-key
DATABASE_URL=postgres://user:pass@localhost/db
GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

## Known Limitations

When running GitHub Actions locally with act, be aware of these limitations:

1. **No Caching** - `actions/cache` is not supported
2. **Limited Artifacts** - Artifacts are saved locally only
3. **No Cross-workflow Artifacts** - Can't download from other workflows
4. **No Registry Push** - Docker push operations are disabled
5. **Simulated GITHUB_TOKEN** - GitHub token is simulated
6. **No GitHub API** - API calls may fail or need mocking
7. **No Pull Requests** - Can't create PRs locally
8. **No Releases** - Can't create GitHub releases

## Health Monitoring

The bridge exposes health endpoints:

- `/health` - Overall health status
- `/ready` - Readiness check
- `/live` - Liveness check
- `/metrics` - Prometheus metrics

## Development

### Project Structure

```
.
|-- cmd/
|   |-- act-test/        # Act validation tool
|   |-- actions-bridge/  # Main bridge worker
|   `-- actions-cli/     # CLI tool
|-- pkg/
|   |-- bridge/          # Core bridge implementation
|   `-- leakdetector/    # Secret leak detection
|-- test/
|   |-- fixtures/        # Test workflows
|   `-- integration/     # Integration tests
|-- Dockerfile           # Multi-stage Docker build
|-- docker-compose.yml   # Docker Compose configuration
|-- prometheus.yml       # Prometheus monitoring config
|-- .env.example         # Example environment file
|-- Makefile
`-- go.mod
```

### Building

```bash
# Build all binaries
make build

# Build specific binary
make build-bridge
make build-cli
make build-act-test

# Run tests
make test

# Run integration tests
make test-integration
```

### Testing

```bash
# Run all tests
make test

# Run with Docker (for act tests)
SKIP_ACT_TESTS=0 make test

# Run specific test
go test -v ./pkg/bridge -run TestWorkspaceIsolation
```

## Security

- Secrets are stored in files with 0600 permissions
- Workspace isolation prevents cross-execution access
- Secure cleanup overwrites secrets before deletion
- Leak detection prevents secrets in logs
- All secrets are sanitized in output

## Troubleshooting

### Docker not found
Ensure Docker is installed and running:
```bash
docker version
```

### Workflow fails locally but works on GitHub
Check the compatibility warnings:
```bash
cub-actions validate workflow.yml
```

### Permission denied errors
Ensure the base directory is writable:
```bash
mkdir -p /var/lib/actions-bridge
chmod 755 /var/lib/actions-bridge
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [nektos/act](https://github.com/nektos/act) - Local GitHub Actions runner
- [ConfigHub](https://confighub.com) - Configuration management platform