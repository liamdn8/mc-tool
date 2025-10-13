# MinIO Debug Command

The `mc-tool debug` command provides advanced debugging capabilities for MinIO servers, specifically designed to detect goroutine leaks and performance issues in production environments.

## Overview

This command analyzes MinIO server goroutines by connecting to the pprof endpoints and provides:

- **Real-time goroutine analysis**
- **Leak detection with configurable thresholds**
- **Continuous monitoring with alerts**
- **Detailed stack trace analysis**
- **JSON and text output formats**

## Prerequisites

### MC Configuration

Before using the debug command, ensure your MinIO server is configured with the `mc` client:

```bash
# Add your MinIO server as an alias
mc alias set minio-prod http://your-minio-server:9000 your-access-key your-secret-key

# Verify the configuration
mc alias list minio-prod
```

The debug command uses the same configuration as other mc-tool commands, loading credentials and endpoints from `~/.mc/config.json`.

## Usage

### Basic Syntax

```bash
mc-tool debug <alias> [flags]
```

### Examples

#### Single Snapshot Analysis
```bash
# Analyze current goroutine state
mc-tool debug minio-prod

# With verbose output
mc-tool debug minio-prod --verbose
```

#### Continuous Monitoring
```bash
# Monitor for 10 minutes with default settings
mc-tool debug minio-prod --monitor 10m

# Custom monitoring with 30-second intervals
mc-tool debug minio-prod --monitor 1h --interval 30s --threshold 100
```

#### JSON Output for Automation
```bash
# JSON format for parsing/automation
mc-tool debug minio-prod --format json
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--monitor` | duration | - | Monitor duration (e.g., 10m, 1h) |
| `--interval` | duration | 10s | Monitoring interval |
| `--threshold` | int | 50 | Goroutine growth threshold for leak detection |
| `--format` | string | text | Output format: text, json |
| `--verbose, -v` | bool | false | Verbose output |
| `--insecure` | bool | false | Skip TLS certificate verification |

## Configuration

### Using MC Aliases

The debug command integrates with the MinIO client configuration:

```bash
# List available aliases
mc alias list

# Add a new alias for debugging
mc alias set debug-server http://localhost:9000 minioadmin minioadmin

# Use the alias with debug command
mc-tool debug debug-server
```

### Multiple Environments

```bash
# Production environment
mc alias set prod https://minio-prod.company.com prod-access-key prod-secret-key
mc-tool debug prod --monitor 30m

# Staging environment  
mc alias set staging http://minio-staging:9000 staging-key staging-secret
mc-tool debug staging --format json

# Local development
mc alias set local http://localhost:9000 minioadmin minioadmin
mc-tool debug local --verbose
```

## Output

### Text Format

```
🔍 MinIO Goroutine Analysis
==========================
📡 Endpoint: http://localhost:9000
🕐 Timestamp: 2025-09-30 17:45:00
📊 Total Goroutines: 127

📈 Goroutines by State:
   running             : 2
   syscall             : 45
   chan receive        : 23
   select              : 57

🏆 Top Functions:
   1. runtime.gopark              : 45
   2. net/http.(*conn).serve      : 23
   3. crypto/tls.(*Conn).Read     : 12
   4. time.Sleep                  : 8
   5. os/signal.signal_recv       : 5

⏱️  Long-running Goroutines (3):
   ID:45 [syscall] 2 minutes - runtime.gopark
   ID:67 [select] 5 minutes - time.Sleep
   ID:89 [chan receive] 10 minutes - net/http.(*conn).serve

🚨 Potential Leaks (2):
   ID:45 [syscall] runtime.gopark
   ID:123 [select] time.Sleep
```

### JSON Format

```json
{
  "total": 127,
  "by_state": {
    "running": 2,
    "syscall": 45,
    "chan receive": 23,
    "select": 57
  },
  "top_functions": [
    {
      "function": "runtime.gopark",
      "count": 45
    }
  ],
  "long_running": [
    {
      "id": 45,
      "state": "syscall",
      "function": "runtime.gopark",
      "duration": "2 minutes"
    }
  ],
  "potential_leaks": [
    {
      "id": 45,
      "state": "syscall",
      "function": "runtime.gopark"
    }
  ],
  "timestamp": "2025-09-30T17:45:00Z",
  "endpoint": "http://localhost:9000"
}
```

## Monitoring Mode

When using `--monitor`, the tool will:

