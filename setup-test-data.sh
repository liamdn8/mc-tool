#!/bin/bash

# MC-Tool Test Data Setup Script
# This script creates test buckets and files for comparing operations

set -e

echo "🔧 Setting up test data for MC-Tool Compare operations..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ROOT_USER="minioadmin"
ROOT_PASSWORD="minioadmin123"

# Function to check if alias exists and is reachable
check_alias() {
    local alias=$1
    if ! mc alias list "$alias" >/dev/null 2>&1; then
        echo -e "${RED}❌ Alias '$alias' not found. Please configure it first.${NC}"
        return 1
    fi
    
    if ! mc admin info "$alias" >/dev/null 2>&1; then
        echo -e "${RED}❌ Cannot connect to '$alias'. Please check server status.${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ Alias '$alias' is configured and reachable${NC}"
    return 0
}

# Function to create bucket if it doesn't exist
create_bucket() {
    local alias=$1
    local bucket=$2
    echo -e "${BLUE}📁 Creating bucket '$bucket' on '$alias'...${NC}"
    
    if mc ls "$alias/$bucket" >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Bucket '$bucket' already exists on '$alias'${NC}"
    else
        mc mb "$alias/$bucket"
        echo -e "${GREEN}✅ Created bucket '$bucket' on '$alias'${NC}"
    fi
}

# Function to upload content as file
upload_content() {
    local alias=$1
    local bucket=$2
    local remote_path=$3
    local content="$4"
    
    echo -e "${BLUE}⬆️  Uploading to '$alias/$bucket/$remote_path'...${NC}"
    echo "$content" | mc pipe "$alias/$bucket/$remote_path"
    echo -e "${GREEN}✅ Uploaded to '$alias/$bucket/$remote_path'${NC}"
}

# Check if mc command is available
if ! command -v mc &> /dev/null; then
    echo -e "${RED}❌ MinIO client 'mc' is not installed or not in PATH${NC}"
    echo -e "${BLUE}💡 Install: wget https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x mc && sudo mv mc /usr/local/bin/${NC}"
    exit 1
fi

echo -e "${GREEN}✅ MinIO client 'mc' is available${NC}"

# Try to find available aliases automatically
echo -e "${YELLOW}🔍 Detecting available aliases...${NC}"
AVAILABLE_ALIASES=()

# Check for site1-site6 pattern
for i in {1..6}; do
    alias_name="site$i"
    if mc alias list "$alias_name" >/dev/null 2>&1 && mc admin info "$alias_name" >/dev/null 2>&1; then
        AVAILABLE_ALIASES+=("$alias_name")
        echo -e "${GREEN}✅ Found working alias: $alias_name${NC}"
    fi
done

