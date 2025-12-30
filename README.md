# mc-tool

A comprehensive MinIO client support tool for comparing buckets, analyzing object distributions, validating bucket configurations, and profiling MinIO servers for performance analysis and memory leak detection.

## Features

- **🌐 Web UI**: Modern bilingual (English/Vietnamese) web interface for easy operation
- **Compare Objects**: Compare objects between two MinIO buckets or paths
- **Analyze Buckets**: Analyze object distribution, versions, and incomplete uploads  
- **Error Trace Capture**: Capture and analyze MinIO error events with grouping by object, version, API, or client
- **Configuration Checklist**: Comprehensive bucket configuration validation including event settings and lifecycle policies
- **Performance Profiling**: CPU, memory, and goroutine profiling for MinIO servers
- **Memory Leak Detection**: Continuous monitoring and automatic leak detection with configurable thresholds
- **⚡ Performance Testing**: Automated PUT replication testing with metrics and analytics

## Quick Start

### Web UI (Recommended for Operators)

```bash
# Start the web interface
mc-tool web

# Or on custom port
mc-tool web --port 3000
```

Then open your browser at `http://localhost:8080` for a user-friendly interface with:
- 🌐 Bilingual support (English & Vietnamese)
- 📊 Interactive dashboard
- 🔄 Visual bucket comparison
- 📈 Bucket analysis with charts
- 🔍 Performance profiling
- ✅ Configuration checklist

#### Web UI Configuration

The web UI can be configured using environment variables:

```bash
# Basic Configuration
PORT=8080                    # Web server port (default: 8080)
REFRESH_INTERVAL=60s         # UI auto-refresh interval (default: 60s)

# Logging
LOG_LEVEL=info              # Logging level: debug, info, warn, error (default: info)
LOG_FORMAT=json             # Log format: text or json (default: text)

# Data Storage
DATA_DIR=./data             # Directory for history and audit logs (default: ./data)
REFERENCE_CONFIGS_DIR=/app/reference-configs  # Reference configs directory

# Alerts (Coming Soon)
ALERT_WEBHOOK_URL=https://hooks.slack.com/...  # Webhook URL for alerts
ALERT_EMAIL_TO=ops@example.com                  # Email address for alerts

# Authentication (Coming Soon)
AUTH_METHOD=none            # Authentication method: none, basic, oauth2, saml
```

Example with environment variables:
```bash
# JSON logging for production
LOG_FORMAT=json LOG_LEVEL=info PORT=8080 mc-tool web

# Debug mode with custom port
LOG_LEVEL=debug PORT=9000 mc-tool web
```

#### Healthcheck Endpoint

For container orchestration (Kubernetes, Docker Swarm), use the `/healthz` endpoint:

```bash
curl http://localhost:8080/healthz
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-13T05:36:14Z",
  "mc_available": true
}
```

Kubernetes example:
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

See [Web UI Documentation](docs/WEB_UI.md) for more details.

### Command Line Interface

```bash
# Compare buckets
mc-tool compare alias1/bucket1 alias2/bucket2

# Analyze bucket
mc-tool analyze alias/bucket

# Capture error traces
mc-tool trace minio-prod --duration 30s

# Profile server
mc-tool profile heap minio-prod --detect-leaks --duration 5m

# Run checklist
mc-tool checklist alias/bucket
```

## Architecture

The application is organized into the following packages for better maintainability:

```
mc-tool/
├── main.go                    # Main application entry point
├── pkg/
│   ├── config/               # MC configuration handling
│   │   └── config.go
│   ├── client/               # MinIO client creation and utilities
│   │   └── client.go
│   ├── compare/              # Object comparison functionality
│   │   └── compare.go
│   ├── analyze/              # Bucket analysis functionality
│   │   └── analyze.go
│   ├── profile/              # Performance profiling and memory leak detection
│   │   └── profile.go
│   ├── validation/           # Bucket configuration validation
│   │   └── validation.go
│   └── web/                  # Web UI server and API
│       ├── server.go
│       └── static/
│           ├── index.html
│           ├── styles.css
│           └── app.js
└── docs/
    └── WEB_UI.md            # Web UI documentation
```

### Package Responsibilities

- **`pkg/config`**: Handles loading MinIO client configuration from `~/.mc/config.json` or `MC_HOST_*` environment variables
- **`pkg/client`**: Creates MinIO clients and parses URLs
- **`pkg/compare`**: Implements object comparison logic and result display
- **`pkg/analyze`**: Provides bucket analysis including object distribution and incomplete uploads
- **`pkg/profile`**: Performance profiling and memory leak detection using mc admin/support profile commands
- **`pkg/perftest`**: Automated performance testing for PUT operations with replication monitoring
- **`pkg/validation`**: Validates bucket configurations (versioning, notifications, lifecycle, encryption, policies)
- **`pkg/web`**: Web server with REST API and bilingual UI for easy operation

## Usage

### Web UI

Start the web interface for easy, visual operation:

