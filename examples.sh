#!/bin/bash

# Example usage script for mc-tool
# This script shows various ways to use the mc-tool for comparing MinIO buckets

echo "=== mc-tool Usage Examples ==="
echo ""

# Check if binary exists
if [ ! -f "./build/mc-tool" ]; then
    echo "Error: Please build the tool first with 'make build'"
    exit 1
fi

TOOL="./build/mc-tool"

echo "Prerequisites:"
echo "1. Install and configure mc client: https://min.io/docs/minio/linux/reference/minio-mc.html"
echo "2. Set up your MinIO aliases using 'mc alias set'"
echo ""

echo "Setting up example aliases (replace with your actual MinIO instances):"
echo "  mc alias set local http://localhost:9000 minioadmin minioadmin"
echo "  mc alias set prod https://minio-prod.example.com ACCESS_KEY SECRET_KEY"
echo "  mc alias set staging https://minio-staging.example.com ACCESS_KEY SECRET_KEY"
echo ""

echo "=== Basic Usage Examples ==="
echo ""

echo "1. Compare two buckets (current versions only):"
echo "   $TOOL compare local/test-bucket prod/test-bucket"
echo ""

echo "2. Compare specific paths within buckets:"
echo "   $TOOL compare local/backup/2024 prod/backup/2024"
echo ""

echo "3. Compare with verbose output:"
echo "   $TOOL compare --verbose local/data prod/data"
echo ""

echo "4. Skip TLS certificate verification:"
echo "   $TOOL compare --insecure local/test-bucket prod/test-bucket"
echo ""

echo "5. Compare all object versions (for versioned buckets):"
echo "   $TOOL compare --versions local/versioned-bucket prod/versioned-bucket"
echo ""

echo "6. Full comparison with versions, verbose output, and insecure connection:"
echo "   $TOOL compare --versions --verbose --insecure staging/important-data prod/important-data"
echo ""

echo "=== What the tool compares ==="
echo ""
echo "Default mode (current versions):"
echo "  ✓ Compares latest version of each object"
echo "  ✓ Uses ETag and file size for comparison"
echo "  ✓ Shows objects that are identical, different, or missing"
echo ""

echo "Versions mode (--versions flag):"
echo "  ✓ Compares ALL versions of each object"
echo "  ✓ Matches objects by version ID"
echo "  ✓ Ensures complete replication including historical versions"
echo ""

echo "=== Output interpretation ==="
echo ""
echo "✓ Identical objects (shown only in verbose mode)"
echo "⚠ Different objects (ETag or size differs)"
echo "- Missing in source"
echo "+ Missing in target"
echo ""

echo "Exit codes:"
echo "  0 - All objects are identical"
echo "  1 - Differences found"
echo ""

echo "=== Common use cases ==="
echo ""
echo "1. Verify replication between MinIO instances:"
echo "   $TOOL compare prod/critical-data backup/critical-data"
echo ""

echo "2. Check if backup is complete:"
echo "   $TOOL compare --versions prod/backup offsite/backup"
echo ""

echo "3. Validate data migration:"
echo "   $TOOL compare old-cluster/data new-cluster/data"
echo ""

echo "4. Monitor sync between regions:"
echo "   $TOOL compare us-east/shared-data eu-west/shared-data"
echo ""

echo "To run any of these examples, replace the alias names with your configured aliases."