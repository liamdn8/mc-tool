#!/bin/bash

# MinIO Site Replication Test Setup
# This script creates 2 MinIO servers for testing site replication

set -e

echo "🚀 Setting up MinIO Site Replication Test Environment"
echo "=================================================="

# Configuration
NETWORK_NAME="minio-replication-network"
SITE1_NAME="minio-site1"
SITE2_NAME="minio-site2"
SITE1_PORT="9001"
SITE2_PORT="9002"
SITE1_CONSOLE_PORT="9011"
SITE2_CONSOLE_PORT="9012"
ROOT_USER="minioadmin"
ROOT_PASSWORD="minioadmin123"

# Get host IP (not localhost)
HOST_IP=$(hostname -I | awk '{print $1}')
echo "📍 Detected Host IP: $HOST_IP"

# Clean up existing containers and network
echo ""
echo "🧹 Cleaning up existing containers..."
docker rm -f $SITE1_NAME $SITE2_NAME 2>/dev/null || true
docker network rm $NETWORK_NAME 2>/dev/null || true

# Create Docker network
echo ""
echo "🌐 Creating Docker network: $NETWORK_NAME"
docker network create $NETWORK_NAME

# Create data directories
echo ""
echo "📁 Creating data directories..."
mkdir -p ./test-data/site1
mkdir -p ./test-data/site2

# Start MinIO Site 1
echo ""
echo "🔧 Starting MinIO Site 1..."
docker run -d \
  --name $SITE1_NAME \
  --network $NETWORK_NAME \
  -p $SITE1_PORT:9000 \
  -p $SITE1_CONSOLE_PORT:9001 \
  -e "MINIO_ROOT_USER=$ROOT_USER" \
  -e "MINIO_ROOT_PASSWORD=$ROOT_PASSWORD" \
  -e "MINIO_SERVER_URL=http://${HOST_IP}:${SITE1_PORT}" \
  -e "MINIO_BROWSER_REDIRECT_URL=http://${HOST_IP}:${SITE1_CONSOLE_PORT}" \
  -v $(pwd)/test-data/site1:/data \
  minio/minio server /data --console-address ":9001"

echo "   ✓ Site 1: http://${HOST_IP}:${SITE1_PORT}"
echo "   ✓ Console 1: http://${HOST_IP}:${SITE1_CONSOLE_PORT}"

# Start MinIO Site 2
echo ""
echo "🔧 Starting MinIO Site 2..."
docker run -d \
  --name $SITE2_NAME \
  --network $NETWORK_NAME \
  -p $SITE2_PORT:9000 \
  -p $SITE2_CONSOLE_PORT:9001 \
  -e "MINIO_ROOT_USER=$ROOT_USER" \
  -e "MINIO_ROOT_PASSWORD=$ROOT_PASSWORD" \
  -e "MINIO_SERVER_URL=http://${HOST_IP}:${SITE2_PORT}" \
  -e "MINIO_BROWSER_REDIRECT_URL=http://${HOST_IP}:${SITE2_CONSOLE_PORT}" \
  -v $(pwd)/test-data/site2:/data \
  minio/minio server /data --console-address ":9001"

echo "   ✓ Site 2: http://${HOST_IP}:${SITE2_PORT}"
echo "   ✓ Console 2: http://${HOST_IP}:${SITE2_CONSOLE_PORT}"

# Wait for MinIO to be ready
echo ""
echo "⏳ Waiting for MinIO servers to be ready..."
sleep 5

# Configure mc aliases
echo ""
echo "🔑 Configuring mc aliases..."

# Remove old aliases if they exist
mc alias remove site1 2>/dev/null || true
mc alias remove site2 2>/dev/null || true

# Add new aliases with accessible IPs
mc alias set site1 http://${HOST_IP}:${SITE1_PORT} $ROOT_USER $ROOT_PASSWORD
mc alias set site2 http://${HOST_IP}:${SITE2_PORT} $ROOT_USER $ROOT_PASSWORD

# Verify connectivity
echo ""
echo "✅ Verifying connectivity..."
mc admin info site1 --json > /dev/null && echo "   ✓ Site 1: Connected"
mc admin info site2 --json > /dev/null && echo "   ✓ Site 2: Connected"

# Display summary
echo ""
echo "=================================================="
echo "✨ Setup Complete!"
echo "=================================================="
echo ""
echo "📋 Connection Details:"
echo "   Host IP: $HOST_IP"
echo ""
echo "   Site 1 API:     http://${HOST_IP}:${SITE1_PORT}"
echo "   Site 1 Console: http://${HOST_IP}:${SITE1_CONSOLE_PORT}"
echo "   Alias: site1"
echo ""
echo "   Site 2 API:     http://${HOST_IP}:${SITE2_PORT}"
echo "   Site 2 Console: http://${HOST_IP}:${SITE2_CONSOLE_PORT}"
echo "   Alias: site2"
echo ""
echo "   Credentials: $ROOT_USER / $ROOT_PASSWORD"
echo ""
echo "🎯 Next Steps:"
echo "   1. Start mc-tool web interface:"
echo "      ./mc-tool web --port 8080"
echo ""
echo "   2. Open browser: http://localhost:8080"
echo ""
echo "   3. Go to 'Sites' page and add sites to replication"
echo ""
echo "   Or use command line:"
echo "   mc admin replicate add site1 site2"
echo ""
echo "📝 Quick Commands:"
echo "   - List aliases:          mc alias list"
echo "   - Check site1:           mc admin info site1"
echo "   - Check site2:           mc admin info site2"
echo "   - View containers:       docker ps"
echo "   - View logs (site1):     docker logs -f $SITE1_NAME"
echo "   - View logs (site2):     docker logs -f $SITE2_NAME"
echo ""
echo "🛑 To stop and cleanup:"
echo "   ./cleanup-minio-test.sh"
echo ""