# If no site1-6, check for common patterns
if [ ${#AVAILABLE_ALIASES[@]} -eq 0 ]; then
    for alias_name in local minio test dev prod site4 site5 site6; do
        if mc alias list "$alias_name" >/dev/null 2>&1 && mc admin info "$alias_name" >/dev/null 2>&1; then
            AVAILABLE_ALIASES+=("$alias_name")
            echo -e "${GREEN}✅ Found working alias: $alias_name${NC}"
        fi
    done
fi

if [ ${#AVAILABLE_ALIASES[@]} -eq 0 ]; then
    echo -e "${RED}❌ No working MinIO aliases found!${NC}"
    echo -e "${BLUE}💡 Please configure at least one alias first:${NC}"
    echo "   mc alias set site1 http://localhost:9001 minioadmin minioadmin"
    echo ""
    echo -e "${BLUE}💡 Or run: ./setup-minio-test.sh to create test environment${NC}"
    exit 1
fi

if [ ${#AVAILABLE_ALIASES[@]} -lt 2 ]; then
    echo -e "${YELLOW}⚠️  Only ${#AVAILABLE_ALIASES[@]} alias found. You need at least 2 for meaningful comparisons.${NC}"
    echo -e "${BLUE}💡 Consider adding more aliases for better testing.${NC}"
fi

echo -e "${GREEN}🎯 Found ${#AVAILABLE_ALIASES[@]} working aliases: ${AVAILABLE_ALIASES[*]}${NC}"

echo ""
echo -e "${YELLOW}📊 Creating test buckets and data...${NC}"

# Test buckets to create
BUCKETS=("test-bucket-1" "test-bucket-2" "shared-bucket" "compare-test")

# Create buckets on all available aliases
for alias in "${AVAILABLE_ALIASES[@]}"; do
    for bucket in "${BUCKETS[@]}"; do
        create_bucket "$alias" "$bucket"
    done
done

echo ""
echo -e "${YELLOW}📝 Creating test scenarios...${NC}"

# Get first few aliases for scenarios
ALIAS1="${AVAILABLE_ALIASES[0]}"
ALIAS2="${AVAILABLE_ALIASES[1]:-$ALIAS1}"
ALIAS3="${AVAILABLE_ALIASES[2]:-$ALIAS1}"

echo -e "${BLUE}📋 Primary aliases for testing: $ALIAS1, $ALIAS2, $ALIAS3${NC}"

# Scenario 1: Files only on first alias
echo -e "${BLUE}📋 Scenario 1: Files only on $ALIAS1${NC}"
upload_content "$ALIAS1" "test-bucket-1" "only-on-first/readme.txt" "This file exists only on $ALIAS1"
upload_content "$ALIAS1" "test-bucket-1" "docs/manual.pdf" "PDF Manual content - only on $ALIAS1"
upload_content "$ALIAS1" "test-bucket-1" "config/settings.json" '{"server": "'$ALIAS1'", "type": "primary", "version": "1.0"}'

# Scenario 2: Files only on second alias (if available)
if [ "$ALIAS2" != "$ALIAS1" ]; then
    echo -e "${BLUE}📋 Scenario 2: Files only on $ALIAS2${NC}"
    upload_content "$ALIAS2" "test-bucket-1" "only-on-second/backup.txt" "This file exists only on $ALIAS2"
    upload_content "$ALIAS2" "test-bucket-1" "backup/data.sql" "Database backup - only on $ALIAS2"
    upload_content "$ALIAS2" "test-bucket-1" "logs/system.log" "System logs from $ALIAS2"
fi

# Scenario 3: Same files with different content
echo -e "${BLUE}📋 Scenario 3: Same files with different content${NC}"
for alias in "${AVAILABLE_ALIASES[@]}"; do
    upload_content "$alias" "shared-bucket" "data/config.json" '{"alias": "'$alias'", "timestamp": "'$(date)'", "version": "1.'$((RANDOM % 10))'", "environment": "test"}'
    upload_content "$alias" "shared-bucket" "data/large-file.txt" "Large file content for $alias - Generated at $(date) - Size simulation: $(head -c 1000 /dev/zero | tr '\0' 'A')"
    upload_content "$alias" "shared-bucket" "status/health.json" '{"alias": "'$alias'", "status": "healthy", "last_check": "'$(date)'", "metrics": {"cpu": '$((RANDOM % 100))', "memory": '$((RANDOM % 100))'}}'
done

# Scenario 4: Identical files across aliases
echo -e "${BLUE}📋 Scenario 4: Identical files across aliases${NC}"
SHARED_CONTENT="This is a shared file that should be identical across all sites. Content created at $(date)"
for alias in "${AVAILABLE_ALIASES[@]}"; do
    upload_content "$alias" "shared-bucket" "shared/common.txt" "$SHARED_CONTENT"
    upload_content "$alias" "shared-bucket" "shared/reference.json" '{"type": "reference", "shared": true, "content": "identical across all sites"}'
done

# Scenario 5: Folder structures with nested content
echo -e "${BLUE}📋 Scenario 5: Complex folder structures${NC}"
for alias in "${AVAILABLE_ALIASES[@]}"; do
    # Common folder structure
    upload_content "$alias" "test-bucket-2" "data/2024/january/report.txt" "Monthly report for January 2024 from $alias"
    upload_content "$alias" "test-bucket-2" "data/2024/february/report.txt" "Monthly report for February 2024 from $alias"
    upload_content "$alias" "test-bucket-2" "data/2024/march/report.txt" "Monthly report for March 2024 from $alias"
    
    # Alias-specific folders
    upload_content "$alias" "test-bucket-2" "site-specific/$alias/config.ini" "[settings]
server_name=$alias
created_at=$(date)
port=$((9000 + ${#alias}))
debug=true"
    
    upload_content "$alias" "test-bucket-2" "site-specific/$alias/logs/access.log" "$(date) - Access log entry for $alias"
    upload_content "$alias" "test-bucket-2" "site-specific/$alias/data/metrics.csv" "timestamp,alias,cpu,memory,disk
$(date),$alias,$((RANDOM % 100)),$((RANDOM % 100)),$((RANDOM % 100))"
done

# Create unique content for first alias
if [ ${#AVAILABLE_ALIASES[@]} -gt 1 ]; then
    echo -e "${BLUE}📋 Scenario 6: Unique content on $ALIAS2${NC}"
    upload_content "$ALIAS2" "test-bucket-2" "unique-to-second/special-data.txt" "This data only exists on $ALIAS2"
    upload_content "$ALIAS2" "test-bucket-2" "unique-to-second/cache/temp.dat" "Temporary cache data for $ALIAS2"
fi

# Scenario 7: Version differences simulation
echo -e "${BLUE}📋 Scenario 7: Version differences${NC}"
for i in "${!AVAILABLE_ALIASES[@]}"; do
    alias="${AVAILABLE_ALIASES[$i]}"
    version="1.$i"
    upload_content "$alias" "compare-test" "app/version.txt" "Application Version: $version
Build: $(date)
Alias: $alias
Features: basic,advanced,premium"

    upload_content "$alias" "compare-test" "app/changelog.md" "# Changelog

## Version $version
- Updated for $alias
- Build date: $(date)
- New features: feature_$i

## Previous versions
- 1.0: Initial release"
done

echo ""
echo -e "${GREEN}🎉 Test data setup completed successfully!${NC}"
echo ""
echo -e "${YELLOW}📊 Test scenarios created:${NC}"
echo -e "${BLUE}1. Files only on $ALIAS1:${NC} test-bucket-1/only-on-first/, test-bucket-1/docs/, test-bucket-1/config/"

if [ "$ALIAS2" != "$ALIAS1" ]; then
    echo -e "${BLUE}2. Files only on $ALIAS2:${NC} test-bucket-1/only-on-second/, test-bucket-1/backup/, test-bucket-1/logs/"
fi

echo -e "${BLUE}3. Different content:${NC} shared-bucket/data/ (config.json, large-file.txt, health.json)"
echo -e "${BLUE}4. Identical files:${NC} shared-bucket/shared/ (common.txt, reference.json)"
echo -e "${BLUE}5. Folder structures:${NC} test-bucket-2/data/2024/ (nested monthly folders)"
echo -e "${BLUE}6. Site-specific:${NC} test-bucket-2/site-specific/ (different per alias)"
echo -e "${BLUE}7. Version differences:${NC} compare-test/app/ (version.txt, changelog.md)"

if [ ${#AVAILABLE_ALIASES[@]} -gt 1 ]; then
    echo -e "${BLUE}8. Unique content:${NC} test-bucket-2/unique-to-second/ (only on $ALIAS2)"
fi

echo ""
echo -e "${GREEN}🔍 You can now test the Compare functionality in MC-Tool Web UI!${NC}"
echo -e "${BLUE}💡 Try comparing:${NC}"
if [ ${#AVAILABLE_ALIASES[@]} -gt 1 ]; then
    echo "   • $ALIAS1 vs $ALIAS2 in test-bucket-1 (files only in source/dest)"
    echo "   • $ALIAS1 vs $ALIAS2 in shared-bucket (different vs identical content)"
    echo "   • $ALIAS1 vs $ALIAS2 in test-bucket-2 (folder differences)"
    echo "   • $ALIAS1 vs $ALIAS2 in compare-test (version differences)"
fi

if [ ${#AVAILABLE_ALIASES[@]} -gt 2 ]; then
    echo "   • $ALIAS2 vs $ALIAS3 in any bucket for more variations"
fi

echo ""
echo -e "${YELLOW}📋 Available aliases for testing: ${AVAILABLE_ALIASES[*]}${NC}"
echo -e "${YELLOW}📋 Available buckets: ${BUCKETS[*]}${NC}"
echo ""
echo -e "${GREEN}✅ Setup complete! Start the web UI and navigate to Operations page.${NC}"