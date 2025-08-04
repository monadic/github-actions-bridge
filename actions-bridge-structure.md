# GitHub Actions Bridge - Project Structure

```
github-actions-bridge/
├── cmd/
│   ├── act-test/          # Phase 0 validation
│   │   └── main.go
│   ├── actions-bridge/    # Main bridge executable
│   │   └── main.go
│   └── actions-cli/       # CLI tool
│       └── main.go
├── pkg/
│   ├── bridge/
│   │   ├── actions_bridge.go      # Main bridge implementation
│   │   ├── workspace_manager.go   # Workspace isolation
│   │   ├── act_wrapper.go         # Act integration
│   │   ├── compatibility.go       # Compatibility checker
│   │   ├── config_injector.go     # Config injection
│   │   ├── secret_handler.go      # Secret management
│   │   ├── execution_context.go   # Execution context
│   │   └── health.go              # Health monitoring
│   └── leakdetector/
│       └── detector.go            # Secret leak detection
├── test/
│   ├── fixtures/
│   │   └── workflows/
│   │       ├── simple.yml
│   │       ├── with-secrets.yml
│   │       └── complex.yml
│   └── integration/
│       ├── bridge_test.go
│       ├── workspace_test.go
│       └── act_test.go
├── go.mod
├── go.sum
├── Dockerfile
├── Makefile
└── README.md
```
