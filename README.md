# GitHub Actions Bridge

**Combine configuration management with GitHub Actions using act and ConfigHub.**

## Why This Matters

Managing configurations and workflows separately leads to complexity and drift. You have workflows in GitHub, configurations in various places, secrets scattered across systems, and no unified way to manage them.

The GitHub Actions Bridge simplifies this by unifying GitHub Actions workflows with ConfigHub's configuration management, creating a single source of truth for both your workflows and configurations.

## Key Benefits

✅ **Unified Configuration** - Manage workflows and configs in one place  
✅ **Local Execution** - Run GitHub Actions anywhere using act  
✅ **Secure Secrets** - ConfigHub handles all secret management  
✅ **Time Travel Testing** - Test workflows with past or future configurations  
✅ **Configuration as Code** - Treat workflows as configuration that evolves with your app

## Getting Started Guide

### Setting Up ConfigHub Worker

To run the GitHub Actions Bridge as a ConfigHub worker:

1. **Login to ConfigHub**
   ```bash
   cub auth login
   ```

2. **Set your working context**
   ```bash
   # Check current context
   cub context get
   
   # Set the space you want to work with
   cub context set --space <your-space-name>
   ```

3. **Create a worker for the bridge**
   ```bash
   # Create a new worker instance
   cub worker create actions-bridge-1
   ```

4. **Get worker credentials**
   ```bash
   # Display environment variables for the worker
   cub worker get-envs actions-bridge-1
   
   # Set them in your shell
   eval "$(cub worker get-envs actions-bridge-1)"

   export CONFIGHUB_URL=https://hub.confighub.com
   ```

5. **Start the bridge worker**
   ```bash   
   # With environment variables set from step 4
   ./bin/actions-bridge   
   ```

6. **Create a Hello World GitHub Actions as a Config Unit**
   ```bash
   cub unit create hello-world examples/hello-world.yml
   ```

Now you can go to https://hub.confighub.com. Then find a Worker named `actions-bridge-1` and set its Target to `docker-desktop`.  Next find a Config Unit named `hello-world` in your space, then you can review the config and click [Apply] to apply the GitHub Actions.

The bridge will now connect to ConfigHub and execute GitHub Actions workflows based on your configurations.

### Creating ConfigHub Actions Resources

When creating GitHub Actions workflows for ConfigHub, you need to add a Kubernetes-style resource header. The bridge will automatically strip these lines before passing the workflow to Act.

**Required Format:**
```yaml
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: your-workflow-name
# Your standard GitHub Actions workflow starts here
name: Your Workflow
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Hello from ConfigHub!"
```

All example workflows in the `examples/` directory are require to include this header before use.

## Quick Example

```bash
# Run workflows with configurations from ConfigHub
cub-actions run .github/workflows/deploy.yml --space production

# Use different configurations for different environments
cub-actions run deploy.yml --space staging --unit webapp

# Test with historical configurations
cub-actions run deploy.yml --as-of "2024-01-01" --space production
```

## Documentation