```bash
mc-tool web
```

Features:
- No command-line knowledge required
- Bilingual interface (English/Vietnamese)
- Real-time job progress tracking
- Interactive forms and dropdowns
- Visual results display

### Compare Objects

```bash
# Compare current versions only
mc-tool compare alias1/bucket1 alias2/bucket2

# Compare all versions including old versions and delete markers
mc-tool compare --versions alias1/bucket1 alias2/bucket2

# Compare with verbose output
mc-tool compare --verbose alias1/bucket1/folder alias2/bucket2/folder

# Skip TLS certificate verification
mc-tool compare --insecure alias1/bucket1 alias2/bucket2
```

### Analyze Bucket

```bash
# Basic bucket analysis
mc-tool analyze alias/bucket

# Analyze with verbose output showing detailed object information
mc-tool analyze --verbose alias/bucket

# Analyze specific path within bucket
mc-tool analyze alias/bucket/path
```

### Configuration Checklist

```bash
# Check bucket configuration including event settings and lifecycle
mc-tool checklist alias/bucket

# Run checklist with verbose output
mc-tool checklist --verbose alias/bucket

# Skip TLS certificate verification
mc-tool checklist --insecure alias/bucket
```

### Performance Testing (NEW)

Automated performance testing for MinIO PUT operations with replication monitoring:

```bash
# Basic test: 100 objects across 3 sites
mc-tool perftest --sites site1,site2,site3 --bucket test-bucket --count 100

# Test with parallelism for higher throughput
mc-tool perftest --sites site1,site2,site3 --bucket test-bucket --count 1000 --parallel 10

# Overwrite test for versioning buckets
mc-tool perftest --sites site1,site2,site3 --bucket test-bucket --count 100 --overwrite

# Custom object size and path
mc-tool perftest --sites site1,site2,site3 --bucket test-bucket --path testdata/ --size 10240 --count 500

# With Prometheus monitoring
mc-tool perftest --sites site1,site2,site3 --bucket test-bucket --count 100 --prometheus http://prometheus:9090
```

The perftest command provides:
- 📊 PUT operation latency metrics
- ⏱️ Replication lag measurement across sites
- 🔍 Object consistency verification
- 📈 Throughput analytics (objects/sec, bytes/sec)
- 📉 Per-site statistics and performance breakdown
- 🔗 Optional Prometheus integration for system metrics correlation

See [Performance Testing Documentation](docs/PERFTEST_FEATURE.md) for details.

### Error Trace Capture and Analysis

Capture and analyze error events from MinIO admin trace:

```bash
# Basic trace capture (5 seconds)
mc-tool trace minio-prod

# Extended trace capture with custom duration
mc-tool trace minio-prod --duration 30s

# Filter by HTTP status codes
mc-tool trace minio-prod --status 403 --status 500

# Filter by error message content
mc-tool trace minio-prod --error-contains "Access Denied"

# Group by API operations
mc-tool trace minio-prod --group-by-api --duration 15s

# Group by client address
mc-tool trace minio-prod --group-by-client --duration 15s

# Group by object AND version (for versioned buckets)
mc-tool trace minio-prod --versions --duration 10s

# Comprehensive trace with all options
mc-tool trace minio-prod --versions --group-by-api --group-by-client --verbose --duration 30s

# Use older mc version for compatibility
mc-tool trace minio-prod --mc-path mc-2021
```

The trace command helps identify:
- 🔍 **Error Patterns**: Groups similar errors together for easier diagnosis
- 📊 **Top Failed Objects**: Shows which objects are experiencing the most errors
- 🔑 **Version-Specific Issues**: With `--versions`, identifies errors affecting specific object versions
- 🌐 **API Analysis**: With `--group-by-api`, shows which operations are failing
- 👥 **Client Analysis**: With `--group-by-client`, identifies problematic clients

### Performance Profiling and Memory Leak Detection

```bash
# Basic heap profile for memory analysis
mc-tool profile heap minio-prod

# CPU profiling for performance analysis
mc-tool profile cpu minio-prod --duration 60s

# Memory leak detection with continuous monitoring
mc-tool profile heap minio-prod --detect-leaks --duration 30m --threshold-mb 100

# Goroutine profiling to detect goroutine leaks
mc-tool profile goroutine minio-prod --output /tmp/goroutines.pprof

# Use older mc version for compatibility
mc-tool profile heap minio-prod --mc-path mc-2021
```

### Configuration Validation

The `checklist` command performs comprehensive validation of:

- ✅ **Bucket Existence**: Verifies the bucket exists
- ✅ **Versioning**: Checks if versioning is enabled and provides recommendations
- ✅ **Event Notifications**: Validates Lambda, Topic, and Queue configurations
- ✅ **Object Lifecycle**: Analyzes lifecycle rules and incomplete multipart upload handling
- ✅ **Server-side Encryption**: Checks encryption configuration
- ✅ **Bucket Policies**: Validates policies and warns about overly permissive settings

