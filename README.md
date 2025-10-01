# mc-tool

A comprehensive MinIO client support tool for comparing buckets, analyzing object distributions, and validating bucket configurations.

## Features

- **Compare Objects**: Compare objects between two MinIO buckets or paths
- **Analyze Buckets**: Analyze object distribution, versions, and incomplete uploads
- **Configuration Checklist**: Comprehensive bucket configuration validation including event settings and lifecycle policies

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
│   └── validation/           # Bucket configuration validation
│       └── validation.go
└── README.md
```

### Package Responsibilities

- **`pkg/config`**: Handles loading MinIO client configuration from `~/.mc/config.json`
- **`pkg/client`**: Creates MinIO clients and parses URLs
- **`pkg/compare`**: Implements object comparison logic and result display
- **`pkg/analyze`**: Provides bucket analysis including object distribution and incomplete uploads
- **`pkg/validation`**: Validates bucket configurations (versioning, notifications, lifecycle, encryption, policies)

## Usage

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
- MinIO client (`mc`) configured with your MinIO/S3 endpoints
- Access to MinIO/S3 instances you want to analyze

The tool reads configuration from `~/.mc/config.json`, so ensure you have configured your MinIO aliases using the `mc` command.

## Examples

### Example 1: Compare Production and Staging Buckets

```bash
mc-tool compare prod/my-bucket staging/my-bucket --verbose
```

### Example 2: Analyze Bucket for Hidden Objects

```bash
mc-tool analyze prod/my-bucket --verbose
```

### Example 3: Validate Bucket Configuration

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

## License

This project is licensed under the MIT License - see the LICENSE file for details.
MinIO client based support tool

A powerful command-line tool for comparing MinIO buckets across different instances with support for versioned objects.

## Features

- Compare objects between two MinIO buckets or paths
- Support for versioned object comparison with `--versions` flag
- Default comparison using ETag and size for current object versions
- Load configuration from existing mc client configuration
- Detailed comparison results with summary statistics
- Verbose mode for detailed object information

## Installation

```bash
git clone <repository-url>
cd mc-tool
go mod tidy
go build -o mc-tool
```

## Usage

### Basic Comparison (Current Versions Only)
```bash
./mc-tool compare alias1/bucket1 alias2/bucket2
./mc-tool compare alias1/bucket1/folder alias2/bucket2/folder
```

### Compare All Object Versions
```bash
./mc-tool compare --versions alias1/bucket1 alias2/bucket2
```

### Verbose Output
```bash
./mc-tool compare --verbose alias1/bucket1 alias2/bucket2
./mc-tool compare --versions --verbose alias1/bucket1 alias2/bucket2
```

### Skip TLS Certificate Verification
```bash
./mc-tool compare --insecure alias1/bucket1 alias2/bucket2
./mc-tool compare --versions --insecure --verbose alias1/bucket1 alias2/bucket2
```

## Configuration

The tool reads MinIO configuration from the standard mc client configuration file located at `~/.mc/config.json`. Make sure you have configured your aliases using the mc client:

```bash
mc alias set myalias https://minio.example.com ACCESS_KEY SECRET_KEY
```

## Comparison Logic

### Default Mode (Current Versions)
- Compares only the latest version of each object
- Uses ETag and file size for comparison
- Objects are considered identical if both ETag and size match

### Versions Mode (`--versions`)
- Compares all versions of each object by version ID
- Each version is compared individually
- Useful for ensuring complete replication including historical versions

## Output

The tool provides:
- ✓ Identical objects (shown only in verbose mode)
- ⚠ Different objects with details about differences
- \- Objects missing in source
- \+ Objects missing in target
- Summary statistics

## Exit Codes

- 0: All objects are identical
- 1: Differences found (different objects, missing objects)

## Examples

```bash
# Compare two buckets on different MinIO instances
./mc-tool compare prod/data-bucket staging/data-bucket

# Compare specific paths with version support
./mc-tool compare --versions prod/backup/2024 staging/backup/2024

# Verbose comparison of current versions
./mc-tool compare --verbose local/test-bucket remote/test-bucket

# Skip TLS verification for self-signed certificates
./mc-tool compare --insecure local/test-bucket remote/test-bucket
```
