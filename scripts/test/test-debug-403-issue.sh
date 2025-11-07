#!/bin/bash

# Demonstration script showing the 403 issue with pprof endpoints

echo "🔍 Testing MinIO Debug Endpoint Access"
echo "======================================"
echo

echo "1. ✅ Testing Regular S3 API (should work):"
echo "-------------------------------------------"
echo "Using awscurl to list buckets:"
/home/liamdn/.local/bin/awscurl --access_key Q3AM3UQ867SPQQA43P2F --secret_key zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG --service s3 --region us-east-1 'https://play.min.io/' | head -3
echo

echo "2. ❌ Testing pprof Endpoint (expected 403):"
echo "--------------------------------------------"
echo "Using curl to access pprof goroutine endpoint:"
curl -s 'https://play.min.io/debug/pprof/goroutine?debug=1' | head -3
echo

echo "3. ❌ Testing with awscurl pprof (also 403):"
echo "-------------------------------------------"
echo "Using awscurl with pprof endpoint:"
/home/liamdn/.local/bin/awscurl --access_key Q3AM3UQ867SPQQA43P2F --secret_key zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG --service s3 --region us-east-1 'https://play.min.io/debug/pprof/goroutine?debug=1' | head -3
echo

echo "4. 🧪 Testing mc-tool debug (improved error handling):"
echo "------------------------------------------------------"
echo "Using mc-tool debug with verbose output:"
./build/mc-tool-portable debug play --verbose 2>&1 | head -15
echo

echo "📝 Analysis:"
echo "============"
echo "• Regular S3 API works fine with proper AWS v4 signing"
echo "• pprof endpoints return 403 on play.min.io (expected security measure)"
echo "• awscurl treats pprof paths as S3 object requests (wrong approach)"
echo "• mc-tool debug now provides better error messages and troubleshooting tips"
echo
echo "🛠️ Solution for Production Use:"
echo "• Use mc-tool debug with local MinIO instances"
echo "• Ensure MINIO_PROFILE=enable environment variable is set"
echo "• Check that pprof endpoints are enabled in MinIO configuration"
echo "• Use insecure flag for self-signed certificates: --insecure"