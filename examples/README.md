# GitHub Actions Bridge Examples

[← Back to README](../README.md) | [← Back to User Guide](../USER_GUIDE.md)

---

This directory contains 15+ example workflows that demonstrate various features of the GitHub Actions Bridge. Each example includes a problem it solves and when to use it.

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
8. **[artifact-handling.yml](artifact-handling.yml)** - Working with artifacts
9. **[docker-compose.yml](docker-compose.yml)** - Integration with Docker Compose
10. **[file-persistence.yml](file-persistence.yml)** - Persistent file handling between workflow runs

### ConfigHub Integration Examples
11. **[config-driven-deployment.yml](config-driven-deployment.yml)** - Deploy using ConfigHub configurations
12. **[time-travel-testing.yml](time-travel-testing.yml)** - Test workflows with historical configurations
13. **[config-triggered-workflow.yml](config-triggered-workflow.yml)** - Workflows triggered by configuration changes
14. **[workflow-diff-testing.yml](workflow-diff-testing.yml)** - Compare different workflow versions
15. **[gitops-preview.yml](gitops-preview.yml)** - Preview GitOps changes without GitHub

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

### 8. artifact-handling.yml
**Problem:** Artifacts in GitHub Actions are not easily accessible in local testing.  
**Solution:** Shows how the bridge handles artifact creation, upload, and download locally.  
**Use When:** Workflows that build binaries, generate reports, or pass data between jobs via artifacts.

### 9. docker-compose.yml
**Problem:** Testing workflows that require multiple services (databases, caches) is complex.  
**Solution:** Demonstrates integration with Docker Compose for realistic local testing environments.  
**Use When:** Testing applications that require PostgreSQL, Redis, or other services.

### 10. file-persistence.yml
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

### 15. gitops-preview.yml
**Problem:** GitOps typically requires git branches and pull requests, adding complexity and delay.  
**Solution:** Preview and apply GitOps changes using ConfigHub spaces as logical environments, no git branches needed.  
**Use When:** Implementing GitOps workflows, promoting configurations between environments, or syncing spaces.

## Running Examples

### Basic Usage

```bash
# Run a simple example
cub-actions run examples/hello-world.yml

# Run with verbose output
cub-actions run examples/hello-world.yml -v

# Dry run to see what would happen
cub-actions run examples/hello-world.yml --dry-run
```

### With Secrets

```bash
# Create a secrets file
echo "API_KEY=your-secret-key" > secrets.env
echo "DATABASE_URL=postgres://localhost/db" >> secrets.env

# Run workflow with secrets
cub-actions run examples/with-secrets.yml --secrets-file secrets.env
```

### With ConfigHub Integration

```bash
# Run with ConfigHub space and unit
cub-actions run examples/config-driven-deployment.yml \
  --space production \
  --unit webapp

# Time travel testing
cub-actions run examples/time-travel-testing.yml \
  --as-of "2024-01-01" \
  --space staging

# Preview GitOps changes
cub-actions run examples/gitops-preview.yml \
  --source-space development \
  --target-space production \
  --dry-run
```

### Advanced Features

```bash
# Compare workflow versions
cub-actions diff examples/workflow-diff-testing.yml \
  --version proposed \
  --space production

# Test config-triggered workflows
cub-actions simulate-trigger examples/config-triggered-workflow.yml \
  --event config.changed \
  --path spec.replicas \
  --old-value 3 \
  --new-value 5
```

## Tips

1. **Validate First**: Always validate your workflow before running:
   ```bash
   cub-actions validate examples/your-workflow.yml
   ```

2. **Check Limitations**: Some GitHub Actions features don't work locally:
   ```bash
   cub-actions list-limitations
   ```

3. **Debug Issues**: Use verbose mode to see detailed execution:
   ```bash
   cub-actions run examples/your-workflow.yml -v
   ```

4. **Test with Different Configs**: Use ConfigHub spaces to test with various configurations:
   ```bash
   cub-actions run examples/config-driven-deployment.yml --space staging
   cub-actions run examples/config-driven-deployment.yml --space production
   ```

5. **Preview Changes**: Always preview before applying:
   ```bash
   cub-actions run examples/gitops-preview.yml --dry-run
   ```

## Understanding the ConfigHub Advantage

These examples demonstrate why combining GitHub Actions with ConfigHub is powerful:

1. **Test Locally with Real Configs**: No more "works on my machine" - test with actual configurations
2. **Time Travel**: Debug historical issues and predict future behavior
3. **Configuration as Code**: Treat workflows as configuration that can be versioned and tested
4. **Secure Secrets**: Never expose secrets in workflows or logs
5. **GitOps Without Git**: Implement GitOps patterns without complex branching strategies

Each example builds on these concepts, showing practical solutions to real-world CI/CD challenges.