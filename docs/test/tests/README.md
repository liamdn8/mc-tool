# MC-Tool Test Suite

## Cấu trúc Test

### 📁 `integration/`
Testcase tích hợp cho tính năng Site Replication Management:
- `smart_removal_test.go` - Test thuật toán xóa site thông minh
- `replication_apis_test.go` - Test các API endpoints
- `ui_integration_test.go` - Test UI/UX integration
- `error_handling_test.go` - Test xử lý lỗi

### 📁 `unit/`
Unit tests cho các modules riêng lẻ:
- `web_server_test.go` - Test web server functions
- `validation_test.go` - Test input validation

### 📁 `fixtures/`
Test data và mock responses:
- `minio_responses.json` - Mock MinIO command responses
- `test_configs.json` - Test configurations

## Chạy Tests

```bash
# Chạy tất cả tests
go test ./tests/...

# Chạy integration tests
go test ./tests/integration/...

# Chạy specific test
go test ./tests/integration/ -run TestSmartSiteRemoval

# Chạy với coverage
go test -cover ./tests/...
```

## Test Requirements

- MinIO test environment (sử dụng Docker)
- Test sites configuration
- Mock HTTP clients cho API testing