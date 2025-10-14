#!/bin/bash

# MinIO Site Replication Test Cleanup
# This script removes all 6 test MinIO servers and cleans up data

set -e

echo "🧹 Cleaning up MinIO Site Replication Test Environment (6 Sites)"
echo "================================================================"

# Configuration
NETWORK_NAME="minio-replication-network"
NUM_SITES=6

# Stop and remove containers
echo ""
echo "🛑 Stopping containers..."
for i in $(seq 1 $NUM_SITES); do
    docker stop "minio-site$i" 2>/dev/null || true
done

echo ""
echo "🗑️  Removing containers..."
for i in $(seq 1 $NUM_SITES); do
    docker rm "minio-site$i" 2>/dev/null || true
done

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
    # Use sudo to remove files with root permissions
    sudo rm -rf ./test-data 2>/dev/null || rm -rf ./test-data 2>/dev/null || true
    echo "   ✓ Data removed"
else
    echo "   ℹ️  Data preserved in ./test-data"
fi

# Remove aliases
echo ""
echo "🔑 Removing mc aliases..."
for i in $(seq 1 $NUM_SITES); do
    mc alias remove "site$i" 2>/dev/null || true
done

echo ""
echo "================================================================"
echo "✅ Cleanup Complete!"
echo "================================================================"
echo ""
