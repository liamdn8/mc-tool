#!/bin/bash

# Full build script: Build React app then Go binary
set -e

echo "Building MC-Tool with React Web UI..."

# Step 1: Build React app
echo "Step 1: Building React web UI..."
./build-web.sh

# Step 2: Build Go binary with embedded React files
echo "Step 2: Building Go binary..."
go build -o mc-tool main.go

echo "✅ Build completed successfully!"
echo "🚀 Run with: ./mc-tool web --port 8080"