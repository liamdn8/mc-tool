# MIME Type Fix - Static Files Serving

## 🐛 Problem

Browser refused to execute JavaScript file due to incorrect MIME type:

```
GET http://localhost:8080/static/app-new.js?v=3 
net::ERR_ABORTED 404 (Not Found)

Refused to execute script from 'http://localhost:8080/static/app-new.js?v=3' 
because its MIME type ('text/plain') is not executable, 
and strict MIME type checking is enabled.
```

### Root Causes

1. **Wrong MIME Type**: `http.FileServer` with `embed.FS` was returning `text/plain` instead of `application/javascript`
2. **Missing File**: `app-new.js` didn't exist (only `app.js`)

---

## ✅ Solutions Applied

### 1. Renamed JavaScript File

**File**: `pkg/web/static/app.js` → `app-new.js`

```bash
cd pkg/web/static
mv app.js app-new.js
```

**Why**: HTML was referencing `app-new.js?v=3` but file was named `app.js`

### 2. Custom Static Handler with MIME Types

**File**: `pkg/web/server.go`

**Before (Broken)**:
```go
mux.Handle("/static/", 
    http.StripPrefix("/static/", 
        http.FileServer(http.FS(staticFS))))
```

**Problem**: Default `http.FileServer` doesn't set correct MIME types for embedded files

**After (Fixed)**:
```go
// Custom static file handler with proper MIME types
staticHandler := http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))
mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    
    // Set correct MIME type based on file extension
    if strings.HasSuffix(path, ".js") {
        w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
    } else if strings.HasSuffix(path, ".css") {
        w.Header().Set("Content-Type", "text/css; charset=utf-8")
    } else if strings.HasSuffix(path, ".html") {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
    } else if strings.HasSuffix(path, ".json") {
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
    } else if strings.HasSuffix(path, ".png") {
        w.Header().Set("Content-Type", "image/png")
    } else if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
        w.Header().Set("Content-Type", "image/jpeg")
    } else if strings.HasSuffix(path, ".svg") {
        w.Header().Set("Content-Type", "image/svg+xml")
    }
    
    // Serve the file
    staticHandler.ServeHTTP(w, r)
})
```

**Benefits**:
- ✅ Correct MIME types for all file types
- ✅ Browser can execute JavaScript
- ✅ CSS loads properly
- ✅ Images display correctly

---

## 🧪 Verification

### Test JavaScript File

```bash
curl -sI http://localhost:8080/static/app-new.js | grep Content-Type
```

**Expected**:
```
Content-Type: application/javascript; charset=utf-8
```

**Result**: ✅ **PASS**

### Test CSS File

```bash
curl -sI http://localhost:8080/static/styles-new.css | grep Content-Type
```

**Expected**:
```
Content-Type: text/css; charset=utf-8
```

**Result**: ✅ **PASS**

### Test HTML File

```bash
curl -sI http://localhost:8080/ | grep Content-Type
```

**Expected**:
```
Content-Type: text/html; charset=utf-8
```

**Result**: ✅ **PASS**

---

## 📊 MIME Type Mapping

| File Extension | MIME Type | Usage |
|----------------|-----------|-------|
| `.js` | `application/javascript; charset=utf-8` | JavaScript files |
| `.css` | `text/css; charset=utf-8` | Stylesheets |
| `.html` | `text/html; charset=utf-8` | HTML pages |
| `.json` | `application/json; charset=utf-8` | JSON data |
| `.png` | `image/png` | PNG images |
| `.jpg`, `.jpeg` | `image/jpeg` | JPEG images |
| `.svg` | `image/svg+xml` | SVG images |

---

## 🔍 Technical Deep Dive

### Why embed.FS Needs Custom MIME Handling

**Problem with Go's `embed.FS`**:

```go
//go:embed static/*
var staticFiles embed.FS
```

When using `http.FileServer(http.FS(staticFS))`:
- Go's `http.DetectContentType()` reads first 512 bytes
- For text files, often defaults to `text/plain`
- Browsers reject `text/plain` for JavaScript (security)

**Solution Pattern**:

1. **Wrap FileServer**: Create handler that wraps original FileServer
2. **Set Headers First**: Add MIME type headers before serving
3. **Delegate to FileServer**: Let original handler serve the file

```go
staticHandler := http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))

mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
    // 1. Set MIME type based on extension
    w.Header().Set("Content-Type", detectMIME(r.URL.Path))
    
    // 2. Let FileServer handle the actual file serving
    staticHandler.ServeHTTP(w, r)
})
```

