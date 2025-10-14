# MC-Tool Site Replication Documentation

## Tổng quan

Thư mục này chứa tài liệu chi tiết về tính năng quản trị Site Replication cho MinIO trong MC-Tool.

## Cấu trúc Tài liệu

### 📋 [site-replication-management.md](./site-replication-management.md)
**Tài liệu Yêu cầu Tính năng Quản trị Site Replication**

- Tổng quan về tính năng và mục đích sử dụng
- Yêu cầu chức năng chi tiết (thêm, xóa, resync sites)
- Yêu cầu giao diện người dùng và Lucid Icons integration
- Yêu cầu kỹ thuật (APIs, commands, error handling)
- Yêu cầu bảo mật, hiệu năng và triển khai
- Acceptance criteria và testing requirements

### 🔧 [smart-removal-algorithm.md](./smart-removal-algorithm.md)
**Đặc tả Kỹ thuật: Thuật toán Xóa Site Thông minh**

- Mô tả chi tiết algorithm xóa site với logic bảo tồn replication group
- Flowchart và implementation details
- Xử lý edge cases và error scenarios
- Test cases và API response formats
- Performance và security considerations

### 🎨 [ui-ux-requirements.md](./ui-ux-requirements.md)  
**Yêu cầu UI/UX: Giao diện Quản lý Site Replication**

- Design system với Lucid Icons integration
- Layout structure và component library
- Responsive design requirements
- Accessibility standards và browser compatibility
- Animation patterns và performance requirements

## Tính năng Chính Đã Implement

### ✅ 1. Lucid Icons Integration
- **Mô tả**: Tích hợp bộ icon Lucid cho giao diện nhất quán
- **Implementation**: Local files để hoạt động offline
- **Icons sử dụng**: Navigation, actions, status indicators
- **File liên quan**: `index.html`, `app.js`

### ✅ 2. Smart Site Removal Logic  
- **Mô tả**: Xóa site individual mà không phá hủy replication group
- **Algorithm**: 
  - Nếu còn 2 sites total → Xóa toàn bộ replication config
  - Nếu còn 3+ sites total → Chỉ xóa site cụ thể, giữ nguyên group
- **Command patterns**:
  - Complete removal: `mc admin replicate rm <alias> --all --force`
  - Individual removal: `mc admin replicate rm <remaining-alias> <target-alias> --force`
- **File implementation**: `pkg/web/server.go` - `handleReplicationRemove()`

## API Endpoints

### Information APIs
- `GET /api/replication/info` - Thông tin replication group
- `GET /api/replication/status` - Trạng thái chi tiết
- `GET /api/replication/compare` - So sánh consistency

### Management APIs  
- `POST /api/replication/add` - Thêm sites vào group
- `POST /api/replication/remove` - **Xóa site với logic thông minh**
- `POST /api/replication/resync` - Resync dữ liệu giữa sites

## Testing Results

### Live Testing Scenario
```bash
# Initial state: 4-site replication group
Sites: [site3, site4, site5, site6]

# Remove site6
curl -X POST http://localhost:8080/api/replication/remove \
  -H "Content-Type: application/json" \
  -d '{"alias": "site6"}'

# Result: 
{
  "success": true,
  "message": "Site 'site6' removed from replication successfully",
  "note": "Remaining sites in replication group: site3, site4, site5"
}

# Verified: Sites [site3, site4, site5] still in replication
# site6 replication status: disabled
```

## Kiến trúc System

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   Frontend (HTML)   │    │   Backend (Go)      │    │   MinIO Cluster     │
│                     │    │                     │    │                     │
│ • Lucid Icons       │────│ • REST APIs         │────│ • Site Replication  │
│ • Responsive UI     │    │ • Smart Remove      │    │ • MC Commands       │
│ • Dynamic Content   │    │ • Error Handling    │    │ • JSON Responses    │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
```

## Development Workflow

### 1. Setup & Configuration
```bash
# Build and run MC-Tool
go build
./mc-tool web --port 8080

# Access web interface
http://localhost:8080
```

### 2. MinIO Aliases Configuration
```bash
# Configure MinIO aliases
mc alias set site1 http://192.168.1.10:9001 admin password
mc alias set site2 http://192.168.1.11:9002 admin password
mc alias set site3 http://192.168.1.12:9003 admin password
```

### 3. Replication Setup
```bash
# Create replication group
mc admin replicate add site1 site2 site3

# Verify setup
mc admin replicate info site1 --json
```

## Troubleshooting

### Common Issues

#### Localhost Endpoints Error
```
❌ Problem: "localhost endpoints cannot connect to each other"
✅ Solution: Use IP addresses or domain names
   Example: mc alias set site1 http://192.168.1.10:9000 access secret
```

#### Connection Refused
```
❌ Problem: "connection refused"  
✅ Solutions:
   1. Check MinIO server is running
   2. Verify firewall settings
   3. Confirm correct port numbers
   4. Test network connectivity
```

#### Replication Already Exists
```
❌ Problem: "site replication already configured"
✅ Solution: Use remove logic to adjust existing groups
   or remove entire config before creating new one
```

## Best Practices

### 1. Network Configuration
- Sử dụng IP addresses thay vì localhost
- Đảm bảo tất cả sites có thể kết nối với nhau
- Cấu hình firewall cho MinIO ports

### 2. Site Management
- Backup data trước khi remove sites
- Test connectivity trước khi add sites
- Monitor replication status thường xuyên

### 3. UI/UX
- Sử dụng Lucid Icons cho consistency
- Implement proper error handling
- Provide clear user feedback

## Future Enhancements

### 1. Advanced Features
- [ ] Batch site removal
- [ ] Site replication health monitoring
- [ ] Automatic failover configuration
- [ ] Cross-region replication setup

### 2. UI Improvements  
- [ ] Real-time status updates
- [ ] Advanced filtering and search
- [ ] Drag-and-drop site management
- [ ] Dark mode support

### 3. Integration
- [ ] Prometheus metrics export
- [ ] Webhook notifications
- [ ] LDAP/SSO authentication
- [ ] Audit logging

## Support & Maintenance

### Documentation Updates
- Cập nhật tài liệu khi có thay đổi API
- Maintain version compatibility notes
- Update testing scenarios

### Code Reviews
- Verify Lucid Icons usage patterns
- Check smart removal logic correctness
- Validate error handling completeness

### Performance Monitoring
- Track API response times
- Monitor UI loading performance
- Optimize icon loading strategies

---

**Version**: 1.0  
**Last Updated**: October 14, 2025  
**Maintainer**: MC-Tool Development Team