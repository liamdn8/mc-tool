# MC-Tool Test Environment Setup

# MC-Tool Test Setup & Operations Page Documentation

## 🎯 Overview

This document describes the completed test environment setup and Operations page functionality for MC-Tool. The implementation includes bucket comparison features, configuration checklists, and automated test environment setup.

## ✅ Completed Features

### 1. Compare Operations
- **Bucket/Path Comparison**: Compare content between two MinIO aliases
- **Visual Results Display**: Command-line style output with clear formatting
- **Dynamic Bucket Selection**: Auto-populated dropdowns for available buckets
- **Path Suggestions**: Autocomplete for paths within selected buckets
- **Site Replication Independence**: Works without site replication configured

### 2. Configuration Checklist
- **Environment Variables**: Check configured environment variables
- **Event Configuration**: Verify event settings
- **Bucket Events**: Validate bucket event configurations
- **Object Lifecycle**: Check lifecycle policies

### 3. Test Environment Setup
- **Automated Scripts**: Complete setup automation
- **Docker Integration**: MinIO container management
- **Alias Configuration**: Automatic MinIO client setup
- **Test Data Generation**: Comprehensive test scenarios

## 🔧 Fixed Issues

### Backend Issues Fixed:
1. **CompareBuckets Service**: 
   - Changed from `mc diff` to `./mc-tool compare`
   - Proper parsing of command-line output format
   - Accurate result categorization (source-only, dest-only, different)
   - Comprehensive summary statistics

2. **API Integration**:
   - Fixed route conflicts in server.go
   - Corrected JSON response format
   - Added proper error handling for command execution

### Frontend Issues Fixed:
1. **Results Display**:
   - Removed excessive color coding
   - Implemented command-line style display
   - Added monospace font for technical output
   - Clear legend for result interpretation

2. **User Experience**:
   - Simplified color scheme (green +, red -, yellow !)
   - Better visual hierarchy
   - Consistent with command-line output format

## 🚀 Usage

### Quick Start
```bash
# Complete setup (all-in-one)
./setup-complete-test-env.sh

# Or step-by-step:
./setup-minio-test.sh      # Setup MinIO containers
./setup-aliases.sh         # Configure mc aliases  
./setup-test-data.sh       # Create test scenarios
./mc-tool web --port 8080  # Start web UI
```

### Web Interface
1. Open browser: http://localhost:8080
2. Navigate to Operations page
3. Select comparison aliases and buckets
4. View results in command-line style format

### Expected Compare Output Format
```
Comparison Results:
==================
+ config/settings.json - Missing in target
+ docs/manual.pdf - Missing in target
+ only-on-first/readme.txt - Missing in target
- logs/system.log - Missing in source
- only-on-second/backup.txt - Missing in source
- backup/data.sql - Missing in source

Summary:
  Identical: 0
  Different: 0
  Missing in source: 3
  Missing in target: 3
  Total compared: 6
```

## 📁 Test Data Scenarios

The setup creates 7 different test scenarios:

1. **Files only in source** (bucket-1): config/, docs/, only-on-first/
2. **Files only in dest** (bucket-1): logs/, only-on-second/, backup/
3. **Different content** (bucket-2): shared files with different sizes
4. **Identical files** (bucket-3): same content on both sides
5. **Complex folder structure** (bucket-4): nested directories
6. **Large files** (bucket-5): files > 1MB for performance testing
7. **Version differences** (bucket-6): same files, different timestamps

## 🔍 Technical Implementation

### Backend Service (operations.go)
```go
func (os *OperationsService) CompareBuckets(sourceAlias, destAlias, path string) (map[string]interface{}, error) {
    // Uses ./mc-tool compare --insecure instead of mc diff
    cmd = exec.Command("./mc-tool", "compare", "--insecure", source, dest)
    
    // Parses command-line output format:
    // + filename - Missing in target
    // - filename - Missing in source  
    // ! filename - Content differs
}
```

