# MC-Tool Web UI Quick Reference

## Start Web UI

```bash
# Default port (8080)
mc-tool web

# Custom port
mc-tool web --port 3000
```

## Access

```
http://localhost:8080
```

## Language Selection

Click the dropdown in the top-right corner:
- 🇬🇧 English
- 🇻🇳 Tiếng Việt

## Pages

### 📊 Dashboard
- View MC client status
- See configured aliases
- Monitor active jobs

### 🔄 Compare
Compare two MinIO buckets:
1. Enter source path: `alias1/bucket1/path`
2. Enter destination path: `alias2/bucket2/path`
3. Check "Recursive" if needed
4. Click "Start Comparison"

### 📈 Analyze
Analyze bucket contents:
1. Select alias
2. Select bucket (auto-loaded)
3. Optional: enter prefix
4. Click "Start Analysis"

### 🔍 Profile
Profile MinIO server:
1. Select alias
2. Choose profile type:
   - Heap (memory)
   - CPU
   - Goroutine
   - Allocs
   - Block
   - Mutex
3. Set duration (e.g., `30s`, `1m`, `5m`)
4. Optional: Enable leak detection
5. Click "Start Profiling"

### ✅ Checklist
Validate bucket configuration:
1. Select alias
2. Select bucket
3. Click "Run Checklist"

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/api/health
```

### List Aliases
```bash
curl http://localhost:8080/api/aliases
```

### List Buckets
```bash
curl http://localhost:8080/api/buckets?alias=minio1
```

### Start Comparison
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

## Troubleshooting

### Port Already in Use
```bash
mc-tool web --port 8090
```

### MC Not Configured
```bash
mc alias list
```

### Jobs Not Running
Ensure `mc-tool` is in PATH

## Tips

- Jobs run in the background - you can track progress in the modal
- Click outside the job modal to close it
- Use the refresh button to reload dashboard data
- All forms show real-time validation
- Results are displayed in an easy-to-read format

## Support

For more details, see:
- [Full Web UI Documentation](docs/WEB_UI.md)
- [Main README](README.md)
