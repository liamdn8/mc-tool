#!/usr/bin/env bash
# Helper script to trigger predictable MinIO errors for mc-tool trace testing.

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

echo "Generating trace-friendly errors against alias '$ALIAS'..."

echo "1) Repeated 404 errors for a non-existent bucket/object"
MISSING_KEYS=(mc-trace-missing-a mc-trace-missing-b mc-trace-missing-c)
for key in "${MISSING_KEYS[@]}"; do
    for attempt in $(seq 1 6); do
        mc cat "$ALIAS/nonexistent-bucket/${key}-${attempt}" >/dev/null 2>&1 || true
        sleep 0.05
    done
done

echo "2) Access denied errors by uploading into a read-only bucket"
TEST_BUCKET="trace-readonly-$(date +%s)"
TMPFILE=$(mktemp)
printf 'trace error generator %s\n' "$(date --iso-8601=seconds)" > "$TMPFILE"

mc mb "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
mc policy set download "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true

BLOCKED_KEYS=(blocked-alpha blocked-beta blocked-gamma)
for key in "${BLOCKED_KEYS[@]}"; do
    for attempt in $(seq 1 4); do
        mc cp "$TMPFILE" "$ALIAS/$TEST_BUCKET/${key}-${attempt}" >/dev/null 2>&1 || true
        sleep 0.05
    done
done

echo "3) Object stat on non-existent keys"
for key in "${MISSING_KEYS[@]}"; do
    mc stat "$ALIAS/nonexistent-bucket/${key}-stat" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "4) Duplicate bucket creation to trigger conflict responses"
mc mb "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
mc mb "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
sleep 0.1

echo "5) Deleting objects that do not exist"
for key in "${BLOCKED_KEYS[@]}"; do
    mc rm "$ALIAS/$TEST_BUCKET/${key}-ghost" >/dev/null 2>&1 || true
    sleep 0.05
done

echo "Cleaning up temporary resources"
mc policy set none "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
mc rm -r --force "$ALIAS/$TEST_BUCKET" >/dev/null 2>&1 || true
rm -f "$TMPFILE"

echo "Done. Re-run './mc-tool trace $ALIAS --duration 10s -v' to capture the generated errors."
