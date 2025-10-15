#!/bin/bash

# MC-Tool Complete Test Environment Setup
#!/bin/bash

# MC-Tool Complete Test Environment Setup
# This script sets up everything needed for testing the Operations page

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logo
echo -e "${CYAN}"
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║               MC-TOOL COMPLETE TEST ENVIRONMENT              ║"
echo "║                 Operations Page Testing Setup                ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Function to print step headers
print_step() {
    echo ""
    echo -e "${PURPLE}▶ $1${NC}"
    echo -e "${BLUE}────────────────────────────────────────────────────────────────${NC}"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if a service is running
service_running() {
    if pgrep -x "$1" > /dev/null; then
        return 0
    fi
    return 1
}

# Function to wait for service
wait_for_service() {
    local host=$1
    local port=$2
    local max_attempts=${3:-30}
    local attempt=1
    
    echo -e "${BLUE}⏳ Waiting for service at $host:$port...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if timeout 2 bash -c "</dev/tcp/$host/$port" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Service is ready at $host:$port${NC}"
            return 0
        fi
        
        echo -e "${YELLOW}⏳ Attempt $attempt/$max_attempts - waiting for $host:$port...${NC}"
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}❌ Service at $host:$port failed to start within $((max_attempts * 2)) seconds${NC}"
    return 1
}

# Step 1: Verify prerequisites
print_step "STEP 1: Verifying Prerequisites"

echo -e "${BLUE}🔍 Checking required commands...${NC}"

MISSING_DEPS=()

if ! command_exists mc; then
    MISSING_DEPS+=("mc")
    echo -e "${RED}❌ MinIO client 'mc' not found${NC}"
else
    echo -e "${GREEN}✅ MinIO client 'mc' available${NC}"
fi

if ! command_exists docker; then
    MISSING_DEPS+=("docker")
    echo -e "${RED}❌ Docker not found${NC}"
else
    echo -e "${GREEN}✅ Docker available${NC}"
fi

if ! command_exists go; then
    MISSING_DEPS+=("go")
    echo -e "${RED}❌ Go not found${NC}"
else
    echo -e "${GREEN}✅ Go available ($(go version | awk '{print $3}'))${NC}"
fi

if ! command_exists npm; then
    MISSING_DEPS+=("npm")
    echo -e "${RED}❌ npm not found${NC}"
else
    echo -e "${GREEN}✅ npm available ($(npm --version))${NC}"
fi

if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
    echo -e "${RED}❌ Missing dependencies: ${MISSING_DEPS[*]}${NC}"
    echo -e "${BLUE}💡 Install missing dependencies and try again${NC}"
    exit 1
fi

# Step 2: Build the application
print_step "STEP 2: Building MC-Tool Application"

if [ ! -f "main.go" ]; then
    echo -e "${RED}❌ main.go not found. Are you in the correct directory?${NC}"
    exit 1
fi

echo -e "${BLUE}🔨 Building web assets...${NC}"
if [ -f "build-web.sh" ]; then
    if ./build-web.sh; then
        echo -e "${GREEN}✅ Web assets built successfully${NC}"
    else
        echo -e "${RED}❌ Web build failed${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  build-web.sh not found, building manually...${NC}"
    cd web
    npm install
    npm run build
    cd ..
fi

echo -e "${BLUE}🔨 Building Go application...${NC}"
if go build -o mc-tool-test ./main.go; then
    echo -e "${GREEN}✅ Go application built successfully${NC}"
else
    echo -e "${RED}❌ Go build failed${NC}"
    exit 1
fi

# Step 3: Setup MinIO instances
print_step "STEP 3: Setting up MinIO Test Environment"

echo -e "${BLUE}🐳 Setting up MinIO instances with Docker...${NC}"

if [ -f "setup-minio-test.sh" ]; then
    if ./setup-minio-test.sh; then
        echo -e "${GREEN}✅ MinIO instances setup completed${NC}"
    else
        echo -e "${RED}❌ MinIO setup failed${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  setup-minio-test.sh not found, setting up manually...${NC}"
    
    # Manual MinIO setup
    HOST_IP=$(hostname -I | awk '{print $1}')
    
    for i in {1..6}; do
        container_name="minio-site$i"
        port=$((9000 + i))
        console_port=$((9900 + i))
        data_dir="/tmp/minio-site$i-data"
        
        echo -e "${BLUE}🐳 Starting $container_name on port $port...${NC}"
        
        # Stop existing container
        docker stop "$container_name" >/dev/null 2>&1 || true
        docker rm "$container_name" >/dev/null 2>&1 || true
        
        # Create data directory
        mkdir -p "$data_dir"
        
        # Start container
        docker run -d \
            --name "$container_name" \
            -p "$port:9000" \
            -p "$console_port:9001" \
            -e MINIO_ROOT_USER=minioadmin \
            -e MINIO_ROOT_PASSWORD=minioadmin123 \
            -v "$data_dir:/data" \
            minio/minio server /data --console-address ":9001"
        
        # Wait for service
        wait_for_service "$HOST_IP" "$port" 15
    done
