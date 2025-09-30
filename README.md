# mc-tool
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
