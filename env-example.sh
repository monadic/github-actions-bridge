# GitHub Actions Bridge Environment Configuration
# Copy this file to .env and fill in your values

# Required: ConfigHub Worker Credentials
CONFIGHUB_WORKER_ID=your-worker-id-here
CONFIGHUB_WORKER_SECRET=your-worker-secret-here

# Optional: ConfigHub API URL (defaults to production)
CONFIGHUB_URL=https://api.confighub.com

# Optional: Bridge Configuration
ACTIONS_BRIDGE_BASE_DIR=/var/lib/actions-bridge
MAX_CONCURRENT_WORKFLOWS=5

# Optional: Act Configuration
ACT_DEFAULT_IMAGE=catthehacker/ubuntu:act-latest

# Optional: Debugging
DEBUG=false
LOG_LEVEL=info

# Optional: Resource Limits
DOCKER_CPU_LIMIT=2
DOCKER_MEMORY_LIMIT=4g

# Optional: Security
ENABLE_LEAK_DETECTION=true
SECURE_CLEANUP_PASSES=3

# Optional: Monitoring Ports
HEALTH_PORT=8080
METRICS_PORT=8080
