# GitHub Actions Bridge

**Test GitHub Actions workflows locally with production configurations before pushing to GitHub.**

## Why This Matters

Ever pushed a workflow change only to watch it fail in CI? Spent hours debugging workflows through commit-push-wait cycles? Wished you could test with real secrets and configurations locally?

The GitHub Actions Bridge solves these problems by bringing GitHub Actions to your local machine, integrated with ConfigHub for secure configuration management.

## Key Benefits

✅ **Test Locally First** - Run workflows on your machine before committing  
✅ **Real Configurations** - Use actual configs from ConfigHub, not mock data  
✅ **Secure Secrets** - Access secrets without exposing them in code  
✅ **Time Travel Testing** - Test workflows with past or future configurations  
✅ **No More "Works on My Machine"** - Test with production-identical settings

## Quick Example

```bash
# Test your deployment workflow locally with production configs
cub-actions run .github/workflows/deploy.yml --space production --dry-run

# See what would have happened if you ran this last week
cub-actions run deploy.yml --as-of "2024-01-01" --space staging

# Compare workflow changes before pushing
cub-actions diff deploy.yml deploy-v2.yml --space production
```

## Documentation

📚 **[User Guide](USER_GUIDE.md)** - Start here if you're new  
🎯 **[Examples](examples/)** - 15+ real-world workflow examples with explanations  
🔧 **[API Reference](#cli-reference)** - Detailed command documentation

## Quick Start

### 1. Install

```bash
# Download latest release (macOS example)
curl -L https://github.com/confighub/actions-bridge/releases/latest/download/cub-actions-darwin-arm64 -o cub-actions
chmod +x cub-actions
sudo mv cub-actions /usr/local/bin/
```

### 2. Verify Setup

```bash
# Check installation
cub-actions version

# Run a simple test
cub-actions run examples/hello-world.yml
```

### 3. Run Your First Workflow

```bash
# Create a test workflow
cat > test.yml << 'EOF'
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
Your Workflow → GitHub Actions Bridge → ConfigHub → Local Execution
                        ↓
                 Workspace Isolation
                 Secret Management  
                 Compatibility Checks
```

The bridge uses [nektos/act](https://github.com/nektos/act) under the hood to execute GitHub Actions locally, adding:
- ConfigHub integration for configurations and secrets
- Workspace isolation for security
- Compatibility checking and warnings
- Advanced features like time-travel testing

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