## Installation

### Option 1: Pre-built Binaries (Recommended)

Pre-built static binaries are available in the `build/` directory after building:

```bash
# Clone the repository
git clone https://github.com/liamdn8/mc-tool.git
cd mc-tool

# Build static binaries for all platforms
make build-all

# For a single static binary (Linux only)
make build-static        # With debug info (14MB)
make build-portable      # Stripped for production (9.7MB)

# Use the portable binary (recommended for production)
./build/mc-tool-portable version
sudo cp build/mc-tool-portable /usr/local/bin/mc-tool
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/liamdn8/mc-tool.git
cd mc-tool

# Install dependencies
make deps

# Build the application
go build -o mc-tool

# Make it executable and optionally move to PATH
chmod +x mc-tool
sudo mv mc-tool /usr/local/bin/  # Optional: install globally
```

### Build Targets

- `make build` - Regular build for development
- `make build-static` - Static binary with debug symbols (14MB)
- `make build-portable` - Static binary stripped for production (9.7MB) 
- `make build-all` - Cross-compile for all platforms
- `make test` - Run all tests
- `make clean` - Clean build artifacts

### Available Binaries

After running `make build-all`, you'll have:
- `mc-tool-linux-amd64-static` - Linux x86_64 (static)
- `mc-tool-linux-arm64-static` - Linux ARM64 (static)
- `mc-tool-darwin-amd64` - macOS Intel
- `mc-tool-darwin-arm64` - macOS Apple Silicon
- `mc-tool-windows-amd64.exe` - Windows 64-bit
- `mc-tool-portable` - Recommended for production (smallest, static)

## Prerequisites

- Go 1.21 or later
- MinIO client (`mc`) configured with your MinIO/S3 endpoints (optional if using environment variables)
- Access to MinIO/S3 instances you want to analyze

### Configuration Options

The tool supports two methods for configuring MinIO aliases:

1. **Configuration file**: `~/.mc/config.json` (traditional method using `mc alias set`)
2. **Environment variables**: `MC_HOST_<alias>` (new method, perfect for containers and CI/CD)

```bash
# Example using environment variables
export MC_HOST_site1=https://accesskey:secretkey@minio1.example.com:9000
export MC_HOST_site2=http://minioadmin:minioadmin@localhost:9000

# Now use the aliases
mc-tool compare site1/bucket site2/bucket
```

For detailed configuration examples, see [docs/CONFIG.md](docs/CONFIG.md).

## Examples

### Example 1: Compare Production and Staging Buckets

```bash
mc-tool compare prod/my-bucket staging/my-bucket --verbose
```

### Example 2: Analyze Bucket for Hidden Objects

```bash
mc-tool analyze prod/my-bucket --verbose
```

### Example 3: Using Environment Variables for Configuration

```bash
# Configure aliases via environment variables
export MC_HOST_prod=https://prodkey:prodsecret@minio-prod.example.com:9000
export MC_HOST_staging=https://stagingkey:stagingsecret@minio-staging.example.com:9000

# Compare without needing config file
mc-tool compare prod/my-bucket staging/my-bucket
```

### Example 3: Validate Bucket Configuration

```bash
mc-tool checklist prod/my-bucket
```

### Example 4: Run Performance Test

```bash
mc-tool perftest --sites site1,site2,site3 --bucket perf-test --count 100 --parallel 5
```

This will output a performance report like:

```
🚀 Starting MinIO Performance Test

Configuration:
  Sites:       [site1 site2 site3]
  Bucket:      perf-test
  Object Count: 100
  Parallelism: 5

================================================================================
Test Results
================================================================================

Duration: 2m15s
Objects:  100 total, 98 successful, 2 failed

Performance Metrics:
  Average PUT Latency:      150ms
  Average Replication Lag:  2.5s
  Throughput:               0.74 objects/sec
  Data Throughput:          758.05 bytes/sec
```

### Example 5: Bucket Configuration Checklist

```bash
mc-tool checklist prod/my-bucket
```

This will output a comprehensive report like:

```
=== Bucket Configuration Checklist ===
Checking bucket: my-bucket

✅ Bucket exists
⚠️  Versioning: Disabled - Consider enabling for data protection
✅ Event Notifications: 3 configurations found
   - Lambda configurations: 1
   - Topic configurations: 2
✅ Object Lifecycle: 2 rules configured
   - Rule 'delete-old-versions': Enabled
   - Rule 'abort-incomplete-uploads': Enabled
✅ Server-side Encryption: AES256 configured
✅ Bucket Policy: Configured
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes following the existing package structure
4. Test your changes
5. Submit a pull request

## Documentation

- [Profile Command Guide](docs/PROFILE_COMMAND.md) - Comprehensive profiling and memory leak detection
- [Debug 403 Resolution](docs/DEBUG_403_RESOLUTION.md) - Understanding authentication differences

## License

This project is licensed under the MIT License - see the LICENSE file for details.
