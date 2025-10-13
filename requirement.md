# Tài liệu yêu cầu cho web tool vận hành MinIO replicated cluster

## 1. Mục tiêu
- Cung cấp giao diện web hỗ trợ đội vận hành quan sát trạng thái các alias MinIO, kiểm tra đồng bộ giữa các site, đánh giá cấu hình bucket và thao tác với quan hệ replication.
- Tích hợp toàn bộ công cụ trong container đi kèm `mc` và cấu hình alias có sẵn, đảm bảo triển khai nhanh cho các môi trường.

## 2. Phạm vi
- Hệ thống chỉ thao tác thông qua `mc` CLI và các alias MinIO được định nghĩa trước.
- Không bao gồm chức năng quản lý người dùng MinIO hay thay đổi cấu hình cluster ngoài các tác vụ replication và checklist bucket.

## 3. Đối tượng sử dụng
- Kỹ sư vận hành chịu trách nhiệm giám sát và bảo trì MinIO replicated cluster.
- Nhân sự hỗ trợ kiểm thử/bảo đảm chất lượng cần xác minh tính ổn định của replication.

## 4. Yêu cầu chức năng

### 4.1 Quản lý trạng thái alias MinIO
- Đọc danh sách alias và thông tin kết nối từ file cấu hình `~/.mc/config.json` được đóng gói trong container.
- Hiển thị trạng thái kết nối của từng alias (online/offline, thời gian phản hồi, phiên bản MinIO, dung lượng tổng/quyền truy cập).
- Cung cấp nút cập nhật trạng thái thủ công và tự động làm mới theo chu kỳ có cấu hình (mặc định 60s).
- Cho phép xem chi tiết alias: cấu hình endpoint, thông tin chứng chỉ TLS, thông số load (số bucket, số object, dung lượng).

### 4.2 So sánh bucket giữa hai site
- Lựa chọn cặp alias nguồn/đích và bucket tương ứng; hỗ trợ auto-suggest danh sách bucket dựa trên `mc ls`.
- Thực hiện đối chiếu danh sách object, so sánh version/timestamp/etag và phát hiện chênh lệch (thiếu, lỗi phiên bản, dung lượng khác).
- Hiển thị kết quả dạng bảng với khả năng lọc theo loại bất đồng bộ, xuất báo cáo CSV/JSON.
- Cho phép đánh dấu một bucket vào danh sách theo dõi để tự động kiểm tra định kỳ và cảnh báo khi có chênh lệch mới.

### 4.3 Checklist cấu hình bucket
- Thu thập cấu hình webhook (notification) và lifecycle cho từng bucket bằng `mc admin`/`mc ilm`.
- So sánh cấu hình thực tế với bộ tiêu chuẩn do vận hành định nghĩa bằng chính các file cấu hình tương thích `mc` (ví dụ file JSON cho `mc ilm import`, script `mc admin config set`) được mount vào container.
- Công cụ đọc các file chuẩn này, chạy chế độ dry-run/parse để trích xuất thông số chuẩn, sau đó đối chiếu với cấu hình đang chạy mà `mc` trả về.
- Trả về trạng thái pass/fail kèm mô tả chi tiết chỗ sai khác và gợi ý cách khắc phục.
- Hỗ trợ tải xuống báo cáo checklist tổng hợp và đánh dấu bucket đã kiểm tra.

### 4.4 Quản lý replication
- Liệt kê các quan hệ replication hiện có, bao gồm trạng thái, rule ID, lịch sử đồng bộ, độ trễ ước tính.
- Cho phép thêm replication mới giữa hai alias/bucket với giao diện wizard: kiểm tra điều kiện (phiên bản MinIO, quyền truy cập, bucket rỗng hoặc không), nhập thông số replication (rule name, filter, lịch).
- Cho phép tạm dừng, tiếp tục, xoá replication và kích hoạt lại (resync) cho rule cụ thể.
- Ghi log chi tiết những thao tác replication và phản hồi lại kết quả/ lỗi của `mc`.

### 4.5 Khả năng tìm kiếm và thông báo
- Tìm kiếm alias, bucket, rule replication theo từ khoá.
- Thiết lập cảnh báo (email/webhook nội bộ) khi phát hiện alias offline, so sánh bucket thất bại hoặc checklist không đạt.

