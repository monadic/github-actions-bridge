.PHONY: all build test clean install docker act-test

# Variables
BINARY_NAME=actions-bridge
WORKER_NAME=cub-worker-actions
GO=go
GOFLAGS=-v
DOCKER_IMAGE=confighub/actions-bridge
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Default target
all: build

# Build all binaries
build: build-bridge build-worker build-act-test

# Build the main bridge
build-bridge:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/actions-bridge

# Build the worker (formerly CLI)
build-worker:
	@echo "Building $(WORKER_NAME)..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(WORKER_NAME) ./cmd/actions-cli

# Build the act test
build-act-test:
	@echo "Building act-test..."
	$(GO) build $(GOFLAGS) -o bin/act-test ./cmd/act-test

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GO) test -v -tags=integration ./test/integration/...

# Run act validation test (Phase 0)
act-test: build-act-test
	@echo "Running act validation..."
	./bin/act-test

# Install binaries
install: build
	@echo "Installing binaries..."
	$(GO) install ./cmd/actions-bridge
	@echo "Note: The worker binary $(WORKER_NAME) should be deployed as a ConfigHub worker"

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf /tmp/act-validation
	rm -rf /tmp/actions-cli
	rm -rf /tmp/actions-bridge

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Installing tools..."
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Lint code
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run bridge locally
run-bridge: build-bridge
	@echo "Running bridge..."
	ACTIONS_BRIDGE_BASE_DIR=/tmp/actions-bridge ./bin/$(BINARY_NAME)

# Run a quick example
example: build-worker
	@echo "Note: To run workflows, use the ConfigHub CLI:"
	@echo "  cub unit create --space dev example test/fixtures/workflows/simple.yml"
	@echo "  cub unit apply --space dev example"

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	$(GO) generate ./...

# Check dependencies
deps:
	@echo "Checking dependencies..."
	$(GO) mod verify
	$(GO) mod tidy -v

# Build release artifacts
release: clean
	@echo "Building release artifacts..."
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-linux-amd64 ./cmd/actions-bridge
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) \
		-o dist/$(WORKER_NAME)-linux-amd64 ./cmd/actions-cli
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-linux-arm64 ./cmd/actions-bridge
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) \
		-o dist/$(WORKER_NAME)-linux-arm64 ./cmd/actions-cli
	# Darwin AMD64
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/actions-bridge
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) \
		-o dist/$(WORKER_NAME)-darwin-amd64 ./cmd/actions-cli
	# Darwin ARM64
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) \
		-o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/actions-bridge
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) \
		-o dist/$(WORKER_NAME)-darwin-arm64 ./cmd/actions-cli
	@echo "Release artifacts built in dist/"

# Help
help:
	@echo "GitHub Actions Bridge - Makefile targets:"
	@echo ""
	@echo "  make build          - Build all binaries"
	@echo "  make test          - Run tests"
	@echo "  make act-test      - Run Phase 0 act validation"
	@echo "  make docker        - Build Docker image"
	@echo "  make install       - Install binaries"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make run-bridge    - Run bridge locally"
	@echo "  make example       - Run example workflow"
	@echo "  make release       - Build release artifacts"
	@echo ""
	@echo "Development:"
	@echo "  make dev-setup     - Setup development environment"
	@echo "  make lint          - Run linter"
	@echo "  make fmt           - Format code"
	@echo "  make deps          - Check dependencies"