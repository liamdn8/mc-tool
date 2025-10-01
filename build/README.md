# MC-Tool Build Artifacts

This directory contains pre-built binaries for the MC-Tool application.

## Recommended Binaries

### For Linux Servers/Containers (Recommended)
- **`mc-tool-portable`** - Smallest static binary (9.7MB), stripped for production use
- **`mc-tool-static`** - Static binary with debug info (14MB), good for development/debugging
- **`mc-tool-linux-amd64-static`** - Static binary for x86_64 Linux systems
- **`mc-tool-linux-arm64-static`** - Static binary for ARM64 Linux systems (Raspberry Pi, ARM servers)

### For Desktop/Development
- **`mc-tool-darwin-amd64`** - macOS Intel (x86_64)
- **`mc-tool-darwin-arm64`** - macOS Apple Silicon (M1/M2/M3)
- **`mc-tool-windows-amd64.exe`** - Windows 64-bit

## Binary Types

### Static Binaries (Linux)
- **No dependencies required** - can run on any Linux system
- **Container-friendly** - perfect for Docker containers
- **Portable** - copy and run anywhere

### Regular Binaries (macOS/Windows)
- Built with Go's standard build process
- May require system libraries (usually available by default)

## Usage

1. Choose the appropriate binary for your platform
2. Make executable (Linux/macOS): `chmod +x mc-tool-*`
3. Run directly: `./mc-tool-linux-amd64-static version`

## Build Information

All binaries include build-time information:
- Version: Current git tag or "dev"
- Commit: Git commit hash
- Build Time: When the binary was built

Check with: `./mc-tool-* version`

## Deployment Recommendations

### Production Servers
Use `mc-tool-portable` for:
- Minimal size (9.7MB)
- Maximum compatibility
- No debugging symbols (secure)

### Development/Debugging
Use `mc-tool-static` for:
- Debug symbols included
- Better error traces
- Development testing

### Containers
```dockerfile
COPY mc-tool-portable /usr/local/bin/mc-tool
RUN chmod +x /usr/local/bin/mc-tool
```

### Direct Download
```bash
# Copy to system PATH
sudo cp mc-tool-portable /usr/local/bin/mc-tool
sudo chmod +x /usr/local/bin/mc-tool
```