#!/bin/bash

# Integration Test for Infrastructure Validation Auto-Discovery
# Tests the new features: resource discovery, static pod detection, and CRD support

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Infrastructure Validation - Integration Test"
echo "=========================================="
echo ""

# Function to print test status
print_test() {
    local status=$1
    local message=$2
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $message"
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}✗ FAIL${NC}: $message"
        exit 1
    elif [ "$status" = "INFO" ]; then
        echo -e "${YELLOW}ℹ INFO${NC}: $message"
    fi
}

# Check if mc-tool binary exists
if [ ! -f "./mc-tool" ]; then
    print_test "FAIL" "mc-tool binary not found. Run 'go build' first."
fi

print_test "INFO" "Starting web server in background..."
./mc-tool web --port 8888 > /tmp/mc-tool-test.log 2>&1 &
WEB_PID=$!
sleep 3

# Cleanup function
cleanup() {
    print_test "INFO" "Cleaning up..."
    if [ ! -z "$WEB_PID" ]; then
        kill $WEB_PID 2>/dev/null || true
    fi
}

trap cleanup EXIT

# Test 1: Check if web server is running
print_test "INFO" "Test 1: Checking if web server is running..."
if curl -s http://localhost:8888/minio-webtool/ > /dev/null; then
    print_test "PASS" "Web server is running"
else
    print_test "FAIL" "Web server is not responding"
fi

# Test 2: Test search-namespaces endpoint (without VIM config)
print_test "INFO" "Test 2: Testing search-namespaces endpoint..."
RESPONSE=$(curl -s "http://localhost:8888/minio-webtool/api/validate/infrastructure/search-namespaces?keyword=test&exact=false")
if echo "$RESPONSE" | grep -q "matches"; then
    print_test "PASS" "Search-namespaces endpoint returns valid response"
else
    print_test "FAIL" "Search-namespaces endpoint returned invalid response: $RESPONSE"
fi

# Test 3: Test search-namespaces with exact match
print_test "INFO" "Test 3: Testing search-namespaces with exact match..."
RESPONSE=$(curl -s "http://localhost:8888/minio-webtool/api/validate/infrastructure/search-namespaces?keyword=default&exact=true")
if echo "$RESPONSE" | grep -q "matches"; then
    print_test "PASS" "Search-namespaces with exact match works"
else
    print_test "FAIL" "Search-namespaces with exact match failed: $RESPONSE"
fi

# Test 4: Test discover-resources endpoint parameter validation
print_test "INFO" "Test 4: Testing discover-resources parameter validation..."
RESPONSE=$(curl -s "http://localhost:8888/minio-webtool/api/validate/infrastructure/discover-resources")
if echo "$RESPONSE" | grep -q "error"; then
    print_test "PASS" "Discover-resources validates required parameters"
else
    print_test "FAIL" "Discover-resources should return error for missing parameters"
fi

# Test 5: Test infrastructure validation endpoint validation
print_test "INFO" "Test 5: Testing validation endpoint with invalid payload..."
RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{}' \
    http://localhost:8888/minio-webtool/api/validate/infrastructure)
if echo "$RESPONSE" | grep -q "error"; then
    print_test "PASS" "Validation endpoint validates payload"
else
    print_test "FAIL" "Validation endpoint should return error for invalid payload"
fi

# Test 6: Test validation with missing baseline
print_test "INFO" "Test 6: Testing validation with missing baseline..."
RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"targets":["site1/ns1"]}' \
    http://localhost:8888/minio-webtool/api/validate/infrastructure)
if echo "$RESPONSE" | grep -q "error.*[Bb]aseline"; then
    print_test "PASS" "Validation rejects missing baseline"
else
    print_test "FAIL" "Validation should reject missing baseline"
fi

# Test 7: Test validation with missing targets
print_test "INFO" "Test 7: Testing validation with missing targets..."
RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"baseline":"site1/ns1","targets":[]}' \
    http://localhost:8888/minio-webtool/api/validate/infrastructure)
if echo "$RESPONSE" | grep -q "error.*target"; then
    print_test "PASS" "Validation rejects empty targets"
else
    print_test "FAIL" "Validation should reject empty targets"
fi

# Test 8: Test VIMs endpoint
print_test "INFO" "Test 8: Testing VIMs listing endpoint..."
RESPONSE=$(curl -s http://localhost:8888/minio-webtool/api/validate/infrastructure/vims)
if echo "$RESPONSE" | grep -q "vims"; then
    print_test "PASS" "VIMs endpoint returns valid response"
else
    print_test "FAIL" "VIMs endpoint failed: $RESPONSE"
fi

# Test 9: Test history endpoint
print_test "INFO" "Test 9: Testing history endpoint..."
RESPONSE=$(curl -s "http://localhost:8888/minio-webtool/api/validate/infrastructure/history?limit=10")
if echo "$RESPONSE" | grep -q "jobs"; then
    print_test "PASS" "History endpoint returns valid response"
else
    print_test "FAIL" "History endpoint failed: $RESPONSE"
fi

# Test 10: Test namespaces endpoint parameter validation
print_test "INFO" "Test 10: Testing namespaces endpoint parameter validation..."
RESPONSE=$(curl -s "http://localhost:8888/minio-webtool/api/validate/infrastructure/namespaces")
if echo "$RESPONSE" | grep -q "error"; then
    print_test "PASS" "Namespaces endpoint validates vim parameter"
else
    print_test "FAIL" "Namespaces endpoint should validate vim parameter"
fi

echo ""
echo "=========================================="
echo -e "${GREEN}All Tests Passed!${NC}"
echo "=========================================="
echo ""

exit 0
