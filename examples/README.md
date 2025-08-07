# GitHub Actions Bridge Examples

[← Back to README](../README.md) | [← Back to User Guide](../USER_GUIDE.md)

---

This directory contains 17 example workflows that demonstrate various features of the GitHub Actions Bridge. Each example includes a problem it solves and when to use it.

## 🚀 Which Tool Should I Use?

**Not sure which examples work with which tool?** → Check our [Examples Compatibility Guide](../EXAMPLES_COMPATIBILITY.md)

- **For local testing**: Use `./bin/cub-local-actions run <example.yml>`
- **For ConfigHub**: Use `cub unit create` and `cub unit apply`

## Quick Start

If you're new here, start with these examples in order:
1. [hello-world.yml](hello-world.yml) - Verify your setup works
2. [with-secrets.yml](with-secrets.yml) - Learn secure secret handling
3. [build-test-deploy.yml](build-test-deploy.yml) - See a complete CI/CD pipeline

Then explore the advanced ConfigHub integration examples for powerful features like time-travel testing and configuration-driven deployments.

## Examples Overview

### Basic Examples
1. **[hello-world.yml](hello-world.yml)** - A simple workflow to test basic functionality
2. **[with-secrets.yml](with-secrets.yml)** - Demonstrates secure secret handling
3. **[environment-variables.yml](environment-variables.yml)** - Shows how to work with environment variables
4. **[multi-job.yml](multi-job.yml)** - Example with multiple dependent jobs

### Advanced Examples
5. **[build-test-deploy.yml](build-test-deploy.yml)** - Complete CI/CD pipeline example
6. **[matrix-builds.yml](matrix-builds.yml)** - Demonstrates matrix strategy for testing multiple versions
7. **[conditional-execution.yml](conditional-execution.yml)** - Shows conditional logic and expressions
8. **[artifact-handling-improved.yml](artifact-handling-improved.yml)** - Working with artifacts (local-compatible version)
9. **[docker-compose-improved.yml](docker-compose-improved.yml)** - Integration with Docker Compose
10. **[file-persistence-improved.yml](file-persistence-improved.yml)** - Persistent file handling between workflow runs

### ConfigHub Integration Examples
11. **[config-driven-deployment.yml](config-driven-deployment.yml)** - Deploy using ConfigHub configurations
12. **[time-travel-testing.yml](time-travel-testing.yml)** - Test workflows with historical configurations
13. **[config-triggered-workflow.yml](config-triggered-workflow.yml)** - Workflows triggered by configuration changes
14. **[workflow-diff-testing.yml](workflow-diff-testing.yml)** - Compare different workflow versions
15. **[gitops-preview-improved.yml](gitops-preview-improved.yml)** - Preview GitOps changes without GitHub

### AI Integration Examples
16. **[claude-orchestrated-ops.yml](claude-orchestrated-ops.yml)** - Claude orchestrates operations using ConfigHub
17. **[worker-calls-claude.yml](worker-calls-claude.yml)** - Workers consult Claude for intelligent decisions

## Why Each Example Matters

### 1. hello-world.yml
**Problem:** Need to verify the GitHub Actions Bridge is working correctly.  
**Solution:** This minimal workflow tests basic execution, ensuring the bridge can run workflows locally.  
**Use When:** Setting up the bridge for the first time or troubleshooting basic connectivity.

### 2. with-secrets.yml
**Problem:** Secrets in GitHub Actions are not available locally, making it impossible to test workflows that depend on them.  
**Solution:** Demonstrates how ConfigHub injects secrets at runtime, allowing secure local testing without exposing credentials.  
**Use When:** Testing workflows that interact with databases, APIs, or other services requiring authentication.

### 3. environment-variables.yml
**Problem:** Environment-specific variables differ between local and GitHub environments.  
**Solution:** Shows how to manage environment variables consistently across local and remote execution.  
**Use When:** Building workflows that need to adapt to different environments (dev, staging, prod).

### 4. multi-job.yml
**Problem:** Complex workflows with job dependencies are hard to test locally.  
**Solution:** Demonstrates job orchestration, dependencies, and data passing between jobs.  
**Use When:** Building pipelines where later stages depend on earlier results.