📚 **[User Guide](USER_GUIDE.md)** - Start here if you're new  
🎯 **[Examples](examples/)** - 15+ real-world workflow examples with explanations  
🔧 **[API Reference](#cli-reference)** - Detailed command documentation

## Prerequisites

Before using GitHub Actions Bridge, you need to install these dependencies:

1. **Docker** - Required for running containers
   - Mac: [Docker Desktop](https://www.docker.com/products/docker-desktop)
   - Linux: Docker Engine
   - Verify: `docker --version`

2. **act** - The GitHub Actions local runner
   - Mac: `brew install act`
   - Linux: `curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash`
   - Windows: `choco install act-cli`
   - Verify: `act --version`

## Quick Start

### 1. Install Dependencies

```bash
# Install act (macOS example)
brew install act

# Verify act is installed
act --version
```

### 2. Install GitHub Actions Bridge

```bash
# Download latest release (macOS example)
curl -L https://github.com/confighub/actions-bridge/releases/latest/download/cub-actions-darwin-arm64 -o cub-actions
chmod +x cub-actions
sudo mv cub-actions /usr/local/bin/
```

### 3. Verify Setup

```bash
# Check installation
cub-actions version

# Run a simple test
cub-actions run examples/hello-world.yml
```

### 4. Run Your First Workflow

```bash
# Create a test workflow
cat > test.yml << 'EOF'
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: my-test
name: My Test
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Hello from local GitHub Actions!"
EOF

# Run it
cub-actions run test.yml
```

## Real-World Examples

### Test with Secrets (No More Hardcoding!)
```bash
# Create secure secrets file
cat > secrets.env << EOF
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=sk_live_xxxxx
EOF

# Run workflow with real secrets
cub-actions run deploy.yml --secrets-file secrets.env
```

### Test Different Configurations
```bash
# Test how your workflow behaves in different environments
cub-actions run deploy.yml --space development
cub-actions run deploy.yml --space staging  
cub-actions run deploy.yml --space production --dry-run
```

### Debug Failed Workflows
```bash
# See exactly what's happening
cub-actions run problematic-workflow.yml -v

# Check if workflow will work locally
cub-actions validate workflow.yml
```

## How It Works

```
ConfigHub Configurations + GitHub Actions Workflows
                    ↓
            GitHub Actions Bridge
                    ↓
    Local Execution via act (nektos/act)
```

The bridge connects three powerful tools:
- **ConfigHub**: Manages all your configurations and secrets
- **GitHub Actions**: Defines your workflows and automation
- **act**: Executes GitHub Actions locally in Docker containers

This combination enables configuration-driven workflows where your CI/CD pipelines adapt based on ConfigHub configurations.

## Installation Options

### Download Pre-built Binary (Recommended)

See platform-specific downloads on the [releases page](https://github.com/confighub/actions-bridge/releases).

### Build from Source

```bash
git clone https://github.com/confighub/actions-bridge
cd actions-bridge
make build
./bin/cub-actions version
```

### Run with Docker

```bash
docker run -v $(pwd):/workspace confighub/actions-bridge run workflow.yml
```

## CLI Reference

### Core Commands

```bash
# Run a workflow
cub-actions run [workflow-file] [flags]

# Validate without running  
cub-actions validate [workflow-file]

# Compare workflow versions
cub-actions diff [workflow1] [workflow2]

# List known limitations
cub-actions list-limitations
```

### Key Flags

- `--space` - ConfigHub space (development, staging, production)
- `--unit` - ConfigHub unit name
- `--dry-run` - Preview what would happen without executing
- `--secrets-file` - File containing secrets
- `--as-of` - Run with historical configuration
- `-v` - Verbose output for debugging

## ConfigHub Integration

When integrated with ConfigHub, the bridge enables powerful features:

- **Configuration-Driven Deployments** - All values come from ConfigHub
- **Time Travel** - Test with past/future configurations  
- **Config-Triggered Workflows** - Auto-run when configs change
- **GitOps Without Git** - Use ConfigHub spaces instead of branches

See the [ConfigHub examples](examples/README.md#confighub-integration-examples) for detailed use cases.

## Known Limitations

Some GitHub Actions features don't work in local execution:

- `actions/cache` - No caching support
- GitHub API calls - Limited or mocked
- Pull request creation - Not supported locally
- Cross-workflow artifacts - Local only

Run `cub-actions list-limitations` for the full list.

## Getting Help

- 📖 **[User Guide](USER_GUIDE.md)** - Comprehensive walkthrough
- 🎯 **[Examples](examples/)** - Learn by doing
- 💬 **[Issues](https://github.com/confighub/actions-bridge/issues)** - Report bugs or request features
- 🤝 **[Contributing](CONTRIBUTING.md)** - Help improve the bridge

## Architecture

<details>
<summary>Technical Details (click to expand)</summary>

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
├── examples/           # 15+ workflow examples
├── cmd/               # CLI and bridge binaries  
├── pkg/               # Core implementation
├── test/              # Test suites
├── USER_GUIDE.md      # Beginner's guide
├── Dockerfile         # Container image
└── docker-compose.yml # Easy deployment
```

</details>

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [nektos/act](https://github.com/nektos/act) - Local GitHub Actions runner
- [ConfigHub](https://confighub.com) - Configuration management platform

---

**Ready to test your workflows locally?** Start with the **[User Guide](USER_GUIDE.md)** or jump into the **[Examples](examples/)**!