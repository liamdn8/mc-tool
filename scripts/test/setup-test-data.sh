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
BUCKETS=("test-bucket-1" "test-bucket-2" "shared-bucket" "compare-test" "performance-test")

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

# Scenario 1: Many files only on first alias (for pagination testing)
echo -e "${BLUE}📋 Scenario 1: Creating 25+ files only on $ALIAS1${NC}"
upload_content "$ALIAS1" "test-bucket-1" "only-on-first/readme.txt" "This file exists only on $ALIAS1"
upload_content "$ALIAS1" "test-bucket-1" "docs/manual.pdf" "PDF Manual content - only on $ALIAS1"
upload_content "$ALIAS1" "test-bucket-1" "config/settings.json" '{"server": "'$ALIAS1'", "type": "primary", "version": "1.0"}'

# Create multiple documents for pagination testing
for i in {1..15}; do
    upload_content "$ALIAS1" "test-bucket-1" "documents/doc_$(printf "%02d" $i).txt" "Document $i - Important content that exists only on $ALIAS1. Created at $(date). This is document number $i in the pagination test series."
done

# Create multiple log files
for i in {1..8}; do
    upload_content "$ALIAS1" "test-bucket-1" "logs/app_$(printf "%02d" $i).log" "$(date) - Application log file $i
ERROR: Connection failed to database
WARN: Retrying connection attempt $i
INFO: Service started successfully
DEBUG: Processing user request $i"
done

# Create multiple config files
for i in {1..5}; do
    upload_content "$ALIAS1" "test-bucket-1" "configs/env_$i.conf" "[environment_$i]
debug=true
port=$((8000 + i))
database_url=db_$i.example.com
created_by=$ALIAS1
timestamp=$(date)"
done

# Scenario 2: Many files only on second alias (if available)
if [ "$ALIAS2" != "$ALIAS1" ]; then
    echo -e "${BLUE}📋 Scenario 2: Creating 20+ files only on $ALIAS2${NC}"
    upload_content "$ALIAS2" "test-bucket-1" "only-on-second/backup.txt" "This file exists only on $ALIAS2"
    upload_content "$ALIAS2" "test-bucket-1" "backup/data.sql" "Database backup - only on $ALIAS2"
    upload_content "$ALIAS2" "test-bucket-1" "logs/system.log" "System logs from $ALIAS2"
    
    # Create multiple backup files
    for i in {1..12}; do
        upload_content "$ALIAS2" "test-bucket-1" "backups/backup_$(date +%Y%m%d)_$(printf "%02d" $i).sql" "-- Database backup $i for $ALIAS2
-- Created: $(date)
-- Tables: users, orders, products, logs
-- Records: $((RANDOM % 10000 + 1000))
CREATE TABLE backup_$i (id INT, data TEXT);"
    done
    
    # Create multiple report files
    for i in {1..8}; do
        upload_content "$ALIAS2" "test-bucket-1" "reports/monthly_report_$(printf "%02d" $i).csv" "date,metric,value,alias
$(date +%Y-%m-%d),sales,$((RANDOM % 50000 + 10000)),$ALIAS2
$(date +%Y-%m-%d),users,$((RANDOM % 1000 + 100)),$ALIAS2
$(date +%Y-%m-%d),revenue,$((RANDOM % 100000 + 20000)),$ALIAS2"
    done
    
    # Create temp files
    for i in {1..6}; do
        upload_content "$ALIAS2" "test-bucket-1" "temp/temp_$i.tmp" "Temporary file $i for $ALIAS2 - $(date) - Random data: $(head -c 50 /dev/urandom | base64 | tr -d '\n')"
    done
fi

