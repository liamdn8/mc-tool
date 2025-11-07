#!/bin/bash

echo "=== Setting up MinIO Site Replication ==="
echo ""

# Check if we have at least 2 sites
SITE_COUNT=$(mc alias list --json | jq -r 'select(.alias != "s3" and .alias != "gcs" and .alias != "play") | .alias' | wc -l)

if [ "$SITE_COUNT" -lt 2 ]; then
    echo "❌ Need at least 2 MinIO sites configured"
    echo "Current aliases: $SITE_COUNT"
    exit 1
fi

echo "Found $SITE_COUNT MinIO aliases"
echo ""

# Get list of sites
SITES=$(mc alias list --json | jq -r 'select(.alias != "s3" and .alias != "gcs" and .alias != "play") | .alias' | tr '\n' ' ')
echo "Sites: $SITES"
echo ""

# Ask user to confirm
echo "⚠️  This will enable site replication between these sites:"
for site in $SITES; do
    URL=$(mc alias list --json | jq -r "select(.alias==\"$site\") | .URL")
    echo "  - $site ($URL)"
done
echo ""
echo "⚠️  Important: All sites must be on the same MinIO version"
echo "⚠️  All sites must have valid TLS certificates (or all use http)"
echo ""
read -p "Continue? (y/N) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    exit 0
fi

# Enable site replication
echo ""
echo "Enabling site replication..."
echo "Running: mc admin replicate add $SITES"
echo ""

mc admin replicate add $SITES

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Site replication enabled successfully!"
    echo ""
    echo "Checking status..."
    FIRST_SITE=$(echo $SITES | awk '{print $1}')
    mc admin replicate info "$FIRST_SITE"
    echo ""
    echo "🎉 Done! Refresh the Web UI to see the updated status."
    echo "   Sites will now show as 'Configured' (green)"
else
    echo ""
    echo "❌ Failed to enable site replication"
    echo ""
    echo "Common issues:"
    echo "  1. MinIO versions are different"
    echo "  2. TLS certificate issues"
    echo "  3. Network connectivity problems"
    echo "  4. Sites already in a different replication group"
    echo ""
    echo "Check MinIO logs for more details"
fi
