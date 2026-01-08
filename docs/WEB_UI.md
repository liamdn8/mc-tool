# MC-Tool Web UI

A modern, bilingual web interface for MC-Tool that makes MinIO operations accessible to operators who aren't familiar with command-line tools.

## Features

### 🌐 Bilingual Support
- **English** and **Vietnamese (Tiếng Việt)** interfaces
- Easy language switching
- Fully translated UI elements

### � Offline Capability
- Works completely offline without internet connection
- Local Lucid Icons integration
- No CDN dependencies
- Embedded static assets

### �📊 Dashboard
- Real-time MC status monitoring
- Configured aliases overview
- Active jobs tracking
- Quick access to all features

### 🔄 Bucket Comparison
- Visual interface for comparing buckets
- Real-time job progress tracking
- Detailed comparison results
- Support for recursive comparisons

### 📈 Bucket Analysis
- Interactive bucket selection
- Object distribution analysis
- Size statistics
- File type breakdown

### 🔍 Performance Profiling
- Memory leak detection
- CPU profiling
- Goroutine monitoring
- Heap analysis
- Custom duration and thresholds
- Support for mc and mc-2021 versions

### ✅ Configuration Checklist
- Bucket configuration validation
- Event notification status
- Lifecycle policy review
- Security settings check

### 🏗️ Infrastructure Validation
- **Multi-cluster Kubernetes validation**
- Compare namespace configurations across clusters
- Detect configuration drift
- **Comprehensive resource tracking**: All resources displayed (matched, mismatched, not found, extra)
- **Extra resource detection**: Identifies unauthorized or missing resources
- **Advanced diff viewer**: ArgoCD-style side-by-side comparison
- **Myers algorithm**: Accurate content-based diffs (same as Git)
- **Interactive badges**: Click to view diffs for Match, Mismatch, or Extra resources
- **Full/Differences toggle**: Focus on changes or view complete files
- **Collapsible hunks**: Expandable change sections with context
- **Resource status indicators**: Color-coded badges for quick assessment
- **Search and filter**: Find specific resources quickly
- **Navigation panel**: Jump to resource types instantly

## Getting Started

### Start Web UI

```bash
# Start on default port 8080
mc-tool web

# Start on custom port
mc-tool web --port 3000
```

### Access the Interface

Open your browser and navigate to:
```
http://localhost:8080
```

## Usage Guide

### 1. Dashboard
- View your configured MinIO aliases
- Check MC client status
- Monitor active background jobs

### 2. Compare Buckets
1. Navigate to the "Compare" page
2. Enter source path (e.g., `minio1/bucket1/folder`)
3. Enter destination path (e.g., `minio2/bucket2/folder`)
4. Choose recursive option if needed
5. Click "Start Comparison"
6. View results in real-time

### 3. Analyze Bucket
1. Go to the "Analyze" page
2. Select an alias from the dropdown
3. Select a bucket (auto-loaded)
4. Optionally specify a prefix
5. Click "Start Analysis"
6. Review object distribution and statistics

### 4. Profile Server
1. Open the "Profile" page
2. Select target alias
3. Choose profile type:
   - **Heap**: Memory usage analysis
   - **CPU**: Performance profiling
   - **Goroutine**: Concurrency analysis
   - **Allocs**: Allocation tracking
   - **Block**: Blocking operations
   - **Mutex**: Lock contention
4. Set duration (e.g., `30s`, `1m`, `5m`)
5. Enable memory leak detection (optional)
6. Configure monitoring interval and threshold
7. Start profiling

### 5. Run Checklist
1. Navigate to "Checklist"
2. Select alias and bucket
3. Run comprehensive configuration check
4. Review results

### 6. Infrastructure Validation
1. Go to "Infrastructure Validation" page
2. **Select Baseline:**
   - Choose VIM (Kubernetes cluster)
   - Choose Namespace
3. **Add Targets:**
   - Click "Add Target"
   - Select VIM and Namespace for comparison
   - Can add multiple targets
4. **Run Validation:**
   - Click "Validate Infrastructure"
   - Monitor real-time progress
5. **Review Results:**
   - **Overview Cards**: Statistics (Total, Match, Mismatch, Not Found, Extra)
   - **Navigation Panel**: Click resource types to jump to sections
   - **Resource Tables**: Grouped by type with status badges
   - **Status Indicators**:
     - 🟢 **Match** (green): Configurations identical
     - 🟢 **Configured** (green): Baseline reference
     - 🟡 **Mismatch** (warning): Configuration differs
     - 🟡 **Extra** (warning): Resource only in target
     - 🟡 **Not Found** (warning): Resource only in baseline
6. **View Diffs:**
   - Click any **Match**, **Mismatch**, or **Extra** badge
   - ArgoCD-style diff viewer opens:
     - Side-by-side comparison
     - VIM/namespace labels
     - Toggle "Show only differences" ↔ "Show full"
     - Collapsible change hunks
     - Copy buttons for baseline and target
   - Close diff viewer to return

