BINARY_NAME=mc-tool
BUILD_DIR=build

.PHONY: build build-static build-portable build-all clean install test help

# Default target
all: build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Build static binary for portability
build-static:
	@echo "Building static $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME)-static main.go
	@echo "Static build completed: $(BUILD_DIR)/$(BINARY_NAME)-static"

# Build portable binary (static + stripped for minimum size)
build-portable:
	@echo "Building portable $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static" -s -w' -o $(BUILD_DIR)/$(BINARY_NAME)-portable main.go
	@echo "Portable build completed: $(BUILD_DIR)/$(BINARY_NAME)-portable (stripped, maximum compatibility)"

# Build for multiple platforms
build-all:
	@echo "Building $(BINARY_NAME) for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux static
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	# macOS
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "Multi-platform builds completed in $(BUILD_DIR)/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
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
	go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) main.go

# Cross-compile for different platforms
build-all:
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "Cross-compilation completed"

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install/update dependencies"
	@echo "  test           - Run all tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage report (generates coverage.html)"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-bench     - Run benchmark tests"
	@echo "  test-single    - Run specific test (use TEST=TestName)"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  dev            - Build with debugging info"
	@echo "  build-all      - Cross-compile for multiple platforms"
	@echo "  help           - Show this help message"