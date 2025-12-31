#!/bin/bash

# Test Environment Variables Validator
# This script tests the new info validator feature

echo "🧪 Testing Environment Variables Validation"
echo "============================================"

# Check if mc-tool web is running
if ! pgrep -f "mc-tool.*web" > /dev/null; then
    echo "❌ Error: mc-tool web server is not running"
    echo "   Please start it with: ./mc-tool-web web --port 8080"
    exit 1
fi

echo ""
echo "📊 Testing /api/operations/validate-bucket-config with empty buckets..."
echo ""

# Test with empty buckets to trigger env vars validation
curl -s -X POST http://localhost:8080/api/operations/validate-bucket-config \
  -H "Content-Type: application/json" \
  -d '{
    "aliases": ["site1"],
    "buckets": [],
    "check_lifecycle": false,
    "check_events": false
  }' | jq '.'

echo ""
echo "✅ Test completed!"
echo ""
echo "Expected output:"
echo "  - env_vars array with MinIO version, commitID, and environment variables"
echo "  - Filtered variables (excluding *_PORT*, *_SERVICE*)"
