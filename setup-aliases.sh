#!/bin/bash

# MC-Tool Aliases Setup Script
# This script configures MinIO client aliases for testing

set -e

echo "🔧 Setting up MinIO client aliases for MC-Tool testing..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ROOT_USER="minioadmin"
ROOT_PASSWORD="minioadmin123"

# Check if mc command is available
if ! command -v mc &> /dev/null; then
    echo -e "${RED}❌ MinIO client 'mc' is not installed or not in PATH${NC}"
    echo -e "${BLUE}💡 Install with:${NC}"
    echo "   wget https://dl.min.io/client/mc/release/linux-amd64/mc"
    echo "   chmod +x mc"
    echo "   sudo mv mc /usr/local/bin/"
    exit 1
fi

echo -e "${GREEN}✅ MinIO client 'mc' is available${NC}"

# Function to configure alias
configure_alias() {
    local alias_name=$1
    local endpoint=$2
    local access_key=${3:-"$ROOT_USER"}
    local secret_key=${4:-"$ROOT_PASSWORD"}
    
    echo -e "${BLUE}🔧 Configuring alias '$alias_name'...${NC}"
    
    # Remove existing alias if it exists
    mc alias remove "$alias_name" >/dev/null 2>&1 || true
    
    # Add new alias
    if mc alias set "$alias_name" "$endpoint" "$access_key" "$secret_key" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Configured alias '$alias_name' -> $endpoint${NC}"
        
        # Test connectivity
        if mc admin info "$alias_name" >/dev/null 2>&1; then
            echo -e "${GREEN}🏓 Connection test successful for '$alias_name'${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠️  Warning: Cannot connect to '$alias_name'. Server might be down.${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ Failed to configure alias '$alias_name'${NC}"
        return 1
    fi
}

# Function to auto-detect running MinIO instances
detect_minio_instances() {
    echo -e "${YELLOW}� Auto-detecting MinIO instances...${NC}"
    
    local found_instances=()
    
    # Check common ports for MinIO
    for port in {9001..9010} {9000..9010}; do
        if timeout 2 bash -c "</dev/tcp/localhost/$port" >/dev/null 2>&1; then
            endpoint="http://localhost:$port"
            echo -e "${BLUE}📡 Found service on $endpoint${NC}"
            
            # Try to configure as test alias
            local alias_name="auto_port_$port"
            if configure_alias "$alias_name" "$endpoint" "$ROOT_USER" "$ROOT_PASSWORD" >/dev/null 2>&1; then
                found_instances+=("$alias_name:$endpoint")
                mc alias remove "$alias_name" >/dev/null 2>&1
            fi
        fi
    done
    
    echo "${found_instances[@]}"
}

# Check for existing aliases
echo -e "${YELLOW}🔍 Checking existing aliases...${NC}"
existing_aliases=$(mc alias list 2>/dev/null | grep -E "site[1-6]" | awk '{print $1}' | tr '\n' ' ' || echo "")