---

## 🛡️ Browser Security Context

### Why Browsers Are Strict About MIME Types

**MIME Type Sniffing Attack Prevention**:

Modern browsers enforce strict MIME type checking to prevent:

```javascript
// Attacker uploads malicious.jpg (actually JavaScript)
<script src="uploads/malicious.jpg"></script>

// Old browsers: Execute as JavaScript (VULNERABLE!)
// Modern browsers: Check MIME type
//   - If image/jpeg: ✓ Safe, don't execute
//   - If application/javascript: ✗ Block, wrong type
```

**Our Error**:
```
MIME type ('text/plain') is not executable
```

Browser refused because:
- Expected: `application/javascript`
- Received: `text/plain`
- Action: **BLOCK** execution for security

---

## 📁 Static Files Structure

```
pkg/web/static/
├── app-new.js       ← Main JavaScript (renamed from app.js)
├── app.js.bk        ← Backup
├── index-new.html   ← Main HTML
├── index.html       ← Old version
├── styles-new.css   ← Main CSS
└── styles.css       ← Old version
```

**Active Files** (referenced in HTML):
- `index-new.html` (served at `/`)
- `app-new.js?v=3` (loaded via `<script>`)
- `styles-new.css?v=3` (loaded via `<link>`)

---

## 🔄 Deployment Process

### Full Rebuild Required

```bash
# 1. Stop server
pkill -f 'mc-tool web'

# 2. Clean build (embeds new static files)
cd /home/liamdn/mc-tool
rm -f mc-tool
go clean
go build -o mc-tool .

# 3. Start server
./mc-tool web --port 8080 > /tmp/mc-tool.log 2>&1 &

# 4. Verify MIME types
sleep 3
curl -sI http://localhost:8080/static/app-new.js | grep Content-Type
curl -sI http://localhost:8080/static/styles-new.css | grep Content-Type

# 5. Open browser (hard refresh)
# Ctrl+Shift+R to clear cache
```

**Why rebuild is required**:
- Static files are embedded with `//go:embed static/*`
- Changes to static files require recompilation
- Embedded files are part of the binary

---

## 🐛 Debugging MIME Type Issues

### Check Response Headers

```bash
# Full headers
curl -I http://localhost:8080/static/app-new.js

# Just Content-Type
curl -sI http://localhost:8080/static/app-new.js | grep Content-Type
```

### Browser DevTools

1. **Open DevTools**: F12
2. **Network Tab**: Reload page
3. **Click on resource**: app-new.js
4. **Headers Tab**: Check `Content-Type`

**Should see**:
```
Content-Type: application/javascript; charset=utf-8
```

**If you see**:
```
Content-Type: text/plain
```
→ MIME type handler not working

### Common Issues

| Issue | Symptom | Solution |
|-------|---------|----------|
| File not found | 404 Not Found | Check filename, rebuild binary |
| Wrong MIME | `text/plain` for .js | Add custom handler |
| Cached old version | Old content loads | Hard refresh (Ctrl+Shift+R) |
| Binary not updated | Changes don't apply | Rebuild with `go build` |

---

## 💡 Lessons Learned

### 1. File Naming Consistency

**Bad**:
```
HTML references: app-new.js
Actual file: app.js
Result: 404 Not Found ❌
```

**Good**:
```
HTML references: app-new.js
Actual file: app-new.js
Result: ✓ File loads ✅
```

### 2. Embedded Files Require Rebuild

```go
//go:embed static/*
var staticFiles embed.FS
```

**Every change to static files needs**:
```bash
go build -o mc-tool .
```

### 3. MIME Types Matter for Security

- Browsers enforce strict MIME type checking
- `text/plain` JavaScript won't execute
- Always set correct `Content-Type` headers

---

## ✅ Final Status

**All MIME Types Correct**:
- ✅ JavaScript: `application/javascript; charset=utf-8`
- ✅ CSS: `text/css; charset=utf-8`
- ✅ HTML: `text/html; charset=utf-8`

**Files Present**:
- ✅ `app-new.js` exists (renamed from app.js)
- ✅ `styles-new.css` exists
- ✅ `index-new.html` exists

**Server Status**:
- ✅ Custom MIME handler implemented
- ✅ Static files serving correctly
- ✅ Browser can execute JavaScript

**Date**: October 13, 2025  
**Status**: ✅ **RESOLVED**  
**Impact**: Critical - JavaScript execution now works
