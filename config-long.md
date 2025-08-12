# Configuration Management Guide

## What This Is

This guide helps AI tools understand how configuration drives this application through ConfigHub - a platform that treats configuration as first-class infrastructure. Think of it as "GitOps for config" where your app's settings, workflows, and operational state are managed declaratively.

## Core Concepts

### ConfigHub Platform
ConfigHub is where your configurations live, version, and deploy. It's like GitHub but for configuration:
- **Spaces**: Isolated environments for different projects or teams
- **Units**: Individual configuration items (like a workflow, a service config, or secrets)
- **Workers**: Services that apply configurations to make things happen
- **Targets**: Where configurations get applied (e.g., docker-desktop, kubernetes)

### This Project's Role
The GitHub Actions Bridge is a **worker** that:
1. Receives workflow configurations from ConfigHub
2. Executes them using GitHub Actions locally
3. Reports status back to ConfigHub
4. Makes workflows manageable as configuration units

## How Configurations Flow

```
Developer → ConfigHub → Bridge Worker → GitHub Actions → Results
    ↑                                                         ↓
    └─────────────── Status & Logs ←─────────────────────────┘
```

### Step by Step
1. Developer creates a workflow file with ConfigHub headers
2. Uses `cub unit create` to register it in ConfigHub
3. Uses `cub unit apply` to execute it
4. Bridge worker receives the configuration
5. Strips ConfigHub headers and runs via `act`
6. Returns execution results to ConfigHub

## Configuration Anatomy

### Workflow Configuration
```yaml
# ConfigHub Resource Header (required)
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: my-deployment
  annotations:
    description: "Deploy my application"
    environment: "production"

# Standard GitHub Actions (passed to act)
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Deploy
        env:
          API_KEY: ${{ secrets.API_KEY }}
        run: |
          ./deploy.sh
```

### Worker Configuration
Workers need credentials to connect to ConfigHub:
```bash
CONFIGHUB_WORKER_ID=worker-abc123
CONFIGHUB_WORKER_SECRET=secret-xyz789
CONFIGHUB_URL=https://hub.confighub.com
```

## SDK Integration

### Using the ConfigHub SDK
The bridge uses ConfigHub's Go SDK to:
```go
// Connect as a worker
client := sdk.NewClient(workerID, workerSecret, hubURL)

// Implement worker protocol
func (b *Bridge) Apply(unit *sdk.Unit) (*sdk.ApplyResult, error) {
    // Extract workflow from unit.Spec
    // Execute with act
    // Return results
}
```

### Key SDK Patterns
- **Authentication**: Worker ID/Secret pairs
- **Protocol**: REST + WebSocket for real-time updates
- **State Management**: ConfigHub tracks desired vs actual state
- **Error Handling**: Structured error responses with remediation hints

## Workers and Bridges

### Worker Types in ConfigHub Ecosystem
1. **Infrastructure Workers**: Apply Terraform, Pulumi configs
2. **Application Workers**: Deploy containers, update services  
3. **Workflow Workers**: Run CI/CD pipelines (like this bridge)
4. **Security Workers**: Apply policies, scan configurations

### How Bridges Connect Systems
Bridges translate between ConfigHub's declarative model and other systems:
- **This Bridge**: ConfigHub ↔ GitHub Actions
- **Kubernetes Bridge**: ConfigHub ↔ K8s Resources
- **Cloud Bridges**: ConfigHub ↔ AWS/GCP/Azure

### Composing Applications
Your app might use multiple workers:
```yaml
# App composed of multiple configurations
my-app:
  - database:     # Managed by Terraform worker
      type: rds
      size: db.t3.medium
  
  - api-service:  # Managed by Kubernetes worker
      image: myapp/api:v2.0
      replicas: 3
  
  - deployment:   # Managed by Actions bridge
      workflow: deploy-api.yml
      triggers: [database, api-service]
```

## Configuration Best Practices

### Structure Your Configs
1. **Separate Concerns**: Different configs for different aspects
2. **Use Spaces**: Dev/staging/prod in separate spaces
3. **Version Everything**: ConfigHub tracks all changes
4. **Link Dependencies**: Express relationships between configs

### Security Considerations
- **Secrets**: Store in ConfigHub, injected at runtime
- **Access Control**: Space-level permissions
- **Audit Trail**: All changes tracked
- **Encryption**: Transit and at-rest

### Testing Configurations
1. **Dry Run**: Many workers support plan/preview modes
2. **Staging Spaces**: Test configs before production
3. **Rollback**: ConfigHub maintains history
4. **Validation**: Workers validate before applying

## Common Patterns

### GitOps Workflow
```bash
# 1. Make changes locally
edit workflow.yml

# 2. Preview changes
cub unit plan my-workflow

# 3. Apply to staging
cub context set --space staging
cub unit apply my-workflow

# 4. Promote to production
cub context set --space production
cub unit apply my-workflow
```

### Multi-Environment Setup
```yaml
# base-workflow.yml
apiVersion: actions.confighub.com/v1alpha1
kind: Actions
metadata:
  name: deploy-app
  
# Override per environment
staging:
  environment: staging
  replicas: 1
  
production:
  environment: production
  replicas: 3
```

### Dependency Management
```yaml
# ConfigHub tracks dependencies
metadata:
  name: frontend-deploy
  depends_on:
    - api-deploy
    - database-migration
```

## Troubleshooting Configurations

### Common Issues
1. **Wrong URL**: Use `hub.confighub.com` not `api.confighub.com`
2. **Missing Target**: Always specify `--target docker-desktop`
3. **Auth Failures**: Check worker credentials
4. **Version Mismatch**: Ensure SDK compatibility

### Debugging Commands
```bash
# Check current context
cub context get

# List configurations
cub unit list

# Get configuration details
cub unit get my-workflow

# View apply history
cub unit history my-workflow
```

## Future Vision

### Where This Is Going
- **More Bridges**: Connect any system to ConfigHub
- **Policy Engine**: Enforce standards across all configs
- **AI Integration**: Suggest optimizations, detect issues
- **Federation**: Connect multiple ConfigHub instances

### Getting Involved
- **Build Workers**: Extend ConfigHub to new platforms
- **Share Configs**: Publish reusable configurations
- **Contribute**: Improve SDK, docs, tools

## Quick Reference

### Essential Commands
```bash
# Setup
cub auth login
cub context set --space my-space

# Work with configs
cub unit create name file.yml --target docker-desktop
cub unit apply name
cub unit get name
cub unit delete name

# Worker management
cub worker create my-worker
cub worker get-envs my-worker
```

### Key Environment Variables
```bash
CONFIGHUB_URL          # Hub API endpoint
CONFIGHUB_WORKER_ID    # Worker authentication
CONFIGHUB_WORKER_SECRET # Worker secret
CONFIGHUB_SPACE        # Default space
```

### File Patterns
- `*.yml` with ConfigHub headers → Configuration units
- `CLAUDE.md` → AI context for the project
- `AGENT.md` → AI agent behavior configuration
- `CONFIG.md` → This file, configuration concepts

---
*CONFIG.md helps AI tools understand the configuration management layer that powers modern applications through ConfigHub.*