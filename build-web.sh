#!/bin/bash

# Build script for MC-Tool Web UI
# This script builds the React app and copies it to the correct location for Go embedding

set -e

echo "🔨 Building React app..."

# Change to web directory
cd web

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
fi

# Build the React app
echo "Building React app..."
npm run build

echo "📁 Copying React build to Go embed location..."
cd ..

# Copy built files to pkg/web/static for embedding
rm -rf pkg/web/static/build
mkdir -p pkg/web/static/build
cp -r web/build/* pkg/web/static/build/

echo "🔧 Building Go application..."
go build -o mc-tool

echo "✅ Build completed successfully!"
echo "🚀 You can now run: ./mc-tool web --port 8080"
echo "React web UI build completed!"
echo "Built files are in pkg/web/static/build/"