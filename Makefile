BINARY_NAME=mc-tool
BUILD_DIR=build
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: build build-static build-portable build-all clean install test test-coverage test-race test-bench test-single dev deps help

# Default target
all: build-static

# Build the application (regular build)
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Build static binary for maximum portability
build-static:
	@echo "Building static $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME) -extldflags '-static'" -o $(BUILD_DIR)/$(BINARY_NAME)-static .
	@echo "Static build completed: $(BUILD_DIR)/$(BINARY_NAME)-static"
	@echo "Binary can run on any Linux system without dependencies"

# Build portable binary (static + stripped for minimum size)
build-portable:
	@echo "Building portable $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME) -extldflags '-static' -s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-portable .
	@echo "Portable build completed: $(BUILD_DIR)/$(BINARY_NAME)-portable"
	@echo "Binary is stripped and optimized for minimum size"

# Build for multiple platforms
build-all:
	@echo "Building $(BINARY_NAME) for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux static (recommended for servers)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME) -extldflags '-static'" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64-static .
	# Linux ARM64 (for ARM servers/containers)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME) -extldflags '-static'" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64-static .
	# macOS Intel
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	# macOS Apple Silicon
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Multi-platform builds completed in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f mc-tool mc-tool-*
	go clean

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test ./... -race -v

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	go test ./... -bench=. -benchmem

# Run specific test
test-single:
	@echo "Running specific test (use TEST=TestName)..."
	go test ./... -run $(TEST) -v

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	go install

# Development build with debugging info
dev: 
	@echo "Building $(BINARY_NAME) with debug info..."
	@mkdir -p $(BUILD_DIR)
	go build -gcflags="all=-N -l" $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-dev .

# Show build info
info:
	@echo "Build Information:"
	@echo "  Binary Name: $(BINARY_NAME)"
	@echo "  Version:     $(VERSION)"
	@echo "  Commit:      $(COMMIT)"
	@echo "  Build Time:  $(BUILD_TIME)"
	@echo "  Build Dir:   $(BUILD_DIR)"

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application (regular build)"
	@echo "  build-static   - Build static binary (no external dependencies)"
	@echo "  build-portable - Build portable binary (static + stripped)"
	@echo "  build-all      - Cross-compile for multiple platforms"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install/update dependencies"
	@echo "  test           - Run all tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage report (generates coverage.html)"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-bench     - Run benchmark tests"
	@echo "  test-single    - Run specific test (use TEST=TestName)"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  dev            - Build with debugging info"
	@echo "  info           - Show build information"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Recommended for deployment: make build-static"