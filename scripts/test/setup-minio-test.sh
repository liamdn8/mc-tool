#!/bin/bash

# MinIO Site Replication Test Setup
# This script creates 6 MinIO servers for testing site replication

set -e

echo "🚀 Setting up MinIO Site Replication Test Environment (6 Sites)"
echo "=============================================================="

# Configuration
NETWORK_NAME="minio-replication-network"
ROOT_USER="minioadmin"
ROOT_PASSWORD="minioadmin123"
NUM_SITES=6

# Get host IP (not localhost)
HOST_IP=$(hostname -I | awk '{print $1}')
echo "📍 Detected Host IP: $HOST_IP"

# Clean up existing containers and network
echo ""
echo "🧹 Cleaning up existing containers..."
for i in $(seq 1 $NUM_SITES); do
    docker rm -f "minio-site$i" 2>/dev/null || true
done
docker network rm $NETWORK_NAME 2>/dev/null || true

# Create Docker network
echo ""
echo "🌐 Creating Docker network: $NETWORK_NAME"
docker network create $NETWORK_NAME

# Create data directories
echo ""
echo "📁 Creating data directories..."
for i in $(seq 1 $NUM_SITES); do
    mkdir -p "./test-data/site$i"
done

# Start MinIO Sites
echo ""
echo "🔧 Starting MinIO Sites..."
for i in $(seq 1 $NUM_SITES); do
    SITE_NAME="minio-site$i"
    SITE_PORT=$((9000 + i))
    CONSOLE_PORT=$((9010 + i))
    
    echo "Starting Site $i on port $SITE_PORT (console: $CONSOLE_PORT)..."
    
    docker run -d \
      --name $SITE_NAME \
      --network $NETWORK_NAME \
      -p $SITE_PORT:9000 \
      -p $CONSOLE_PORT:9001 \
      -e "MINIO_ROOT_USER=$ROOT_USER" \
      -e "MINIO_ROOT_PASSWORD=$ROOT_PASSWORD" \
      -e "MINIO_SERVER_URL=http://${HOST_IP}:${SITE_PORT}" \
      -e "MINIO_BROWSER_REDIRECT_URL=http://${HOST_IP}:${CONSOLE_PORT}" \
      -v $(pwd)/test-data/site$i:/data \
      minio/minio server /data --console-address ":9001"

    echo "   ✓ Site $i: http://${HOST_IP}:${SITE_PORT}"
    echo "   ✓ Console $i: http://${HOST_IP}:${CONSOLE_PORT}"
done

# Wait for MinIO to be ready
echo ""
echo "⏳ Waiting for MinIO servers to be ready..."
sleep 10

# Configure mc aliases
echo ""
echo "🔑 Configuring mc aliases..."

# Remove old aliases if they exist
for i in $(seq 1 $NUM_SITES); do
    mc alias remove "site$i" 2>/dev/null || true
done

# Add new aliases with accessible IPs
for i in $(seq 1 $NUM_SITES); do
    SITE_PORT=$((9000 + i))
    mc alias set "site$i" "http://${HOST_IP}:${SITE_PORT}" $ROOT_USER $ROOT_PASSWORD
done

# Verify connectivity
echo ""
echo "✅ Verifying connectivity..."
for i in $(seq 1 $NUM_SITES); do
    if mc admin info "site$i" --json > /dev/null 2>&1; then
        echo "   ✓ Site $i: Connected"
    else
        echo "   ✗ Site $i: Failed to connect"
    fi
done

# Display summary
echo ""
echo "=============================================================="
echo "✨ Setup Complete!"
echo "=============================================================="
echo ""
echo "📋 Connection Details:"
echo "   Host IP: $HOST_IP"
echo ""

for i in $(seq 1 $NUM_SITES); do
    SITE_PORT=$((9000 + i))
    CONSOLE_PORT=$((9010 + i))
    echo "   Site $i API:     http://${HOST_IP}:${SITE_PORT}"
    echo "   Site $i Console: http://${HOST_IP}:${CONSOLE_PORT}"
    echo "   Alias: site$i"
    echo ""
done

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
echo "   mc admin replicate add site1 site2 site3 site4 site5 site6"
echo ""
echo "📝 Quick Commands:"
echo "   - List aliases:          mc alias list"
for i in $(seq 1 $NUM_SITES); do
    echo "   - Check site$i:           mc admin info site$i"
done
echo "   - View containers:       docker ps"
for i in $(seq 1 $NUM_SITES); do
    echo "   - View logs (site$i):     docker logs -f minio-site$i"
done
echo ""
echo "🛑 To stop and cleanup:"
echo "   ./cleanup-minio-test.sh"
echo ""