fi

# Step 4: Configure aliases
print_step "STEP 4: Configuring MinIO Client Aliases"

if [ -f "setup-aliases.sh" ]; then
    echo -e "${BLUE}🔧 Running aliases setup script...${NC}"
    if ./setup-aliases.sh; then
        echo -e "${GREEN}✅ Aliases configured successfully${NC}"
    else
        echo -e "${YELLOW}⚠️  Aliases setup had issues, continuing...${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  setup-aliases.sh not found, configuring manually...${NC}"
    
    # Manual alias configuration
    HOST_IP=$(hostname -I | awk '{print $1}')
    
    for i in {1..6}; do
        alias_name="site$i"
        port=$((9000 + i))
        endpoint="http://$HOST_IP:$port"
        
        echo -e "${BLUE}🔧 Configuring alias '$alias_name'...${NC}"
        
        # Remove existing alias
        mc alias remove "$alias_name" >/dev/null 2>&1 || true
        
        # Add new alias
        if mc alias set "$alias_name" "$endpoint" "minioadmin" "minioadmin123" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Configured $alias_name -> $endpoint${NC}"
        else
            echo -e "${RED}❌ Failed to configure $alias_name${NC}"
        fi
    done
fi

# Step 5: Create test data
print_step "STEP 5: Creating Test Data"

if [ -f "setup-test-data.sh" ]; then
    echo -e "${BLUE}📊 Running test data creation script...${NC}"
    if ./setup-test-data.sh; then
        echo -e "${GREEN}✅ Test data created successfully${NC}"
    else
        echo -e "${YELLOW}⚠️  Test data creation had issues, continuing...${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  setup-test-data.sh not found, creating basic test data...${NC}"
    
    # Create basic test buckets and data
    for alias in site1 site2 site3; do
        if mc admin info "$alias" >/dev/null 2>&1; then
            echo -e "${BLUE}📦 Creating test buckets for $alias...${NC}"
            
            # Create buckets
            mc mb "$alias/test-bucket" >/dev/null 2>&1 || true
            mc mb "$alias/compare-test" >/dev/null 2>&1 || true
            
            # Add some test files
            echo "Test content for $alias" | mc pipe "$alias/test-bucket/test-file.txt"
            echo "Compare test data" | mc pipe "$alias/compare-test/sample.txt"
            
            echo -e "${GREEN}✅ Created basic test data for $alias${NC}"
        fi
    done
fi

# Step 6: Start the application
print_step "STEP 6: Starting MC-Tool Web Application"

echo -e "${BLUE}🚀 Starting MC-Tool web server...${NC}"

# Kill existing instances
pkill -f "mc-tool.*web" >/dev/null 2>&1 || true
sleep 2

# Start in background
nohup ./mc-tool-test web --port 8080 > mc-tool.log 2>&1 &
WEB_PID=$!

echo -e "${BLUE}⏳ Waiting for web server to start...${NC}"
if wait_for_service "localhost" "8080" 15; then
    echo -e "${GREEN}✅ MC-Tool web server is running on http://localhost:8080${NC}"
    echo -e "${BLUE}📋 Process ID: $WEB_PID${NC}"
    echo -e "${BLUE}📄 Logs: tail -f mc-tool.log${NC}"
else
    echo -e "${RED}❌ Failed to start web server${NC}"
    echo -e "${BLUE}💡 Check logs: cat mc-tool.log${NC}"
    exit 1
fi

# Step 7: Final verification and instructions
print_step "STEP 7: Environment Verification & Instructions"

echo -e "${BLUE}🔍 Verifying complete setup...${NC}"

# Check aliases
working_aliases=()
for alias in site1 site2 site3 site4 site5 site6; do
    if mc admin info "$alias" >/dev/null 2>&1; then
        working_aliases+=("$alias")
    fi
done

echo -e "${GREEN}✅ Working MinIO aliases: ${#working_aliases[@]} (${working_aliases[*]})${NC}"

# Check web server
if curl -s "http://localhost:8080" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Web server is responding${NC}"
else
    echo -e "${YELLOW}⚠️  Web server check failed${NC}"
