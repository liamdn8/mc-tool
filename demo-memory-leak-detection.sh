#!/bin/bash

# Memory Leak Detection Demo Script
# This demonstrates how mc-tool profile would detect memory leaks

echo "=== MinIO Memory Leak Detection Demo ==="
echo "📊 Simulating mc-tool profile heap minio-prod --detect-leaks --duration 5m"
echo

echo "🔍 Starting memory leak detection for alias: minio-prod"
echo "📈 Monitor interval: 10s"
echo "🚨 Memory threshold: 50 MB"
echo

# Simulate memory samples over time showing a leak
echo "📊 Taking sample 1..."
echo "[15:04:05] Alloc: 145.2 MB, Total: 2.1 GB, Sys: 234.5 MB, Goroutines: 1847, Objects: 145672"
sleep 1

echo "📊 Taking sample 2..."
echo "[15:04:15] Alloc: 167.8 MB, Total: 2.2 GB, Sys: 245.1 MB, Goroutines: 1923, Objects: 167834"
sleep 1

echo "📊 Taking sample 3..."
echo "[15:04:25] Alloc: 198.4 MB, Total: 2.4 GB, Sys: 267.2 MB, Goroutines: 2156, Objects: 198423"
sleep 1

echo "🚨 POTENTIAL LEAK DETECTED: Memory grew by 52.6 MB in the last interval"
sleep 1

echo "📊 Taking sample 4..."
echo "[15:04:35] Alloc: 234.7 MB, Total: 2.6 GB, Sys: 289.8 MB, Goroutines: 2387, Objects: 234891"
sleep 1

echo "🚨 GOROUTINE LEAK DETECTED: 231 new goroutines in the last interval"
sleep 1

echo "📊 Taking sample 5..."
echo "[15:04:45] Alloc: 278.3 MB, Total: 2.9 GB, Sys: 312.4 MB, Goroutines: 2634, Objects: 278456"
sleep 1

echo "⏰ Monitoring duration completed"
echo

echo "=== Memory Leak Analysis Report ==="
echo "📊 Total samples: 5"
echo "⏱️  Monitoring duration: 45s"
echo

echo "📈 Memory Growth Analysis:"
echo "  Initial Memory: 145.2 MB"
echo "  Final Memory: 278.3 MB"
echo "  Total Growth: 133.1 MB"
echo "  Hourly Growth Rate: 10,648.0 MB/hour"
echo

echo "🔄 Goroutine Analysis:"
echo "  Initial Goroutines: 1847"
echo "  Final Goroutines: 2634"
echo "  Growth: 787"
echo

echo "📦 Object Analysis:"
echo "  Initial Objects: 145672"
echo "  Final Objects: 278456"
echo "  Growth: 132784"
echo

echo "🕵️  Leak Detection Results:"
echo "  🚨 MEMORY LEAK LIKELY: High memory growth rate (10,648.0 MB/hour)"
echo "  🚨 GOROUTINE LEAK DETECTED: Significant goroutine growth (787)"
echo "  🚨 OBJECT LEAK DETECTED: Significant object growth (132784)"
echo

echo "💡 Recommendations:"
echo "  • Take a detailed heap profile for analysis: mc-tool profile heap minio-prod"
echo "  • Consider enabling GC profiling: mc-tool profile allocs minio-prod"
echo "  • Monitor for longer periods to confirm trends"
echo "  • Check application logs for error patterns"
echo

echo "=== Example Commands ==="
echo "# Basic heap profile"
echo "mc-tool profile heap minio-prod --duration 1m"
echo

echo "# Continuous leak monitoring"
echo "mc-tool profile heap minio-prod --detect-leaks --duration 30m --threshold-mb 100"
echo

echo "# CPU profiling for performance analysis"
echo "mc-tool profile cpu minio-prod --duration 60s --output /tmp/cpu.pprof"
echo

echo "# Goroutine leak detection"
echo "mc-tool profile goroutine minio-prod --duration 5m --output /tmp/goroutines.pprof"
echo

echo "# Using older mc version for compatibility"
echo "mc-tool profile heap minio-prod --mc-path mc-2021 --duration 30s"
echo

echo "=== Profile Types Available ==="
echo "• cpu:       CPU profiling to identify performance bottlenecks"
echo "• heap:      Memory heap profiling for memory leak detection"
echo "• goroutine: Goroutine profiling to find goroutine leaks"
echo "• allocs:    Allocation profiling for GC pressure analysis"
echo "• block:     Blocking profiling for synchronization issues"
echo "• mutex:     Mutex contention profiling"
echo

echo "✅ Demo completed! The mc-tool profile command provides comprehensive"
echo "   memory leak detection and performance profiling for MinIO servers."