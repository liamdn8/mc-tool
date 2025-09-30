#!/bin/bash

# Build script for mc-tool
echo "Building mc-tool..."

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Build the application
echo "Running go mod tidy..."
go mod tidy

echo "Building application..."
go build -o mc-tool main.go

if [ $? -eq 0 ]; then
    echo "Build successful! The 'mc-tool' binary is ready to use."
    echo ""
    echo "Usage examples:"
    echo "  ./mc-tool compare alias1/bucket1 alias2/bucket2"
    echo "  ./mc-tool compare --versions alias1/bucket1 alias2/bucket2"
    echo "  ./mc-tool compare --verbose alias1/bucket1/path alias2/bucket2/path"
else
    echo "Build failed!"
    exit 1
fi