fi

# Check test buckets
test_buckets=0
for alias in "${working_aliases[@]}"; do
    bucket_count=$(mc ls "$alias" 2>/dev/null | wc -l || echo "0")
    test_buckets=$((test_buckets + bucket_count))
done
echo -e "${GREEN}✅ Total test buckets available: $test_buckets${NC}"

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                     🎉 SETUP COMPLETE! 🎉                       ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${GREEN}🌐 Web Interface:${NC} http://localhost:8080"
echo -e "${GREEN}📊 Operations Page:${NC} http://localhost:8080/#/operations"
echo ""

echo -e "${YELLOW}🔧 Available MinIO Aliases:${NC}"
for alias in "${working_aliases[@]}"; do
    endpoint=$(mc alias list "$alias" 2>/dev/null | awk '{print $2}' | head -1)
    echo -e "  • ${GREEN}$alias${NC} -> $endpoint"
done

echo ""
echo -e "${BLUE}📋 What you can test now:${NC}"
echo "  1. 🔍 Compare Operations:"
echo "     - Navigate to Operations page"
echo "     - Select two different aliases"
echo "     - Compare buckets/paths"
echo "     - View differences in organized blocks"
echo ""
echo "  2. ✅ Configuration Checklist:"
echo "     - Check environment variables"
echo "     - Verify event configurations"
echo "     - Review bucket lifecycle settings"
echo ""
echo "  3. 🛠️ General Operations:"
echo "     - Works with or without site replication"
echo "     - Dynamic bucket/path suggestions"
echo "     - Real-time connectivity checking"

echo ""
echo -e "${PURPLE}🔧 Useful Commands:${NC}"
echo "  • View logs: tail -f mc-tool.log"
echo "  • Stop web server: kill $WEB_PID"
echo "  • Stop MinIO containers: docker stop \$(docker ps -q --filter name=minio-site)"
echo "  • Restart setup: $0"

echo ""
echo -e "${CYAN}🚀 Ready for testing! Open your browser and start exploring!${NC}"

set -e

echo "🚀 MC-Tool Test Environment Setup"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check prerequisites
echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"

if ! command -v mc &> /dev/null; then
    echo -e "${RED}❌ MinIO client 'mc' is not installed${NC}"
    echo -e "${BLUE}💡 Install with: wget https://dl.min.io/client/mc/release/linux-amd64/mc && chmod +x mc && sudo mv mc /usr/local/bin/${NC}"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}⚠️  Docker not found. You'll need to run MinIO manually or install Docker.${NC}"
    DOCKER_AVAILABLE=false
else
    echo -e "${GREEN}✅ Docker is available${NC}"
    DOCKER_AVAILABLE=true
fi

echo -e "${GREEN}✅ MinIO client 'mc' is available${NC}"

# Function to start MinIO with Docker
start_minio_docker() {
    local name=$1
    local port=$2
    local console_port=$3
    
    echo -e "${BLUE}🐳 Starting MinIO container '$name' on port $port...${NC}"
    
    # Check if container already exists
    if docker ps -a --format "table {{.Names}}" | grep -q "^$name$"; then
        echo -e "${YELLOW}⚠️  Container '$name' already exists. Removing...${NC}"
        docker rm -f "$name" >/dev/null 2>&1
    fi
    
    # Create data directory
    mkdir -p "./test-data/$name"
    
    # Start MinIO container
    docker run -d \
        --name "$name" \
        -p "$port:9000" \
        -p "$console_port:9001" \
        -e "MINIO_ROOT_USER=admin" \
        -e "MINIO_ROOT_PASSWORD=password123" \
        -v "$(pwd)/test-data/$name:/data" \
        quay.io/minio/minio server /data --console-address ":9001" >/dev/null
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Started MinIO container '$name'${NC}"
        echo -e "${BLUE}   API: http://localhost:$port${NC}"
        echo -e "${BLUE}   Console: http://localhost:$console_port${NC}"
    else
        echo -e "${RED}❌ Failed to start MinIO container '$name'${NC}"
        return 1
    fi
}

# Function to wait for MinIO to be ready
wait_for_minio() {
    local alias=$1
    local max_attempts=30
    local attempt=1
    
    echo -e "${YELLOW}⏳ Waiting for MinIO '$alias' to be ready...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if mc ping "$alias" --count 1 >/dev/null 2>&1; then
            echo -e "${GREEN}✅ MinIO '$alias' is ready!${NC}"
            return 0
        fi
        
        echo -e "${YELLOW}   Attempt $attempt/$max_attempts...${NC}"
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}❌ MinIO '$alias' failed to start after $max_attempts attempts${NC}"
    return 1
}

