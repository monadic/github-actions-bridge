#!/bin/bash
# GitHub Actions Bridge Quick Start Script

set -e

echo "🚀 GitHub Actions Bridge Quick Start"
echo "===================================="
echo

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}✗ $1 is not installed${NC}"
        return 1
    else
        echo -e "${GREEN}✓ $1 is installed${NC}"
        return 0
    fi
}

echo "Checking prerequisites..."
MISSING_DEPS=0

check_command go || MISSING_DEPS=1
check_command docker || MISSING_DEPS=1
check_command cub || echo -e "${YELLOW}  (ConfigHub CLI recommended but not required)${NC}"

if [ $MISSING_DEPS -eq 1 ]; then
    echo
    echo -e "${RED}Please install missing dependencies before continuing.${NC}"
    exit 1
fi

# Check Docker is running
if ! docker info &> /dev/null; then
    echo -e "${RED}✗ Docker is not running${NC}"
    echo "Please start Docker and try again."
    exit 1
else
    echo -e "${GREEN}✓ Docker is running${NC}"
fi

echo

# Setup .env if it doesn't exist
if [ ! -f .env ]; then
    echo "Setting up environment..."
    cp .env.example .env
    echo -e "${YELLOW}Created .env file from template${NC}"
    echo
    echo "Please edit .env and add your ConfigHub credentials:"
    echo "  CONFIGHUB_WORKER_ID=your-worker-id"
    echo "  CONFIGHUB_WORKER_SECRET=your-worker-secret"
    echo
    read -p "Press Enter once you've updated .env file..."
fi

# Load environment
source .env

# Verify credentials
if [ -z "$CONFIGHUB_WORKER_ID" ] || [ "$CONFIGHUB_WORKER_ID" == "your-worker-id-here" ]; then
    echo -e "${RED}✗ CONFIGHUB_WORKER_ID not set in .env${NC}"
    exit 1
fi

if [ -z "$CONFIGHUB_WORKER_SECRET" ] || [ "$CONFIGHUB_WORKER_SECRET" == "your-worker-secret-here" ]; then
    echo -e "${RED}✗ CONFIGHUB_WORKER_SECRET not set in .env${NC}"
    exit 1
fi

echo -e "${GREEN}✓ ConfigHub credentials configured${NC}"
echo

# Build or run?
echo "Choose an option:"
echo "1) Build from source"
echo "2) Run with Docker Compose"
echo "3) Run act validation test"
echo
read -p "Enter your choice (1-3): " choice

case $choice in
    1)
        echo
        echo "Building from source..."
        make build
        
        echo
        echo -e "${GREEN}✓ Build complete!${NC}"
        echo
        echo "To run the bridge:"
        echo "  ./bin/actions-bridge"
        echo
        echo "To run a workflow:"
        echo "  ./bin/cub-actions run .github/workflows/deploy.yml --space staging --unit webapp"
        ;;
        
    2)
        echo
        echo "Starting with Docker Compose..."
        docker-compose up -d
        
        echo
        echo "Waiting for bridge to be ready..."
        sleep 5
        
        # Check health
        if curl -s http://localhost:8080/health > /dev/null; then
            echo -e "${GREEN}✓ Bridge is running!${NC}"
            echo
            echo "Services:"
            echo "  - Bridge: http://localhost:8080/health"
            echo "  - Metrics: http://localhost:8080/metrics"
            
            if docker-compose ps | grep -q prometheus; then
                echo "  - Prometheus: http://localhost:9090"
            fi
            
            if docker-compose ps | grep -q grafana; then
                echo "  - Grafana: http://localhost:3000 (admin/admin)"
            fi
        else
            echo -e "${RED}✗ Bridge failed to start${NC}"
            echo "Check logs with: docker-compose logs actions-bridge"
            exit 1
        fi
        ;;
        
    3)
        echo
        echo "Running act validation test..."
        make act-test
        ;;
        
    *)
        echo -e "${RED}Invalid choice${NC}"
        exit 1
        ;;
esac

echo
echo "📚 Next steps:"
echo "  - Read the documentation: README.md"
echo "  - Try example workflows: test/fixtures/workflows/"
echo "  - Run 'cub-actions --help' for CLI usage"
echo
echo "Happy workflow testing! 🎉"
