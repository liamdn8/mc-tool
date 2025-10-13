# MC-Tool Web UI

A modern, bilingual web interface for MC-Tool that makes MinIO operations accessible to operators who aren't familiar with command-line tools.

## Features

### 🌐 Bilingual Support
- **English** and **Vietnamese (Tiếng Việt)** interfaces
- Easy language switching
- Fully translated UI elements

### 📊 Dashboard
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

## API Endpoints

The web UI exposes the following REST API endpoints:

- `GET /api/health` - Health check
- `GET /api/aliases` - List configured aliases
- `GET /api/buckets?alias=<name>` - List buckets for alias
- `POST /api/compare` - Start comparison job
- `POST /api/analyze` - Start analysis job
- `POST /api/profile` - Start profiling job
- `POST /api/checklist` - Start checklist job
- `GET /api/jobs/<id>` - Get job status
- `GET /api/mc-config` - Check MC configuration

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
