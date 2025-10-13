#!/bin/bash

# MinIO Site Replication Test Cleanup
# This script removes the test MinIO servers and cleans up data

set -e

echo "🧹 Cleaning up MinIO Site Replication Test Environment"
echo "=================================================="

# Configuration
NETWORK_NAME="minio-replication-network"
SITE1_NAME="minio-site1"
SITE2_NAME="minio-site2"

# Stop and remove containers
echo ""
echo "🛑 Stopping containers..."
docker stop $SITE1_NAME $SITE2_NAME 2>/dev/null || true

echo ""
echo "🗑️  Removing containers..."
docker rm $SITE1_NAME $SITE2_NAME 2>/dev/null || true

# Remove network
echo ""
echo "🌐 Removing network..."
docker network rm $NETWORK_NAME 2>/dev/null || true

# Ask about data cleanup
echo ""
read -p "❓ Do you want to delete test data directories? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️  Removing test data..."
    rm -rf ./test-data
    echo "   ✓ Data removed"
else
    echo "   ℹ️  Data preserved in ./test-data"
fi

# Remove aliases
echo ""
echo "🔑 Removing mc aliases..."
mc alias remove site1 2>/dev/null || true
mc alias remove site2 2>/dev/null || true

echo ""
echo "=================================================="
echo "✅ Cleanup Complete!"
echo "=================================================="
echo ""