# Ask user for setup type
echo ""
echo -e "${YELLOW}🤔 How would you like to set up MinIO instances?${NC}"
echo "1) Docker containers (automatic)"
echo "2) Manual configuration (you provide endpoints)"
echo "3) Skip MinIO setup (already configured)"
echo ""
read -p "Choose option (1-3): " setup_choice

case $setup_choice in
    1)
        if [ "$DOCKER_AVAILABLE" = false ]; then
            echo -e "${RED}❌ Docker is not available. Please choose option 2 or 3.${NC}"
            exit 1
        fi
        
        echo -e "${BLUE}🐳 Setting up MinIO with Docker containers...${NC}"
        
        # Start MinIO containers
        start_minio_docker "minio-site4" "9004" "9014"
        start_minio_docker "minio-site5" "9005" "9015" 
        start_minio_docker "minio-site6" "9006" "9016"
        
        echo -e "${YELLOW}⏳ Waiting for containers to start...${NC}"
        sleep 5
        
        # Configure aliases
        echo -e "${BLUE}🔧 Configuring MinIO aliases...${NC}"
        mc config host add site4 http://localhost:9004 admin password123
        mc config host add site5 http://localhost:9005 admin password123
        mc config host add site6 http://localhost:9006 admin password123
        
        # Wait for MinIO instances to be ready
        wait_for_minio site4
        wait_for_minio site5  
        wait_for_minio site6
        ;;
        
    2)
        echo -e "${BLUE}🔧 Manual MinIO configuration...${NC}"
        echo "Please configure your MinIO aliases manually:"
        echo ""
        echo "Examples:"
        echo "  mc config host add site4 http://localhost:9004 admin password123"
        echo "  mc config host add site5 http://localhost:9005 admin password123"  
        echo "  mc config host add site6 http://localhost:9006 admin password123"
        echo ""
        echo "Or run: ./setup-aliases.sh"
        echo ""
        read -p "Press Enter when aliases are configured..."
        ;;
        
    3)
        echo -e "${BLUE}⏭️  Skipping MinIO setup...${NC}"
        ;;
        
    *)
        echo -e "${RED}❌ Invalid choice. Exiting.${NC}"
        exit 1
        ;;
esac

# Verify aliases are configured
echo ""
echo -e "${YELLOW}🔍 Verifying MinIO aliases...${NC}"

REQUIRED_ALIASES=("site4" "site5" "site6")
for alias in "${REQUIRED_ALIASES[@]}"; do
    if mc ping "$alias" --count 1 >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Alias '$alias' is working${NC}"
    else
        echo -e "${RED}❌ Alias '$alias' is not working${NC}"
        echo -e "${YELLOW}💡 Please check your MinIO configuration for '$alias'${NC}"
        exit 1
    fi
done

# Create test data
echo ""
echo -e "${YELLOW}📊 Creating test data...${NC}"

if [ -f "./setup-test-data.sh" ]; then
    ./setup-test-data.sh
else
    echo -e "${RED}❌ setup-test-data.sh not found${NC}"
    exit 1
fi

# Start web UI
echo ""
echo -e "${YELLOW}🌐 Starting MC-Tool Web UI...${NC}"

if [ -f "./mc-tool" ]; then
    echo -e "${GREEN}🚀 Starting web server on port 8080...${NC}"
    echo -e "${BLUE}🌍 Open your browser at: http://localhost:8080${NC}"
    echo -e "${BLUE}📋 Navigate to Operations page to test compare functionality${NC}"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
    echo ""
    
    # Start the web server
    ./mc-tool web --port 8080
else
    echo -e "${RED}❌ mc-tool binary not found. Please build it first with: go build${NC}"
    exit 1
fi

# Cleanup function
cleanup() {
    echo ""
    echo -e "${YELLOW}🧹 Cleaning up...${NC}"
    
    if [ "$setup_choice" = "1" ] && [ "$DOCKER_AVAILABLE" = true ]; then
        echo -e "${BLUE}🐳 Stopping Docker containers...${NC}"
        docker stop minio-site4 minio-site5 minio-site6 >/dev/null 2>&1 || true
        docker rm minio-site4 minio-site5 minio-site6 >/dev/null 2>&1 || true
        echo -e "${GREEN}✅ Containers stopped and removed${NC}"
    fi
    
    echo -e "${GREEN}👋 Goodbye!${NC}"
}

# Set up cleanup on script exit
trap cleanup EXIT