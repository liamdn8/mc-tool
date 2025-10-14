#!/bin/bash

# Build script for React web UI
set -e

echo "Building React web UI..."

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

# Copy built files to pkg/web/static for embedding
echo "Copying built files..."
rm -rf ../pkg/web/static/build
mkdir -p ../pkg/web/static/build
cp -r build/* ../pkg/web/static/build/

echo "React web UI build completed!"
echo "Built files are in pkg/web/static/build/"