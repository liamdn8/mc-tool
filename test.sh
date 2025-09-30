#!/bin/bash

# Test script for mc-tool
# This script demonstrates basic functionality and validates the tool

echo "=== mc-tool Test Script ==="
echo ""

# Check if the binary exists
if [ ! -f "./mc-tool" ] && [ ! -f "./build/mc-tool" ]; then
    echo "Error: mc-tool binary not found. Please run 'make build' first."
    exit 1
fi

# Use the binary from build directory if it exists, otherwise current directory
BINARY="./mc-tool"
if [ -f "./build/mc-tool" ]; then
    BINARY="./build/mc-tool"
fi

echo "Using binary: $BINARY"
echo ""

# Test 1: Show help
echo "Test 1: Showing help message"
echo "=============================="
$BINARY --help
echo ""

# Test 2: Show compare command help
echo "Test 2: Showing compare command help"
echo "====================================="
$BINARY compare --help
echo ""

# Test 3: Test with invalid arguments (should show error)
echo "Test 3: Testing with invalid arguments"
echo "======================================"
echo "Running: $BINARY compare (should show error about missing arguments)"
$BINARY compare 2>/dev/null || echo "✓ Correctly shows error for missing arguments"
echo ""

# Test 4: Test with invalid URL format
echo "Test 4: Testing with invalid URL format"
echo "======================================="
echo "Running: $BINARY compare invalid-url another-invalid-url"
$BINARY compare invalid-url another-invalid-url 2>/dev/null || echo "✓ Correctly shows error for invalid URL format"
echo ""

# Test 5: Test with non-existent alias (will fail but shows proper error handling)
echo "Test 5: Testing with non-existent alias"
echo "======================================="
echo "Running: $BINARY compare nonexistent/bucket1 nonexistent/bucket2"
$BINARY compare nonexistent/bucket1 nonexistent/bucket2 2>/dev/null || echo "✓ Correctly shows error for non-existent alias"
echo ""

echo "=== Test Results ==="
echo "Basic functionality tests completed."
echo ""
echo "To test with real MinIO instances:"
echo "1. Configure your aliases using 'mc alias set'"
echo "2. Run: $BINARY compare alias1/bucket alias2/bucket"
echo "3. Add --versions flag to compare all object versions"
echo "4. Add --verbose flag for detailed output"
echo ""
echo "Example with real aliases:"
echo "  $BINARY compare local/test-bucket prod/test-bucket"
echo "  $BINARY compare --versions --verbose staging/backup prod/backup"