### 5. build-test-deploy.yml
**Problem:** Full CI/CD pipelines are risky to test in production.  
**Solution:** Complete example showing build, test, and deployment stages that can be safely tested locally.  
**Use When:** Implementing or modifying deployment pipelines.

### 6. matrix-builds.yml
**Problem:** Testing across multiple versions/platforms is expensive and time-consuming in CI.  
**Solution:** Run matrix builds locally to validate compatibility before pushing.  
**Use When:** Supporting multiple Node/Python/Go versions or operating systems.

### 7. conditional-execution.yml
**Problem:** Complex conditional logic is hard to test without triggering actual conditions.  
**Solution:** Test different execution paths by simulating various conditions locally.  
**Use When:** Building workflows with branching logic based on events, inputs, or repository state.

### 8. artifact-handling-improved.yml
**Problem:** Artifacts in GitHub Actions are not easily accessible in local testing.  
**Solution:** Shows how the bridge handles artifact creation, upload, and download locally.  
**Use When:** Workflows that build binaries, generate reports, or pass data between jobs via artifacts.

### 9. docker-compose-improved.yml
**Problem:** Testing workflows that require multiple services (databases, caches) is complex.  
**Solution:** Demonstrates integration with Docker Compose for realistic local testing environments.  
**Use When:** Testing applications that require PostgreSQL, Redis, or other services.

### 10. file-persistence-improved.yml
**Problem:** Workflows that generate configuration files need persistent storage between runs.  
**Solution:** Shows how to persist files across workflow executions, mimicking the custom-bridge pattern.  
**Use When:** Building workflows that maintain state or generate incremental configurations.

### 11. config-driven-deployment.yml
**Problem:** Traditional GitHub Actions mix configuration with workflow logic, making it hard to test with different configs.  
**Solution:** Demonstrates pure configuration-driven deployments where all values come from ConfigHub.  
**Use When:** You want to test deployments with production-like configurations without modifying workflow files.

### 12. time-travel-testing.yml
**Problem:** "What would have happened if we ran this deployment last week?" - Impossible to answer with traditional CI/CD.  
**Solution:** Test workflows with historical configurations to understand past behavior or predict future changes.  
**Use When:** Debugging issues ("it worked yesterday"), planning changes, or auditing historical deployments.

### 13. config-triggered-workflow.yml
**Problem:** Configuration changes often require manual workflow triggers, leading to drift between config and deployment.  
**Solution:** Automatically trigger workflows when configurations change, ensuring configs and deployments stay in sync.  
**Use When:** Implementing true GitOps where configuration changes automatically trigger appropriate workflows.

### 14. workflow-diff-testing.yml
**Problem:** Changing workflows is risky - you don't know the impact until it runs in production.  
**Solution:** Compare workflow versions side-by-side, seeing exactly what would change in behavior, timing, and resources.  
**Use When:** Modifying critical workflows, adding new features, or evaluating workflow optimization proposals.

### 15. gitops-preview-improved.yml
**Problem:** GitOps typically requires git branches and pull requests, adding complexity and delay.  
**Solution:** Preview and apply GitOps changes using ConfigHub spaces as logical environments, no git branches needed.  
**Use When:** Implementing GitOps workflows, promoting configurations between environments, or syncing spaces.

### 16. claude-orchestrated-ops.yml
**Problem:** Complex operational decisions require human expertise, causing delays and inconsistency.  
**Solution:** Claude AI orchestrates operations by analyzing requests, making decisions, and updating ConfigHub with full audit trails.  
**Use When:** Automating operational workflows, scaling decisions, or handling complex deployment scenarios with AI assistance.

### 17. worker-calls-claude.yml
**Problem:** Workers encounter situations requiring intelligent decision-making beyond simple rules.  
**Solution:** Workers can consult Claude for real-time advice on deployments, using ConfigHub to track all decisions and maintain audit trails.  
**Use When:** Deployment decisions need context-aware intelligence, anomaly detection requires expert analysis, or risk assessment needs AI assistance.