### Frontend Display (OperationsPage.js)
```javascript
const renderCompareResults = () => {
    // Command-line style display with monospace font
    // Color coding: green (+), red (-), yellow (!)
    // Comprehensive summary section
}
```

## 🛠️ Script Details

### setup-complete-test-env.sh
- **Prerequisites check**: Verifies mc, docker, go, npm
- **Build process**: Compiles web assets and Go binary
- **MinIO setup**: Creates 6 container instances
- **Alias configuration**: Auto-configures mc client
- **Test data**: Generates 7 test scenarios
- **Health checks**: Validates all components

### setup-test-data.sh
- **Dynamic alias detection**: Works with site1-6 or custom aliases
- **Comprehensive scenarios**: 7 different test cases
- **Auto-cleanup**: Removes existing buckets before creation
- **Error handling**: Checks connectivity before operations
- **Progress feedback**: Colored output with status updates

### setup-aliases.sh
- **Auto-detection**: Finds existing MinIO instances
- **Docker integration**: Works with setup-minio-test.sh containers
- **Flexible configuration**: Manual, auto-detect, or Docker modes
- **Connectivity testing**: Validates aliases after configuration

## 📊 Performance & Reliability

### Improvements Made:
1. **Better error handling**: Graceful degradation on failures
2. **Timeout handling**: Prevents hanging operations
3. **Resource cleanup**: Proper cleanup of Docker containers
4. **Dynamic configuration**: Adapts to different environments
5. **Progress tracking**: Clear feedback during setup

### Test Results:
- ✅ Web UI loads successfully (228 KiB bundle)
- ✅ Compare operations execute correctly
- ✅ Results display matches command-line output
- ✅ Works with/without site replication
- ✅ Dynamic bucket/path suggestions functional

## 🎨 UI/UX Improvements

### Before:
- Excessive color blocks (yellow, blue, red backgrounds)
- Separate sections for each result type
- No connection to command-line experience
- Confusing visual hierarchy

### After:
- Clean monospace display mimicking terminal
- Unified results section with symbolic prefixes
- Clear legend explaining symbols
- Consistent with existing command-line tools

## 🔮 Future Enhancements

1. **Real-time comparison**: Live updates during large comparisons
2. **Export functionality**: Save results to file
3. **Comparison history**: Track previous comparisons
4. **Advanced filtering**: Filter results by file type/size
5. **Batch operations**: Compare multiple bucket pairs

## 📝 Notes

- Scripts are designed to be idempotent (can run multiple times safely)
- All test data is automatically generated, no manual file creation needed
- Web UI gracefully handles missing MinIO instances
- Compare functionality works independently of site replication status
- Command-line and web interface produce identical comparison logic

---

**Status**: ✅ All requested features implemented and tested
**Last Updated**: October 15, 2025
**Version**: 1.0 - Production Ready

## 🚀 Quick Start

Run the complete setup script:
```bash
./setup-complete-test-env.sh
```

This script will:
1. Check prerequisites (mc, docker)
2. Set up MinIO instances (Docker or manual)
3. Configure aliases
4. Create test data
5. Start the web UI

## 📁 Scripts Overview

### `setup-complete-test-env.sh` (Recommended)
**Complete automated setup**
- Sets up everything needed for testing
- Offers multiple setup options (Docker, manual, skip)
- Creates test data and starts web UI
- Includes cleanup on exit

### `setup-aliases.sh`
**MinIO client alias configuration**
- Configures mc aliases for site4, site5, site6
- Tests connectivity
- Customizable endpoints and credentials

### `setup-test-data.sh`  
**Test data creation**
- Creates test buckets and files
- Sets up various comparison scenarios
- Requires aliases to be configured first

## 🔧 Prerequisites

1. **MinIO Client (mc)**
   ```bash
   wget https://dl.min.io/client/mc/release/linux-amd64/mc
   chmod +x mc
   sudo mv mc /usr/local/bin/
   ```

2. **Docker (optional, for automatic MinIO setup)**
   - For Docker-based MinIO instances
   - Not required if you have existing MinIO servers

3. **MC-Tool built**
   ```bash
   go build -o mc-tool .
   # or
   ./build-web.sh
   ```

