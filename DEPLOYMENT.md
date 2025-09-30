# mc-tool Deployment Guide

## Available Binary Types

### 1. Standard Binary (`mc-tool`)
- **Use case**: Development and local testing
- **Dependencies**: Requires system libraries
- **Size**: ~14MB
- **Command**: `make build`

### 2. Static Binary (`mc-tool-static`)
- **Use case**: Production deployment on same OS type
- **Dependencies**: None (statically linked)
- **Size**: ~14MB (with debug symbols)
- **Command**: `make build-static`

### 3. Portable Binary (`mc-tool-portable`) ⭐ **RECOMMENDED**
- **Use case**: Maximum compatibility across environments
- **Dependencies**: None (statically linked + stripped)
- **Size**: ~9.5MB (smallest)
- **Command**: `make build-portable`

## Fixing "Required File Not Found" Errors

The portable binary (`mc-tool-portable`) eliminates common deployment issues:

### ❌ Common Errors with Dynamic Binaries
```bash
./mc-tool: error while loading shared libraries: libc.so.6: cannot open shared object file
./mc-tool: /lib/x86_64-linux-gnu/libc.so.6: version 'GLIBC_2.34' not found
```

### ✅ Solution: Use Static/Portable Binary
```bash
# Copy the portable binary to any Linux system
scp build/mc-tool-portable user@target-server:/usr/local/bin/mc-tool
chmod +x /usr/local/bin/mc-tool

# No additional dependencies needed!
mc-tool --help
```

## Deployment Options

### Option 1: Direct Copy (Recommended)
```bash
# Build portable version
make build-portable

# Copy to target system
scp build/mc-tool-portable user@server:/opt/mc-tool

# Make executable and test
ssh user@server "chmod +x /opt/mc-tool && /opt/mc-tool --help"
```

### Option 2: System-wide Installation
```bash
# Copy to system binary directory
sudo cp build/mc-tool-portable /usr/local/bin/mc-tool
sudo chmod +x /usr/local/bin/mc-tool

# Verify installation
mc-tool --help
```

### Option 3: Docker Container
```dockerfile
FROM scratch
COPY build/mc-tool-portable /mc-tool
ENTRYPOINT ["/mc-tool"]
```

## Compatibility Matrix

| Binary Type | Linux | Docker | Alpine | CentOS 7 | Ubuntu 18+ | RHEL 8+ |
|-------------|-------|--------|--------|----------|------------|---------|
| `mc-tool` | ⚠️ | ⚠️ | ❌ | ⚠️ | ⚠️ | ⚠️ |
| `mc-tool-static` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `mc-tool-portable` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

**Legend:**
- ✅ = Works without issues
- ⚠️ = May require compatible system libraries
- ❌ = Likely compatibility issues

## Verification Commands

After deployment, verify the binary works correctly:

```bash
# Check binary type
file /path/to/mc-tool
# Should show: "statically linked"

# Check for dependencies
ldd /path/to/mc-tool
# Should show: "not a dynamic executable"

# Test functionality
/path/to/mc-tool --help
/path/to/mc-tool compare --help
/path/to/mc-tool analyze --help
```

## Build All Versions

```bash
# Build all binary types
make build          # Standard binary
make build-static   # Static binary
make build-portable # Portable binary (recommended)
make build-all      # All platforms (Linux, macOS, Windows)
```

## Troubleshooting

### Issue: Permission Denied
```bash
chmod +x /path/to/mc-tool
```

### Issue: Command Not Found
```bash
# Add to PATH or use full path
export PATH=$PATH:/path/to/directory
# Or use absolute path
/full/path/to/mc-tool --help
```

### Issue: Architecture Mismatch
- The binaries are built for x86_64 (AMD64)
- For ARM64 systems, rebuild with: `GOARCH=arm64 make build-portable`

## Best Practices

1. **Use `mc-tool-portable`** for production deployments
2. **Verify binary integrity** after copying to target systems
3. **Test on target environment** before full deployment
4. **Use absolute paths** in scripts and automation
5. **Set proper permissions** (755 for executables)

The portable binary eliminates "required file not found" errors by including all dependencies statically!