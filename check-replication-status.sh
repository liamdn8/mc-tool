#!/bin/bash

echo "=== MinIO Site Replication Status Check ==="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check mc is installed
if ! command -v mc &> /dev/null; then
    echo -e "${RED}❌ MinIO Client (mc) is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ MinIO Client (mc) is installed${NC}"
echo ""

# Get all aliases
echo "📋 Configured MinIO Aliases:"
echo "─────────────────────────────────────────"
mc alias list --json | jq -r '.alias' | while read -r alias; do
    if [ "$alias" != "s3" ] && [ "$alias" != "gcs" ] && [ "$alias" != "play" ]; then
        URL=$(mc alias list --json | jq -r "select(.alias==\"$alias\") | .URL")
        
        # Check replication status
        REPL_STATUS=$(mc admin replicate info "$alias" --json 2>/dev/null)
        
        if [ $? -eq 0 ]; then
            ENABLED=$(echo "$REPL_STATUS" | jq -r '.enabled // false')
            if [ "$ENABLED" = "true" ]; then
                STATUS="${GREEN}✓ Replication Enabled${NC}"
                SITE_NAME=$(echo "$REPL_STATUS" | jq -r '.siteName // "N/A"')
                SITE_COUNT=$(echo "$REPL_STATUS" | jq -r '.sites | length')
                echo -e "  ${GREEN}●${NC} $alias"
                echo -e "    URL: $URL"
                echo -e "    Status: $STATUS"
                echo -e "    Site Name: $SITE_NAME"
                echo -e "    Sites in Group: $SITE_COUNT"
            else
                echo -e "  ${YELLOW}●${NC} $alias"
                echo -e "    URL: $URL"
                echo -e "    Status: ${YELLOW}⚠ Replication Disabled${NC}"
            fi
        else
            echo -e "  ${RED}●${NC} $alias"
            echo -e "    URL: $URL"
            echo -e "    Status: ${RED}✗ Not Configured${NC}"
        fi
        echo ""
    fi
done

echo ""
echo "=== Replication Groups ==="
echo "─────────────────────────────────────────"

# Check for replication groups
FOUND_GROUP=false
mc alias list --json | jq -r '.alias' | while read -r alias; do
    if [ "$alias" != "s3" ] && [ "$alias" != "gcs" ] && [ "$alias" != "play" ]; then
        REPL_INFO=$(mc admin replicate info "$alias" --json 2>/dev/null)
        if [ $? -eq 0 ]; then
            ENABLED=$(echo "$REPL_INFO" | jq -r '.enabled // false')
            if [ "$ENABLED" = "true" ]; then
                if [ "$FOUND_GROUP" = false ]; then
                    echo -e "${GREEN}Found Site Replication Group:${NC}"
                    echo "$REPL_INFO" | jq -r '.sites[] | "  - \(.name) (\(.endpoint))"'
                    FOUND_GROUP=true
                    break
                fi
            fi
        fi
    fi
done

if [ "$FOUND_GROUP" = false ]; then
    echo -e "${YELLOW}No active site replication groups found${NC}"
fi

echo ""
echo "=== Quick Setup Guide ==="
echo "─────────────────────────────────────────"
echo "To enable site replication between sites:"
echo "  mc admin replicate add site1 site2 [site3...]"
echo ""
echo "To check replication status:"
echo "  mc admin replicate info site1"
echo ""
echo "To view in Web UI:"
echo "  ./mc-tool-new web"
echo "  Open: http://localhost:8080"
echo ""
