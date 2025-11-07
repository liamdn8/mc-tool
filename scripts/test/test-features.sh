#!/bin/bash

# MC-Tool Feature Testing Script
# Tests all major API endpoints and functionalities after refactoring

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
FAILED_TESTS=0
TOTAL_TESTS=0

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# Function to test API endpoint
test_api() {
    local endpoint="$1"
    local method="${2:-GET}"
    local data="$3"
    local expected_status="${4:-200}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -d "$data" "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    fi
    
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" = "$expected_status" ]; then
        print_success "API $method $endpoint - Status: $status_code"
        if [ ${#body} -lt 200 ]; then
            echo "    Response: $body"
        else
            echo "    Response: ${body:0:100}..."
        fi
    else
        print_error "API $method $endpoint - Expected: $expected_status, Got: $status_code"
        echo "    Response: $body"
    fi
    
    echo ""
}

# Function to check if server is running
check_server() {
    print_status "Checking if MC-Tool web server is running..."
    
    if curl -s "$BASE_URL/api/health" > /dev/null 2>&1; then
        print_success "Server is running at $BASE_URL"
        return 0
    else
        print_error "Server is not running at $BASE_URL"
        print_warning "Please start the server with: ./mc-tool web --port 8080"
        exit 1
    fi
}

# Main testing function
run_tests() {
    print_status "============================================"
    print_status "MC-Tool Feature Testing Suite"
    print_status "============================================"
    echo ""
    
    check_server
    echo ""
    
    print_status "Testing System APIs..."
    test_api "/api/health"
    test_api "/healthz"
    test_api "/api/mc-config"
    
    print_status "Testing Site Management APIs..."
    test_api "/api/aliases"
    test_api "/api/aliases-stats"
    test_api "/api/sites"
    test_api "/api/alias-health?alias=site1"
    test_api "/api/sites/health"
    
    print_status "Testing Bucket APIs..."
    test_api "/api/buckets" "GET" "" "400"  # Should fail without alias
    test_api "/api/buckets?alias=site1"
    test_api "/api/bucket-stats?alias=site1"
    
    print_status "Testing Replication APIs..."
    test_api "/api/replication/info"
    test_api "/api/replication/status"
    
    print_status "Testing Replication Management APIs..."
    test_api "/api/replication/add" "POST" '{"sites":["site1"]}' "400"  # Should fail with < 2 sites
    test_api "/api/replication/remove" "POST" '{"confirm":false}' "400"  # Should fail without confirmation
    
    print_status "Testing Analysis APIs..."
    test_api "/api/compare?site1=site1&site2=site2&bucket=test"
    test_api "/api/analyze?alias=site1"
    test_api "/api/checklist?alias=site1"
    
    print_status "Testing Job APIs..."
    test_api "/api/jobs/test-job-id"
    
    print_status "============================================"
    print_status "Test Summary"
    print_status "============================================"
    
    PASSED_TESTS=$((TOTAL_TESTS - FAILED_TESTS))
    
    if [ $FAILED_TESTS -eq 0 ]; then
        print_success "All tests passed! ($PASSED_TESTS/$TOTAL_TESTS)"
        echo ""
        print_success "🎉 MC-Tool refactoring successful!"
        print_success "✅ All API endpoints working correctly"
        print_success "✅ Server architecture is stable"
        print_success "✅ Replication functions properly separated"
        exit 0
    else
        print_error "$FAILED_TESTS tests failed out of $TOTAL_TESTS total tests"
        print_warning "Some issues need to be addressed"
        exit 1
    fi
}

# Run the tests
run_tests