### Infrastructure Validation Features

**Enhanced Resource Display:**
- ✅ Shows ALL resources including perfectly matched ones
- ✅ Extra resource detection (resources in target not in baseline)
- ✅ Clear baseline status ("Configured" instead of "-")
- ✅ Interactive table with search, filter, and pagination
- ✅ Status count badges in table headers

**Advanced Diff Viewer:**
- ✅ Myers diff algorithm for accurate comparison
- ✅ ArgoCD-style side-by-side view
- ✅ Collapsible hunks with context lines
- ✅ Full file view vs differences-only mode
- ✅ Syntax highlighting for YAML
- ✅ Separate copy buttons for each side
- ✅ VIM/namespace labels in header

**Navigation & UX:**
- ✅ Sidebar navigation with auto-scroll
- ✅ Contents panel integration
- ✅ Mobile-responsive design
- ✅ Real-time validation progress

## API Endpoints

The web UI exposes the following REST API endpoints:

**General:**
- `GET /api/health` - Health check
- `GET /api/mc-config` - Check MC configuration

**MinIO Operations:**
- `GET /api/aliases` - List configured aliases
- `GET /api/buckets?alias=<name>` - List buckets for alias
- `POST /api/compare` - Start comparison job
- `POST /api/analyze` - Start analysis job
- `POST /api/profile` - Start profiling job
- `POST /api/checklist` - Start checklist job
- `GET /api/jobs/<id>` - Get job status

**Infrastructure Validation:**
- `GET /api/validate/infrastructure/vims` - List available Kubernetes clusters (VIMs)
- `GET /api/validate/infrastructure/namespaces?vim=<name>` - List namespaces in VIM
- `POST /api/validate/infrastructure` - Start validation job
  ```json
  {
    "baseline": "site1/app-staging",
    "targets": ["site2/app-staging", "site3/app-dev"]
  }
  ```
- `GET /api/validate/infrastructure/diff?baseline=<site/ns>&target=<site/ns>&resource_type=<type>&resource_name=<name>` - Get resource diff

## Language Support

### Switching Languages

Click the language selector in the top-right corner:
- 🇬🇧 English
- 🇻🇳 Tiếng Việt

### Supported Translations

All UI elements are translated, including:
- Navigation menus
- Form labels and placeholders
- Button text
- Status messages
- Help text
- Error messages

## Architecture

### Frontend
- Pure HTML5, CSS3, and vanilla JavaScript
- No external dependencies
- Responsive design
- Real-time job updates via polling

### Backend
- Go HTTP server with embedded static files
- RESTful API
- Background job management
- Integration with mc-tool CLI commands

### Job Management
- Asynchronous job execution
- Real-time status updates
- Progress tracking
- Output capture
- Error handling

## Development

### File Structure

```
pkg/web/
├── server.go           # Web server and API handlers
└── static/
    ├── index.html      # Main HTML structure
    ├── styles.css      # Responsive CSS styling
    └── app.js          # Frontend logic and i18n
```

### Building with Web UI

```bash
# Regular build
go build -o mc-tool .

# Portable build
make build-portable

# All platforms
make build-all
```

### Customization

#### Adding New Languages

Edit `pkg/web/static/app.js` and add translations to the `translations` object:

```javascript
const translations = {
    en: { /* English translations */ },
    vi: { /* Vietnamese translations */ },
    ja: { /* Add Japanese */ }
};
```

#### Custom Styling

Modify CSS variables in `pkg/web/static/styles.css`:

```css
:root {
    --primary-color: #2563eb;  /* Change primary color */
    --sidebar-bg: #1e293b;     /* Change sidebar background */
    /* ... other variables */
}
```

## Browser Compatibility

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Opera 76+

## Security Considerations

- The web UI runs locally by default
- No authentication required for localhost access
- For production deployment:
  - Use reverse proxy with authentication
  - Enable HTTPS
  - Restrict network access
  - Configure CORS policies

## Troubleshooting

### Port Already in Use

```bash
# Use different port
mc-tool web --port 8090
```

### MC Not Configured

Ensure MinIO Client is properly configured:

```bash
mc alias list
```

### Jobs Not Completing

Check that `mc-tool` binary is in PATH and accessible from the web server.

## Examples

### Start Web UI

```bash
# Default port
./mc-tool web

# Custom port
./mc-tool web --port 3000
```

### Compare via API

```bash
curl -X POST http://localhost:8080/api/compare \
  -H "Content-Type: application/json" \
  -d '{
    "source": "minio1/bucket1",
    "destination": "minio2/bucket2",
    "recursive": true
  }'
```

### Check Job Status

```bash
curl http://localhost:8080/api/jobs/compare-1696956789
```

## Contributing

Contributions are welcome! Areas for improvement:
- Additional language support
- Enhanced visualizations
- More profile types
- Advanced filtering options
- Export functionality

## License

Same as mc-tool main project.