## 🎯 Test Scenarios Created

The test data setup creates these comparison scenarios:

### 📊 **Scenario 1: Files only on source**
- `test-bucket-1/only-on-site4.txt` (only on site4)
- `test-bucket-1/docs/` folder (only on site4)

### 📊 **Scenario 2: Files only on destination**  
- `test-bucket-1/only-on-site5.txt` (only on site5)
- `test-bucket-1/backup/` folder (only on site5)
- `test-bucket-2/unique-folder/` (only on site5)

### 📊 **Scenario 3: Different content**
- `shared-bucket/data/large-file.txt` (different content on each site)
- `shared-bucket/config/app.json` (different versions)

### 📊 **Scenario 4: Identical files**
- `shared-bucket/shared/common.txt` (same on all sites)

### 📊 **Scenario 5: Folder structures**
- `test-bucket-2/data/2024/` (nested folders)
- `test-bucket-2/site-specific/` (site-specific folders)

## 🌐 Testing Compare Operations

1. **Start the web UI:**
   ```bash
   ./mc-tool web --port 8080
   ```

2. **Open browser:** http://localhost:8080

3. **Navigate to Operations page**

4. **Try these comparisons:**
   - **site4 vs site5** in `test-bucket-1` (files only in source/dest)
   - **site4 vs site6** in `shared-bucket` (different content)  
   - **site5 vs site6** in `test-bucket-2` (folder differences)

## 🔧 Manual Setup

If you prefer manual setup:

1. **Configure aliases:**
   ```bash
   mc config host add site4 http://localhost:9004 admin password123
   mc config host add site5 http://localhost:9005 admin password123
   mc config host add site6 http://localhost:9006 admin password123
   ```

2. **Create test data:**
   ```bash
   ./setup-test-data.sh
   ```

3. **Start web UI:**
   ```bash
   ./mc-tool web --port 8080
   ```

## 🐳 Docker Setup

For Docker-based MinIO instances:

```bash
# Site 4
docker run -d --name minio-site4 \
  -p 9004:9000 -p 9014:9001 \
  -e "MINIO_ROOT_USER=admin" \
  -e "MINIO_ROOT_PASSWORD=password123" \
  quay.io/minio/minio server /data --console-address ":9001"

# Site 5  
docker run -d --name minio-site5 \
  -p 9005:9000 -p 9015:9001 \
  -e "MINIO_ROOT_USER=admin" \
  -e "MINIO_ROOT_PASSWORD=password123" \
  quay.io/minio/minio server /data --console-address ":9001"

# Site 6
docker run -d --name minio-site6 \
  -p 9006:9000 -p 9016:9001 \
  -e "MINIO_ROOT_USER=admin" \
  -e "MINIO_ROOT_PASSWORD=password123" \
  quay.io/minio/minio server /data --console-address ":9001"
```

## 🧹 Cleanup

**Manual cleanup:**
```bash
# Remove Docker containers
docker stop minio-site4 minio-site5 minio-site6
docker rm minio-site4 minio-site5 minio-site6

# Remove aliases
mc config host remove site4
mc config host remove site5  
mc config host remove site6

# Remove test data (optional)
rm -rf test-data/
```

**Automatic cleanup:** The complete setup script includes cleanup on exit (Ctrl+C).

## 🎛️ Operations Page Features

The updated Operations page now supports:

✅ **Works without site replication**
- Compare operations work with any configured aliases
- Checklist operations work independently
- Warning shown for replication-specific operations

✅ **Improved Compare interface**
- Source/destination alias selection
- Bucket dropdown (populated from source alias)
- Path suggestions with autocomplete
- Smart validation

✅ **Better result display**
- Color-coded sections (source-only, dest-only, different)
- Detailed diff information
- Summary statistics

## 🤝 Support

If you encounter issues:
1. Check that aliases are configured: `mc config host list`
2. Test connectivity: `mc ping site4`
3. Verify buckets exist: `mc ls site4`
4. Check web UI console for errors

Happy testing! 🎉