#!/bin/bash

# Demo script showing mc-tool debug command output formats

echo "🧪 MC-Tool Debug Command Output Examples"
echo "========================================"
echo

echo "1. 📊 NORMAL TEXT OUTPUT (Single Snapshot)"
echo "----------------------------------------"
cat << 'EOF'
🔧 Using alias: minio-prod
📡 Endpoint: https://minio.company.com:9000
🔑 Access Key: minioadmin
📊 Format: text

🔍 MinIO Goroutine Analysis Report
==================================
📅 Timestamp: 2024-10-01 14:30:15 UTC
📡 Endpoint: https://minio.company.com:9000

📈 Summary Statistics:
   Total Goroutines: 127
   Active (running): 8
   Waiting (select): 45
   Blocked (chan): 12
   System calls: 23
   Network I/O: 18
   Other states: 21

⏱️ Long-Running Goroutines (>5m):
   ID:45 [select] 2h15m - net/http.(*Server).Serve
   ID:23 [syscall] 1h45m - runtime.epollwait
   ID:78 [chan receive] 45m - github.com/minio/minio/cmd.(*xlSets).HealObjects

🚨 Potential Leaks (3):
   ID:156 [select] runtime.gopark
      Stack: runtime.gopark(0x4f2c20, 0xc000432e70, 0x41d, 0x17)
   ID:189 [chan receive] sync.(*WaitGroup).Wait
      Stack: sync.(*WaitGroup).Wait(0xc0004a8480)
   ID:203 [select] time.Sleep
      Stack: time.Sleep(0x3b9aca00)

🔝 Top Functions:
   1. runtime.gopark: 34 calls
   2. net/http.(*conn).serve: 18 calls
   3. runtime.selectgo: 15 calls
   4. sync.(*WaitGroup).Wait: 12 calls
   5. github.com/minio/minio/cmd.serve: 8 calls

✅ Analysis completed successfully!
EOF

echo
echo "2. 📋 JSON OUTPUT FORMAT"
echo "------------------------"
cat << 'EOF'
{
  "timestamp": "2024-10-01T14:30:15Z",
  "endpoint": "https://minio.company.com:9000",
  "summary": {
    "total": 127,
    "by_state": {
      "running": 8,
      "select": 45,
      "chan": 12,
      "syscall": 23,
      "IO wait": 18,
      "other": 21
    }
  },
  "long_running": [
    {
      "id": 45,
      "state": "select",
      "duration": "2h15m",
      "function": "net/http.(*Server).Serve"
    },
    {
      "id": 23,
      "state": "syscall", 
      "duration": "1h45m",
      "function": "runtime.epollwait"
    }
  ],
  "potential_leaks": [
    {
      "id": 156,
      "state": "select",
      "function": "runtime.gopark",
      "stack": "runtime.gopark(0x4f2c20, 0xc000432e70, 0x41d, 0x17)"
    },
    {
      "id": 189,
      "state": "chan receive",
      "function": "sync.(*WaitGroup).Wait", 
      "stack": "sync.(*WaitGroup).Wait(0xc0004a8480)"
    }
  ],
  "top_functions": [
    {"function": "runtime.gopark", "count": 34},
    {"function": "net/http.(*conn).serve", "count": 18},
    {"function": "runtime.selectgo", "count": 15}
  ]
}
EOF

echo
echo "3. 🔄 MONITORING MODE OUTPUT"
echo "----------------------------"
cat << 'EOF'
🔍 Monitoring MinIO goroutines for 10m (interval: 30s)
📡 Endpoint: https://minio.company.com:9000
🎯 Leak threshold: 50 goroutines

[14:30:15] 📊 Baseline: 127 goroutines
[14:30:45] 📈 Current: 134 goroutines (+7)
[14:31:15] 📈 Current: 145 goroutines (+18) 
[14:31:45] 🚨 ALERT: 178 goroutines (+51) - LEAK DETECTED!
           📍 Growth pattern: Steady increase over 2 minutes
           🔍 Suspected cause: Channel receiver stuck
           📋 Top growing functions:
              - sync.(*WaitGroup).Wait: +15 instances
              - runtime.chanrecv: +12 instances

[14:32:15] 🚨 ALERT: 195 goroutines (+68) - LEAK CONFIRMED!
           ⚠️  Memory leak pattern detected
           📊 Growth rate: 34 goroutines/minute
           🛑 Recommendation: Investigate channel operations

📝 Final Report:
   • Total monitoring time: 10m
   • Goroutine growth: 127 → 245 (+118)
   • Leak episodes: 2
   • Peak growth rate: 34/min
   • Status: 🚨 LEAK DETECTED
EOF

echo
echo "4. 🔒 INSECURE CONNECTION OUTPUT"
echo "--------------------------------"
cat << 'EOF'
🔧 Using alias: dev-minio
📡 Endpoint: https://dev.company.com:9000
🔑 Access Key: minioadmin
⚠️  TLS certificate verification disabled
📊 Format: text

🔍 MinIO Goroutine Analysis Report
==================================
📅 Timestamp: 2024-10-01 14:30:15 UTC
📡 Endpoint: https://dev.company.com:9000 (insecure)

📈 Summary Statistics:
   Total Goroutines: 45
   Active (running): 4
   Waiting (select): 20
   Blocked (chan): 8
   System calls: 13

✅ No memory leaks detected
📊 System appears healthy
EOF

echo
echo "5. ❌ ERROR SCENARIOS"
echo "--------------------"
echo "Connection Error:"
cat << 'EOF'
🔧 Using alias: offline-minio
📡 Endpoint: https://offline.company.com:9000
🔑 Access Key: minioadmin

2024/10/01 14:30:15 Debug analysis failed: failed to analyze goroutines: dial tcp: connect: connection refused
EOF

echo
echo "Authentication Error:"
cat << 'EOF'
🔧 Using alias: minio-prod
📡 Endpoint: https://minio.company.com:9000
🔑 Access Key: wrong-key

2024/10/01 14:30:15 Debug analysis failed: failed to analyze goroutines: unexpected status code: 403
EOF

echo
echo "Missing Alias Error:"
cat << 'EOF'
Error: alias 'nonexistent' not found in MC configuration
Available aliases: local, play, minio-prod, dev-minio

Use 'mc alias set <name> <endpoint> <access-key> <secret-key>' to add new aliases
EOF

echo
echo "✅ mc-tool debug command provides comprehensive goroutine analysis"
echo "   with multiple output formats and monitoring capabilities!"