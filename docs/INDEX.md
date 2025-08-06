# GitHub Actions Bridge - Documentation Index

This index provides a complete overview of all documentation for the GitHub Actions Bridge project.

## 📚 Documentation Structure

### Core Documentation
- **[README.md](../README.md)** - Project overview, quick start, and examples
- **[USER_GUIDE.md](../USER_GUIDE.md)** - Comprehensive user guide and tutorials
- **[CLI_REFERENCE.md](../CLI_REFERENCE.md)** - Complete CLI command reference
- **[CONTRIBUTING.md](../CONTRIBUTING.md)** - How to contribute to the project

### Technical Documentation
- **[SECURITY.md](../SECURITY.md)** - Security considerations and best practices
- **[SDK_VALIDATION.md](../SDK_VALIDATION.md)** - ConfigHub SDK dependency analysis
- **[SDK_REQUESTS.md](../SDK_REQUESTS.md)** - Feature requests for ConfigHub SDK team
- **[ENTERPRISE_FEATURES.md](../ENTERPRISE_FEATURES.md)** - Features deliberately delegated to ConfigHub SaaS
- **[YAML_FORMATS.md](../YAML_FORMATS.md)** - YAML format specifications and validation

### Examples and Guides
- **[examples/README.md](../examples/README.md)** - Detailed guide to all 17+ workflow examples
- **[examples/](../examples/)** - Actual workflow files with comments

### Configuration Files
- **[docker-compose.yml](../docker-compose.yml)** - Default Docker deployment (with security warnings)
- **[docker-compose.secure.yml](../docker-compose.secure.yml)** - Production-ready secure deployment
- **[prometheus.yml](../prometheus.yml)** - Prometheus monitoring configuration

### Development Files
- **[Makefile](../Makefile)** - Build targets and development commands
- **[go.mod](../go.mod)** - Go module dependencies
- **[Dockerfile](../Dockerfile)** - Container image definition

## 🗺️ Documentation Map

```
Start Here
    ↓
README.md → Quick Start → Examples
    ↓                        ↓
USER_GUIDE.md          examples/README.md
    ↓
For Production?
    ↓
SECURITY.md → docker-compose.secure.yml
    ↓
Need Enterprise Features?
    ↓
ENTERPRISE_FEATURES.md → ConfigHub SaaS
    ↓
Contributing?
    ↓
CONTRIBUTING.md → SDK_REQUESTS.md
```

## 📖 Reading Order for New Users

1. **[README.md](../README.md)** - Understand what the project does
2. **[examples/hello-world.yml](../examples/hello-world.yml)** - See a simple example
3. **[USER_GUIDE.md](../USER_GUIDE.md)** - Learn how to use it
4. **[SECURITY.md](../SECURITY.md)** - Understand security implications
5. **[ENTERPRISE_FEATURES.md](../ENTERPRISE_FEATURES.md)** - Learn about advanced features

## 🔍 Finding Specific Information

### "How do I..."
- **Run a workflow locally?** → [README.md Quick Start](../README.md#quick-start)
- **Use secrets?** → [USER_GUIDE.md Secrets](../USER_GUIDE.md#using-secrets)
- **Deploy to production?** → [SECURITY.md](../SECURITY.md) + [docker-compose.secure.yml](../docker-compose.secure.yml)
- **Get enterprise features?** → [ENTERPRISE_FEATURES.md](../ENTERPRISE_FEATURES.md)

### "What about..."
- **Security risks?** → [SECURITY.md](../SECURITY.md)
- **ConfigHub SDK?** → [SDK_VALIDATION.md](../SDK_VALIDATION.md)
- **Missing features?** → [ENTERPRISE_FEATURES.md](../ENTERPRISE_FEATURES.md)
- **Contributing?** → [CONTRIBUTING.md](../CONTRIBUTING.md)

### "I found a..."
- **Bug** → [GitHub Issues](https://github.com/confighub/actions-bridge/issues)
- **Security issue** → [SECURITY.md#reporting-security-issues](../SECURITY.md#reporting-security-issues)
- **Missing SDK feature** → [SDK_REQUESTS.md](../SDK_REQUESTS.md)

## 📊 Documentation Status

| Document | Purpose | Status | Last Updated |
|----------|---------|--------|--------------|
| README.md | Project overview | ✅ Complete | Current |
| USER_GUIDE.md | User tutorials | ✅ Complete | Current |
| SECURITY.md | Security guide | ✅ Complete | Current |
| ENTERPRISE_FEATURES.md | Feature delegation | ✅ Complete | Current |
| SDK_VALIDATION.md | SDK analysis | ✅ Complete | Current |
| SDK_REQUESTS.md | SDK improvements | ✅ Complete | Current |
| CONTRIBUTING.md | Contribution guide | ⚠️ Needs creation | - |

## 🔗 External Resources

- **ConfigHub Documentation**: https://docs.confighub.com
- **ConfigHub SDK**: https://github.com/confighub/sdk
- **act Documentation**: https://github.com/nektos/act
- **GitHub Actions Docs**: https://docs.github.com/actions

## 📝 Documentation Principles

1. **Honest** - No exaggerated claims or fake features
2. **Practical** - Real examples that actually work
3. **Clear** - Explicit about what's included vs enterprise features
4. **Connected** - All docs reference related documentation

---

*This documentation index is maintained as part of the GitHub Actions Bridge project. For updates or corrections, please submit a pull request.*