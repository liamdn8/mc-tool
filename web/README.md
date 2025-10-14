# MC-Tool React Web UI

Dự án đã được nâng cấp từ web tĩnh sang React. Web UI hiện được xây dựng bằng React và vẫn được serve qua Go server như trước đây.

## Cấu trúc project

```
mc-tool/
├── web/                    # React source code (ngang hàng với pkg/)
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── pages/         # Page components  
│   │   ├── utils/         # Utilities (API, i18n, helpers)
│   │   ├── App.js         # Main App component
│   │   ├── index.js       # Entry point
│   │   └── styles.css     # CSS styles
│   ├── public/
│   │   └── index.html     # HTML template
│   ├── build/             # Built files (generated)
│   ├── package.json       # NPM dependencies
│   └── webpack.config.js  # Webpack configuration
├── pkg/web/
│   ├── server.go          # Go web server (cập nhật để serve React)
│   └── static/build/      # Built React files (embedded vào Go)
├── build-web.sh           # Script build React app
└── build-full.sh          # Script build toàn bộ (React + Go)
```

## Build và chạy

### 1. Build toàn bộ dự án (React + Go)

```bash
./build-full.sh
```

### 2. Chạy web server

```bash
./mc-tool web --port 8080
```

### 3. Chỉ build React (nếu cần)

```bash
./build-web.sh
```

## Development workflow

1. **Để phát triển React app:**
   ```bash
   cd web/
   npm install
   npm run dev     # Webpack development mode with watch
   ```

2. **Để test với Go server:**
   ```bash
   ./build-full.sh
   ./mc-tool web --port 8080
   ```

## Tính năng

- ✅ React-based UI với component architecture hiện đại
- ✅ Multi-language support (English/Vietnamese) 
- ✅ Responsive design
- ✅ Real-time data loading từ API
- ✅ Tất cả tính năng gốc được giữ nguyên:
  - Site management
  - Bucket overview  
  - Replication status
  - Consistency checking
  - Automated operations

## Build process

1. React app được build thành static files trong `web/build/`
2. Files được copy vào `pkg/web/static/build/` 
3. Go server embed files này qua `//go:embed static/build/*`
4. Server serve React app tại root URL `/`
5. API endpoints vẫn hoạt động như cũ tại `/api/*`

## Lưu ý

- Cần Node.js 18+ để build React app
- Build process tự động copy files từ React build vào Go embed directory
- Final binary vẫn là single file như trước, chỉ có web UI được nâng cấp lên React