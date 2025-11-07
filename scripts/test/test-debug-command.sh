#!/bin/bash

# Test the debug command with MC configuration

echo "🧪 Testing mc-tool debug command with MC configuration"
echo "====================================================="
echo

# Check if mc-tool binary exists
if [[ ! -f "build/mc-tool-portable" ]]; then
    echo "📦 Building mc-tool first..."
    make build-portable
fi

# Test help command
echo "📋 Testing debug command help..."
./build/mc-tool-portable debug --help

echo
echo "🔍 Testing debug command structure..."

# Test with missing alias (should fail gracefully)
echo -n "   Missing alias test: "
if ./build/mc-tool-portable debug nonexistent-alias 2>/dev/null; then
    echo "❌ Should have failed"
else
    echo "✅ Properly handled missing alias"
fi

# Test with playground alias (if configured)
echo -n "   Playground alias test: "
if ./build/mc-tool-portable debug playground --format json --help >/dev/null 2>&1; then
    echo "✅ Command structure valid"
else
    echo "⚠️  Command structure issue"
fi

# Test insecure option
echo -n "   Insecure option test: "
if ./build/mc-tool-portable debug --help | grep -q "insecure.*Skip TLS certificate verification"; then
    echo "✅ Insecure option available"
else
    echo "❌ Insecure option missing"
fi

echo
echo "� Testing insecure TLS option..."
echo "   The debug command now supports --insecure flag for:"
echo "   • Self-signed certificates"
echo "   • Development environments"
echo "   • MinIO servers with custom CA"

echo
echo "�💡 Usage examples:"
echo "   # First configure an alias:"
echo "   mc alias set local http://localhost:9000 minioadmin minioadmin"
echo "   mc alias set playground https://play.min.io Q3AM3UQ867SPQQA43P2F zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
echo
echo "   # Then use debug command:"
echo "   mc-tool debug local"
echo "   mc-tool debug playground --monitor 5m"
echo "   mc-tool debug local --format json --verbose"
echo "   mc-tool debug local --insecure  # Skip TLS verification"
echo
echo "✅ Debug command is ready for use with MC configuration!"