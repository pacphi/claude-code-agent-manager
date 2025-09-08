.PHONY: build test install clean run help cross-compile benchmark benchmark-quick benchmark-profile benchmark-clean deps-upgrade deadcode ci

# Variables
BINARY_NAME := agent-manager
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)
GOFLAGS := -v

# Default target
all: build

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^## [a-zA-Z_-]+:' $(MAKEFILE_LIST) | \
		sed 's/^## \([^:]*\): \(.*\)/  \1|\2/' | \
		column -t -s '|' | \
		sed 's/^/  /'

## build: Build the agent-manager binary
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) cmd/agent-manager/main.go
	@echo "Build complete: bin/$(BINARY_NAME)"

## install: Install agent-manager to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp bin/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete: /usr/local/bin/$(BINARY_NAME)"

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v -cover ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## run: Build and run agent-manager with install command
run: build
	@echo "Running $(BINARY_NAME)..."
	@./bin/$(BINARY_NAME) install

## validate: Validate the configuration file
validate: build
	@echo "Validating configuration..."
	@./bin/$(BINARY_NAME) validate

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

## deps-upgrade: Upgrade all dependencies to their latest versions
deps-upgrade:
	@echo "Upgrading all dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies upgraded to latest versions"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

## lint: Run Go linter (Note: Linting is automatically run in CI via GitHub Actions)
lint:
	@echo "Running linter locally (optional - linting is handled in CI)..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo ""; \
		echo "golangci-lint is not installed locally."; \
		echo "This is OK - linting is automatically handled in GitHub Actions CI."; \
		echo ""; \
		echo "If you want to run linting locally, install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		echo ""; \
	fi

## vet: Run Go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## deadcode: Run deadcode analysis with test flag
deadcode:
	@echo "Running deadcode analysis..."
	@if command -v deadcode &> /dev/null; then \
		deadcode -test ./...; \
	else \
		echo ""; \
		echo "deadcode is not installed."; \
		echo "Install with: go install golang.org/x/tools/cmd/deadcode@latest"; \
		echo ""; \
		exit 1; \
	fi

## ci: Run all CI checks (format, vet, lint, deadcode, test)
ci: fmt vet lint deadcode test
	@echo "All CI checks completed successfully"

## cross-compile: Build binaries for multiple platforms
cross-compile:
	@echo "Building for multiple platforms..."
	@mkdir -p bin/

	@echo "Building for Darwin AMD64..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-amd64 cmd/agent-manager/main.go

	@echo "Building for Darwin ARM64..."
	@GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-arm64 cmd/agent-manager/main.go

	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-amd64 cmd/agent-manager/main.go

	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-arm64 cmd/agent-manager/main.go

	@echo "Building for Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-windows-amd64.exe cmd/agent-manager/main.go

	@echo "Cross-compilation complete. Binaries in bin/"

## release: Create release artifacts
release: clean cross-compile
	@echo "Creating release artifacts..."
	@mkdir -p releases/
	@for file in bin/*; do \
		if [ -f "$$file" ]; then \
			base=$$(basename $$file); \
			tar czf "releases/$$base-$(VERSION).tar.gz" -C bin/ $$base; \
			echo "Created: releases/$$base-$(VERSION).tar.gz"; \
		fi \
	done
	@echo "Release artifacts created in releases/"

## benchmark: Run performance benchmarks
benchmark:
	@echo "Running performance benchmarks..."
	@./scripts/benchmark.sh

## benchmark-quick: Run quick performance benchmarks
benchmark-quick:
	@echo "Running quick performance benchmarks..."
	@./scripts/benchmark.sh --quick

## benchmark-profile: Run benchmarks with detailed profiling
benchmark-profile:
	@echo "Running benchmarks with profiling..."
	@./scripts/benchmark.sh --profile

## benchmark-clean: Clean benchmark results
benchmark-clean:
	@echo "Cleaning benchmark results..."
	@./scripts/benchmark.sh --clean

## dev: Run in development mode with hot reload (requires entr)
dev:
	@if command -v entr &> /dev/null; then \
		find . -name "*.go" | entr -r make run; \
	else \
		echo "entr not installed. Install with: brew install entr (macOS) or apt-get install entr (Linux)"; \
		exit 1; \
	fi

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest
	@echo "Docker image built: $(BINARY_NAME):$(VERSION)"

# Check if all required tools are installed
check-tools:
	@echo "Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "Git is not installed"; exit 1; }
	@echo "All required tools are installed"

.DEFAULT_GOAL := help