# MinIO Profiling and Memory Leak Detection

## Overview

The `mc-tool profile` command replaces the previous debug functionality and provides comprehensive performance profiling and memory leak detection for MinIO servers using the MinIO client's `mc admin profile` or `mc support profile` commands.

## Features

### Profile Types
- **cpu**: CPU profiling to identify performance bottlenecks
- **heap**: Memory heap profiling for memory leak detection  
- **goroutine**: Goroutine profiling to find goroutine leaks
- **allocs**: Allocation profiling for GC pressure analysis
- **block**: Blocking profiling for synchronization issues
- **mutex**: Mutex contention profiling

### Advanced Features
- **Automatic version detection**: Supports both newer `mc support profile` and older `mc admin profile` commands
- **Memory leak monitoring**: Continuous monitoring with configurable thresholds
- **Growth rate analysis**: Calculates hourly memory/goroutine growth rates
- **Immediate leak alerts**: Real-time detection of memory spikes
- **Profile persistence**: Save profiles to files for detailed analysis
- **Multiple mc versions**: Support for mc, mc-2021, and custom binary paths

## Command Syntax

```bash
mc-tool profile <type> <alias> [flags]
```

### Flags
- `--duration string`: Profile duration (default: 30s)
- `--detect-leaks`: Enable memory leak detection monitoring
- `--monitor-interval string`: Monitoring interval for leak detection (default: 10s)
- `--threshold-mb int`: Memory growth threshold in MB (default: 50)
- `--output string`: Output file path for profile data
- `--mc-path string`: Path to mc binary (default: "mc")
- `--verbose, -v`: Verbose output
- `--insecure`: Skip TLS certificate verification

## Basic Usage Examples

### Single Profile Collection
```bash
# Basic heap profile for memory analysis
mc-tool profile heap minio-prod

# CPU profile with custom duration
mc-tool profile cpu minio-prod --duration 60s

# Goroutine profile saved to file
mc-tool profile goroutine minio-prod --output /tmp/goroutines.pprof
```

### Memory Leak Detection
```bash
# Basic leak detection with heap profiling
mc-tool profile heap minio-prod --detect-leaks --duration 10m

# Continuous monitoring with custom threshold
mc-tool profile heap minio-prod --detect-leaks --threshold-mb 100 --duration 1h

# High-frequency monitoring for critical systems
mc-tool profile heap minio-prod --detect-leaks --monitor-interval 5s --duration 30m
```

### Version Compatibility
```bash
# Use older mc version for compatibility
mc-tool profile heap minio-prod --mc-path mc-2021

# Use custom mc binary path
mc-tool profile cpu minio-prod --mc-path /opt/minio/mc-custom
```

## Memory Leak Detection Output

### Real-time Monitoring
```
🔍 Starting memory leak detection for alias: minio-prod
📈 Monitor interval: 10s
🚨 Memory threshold: 50 MB

📊 Taking sample 1...
[15:04:05] Alloc: 145.2 MB, Total: 2.1 GB, Sys: 234.5 MB, Goroutines: 1847, Objects: 145672

📊 Taking sample 2...
[15:04:15] Alloc: 167.8 MB, Total: 2.2 GB, Sys: 245.1 MB, Goroutines: 1923, Objects: 167834

🚨 POTENTIAL LEAK DETECTED: Memory grew by 52.6 MB in the last interval
```

### Analysis Report
```
=== Memory Leak Analysis Report ===
📊 Total samples: 30
⏱️  Monitoring duration: 5m0s

📈 Memory Growth Analysis:
  Initial Memory: 145.2 MB
  Final Memory: 278.3 MB
  Total Growth: 133.1 MB
  Hourly Growth Rate: 1,597.2 MB/hour

🔄 Goroutine Analysis:
  Initial Goroutines: 1847
  Final Goroutines: 2634
  Growth: 787

🕵️  Leak Detection Results:
  🚨 MEMORY LEAK LIKELY: High memory growth rate (1,597.2 MB/hour)
  🚨 GOROUTINE LEAK DETECTED: Significant goroutine growth (787)

💡 Recommendations:
  • Take a detailed heap profile for analysis: mc-tool profile heap minio-prod
  • Consider enabling GC profiling: mc-tool profile allocs minio-prod
  • Monitor for longer periods to confirm trends
```

