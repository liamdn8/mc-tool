#!/bin/bash

# MinIO Performance Monitoring Wrapper
# Simplified interface for checking MinIO goroutine performance

set -euo pipefail

usage() {
    cat << EOF
MinIO Performance Monitor

USAGE:
    minio-perf [ENDPOINT] [ACCESS_KEY] [SECRET_KEY]

EXAMPLES:
    # Check local MinIO
    minio-perf http://localhost:9000 minioadmin minioadmin

    # Check MinIO playground
    minio-perf https://play.min.io Q3AM3UQ867SPQQA43P2F zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG

    # Interactive mode (no arguments)
    minio-perf

WHAT IT CHECKS:
    ✓ Server health (liveness/readiness)
    ✓ Goroutine profiles
    ✓ Memory heap usage
    ✓ Cluster metrics
    ✓ Prometheus metrics (if available)
EOF
}

if [[ $# -eq 0 ]]; then
    echo "MinIO Performance Monitor - Interactive Mode"
    echo "==========================================="
    echo
    
    read -p "MinIO Endpoint (e.g., http://localhost:9000): " endpoint
    read -p "Access Key: " access_key
    read -s -p "Secret Key: " secret_key
    echo
    echo
    
    exec /usr/local/bin/scripts/minio-perf.sh -e "$endpoint" -a "$access_key" -s "$secret_key" -f table
elif [[ $# -eq 3 ]]; then
    exec /usr/local/bin/scripts/minio-perf.sh -e "$1" -a "$2" -s "$3" -f table
else
    usage
    exit 1
fi