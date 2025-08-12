#!/bin/bash

# Start the GitHub Actions Bridge with proper logging
# Usage: ./scripts/start-bridge.sh [worker-name]

set -e

WORKER_NAME="${1:-actions-bridge-1}"
LOG_FILE="bridge-$(date +%Y%m%d-%H%M%S).log"

echo "=== GitHub Actions Bridge Startup ==="
echo "Worker: $WORKER_NAME"
echo "Log file: $LOG_FILE"
echo ""

# Check prerequisites
echo "Checking prerequisites..."
if ! docker ps >/dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker Desktop first."
    exit 1
fi
echo "✅ Docker is running"

if ! command -v act >/dev/null 2>&1; then
    echo "❌ act is not installed. Please install with: brew install act"
    exit 1
fi
echo "✅ act is installed"

if ! command -v cub >/dev/null 2>&1; then
    echo "❌ cub CLI not found. Add to PATH: export PATH=\"\$HOME/.confighub/bin:\$PATH\""
    exit 1
fi
echo "✅ cub CLI found"

# Check authentication
echo ""
echo "Checking ConfigHub authentication..."
if ! cub context get >/dev/null 2>&1; then
    echo "❌ Not authenticated. Please run: cub auth login"
    exit 1
fi
echo "✅ Authenticated to ConfigHub"

# Get worker credentials
echo ""
echo "Getting worker credentials..."
if ! cub worker get "$WORKER_NAME" >/dev/null 2>&1; then
    echo "❌ Worker '$WORKER_NAME' not found"
    echo "Available workers:"
    cub worker list | head -10
    exit 1
fi

# Export credentials
eval "$(cub worker get-envs $WORKER_NAME)"
export CONFIGHUB_URL=https://hub.confighub.com

echo "✅ Worker credentials loaded"
echo "   ID: ${CONFIGHUB_WORKER_ID:0:20}..."

# Stop any existing bridge
if pgrep -f actions-bridge >/dev/null; then
    echo ""
    echo "Stopping existing bridge..."
    pkill -f actions-bridge
    sleep 2
fi

# Start the bridge
echo ""
echo "Starting bridge..."
nohup ./bin/actions-bridge > "$LOG_FILE" 2>&1 &
BRIDGE_PID=$!

echo "✅ Bridge started with PID: $BRIDGE_PID"

# Wait for bridge to be ready
echo ""
echo "Waiting for bridge to connect..."
sleep 5

# Check if bridge is running
if ! ps -p $BRIDGE_PID >/dev/null; then
    echo "❌ Bridge failed to start. Check logs:"
    tail -20 "$LOG_FILE"
    exit 1
fi

# Check health endpoint
if curl -s http://localhost:8080/health | grep -q "healthy"; then
    echo "✅ Bridge is healthy"
else
    echo "⚠️  Health check failed, but bridge is running"
fi

# Check worker status
WORKER_STATUS=$(cub worker get "$WORKER_NAME" | grep Condition | awk '{print $2}')
echo "✅ Worker status: $WORKER_STATUS"

echo ""
echo "=== Bridge Started Successfully ==="
echo "Monitor logs with: tail -f $LOG_FILE"
echo "Or use: ./scripts/watch-bridge.sh $LOG_FILE"
echo ""
echo "To test:"
echo "  cub unit create test examples/hello-world.yml --target docker-desktop"
echo "  cub unit apply test"