## Advanced Use Cases

### Production Monitoring Workflow
```bash
# 1. Quick health check
mc-tool profile heap production-cluster --duration 30s

# 2. If issues detected, start continuous monitoring
mc-tool profile heap production-cluster --detect-leaks --duration 4h --threshold-mb 200

# 3. Collect detailed profiles for analysis
mc-tool profile cpu production-cluster --duration 5m --output /analysis/cpu-$(date +%Y%m%d-%H%M%S).pprof
mc-tool profile heap production-cluster --duration 2m --output /analysis/heap-$(date +%Y%m%d-%H%M%S).pprof
mc-tool profile goroutine production-cluster --output /analysis/goroutines-$(date +%Y%m%d-%H%M%S).pprof
```

### Development and Testing
```bash
# Monitor during load testing
mc-tool profile heap dev-cluster --detect-leaks --duration 1h --monitor-interval 30s

# Profile allocation patterns
mc-tool profile allocs dev-cluster --duration 2m --output /dev/allocs.pprof

# Check for blocking operations
mc-tool profile block dev-cluster --duration 1m --output /dev/block.pprof
```

## Troubleshooting

### MC Version Compatibility
The tool automatically detects and handles different mc versions:

1. **Newer mc versions**: Uses `mc support profile` command
2. **Older mc versions**: Falls back to `mc admin profile` command
3. **Version detection failure**: Use `--mc-path mc-2021` for older versions

### Common Issues

#### "Deprecated command" Error
```bash
# Error: mc: <ERROR> Deprecated command. Please use 'mc support profile' instead.
# Solution: Use older mc version
mc-tool profile heap minio-prod --mc-path mc-2021
```

#### Permission Denied
```bash
# Ensure alias has admin permissions for profiling
mc alias list
mc admin user list minio-prod
```

#### No Profile Data
```bash
# Increase duration for better data collection
mc-tool profile heap minio-prod --duration 60s --verbose
```

## Integration with Docker

The Docker image includes both mc versions for maximum compatibility:

```bash
# Run profiling in container
docker run --rm mc-tool:latest mc-tool profile heap minio-prod --mc-path mc-2021

# Mount output directory for profile files
docker run --rm -v /host/profiles:/profiles mc-tool:latest \
  mc-tool profile cpu minio-prod --output /profiles/cpu.pprof
```

## Performance Impact

Profile collection has minimal impact on MinIO performance:
- **CPU profiling**: <1% CPU overhead during collection
- **Heap profiling**: ~100-500KB memory overhead
- **Goroutine profiling**: Minimal impact, snapshot-based
- **Continuous monitoring**: <0.1% overhead with 10s intervals

## Profile Analysis

Collected profiles can be analyzed using standard Go tools:

```bash
# Analyze CPU profile
go tool pprof cpu.pprof

# Analyze heap profile  
go tool pprof heap.pprof

# Web interface for visualization
go tool pprof -http=:8080 heap.pprof
```

## Migration from Debug Command

The profile command replaces the previous debug functionality with enhanced capabilities:

| Old Debug Command | New Profile Command |
|-------------------|-------------------|
| `mc-tool debug alias` | `mc-tool profile goroutine alias` |
| `mc-tool debug alias --monitor 10m` | `mc-tool profile heap alias --detect-leaks --duration 10m` |
| `mc-tool debug alias --format json` | `mc-tool profile heap alias --output profile.json` |

The new profile command provides more accurate leak detection using actual MinIO profiling data rather than HTTP endpoint scraping.