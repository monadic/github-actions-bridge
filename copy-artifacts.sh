#!/bin/bash
clear
echo "📋 COPY THESE ARTIFACTS FROM CLAUDE:"
echo "===================================="
echo
echo "Click each artifact name in Claude, press Ctrl+A, Ctrl+C, then paste into the file:"
echo
artifacts=(
  "go-mod-file                → go.mod"
  "actions-bridge-impl        → pkg/bridge/actions_bridge.go"
  "workspace-manager          → pkg/bridge/workspace_manager.go"
  "act-wrapper               → pkg/bridge/act_wrapper.go"
  "compatibility-checker      → pkg/bridge/compatibility.go"
  "config-injector           → pkg/bridge/config_injector.go"
  "secret-handler            → pkg/bridge/secret_handler.go"
  "health-monitoring         → pkg/bridge/health.go"
  "leak-detector-pkg         → pkg/leakdetector/detector.go"
  "act-test-main            → cmd/act-test/main.go"
  "actions-bridge-main      → cmd/actions-bridge/main.go"
  "actions-cli              → cmd/actions-cli/main.go"
  "makefile                 → Makefile"
  "dockerfile               → Dockerfile"
  "docker-compose           → docker-compose.yml"
  "docker-config            → docker/config.yaml"
  "test-workflow-simple     → test/fixtures/workflows/simple.yml"
  "test-workflow-secrets    → test/fixtures/workflows/with-secrets.yml"
  "test-workflow-complex    → test/fixtures/workflows/complex.yml"
  "integration-tests        → test/integration/bridge_test.go"
  "test-spec-verification   → test/verify_spec.sh"
  "env-example              → .env.example"
  "gitignore                → .gitignore"
  "quickstart-script        → quickstart.sh"
  "license                  → LICENSE"
  "readme                   → README.md"
)

for i in "${!artifacts[@]}"; do
  printf "[%2d] ${artifacts[$i]}\n" $((i+1))
done

echo
echo "✅ After copying all files, run: make build"
