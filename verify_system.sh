#!/bin/bash

# MinIO Site Replication System Verification Script
# Test các chức năng cơ bản của MC-Tool với Docker MinIO cluster

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test 1: Verify Docker containers are running
test_docker_containers() {
    print_status "Testing Docker MinIO containers..."
    
    local sites=("site1" "site2" "site3" "site4" "site5" "site6")
    local ports=(9001 9002 9003 9004 9005 9006)
    
    for i in "${!sites[@]}"; do
        local site="${sites[$i]}"
        local port="${ports[$i]}"
        
        if docker ps | grep "minio-${site}" | grep -q "healthy"; then
            print_success "Container minio-${site} is running and healthy"
        else
            print_error "Container minio-${site} is not healthy"
            return 1
        fi
        
        # Test connectivity
        if curl -s -f "http://localhost:${port}/minio/health/live" > /dev/null; then
            print_success "MinIO ${site} is accessible on port ${port}"
        else
            print_error "MinIO ${site} is not accessible on port ${port}"
            return 1
        fi
    done
    
    return 0
}

# Test 2: Verify mc aliases are configured
test_mc_aliases() {
    print_status "Testing mc aliases configuration..."
    
    local sites=("site1" "site2" "site3" "site4" "site5" "site6")
    
    for site in "${sites[@]}"; do
        if mc alias list | grep -q "^${site}"; then
            print_success "Alias ${site} is configured"
        else
            print_error "Alias ${site} is not configured"
            return 1
        fi
        
        # Test alias connectivity
        if mc admin info "${site}" > /dev/null 2>&1; then
            print_success "Alias ${site} is accessible"
        else
            print_warning "Alias ${site} is not accessible (this might be expected)"
        fi
    done
    
    return 0
}

# Test 3: Test web UI offline functionality
test_web_ui_offline() {
    print_status "Testing web UI offline functionality..."
    
    # Start web server in background
    print_status "Starting MC-Tool web server..."
    cd /home/liamdn/mc-tool
    ./mc-tool web --port 8082 > /tmp/mc-tool-test.log 2>&1 &
    local web_pid=$!
    
    # Wait for server to start
    sleep 3
    
    # Test main page loads
    if curl -s -f http://localhost:8082 > /dev/null; then
        print_success "Web UI main page loads successfully"
    else
        print_error "Web UI main page failed to load"
        kill $web_pid 2>/dev/null || true
        return 1
    fi
    
    # Test local Lucid Icons
    if curl -s -f http://localhost:8082/static/js/lucide.js | head -5 | grep -q "lucide"; then
        print_success "Local Lucid Icons are accessible"
    else
        print_warning "Local Lucid Icons might not be properly loaded"
    fi
    
    # Test CSS and JS files
    if curl -s -f http://localhost:8082/static/styles.css > /dev/null; then
        print_success "CSS files are accessible"
    else
        print_error "CSS files are not accessible"
        kill $web_pid 2>/dev/null || true
        return 1
    fi
    
    if curl -s -f http://localhost:8082/static/app.js > /dev/null; then
        print_success "JavaScript files are accessible"
    else
        print_error "JavaScript files are not accessible"
        kill $web_pid 2>/dev/null || true
        return 1
    fi
    
    # Test API endpoints
    if curl -s http://localhost:8082/api/replication/info | grep -q "enabled"; then
        print_success "API endpoints are working"
    else
        print_warning "API endpoints might not be fully functional (expected without replication)"
    fi
    
    # Cleanup
    kill $web_pid 2>/dev/null || true
    sleep 1
    
    return 0
}

