# MinIO Site Replication Management Tool

## 🎯 Mục đích

Tool này được thiết kế để **quản lý và maintain các MinIO instances** trong một Site Replication setup. Giúp bạn:

- ✅ Giám sát trạng thái site replication của tất cả MinIO aliases
- ✅ So sánh và phát hiện configuration drift giữa các sites
- ✅ Tự động đồng bộ cấu hình (policies, lifecycle rules, etc.)
- ✅ Health monitoring real-time
- ✅ Consistency checking

## 🚀 Quick Start

### 1. Cài đặt MinIO Client (mc)

```bash
# Download mc
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc
sudo mv mc /usr/local/bin/

# Verify
mc --version
```

### 2. Cấu hình MinIO Aliases

```bash
# Thêm các MinIO instances vào mc config
mc alias set site1 http://localhost:9001 minioadmin minioadmin
mc alias set site2 http://localhost:9002 minioadmin minioadmin

# Kiểm tra
mc alias list
```

### 3. Cấu hình Site Replication (Tùy chọn)

Nếu muốn enable site replication giữa các MinIO instances:

```bash
# Từ site1, add site2 vào replication group
mc admin replicate add site1 site2

# Kiểm tra status
mc admin replicate info site1
```

### 4. Chạy Tool

```bash
# Build
go build -ldflags="-s -w" -o mc-tool main.go

# Run
./mc-tool web

# Mở browser
# http://localhost:8080
```

## 📊 Giao diện

### 1. Overview Dashboard

Hiển thị:
- **Replication Group Info**: Thông tin tổng quan
- **Total Sites**: Số lượng MinIO aliases
- **Synced Buckets**: Số buckets đã đồng bộ
- **Total Objects**: Tổng số objects
- **Health Status**: Tình trạng tổng thể

### 2. Configured MinIO Aliases

Danh sách tất cả MinIO aliases với:
- **Alias name**: Tên alias trong mc config
- **Endpoint**: URL của MinIO instance
- **Replication Status**: 
  - ✅ **Configured** (màu xanh): Site replication đã được cấu hình
  - ⚠️ **Disabled** (màu vàng): Site replication bị tắt
  - ❌ **Not Configured** (màu đỏ): Chưa cấu hình site replication
- **Health**: Online/Offline status
- **Buckets**: Số lượng buckets
- **Objects**: Số lượng objects
- **Servers**: Số lượng server nodes (nếu là distributed setup)
- **Deployment ID**: ID deployment (nếu có)
- **Site Name**: Tên site trong replication group (nếu có)

### 3. Replication Status

Monitor chi tiết từng site:
- Replicated buckets
- Pending objects
- Failed objects
- Last sync time
- Health status

### 4. Consistency Check

So sánh cấu hình giữa các sites:
- **Bucket Policies**: Kiểm tra xem policies có giống nhau không
- **Lifecycle Rules**: So sánh ILM policies
- **Versioning**: Kiểm tra versioning status
- Visual diff để dễ nhận biết sự khác biệt

### 5. Automated Operations

4 operations tự động:
1. **Sync Bucket Policies**: Đồng bộ bucket policies
2. **Sync Lifecycle Policies**: Đồng bộ ILM rules
3. **Validate Consistency**: Kiểm tra tính nhất quán
4. **Health Check**: Verify tất cả sites đều healthy

## 🎨 Trạng thái Replication

Tool tự động phát hiện và hiển thị 3 trạng thái:

### ✅ Configured (Xanh)
- Site replication đã được cấu hình
- Sites đang trong một replication group
- Có thể sync configurations tự động

### ⚠️ Disabled (Vàng)
- Site replication có cấu hình nhưng đang tắt
- Cần enable lại để sử dụng

### ❌ Not Configured (Đỏ)
- Chưa cấu hình site replication
- Đang hoạt động như standalone MinIO instance
- Cần setup site replication nếu muốn sync

## 🔧 Use Cases

### Case 1: Quản lý MinIO Standalone Instances

Nếu bạn có nhiều MinIO instances **CHƯA** cấu hình site replication:

```bash
# Add aliases
mc alias set prod1 http://prod1.example.com minioadmin password1
mc alias set prod2 http://prod2.example.com minioadmin password2
mc alias set dev1 http://dev.example.com minioadmin password3

# Tool sẽ hiển thị tất cả với status "Not Configured" (đỏ)
# Bạn có thể:
# - View buckets trên mỗi instance
# - Compare configurations manually
# - Decide which ones to set up replication
```

### Case 2: Quản lý Site Replication Group

Nếu bạn đã có site replication setup:

```bash
# Các sites đã được config replication
mc admin replicate add site1 site2 site3

# Tool sẽ:
# - Auto-detect replication group
# - Hiển thị sites với status "Configured" (xanh)
# - Monitor replication status
# - Detect configuration drift
# - Suggest sync operations
```

### Case 3: Mixed Environment

Một số sites có replication, một số không:

```bash
# Site1 và Site2 có replication
mc admin replicate add site1 site2

# Site3 standalone
mc alias set site3 http://site3.example.com admin pass

# Tool sẽ hiển thị:
# - site1, site2: Configured (xanh)
# - site3: Not Configured (đỏ)
```

## 📖 API Endpoints

### GET /api/replication/info
Lấy thông tin về tất cả MinIO aliases và replication status:

```json
{
  "enabled": true,
  "aliases": [
    {
      "alias": "site1",
      "endpoint": "http://localhost:9001",
      "healthy": true,
      "replicationEnabled": true,
      "replicationStatus": "configured",
      "siteName": "site1-name",
      "deploymentID": "abc-123",
      "serverCount": 4
    }
  ],
  "totalAliases": 2,
  "replicationGroup": {...}
}
```

### GET /api/replication/status
Monitor replication status:

```json
{
  "status": "healthy",
  "sites": {
    "site1": {
      "replicatedBuckets": 5,
      "pendingObjects": 0,
      "failedObjects": 0,
      "lastSyncTime": "2025-10-10T10:00:00Z",
      "healthy": true
    }
  }
}
```

### GET /api/replication/compare
So sánh configurations:

```json
{
  "buckets": {
    "mybucket": {
      "existsOn": ["site1", "site2"],
      "policy": {
        "consistent": false,
        "values": {
          "site1": "policy1",
          "site2": "policy2"
        }
      },
      "lifecycle": {
        "consistent": true
      }
    }
  }
}
```

## 🛠️ Development

### Build từ source

```bash
git clone https://github.com/yourusername/mc-tool
cd mc-tool
go build -o mc-tool main.go
```

### Run in development mode

```bash
go run main.go web
```

### Build portable binary

```bash
go build -ldflags="-s -w" -o mc-tool-portable main.go
```

## 🧪 Testing với Docker

Xem file `docker-compose.yml` và `setup-minio.sh` để test với 2 MinIO instances local.

```bash
# Start MinIO instances
docker-compose up -d

# Setup aliases
./setup-minio.sh

# Run tool
./mc-tool web
```

## 📝 Notes

- Tool **KHÔNG** tự động enable site replication. Bạn phải dùng `mc admin replicate add` để setup.
- Tool chỉ **monitor và manage** các MinIO instances đã được add vào mc config.
- Consistency check không modify data, chỉ so sánh và báo cáo.
- Sync operations (đang development) sẽ cho phép auto-sync configurations.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see LICENSE file for details.
