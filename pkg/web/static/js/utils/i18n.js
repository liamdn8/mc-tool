// i18n Translations
const translations = {
    en: {
        overview: "Overview",
        sites: "Sites",
        buckets: "Buckets",
        replication: "Replication Status",
        consistency: "Consistency Check",
        operations: "Operations",
        site_replication_overview: "Site Replication Overview",
        add_site: "Add Site",
        replication_group: "Replication Group",
        total_sites: "Total Sites",
        synced_buckets: "Synced Buckets",
        total_objects: "Total Objects",
        replication_health: "Health",
        healthy: "Healthy",
        configured_aliases: "Configured MinIO Aliases",
        manage_sites: "Manage Sites",
        buckets_overview: "Buckets Overview",
        replication_status: "Replication Status",
        refresh: "Refresh",
        consistency_check: "Consistency Check",
        run_check: "Run Check",
        automated_operations: "Automated Operations",
        sync_bucket_policies: "Sync Bucket Policies",
        sync_bucket_policies_desc: "Automatically sync bucket policies across all sites",
        sync_lifecycle: "Sync Lifecycle Policies",
        sync_lifecycle_desc: "Sync ILM policies across all sites",
        validate_consistency: "Validate Consistency",
        validate_consistency_desc: "Check configuration consistency across sites",
        health_check: "Health Check",
        health_check_desc: "Verify all sites are healthy and reachable",
        execute: "Execute",
        operation_status: "Operation Status",
        replication_enabled: "Replication Enabled",
        replication_disabled: "Replication Disabled",
        not_configured: "Not Configured",
        configured: "Configured",
        alias: "Alias",
        endpoint: "Endpoint",
        status: "Status",
        servers: "Servers",
        site_replication_config: "Site Replication Configuration",
        setup_replication: "Setup Site Replication",
        setup_replication_desc: "Select aliases in order and click 'Add Sites' to create site replication cluster.",
        select_aliases: "Select Aliases (minimum 2)",
        selected_order: "Selected Order",
        no_selection: "No aliases selected",
        add_sites: "Add Sites to Replication",
        manage_replication: "Manage Site Replication", 
        manage_replication_desc: "Manage sites in your replication cluster.",
        add_sites_to_cluster: "Add Sites to Existing Cluster",
        add_to_cluster: "Add to Cluster",
        current_cluster: "Current Cluster Sites",
        remove_selected: "Remove Selected",
        resync_from: "Resync From",
        resync_to: "Resync To",
        remove: "Remove",
    },
    vi: {
        overview: "Tổng quan",
        sites: "Các Site",
        buckets: "Buckets",
        replication: "Trạng thái Replication",
        consistency: "Kiểm tra Nhất quán",
        operations: "Thao tác",
        site_replication_overview: "Tổng quan Site Replication",
        add_site: "Thêm Site",
        replication_group: "Nhóm Replication",
        total_sites: "Tổng số Site",
        synced_buckets: "Buckets đã đồng bộ",
        total_objects: "Tổng số Objects",
        replication_health: "Tình trạng",
        healthy: "Tốt",
        configured_aliases: "MinIO Aliases đã cấu hình",
        manage_sites: "Quản lý Sites",
        buckets_overview: "Tổng quan Buckets",
        replication_status: "Trạng thái Replication",
        refresh: "Làm mới",
        consistency_check: "Kiểm tra Nhất quán",
        run_check: "Chạy kiểm tra",
        automated_operations: "Thao tác Tự động",
        sync_bucket_policies: "Đồng bộ Bucket Policies",
        sync_bucket_policies_desc: "Tự động đồng bộ bucket policies trên tất cả các site",
        sync_lifecycle: "Đồng bộ Lifecycle Policies",
        sync_lifecycle_desc: "Đồng bộ ILM policies trên tất cả các site",
        validate_consistency: "Kiểm tra Nhất quán",
        validate_consistency_desc: "Kiểm tra tính nhất quán của cấu hình giữa các site",
        health_check: "Kiểm tra Sức khỏe",
        health_check_desc: "Xác minh tất cả các site đều khỏe mạnh và có thể truy cập",
        execute: "Thực thi",
        operation_status: "Trạng thái Thao tác",
        replication_enabled: "Replication đã bật",
        replication_disabled: "Replication đã tắt",
        not_configured: "Chưa cấu hình",
        configured: "Đã cấu hình",
        alias: "Alias",
        endpoint: "Endpoint",
        status: "Trạng thái",
        servers: "Servers",
        site_replication_config: "Cấu hình Site Replication",
        setup_replication: "Thiết lập Site Replication",
        setup_replication_desc: "Chọn các alias theo thứ tự và nhấn 'Thêm Sites' để tạo cluster replication.",
        select_aliases: "Chọn Aliases (tối thiểu 2)",
        selected_order: "Thứ tự đã chọn",
        no_selection: "Chưa chọn alias nào",
        add_sites: "Thêm Sites vào Replication",
        manage_replication: "Quản lý Site Replication",
        manage_replication_desc: "Quản lý các site trong cluster replication của bạn.",
        add_sites_to_cluster: "Thêm Sites vào Cluster Hiện có",
        add_to_cluster: "Thêm vào Cluster",
        current_cluster: "Sites trong Cluster Hiện tại",
        remove_selected: "Xóa Đã chọn",
        resync_from: "Resync Từ",
        resync_to: "Resync Đến",
        remove: "Xóa",
    }
};

let currentLang = 'en';

export function getCurrentLang() {
    return currentLang;
}

export function setCurrentLang(lang) {
    currentLang = lang;
    updateI18n();
}

export function translate(key) {
    return translations[currentLang][key] || key;
}

export function updateI18n() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.dataset.i18n;
        if (translations[currentLang][key]) {
            el.textContent = translations[currentLang][key];
        }
    });
}

export { translations };