# Scenario 3: Many files with different content (30+ files for pagination testing)
echo -e "${BLUE}📋 Scenario 3: Creating 30+ files with different content${NC}"
for alias in "${AVAILABLE_ALIASES[@]}"; do
    upload_content "$alias" "shared-bucket" "data/config.json" '{"alias": "'$alias'", "timestamp": "'$(date)'", "version": "1.'$((RANDOM % 10))'", "environment": "test"}'
    upload_content "$alias" "shared-bucket" "data/large-file.txt" "Large file content for $alias - Generated at $(date) - Size simulation: $(head -c 1000 /dev/zero | tr '\0' 'A')"
    upload_content "$alias" "shared-bucket" "status/health.json" '{"alias": "'$alias'", "status": "healthy", "last_check": "'$(date)'", "metrics": {"cpu": '$((RANDOM % 100))', "memory": '$((RANDOM % 100))'}}'
    
    # Create multiple API response files with different content
    for i in {1..15}; do
        upload_content "$alias" "shared-bucket" "api/response_$(printf "%02d" $i).json" '{
    "alias": "'$alias'",
    "endpoint": "/api/data/'$i'",
    "response_time": '$((RANDOM % 500 + 50))',
    "timestamp": "'$(date)'",
    "data": {
        "id": '$i',
        "status": "active",
        "server": "'$alias'",
        "random_value": '$((RANDOM % 1000))'
    },
    "version": "2.'$i'.'$((RANDOM % 10))'"
}'
    done
    
    # Create different database files
    for i in {1..8}; do
        upload_content "$alias" "shared-bucket" "database/table_$(printf "%02d" $i).sql" "-- Table $i for $alias
-- Generated: $(date)
-- Rows: $((RANDOM % 1000 + 100))

CREATE TABLE table_$i (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    alias VARCHAR(50) DEFAULT '$alias',
    created_at TIMESTAMP DEFAULT NOW(),
    random_data INTEGER DEFAULT $((RANDOM % 10000))
);"
    done
    
    # Create different log entries
    for i in {1..10}; do
        upload_content "$alias" "shared-bucket" "logs/service_$(printf "%02d" $i).log" "$(date) [$alias] [service_$i] INFO: Service started
$(date) [$alias] [service_$i] DEBUG: Configuration loaded from $alias
$(date) [$alias] [service_$i] WARN: Memory usage at $((RANDOM % 80 + 10))%
$(date) [$alias] [service_$i] INFO: Processing request #$((RANDOM % 1000))
$(date) [$alias] [service_$i] ERROR: Connection timeout after $((RANDOM % 30 + 5))s"
    done
done

# Scenario 4: Identical files across aliases
echo -e "${BLUE}📋 Scenario 4: Identical files across aliases${NC}"
SHARED_CONTENT="This is a shared file that should be identical across all sites. Content created at $(date)"
for alias in "${AVAILABLE_ALIASES[@]}"; do
    upload_content "$alias" "shared-bucket" "shared/common.txt" "$SHARED_CONTENT"
    upload_content "$alias" "shared-bucket" "shared/reference.json" '{"type": "reference", "shared": true, "content": "identical across all sites"}'
done

# Scenario 5: Large folder structures with many nested files (50+ files)
echo -e "${BLUE}📋 Scenario 5: Creating complex folder structures with 50+ files${NC}"
for alias in "${AVAILABLE_ALIASES[@]}"; do
    # Monthly reports for a year
    for month in {01..12}; do
        upload_content "$alias" "test-bucket-2" "data/2024/$month/monthly_report.txt" "Monthly report for 2024-$month from $alias
Generated: $(date)
Records processed: $((RANDOM % 50000 + 10000))
Revenue: $((RANDOM % 1000000 + 100000))
Users: $((RANDOM % 10000 + 1000))"
        
        upload_content "$alias" "test-bucket-2" "data/2024/$month/summary.json" '{
    "month": "'$month'",
    "year": "2024",
    "alias": "'$alias'",
    "metrics": {
        "total_users": '$((RANDOM % 10000 + 1000))',
        "active_sessions": '$((RANDOM % 5000 + 500))',
        "revenue": '$((RANDOM % 1000000 + 100000))',
        "errors": '$((RANDOM % 100))'
    },
    "generated_at": "'$(date)'"
}'
    done
    
    # Weekly data files for each quarter
    for quarter in {1..4}; do
        for week in {01..13}; do
            upload_content "$alias" "test-bucket-2" "data/2024/Q$quarter/week_$week.csv" "date,transactions,revenue,users,alias