if [ -n "$existing_aliases" ]; then
    echo -e "${GREEN}✅ Found existing site aliases: $existing_aliases${NC}"
    echo -e "${BLUE}💡 Testing connectivity...${NC}"
    
    working_aliases=()
    for alias in $existing_aliases; do
        if mc admin info "$alias" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ $alias is working${NC}"
            working_aliases+=("$alias")
        else
            echo -e "${RED}❌ $alias is not reachable${NC}"
        fi
    done
    
    if [ ${#working_aliases[@]} -gt 0 ]; then
        echo -e "${GREEN}🎯 Found ${#working_aliases[@]} working aliases: ${working_aliases[*]}${NC}"
        echo -e "${BLUE}💡 You can skip to: ./setup-test-data.sh${NC}"
        exit 0
    fi
fi

# Check if setup-minio-test.sh created containers
echo -e "${YELLOW}🐳 Checking for existing MinIO containers...${NC}"
if command -v docker &> /dev/null; then
    containers=$(docker ps --format "table {{.Names}}" | grep -E "minio-site[1-6]" | tr '\n' ' ' || echo "")
    if [ -n "$containers" ]; then
        echo -e "${GREEN}✅ Found MinIO containers: $containers${NC}"
        echo -e "${BLUE}🔧 Configuring aliases for existing containers...${NC}"
        
        # Get host IP
        HOST_IP=$(hostname -I | awk '{print $1}')
        
        success_count=0
        for i in {1..6}; do
            container_name="minio-site$i"
            if docker ps --format "table {{.Names}}" | grep -q "$container_name"; then
                port=$((9000 + i))
                endpoint="http://$HOST_IP:$port"
                alias_name="site$i"
                
                if configure_alias "$alias_name" "$endpoint" "$ROOT_USER" "$ROOT_PASSWORD"; then
                    success_count=$((success_count + 1))
                fi
            fi
        done
        
        if [ $success_count -gt 0 ]; then
            echo -e "${GREEN}🎉 Successfully configured $success_count aliases!${NC}"
            echo -e "${BLUE}💡 Run: ./setup-test-data.sh to create test data${NC}"
            exit 0
        fi
    fi
fi

echo ""
echo -e "${YELLOW}📝 No existing setup found. Manual configuration needed.${NC}"

# Offer setup options
echo ""
echo -e "${YELLOW}🤔 How would you like to configure MinIO?${NC}"
echo "1) Auto-setup with Docker (creates 6 MinIO instances)"
echo "2) Auto-detect running instances"
echo "3) Manual configuration"
echo ""
read -p "Choose option (1-3): " setup_choice

case $setup_choice in
    1)
        echo -e "${BLUE}🐳 Setting up MinIO with Docker...${NC}"
        if [ -f "./setup-minio-test.sh" ]; then
            ./setup-minio-test.sh
            echo -e "${GREEN}✅ Docker setup complete! Aliases should now be configured.${NC}"
        else
            echo -e "${RED}❌ setup-minio-test.sh not found${NC}"
            exit 1
        fi
        ;;
        
    2)
        echo -e "${BLUE}🔍 Auto-detecting instances...${NC}"
        instances=($(detect_minio_instances))
        
        if [ ${#instances[@]} -eq 0 ]; then
            echo -e "${RED}❌ No MinIO instances detected${NC}"
            echo -e "${BLUE}� Make sure MinIO is running on standard ports (9001-9010)${NC}"
            exit 1
        fi
        
        echo -e "${GREEN}✅ Found ${#instances[@]} potential MinIO instances${NC}"
        
        # Configure aliases for detected instances
        for i in "${!instances[@]}"; do
            instance="${instances[$i]}"
            IFS=':' read -r _ endpoint <<< "$instance"
            alias_name="site$((i + 1))"
            configure_alias "$alias_name" "$endpoint" "$ROOT_USER" "$ROOT_PASSWORD"
        done
        ;;
        
    3)
        echo -e "${BLUE}🔧 Manual configuration...${NC}"
        echo ""
        echo "Please provide MinIO server details:"
        echo ""
        
        for i in {1..3}; do
            echo -e "${BLUE}Site $i configuration:${NC}"
            read -p "  Endpoint (e.g., http://localhost:900$i): " endpoint
            if [ -z "$endpoint" ]; then
                echo -e "${YELLOW}  Skipping site $i${NC}"
                continue
            fi
            
            read -p "  Access Key (default: $ROOT_USER): " access_key
            access_key=${access_key:-$ROOT_USER}
            
            read -p "  Secret Key (default: $ROOT_PASSWORD): " secret_key
            secret_key=${secret_key:-$ROOT_PASSWORD}
            
            configure_alias "site$i" "$endpoint" "$access_key" "$secret_key"
            echo ""
        done
        ;;
        
    *)
        echo -e "${RED}❌ Invalid choice. Exiting.${NC}"
        exit 1
        ;;
esac

# Final verification
echo ""
echo -e "${YELLOW}🔍 Final verification...${NC}"
working_count=0
for i in {1..6}; do
    alias_name="site$i"
    if mc alias list "$alias_name" >/dev/null 2>&1 && mc admin info "$alias_name" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ $alias_name is working${NC}"
        working_count=$((working_count + 1))
    fi
done

echo ""
if [ $working_count -gt 0 ]; then
    echo -e "${GREEN}🎉 Alias configuration completed!${NC}"
    echo -e "${GREEN}🎯 $working_count working aliases configured${NC}"
    echo ""
    echo -e "${YELLOW}📋 Configured aliases:${NC}"
    mc alias list 2>/dev/null | grep -E "(site[1-6])" || echo -e "${YELLOW}No site aliases found in list output${NC}"
    
    echo ""
    echo -e "${BLUE}💡 Next steps:${NC}"
    echo "1. Create test data: ./setup-test-data.sh"
    echo "2. Start web UI: ./mc-tool web --port 8080"
    echo "3. Open browser: http://localhost:8080"
    echo "4. Navigate to Operations page to test compare functionality"
else
    echo -e "${RED}❌ No working aliases configured${NC}"
    echo -e "${BLUE}💡 Please check your MinIO servers and try again${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ Ready for testing!${NC}"