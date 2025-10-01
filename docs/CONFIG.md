# Example configuration for mc-tool

This tool reads MinIO configuration from the standard mc client configuration file.

## Setting up MinIO aliases

Before using mc-tool, you need to configure your MinIO aliases using the mc client:

```bash
# Add a MinIO alias
mc alias set minio1 https://minio1.example.com ACCESS_KEY SECRET_KEY

# Add another MinIO alias  
mc alias set minio2 https://minio2.example.com ACCESS_KEY SECRET_KEY

# List configured aliases
mc alias list
```

## Example mc config.json structure

The configuration file is typically located at `~/.mc/config.json`:

```json
{
  "version": "10",
  "aliases": {
    "local": {
      "url": "http://localhost:9000",
      "accessKey": "minioadmin",
      "secretKey": "minioadmin",
      "api": "s3v4",
      "path": "auto"
    },
    "prod": {
      "url": "https://minio-prod.example.com",
      "accessKey": "your-access-key",
      "secretKey": "your-secret-key", 
      "api": "s3v4",
      "path": "auto"
    },
    "staging": {
      "url": "https://minio-staging.example.com",
      "accessKey": "your-access-key",
      "secretKey": "your-secret-key",
      "api": "s3v4", 
      "path": "auto",
      "insecure": true
    }
  }
}
```

## SSL/TLS Configuration

The tool supports both secure and insecure connections:

### Method 1: Configure in mc config
Add `"insecure": true` to your alias configuration for self-signed certificates:

```json
{
  "staging": {
    "url": "https://minio-staging.example.com",
    "accessKey": "your-access-key",
    "secretKey": "your-secret-key",
    "api": "s3v4",
    "path": "auto",
    "insecure": true
  }
}
```

### Method 2: Use command line flag
Use the `--insecure` flag to skip certificate verification (overrides config setting):

```bash
./mc-tool compare --insecure alias1/bucket alias2/bucket
```

### For production environments:
1. Use proper SSL certificates
2. Remove the `insecure` setting from config
3. Set up certificate trust chains properly

## Usage Examples

```bash
# Compare buckets between local and production
./mc-tool compare local/test-bucket prod/test-bucket

# Compare with version support
./mc-tool compare --versions prod/backup staging/backup

# Compare specific paths
./mc-tool compare prod/data/2024/01 staging/data/2024/01

# Verbose output for detailed information
./mc-tool compare --verbose --versions local/bucket prod/bucket

# Skip TLS verification for self-signed certificates
./mc-tool compare --insecure staging/bucket prod/bucket
```