2024-$((quarter*3-2))-$(printf "%02d" $((week))),$(((RANDOM % 1000) + 100)),$((RANDOM % 50000 + 5000)),$((RANDOM % 1000 + 100)),$alias
2024-$((quarter*3-1))-$(printf "%02d" $((week))),$(((RANDOM % 1000) + 100)),$((RANDOM % 50000 + 5000)),$((RANDOM % 1000 + 100)),$alias
2024-$((quarter*3))-$(printf "%02d" $((week))),$(((RANDOM % 1000) + 100)),$((RANDOM % 50000 + 5000)),$((RANDOM % 1000 + 100)),$alias"
        done
    done
    
    # Alias-specific folders with multiple files
    upload_content "$alias" "test-bucket-2" "site-specific/$alias/config.ini" "[settings]
server_name=$alias
created_at=$(date)
port=$((9000 + ${#alias}))
debug=true"
    
    upload_content "$alias" "test-bucket-2" "site-specific/$alias/logs/access.log" "$(date) - Access log entry for $alias"
    upload_content "$alias" "test-bucket-2" "site-specific/$alias/data/metrics.csv" "timestamp,alias,cpu,memory,disk
$(date),$alias,$((RANDOM % 100)),$((RANDOM % 100)),$((RANDOM % 100))"
    
    # Multiple cache files
    for i in {1..8}; do
        upload_content "$alias" "test-bucket-2" "site-specific/$alias/cache/cache_$i.dat" "Cache file $i for $alias
Created: $(date)
Size: $((RANDOM % 1000000 + 100000)) bytes
Type: session_cache
TTL: $((RANDOM % 3600 + 300)) seconds"
    done
    
    # Multiple user data files
    for i in {1..6}; do
        upload_content "$alias" "test-bucket-2" "site-specific/$alias/users/user_$i.json" '{
    "user_id": '$i',
    "alias": "'$alias'",
    "created_at": "'$(date)'",
    "profile": {
        "name": "User '$i'",
        "email": "user'$i'@'$alias'.example.com",
        "last_login": "'$(date)'",
        "preferences": {
            "theme": "dark",
            "language": "en",
            "notifications": true
        }
    },
    "stats": {
        "login_count": '$((RANDOM % 1000 + 10))',
        "last_activity": "'$(date)'"
    }
}'
    done
done

# Create extensive unique content for first alias
if [ ${#AVAILABLE_ALIASES[@]} -gt 1 ]; then
    echo -e "${BLUE}📋 Scenario 6: Creating 15+ unique files on $ALIAS2${NC}"
    upload_content "$ALIAS2" "test-bucket-2" "unique-to-second/special-data.txt" "This data only exists on $ALIAS2"
    upload_content "$ALIAS2" "test-bucket-2" "unique-to-second/cache/temp.dat" "Temporary cache data for $ALIAS2"
    
    # Multiple unique documents
    for i in {1..10}; do
        upload_content "$ALIAS2" "test-bucket-2" "unique-to-second/documents/unique_doc_$(printf "%02d" $i).txt" "Unique document $i for $ALIAS2
This content is special and only exists on this alias.
Created: $(date)
Document ID: UNIQUE-$ALIAS2-$(printf "%03d" $i)
Content: This is a special document that demonstrates files that exist only on $ALIAS2.
Random data: $(head -c 100 /dev/urandom | base64 | tr -d '\n' | head -c 50)"
    done
    
    # Unique configuration files
    for i in {1..5}; do
        upload_content "$ALIAS2" "test-bucket-2" "unique-to-second/configs/app_config_$i.yaml" "# Application Configuration $i for $ALIAS2
# This config only exists on $ALIAS2

app:
  name: \"unique-app-$i\"
  version: \"1.$i.0\"
  environment: \"$ALIAS2\"
  
database:
  host: \"$ALIAS2-db-$i.local\"
  port: $((5432 + i))
  name: \"app_$i\"
  
cache:
  enabled: true
  ttl: $((3600 * i))
  
features:
  feature_$i: true
  advanced_mode: $([ $((i % 2)) -eq 0 ] && echo true || echo false)
"
    done
fi

# Scenario 7: Extensive version differences simulation (20+ files)
echo -e "${BLUE}📋 Scenario 7: Creating 20+ files with version differences${NC}"
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

    # Multiple API version files
    for api_ver in {1..8}; do
        upload_content "$alias" "compare-test" "api/v$api_ver/endpoint.json" '{
    "version": "'$version'",
    "api_version": "v'$api_ver'",
    "alias": "'$alias'",
    "endpoints": [
        "/api/v'$api_ver'/users",
        "/api/v'$api_ver'/orders", 
        "/api/v'$api_ver'/products"
    ],
    "features": {
        "pagination": true,
        "filtering": true,
        "sorting": '$([ $((api_ver % 2)) -eq 0 ] && echo true || echo false)',
        "caching": '$([ $api_ver -gt 4 ] && echo true || echo false)'
    },
    "build_info": {
        "compiled_at": "'$(date)'",
        "compiler_version": "1.'$i'.'$api_ver'",
        "target_alias": "'$alias'"
    }
}'
    done
    
    # Multiple library version files
    for lib in {1..6}; do
        upload_content "$alias" "compare-test" "libraries/lib_$(printf "%02d" $lib).js" "/**
 * Library $lib for $alias
 * Version: $version.$lib
 * Compiled: $(date)
 * 
 * This library has different implementations per alias
 */

const LIB_$lib = {
    version: '$version.$lib',
    alias: '$alias',
    buildNumber: $((RANDOM % 1000 + 100)),
    features: {
        caching: $([ $((lib % 2)) -eq 0 ] && echo true || echo false),
        logging: true,
        encryption: $([ $lib -gt 3 ] && echo true || echo false)
    },
    
    init: function() {
        console.log('Library $lib initialized for $alias');
        console.log('Version: ' + this.version);
        return this.buildNumber + $i;
    }
};"
    done
    
    # Multiple configuration files with version differences
    for config_num in {1..4}; do
        upload_content "$alias" "compare-test" "configs/service_$config_num.toml" "# Service $config_num Configuration for $alias
# Version: $version
# Generated: $(date)

[service]
name = \"service_$config_num\"
version = \"$version\"
alias = \"$alias\"
port = $((8000 + config_num + i))

[database]
host = \"$alias-db$config_num.local\"
port = $((5432 + config_num))
pool_size = $((10 + i * 5))

[features]
feature_$config_num = true
advanced_mode = $([ $((config_num % 2)) -eq 0 ] && echo true || echo false)
debug_enabled = $([ $i -eq 0 ] && echo true || echo false)

[cache]
enabled = $([ $config_num -gt 2 ] && echo true || echo false)
ttl = $((3600 * config_num))
        max_size = \"$((100 + i * 50))MB\"
"
    done
done

# Scenario 8: Performance testing with many small files (50+ files)
echo -e "${BLUE}📋 Scenario 8: Creating performance test data with 50+ small files${NC}"
for alias in "${AVAILABLE_ALIASES[@]}"; do
    # Create many small files in performance-test bucket if it doesn't exist  
    if ! mc ls "$alias/performance-test" >/dev/null 2>&1; then
        mc mb "$alias/performance-test" 2>/dev/null || true
    fi
    
    # Generate different types of small files
    for i in {001..025}; do
        upload_content "$alias" "performance-test" "batch_1/file_$i.txt" "Performance test file $i for $alias
Content varies per alias: $((RANDOM % 1000 + i))
Generated: $(date)
Batch: 1
Size category: small"
    done
    
    for i in {026..050}; do
        upload_content "$alias" "performance-test" "batch_2/data_$i.json" '{
    "id": '$i',
    "alias": "'$alias'",
    "batch": 2,
    "performance_data": {
        "random_value": '$((RANDOM % 10000))',
        "timestamp": "'$(date)'",
        "iteration": '$i'
    },
    "metadata": {
        "size": "medium",
        "type": "performance_test"
    }
}'
    done
    
    # Create different CSV files for each alias
    for i in {051..070}; do
        upload_content "$alias" "performance-test" "batch_3/metrics_$i.csv" "id,alias,value,timestamp,batch
$i,$alias,$((RANDOM % 1000)),'$(date)',3
$((i+1000)),$alias,$((RANDOM % 1000)),'$(date)',3
$((i+2000)),$alias,$((RANDOM % 1000)),'$(date)',3"
    done
done

echo ""
echo -e "${GREEN}🎉 Test data setup completed successfully!${NC}"
echo ""
echo -e "${YELLOW}📊 Extensive test scenarios created for pagination testing:${NC}"
echo -e "${BLUE}1. Files only on $ALIAS1:${NC} test-bucket-1/ (25+ files: documents, logs, configs)"

if [ "$ALIAS2" != "$ALIAS1" ]; then
    echo -e "${BLUE}2. Files only on $ALIAS2:${NC} test-bucket-1/ (20+ files: backups, reports, temp files)"
fi

echo -e "${BLUE}3. Different content:${NC} shared-bucket/ (30+ files: api responses, database files, logs)"
echo -e "${BLUE}4. Identical files:${NC} shared-bucket/shared/ (common.txt, reference.json)"
echo -e "${BLUE}5. Large folder structures:${NC} test-bucket-2/ (50+ files: monthly/quarterly data, user files)"
echo -e "${BLUE}6. Site-specific content:${NC} test-bucket-2/site-specific/ (cache, user data per alias)"
echo -e "${BLUE}7. Version differences:${NC} compare-test/ (20+ files: APIs, libraries, configs)"

if [ ${#AVAILABLE_ALIASES[@]} -gt 1 ]; then
    echo -e "${BLUE}8. Unique extensive content:${NC} test-bucket-2/unique-to-second/ (15+ files only on $ALIAS2)"
fi

echo ""
echo -e "${GREEN}📈 Pagination test scenarios:${NC}"
echo -e "${BLUE}• Small pages (10 items):${NC} Perfect for testing basic pagination"
echo -e "${BLUE}• Medium pages (25 items):${NC} Good for performance testing"
echo -e "${BLUE}• Large pages (50+ items):${NC} Test edge cases and max page sizes"
echo -e "${BLUE}• Different categories:${NC} Only in source, only in dest, different content"

echo ""
echo -e "${GREEN}🔍 You can now test the Compare functionality with extensive pagination!${NC}"
echo -e "${BLUE}💡 Recommended test comparisons:${NC}"
if [ ${#AVAILABLE_ALIASES[@]} -gt 1 ]; then
    echo "   • $ALIAS1 vs $ALIAS2 in test-bucket-1 (45+ files total - great for pagination)"
    echo "   • $ALIAS1 vs $ALIAS2 in shared-bucket (30+ different files)"
    echo "   • $ALIAS1 vs $ALIAS2 in test-bucket-2 (50+ files with complex structure)"
    echo "   • $ALIAS1 vs $ALIAS2 in compare-test (20+ version differences)"
fi

if [ ${#AVAILABLE_ALIASES[@]} -gt 2 ]; then
    echo "   • $ALIAS2 vs $ALIAS3 in any bucket for different pagination patterns"
fi

echo ""
echo -e "${YELLOW}📊 Total files created per bucket (approximate):${NC}"
echo -e "${BLUE}• test-bucket-1:${NC} 45+ files (25 on $ALIAS1, 20+ on $ALIAS2)"
echo -e "${BLUE}• shared-bucket:${NC} 30+ different files + identical files"
echo -e "${BLUE}• test-bucket-2:${NC} 50+ files with complex nesting + unique content"
echo -e "${BLUE}• compare-test:${NC} 20+ files with version differences"
echo -e "${BLUE}• performance-test:${NC} 70+ small files in 3 batches for performance testing"

echo ""
echo -e "${YELLOW}📋 Available aliases for testing: ${AVAILABLE_ALIASES[*]}${NC}"
echo -e "${YELLOW}📋 Available buckets: ${BUCKETS[*]}${NC}"
echo ""
echo -e "${GREEN}✅ Setup complete! Start the web UI and test pagination in the Compare operations.${NC}"
echo -e "${BLUE}💡 Try different page sizes: 10 (default), 25, 50, 100 to test pagination performance.${NC}"