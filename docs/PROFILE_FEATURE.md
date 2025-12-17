# MC-Tool Profile Feature - Implementation Summary

## Overview
Đã hoàn thành việc cải thiện output của `mc-tool profile` và tích hợp vào Web UI trong nhóm Tracing.

## Changes Made

### 1. CLI Output Improvements (`pkg/profile/mc21_profile.go`)

**Before**: Mỗi loại profile hiển thị 6-7 commands dài dòng
**After**: Mỗi loại chỉ hiển thị 2-3 commands quan trọng nhất

#### CPU Profile Commands (giảm từ 6 → 3):
```bash
# Web UI (recommended):
go tool pprof -http=:8080 <file>

# Top 10 CPU-intensive functions:
go tool pprof -top -nodecount=10 <file>

# Flamegraph SVG:
go tool pprof -svg <file> > cpu-flamegraph.svg
```

#### Memory Profile Commands (giảm từ 7 → 3):
```bash
# Web UI (recommended):
go tool pprof -http=:8080 <file>

# Top 10 memory allocations:
go tool pprof -top -nodecount=10 -alloc_space <file>

# Top 10 in-use memory:
go tool pprof -top -nodecount=10 -inuse_space <file>
```

#### Block/Mutex/Goroutine Profile Commands (giảm từ 2-3 → 2):
```bash
# <Type> - Web UI:
go tool pprof -http=:8080 <file>

# <Type> - Top 10:
go tool pprof -top -nodecount=10 <file>
```

#### Additional Commands (giảm từ 3-4 → 2):
```bash
# View cluster info:
cat <output_dir>/cluster.info

# List all profiles:
ls -lh <output_dir>/*.pprof
```

### 2. Web UI Integration

#### Backend API (`pkg/web/`)

**Files Modified**:
- `handlers/operations.go`: Added `ProfileRequest` struct and `HandleProfile()` handler
- `services/operations.go`: Added `ProfileCaptureOptions` and `RunProfileCapture()` service
- `server.go`: Added route `/api/operations/profile`

**API Endpoint**: `POST /api/operations/profile`

**Request Payload**:
```json
{
  "alias": "site1",
  "duration": "30s",
  "profileType": "cpu,mem"
}
```

**Response**:
```json
{
  "success": true,
  "alias": "site1",
  "duration": "30s",
  "outputDir": "/tmp/profile-20251125-130542",
  "files": {
    "cpu": ["profile-127.0.0.1:9000-cpu.pprof"],
    "mem": ["profile-127.0.0.1:9000-mem.pprof"],
    "other": []
  },
  "commands": {
    "cpu": [
      {
        "label": "Web UI (recommended)",
        "command": "go tool pprof -http=:8080 /tmp/profile-20251125-130542/profile-127.0.0.1:9000-cpu.pprof"
      },
      {
        "label": "Top 10 CPU-intensive functions",
        "command": "go tool pprof -top -nodecount=10 /tmp/profile-20251125-130542/profile-127.0.0.1:9000-cpu.pprof"
      },
      {
        "label": "Flamegraph SVG",
        "command": "go tool pprof -svg /tmp/profile-20251125-130542/profile-127.0.0.1:9000-cpu.pprof > cpu-flamegraph.svg"
      }
    ],
    "mem": [...],
    "other": [...]
  }
}
```

#### Frontend Components

**Files Created**:
- `web/src/components/operations/ProfileOperations.js`: React component for profiling UI

**Files Modified**:
- `web/src/utils/api.js`: Added `runProfileCapture()` function
- `web/src/App.js`: Added ProfileOperations import and route `/tracing/profiler`
- `web/src/components/Sidebar.js`: Added "Profiler" menu item under Tracing group

**UI Features**:
1. **Form Input**:
   - Alias selection (required)
   - Duration (default: 30s, max: 5m)
   - Profile types (default: cpu,mem)

2. **Results Display**:
   - Profile summary (output directory, duration, file counts)
   - Grouped commands by type (CPU, Memory, Other)
   - Copy-to-clipboard functionality for each command
   - Visual feedback when command is copied

3. **Navigation**:
   - Accessible via Tracing → Profiler in sidebar
   - URL: `/tracing/profiler`

### 3. Key Implementation Details

#### Command Generation Logic (`buildProfileCommands()`)
```go
// For each profile type, generate 2-3 essential commands
// CPU: Web UI + Top 10 + Flamegraph
// Memory: Web UI + Top 10 alloc + Top 10 inuse
// Other: Web UI + Top 10
```

#### React Component Structure
```jsx
ProfileOperations
├── Form (alias, duration, profileType)
├── Loading State
├── Error Display
└── Results
    ├── Summary Card
    └── Command Groups
        ├── CPU Commands
        ├── Memory Commands
        └── Other Commands
```

## Testing

### CLI Testing
```bash
./mc-tool profile site1 --duration 5s --type cpu
```

**Output**:
- ✅ Displays only 3 CPU commands (vs 6 before)
- ✅ Additional section shows only 2 commands (vs 3-4 before)
- ✅ Clean, readable output

### Web UI Testing
1. ✅ Build successful: `npm run build`
2. ✅ Server starts: `./mc-tool web --port 8080`
3. ✅ Route accessible: `http://localhost:8080/tracing/profiler`
4. ✅ API endpoint working: `POST /api/operations/profile`

## Files Summary

### Modified Files
1. `pkg/profile/mc21_profile.go` - Simplified command output
2. `pkg/web/handlers/operations.go` - Added profile handler
3. `pkg/web/services/operations.go` - Added profile service
4. `pkg/web/server.go` - Added profile route
5. `web/src/utils/api.js` - Added API function
6. `web/src/App.js` - Added route and import
7. `web/src/components/Sidebar.js` - Added menu item

### Created Files
1. `web/src/components/operations/ProfileOperations.js` - New component

## Usage Examples

### CLI
```bash
# CPU profile only
./mc-tool profile site1 --duration 30s --type cpu

# CPU + Memory
./mc-tool profile site1 --duration 1m --type cpu,mem

# All profiles
./mc-tool profile site1 --duration 2m --type cpu,mem,block,mutex,goroutines
```

### Web UI
1. Navigate to Tracing → Profiler
2. Enter alias (e.g., "site1")
3. Set duration (e.g., "30s")
4. Choose profile types (e.g., "cpu,mem")
5. Click "Start Profile"
6. Wait for results
7. Copy commands to run analysis

## Benefits

1. **CLI Output**: 
   - Giảm từ ~250 dòng xuống ~50 dòng cho typical profile
   - Chỉ hiển thị commands thực sự cần thiết
   - Dễ đọc, dễ copy-paste

2. **Web UI**:
   - Tích hợp vào Tracing group (logical grouping)
   - User-friendly interface
   - One-click copy commands
   - Visual feedback
   - Mobile-responsive

3. **Developer Experience**:
   - Clean code separation (CLI vs Web)
   - Reusable command generation logic
   - Consistent API design
   - Good error handling

## Future Enhancements (Optional)

1. Direct pprof visualization in Web UI (embed go tool pprof -http)
2. Profile comparison (before/after)
3. Scheduled profiling
4. Profile history/archive
5. Auto-detect performance issues from profile data
6. Integration with alerting system

## Notes

- mc21 binary required for profiling
- Profiles stored in /tmp/profile-YYYYMMDD-HHMMSS
- Web UI timeout: 5 minutes max duration
- CLI supports all mc21 profile types
- Commands are ready-to-run (no placeholders)
