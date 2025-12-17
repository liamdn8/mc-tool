# Thông tin chung

Tên tính năng: Công cụ thực hiện Profiling MinIO (MinIO Profiler Tool)
Mục tiêu: Hỗ trợ thu thập CPU/Heap/Goroutine/Block/Mutex profile từ MinIO trong môi trường production mà không cần debug mode hoặc thay đổi cấu hình volume.

# Mục đích

Tự động hóa quy trình chạy profiling MinIO.
Giảm thời gian thao tác thủ công bằng mc.
Dễ dàng tích hợp vào runbook vận hành.
Mô phỏng chính xác cơ chế gọi API profiling mà mc version 2021 (đã cài mc21) sử dụng.

# Các bước thực hiện manual

1. chạy profile start

```bash
mc21 --config-dir ~/.mc admin profile start <alias> --insecure --type <types>
```

trong đó: `--type value` bao gồm các option lựa chọn: 'cpu', 'mem', 'block', 'mutex', 'trace', 'threads' và 'goroutines' (default: "cpu,mem,block,goroutines")

2. chờ X giây

3. chạy profile stop để download file profile.zip

```bash
cd /tmp
mc21 --config-dir ~/.mc admin profile stop <alias> --insecure
```

output: 
```bash
mc21: Profile data successfully downloaded as profile.zip
```

4. unzip kết quả profile

```bash
cd /tmp
unzip profile.zip
```

output:
```bash
Archive:  profile.zip
  inflating: cluster.info            
  inflating: profile-127.0.0.1:9000-mem.pprof  
  inflating: profile-127.0.0.1:9000-mem-before.pprof  
  inflating: profile-127.0.0.1:9000-block.pprof  
  inflating: profile-127.0.0.1:9000-goroutines.txt  
  inflating: profile-127.0.0.1:9000-goroutines-before.txt  
  inflating: profile-127.0.0.1:9000-goroutines-before,debug=2.txt  
  inflating: profile-127.0.0.1:9000-cpu.pprof 
  ....
```

5. thực hiện pprof

```bash
go tool pprof <file>
```

trong đó file bao gồm:
- memory: file có đuôi mem.pprof
- cpu: file có đuôi cpu.pprof

# Yêu cầu

mc-tool hỗ trợ thêm tính năng profile giúp thực hiện các thao tác trong phần manual trong 1 command
- hỗ trợ option --duration để điều chỉnh thời gian sleep tự động
- hỗ trợ option --insecure
- hỗ trợ option --out để chỉ định được thư mục output của profile
- hỗ trợ thực hiện pprof

Sau khi hoàn thành lấy profile, mc-tool output ra các copy-paste các command để thực hiện go tool pprof

mỗi lần thực thi profile của mc-tool giải nén profile.zip ra các thư mục khác nhau theo thời gian thực hiện
