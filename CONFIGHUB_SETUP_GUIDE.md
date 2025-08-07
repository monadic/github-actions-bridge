# ConfigHub Setup Guide for GitHub Actions Bridge

This guide will walk you through setting up ConfigHub to run GitHub Actions workflows using the `cub` CLI and actions-bridge worker.

## Prerequisites

Before starting, ensure you have:
- ✅ ConfigHub account (sign up at https://confighub.com)
- ✅ `cub` CLI installed (via the install script)
- ✅ `act` installed (required for running workflows)
- ✅ Docker running
- ✅ Built the actions-bridge binary (`make build`)

## Step-by-Step Setup Process

### Step 1: Authenticate with ConfigHub

First, make sure `cub` is in your PATH:

```bash
# Check if cub is available
which cub

# If not found, add it to your PATH:
export PATH="$HOME/.confighub/bin:$PATH"

# Make it permanent by adding to your shell profile:
echo 'export PATH="$HOME/.confighub/bin:$PATH"' >> ~/.zshrc  # or ~/.bashrc
source ~/.zshrc  # reload configuration
```

Now log in to ConfigHub. This will open your browser for authentication:

```bash
cub auth login
```

**What happens:**
1. Your browser opens to the ConfigHub login page
2. Enter your credentials and complete authentication
3. The browser will show "Authentication successful"
4. Return to your terminal

**Verify you're logged in:**
```bash
cub context get
```

You should see something like:
```
Organization: your-org
Space: default
User: you@example.com
```

**Troubleshooting:**
- If browser doesn't open: Copy the URL shown and paste it manually
- If authentication fails: Check your ConfigHub account is active
- If context shows no organization: Contact your ConfigHub admin

### Step 2: Create a Test Space

Create a dedicated space for testing:

```bash
# Create the space
cub space create alexis-actions-test

# Set it as your current space
cub context set --space alexis-actions-test

# Verify
cub context get
```

Expected output:
```
Organization: your-org
Space: alexis-actions-test
User: you@example.com
```

### Step 3: Create and Configure a Worker

The worker connects ConfigHub to your local GitHub Actions execution:

```bash
# Create a worker
cub worker create actions-bridge-1

# Get the worker credentials (IMPORTANT: Save these!)
cub worker get actions-bridge-1
```

This will show output like:
```
Worker ID: 12345678-1234-1234-1234-123456789012
Worker Secret: ****************************************
Status: created
```

### Step 4: Set Worker Environment Variables

Set up the environment for the worker:

```bash
# Option 1: Use the helper command
eval "$(cub worker get-envs actions-bridge-1)"

# Option 2: Set manually (replace with your actual values)
export CONFIGHUB_WORKER_ID="your-worker-id"
export CONFIGHUB_WORKER_SECRET="your-worker-secret"

# IMPORTANT: Use the correct URL - NOT api.confighub.com!
export CONFIGHUB_URL="https://hub.confighub.com"

# Verify environment is set
echo $CONFIGHUB_WORKER_ID
# Should show your worker ID
```

### Step 5: Start the Actions Bridge Worker

In a **new terminal window**, start the worker:

```bash
# Navigate to the project directory
cd /path/to/github-actions-bridge

# Start the worker
./bin/actions-bridge
```

You should see:
```
2025/08/07 10:30:45 Starting GitHub Actions Bridge worker...
2025/08/07 10:30:45 Worker ID: [your-worker-id]
2025/08/07 10:30:45 Connected to ConfigHub
2025/08/07 10:30:45 Listening for jobs...
```

**IMPORTANT: Keep this terminal running!**

### Step 6: Test with a Simple Workflow

Back in your original terminal, let's test with hello-world:

```bash
# IMPORTANT: You must specify a target when creating units!
# List available targets first
cub target list

# Create a unit with the docker-desktop target
cub unit create hello examples/hello-world.yml --target docker-desktop

# Run it
cub unit apply hello
```

**Critical Notes:**
- You MUST use `--target docker-desktop` (or `--target podman-local` if using Podman)
- Without a target, you'll get: "cannot invoke action on a unit without a target"
- The target connects your unit to the worker that will execute it

Watch the worker terminal - you should see:
```
2025/08/07 10:31:00 Received job: hello
2025/08/07 10:31:00 Executing workflow...
[Hello World/greet] 🏁 Job succeeded
2025/08/07 10:31:05 Job completed successfully
```

### Step 7: View Results

Check the execution results:

```bash
# View unit status
cub unit get hello

# View execution logs
cub unit get hello --extended
```

## Running More Examples

Now you can run any ConfigHub-compatible example:

```bash
# With secrets
cub unit create secrets-test examples/with-secrets.yml --target docker-desktop
cub unit apply secrets-test

# Multi-job workflow
cub unit create multi-job examples/multi-job.yml --target docker-desktop
cub unit apply multi-job

# Build and deploy pipeline
cub unit create pipeline examples/build-test-deploy.yml --target docker-desktop
cub unit apply pipeline
```

## Common Issues and Solutions

### "cub: command not found"
- The installer puts cub at `~/.confighub/bin/cub`
- Add to PATH: `export PATH="$HOME/.confighub/bin:$PATH"`
- Or use full path: `~/.confighub/bin/cub auth login`

### "cannot invoke action on a unit without a target"
- You MUST specify `--target` when creating units
- Use: `cub unit create <name> <file> --target docker-desktop`
- Check available targets: `cub target list`

### "dial tcp: lookup api.confighub.com: no such host"
- Wrong URL! Use `export CONFIGHUB_URL="https://hub.confighub.com"`
- NOT `api.confighub.com` - this doesn't exist
- The correct URL is `hub.confighub.com`

### "Worker not found" error
- Check the worker is still running in the other terminal
- Verify with: `cub worker list`
- Restart the worker if needed

### "Authentication failed"
- Run `cub auth login` again
- Check your ConfigHub account is active

### "Space not found"
- Verify current context: `cub context get`
- Set correct space: `cub context set --space alexis-actions-test`

### Worker crashes on startup
- Check Docker is running: `docker ps`
- Verify environment variables are set correctly
- Check the worker logs for specific errors

## Next Steps

1. **Explore advanced features:**
   - Time-travel testing with `--restore`
   - Configuration-driven deployments
   - GitOps workflows

2. **Set up production spaces:**
   - Create separate spaces for dev/staging/prod
   - Configure appropriate access controls

3. **Integrate with your workflows:**
   - Copy your `.github/workflows` files
   - Adapt them for ConfigHub execution

## Quick Reference

```bash
# Authentication
cub auth login                          # Log in
cub context get                         # Check current context

# Spaces
cub space create <name>                 # Create space
cub context set --space <name>          # Switch space

# Workers
cub worker create <name>                # Create worker
cub worker get-envs <name>              # Get env vars
cub worker list                         # List workers

# Units (workflows)
cub unit create <name> <file>           # Create unit
cub unit apply <name>                   # Run unit
cub unit get <name> --extended          # View logs

# Start worker
./bin/actions-bridge                    # Run in separate terminal
```

Remember: The worker must be running for ConfigHub to execute workflows!