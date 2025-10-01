#!/bin/bash

# MC-Tool Binary Verification Script
echo "=== MC-Tool Binary Verification ==="
echo

BUILD_DIR="build"
FAILED=0

if [ ! -d "$BUILD_DIR" ]; then
    echo "❌ Build directory not found. Run 'make build-all' first."
    exit 1
fi

echo "📁 Build directory contents:"
ls -lh "$BUILD_DIR/" | grep -v "^total"
echo

# Test each executable binary
for binary in "$BUILD_DIR"/mc-tool-*; do
    if [ -f "$binary" ] && [ -x "$binary" ]; then
        filename=$(basename "$binary")
        echo -n "🧪 Testing $filename... "
        
        # Skip non-Linux binaries on Linux
        if [[ "$filename" == *"windows"* ]] || [[ "$filename" == *"darwin"* ]]; then
            echo "⚠️  SKIPPED (Cross-platform binary)"
            continue
        fi
        
        # Test version command
        if timeout 5s "$binary" version >/dev/null 2>&1; then
            echo "✅ PASSED"
        else
            echo "❌ FAILED"
            ((FAILED++))
        fi
    fi
done

echo

# Test static linking for Linux binaries
echo "🔗 Checking static linking:"
for binary in "$BUILD_DIR"/mc-tool-*static* "$BUILD_DIR"/mc-tool-portable; do
    if [ -f "$binary" ]; then
        filename=$(basename "$binary")
        echo -n "   $filename: "
        
        if ldd "$binary" 2>&1 | grep -q "not a dynamic executable"; then
            echo "✅ STATIC"
        else
            echo "❌ DYNAMIC"
            ((FAILED++))
        fi
    fi
done

echo

# Show version information from recommended binary
if [ -f "$BUILD_DIR/mc-tool-portable" ]; then
    echo "📋 Version Information (from mc-tool-portable):"
    "$BUILD_DIR/mc-tool-portable" version | sed 's/^/   /'
    echo
fi

# Summary
if [ $FAILED -eq 0 ]; then
    echo "🎉 All binaries verified successfully!"
    echo "✅ Ready for deployment"
else
    echo "❌ $FAILED verification(s) failed"
    exit 1
fi

echo
echo "💡 Deployment recommendations:"
echo "   • Production servers: mc-tool-portable (smallest, no debug info)"
echo "   • Development: mc-tool-static (includes debug symbols)"
echo "   • Containers: mc-tool-portable"
echo "   • Cross-platform: Choose appropriate binary for target OS"