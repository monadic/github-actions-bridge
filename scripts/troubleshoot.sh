#!/bin/bash

# Troubleshooting script for GitHub Actions Bridge
# Usage: ./scripts/troubleshoot.sh

echo "=== GitHub Actions Bridge Troubleshooting ==="
echo "Timestamp: $(date)"
echo ""

# Function to check status
check_status() {
    if eval "$2" >/dev/null 2>&1; then
        echo "✅ $1"
        return 0
    else
        echo "❌ $1"
        return 1
    fi
}

# System checks
echo "System Checks:"
check_status "Docker running" "docker ps"
check_status "act installed" "act --version"
check_status "cub CLI available" "which cub"
check_status "Bridge binary exists" "test -f ./bin/actions-bridge"

# ConfigHub checks
echo ""
echo "ConfigHub Status:"
if check_status "Authenticated" "cub context get"; then
    echo "   $(cub context get | grep 'User Email' | head -1)"
    echo "   $(cub context get | grep 'Space' | head -1)"
fi

# Worker checks
echo ""
echo "Worker Configuration:"
if [ -n "$CONFIGHUB_WORKER_ID" ]; then
    echo "✅ CONFIGHUB_WORKER_ID set: ${CONFIGHUB_WORKER_ID:0:20}..."
else
    echo "❌ CONFIGHUB_WORKER_ID not set"
fi

if [ -n "$CONFIGHUB_URL" ]; then
    echo "✅ CONFIGHUB_URL: $CONFIGHUB_URL"
    if [ "$CONFIGHUB_URL" != "https://hub.confighub.com" ]; then
        echo "   ⚠️  WARNING: URL should be https://hub.confighub.com"
    fi
else
    echo "❌ CONFIGHUB_URL not set"
fi

# Bridge process check
echo ""
echo "Bridge Process:"
if pgrep -f actions-bridge >/dev/null; then
    PID=$(pgrep -f actions-bridge | head -1)
    echo "✅ Bridge running (PID: $PID)"
    
    # Check health endpoint
    if curl -s http://localhost:8080/health | grep -q "healthy"; then
        echo "✅ Health endpoint responding"
    else
        echo "❌ Health endpoint not responding"
    fi
else
    echo "❌ Bridge not running"
fi

# Target and worker mapping
echo ""
echo "Target Configuration:"
echo "Docker Desktop Target:"
cub target list 2>/dev/null | grep -E "(SLUG|docker-desktop)" | head -2

# Recent logs check
echo ""
echo "Recent Bridge Activity:"
if [ -f bridge.log ]; then
    echo "Last 5 non-heartbeat events:"
    grep -v heartbeat bridge.log | tail -5
elif [ -f bridge-new.log ]; then
    echo "Last 5 non-heartbeat events:"
    grep -v heartbeat bridge-new.log | tail -5
else
    echo "No log files found"
fi

# Common issues
echo ""
echo "Common Issues to Check:"
echo "1. Worker-Target Mismatch: Ensure you're using the worker shown in 'cub target list'"
echo "2. Docker Not Running: Start Docker Desktop"
echo "3. Wrong ConfigHub URL: Must be https://hub.confighub.com"
echo "4. Expired Token: Run 'cub auth login' again"
echo "5. No Target Specified: Always use --target docker-desktop"

echo ""
echo "=== End of Troubleshooting ==="