1. **Establish baseline** - Record initial goroutine count
2. **Continuous monitoring** - Check at specified intervals
3. **Leak detection** - Alert when growth exceeds threshold
4. **Summary report** - Generate final analysis

### Monitoring Output

```
🔍 Monitoring MinIO goroutines for 10m0s (interval: 10s)
📡 Endpoint: http://minio-prod.company.com:9000
🎯 Leak threshold: 50 goroutines

📊 Baseline: 127 goroutines
✅ Current: 129 goroutines (growth: +2)
✅ Current: 131 goroutines (growth: +4)
🚨 POTENTIAL LEAK DETECTED!
   Growth: +78 goroutines (baseline: 127, current: 205)

📊 Growth Analysis:
   syscall: +45 (was 45, now 90)
   select: +23 (was 57, now 80)
   chan receive: +10 (was 23, now 33)

📈 Summary Report
================
Duration: 10m0s
Measurements: 60
Baseline: 127 goroutines
Final: 205 goroutines
Net Growth: +78 goroutines
Peak: 218 goroutines
```

## Leak Detection

The debug command uses several heuristics to detect potential goroutine leaks:

### Detection Criteria

1. **Growth threshold** - Significant increase from baseline
2. **Long-running goroutines** - Goroutines running for extended periods
3. **Suspicious patterns** - Common leak patterns like:
   - `runtime.gopark`
   - `sync.runtime_Semacquire`
   - `time.Sleep` (excessive)
   - `chan receive/send` (blocked)

### Common Leak Patterns

| Pattern | Description | Likely Cause |
|---------|-------------|--------------|
| `runtime.gopark` | Goroutine parked waiting | Resource contention |
| `sync.runtime_Semacquire` | Waiting for semaphore | Deadlock or contention |
| `time.Sleep` | Sleeping goroutine | Inefficient polling |
| `chan receive` | Blocked on channel | Producer/consumer imbalance |
| `syscall` | System call blocked | Network or I/O issues |

## Prerequisites

### MinIO Server Configuration

The debug command requires access to MinIO's pprof endpoints:

```bash
# Ensure pprof is enabled (usually enabled by default)
export MINIO_PROMETHEUS_AUTH_TYPE="public"
```

### Required Endpoints

The command accesses:
- `/debug/pprof/goroutine?debug=1` - Goroutine profiles
- Authentication via AWS Signature v4

### Network Access

Ensure the MinIO server is accessible and pprof endpoints are not blocked by firewalls.

## Troubleshooting

### Common Issues

1. **Authentication failed**
   - Verify access key and secret key
   - Check MinIO server authentication settings

2. **Connection refused**
   - Verify endpoint URL and port
   - Check network connectivity
   - Ensure MinIO server is running

3. **403 Forbidden**
   - pprof endpoints may be disabled
   - Check MinIO server configuration

4. **Empty goroutine data**
   - Server may be under heavy load
   - Try increasing timeout values

### Debug Tips

```bash
# Test basic connectivity first
curl http://localhost:9000/minio/health/live

# Verify pprof endpoint accessibility
curl http://localhost:9000/debug/pprof/

# Use verbose mode for debugging
mc-tool debug http://localhost:9000 \
  --access-key minioadmin \
  --secret-key minioadmin \
  --verbose
```

## Production Usage

### Best Practices

1. **Regular monitoring** - Schedule periodic checks
2. **Baseline establishment** - Know normal goroutine counts
3. **Alert thresholds** - Set appropriate growth limits
4. **Automation** - Use JSON output for automated analysis

### Example Monitoring Script

```bash
#!/bin/bash
# Production monitoring script

ALIAS="minio-prod"
THRESHOLD=100

# Run debug analysis
mc-tool debug "$ALIAS" \
  --format json > /tmp/minio-debug.json

# Check for leaks
GOROUTINES=$(jq '.total' /tmp/minio-debug.json)
LEAKS=$(jq '.potential_leaks | length' /tmp/minio-debug.json)

if [ "$LEAKS" -gt 0 ]; then
  echo "ALERT: $LEAKS potential goroutine leaks detected"
  echo "Total goroutines: $GOROUTINES"
  # Send alert to monitoring system
fi
```

## Integration

### With Monitoring Systems

The JSON output can be easily integrated with:
- **Prometheus** - Custom metrics
- **Grafana** - Dashboards and alerts
- **ELK Stack** - Log analysis
- **Custom monitoring** - API integration

### Example Prometheus Integration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'minio-debug'
    script_path: '/scripts/minio-debug.sh'
    scrape_interval: 60s
```