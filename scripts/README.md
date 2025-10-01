# MinIO Performance Monitoring

This directory contains scripts for monitoring MinIO server performance, particularly focusing on goroutine analysis and server health.

## Tools

### 1. minio-perf.sh
Comprehensive performance monitoring script that checks various MinIO endpoints.

**Features:**
- Health checks (liveness/readiness)
- Goroutine profiling
- Memory heap analysis
- Cluster metrics
- Prometheus metrics (if enabled)

**Usage:**
```bash
# Basic usage
./minio-perf.sh -e http://localhost:9000 -a minioadmin -s minioadmin

# Table format output
./minio-perf.sh -e http://localhost:9000 -a minioadmin -s minioadmin -f table

# JSON output with verbose logging
./minio-perf.sh -e http://localhost:9000 -a minioadmin -s minioadmin -f json -v
```

### 2. minio-perf-wrapper.sh (Available as `minio-perf` command)
Simplified wrapper for the performance monitoring script.

**Usage:**
```bash
# With arguments
minio-perf http://localhost:9000 minioadmin minioadmin

# Interactive mode
minio-perf
```

## Docker Usage

When using the Docker container, these tools are pre-installed:

```bash
# Run the container
docker run -it mc-tool

# Use the performance monitor
minio-perf http://your-minio:9000 your-access-key your-secret-key

# Or use the full script with options
/usr/local/bin/scripts/minio-perf.sh --help
```

## Monitored Endpoints

| Endpoint | Description | Purpose |
|----------|-------------|---------|
| `/minio/health/live` | Liveness check | Verify server is running |
| `/minio/health/ready` | Readiness check | Verify server is ready for requests |
| `/minio/v2/metrics/cluster` | Cluster metrics | Overall cluster health and stats |
| `/debug/pprof/goroutine` | Goroutine profile | Monitor goroutine usage and potential leaks |
| `/debug/pprof/heap` | Heap profile | Memory usage analysis |
| `/minio/prometheus/metrics` | Prometheus metrics | Detailed performance metrics |

## Output Formats

### JSON Format
Structured output suitable for parsing and automation:
```json
{
  "timestamp": "2025-09-30 17:45:00",
  "endpoint": "http://localhost:9000",
  "checks": [
    {
      "endpoint": "/minio/health/live",
      "description": "Liveness Check",
      "status": "success"
    }
  ]
}
```

### Table Format
Human-readable tabular output:
```
MinIO Performance Check Results
===============================
Endpoint: http://localhost:9000
Timestamp: 2025-09-30 17:45:00

ENDPOINT                       STATUS               DESCRIPTION
--------                       ------               -----------
/minio/health/live            success              Liveness Check
/minio/health/ready           success              Readiness Check
```

## Dependencies

The scripts require:
- `awscurl` - For authenticated HTTP requests to MinIO
- `jq` - For JSON processing
- `bash` - For script execution

All dependencies are pre-installed in the Docker container.

## Troubleshooting

### awscurl not found
```bash
# Install via pip
pip3 install awscurl
```

### Authentication errors
- Verify access key and secret key
- Ensure the MinIO server allows the API endpoints
- Check if the server has authentication enabled

### Connection errors
- Verify the endpoint URL is correct
- Check network connectivity
- Ensure MinIO server is running and accessible