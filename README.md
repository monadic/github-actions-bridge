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
# Create a GitHub Actions workflow as a ConfigHub unit
cub unit create --space production deploy-workflow .github/workflows/deploy.yml

# Apply the workflow in different environments
cub unit apply --space production deploy-workflow
cub unit apply --space staging webapp

# Test with historical configurations
cub unit apply --space production deploy-workflow --restore 1
```

## Documentation

📚 **[User Guide](USER_GUIDE.md)** - Start here if you're new  
🎯 **[Examples](examples/)** - 15+ real-world workflow examples with explanations  
🔧 **[CLI Reference](CLI_REFERENCE.md)** - Complete command documentation

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

3. **Security Note** - Read [SECURITY.md](SECURITY.md) for production deployment guidance

## Quick Start

### 1. Install Dependencies

```bash
# Install act (macOS example)
brew install act

# Verify act is installed
act --version
```

### 2. Install ConfigHub CLI and Bridge Worker

```bash
# Install ConfigHub CLI
curl -fsSL https://hub.confighub.com/cub/install.sh | bash

# The actions-bridge worker will be installed as part of ConfigHub
# Verify installation
cub --version
```

### 3. Verify Setup

```bash
# Login to ConfigHub
cub auth login

# Create and run a test workflow
cub unit create --space default hello-world examples/hello-world.yml
cub unit apply --space default hello-world
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

# Create a ConfigHub unit and run it
cub unit create --space default my-test test.yml
cub unit apply --space default my-test
```

## Real-World Examples

### Test with Secrets (No More Hardcoding!)
```bash
# Create secure secrets file
cat > secrets.env << EOF
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=sk_live_xxxxx
EOF

# Run workflow with secrets
./bin/cub-worker-actions run examples/with-secrets.yml --secrets-file secrets.env

# Or use ConfigHub for production
cub unit create --space production deploy deploy.yml
cub unit apply --space production deploy
```

### Test Different Configurations
```bash
# Test how your workflow behaves in different environments
cub unit create --space development deploy deploy.yml
cub unit create --space staging deploy deploy.yml
cub unit create --space production deploy deploy.yml

# Apply in different environments
cub unit apply --space development deploy
cub unit apply --space staging deploy
# Preview without applying
cub unit get --space production deploy --extended
```

### Debug Failed Workflows
```bash
# See exactly what's happening
cub unit create --space debug problematic problematic-workflow.yml --verbose
cub unit apply --space debug problematic --debug

# Check workflow status and logs
cub unit-event list --space debug --where "UnitSlug = 'problematic'"
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
# The bridge will be built as a ConfigHub worker
./bin/actions-bridge --version
```

### Run with Docker

```bash
# Development (with security warnings)
docker-compose up -d

# Production (secure configuration)
docker-compose -f docker-compose.secure.yml up -d
```

⚠️ **Security Note**: The default `docker-compose.yml` mounts the Docker socket for convenience but this is a security risk. For production use, see [SECURITY.md](SECURITY.md) and use `docker-compose.secure.yml`.

## Quick CLI Examples

### Local Execution with cub-worker-actions

```bash
# Run a workflow
./bin/cub-worker-actions run examples/hello-world.yml

# Validate a workflow
./bin/cub-worker-actions validate examples/build.yml

# Run with secrets
./bin/cub-worker-actions run examples/deploy.yml --secrets-file secrets.env
```

See [CLI_REFERENCE.md](CLI_REFERENCE.md) for complete command documentation.

### ConfigHub Integration

When using with ConfigHub:

```bash
# Create a workflow unit
cub unit create --space [space] [unit-name] [workflow-file]

# Apply (run) a workflow  
cub unit apply --space [space] [unit-name]

# View workflow details
cub unit get --space [space] [unit-name] --extended
```

## ConfigHub Integration

When integrated with ConfigHub, the bridge enables powerful features:

- **Configuration-Driven Deployments** - All values come from ConfigHub
- **Time Travel** - Test with past/future configurations  
- **Config-Triggered Workflows** - Auto-run when configs change
- **GitOps Without Git** - Use ConfigHub spaces instead of branches

See the [ConfigHub examples](examples/README.md#confighub-integration-examples) for detailed use cases.

## Examples Status

We provide 17+ example workflows in the `examples/` directory:

✅ **Working Examples** (15):
- Basic workflows (hello-world, environment-variables)
- Complex workflows (matrix-builds, multi-job, conditional-execution)
- Docker workflows (docker-compose-improved)
- ConfigHub integration (config-driven-deployment, time-travel-testing)
- AI integration (claude-orchestrated-ops, worker-calls-claude)

⚠️ **GitHub-Specific** (2):
- artifact-handling.yml (uses actions/upload-artifact)
- file-persistence.yml (uses actions/upload-artifact)

💡 **Improved Versions**: We provide local-compatible versions of GitHub-specific workflows (e.g., artifact-handling-improved.yml)

## Known Limitations

Some GitHub Actions features don't work in local execution:

- `actions/cache` - No caching support
- `actions/upload-artifact` & `actions/download-artifact` - Use local alternatives
- GitHub API calls - Limited or mocked
- Pull request creation - Not supported locally
- Cross-workflow artifacts - Local only

See the [act documentation](https://github.com/nektos/act#known-issues) for the full list of limitations.

## Documentation

**📚 [Complete Documentation Index](docs/INDEX.md)** - Full documentation map and guides

### Getting Started
- 📖 **[User Guide](USER_GUIDE.md)** - Comprehensive walkthrough
- 🎯 **[Examples](examples/)** - Learn by doing with 17+ examples
- 🔒 **[Security](SECURITY.md)** - Security considerations and best practices

### Technical Details
- 🏢 **[Enterprise Features](ENTERPRISE_FEATURES.md)** - Features provided by ConfigHub SaaS
- 📦 **[SDK Validation](SDK_VALIDATION.md)** - ConfigHub SDK dependency details
- 📝 **[SDK Requests](SDK_REQUESTS.md)** - Feature requests for ConfigHub SDK

### Contributing
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

## Dependencies

### Core Dependencies
- **[nektos/act](https://github.com/nektos/act)** - Local GitHub Actions runner (Apache 2.0)
- **[ConfigHub SDK](https://github.com/confighub/sdk)** - Official worker protocol SDK (MIT)
- **[spf13/cobra](https://github.com/spf13/cobra)** - CLI framework (Apache 2.0)

All dependencies are open source with permissive licenses.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [nektos/act](https://github.com/nektos/act) - Local GitHub Actions runner
- [ConfigHub](https://confighub.com) - Configuration management platform

---

**Ready to test your workflows locally?** Start with the **[User Guide](USER_GUIDE.md)** or jump into the **[Examples](examples/)**!