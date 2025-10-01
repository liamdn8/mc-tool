#!/bin/bash

# Quick test to check Docker image size optimization

echo "🐳 Testing Docker Image Size Optimization"
echo "========================================="
echo

# Build the optimized image
echo "📦 Building optimized image..."
if docker build -t mc-tool:optimized . >/dev/null 2>&1; then
    SIZE=$(docker images mc-tool:optimized --format "{{.Size}}")
    echo "✅ Build successful!"
    echo "📊 Image size: $SIZE"
else
    echo "❌ Build failed"
    exit 1
fi

echo
echo "🧪 Testing functionality..."

# Test awscurl availability
echo -n "   awscurl import: "
if docker run --rm mc-tool:optimized -c "python3 -c 'import awscurl; print(\"OK\")'" 2>/dev/null | grep -q "OK"; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
fi

# Test mc-tool
echo -n "   mc-tool version: "
if docker run --rm mc-tool:optimized -c "mc-tool version" >/dev/null 2>&1; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
fi

# Test mc client
echo -n "   mc client: "
if docker run --rm mc-tool:optimized -c "mc --version" >/dev/null 2>&1; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
fi

# Test jq
echo -n "   jq tool: "
if docker run --rm mc-tool:optimized -c "echo '{}' | jq ." >/dev/null 2>&1; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
fi

echo
echo "🎯 Optimization complete!"
echo "   Image: mc-tool:optimized"
echo "   Size: $SIZE"
echo
echo "💡 To use:"
echo "   docker run -it mc-tool:optimized"