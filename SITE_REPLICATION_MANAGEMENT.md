# Site Replication Management Features

## Tổng quan / Overview

Các tính năng quản lý Site Replication cho phép bạn thực hiện các thao tác quản lý cluster MinIO Site Replication trực tiếp từ giao diện web.

Site Replication Management features allow you to perform MinIO Site Replication cluster management operations directly from the web interface.

---

## Tính năng / Features

### 1. Thêm Sites vào Replication / Add Sites to Replication

**Mô tả / Description:**
- Cho phép chọn nhiều alias theo thứ tự để tạo cluster site replication
- Thực hiện lệnh `mc admin replicate add` ở backend

**Cách sử dụng / Usage:**

1. Vào trang **Sites** từ menu bên trái
2. Nếu chưa cấu hình replication, bạn sẽ thấy form "Setup Site Replication"
3. Chọn các alias muốn thêm vào cluster (tối thiểu 2 sites)
4. Thứ tự chọn sẽ được hiển thị trong "Selected Order"
5. Nhấn nút **"Add Sites to Replication"**
6. Xác nhận và đợi quá trình hoàn tất

**Backend Command:**
```bash
mc admin replicate add <alias1> <alias2> [alias3] ...
```

**API Endpoint:**
```
POST /api/replication/add
Content-Type: application/json

{
  "aliases": ["alias1", "alias2", "alias3"]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Site replication added successfully",
  "output": "..."
}
```

---

### 2. Xóa Site khỏi Replication / Remove Site from Replication

**Mô tả / Description:**
- Cho phép xóa một alias khỏi cluster replication
- Thực hiện lệnh `mc admin replicate rm` với flag `--force`

**Cách sử dụng / Usage:**

1. Vào trang **Sites** 
2. Trong phần "Manage Site Replication", tìm site muốn xóa
3. Nhấn nút **"Remove"** (biểu tượng thùng rác màu đỏ)
4. Xác nhận trong hộp thoại cảnh báo
5. Đợi quá trình hoàn tất

**Cảnh báo / Warning:**
⚠️ Xóa site sẽ dừng replication cho site đó. Dữ liệu vẫn được giữ nguyên nhưng sẽ không còn được đồng bộ.

**Backend Command:**
```bash
mc admin replicate rm <alias> --force
```

**API Endpoint:**
```
POST /api/replication/remove
Content-Type: application/json

{
  "alias": "alias1"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Site removed from replication successfully",
  "output": "..."
}
```

---

### 3. Resync Dữ liệu / Resync Data

**Mô tả / Description:**
- Cho phép đồng bộ lại dữ liệu giữa các sites
- Hỗ trợ 2 hướng:
  - **Resync FROM**: Kéo dữ liệu từ site nguồn (pull)
  - **Resync TO**: Đẩy dữ liệu đến site đích (push)

**Cách sử dụng / Usage:**

#### Resync FROM (Pull data)

1. Vào trang **Sites**
2. Tìm site muốn làm nguồn để kéo dữ liệu
3. Nhấn nút **"Resync From"** (biểu tượng mũi tên xuống)
4. Chọn site đích trong hộp thoại
5. Xác nhận thao tác
6. Quá trình resync sẽ bắt đầu

**Ví dụ / Example:**
```
Site A (source) → Site B (target)
Resync FROM Site A: Kéo dữ liệu từ A về B
```

#### Resync TO (Push data)

1. Vào trang **Sites**
2. Tìm site muốn làm đích để đẩy dữ liệu
3. Nhấn nút **"Resync To"** (biểu tượng mũi tên lên)
4. Chọn site nguồn trong hộp thoại
5. Xác nhận thao tác
6. Quá trình resync sẽ bắt đầu

**Ví dụ / Example:**
```
Site A (source) → Site B (target)
Resync TO Site B: Đẩy dữ liệu từ A sang B
```

**Backend Command:**
```bash
# Resync FROM
mc admin replicate resync start --deployment-id <target-deployment-id> <source-alias>

# Resync TO  
mc admin replicate resync start --deployment-id <target-deployment-id> <source-alias>
```

**API Endpoint:**
```
POST /api/replication/resync
Content-Type: application/json

{
  "source_alias": "alias1",
  "target_alias": "alias2",
  "direction": "resync-from" | "resync-to"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Resync started successfully",
  "output": "..."
}
```

---

## Giao diện / UI Components

### Trạng thái Site / Site Status

Mỗi site hiển thị trạng thái replication:
- ✅ **Active** (badge xanh): Site đang tham gia replication
- ⚠️ **Inactive** (badge vàng): Site không active trong replication

### Nút thao tác / Action Buttons

1. **Resync From** (icon mũi tên xuống):
   - Tooltip: "Resync FROM this site (pull data)"
   - Kéo dữ liệu từ site này

2. **Resync To** (icon mũi tên lên):
   - Tooltip: "Resync TO this site (push data)"
   - Đẩy dữ liệu đến site này

3. **Remove** (icon thùng rác đỏ):
   - Tooltip: "Remove from replication cluster"
   - Xóa site khỏi cluster

### Hộp thoại chọn Site / Site Selection Dialog

Khi thực hiện resync, một modal sẽ hiện ra cho phép chọn site nguồn/đích:
- Hiển thị danh sách các site khác
- Mỗi site hiển thị tên và URL
- Click vào site để chọn
- Có nút "Cancel" để hủy thao tác

