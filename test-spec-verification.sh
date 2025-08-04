#!/bin/bash
# Verify the implementation matches the specification

set -e

echo "🔍 GitHub Actions Bridge Specification Verification"
echo "=================================================="
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

# Test function
test_feature() {
    local name=$1
    local test_cmd=$2
    
    echo -n "Testing: $name... "
    if eval $test_cmd &> /dev/null; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((FAILED++))
    fi
}

# Phase 0: Act Validation
echo "Phase 0: Act Validation"
echo "----------------------"
test_feature "Act test binary exists" "test -f cmd/act-test/main.go"
test_feature "Act wrapper implementation" "test -f pkg/bridge/act_wrapper.go"
test_feature "Basic workflow fixture" "test -f test/fixtures/workflows/simple.yml"
echo

# Phase 1: Bridge Foundation
echo "Phase 1: Bridge Foundation"
echo "-------------------------"
test_feature "Bridge interface implementation" "grep -q 'func.*Info' pkg/bridge/actions_bridge.go"
test_feature "Apply method" "grep -q 'func.*Apply' pkg/bridge/actions_bridge.go"
test_feature "Refresh method" "grep -q 'func.*Refresh' pkg/bridge/actions_bridge.go"
test_feature "Destroy method" "grep -q 'func.*Destroy' pkg/bridge/actions_bridge.go"
test_feature "Import method" "grep -q 'func.*Import' pkg/bridge/actions_bridge.go"
test_feature "Finalize method" "grep -q 'func.*Finalize' pkg/bridge/actions_bridge.go"
test_feature "Workspace Manager" "test -f pkg/bridge/workspace_manager.go"
test_feature "Secure cleanup implementation" "grep -q 'SecureCleanup' pkg/bridge/workspace_manager.go"
test_feature "Compatibility checker" "test -f pkg/bridge/compatibility.go"
echo

# Phase 2: ConfigHub Integration
echo "Phase 2: ConfigHub Integration"
echo "-----------------------------"
test_feature "Config injection" "test -f pkg/bridge/config_injector.go"
test_feature "Secret handler" "test -f pkg/bridge/secret_handler.go"
test_feature "File-based secrets" "grep -q 'prepareSecrets' pkg/bridge/secret_handler.go"
test_feature "Leak detector" "grep -q 'LeakDetector' pkg/bridge/secret_handler.go"
test_feature "Multiple config formats" "grep -q 'ConfigFormat' pkg/bridge/config_injector.go"
echo

# Phase 3: Advanced Features
echo "Phase 3: Advanced Features"
echo "-------------------------"
test_feature "Execution context" "grep -q 'ExecutionContext' pkg/bridge/act_wrapper.go"
test_feature "GitHub context simulation" "grep -q 'PrepareEvent' pkg/bridge/act_wrapper.go"
test_feature "Enhanced CLI" "test -f cmd/actions-cli/main.go"
test_feature "Health monitoring" "test -f pkg/bridge/health.go"
test_feature "Prometheus metrics" "grep -q 'prometheus' pkg/bridge/health.go"
echo

# Key Architecture Components
echo "Key Architecture Components"
echo "--------------------------"
test_feature "Main bridge executable" "test -f cmd/actions-bridge/main.go"
test_feature "CLI tool" "test -f cmd/actions-cli/main.go"
test_feature "Integration tests" "test -f test/integration/bridge_test.go"
test_feature "Makefile" "test -f Makefile"
test_feature "Dockerfile" "test -f Dockerfile"
test_feature "Docker Compose" "test -f docker-compose.yml"
echo

# Documentation
echo "Documentation"
echo "------------"
test_feature "README.md" "test -f README.md"
test_feature "Example workflows" "test -d test/fixtures/workflows"
test_feature "Environment example" "test -f .env.example"
test_feature "Quick start script" "test -f quickstart.sh"
echo

# Specification Requirements
echo "Specification Requirements"
echo "-------------------------"
test_feature "Workspace isolation per execution" "grep -q 'CreateWorkspace' pkg/bridge/workspace_manager.go"
test_feature "Never use env vars for secrets" "grep -q 'secretsPath' pkg/bridge/actions_bridge.go"
test_feature "Full bridge pattern" "grep -q 'BridgeWorker' pkg/bridge/actions_bridge.go"
test_feature "Act limitations documented" "grep -q 'KnownLimitations' pkg/bridge/compatibility.go"
test_feature "Secure delete implementation" "grep -q 'secureDelete' pkg/bridge/workspace_manager.go"
test_feature "Auto-cleanup timeout" "grep -q 'time.Sleep' pkg/bridge/workspace_manager.go"
echo

# Summary
echo "======================================"
echo -e "Total Tests: $((PASSED + FAILED))"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All specification requirements verified!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some specification requirements not met${NC}"
    exit 1
fi
