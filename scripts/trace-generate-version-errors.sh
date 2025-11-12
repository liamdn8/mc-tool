#!/usr/bin/env bash
# Helper script to trigger errors with versioned objects for mc-tool trace --versions testing.

set -uo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: $0 <mc-alias>"
    exit 1
fi

ALIAS=$1

if ! mc alias list | grep -q "^$ALIAS"; then
    echo "Alias '$ALIAS' not found in mc configuration." >&2
    exit 1
fi

echo "Generating versioned object errors against alias '$ALIAS'..."

TEST_BUCKET="trace-versions-test-$(date +%s)"
TMPFILE=$(mktemp)
TMPFILE2=$(mktemp)
printf 'version 1 content\n' > "$TMPFILE"
printf 'version 2 content\n' > "$TMPFILE2"

echo "1) Creating versioned bucket"
mc mb "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
mc version enable "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
sleep 0.2

echo "2) Creating objects with multiple versions"
OBJECTS=(doc1.txt doc2.txt doc3.txt)

for obj in "${OBJECTS[@]}"; do
    # Create initial version
    mc cp "$TMPFILE" "$ALIAS/$TEST_BUCKET/${obj}" >/dev/null 2>&1 || true
    sleep 0.1
    # Create second version
    mc cp "$TMPFILE2" "$ALIAS/$TEST_BUCKET/${obj}" >/dev/null 2>&1 || true
    sleep 0.1
done

echo "3) Getting version IDs"
declare -A VERSION_IDS

for obj in "${OBJECTS[@]}"; do
    # Get the list of versions for each object
    VERSIONS=$(mc ls --versions "$ALIAS/$TEST_BUCKET/${obj}" 2>/dev/null | awk '{print $6}' | grep -v '^$' || true)
    if [ -n "$VERSIONS" ]; then
        # Store the version IDs
        VERSION_IDS[$obj]="$VERSIONS"
        echo "   Object: $obj has versions: $VERSIONS"
    fi
done

echo "4) Attempting to access non-existent version IDs (should generate 404 errors)"
FAKE_VERSIONS=("00000000-0000-0000-0000-000000000001" "00000000-0000-0000-0000-000000000002" "00000000-0000-0000-0000-000000000003")

for obj in "${OBJECTS[@]}"; do
    for fake_ver in "${FAKE_VERSIONS[@]}"; do
        # Try to get objects with fake version IDs
        mc cat "$ALIAS/$TEST_BUCKET/${obj}?versionId=${fake_ver}" >/dev/null 2>&1 || true
        sleep 0.05
    done
done

echo "5) Attempting to delete specific versions (should generate permission/error events)"
for obj in "${OBJECTS[@]}"; do
    if [ -n "${VERSION_IDS[$obj]:-}" ]; then
        for ver in ${VERSION_IDS[$obj]}; do
            # Try to stat the specific version (this will appear in trace)
            mc stat "$ALIAS/$TEST_BUCKET/${obj}?versionId=${ver}" >/dev/null 2>&1 || true
            sleep 0.05
        done
    fi
done

echo "6) Attempting operations on deleted versions"
# Delete objects to create delete markers
for obj in "${OBJECTS[@]}"; do
    mc rm "$ALIAS/$TEST_BUCKET/${obj}" >/dev/null 2>&1 || true
    sleep 0.1
done

echo "7) Getting delete marker version IDs"
declare -A DELETE_MARKER_IDS

for obj in "${OBJECTS[@]}"; do
    # List versions to find delete markers
    DELETE_MARKERS=$(mc ls --versions "$ALIAS/$TEST_BUCKET/${obj}" 2>/dev/null | grep "DEL" | awk '{print $6}' | grep -v '^$' || true)
    if [ -n "$DELETE_MARKERS" ]; then
        DELETE_MARKER_IDS[$obj]="$DELETE_MARKERS"
        echo "   Object: $obj has delete marker(s): $DELETE_MARKERS"
    fi
done

echo "8) Attempting to access delete marker versions (should trigger 405 Method Not Allowed)"
# Try to access delete markers using mc commands - this will properly authenticate and show in trace
for obj in "${OBJECTS[@]}"; do
    if [ -n "${DELETE_MARKER_IDS[$obj]:-}" ]; then
        for ver in ${DELETE_MARKER_IDS[$obj]}; do
            echo "   Attempting to access delete marker: ${obj} version ${ver}"
            
            # Try to stat the delete marker version (HEAD operation)
            mc stat "$ALIAS/$TEST_BUCKET/${obj}?versionId=${ver}" >/dev/null 2>&1 || true
            sleep 0.1
            
            # Try to cat/read the delete marker version (GET operation) - should fail with 405
            mc cat "$ALIAS/$TEST_BUCKET/${obj}?versionId=${ver}" >/dev/null 2>&1 || true
            sleep 0.1
            
            # Try to copy the delete marker version - should also fail
            mc cp "$ALIAS/$TEST_BUCKET/${obj}?versionId=${ver}" /tmp/test-delete-marker 2>&1 || true
            sleep 0.1
        done
    fi
done

# Try to access the deleted versions (without version ID)
echo "9) Attempting to access objects with delete markers"
for obj in "${OBJECTS[@]}"; do
    mc cat "$ALIAS/$TEST_BUCKET/${obj}" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "10) Triggering 405 with unsupported HTTP methods"
# MinIO may return 405 for certain unsupported operations

if command -v curl &> /dev/null; then
    ENDPOINT=$(mc alias list "$ALIAS" 2>/dev/null | grep -A 1 "^$ALIAS" | grep "URL" | awk '{print $3}')
    
    if [ -n "$ENDPOINT" ]; then
        echo "   Using curl to attempt unsupported HTTP methods at: $ENDPOINT"
        
        for obj in "${OBJECTS[@]}"; do
            # Try PATCH method (not supported by S3 API)
            curl -X PATCH "${ENDPOINT}/${TEST_BUCKET}/${obj}" \
                 -s -o /dev/null -w "PATCH response: %{http_code}\n" 2>&1 || true
            sleep 0.05
            
            # Try TRACE method (typically disabled)
            curl -X TRACE "${ENDPOINT}/${TEST_BUCKET}/${obj}" \
                 -s -o /dev/null -w "TRACE response: %{http_code}\n" 2>&1 || true
            sleep 0.05
        done
    fi
fi

echo "   Note: 405 errors depend on server configuration and may not always trigger"

echo "Cleaning up temporary resources"
mc rm -r --force --versions "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
rm -f "$TMPFILE" "$TMPFILE2"

echo ""
echo "Done! Now run:"
echo "  ./mc-tool trace $ALIAS --versions --duration 10s --verbose"
echo ""
echo "This should show errors grouped by object AND version ID."
