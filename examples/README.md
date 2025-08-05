# GitHub Actions Bridge Examples

This directory contains example workflows that demonstrate various features of the GitHub Actions Bridge.

## Basic Examples

- **[hello-world.yml](hello-world.yml)** - Simple workflow to test your setup
- **[with-secrets.yml](with-secrets.yml)** - Using secrets in workflows
- **[environment-variables.yml](environment-variables.yml)** - Working with environment variables
- **[multi-job.yml](multi-job.yml)** - Workflow with multiple jobs

## Advanced Examples

- **[build-test-deploy.yml](build-test-deploy.yml)** - Complete CI/CD pipeline example
- **[matrix-builds.yml](matrix-builds.yml)** - Matrix strategy for testing multiple versions
- **[conditional-execution.yml](conditional-execution.yml)** - Using conditions and expressions
- **[artifact-handling.yml](artifact-handling.yml)** - Creating and using artifacts
- **[docker-compose.yml](docker-compose.yml)** - Working with Docker Compose
- **[file-persistence.yml](file-persistence.yml)** - Saving outputs to persistent files

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

### With Environment Variables

```bash
# Set environment variables and run
export ENVIRONMENT=production
export VERSION=1.2.3
cub-actions run examples/environment-variables.yml
```

## ConfigHub Integration Examples

These examples show how to use the bridge with ConfigHub:

- **[configmap-example.yml](configmap-example.yml)** - Generating Kubernetes ConfigMaps
- **[terraform-output.yml](terraform-output.yml)** - Capturing Terraform outputs
- **[config-validation.yml](config-validation.yml)** - Validating configuration files

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