## Running Examples

### Basic Usage

```bash
# Run a simple example
cub unit create --space default hello examples/hello-world.yml --target docker-desktop
cub unit apply --space default hello

# Run with verbose output
cub unit apply --space default hello --debug

# Dry run to see what would happen
cub unit create --space default hello examples/hello-world.yml --target docker-desktop --dry-run
```

### With Secrets

```bash
# Create a secrets file (for local testing)
echo "API_KEY=your-secret-key" > secrets.env
echo "DATABASE_URL=postgres://localhost/db" >> secrets.env

# Run workflow with ConfigHub managing secrets
cub unit create --space default secrets-demo examples/with-secrets.yml --target docker-desktop
cub unit apply --space default secrets-demo
# Secrets are securely managed by ConfigHub
```

### With ConfigHub Integration

```bash
# Run with ConfigHub space and unit
cub unit create --space production webapp examples/config-driven-deployment.yml --target docker-desktop
cub unit apply --space production webapp

# Time travel testing
cub unit create --space staging time-travel examples/time-travel-testing.yml --target docker-desktop
cub unit apply --space staging time-travel --restore 1

# Preview GitOps changes
cub unit create --space production gitops examples/gitops-preview-improved.yml --target docker-desktop
cub unit get --space production gitops --extended
# Use ConfigHub's space management for GitOps workflows
```

### Advanced Features

```bash
# Compare workflow versions
cub revision list --space production --where "UnitSlug = 'workflow-diff'"
cub revision diff --space production workflow-diff --from 1 --to 2

# Test config-triggered workflows
cub unit create --space production config-trigger examples/config-triggered-workflow.yml --target docker-desktop
cub trigger create --space production config-change \
  --unit config-trigger \
  --event "config.changed"
```

### AI Integration Examples

```bash
# Claude orchestrates operations
cub unit create --space production claude-ops examples/claude-orchestrated-ops.yml --target docker-desktop
# Set the operation request in ConfigHub
cub unit set --space production claude-ops \
  --key operation-request \
  --value "Check system health and scale if needed"
cub unit apply --space production claude-ops

# Worker asks Claude for deployment advice
cub unit create --space production claude-worker examples/worker-calls-claude.yml --target docker-desktop
cub unit set --space production claude-worker \
  --key deployment-stage --value production \
  --key anomaly-type --value high-error-rate
cub unit apply --space production claude-worker
```

## Tips

1. **Validate First**: Always validate your workflow before running:
   ```bash
   cub unit create --space default test examples/your-workflow.yml --target docker-desktop --dry-run
   ```

2. **Check Limitations**: Some GitHub Actions features don't work locally:
   ```bash
   # Check ConfigHub and act documentation for limitations
   cub help
   ```

3. **Debug Issues**: Use verbose mode to see detailed execution:
   ```bash
   cub unit apply --space default your-workflow --debug
   ```

4. **Test with Different Configs**: Use ConfigHub spaces to test with various configurations:
   ```bash
   cub unit create --space staging deploy examples/config-driven-deployment.yml --target docker-desktop
   cub unit create --space production deploy examples/config-driven-deployment.yml --target docker-desktop
   cub unit apply --space staging deploy
   cub unit apply --space production deploy
   ```

5. **Preview Changes**: Always preview before applying:
   ```bash
   cub unit create --space default preview examples/gitops-preview-improved.yml --target docker-desktop --dry-run
   cub unit get --space default preview --extended
   ```

## Understanding the ConfigHub Advantage

These examples demonstrate why combining GitHub Actions with ConfigHub is powerful:

1. **Test Locally with Real Configs**: No more "works on my machine" - test with actual configurations
2. **Time Travel**: Debug historical issues and predict future behavior
3. **Configuration as Code**: Treat workflows as configuration that can be versioned and tested
4. **Secure Secrets**: Never expose secrets in workflows or logs
5. **GitOps Without Git**: Implement GitOps patterns without complex branching strategies

Each example builds on these concepts, showing practical solutions to real-world CI/CD challenges.