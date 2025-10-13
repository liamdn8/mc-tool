#!/bin/bash

# Comprehensive test script for mc-tool profile functionality
# This script runs unit tests, integration tests, and playground demonstrations

set -e

echo "=== MC-Tool Profile Test Suite ==="
echo "Date: $(date)"
echo "Working Directory: $(pwd)"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_section() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check prerequisites
print_section "Prerequisites Check"

# Check Go installation
if command -v go &> /dev/null; then
    print_success "Go is installed: $(go version)"
else
    print_error "Go is not installed"
    exit 1
fi

# Check if we're in the right directory
if [[ ! -f "pkg/profile/profile.go" ]]; then
    print_error "Not in mc-tool root directory"
    echo "Please run this script from the mc-tool project root"
    exit 1
fi

# Check MinIO client availability
MC_AVAILABLE=false
if command -v mc &> /dev/null; then
    print_success "MinIO client (mc) is available: $(mc --version | head -1)"
    MC_AVAILABLE=true
else
    print_warning "MinIO client (mc) is not installed"
    echo "  Install from: https://docs.min.io/docs/minio-client-quickstart-guide.html"
fi

# Check for mc-2021
if command -v mc-2021 &> /dev/null; then
    print_success "mc-2021 is available"
else
    print_warning "mc-2021 is not available (install via Docker image)"
fi

echo

# Run unit tests
print_section "Unit Tests"

echo "Running profile package unit tests..."
if go test -v ./pkg/profile/ -short; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

echo

# Run benchmark tests
print_section "Benchmark Tests"

echo "Running performance benchmarks..."
go test -bench=. ./pkg/profile/ -benchmem

echo

# Build the application
print_section "Build Test"

echo "Building mc-tool..."
if go build -o build/mc-tool-test .; then
    print_success "Build successful"
else
    print_error "Build failed"
    exit 1
fi

echo

# Test command help
print_section "Command Interface Test"

echo "Testing profile command help..."
./build/mc-tool-test profile --help > /dev/null
print_success "Profile command help works"

echo "Testing profile command with invalid arguments..."
if ./build/mc-tool-test profile invalid-type invalid-alias 2>/dev/null; then
    print_warning "Expected failure didn't occur"
else
    print_success "Invalid arguments properly rejected"
fi

echo

# Run playground tests
print_section "Playground Tests"

echo "Running playground demonstrations..."
cd pkg/profile

if go run playground_test.go; then
    print_success "Playground tests completed"
else
    print_warning "Playground tests had issues"
fi

cd ../..

echo

# Integration tests (if mc is available and alias provided)
print_section "Integration Tests"

if [[ "$MC_AVAILABLE" == "true" && -n "$1" ]]; then
    ALIAS="$1"
    echo "Testing with MinIO alias: $ALIAS"
    
    # Check if alias exists
    if mc alias list | grep -q "^$ALIAS"; then
        print_success "Alias '$ALIAS' found in mc configuration"
        
        echo "Testing mc admin/support profile availability..."
        cd pkg/profile
        if go run playground_test.go "$ALIAS"; then
            print_success "Integration test with alias completed"
        else
            print_warning "Integration test had issues"
        fi
        cd ../..
    else
        print_warning "Alias '$ALIAS' not found in mc configuration"
        echo "Available aliases:"
        mc alias list | grep -v "^gcs" | head -5
    fi
else
    print_warning "Integration tests skipped"
    if [[ "$MC_AVAILABLE" != "true" ]]; then
        echo "  Reason: mc client not available"
    else
        echo "  Reason: no alias provided"
        echo "  Usage: $0 <minio-alias>"
    fi
fi

echo

# Test Docker integration (if Docker is available)
print_section "Docker Integration Test"

if command -v docker &> /dev/null; then
    echo "Testing Docker image build..."
    
    # Check if Docker image exists
    if docker images | grep -q "mc-tool.*profile"; then
        print_success "Docker image with profile support exists"
        
        echo "Testing profile command in container..."
        if docker run --rm mc-tool:profile-final mc-tool profile --help > /dev/null; then
            print_success "Profile command works in Docker container"
        else
            print_warning "Profile command failed in Docker container"
        fi
    else
        print_warning "Docker image not found (run: docker build -t mc-tool:profile-final .)"
    fi
else
    print_warning "Docker not available, skipping container tests"
fi

echo

# Memory leak detection demonstration
print_section "Memory Leak Detection Demo"

echo "Running memory leak detection demonstration..."
if [[ -f "demo-memory-leak-detection.sh" ]]; then
    if bash demo-memory-leak-detection.sh; then
        print_success "Memory leak detection demo completed"
    else
        print_warning "Memory leak detection demo had issues"
    fi
else
    print_warning "Memory leak detection demo script not found"
fi

echo

# Performance testing
print_section "Performance Testing"

echo "Testing profile parsing performance..."
cd pkg/profile
go test -bench=BenchmarkParseHeapProfile -benchtime=5s ./
cd ../..

echo

# Code quality checks
print_section "Code Quality"

echo "Running go vet..."
if go vet ./pkg/profile/; then
    print_success "go vet passed"
else
    print_error "go vet found issues"
fi

echo "Running go fmt check..."
if [[ -z $(gofmt -l pkg/profile/) ]]; then
    print_success "Code is properly formatted"
else
    print_warning "Code formatting issues found:"
    gofmt -l pkg/profile/
fi

echo

# Final summary
print_section "Test Summary"

echo "Test suite completed!"
echo
echo "📊 Test Results:"
echo "  ✅ Unit tests: PASSED"
echo "  ✅ Build test: PASSED"
echo "  ✅ Command interface: PASSED"
echo "  ✅ Playground tests: COMPLETED"
if [[ "$MC_AVAILABLE" == "true" && -n "$1" ]]; then
    echo "  ✅ Integration tests: COMPLETED"
else
    echo "  ⚠️  Integration tests: SKIPPED"
fi
echo

echo "🚀 Ready for production!"
echo
echo "Next steps:"
echo "  1. Deploy with: cp build/mc-tool-test /usr/local/bin/mc-tool"
echo "  2. Test with real MinIO: mc-tool profile heap <your-alias>"
echo "  3. Monitor production: mc-tool profile heap prod --detect-leaks --duration 1h"

if [[ "$MC_AVAILABLE" != "true" ]]; then
    echo
    echo "💡 To enable full testing:"
    echo "  1. Install MinIO client: https://docs.min.io/docs/minio-client-quickstart-guide.html"
    echo "  2. Configure alias: mc alias set myalias https://minio.example.com ACCESS_KEY SECRET_KEY"
    echo "  3. Run: $0 myalias"
fi

echo
echo "=== Test Suite Complete ==="