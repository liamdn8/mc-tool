#!/bin/bash

# MinIO Performance Monitoring Script
# Uses awscurl to check MinIO server performance including goroutines

set -euo pipefail

# Default values
MINIO_ENDPOINT=""
ACCESS_KEY=""
SECRET_KEY=""
VERBOSE=false
OUTPUT_FORMAT="json"

usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Monitor MinIO server performance including goroutine information.

OPTIONS:
    -e, --endpoint ENDPOINT    MinIO server endpoint (e.g., https://minio.example.com)
    -a, --access-key KEY      MinIO access key
    -s, --secret-key SECRET   MinIO secret key
    -f, --format FORMAT       Output format: json, table (default: json)
    -v, --verbose             Enable verbose output
    -h, --help                Show this help message

EXAMPLES:
    # Check goroutines on local MinIO
    $0 -e http://localhost:9000 -a minioadmin -s minioadmin

    # Check performance with custom credentials
    $0 -e https://play.min.io -a Q3AM3UQ867SPQQA43P2F -s zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG

    # Table format output
    $0 -e http://localhost:9000 -a minioadmin -s minioadmin -f table

PERFORMANCE ENDPOINTS CHECKED:
    - /minio/v2/metrics/cluster    - Cluster metrics
    - /minio/prometheus/metrics    - Prometheus metrics (if enabled)
    - /debug/pprof/goroutine       - Goroutine profile
    - /debug/pprof/heap            - Heap profile
    - /minio/health/live           - Liveness check
    - /minio/health/ready          - Readiness check
EOF
}

log() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >&2
    fi
}

error() {
    echo "ERROR: $*" >&2
    exit 1
}

check_dependencies() {
    for cmd in awscurl jq; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            error "Required command '$cmd' not found. Please install it."
        fi
    done
}

make_request() {
    local path="$1"
    local description="$2"
    
    log "Checking $description at $path"
    
    local url="${MINIO_ENDPOINT}${path}"
    local response
    local http_code
    
    if ! response=$(awscurl --service s3 \
        --access_key "$ACCESS_KEY" \
        --secret_key "$SECRET_KEY" \
        --region us-east-1 \
        "$url" 2>/dev/null); then
        echo "{\"endpoint\":\"$path\",\"status\":\"error\",\"description\":\"$description\",\"error\":\"Request failed\"}"
        return 1
    fi
    
    # Try to parse as JSON, fallback to plain text
    if echo "$response" | jq . >/dev/null 2>&1; then
        echo "$response" | jq -c ". + {\"endpoint\":\"$path\",\"description\":\"$description\"}"
    else
        # For non-JSON responses (like pprof data), create JSON wrapper
        local line_count=$(echo "$response" | wc -l)
        local char_count=$(echo "$response" | wc -c)
        echo "{\"endpoint\":\"$path\",\"description\":\"$description\",\"status\":\"success\",\"response_lines\":$line_count,\"response_chars\":$char_count,\"sample\":$(echo "$response" | head -n 3 | jq -Rs .)}"
    fi
}

check_goroutines() {
    log "Checking MinIO goroutine performance"
    
    local endpoints=(
        "/minio/health/live:Liveness Check"
        "/minio/health/ready:Readiness Check"
        "/minio/v2/metrics/cluster:Cluster Metrics"
        "/debug/pprof/goroutine?debug=1:Goroutine Profile"
        "/debug/pprof/heap?debug=1:Heap Profile"
        "/minio/prometheus/metrics:Prometheus Metrics"
    )
    
    local results=()
    
    for endpoint_desc in "${endpoints[@]}"; do
        IFS=':' read -r endpoint description <<< "$endpoint_desc"
        result=$(make_request "$endpoint" "$description")
        results+=("$result")
    done
    
    if [[ "$OUTPUT_FORMAT" == "json" ]]; then
        printf '%s\n' "${results[@]}" | jq -s '{
            "timestamp": now | strftime("%Y-%m-%d %H:%M:%S"),
            "endpoint": "'"$MINIO_ENDPOINT"'",
            "checks": .
        }'
    else
        echo "MinIO Performance Check Results"
        echo "=============================="
        echo "Endpoint: $MINIO_ENDPOINT"
        echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
        echo ""
        
        printf "%-30s %-20s %-50s\n" "ENDPOINT" "STATUS" "DESCRIPTION"
        printf "%-30s %-20s %-50s\n" "--------" "------" "-----------"
        
        for result in "${results[@]}"; do
            local endpoint=$(echo "$result" | jq -r '.endpoint // "unknown"')
            local status=$(echo "$result" | jq -r '.status // "unknown"')
            local description=$(echo "$result" | jq -r '.description // "unknown"')
            
            printf "%-30s %-20s %-50s\n" "$endpoint" "$status" "$description"
        done
    fi
}

main() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--endpoint)
                MINIO_ENDPOINT="$2"
                shift 2
                ;;
            -a|--access-key)
                ACCESS_KEY="$2"
                shift 2
                ;;
            -s|--secret-key)
                SECRET_KEY="$2"
                shift 2
                ;;
            -f|--format)
                OUTPUT_FORMAT="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
    
    # Validate required parameters
    if [[ -z "$MINIO_ENDPOINT" ]]; then
        error "MinIO endpoint is required. Use -e or --endpoint."
    fi
    
    if [[ -z "$ACCESS_KEY" ]]; then
        error "Access key is required. Use -a or --access-key."
    fi
    
    if [[ -z "$SECRET_KEY" ]]; then
        error "Secret key is required. Use -s or --secret-key."
    fi
    
    if [[ ! "$OUTPUT_FORMAT" =~ ^(json|table)$ ]]; then
        error "Output format must be 'json' or 'table'"
    fi
    
    # Remove trailing slash from endpoint
    MINIO_ENDPOINT="${MINIO_ENDPOINT%/}"
    
    check_dependencies
    check_goroutines
}

main "$@"