#!/bin/bash

echo "🚀 Setting up MinIO Site Replication Test Environment"
echo ""

# Start MinIO servers
echo "📦 Starting 2 MinIO servers..."
cd /home/liamdn/mc-tool
docker-compose up -d

echo ""
echo "⏳ Waiting for MinIO servers to be ready..."
sleep 10

# Configure mc aliases
echo ""
echo "🔧 Configuring mc aliases..."

# Add site1
mc alias set site1 http://localhost:9001 minioadmin minioadmin
echo "✅ Added site1 (localhost:9001)"

# Add site2
mc alias set site2 http://localhost:9002 minioadmin minioadmin
echo "✅ Added site2 (localhost:9002)"

echo ""
echo "🪣 Creating test buckets..."

# Create buckets on site1
mc mb site1/test-bucket-1
mc mb site1/test-bucket-2
mc mb site1/shared-bucket

# Create buckets on site2
mc mb site2/test-bucket-3
mc mb site2/shared-bucket

echo ""
echo "📝 Adding some test data..."

# Add test objects
echo "Test file 1" | mc pipe site1/test-bucket-1/file1.txt
echo "Test file 2" | mc pipe site1/test-bucket-2/file2.txt
echo "Shared file" | mc pipe site1/shared-bucket/shared.txt

echo "Test file 3" | mc pipe site2/test-bucket-3/file3.txt
echo "Shared file 2" | mc pipe site2/shared-bucket/shared2.txt

echo ""
echo "✅ Setup complete!"
echo ""
echo "📊 MinIO Servers:"
echo "  - Site 1: http://localhost:9001 (Console: http://localhost:9091)"
echo "  - Site 2: http://localhost:9002 (Console: http://localhost:9092)"
echo ""
echo "🔑 Credentials:"
echo "  - Username: minioadmin"
echo "  - Password: minioadmin"
echo ""
echo "🎯 MC Aliases configured:"
echo "  - site1"
echo "  - site2"
echo ""
echo "Now you can test the Site Replication UI at http://localhost:8080"
