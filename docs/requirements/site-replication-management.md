# Tài liệu Yêu cầu Tính năng Quản trị Site Replication cho MinIO

## Tổng quan

MC-Tool cung cấp giao diện web để quản lý site replication cho MinIO, cho phép người dùng dễ dàng thiết lập, giám sát và quản lý nhân bản dữ liệu giữa nhiều site MinIO.

## 1. Yêu cầu Chức năng

### 1.1 Quản lý Site Replication

#### 1.1.1 Xem thông tin Replication Group
- **Mục đích**: Hiển thị trạng thái tổng quan của replication group
- **Chức năng**:
  - Liệt kê tất cả sites trong replication group
  - Hiển thị trạng thái healthy/unhealthy của từng site
  - Hiển thị deployment ID và endpoint của từng site
  - Hiển thị tổng số sites được cấu hình
  - Hiển thị trạng thái enable/disable của replication

#### 1.1.2 Thêm Sites vào Replication Group
- **Mục đích**: Thiết lập site replication mới hoặc thêm sites vào group hiện tại
- **Yêu cầu**:
  - Cần tối thiểu 2 sites để thiết lập replication
  - Tất cả sites phải có thể kết nối được với nhau
  - Sites phải sử dụng endpoint có thể truy cập (không phải localhost)
- **Xử lý lỗi**:
  - Cung cấp thông báo chi tiết khi thiết lập thất bại
  - Hướng dẫn người dùng cách khắc phục lỗi phổ biến
  - Kiểm tra và báo lỗi kết nối mạng

#### 1.1.3 Xóa Sites khỏi Replication Group ⭐ **Tính năng chính**
- **Mục đích**: Xóa một site cụ thể khỏi replication group mà không phá hủy toàn bộ configuration
- **Logic thông minh**:
  - Nếu chỉ còn 2 sites (xóa 1 thì còn 1): Xóa toàn bộ replication configuration
  - Nếu có 3+ sites (xóa 1 thì còn 2+): Chỉ xóa site cụ thể, giữ nguyên replication group
- **Thuật toán**:
  ```
  1. Lấy thông tin replication group hiện tại
  2. Xác định danh sách sites còn lại sau khi xóa
  3. Nếu remainingSites.length == 1:
     - Sử dụng: mc admin replicate rm [alias] --all --force
  4. Nếu remainingSites.length > 1:
     - Sử dụng: mc admin replicate rm [remaining-alias] [target-alias] --force
  ```
- **Response**:
  - Hiển thị danh sách sites còn lại trong group
  - Thông báo rõ ràng về hành động đã thực hiện

#### 1.1.4 Resync dữ liệu giữa Sites
- **Mục đích**: Đồng bộ hóa dữ liệu giữa các sites
- **Chức năng**:
  - Resync FROM source (pull data from source)
  - Resync TO target (push data to target)
  - Theo dõi trạng thái resync process

### 1.2 Giám sát và Báo cáo

#### 1.2.1 Replication Status
- **Chức năng**:
  - Hiển thị số lượng buckets được replicated
  - Hiển thị số objects đang pending
  - Hiển thị số objects failed
  - Thời gian sync cuối cùng

#### 1.2.2 Consistency Check
- **Chức năng**:
  - So sánh configuration giữa các sites
  - Kiểm tra bucket policies
  - Kiểm tra lifecycle rules
  - Kiểm tra versioning settings
  - Đánh dấu inconsistencies

## 2. Yêu cầu Giao diện Người dùng

### 2.1 Thiết kế UI/UX

#### 2.1.1 Lucid Icons Integration ⭐ **Tính năng chính**
- **Mục đích**: Sử dụng bộ icon Lucid để có giao diện nhất quán và đẹp mắt
- **Implementation**:
  - CDN: `https://unpkg.com/lucide@latest/dist/umd/lucide.js`
  - Cách sử dụng: `<i data-lucide="icon-name" width="size" height="size"></i>`
- **Icons được sử dụng**:
  - Navigation: `layout-dashboard`, `globe`, `folder`, `repeat`, `check-circle`, `settings`
  - Header: `package`, `refresh-cw`
  - Actions: `plus`, `download`, `upload`, `trash-2`
- **Dynamic Initialization**: Gọi `lucide.createIcons()` sau khi thêm nội dung động

#### 2.1.2 Layout Structure
- **Header**: Logo, title, language selector, refresh button
- **Sidebar**: Navigation menu với các trang chính
- **Main Content**: Nội dung trang được chọn
- **Status Indicator**: Hiển thị trạng thái MC tool

#### 2.1.3 Responsive Design
- Hỗ trợ desktop và mobile
- Sidebar có thể collapse trên mobile
- Grid layout responsive cho cards và tables

### 2.2 Trang chính

#### 2.2.1 Overview Page
- Replication Group summary
- Quick stats (sites count, status)
- Recent activities
- Quick actions (Add Site, Refresh)

#### 2.2.2 Sites Page
- Danh sách tất cả sites
- Trạng thái của từng site
- Actions: Add, Remove, Resync

#### 2.2.3 Replication Status Page
- Chi tiết trạng thái replication
- Metrics và charts
- Health indicators

## 3. Yêu cầu Kỹ thuật

### 3.1 Backend APIs

