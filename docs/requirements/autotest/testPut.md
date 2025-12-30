# SOFTWARE REQUIREMENTS SPECIFICATION (SRS)

## Automated MinIO PUT Replication Test Tool

---

## 1. Giới thiệu

### 1.1. Mục đích

Tài liệu SRS này mô tả các yêu cầu chức năng và phi chức năng cho **công cụ kiểm thử tự động hành vi PUT object kết hợp Site Replication của MinIO**, sử dụng **minio mc client** làm công cụ thao tác chính.

Công cụ hướng tới:

* Đánh giá **replication latency / throughput**
* Phát hiện **inconsistency** giữa các site
* Hỗ trợ **benchmark, regression test và CI/CD**

---

### 1.2. Phạm vi hệ thống

Công cụ kiểm thử:

* Áp dụng cho **MinIO cluster nhiều site (6 instance)**
* Workload chính: **PUT object kích thước nhỏ (1–10 KiB)**
* Test **replication behaviour**, không test internal queue

Ngoài phạm vi:

* GET / HEAD / LIST (test riêng)
* MinIO admin/internal API
* Warp benchmark

---

### 1.3. Định nghĩa & thuật ngữ

| Thuật ngữ       | Giải thích                                                    |
| --------------- | ------------------------------------------------------------- |
| Site            | Một MinIO instance trong mô hình site replication             |
| Primary site    | Site nhận PUT ban đầu                                         |
| Replication lag | Thời gian từ PUT đến khi object xuất hiện trên site cuối cùng |
| Completion time | Thời điểm toàn bộ site đạt trạng thái一致                       |
| mc              | MinIO Client                                                  |

---

## 2. Tổng quan hệ thống

### 2.1. Mô hình tổng thể

Công cụ hoạt động theo mô hình **client-side orchestration**:

* Điều phối các lệnh `mc cp`
* Theo dõi trạng thái object bằng `mc stat`
* Tính toán thời gian replication

Không yêu cầu agent trên MinIO server.

---

### 2.2. Đối tượng sử dụng

* DevOps Engineer
* SRE
* Performance/Test Engineer

---

## 3. Yêu cầu chức năng

### 3.1. Quản lý cấu hình

**FR-01:** Hệ thống phải cho phép cấu hình danh sách site MinIO (alias mc). --> đã có
**FR-02:** Hệ thống phải cho phép cấu hình bucket test và object path thực hiện test.
**FR-03:** Hệ thống phải cho phép cấu hình kích thước file test.
**FR-04:** Hệ thống phải cho phép cấu hình số lượng object PUT.

---

### 3.2. Thực thi testcase PUT

**FR-05:** Hệ thống phải hỗ trợ PUT N object (cho phép nhập từ input, có hỗ trợ select nhanh các giá trị 1, 10, 100, 1000, 2000, 5000, 10000).

**FR-06:** Hệ thống phải hỗ trợ PUT song song bằng nhiều process (cho phép chọn số process, default = 1, cho phép select nhanh một trong các giá trị: 1, 2, 5, 10, 20).

**FR-07:** Hệ thống phải hỗ trợ PUT overwrite cùng key trong cả 2 trường hợp bucket enable/disable versioning (tự động detech).

---

### 3.3. Đo replication

**FR-08:** Hệ thống phải ghi nhận thời gian bắt đầu PUT.

**FR-09:** Hệ thống phải poll trạng thái object trên tất cả site bằng `mc stat`.

**FR-10:** Hệ thống phải xác định site cuối cùng hoàn tất replication.

**FR-11:** Hệ thống phải tính replication completion time cho từng object.

**FR-12:** Hệ thống phải tính query được prometheus (nếu được cấu hình)

---

### 3.4. Xử lý tình huống lỗi & gián đoạn

**FR-14:** Hệ thống phải phát hiện site không phản hồi khi poll.

**FR-15:** Hệ thống phải tiếp tục poll khi site bị gián đoạn tạm thời.

**FR-16:** Hệ thống phải đánh dấu testcase fail nếu object không replicate sau timeout cấu hình.

---

### 3.5. Đánh giá & xác nhận kết quả

**FR-17:** Hệ thống phải xác nhận object tồn tại đầy đủ trên tất cả site.

**FR-18:** Hệ thống phải phát hiện inconsistency giữa các site.

**FR-19:** Hệ thống phải hỗ trợ so sánh trạng thái bằng `mc diff` (tuỳ chọn).

---

### 3.6. Báo cáo kết quả

**FR-20:** Hệ thống phải sinh báo cáo kết quả theo testcase.

**FR-21:** Báo cáo phải bao gồm:

* PUT latency
* Replication completion time
* Trạng thái PASS / FAIL

<!-- **FR-22:** Hệ thống phải hỗ trợ export báo cáo dạng JSON hoặc CSV. -->

---

## 4. Yêu cầu phi chức năng

### 4.1. Hiệu năng

**NFR-01:** Công cụ không được tạo tải vượt quá workload test đã cấu hình.

---

### 4.2. Độ tin cậy

**NFR-02:** Công cụ phải xử lý retry polling một cách an toàn.

---

### 4.3. Khả năng mở rộng

**NFR-03:** Công cụ phải hỗ trợ mở rộng số lượng site > 6.

---

### 4.4. Tính tự động hoá

**NFR-04:** Công cụ phải chạy được ở chế độ non-interactive.

---

### 4.5. Khả năng triển khai

**NFR-06:** Công cụ chỉ phụ thuộc `minio mc` và shell runtime.

---

## 5. Ràng buộc & giả định

* Thời gian hệ thống giữa các site được đồng bộ (NTP).
* Quyền truy cập bucket đã được cấp cho mc.
* Không can thiệp cấu hình MinIO trong quá trình test.

---

## 6. Tiêu chí chấp nhận

* Công cụ thực thi được toàn bộ testcase PUT đã định nghĩa.
* Kết quả replication có thể đo, ghi log và export.
* Không phát hiện inconsistency khi hệ thống hoạt động bình thường.

---

**Tài liệu SRS này là cơ sở để thiết kế kiến trúc, viết script và triển khai công cụ test tự động.**