## 5. Yêu cầu giao diện người dùng
- Dashboard tổng quan hiển thị trạng thái alias, số lượng cảnh báo replication, số bucket cần chú ý.
- Bảng dữ liệu hỗ trợ sắp xếp, lọc, phân trang; hiển thị rõ ràng trạng thái bằng màu sắc.
- Trang chi tiết alias/bucket có biểu đồ đơn giản (progress bar, sparkline) để minh hoạ dung lượng và số object.
- Hỗ trợ theme sáng/tối, ngôn ngữ tiếng Việt là mặc định; chuẩn bị cơ chế i18n để mở rộng.

## 6. Tích hợp và triển khai
- Ứng dụng web chạy dưới dạng container duy nhất bao gồm backend viết bằng Go, frontend tĩnh, `mc`, và file cấu hình alias (`config.json`) mount vào `/root/.mc/config.json`.
- Cho phép cấu hình các tham số (chu kỳ refresh, endpoint cảnh báo, chuẩn checklist) thông qua biến môi trường khi khởi động container.
- Đảm bảo container có endpoint healthcheck để hệ thống giám sát có thể theo dõi.
- Mount thêm thư mục `/app/reference-configs` chứa các file cấu hình chuẩn (định dạng `mc` hỗ trợ) để phục vụ checklist.
- Không dùng cơ sở dữ liệu; lưu cache tạm thời trong bộ nhớ và ghi lịch sử/báo cáo ra thư mục volume dạng file JSONL/YAML để thuận tiện backup.

## 7. Yêu cầu phi chức năng
- **Hiệu năng**: Có thể xử lý so sánh bucket với ít nhất 100k object trong vòng <5 phút; dùng tác vụ nền và cập nhật tiến độ theo thời gian thực.
- **Độ tin cậy**: Nếu lệnh `mc` thất bại, hiển thị thông báo rõ ràng và cho phép retry; ghi lại lịch sử lỗi.
- **Bảo mật**: Mã hoá thông tin xác thực alias trong container; hạn chế truy cập giao diện bằng OAuth2/SAML hoặc basic auth cấu hình qua biến môi trường.
- **Khả năng mở rộng**: Thiết kế backend theo hướng service layer để dễ mở rộng thêm loại checklist hoặc plugin tích hợp khác.
- **Quan sát được**: Xuất log chuẩn JSON, cung cấp metrics Prometheus (số alias online, số replication bị lỗi, thời gian so sánh).

## 8. Báo cáo và audit
- Lưu lịch sử các lần so sánh bucket và kết quả checklist bằng file JSONL (mỗi dòng một bản ghi) trong volume được mount, kèm timestamp và người thực hiện.
- Ghi nhận mọi thao tác quản trị (add/remove/resync replication) kèm người thực hiện và thời gian.
- Cho phép tải xuống báo cáo hoạt động theo ngày/tuần/tháng.

## 9. Giả định và ràng buộc
- Các alias được cấu hình đầy đủ quyền cần thiết để thực hiện các thao tác yêu cầu (admin hoặc equivalent).
- Hệ thống có quyền truy cập mạng tới tất cả endpoint MinIO tương ứng.
- Đội vận hành cung cấp bộ tiêu chuẩn checklist bucket dưới dạng file cấu hình extern.
- Có volume (hoặc bind mount) dùng để lưu trữ file lịch sử và checklist, đảm bảo không mất dữ liệu khi container restart.
- Các file chuẩn checklist tuân thủ đúng định dạng mà `mc` có thể áp dụng trực tiếp (JSON/TOML/script CLI) để đảm bảo tính khả dụng khi cần reapply cấu hình.

## 10. Cần làm rõ thêm
- Quy trình quản lý version, review và phân quyền cập nhật bộ file cấu hình chuẩn tương thích `mc`.
- Quy trình xác thực người dùng cuối cùng (SSO hay tài khoản riêng).
- Cách đồng bộ file lịch sử giữa các môi trường (ví dụ đẩy sang object storage hoặc mount NFS).