#### 3.1.1 Information APIs
```
GET /api/replication/info
- Trả về thông tin tổng quan về replication group
- Response: enabled, aliases[], replicationGroup{}

GET /api/replication/status  
- Trả về trạng thái chi tiết của replication
- Response: sites{}, replicatedBuckets, pendingObjects

GET /api/replication/compare
- So sánh configuration giữa các sites
- Response: buckets{}, consistency status
```

#### 3.1.2 Management APIs
```
POST /api/replication/add
Body: {"aliases": ["site1", "site2", ...]}
- Thêm sites vào replication group

POST /api/replication/remove ⭐ **API chính**
Body: {"alias": "site1"}
- Xóa site khỏi replication group với logic thông minh
- Response: success, message, note (danh sách sites còn lại)

POST /api/replication/resync
Body: {"source_alias": "site1", "target_alias": "site2", "direction": "resync-from|resync-to"}
- Thực hiện resync dữ liệu giữa sites
```

### 3.2 MinIO Commands Integration

#### 3.2.1 Core Commands
- `mc admin replicate add [aliases...]`: Thiết lập replication
- `mc admin replicate rm [alias] --all --force`: Xóa toàn bộ replication
- `mc admin replicate rm [remaining-alias] [target-alias] --force`: Xóa site cụ thể
- `mc admin replicate info [alias] --json`: Lấy thông tin replication
- `mc admin replicate resync start --deployment-id [target] [source]`: Resync

### 3.3 Error Handling

#### 3.3.1 Common Error Scenarios
- **Localhost endpoints**: Hướng dẫn cấu hình IP addresses
- **Connection refused**: Kiểm tra MinIO server status
- **Network issues**: Hướng dẫn troubleshooting
- **Permission errors**: Kiểm tra credentials và permissions

#### 3.3.2 User-friendly Messages
- Cung cấp thông báo lỗi chi tiết bằng tiếng Việt và tiếng Anh
- Đưa ra hướng dẫn khắc phục cụ thể
- Hiển thị technical details khi cần thiết

## 4. Yêu cầu Bảo mật

### 4.1 Authentication & Authorization
- Sử dụng MinIO credentials để authenticate
- Kiểm tra permissions trước khi thực hiện operations
- Secure API endpoints

### 4.2 Network Security
- HTTPS support cho production
- CORS configuration
- Input validation và sanitization

## 5. Yêu cầu Hiệu năng

### 5.1 Response Time
- API responses < 5 seconds cho operations thông thường
- UI loading < 2 seconds
- Real-time status updates

### 5.2 Scalability
- Hỗ trợ tối thiểu 10 sites trong replication group
- Efficient polling cho status updates
- Lazy loading cho large datasets

## 6. Yêu cầu Khả năng Sử dụng

### 6.1 Đa ngôn ngữ
- Hỗ trợ tiếng Việt và tiếng Anh
- Dynamic language switching
- Localized error messages

### 6.2 Accessibility
- Keyboard navigation support
- Screen reader compatibility
- High contrast mode support

## 7. Yêu cầu Triển khai

### 7.1 Environment Requirements
- Go 1.19+ cho backend
- Modern browsers cho frontend
- MinIO client (mc) tool installed
- Network connectivity giữa các MinIO sites

### 7.2 Configuration
- Environment variables cho MinIO credentials
- Configurable timeouts và intervals
- Logging configuration

## 8. Acceptance Criteria

### 8.1 Functional Testing
- ✅ Có thể thiết lập replication group với 2+ sites
- ✅ Có thể xóa individual site mà không phá hủy group (khi có 3+ sites)
- ✅ Có thể xóa toàn bộ replication khi chỉ còn 2 sites
- ✅ UI hiển thị chính xác trạng thái và danh sách sites
- ✅ Lucid Icons hiển thị đúng trên tất cả components

### 8.2 Error Handling Testing
- ✅ Xử lý đúng localhost endpoint errors
- ✅ Xử lý đúng connection refused errors
- ✅ Hiển thị user-friendly error messages

### 8.3 UI/UX Testing
- ✅ Responsive design trên mobile và desktop
- ✅ Consistent icon usage với Lucid Icons
- ✅ Smooth navigation giữa các trang
- ✅ Real-time status updates

## 9. Ghi chú Kỹ thuật

### 9.1 Smart Remove Logic Implementation
Logic xóa site thông minh được implement trong `handleReplicationRemove()`:

```go
// Lấy thông tin replication group
replicateInfoCmd := exec.Command("mc", "admin", "replicate", "info", req.Alias, "--json")

// Parse remaining sites
var remainingSites []string
// ... logic parse sites ...

// Decision logic
if len(remainingSites) == 1 {
    // Remove entire config
    cmd := exec.Command("mc", "admin", "replicate", "rm", req.Alias, "--all", "--force")
} else {
    // Remove specific site
    remainingAlias := remainingSites[0]
    cmd := exec.Command("mc", "admin", "replicate", "rm", remainingAlias, req.Alias, "--force")
}
```

### 9.2 Lucid Icons Integration
```html
<!-- CDN Include -->
<script src="https://unpkg.com/lucide@latest/dist/umd/lucide.js"></script>

<!-- Icon Usage -->
<i data-lucide="trash-2" width="16" height="16"></i>

<!-- JavaScript Initialization -->
<script>
if (typeof lucide !== 'undefined') {
    lucide.createIcons();
}
</script>
```

Tài liệu này mô tả đầy đủ các yêu cầu cho tính năng quản trị site replication, bao gồm cả logic thông minh để xóa sites và integration Lucid Icons cho giao diện nhất quán.