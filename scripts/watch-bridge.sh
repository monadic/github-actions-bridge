#!/bin/bash

# Watch bridge logs with better visibility
# Usage: ./scripts/watch-bridge.sh

echo "=== GitHub Actions Bridge Log Monitor ==="
echo "Filtering out heartbeats for clarity"
echo "Press Ctrl+C to stop"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Find the most recent log file
LOG_FILE="${1:-bridge.log}"
if [ ! -f "$LOG_FILE" ]; then
    LOG_FILE="bridge-new.log"
fi

if [ ! -f "$LOG_FILE" ]; then
    echo -e "${RED}Error: No log file found${NC}"
    echo "Usage: $0 [logfile]"
    exit 1
fi

echo "Watching: $LOG_FILE"
echo "========================================"

# Watch the log file, filtering and highlighting important events
tail -f "$LOG_FILE" | while read line; do
    # Skip heartbeat messages
    if echo "$line" | grep -q "heartbeat\|Heartbeat"; then
        continue
    fi
    
    # Highlight different types of messages
    if echo "$line" | grep -q "ERROR\|error\|failed"; then
        echo -e "${RED}$line${NC}"
    elif echo "$line" | grep -q "Successfully\|completed\|SUCCESS"; then
        echo -e "${GREEN}$line${NC}"
    elif echo "$line" | grep -q "Received APPLY\|Processing bridge event\|Queueing bridge"; then
        echo -e "${YELLOW}$line${NC}"
    else
        echo "$line"
    fi
done