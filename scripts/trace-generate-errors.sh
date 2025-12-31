#!/usr/bin/env bash
# Helper script to generate various MinIO API calls for mc-tool trace testing.
# Creates both successful operations and errors to test trace capture functionality.

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

echo "Generating diverse API calls against alias '$ALIAS' for trace testing..."
echo "This includes both successful operations and errors to test trace capture."
echo ""

TEST_BUCKET="trace-test-$(date +%s)"
TMPFILE=$(mktemp)
TMPFILE2=$(mktemp)
printf 'trace test data %s\n' "$(date --iso-8601=seconds)" > "$TMPFILE"
printf 'second test file %s\n' "$(date --iso-8601=seconds)" > "$TMPFILE2"

echo "═══════════════════════════════════════════════════════════════"
echo "SUCCESSFUL OPERATIONS (for trace all requests testing)"
echo "═══════════════════════════════════════════════════════════════"

echo "[1/10] Creating test bucket - MakeBucket API"
mc mb "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
sleep 0.1

echo "[2/10] Uploading objects - PutObject API"
SUCCESS_KEYS=(data-alpha data-beta data-gamma data-delta)
for key in "${SUCCESS_KEYS[@]}"; do
    for i in $(seq 1 3); do
        mc cp "$TMPFILE" "$ALIAS/$TEST_BUCKET/${key}-${i}.txt" >/dev/null 2>&1 || true
        sleep 0.05
    done
done

echo "[3/10] Listing buckets - ListBuckets API"
for i in $(seq 1 5); do
    mc ls "$ALIAS" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "[4/10] Listing objects - ListObjects API"
for i in $(seq 1 5); do
    mc ls "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "[5/10] Reading objects - GetObject API"
for key in "${SUCCESS_KEYS[@]}"; do
    mc cat "$ALIAS/$TEST_BUCKET/${key}-1.txt" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "[6/10] Getting object metadata - HeadObject API"
for key in "${SUCCESS_KEYS[@]}"; do
    mc stat "$ALIAS/$TEST_BUCKET/${key}-1.txt" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "[7/10] Copying objects - CopyObject API"
mc cp "$ALIAS/$TEST_BUCKET/data-alpha-1.txt" "$ALIAS/$TEST_BUCKET/copy-alpha.txt" >/dev/null 2>&1 || true
mc cp "$ALIAS/$TEST_BUCKET/data-beta-1.txt" "$ALIAS/$TEST_BUCKET/copy-beta.txt" >/dev/null 2>&1 || true
sleep 0.1

echo "[8/10] Setting bucket policies - PutBucketPolicy API"
mc policy set download "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
sleep 0.1

echo "[9/10] Getting bucket policies - GetBucketPolicy API"
mc policy get "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
sleep 0.1

echo "[10/10] Uploading with metadata - PutObject with metadata"
mc cp --attr "X-Test-Trace=true;X-Timestamp=$(date +%s)" "$TMPFILE2" "$ALIAS/$TEST_BUCKET/with-metadata.txt" >/dev/null 2>&1 || true
sleep 0.1

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "ERROR OPERATIONS (for error filtering testing)"
echo "═══════════════════════════════════════════════════════════════"

MISSING_KEYS=(missing-a missing-b missing-c)

echo "[1/8] 404 Not Found - GetObject on non-existent objects"
for key in "${MISSING_KEYS[@]}"; do
    for attempt in $(seq 1 4); do
        mc cat "$ALIAS/nonexistent-bucket/${key}-${attempt}" >/dev/null 2>&1 || true
        sleep 0.05
    done
done

echo "[2/8] 403 Access Denied - PutObject to read-only bucket"
BLOCKED_KEYS=(blocked-alpha blocked-beta blocked-gamma)
for key in "${BLOCKED_KEYS[@]}"; do
    for attempt in $(seq 1 3); do
        mc cp "$TMPFILE" "$ALIAS/$TEST_BUCKET/${key}-${attempt}" >/dev/null 2>&1 || true
        sleep 0.05
    done
done

echo "[3/8] 404 Not Found - HeadObject on non-existent keys"
for key in "${MISSING_KEYS[@]}"; do
    mc stat "$ALIAS/$TEST_BUCKET/${key}-ghost" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "[4/8] 409 Conflict - Duplicate bucket creation"
mc mb "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
mc mb "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
sleep 0.1

echo "[5/8] 404 Not Found - DeleteObject on non-existent objects"
for key in "${BLOCKED_KEYS[@]}"; do
    mc rm "$ALIAS/$TEST_BUCKET/${key}-phantom" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "[6/8] 404 Not Found - ListObjects on non-existent bucket"
for i in $(seq 1 3); do
    mc ls "$ALIAS/fake-bucket-xyz" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "[7/8] 404 Not Found - CopyObject from non-existent source"
mc cp "$ALIAS/$TEST_BUCKET/does-not-exist.txt" "$ALIAS/$TEST_BUCKET/copy-failed.txt" >/dev/null 2>&1 || true
sleep 0.1

echo "[8/8] Mixed operations - rapid successive calls"
for i in $(seq 1 10); do
    mc cat "$ALIAS/$TEST_BUCKET/data-alpha-1.txt" >/dev/null 2>&1 || true
    mc cat "$ALIAS/$TEST_BUCKET/nonexistent.txt" >/dev/null 2>&1 || true
    sleep 0.03
done

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "CLEANUP"
echo "═══════════════════════════════════════════════════════════════"

echo "Cleaning up temporary resources..."
mc policy set none "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
mc rm -r --force "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
rm -f "$TMPFILE" "$TMPFILE2"

echo ""
echo "✓ Done! Generated diverse API calls including:"
echo "  - Successful operations: PutObject, GetObject, ListObjects, HeadObject, CopyObject, etc."
echo "  - Error operations: 404 Not Found, 403 Access Denied, 409 Conflict"
echo ""
echo "Next steps:"
echo "  1. Run trace with all requests (default):"
echo "     ./mc-tool web → /minio-webtool/tracing/analyzer"
echo "  2. Run trace with errors only (check the 'Trace errors only' checkbox)"
echo "  3. Or use CLI: ./mc-tool trace $ALIAS --duration 10s -v"
echo ""
