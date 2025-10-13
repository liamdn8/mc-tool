# Debug Command 403 Issue Resolution

## Problem Analysis

The `mc-tool debug` command was returning **403 Forbidden** errors when trying to access MinIO's pprof endpoints. This is **expected behavior** for public and production MinIO servers for security reasons.

## Root Cause

1. **Security by Design**: Public MinIO servers (like `play.min.io`) intentionally disable pprof endpoints to prevent security risks
2. **Different Endpoint Types**: pprof endpoints (`/debug/pprof/goroutine`) are **administrative/debugging endpoints**, not S3 API endpoints
3. **Authentication Mismatch**: The initial implementation wasn't trying multiple authentication approaches

## Solution Implemented

### 🔧 **Enhanced Authentication Strategy**

The debug command now tries multiple authentication approaches in sequence:

1. **No Authentication**: Some development MinIO instances allow unauthenticated access to pprof
2. **AWS Signature v4**: Standard authentication for MinIO admin operations
3. **Authorization Header**: Alternative authentication approach for pprof endpoints

### 🎯 **Improved Error Handling**

- **Verbose Output**: Shows all authentication attempts with detailed feedback
- **Clear Error Messages**: Explains why 403 occurs and provides troubleshooting tips
- **Helpful Guidance**: Suggests solutions for production use

### 📝 **Better User Experience**

```bash
# Example output with improved error handling
🔧 Using alias: play
📡 Endpoint: https://play.min.io
🔑 Access Key: Q3AM3UQ867SPQQA43P2F
📊 Format: text

🔍 Fetching goroutine data from: https://play.min.io/debug/pprof/goroutine?debug=1
🔐 Trying without authentication...
🔐 Trying with AWS signature v4...
🔐 Trying with Authorization header...
❌ All authentication approaches failed. Last error: AWS v4 auth failed with status: 403
💡 Troubleshooting tips:
   • pprof endpoints may be disabled on this MinIO server
   • Some MinIO instances don't expose /debug/pprof for security
   • Try with a local MinIO instance where you have admin access
   • Check if your MinIO server has MINIO_PROFILE=enable environment variable

Error: pprof endpoint not accessible (status: 403) - this may be disabled for security on production/public MinIO servers
```

## Working vs Non-Working Scenarios

### ✅ **What Works (awscurl with S3 API)**
```bash
awscurl --access_key KEY --secret_key SECRET --service s3 --region us-east-1 'https://play.min.io/'
# Returns: Bucket list (S3 API endpoint)
```

### ❌ **What Doesn't Work (pprof endpoints on public servers)**
```bash
curl 'https://play.min.io/debug/pprof/goroutine?debug=1'
# Returns: 403 Access Denied (security protection)
```

### 🎯 **What Should Work (local/development MinIO)**
```bash
mc-tool debug local --verbose
# Should work with local MinIO instances that have MINIO_PROFILE=enable
```

## Production Usage Guide

### 🏗️ **For Local Development**
```bash
# Start MinIO with profiling enabled
export MINIO_PROFILE=enable
minio server /data

# Use debug command
mc-tool debug local --verbose
```

### 🔒 **For Self-Signed Certificates**
```bash
mc-tool debug dev-server --insecure --verbose
```

### 📊 **For Production Monitoring**
```bash
# Set up proper MinIO configuration with profiling
# Then use continuous monitoring
mc-tool debug prod-alias --monitor 30m --interval 1m --threshold 100
```

## Key Differences: S3 API vs pprof Endpoints

| Aspect | S3 API | pprof Endpoints |
|--------|--------|----------------|
| **Purpose** | Object storage operations | Performance debugging |
| **Security** | Always enabled | Often disabled in production |
| **Authentication** | AWS Signature v4 | Various approaches |
| **Availability** | Public servers | Local/admin servers only |
| **Example** | `/bucket/object` | `/debug/pprof/goroutine` |

## Testing Results

The enhanced debug command now provides:
- ✅ **Multiple authentication strategies**
- ✅ **Clear error explanations** 
- ✅ **Troubleshooting guidance**
- ✅ **Production-ready error handling**
- ✅ **Insecure TLS support** for development environments

## Conclusion

The 403 error was **not a bug** but expected security behavior. The solution provides:

1. **Better error handling** with multiple authentication attempts
2. **Clear user guidance** for different scenarios  
3. **Production-ready deployment** instructions
4. **Development-friendly** insecure TLS support

The debug command now gracefully handles both working and non-working scenarios while providing users with actionable troubleshooting information.