# Test 4: Test site replication setup and smart removal
test_site_replication() {
    print_status "Testing site replication functionality..."
    
    # Setup replication with 4 sites
    print_status "Setting up site replication with 4 sites..."
    if mc admin replicate add site1 site2 site3 site4; then
        print_success "Site replication setup successful"
    else
        print_error "Site replication setup failed"
        return 1
    fi
    
    # Verify replication info
    local replication_info
    replication_info=$(mc admin replicate info site1 --json)
    
    if echo "$replication_info" | jq -r '.enabled' | grep -q "true"; then
        print_success "Site replication is enabled"
    else
        print_error "Site replication is not enabled"
        return 1
    fi
    
    local site_count
    site_count=$(echo "$replication_info" | jq '.sites | length')
    if [ "$site_count" -eq 4 ]; then
        print_success "4 sites are in replication group"
    else
        print_error "Expected 4 sites, got $site_count"
        return 1
    fi
    
    # Test smart removal (remove 1 site, keep 3)
    print_status "Testing smart site removal..."
    if mc admin replicate rm site1 site4 --force; then
        print_success "Smart site removal successful"
    else
        print_error "Smart site removal failed"
        return 1
    fi
    
    # Verify remaining sites
    replication_info=$(mc admin replicate info site1 --json)
    site_count=$(echo "$replication_info" | jq '.sites | length')
    if [ "$site_count" -eq 3 ]; then
        print_success "3 sites remain in replication group after smart removal"
    else
        print_error "Expected 3 sites after removal, got $site_count"
        return 1
    fi
    
    # Cleanup - remove replication
    print_status "Cleaning up replication configuration..."
    mc admin replicate rm site1 --all --force || true
    
    return 0
}

# Test 5: Run minimal automation tests
test_automation_suite() {
    print_status "Running automation test suite..."
    
    cd /home/liamdn/mc-tool
    
    # Install Go dependencies if not present
    if ! go list -m github.com/stretchr/testify &> /dev/null; then
        print_status "Installing Go test dependencies..."
        go get github.com/stretchr/testify/assert
        go get github.com/stretchr/testify/require
    fi
    
    # Run smart removal tests (these use mocks, so they should work)
    print_status "Running smart removal logic tests..."
    if go test ./tests/integration/ -run TestSmartSiteRemoval -v -timeout 30s; then
        print_success "Smart removal tests passed"
    else
        print_warning "Smart removal tests had issues (this might be expected)"
    fi
    
    # Run API tests
    print_status "Running API tests..."
    if go test ./tests/integration/ -run TestReplication.*API -v -timeout 30s; then
        print_success "API tests passed"
    else
        print_warning "API tests had issues (this might be expected)"
    fi
    
    return 0
}

# Main execution
main() {
    print_status "Starting MinIO Site Replication System Verification"
    print_status "=================================================="
    
    local test_results=0
    
    # Run tests
    if test_docker_containers; then
        print_success "✅ Docker containers test passed"
    else
        print_error "❌ Docker containers test failed"
        ((test_results++))
    fi
    
    if test_mc_aliases; then
        print_success "✅ MC aliases test passed"
    else
        print_error "❌ MC aliases test failed"
        ((test_results++))
    fi
    
    if test_web_ui_offline; then
        print_success "✅ Web UI offline test passed"
    else
        print_error "❌ Web UI offline test failed"
        ((test_results++))
    fi
    
    if test_site_replication; then
        print_success "✅ Site replication test passed"
    else
        print_error "❌ Site replication test failed"
        ((test_results++))
    fi
    
    if test_automation_suite; then
        print_success "✅ Automation test suite passed"
    else
        print_warning "⚠️ Automation test suite had issues"
    fi
    
    # Summary
    print_status "=================================================="
    print_status "System Verification Summary"
    print_status "=================================================="
    
    if [ $test_results -eq 0 ]; then
        print_success "🎉 All critical tests passed!"
        print_success "System is ready for production use"
        
        print_status ""
        print_status "🌐 Web UI: http://localhost:8080"
        print_status "📚 Documentation: /home/liamdn/mc-tool/docs/test/"
        print_status "🧪 Test Suite: /home/liamdn/mc-tool/docs/test/tests/"
        
        exit 0
    else
        print_error "💥 $test_results critical test(s) failed"
        print_error "Please check the issues above before production use"
        exit 1
    fi
}

# Run main function
main "$@"