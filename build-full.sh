#!/bin/bash
# Full build script: Build React app then Go static binary

set -euo pipefail

echo "=============================="
echo " Building MC-Tool with React UI"
echo "=============================="

# Step 1: Build React app
echo "Step 1: Building React web UI..."
./build-web.sh

# Step 2: Build Go static binary with embedded React files
echo "Step 2: Building Go static binary..."

# Optional: clean old build
rm -f mc-tool

# Choose output name and version info
APP_NAME="mc-tool"
BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Static build (no libc dependency)
CGO_ENABLED=0 go build -a -installsuffix cgo \
  -ldflags "-s -w \
    -X 'main.BuildTime=${BUILD_TIME}' \
    -X 'main.GitCommit=${GIT_COMMIT}'" \
  -o "${APP_NAME}" main.go

echo
echo "✅ Build completed successfully!"
echo "📦 Output binary: ${APP_NAME}"
echo "🚀 Run with: ./${APP_NAME} web --port 8080"
