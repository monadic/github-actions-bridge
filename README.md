# GitHub Actions Bridge

**Run GitHub Actions workflows locally or through ConfigHub's configuration management platform.**

## 📍 Quick Navigation

**New to the project?** → Start with our [🚀 Quick Start Guide](USER_GUIDE.md#-quick-start---choose-your-path)  
**Want to see examples?** → Browse [📂 Examples](examples/) with [compatibility guide](EXAMPLES_COMPATIBILITY.md)  
**Need help?** → Check the [📖 User Guide](USER_GUIDE.md) or [🔧 CLI Reference](CLI_REFERENCE.md)  
**Having issues?** → See [🔍 Troubleshooting](USER_GUIDE.md#troubleshooting)

## 🔍 Two Ways to Use This Project

This project provides two distinct workflows for running GitHub Actions:

| Workflow | Tool | Use Case |
|----------|------|----------|
| **Local Development** | `cub-local-actions` | Test workflows on your machine without ConfigHub |
| **ConfigHub Integration** | `cub` | Production workflows with centralized configuration |

### Which workflow should I use?

```
Do you need ConfigHub features?
├─ No → Use cub-local-actions (Local Development)
└─ Yes → Use cub (ConfigHub Integration)
   └─ Features: Central config, team collaboration, 
      time travel, triggers, audit trails
```

## Why This Matters

Managing configurations and workflows separately leads to complexity and drift. The GitHub Actions Bridge solves this by:
- **Local Development**: Test GitHub Actions workflows instantly on your machine
- **ConfigHub Integration**: Unify workflows with configuration management for production

## Key Benefits

✅ **Local Testing** - Run workflows immediately without pushing to GitHub  
✅ **ConfigHub Integration** - Centralize workflow and configuration management  
✅ **Secure Secrets** - Local files for dev, ConfigHub for production  
✅ **Time Travel Testing** - Test with historical configurations (ConfigHub only)  
✅ **No Vendor Lock-in** - Use locally or with ConfigHub as needed

---

## 🚀 Workflow 1: Local Development (Quick Start)

Use `cub-local-actions` to test workflows on your local machine without any external dependencies.

### Prerequisites

1. **Docker** - Required for running containers
   - Mac: [Docker Desktop](https://www.docker.com/products/docker-desktop)
   - Linux: Docker Engine
   - Verify: `docker --version`

2. **act** - The GitHub Actions local runner
   - Mac: `brew install act`
   - Linux: `curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash`
   - Verify: `act --version`

### Installation

```bash
# Clone and build
git clone https://github.com/confighub/actions-bridge
cd actions-bridge
make build

# The local CLI will be at ./bin/cub-local-actions
```

### Usage

```bash
# Run a workflow
./bin/cub-local-actions run examples/hello-world.yml

# Validate without running
./bin/cub-local-actions validate examples/build.yml

# Run with secrets
./bin/cub-local-actions run examples/deploy.yml --secrets-file secrets.env

# Watch mode for development
./bin/cub-local-actions run workflow.yml --watch
```

**What to try first?**
1. Start with `hello-world.yml` - it always works
2. Then try `multi-job.yml` or `build-test-deploy.yml`
3. Check [Examples Compatibility Guide](EXAMPLES_COMPATIBILITY.md) to see which examples work locally

**💡 Tip**: If you're new, follow our [step-by-step Quick Start](USER_GUIDE.md#local-testing-quickstart) instead.

---

## 🏢 Workflow 2: ConfigHub Integration

Use the `cub` CLI to manage workflows through ConfigHub for production use with centralized configuration.

### Prerequisites

Same as local development, plus:
- ConfigHub account (https://hub.confighub.com)
- `cub` CLI installed

### Installation

```bash
# Install ConfigHub CLI
curl -fsSL https://hub.confighub.com/cub/install.sh | bash

# IMPORTANT: Add cub to your PATH
# The installer places cub at ~/.confighub/bin/cub

# For current session:
export PATH="$HOME/.confighub/bin:$PATH"

# For permanent installation, add to your shell profile:
echo 'export PATH="$HOME/.confighub/bin:$PATH"' >> ~/.zshrc  # For zsh (default on macOS)
# OR
echo 'export PATH="$HOME/.confighub/bin:$PATH"' >> ~/.bashrc # For bash

# Reload your shell configuration:
source ~/.zshrc  # or source ~/.bashrc

# Verify installation
cub version
```

**Troubleshooting PATH Issues:**
- If `cub: command not found`, the PATH isn't set correctly
- Check installation location: `ls ~/.confighub/bin/cub`
- Use full path if needed: `~/.confighub/bin/cub version`

### Setup

1. **Login to ConfigHub**
   ```bash
   cub auth login
   ```

2. **Set your working space**
   ```bash
   cub context set --space <your-space-name>
   ```

3. **Create a worker**
   ```bash
   cub worker create actions-bridge-1
   eval "$(cub worker get-envs actions-bridge-1)"
   ```

4. **Start the bridge worker**
   ```bash
   ./bin/actions-bridge
   ```

### Usage

```bash
# Create a workflow unit
cub unit create --space production hello examples/hello-world.yml

# Apply (run) the workflow
cub unit apply --space production hello

# Use different environments
cub unit apply --space staging hello
cub unit apply --space development hello

# Time travel with previous configurations
cub unit apply --space production hello --restore 1
```

### ConfigHub Integration Examples

Some examples require or simulate ConfigHub features:
- `config-driven-deployment.yml` - Basic ConfigHub integration (partially simulated)
- `time-travel-testing.yml` - Simulated with date comparisons
- `claude-orchestrated-ops.yml` - AI responses are mocked

See the full list and implementation status in our [Examples Compatibility Guide](EXAMPLES_COMPATIBILITY.md).

---

## 📋 Workflow Format

All workflows use the ConfigHub Actions format (Kubernetes-style header):

```yaml
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: your-workflow-name
# Standard GitHub Actions workflow follows
name: Your Workflow
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Hello World!"
```

**Note**: The header is required but automatically stripped when running locally.

---

## 📚 Documentation

**📚 [Complete Documentation Index](docs/INDEX.md)** - Full documentation map and guides

### Getting Started
- 📖 **[User Guide](USER_GUIDE.md)** - Comprehensive walkthrough for both workflows
- 🆕 **[ConfigHub Setup Guide](CONFIGHUB_SETUP_GUIDE.md)** - Step-by-step ConfigHub setup with troubleshooting
- 🎯 **[Examples](examples/)** - 17 workflow examples (some simulated)
- 📊 **[Examples Compatibility](EXAMPLES_COMPATIBILITY.md)** - Which examples work where
- 🔒 **[Security](SECURITY.md)** - Security considerations and best practices

### Reference
- 🔧 **[CLI Reference](CLI_REFERENCE.md)** - Complete `cub-local-actions` documentation
- 📝 **[YAML Formats](YAML_FORMATS.md)** - Workflow format specifications

### Advanced
- 🏢 **[Enterprise Features](ENTERPRISE_FEATURES.md)** - Features provided by ConfigHub SaaS
- 📦 **[SDK Validation](SDK_VALIDATION.md)** - ConfigHub SDK dependency details

### About the cub CLI

The `cub` command-line interface is ConfigHub's comprehensive tool. For details:
- Run `cub --help-overview` for complete documentation
- Standard pattern: `cub <entity> <verb> [flags] [arguments]`

---

## 🏗️ Architecture

```
┌─────────────────┐     ┌──────────────────┐
│ Local Workflow  │     │ ConfigHub Flow   │
├─────────────────┤     ├──────────────────│
│ cub-local-      │     │ cub CLI          │
│ actions         │     │   ↓              │
│   ↓             │     │ ConfigHub API    │
│ act (local)     │     │   ↓              │
│   ↓             │     │ actions-bridge   │
│ Docker          │     │   ↓              │
└─────────────────┘     │ act (local)      │
                        │   ↓              │
                        │ Docker           │
                        └──────────────────┘
```

<details>
<summary>Technical Architecture Details (click to expand)</summary>

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
+----------------------------------+
|    Workspace Manager             |
|    - Isolation per execution     |
|    - Secure cleanup              |
+----------------------------------+
|    Act Wrapper                   |
|    - nektos/act integration      |
|    - Compatibility layer         |
+----------------------------------+
```

### Project Structure

```
github-actions-bridge/
├── examples/           # 17+ workflow examples
├── cmd/               # CLI and bridge binaries  
├── pkg/               # Core implementation
├── test/              # Test suites
├── USER_GUIDE.md      # Comprehensive guide
├── Dockerfile         # Container image
└── docker-compose.yml # Easy deployment
```

</details>

---

## 🚨 Known Limitations

Some GitHub Actions features don't work locally:
- `actions/cache` - No caching support
- `actions/upload-artifact` to GitHub - Use local alternatives
- GitHub API calls - Limited or mocked
- Pull request creation - Not supported

See [act documentation](https://github.com/nektos/act#known-issues) for details.

---

## Package Dependencies

### Module Information
- **Module Path**: `github.com/confighub/actions-bridge`
- **Go Version**: 1.24.3 (requires Go 1.24+)
- **License**: MIT
- **Documentation Updated**: August 7, 2025

#### Version Context
As of August 7, 2025:
- **Latest Go stable version**: 1.24.6 (released August 6, 2025)
- **Go 1.25**: Expected release August 2025
- This project uses Go 1.24.3 which is a valid recent version

### Core Dependencies

#### ConfigHub SDK
- **Package**: `github.com/confighub/sdk`
- **Version**: `v0.0.0-20250804044729-f1517379cea0` (pseudo-version)
- **Repository**: [https://github.com/confighub/sdk](https://github.com/confighub/sdk)
- **License**: MIT
- **Note**: This project uses a pseudo-version (timestamp-based) rather than a tagged release. This indicates the SDK is pinned to a specific commit.

#### GitHub Actions Local Runner
- **Package**: `github.com/nektos/act`
- **Version**: `v0.2.80`
- **Repository**: [https://github.com/nektos/act](https://github.com/nektos/act)
- **License**: Apache 2.0
- **Purpose**: Enables local execution of GitHub Actions workflows

#### CLI Framework
- **Package**: `github.com/spf13/cobra`
- **Version**: `v1.9.1`
- **Repository**: [https://github.com/spf13/cobra](https://github.com/spf13/cobra)
- **License**: Apache 2.0
- **Purpose**: Powers the command-line interface

### Additional Direct Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/google/uuid` | v1.6.0 | UUID generation for workspace management |
| `github.com/prometheus/client_golang` | v1.22.0 | Metrics and monitoring |
| `github.com/stretchr/testify` | v1.10.0 | Testing framework |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML parsing for workflows |

### Version Management

#### Pseudo-versions Explained
The ConfigHub SDK uses a **pseudo-version** format: `v0.0.0-20250804044729-f1517379cea0`

This format breaks down as:
- `v0.0.0` - Base version (no official release tag)
- `20250804044729` - Timestamp (August 4, 2025 at 04:47:29 UTC)
- `f1517379cea0` - Short commit hash

Note: This timestamp is from 3 days ago (August 4, 2025), indicating active development.

**Important**: Pseudo-versions indicate the dependency is pinned to a specific commit rather than a stable release. This may mean:
- The SDK is under active development
- The project requires features not yet in a tagged release
- Extra care should be taken when updating this dependency

#### Dependency Verification
To verify the exact versions used in this project:
```bash
# View all dependencies
go list -m all

# View specific dependency details
go list -m github.com/confighub/sdk
go list -m github.com/nektos/act
```

### Docker Image Note
The project also has a runtime dependency on Docker being installed and running, as `nektos/act` requires Docker to execute GitHub Actions workflows in containers.

All dependencies are open source with permissive licenses.

---

## Installation Options

### Download Pre-built Binary (Recommended)

See platform-specific downloads on the [releases page](https://github.com/confighub/actions-bridge/releases).

### Build from Source

```bash
git clone https://github.com/confighub/actions-bridge
cd actions-bridge
make build
```

### Run with Docker

```bash
# Development (with security warnings)
docker-compose up -d

# Production (secure configuration)
docker-compose -f docker-compose.secure.yml up -d
```

⚠️ **Security Note**: The default `docker-compose.yml` mounts the Docker socket for convenience but this is a security risk. For production use, see [SECURITY.md](SECURITY.md) and use `docker-compose.secure.yml`.

---

## Quick Decision Guide

### Use Local Development when:
- Testing workflows during development
- Quick iteration needed
- No need for central configuration
- Working offline

### Use ConfigHub Integration when:
- Managing production workflows
- Team collaboration required
- Need audit trails and versioning
- Want configuration-driven deployments

---

## Contributing

- 💬 **[Issues](https://github.com/confighub/actions-bridge/issues)** - Report bugs or request features

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [nektos/act](https://github.com/nektos/act) - Local GitHub Actions runner
- [ConfigHub](https://confighub.com) - Configuration management platform

---

**Ready to get started?** Choose your workflow:
- 🚀 **[Local Development](#-workflow-1-local-development-quick-start)** - Start testing in seconds
- 🏢 **[ConfigHub Integration](#-workflow-2-confighub-integration)** - Production-ready workflow management