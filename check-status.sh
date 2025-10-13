#!/bin/bash

echo "🔍 MinIO Site Replication Status Check"
echo ""

# Check if containers are running
echo "📦 Docker Containers:"
docker ps --filter "name=minio-site" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""

# Check mc aliases
echo "🔧 MC Aliases:"
echo "site1: $(mc alias list | grep -A1 'site1' | grep URL | awk '{print $3}')"
echo "site2: $(mc alias list | grep -A1 'site2' | grep URL | awk '{print $3}')"
echo ""

# Check buckets on site1
echo "🪣 Site 1 Buckets:"
mc ls site1 2>/dev/null || echo "  ❌ Cannot connect to site1"
echo ""

# Check buckets on site2
echo "🪣 Site 2 Buckets:"
mc ls site2 2>/dev/null || echo "  ❌ Cannot connect to site2"
echo ""

# Check objects
echo "📄 Objects on Site 1:"
echo "  test-bucket-1:"
mc ls site1/test-bucket-1 2>/dev/null | sed 's/^/    /'
echo "  test-bucket-2:"
mc ls site1/test-bucket-2 2>/dev/null | sed 's/^/    /'
echo "  shared-bucket:"
mc ls site1/shared-bucket 2>/dev/null | sed 's/^/    /'
echo ""

echo "📄 Objects on Site 2:"
echo "  test-bucket-3:"
mc ls site2/test-bucket-3 2>/dev/null | sed 's/^/    /'
echo "  shared-bucket:"
mc ls site2/shared-bucket 2>/dev/null | sed 's/^/    /'
echo ""

# Summary
echo "📊 Summary:"
SITE1_BUCKETS=$(mc ls site1 2>/dev/null | wc -l)
SITE2_BUCKETS=$(mc ls site2 2>/dev/null | wc -l)
SITE1_OBJECTS=$(mc ls --recursive site1 2>/dev/null | wc -l)
SITE2_OBJECTS=$(mc ls --recursive site2 2>/dev/null | wc -l)

echo "  Site 1: $SITE1_BUCKETS buckets, $SITE1_OBJECTS objects"
echo "  Site 2: $SITE2_BUCKETS buckets, $SITE2_OBJECTS objects"
echo ""

# Check shared bucket
echo "🔄 Shared Bucket Analysis:"
SITE1_SHARED=$(mc ls site1/shared-bucket 2>/dev/null | wc -l)
SITE2_SHARED=$(mc ls site2/shared-bucket 2>/dev/null | wc -l)
echo "  Site 1 shared-bucket: $SITE1_SHARED objects"
echo "  Site 2 shared-bucket: $SITE2_SHARED objects"
if [ "$SITE1_SHARED" -eq "$SITE2_SHARED" ]; then
    echo "  ✅ Object count matches"
else
    echo "  ⚠️  Object count differs - not synced"
fi
echo ""

echo "🌐 Access URLs:"
echo "  Site 1 Console: http://localhost:9091"
echo "  Site 2 Console: http://localhost:9092"
echo "  MC-Tool Web UI: http://localhost:8080"