---

## Thông báo / Notifications

Hệ thống hiển thị thông báo cho mỗi thao tác:

- ✅ **Success** (xanh lá): Thao tác thành công
- ✗ **Error** (đỏ): Có lỗi xảy ra
- ℹ **Info** (xanh dương): Thông tin

Thông báo tự động biến mất sau 5 giây.

---

## Bảo mật / Security

### Xác nhận thao tác / Confirmation Dialogs

Các thao tác quan trọng yêu cầu xác nhận:

1. **Remove Site:**
   ```
   Are you sure you want to remove "alias1" from the replication cluster?
   
   This will stop replication for this site.
   ```

2. **Resync:**
   ```
   Resync FROM "alias1" TO "alias2"?
   
   This will pull data from alias1 to alias2.
   ```

### Logging

Tất cả thao tác được ghi log với structured logging:

```json
{
  "timestamp": "2025-10-13T12:00:00Z",
  "level": "info",
  "message": "Adding site replication",
  "fields": {
    "aliases": ["alias1", "alias2", "alias3"]
  }
}
```

---

## Xử lý lỗi / Error Handling

### Frontend Validation

- Kiểm tra tối thiểu 2 aliases khi add sites
- Disable nút khi không đủ điều kiện
- Hiển thị thông báo lỗi rõ ràng

### Backend Validation

- Validate request body
- Kiểm tra alias tồn tại
- Kiểm tra direction (resync-from/resync-to)

### Error Messages

```json
{
  "error": "At least 2 aliases are required",
  "error": "Alias is required",
  "error": "Direction must be 'resync-from' or 'resync-to'",
  "error": "Failed to add replication: <output>"
}
```

---

## Workflow ví dụ / Example Workflow

### Thiết lập Site Replication mới

1. **Chuẩn bị:**
   - Đảm bảo có ít nhất 2 MinIO aliases đã cấu hình
   - Kiểm tra kết nối đến các sites

2. **Thêm Sites:**
   ```
   Chọn: minio1, minio2, minio3 (theo thứ tự)
   → Click "Add Sites to Replication"
   → Xác nhận
   → Đợi hoàn tất
   ```

3. **Xác minh:**
   - Refresh trang Sites
   - Kiểm tra status badge (Active)
   - Kiểm tra trang Replication Status

### Resync sau sự cố

1. **Phát hiện vấn đề:**
   - Site B bị mất dữ liệu
   - Cần đồng bộ lại từ Site A

2. **Thực hiện Resync:**
   ```
   Vào Sites → Tìm Site A
   → Click "Resync From"
   → Chọn Site B
   → Xác nhận
   → Monitor progress
   ```

3. **Kiểm tra:**
   - Vào Consistency Check
   - Run check để xác minh
   - Kiểm tra bucket counts

---

## API Reference

### POST /api/replication/add

**Request:**
```json
{
  "aliases": ["alias1", "alias2", "alias3"]
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Site replication added successfully",
  "output": "Successfully added sites to replication..."
}
```

**Error Response (400/500):**
```json
{
  "error": "At least 2 aliases are required"
}
```

---

### POST /api/replication/remove

**Request:**
```json
{
  "alias": "alias1"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Site removed from replication successfully",
  "output": "Successfully removed site from replication..."
}
```

**Error Response (400/500):**
```json
{
  "error": "Alias is required"
}
```

---

### POST /api/replication/resync

**Request:**
```json
{
  "source_alias": "alias1",
  "target_alias": "alias2",
  "direction": "resync-from"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Resync started successfully",
  "output": "Resync started for deployment..."
}
```

**Error Response (400/500):**
```json
{
  "error": "Direction must be 'resync-from' or 'resync-to'"
}
```

---

## Troubleshooting

### "At least 2 aliases are required"
- **Nguyên nhân:** Chưa chọn đủ 2 sites
- **Giải pháp:** Chọn thêm ít nhất 1 alias nữa

### "Failed to add replication: ..."
- **Nguyên nhân:** Lệnh mc admin thất bại
- **Giải pháp:** 
  - Kiểm tra mc client đã cài đặt
  - Kiểm tra aliases đã cấu hình đúng
  - Xem output message để biết chi tiết

### "No other sites available for resync"
- **Nguyên nhân:** Chỉ có 1 site trong cluster
- **Giải pháp:** Thêm site khác vào cluster trước

---

## Best Practices

1. **Thêm Sites:**
   - Thêm sites theo thứ tự ưu tiên (master first)
   - Đảm bảo tất cả sites đều healthy trước khi add
   - Backup cấu hình trước khi thay đổi

2. **Remove Sites:**
   - Chỉ remove khi thực sự cần thiết
   - Backup dữ liệu site trước khi remove
   - Notify team trước khi thực hiện

3. **Resync:**
   - Chỉ resync khi cần thiết (sau sự cố, recovery)
   - Monitor network bandwidth trong quá trình resync
   - Sử dụng resync-from để pull từ master site

---

## Changelog

**Version 1.0.0** (October 13, 2025)
- ✅ Thêm tính năng Add Sites to Replication
- ✅ Thêm tính năng Remove Site from Replication
- ✅ Thêm tính năng Resync (FROM/TO)
- ✅ UI với modal selection dialog
- ✅ Notification system
- ✅ Structured logging
- ✅ Error handling và validation
- ✅ Bilingual support (EN/VI)
