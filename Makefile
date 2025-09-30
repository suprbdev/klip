# klip - A tiny cross-platform clipboard manager
# Makefile for building, testing, and managing the project

# Variables
BINARY_NAME=klip
MAIN_PACKAGE=./cmd/klip
BUILD_DIR=build
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(MAIN_PACKAGE)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linting
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running go vet instead..."; \
		go vet ./...; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run a quick test of the built binary
.PHONY: test-binary
test-binary: build
	@echo "Testing the built binary..."
	@$(BUILD_DIR)/$(BINARY_NAME) --help 2>/dev/null || echo "Binary built successfully (no --help flag available)"
	@echo "Testing basic functionality..."
	@$(BUILD_DIR)/$(BINARY_NAME) set --no-clip test-makefile "Hello from Makefile"
	@$(BUILD_DIR)/$(BINARY_NAME) get test-makefile --no-clip --raw
	@$(BUILD_DIR)/$(BINARY_NAME) rm test-makefile
	@echo "Basic functionality test passed!"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Clean everything including Go cache
.PHONY: clean-all
clean-all: clean
	@echo "Cleaning Go cache..."
	go clean -cache -modcache -testcache
	@echo "Full clean complete"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary (default)"
	@echo "  build-all     - Build for all platforms (Linux, macOS, Windows)"
	@echo "  build-linux   - Build for Linux (amd64, arm64)"
	@echo "  build-darwin  - Build for macOS (amd64, arm64)"
	@echo "  build-windows - Build for Windows (amd64)"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-binary   - Test the built binary functionality"
	@echo "  lint          - Run linter (golangci-lint or go vet)"
	@echo "  fmt           - Format code"
	@echo "  clean         - Remove build artifacts"
	@echo "  clean-all     - Remove build artifacts and Go cache"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION       - Set version string (default: git describe or 'dev')"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=v1.0.0"
	@echo "  make test-coverage"
	@echo "  make build-all"
