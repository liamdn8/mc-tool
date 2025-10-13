// Internationalization
const translations = {
    en: {
        dashboard: 'Dashboard',
        compare: 'Compare',
        analyze: 'Analyze',
        profile: 'Profile',
        checklist: 'Checklist',
        dashboard_title: 'Dashboard',
        compare_title: 'Compare Buckets',
        analyze_title: 'Analyze Bucket',
        profile_title: 'Profile MinIO Server',
        checklist_title: 'Bucket Checklist',
        refresh: 'Refresh',
        mc_status: 'MC Status',
        aliases: 'Aliases',
        active_jobs: 'Active Jobs',
        configured_aliases: 'Configured Aliases',
        loading: 'Loading...',
        source_path: 'Source Path',
        destination_path: 'Destination Path',
        source_help: 'Example: minio1/bucket1/folder',
        destination_help: 'Example: minio2/bucket2/folder',
        recursive: 'Recursive',
        start_comparison: 'Start Comparison',
        select_alias: 'Select Alias',
        select_bucket: 'Select Bucket',
        prefix_optional: 'Prefix (Optional)',
        start_analysis: 'Start Analysis',
        profile_type: 'Profile Type',
        duration: 'Duration',
        duration_help: 'Example: 30s, 1m, 5m',
        detect_memory_leaks: 'Detect Memory Leaks',
        monitor_interval: 'Monitor Interval',
        threshold_mb: 'Threshold (MB)',
        start_profiling: 'Start Profiling',
        run_checklist: 'Run Checklist',
        job_status: 'Job Status',
        job_id: 'Job ID',
        status: 'Status',
        progress: 'Progress',
        message: 'Message',
        configured: 'Configured',
        not_configured: 'Not Configured',
        buckets: 'Buckets',
        objects: 'Objects',
        source: 'Source',
        destination: 'Destination',
        path_optional: 'Path (Optional)',
        path_help: 'Leave empty for root, or enter path like folder/subfolder',
        only_in_source: 'Only in Source',
        only_in_destination: 'Only in Destination',
        different: 'Different',
        identical: 'Identical',
        total_objects: 'Total Objects',
        total_size: 'Total Size',
        summary: 'Summary',
        results: 'Results',
        error: 'Error',
        success: 'Success',
        failed: 'Failed',
        completed: 'Completed',
        running: 'Running',
        pending: 'Pending',
    },
    vi: {
        dashboard: 'Tổng quan',
        compare: 'So sánh',
        analyze: 'Phân tích',
        profile: 'Phân tích hiệu năng',
        checklist: 'Kiểm tra',
        dashboard_title: 'Tổng quan',
        compare_title: 'So sánh Bucket',
        analyze_title: 'Phân tích Bucket',
        profile_title: 'Phân tích hiệu năng MinIO',
        checklist_title: 'Kiểm tra cấu hình Bucket',
        refresh: 'Làm mới',
        mc_status: 'Trạng thái MC',
        aliases: 'Alias',
        active_jobs: 'Công việc đang chạy',
        configured_aliases: 'Alias đã cấu hình',
        loading: 'Đang tải...',
        source_path: 'Đường dẫn nguồn',
        destination_path: 'Đường dẫn đích',
        source_help: 'Ví dụ: minio1/bucket1/folder',
        destination_help: 'Ví dụ: minio2/bucket2/folder',
        recursive: 'Đệ quy',
        start_comparison: 'Bắt đầu so sánh',
        select_alias: 'Chọn Alias',
        select_bucket: 'Chọn Bucket',
        prefix_optional: 'Tiền tố (Tùy chọn)',
        start_analysis: 'Bắt đầu phân tích',
        profile_type: 'Loại phân tích',
        duration: 'Thời gian',
        duration_help: 'Ví dụ: 30s, 1m, 5m',
        detect_memory_leaks: 'Phát hiện rò rỉ bộ nhớ',
        monitor_interval: 'Khoảng thời gian giám sát',
        threshold_mb: 'Ngưỡng (MB)',
        start_profiling: 'Bắt đầu phân tích',
        run_checklist: 'Chạy kiểm tra',
        job_status: 'Trạng thái công việc',
        job_id: 'Mã công việc',
        status: 'Trạng thái',
        progress: 'Tiến độ',
        message: 'Thông báo',
        configured: 'Đã cấu hình',
        not_configured: 'Chưa cấu hình',
        buckets: 'Bucket',
        objects: 'Đối tượng',
        source: 'Nguồn',
        destination: 'Đích',
        path_optional: 'Đường dẫn (Tùy chọn)',
        path_help: 'Để trống cho thư mục gốc, hoặc nhập đường dẫn như folder/subfolder',
        only_in_source: 'Chỉ có ở nguồn',
        only_in_destination: 'Chỉ có ở đích',
        different: 'Khác nhau',
        identical: 'Giống nhau',
        total_objects: 'Tổng số đối tượng',
        total_size: 'Tổng dung lượng',
        summary: 'Tóm tắt',
        results: 'Kết quả',
        error: 'Lỗi',
        success: 'Thành công',
        failed: 'Thất bại',
        completed: 'Hoàn thành',
        running: 'Đang chạy',
        pending: 'Chờ xử lý',
    }
};

// Current language
let currentLang = localStorage.getItem('language') || 'en';

// API Base URL
const API_BASE = '/api';

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    initializeLanguage();
    initializeNavigation();
    initializeRefresh();
    initializeForms();
    loadDashboard();
    loadAliases();
});

// Language Management
function initializeLanguage() {
    const selector = document.getElementById('languageSelector');
    selector.value = currentLang;
    selector.addEventListener('change', (e) => {
        currentLang = e.target.value;
        localStorage.setItem('language', currentLang);
        updateLanguage();
    });
    updateLanguage();
}

function updateLanguage() {
    document.querySelectorAll('[data-i18n]').forEach(element => {
        const key = element.getAttribute('data-i18n');
        if (translations[currentLang][key]) {
            if (element.tagName === 'INPUT' && element.type === 'submit') {
                element.value = translations[currentLang][key];
            } else {
                element.textContent = translations[currentLang][key];
            }
        }
    });
}

// Utility functions
function formatNumber(num) {
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Navigation
function initializeNavigation() {
    const navItems = document.querySelectorAll('.nav-item');
    navItems.forEach(item => {
        item.addEventListener('click', () => {
            const page = item.getAttribute('data-page');
            switchPage(page);
        });
    });
}

function switchPage(pageName) {
    // Update navigation
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
        if (item.getAttribute('data-page') === pageName) {
            item.classList.add('active');
        }
    });

    // Update pages
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
    });
    document.getElementById(`${pageName}-page`).classList.add('active');

    // Load page-specific data
    if (pageName === 'analyze' || pageName === 'profile' || pageName === 'checklist') {
        loadAliasesForSelect(`${pageName}Alias`);
    }
    
    if (pageName === 'compare') {
        loadAliasesForSelect('compareSourceAlias');
        loadAliasesForSelect('compareDestAlias');
    }
}

// Refresh
function initializeRefresh() {
    document.getElementById('refreshBtn').addEventListener('click', () => {
        loadDashboard();
        loadAliases();
    });
}

// Dashboard
async function loadDashboard() {
    try {
        // Check MC status
        const mcConfig = await fetch(`${API_BASE}/mc-config`).then(r => r.json());
        const mcStatusEl = document.getElementById('mcStatus');
        if (mcConfig.configured) {
            mcStatusEl.textContent = translations[currentLang].configured;
            mcStatusEl.className = 'status-value';
            mcStatusEl.style.color = 'var(--success-color)';
        } else {
            mcStatusEl.textContent = translations[currentLang].not_configured;
            mcStatusEl.style.color = 'var(--danger-color)';
        }

        // Get aliases count
        const aliasesData = await fetch(`${API_BASE}/aliases`).then(r => r.json());
        document.getElementById('aliasCount').textContent = aliasesData.aliases?.length || 0;

    } catch (error) {
        console.error('Failed to load dashboard:', error);
        showError('Failed to load dashboard data');
    }
}

// Aliases
async function loadAliases() {
    const aliasesList = document.getElementById('aliasesList');
    aliasesList.innerHTML = `<p class="loading">${translations[currentLang].loading}</p>`;

    try {
        // Load alias list with detailed info
        const response = await fetch(`${API_BASE}/aliases`);
        const data = await response.json();

        if (data.aliases && data.aliases.length > 0) {
            // Display aliases immediately
            aliasesList.innerHTML = data.aliases.map(alias => {
                return `
                <div class="alias-item" id="alias-${alias.name}" onclick="openBucketModal('${alias.name}')" style="cursor: pointer;">
                    <div class="alias-header">
                        <h4>${alias.name}</h4>
                        <span class="alias-status checking">Checking...</span>
                    </div>
                    <p class="alias-url">${alias.url}</p>
                    <div class="alias-details">
                        <div class="detail-item">
                            <span class="detail-label">Buckets:</span>
                            <span class="detail-value bucket-count">-</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">API:</span>
                            <span class="detail-value">${alias.api || 'S3v4'}</span>
                        </div>
                    </div>
                </div>
            `}).join('');

            // Check health and load bucket count for each alias
            data.aliases.forEach(alias => {
                checkAliasHealth(alias.name, alias.url);
                loadBucketCount(alias.name);
            });
        } else {
            aliasesList.innerHTML = `<p class="loading">No aliases configured</p>`;
        }
    } catch (error) {
        console.error('Failed to load aliases:', error);
        aliasesList.innerHTML = `<p class="loading" style="color: var(--danger-color)">Failed to load aliases: ${error.message}</p>`;
    }
}

async function loadBucketCount(aliasName) {
    const aliasElement = document.getElementById(`alias-${aliasName}`);
    if (!aliasElement) return;

    const bucketCountElement = aliasElement.querySelector('.bucket-count');
    
    try {
        const response = await fetch(`${API_BASE}/buckets?alias=${aliasName}`);
        const data = await response.json();
        
        const count = data.buckets ? data.buckets.length : 0;
        bucketCountElement.textContent = count;
        bucketCountElement.style.color = 'var(--text-primary)';
    } catch (error) {
        console.error(`Failed to load bucket count for ${aliasName}:`, error);
        bucketCountElement.textContent = 'Error';
        bucketCountElement.style.color = 'var(--danger-color)';
        bucketCountElement.style.fontSize = '0.75rem';
    }
}

async function checkAliasHealth(aliasName, aliasUrl) {
    const aliasElement = document.getElementById(`alias-${aliasName}`);
    if (!aliasElement) return;

    const statusElement = aliasElement.querySelector('.alias-status');
    
    try {
        // Use mc admin info to check if alias is reachable
        const response = await fetch(`${API_BASE}/alias-health?alias=${aliasName}`);
        const data = await response.json();
        
        if (data.healthy) {
            statusElement.className = 'alias-status online';
            statusElement.textContent = 'Online';
        } else {
            statusElement.className = 'alias-status offline';
            statusElement.textContent = 'Offline';
        }
    } catch (error) {
        console.error(`Failed to check health for ${aliasName}:`, error);
        statusElement.className = 'alias-status error';
        statusElement.textContent = 'Unknown';
    }
}

async function loadAliasesForSelect(selectId) {
    const select = document.getElementById(selectId);
    try {
        const response = await fetch(`${API_BASE}/aliases`);
        const data = await response.json();

        select.innerHTML = '<option value="">-- Select Alias --</option>';
        if (data.aliases) {
            data.aliases.forEach(alias => {
                const option = document.createElement('option');
                option.value = alias.name;
                option.textContent = `${alias.name} (${alias.url})`;
                select.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Failed to load aliases:', error);
    }
}

// Load buckets when alias is selected
async function loadBucketsForSelect(alias, selectId) {
    const select = document.getElementById(selectId);
    select.innerHTML = '<option value="">Loading...</option>';

    try {
        const response = await fetch(`${API_BASE}/buckets?alias=${alias}`);
        const data = await response.json();

        select.innerHTML = '<option value="">-- Select Bucket --</option>';
        if (data.buckets) {
            data.buckets.forEach(bucket => {
                const option = document.createElement('option');
                option.value = bucket;
                option.textContent = bucket;
                select.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Failed to load buckets:', error);
        select.innerHTML = '<option value="">Error loading buckets</option>';
    }
}

// Forms
function initializeForms() {
    // Compare Form
    document.getElementById('compareForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const sourceAlias = document.getElementById('compareSourceAlias').value;
        const sourceBucket = document.getElementById('compareSourceBucket').value;
        const sourcePath = document.getElementById('compareSourcePath').value;
        
        const destAlias = document.getElementById('compareDestAlias').value;
        const destBucket = document.getElementById('compareDestBucket').value;
        const destPath = document.getElementById('compareDestPath').value;
        
        // Build full paths
        const source = sourcePath 
            ? `${sourceAlias}/${sourceBucket}/${sourcePath}`
            : `${sourceAlias}/${sourceBucket}`;
            
        const destination = destPath
            ? `${destAlias}/${destBucket}/${destPath}`
            : `${destAlias}/${destBucket}`;
        
        const recursive = document.getElementById('compareRecursive').checked;

        await submitCompare(source, destination, recursive);
    });

    // Analyze Form
    document.getElementById('analyzeForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const alias = document.getElementById('analyzeAlias').value;
        const bucket = document.getElementById('analyzeBucket').value;
        const prefix = document.getElementById('analyzePrefix').value;

        await submitAnalyze(alias, bucket, prefix);
    });

    // Profile Form
    document.getElementById('profileForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = {
            alias: document.getElementById('profileAlias').value,
            profile_type: document.getElementById('profileType').value,
            duration: document.getElementById('profileDuration').value,
            detect_leaks: document.getElementById('profileDetectLeaks').checked,
            monitor_interval: document.getElementById('profileMonitorInterval').value,
            threshold_mb: parseInt(document.getElementById('profileThreshold').value)
        };

        await submitProfile(formData);
    });

    // Checklist Form
    document.getElementById('checklistForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const alias = document.getElementById('checklistAlias').value;
        const bucket = document.getElementById('checklistBucket').value;

        await submitChecklist(alias, bucket);
    });

    // Alias change handlers
    document.getElementById('analyzeAlias').addEventListener('change', (e) => {
        if (e.target.value) {
            loadBucketsForSelect(e.target.value, 'analyzeBucket');
        }
    });

    document.getElementById('checklistAlias').addEventListener('change', (e) => {
        if (e.target.value) {
            loadBucketsForSelect(e.target.value, 'checklistBucket');
        }
    });

    document.getElementById('compareSourceAlias').addEventListener('change', (e) => {
        if (e.target.value) {
            loadBucketsForSelect(e.target.value, 'compareSourceBucket');
        }
    });

    document.getElementById('compareDestAlias').addEventListener('change', (e) => {
        if (e.target.value) {
            loadBucketsForSelect(e.target.value, 'compareDestBucket');
        }
    });

    // Profile leak detection toggle
    document.getElementById('profileDetectLeaks').addEventListener('change', (e) => {
        document.getElementById('leakDetectionOptions').style.display = 
            e.target.checked ? 'block' : 'none';
    });

    // Modal close
    document.querySelector('.modal-close').addEventListener('click', () => {
        document.getElementById('jobModal').classList.remove('active');
    });
}

// Submit handlers
async function submitCompare(source, destination, recursive) {
    try {
        const response = await fetch(`${API_BASE}/compare`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ source, destination, recursive })
        });

        const data = await response.json();
        if (data.job_id) {
            showJobModal(data.job_id);
            pollJobStatus(data.job_id, 'compareResults');
        }
    } catch (error) {
        showError('Failed to start comparison: ' + error.message);
    }
}

async function submitAnalyze(alias, bucket, prefix) {
    try {
        const response = await fetch(`${API_BASE}/analyze`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ alias, bucket, prefix })
        });

        const data = await response.json();
        if (data.job_id) {
            showJobModal(data.job_id);
            pollJobStatus(data.job_id, 'analyzeResults');
        }
    } catch (error) {
        showError('Failed to start analysis: ' + error.message);
    }
}

async function submitProfile(formData) {
    try {
        const response = await fetch(`${API_BASE}/profile`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });

        const data = await response.json();
        if (data.job_id) {
            showJobModal(data.job_id);
            pollJobStatus(data.job_id, 'profileResults');
        }
    } catch (error) {
        showError('Failed to start profiling: ' + error.message);
    }
}

async function submitChecklist(alias, bucket) {
    try {
        const response = await fetch(`${API_BASE}/checklist`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ alias, bucket })
        });

        const data = await response.json();
        if (data.job_id) {
            showJobModal(data.job_id);
            pollJobStatus(data.job_id, 'checklistResults');
        }
    } catch (error) {
        showError('Failed to start checklist: ' + error.message);
    }
}

// Job Management
function showJobModal(jobId) {
    const modal = document.getElementById('jobModal');
    modal.classList.add('active');
    document.getElementById('modalJobId').textContent = jobId;
}

async function pollJobStatus(jobId, resultsElementId) {
    const pollInterval = setInterval(async () => {
        try {
            const response = await fetch(`${API_BASE}/jobs/${jobId}`);
            const job = await response.json();

            updateJobModal(job);

            if (job.status === 'completed' || job.status === 'failed') {
                clearInterval(pollInterval);
                displayResults(job, resultsElementId);
            }
        } catch (error) {
            clearInterval(pollInterval);
            showError('Failed to get job status: ' + error.message);
        }
    }, 1000);
}

function updateJobModal(job) {
    document.getElementById('modalJobStatus').textContent = job.status;
    document.getElementById('modalJobStatus').className = `badge badge-${job.status}`;
    document.getElementById('modalJobProgress').textContent = job.progress;
    document.getElementById('modalProgressBar').style.width = `${job.progress}%`;
    document.getElementById('modalJobMessage').textContent = job.message || '';

    if (job.output && job.output.length > 0) {
        document.getElementById('modalJobOutput').innerHTML = 
            job.output.map(line => `<div>${escapeHtml(line)}</div>`).join('');
    }

    if (job.result) {
        document.getElementById('modalJobResult').innerHTML = 
            `<pre>${JSON.stringify(job.result, null, 2)}</pre>`;
    }

    if (job.error) {
        document.getElementById('modalJobResult').innerHTML = 
            `<div class="alert alert-error">${escapeHtml(job.error)}</div>`;
    }
}

function displayResults(job, resultsElementId) {
    const resultsEl = document.getElementById(resultsElementId);
    resultsEl.classList.remove('empty');

    if (job.status === 'failed') {
        resultsEl.innerHTML = `
            <div class="alert alert-error">
                <strong>${translations[currentLang].error}:</strong> ${escapeHtml(job.error)}
            </div>
        `;
        return;
    }

    if (!job.result) {
        resultsEl.innerHTML = `
            <div class="alert alert-info">
                ${translations[currentLang].completed}
            </div>
        `;
        return;
    }

    // Format results based on job type
    if (job.type === 'compare') {
        displayCompareResults(job.result, resultsEl);
    } else if (job.type === 'analyze') {
        displayAnalyzeResults(job.result, resultsEl);
    } else if (job.type === 'profile') {
        displayProfileResults(job.result, resultsEl);
    } else if (job.type === 'checklist') {
        displayChecklistResults(job.result, resultsEl);
    } else {
        // Generic result display
        if (job.result.output && typeof job.result.output === 'string') {
            resultsEl.innerHTML = `
                <div class="alert alert-success">
                    <strong>${translations[currentLang].success}:</strong> ${translations[currentLang].completed}
                </div>
                <div class="results-output">
                    <pre>${escapeHtml(job.result.output)}</pre>
                </div>
            `;
        } else {
            resultsEl.innerHTML = `<pre>${JSON.stringify(job.result, null, 2)}</pre>`;
        }
    }
}

function displayCompareResults(result, element) {
    // Check if we have output text to display
    if (result.output && typeof result.output === 'string') {
        element.innerHTML = `
            <div class="alert alert-success">
                <strong>${translations[currentLang].success}:</strong> ${translations[currentLang].completed}
            </div>
            <div class="results-output">
                <pre>${escapeHtml(result.output)}</pre>
            </div>
        `;
    } else {
        const summary = result.summary || {};
        element.innerHTML = `
            <div class="alert alert-success">
                <strong>${translations[currentLang].success}:</strong> ${translations[currentLang].completed}
            </div>
            <div class="stats-grid">
                <div class="stat-item">
                    <h4>${translations[currentLang].only_in_source}</h4>
                    <p>${summary.only_in_source || 0}</p>
                </div>
                <div class="stat-item">
                    <h4>${translations[currentLang].only_in_destination}</h4>
                    <p>${summary.only_in_destination || 0}</p>
                </div>
                <div class="stat-item">
                    <h4>${translations[currentLang].different}</h4>
                    <p>${summary.different || 0}</p>
                </div>
                <div class="stat-item">
                    <h4>${translations[currentLang].identical}</h4>
                    <p>${summary.identical || 0}</p>
                </div>
            </div>
            <details style="margin-top: 1rem;">
                <summary style="cursor: pointer; font-weight: 600;">View Detailed Results</summary>
                <pre style="margin-top: 1rem;">${JSON.stringify(result, null, 2)}</pre>
            </details>
        `;
    }
}

function displayAnalyzeResults(result, element) {
    // Check if we have output text to display
    if (result.output && typeof result.output === 'string') {
        element.innerHTML = `
            <div class="alert alert-success">
                <strong>${translations[currentLang].success}:</strong> Analysis completed
            </div>
            <div class="results-output">
                <pre>${escapeHtml(result.output)}</pre>
            </div>
        `;
    } else {
        element.innerHTML = `
            <div class="alert alert-success">
                <strong>${translations[currentLang].success}:</strong> Analysis completed
            </div>
            <div class="stats-grid">
                <div class="stat-item">
                    <h4>${translations[currentLang].total_objects}</h4>
                    <p>${result.total_objects || 0}</p>
                </div>
                <div class="stat-item">
                    <h4>${translations[currentLang].total_size}</h4>
                    <p>${formatBytes(result.total_size || 0)}</p>
                </div>
            </div>
            <details style="margin-top: 1rem;">
                <summary style="cursor: pointer; font-weight: 600;">View Detailed Results</summary>
                <pre style="margin-top: 1rem;">${JSON.stringify(result, null, 2)}</pre>
            </details>
        `;
    }
}

function displayProfileResults(result, element) {
    if (result.output && typeof result.output === 'string') {
        element.innerHTML = `
            <div class="alert alert-success">
                <strong>${translations[currentLang].success}:</strong> Profiling completed
            </div>
            <div class="results-output">
                <pre>${escapeHtml(result.output)}</pre>
            </div>
        `;
    } else {
        element.innerHTML = `
            <div class="alert alert-success">
                <strong>${translations[currentLang].success}:</strong> Profiling completed
            </div>
            <pre style="margin-top: 1rem;">${JSON.stringify(result, null, 2)}</pre>
        `;
    }
}

function displayChecklistResults(result, element) {
    if (result.output && typeof result.output === 'string') {
        element.innerHTML = `
            <div class="alert alert-success">
                <strong>${translations[currentLang].success}:</strong> Checklist completed
            </div>
            <div class="results-output">
                <pre>${escapeHtml(result.output)}</pre>
            </div>
        `;
    } else {
        element.innerHTML = `
            <div class="alert alert-success">
                <strong>${translations[currentLang].success}:</strong> Checklist completed
            </div>
            <pre style="margin-top: 1rem;">${JSON.stringify(result, null, 2)}</pre>
        `;
    }
}

// Utilities
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showError(message) {
    alert(message); // Simple alert for now, can be enhanced with better UI
}

// Bucket Modal
let currentBucketAlias = '';
let currentBucketPage = 1;
let totalBuckets = 0;
const bucketsPerPage = 10;

function openBucketModal(aliasName) {
    currentBucketAlias = aliasName;
    currentBucketPage = 1;
    
    const modal = document.getElementById('bucketModal');
    document.getElementById('bucketModalTitle').textContent = `${aliasName} - ${translations[currentLang].buckets}`;
    modal.classList.add('active');
    
    loadBucketDetails(aliasName, 1);
}

function closeBucketModal() {
    document.getElementById('bucketModal').classList.remove('active');
    currentBucketAlias = '';
    currentBucketPage = 1;
}

async function loadBucketDetails(aliasName, page) {
    const container = document.getElementById('bucketsList');
    container.innerHTML = `<div style="text-align: center; padding: 2rem;">${translations[currentLang].loading}</div>`;
    
    try {
        const response = await fetch(`/api/buckets?alias=${encodeURIComponent(aliasName)}`);
        if (!response.ok) throw new Error('Failed to load buckets');
        
        const data = await response.json();
        const buckets = data.buckets || [];
        totalBuckets = buckets.length;
        
        const startIndex = (page - 1) * bucketsPerPage;
        const endIndex = startIndex + bucketsPerPage;
        const pageBuckets = buckets.slice(startIndex, endIndex);
        
        if (pageBuckets.length === 0) {
            container.innerHTML = `<div style="text-align: center; padding: 2rem; color: var(--text-secondary);">No buckets found</div>`;
            renderBucketPagination();
            return;
        }
        
        container.innerHTML = '';
        pageBuckets.forEach(bucket => {
            const card = createBucketCard(aliasName, bucket);
            container.appendChild(card);
        });
        
        renderBucketPagination();
        
        // Load object counts and checklist status asynchronously
        pageBuckets.forEach(bucket => {
            loadBucketObjectCount(aliasName, bucket);
            loadBucketChecklistStatus(aliasName, bucket);
        });
        
    } catch (error) {
        container.innerHTML = `<div style="text-align: center; padding: 2rem; color: var(--danger-color);">Error: ${error.message}</div>`;
    }
}

function createBucketCard(aliasName, bucketName) {
    const card = document.createElement('div');
    card.className = 'bucket-card';
    card.id = `bucket-${bucketName}`;
    
    card.innerHTML = `
        <div class="bucket-card-layout">
            <div class="bucket-info-section">
                <div class="bucket-name-large">${bucketName}</div>
                <div class="bucket-meta">
                    <div class="meta-item">
                        <span class="meta-label">Objects:</span>
                        <span class="meta-value" id="bucket-${bucketName}-objects">...</span>
                    </div>
                </div>
            </div>
            <div class="bucket-checklist-section" id="bucket-${bucketName}-details">
                <div class="checklist-placeholder">
                    <span class="checklist-badge checking">Checking...</span>
                </div>
            </div>
        </div>
    `;
    
    return card;
}

async function loadBucketObjectCount(aliasName, bucketName) {
    try {
        const response = await fetch(`/api/bucket-stats?alias=${encodeURIComponent(aliasName)}&bucket=${encodeURIComponent(bucketName)}`);
        if (!response.ok) throw new Error('Failed');
        
        const stats = await response.json();
        const element = document.getElementById(`bucket-${bucketName}-objects`);
        if (element) {
            const count = stats.objects || 0;
            element.innerHTML = `<strong>${count.toLocaleString()}</strong>`;
        }
    } catch (error) {
        const element = document.getElementById(`bucket-${bucketName}-objects`);
        if (element) {
            element.innerHTML = '<span style="color: var(--danger-color);">Error</span>';
        }
    }
}

async function loadBucketChecklistStatus(aliasName, bucketName) {
    try {
        const response = await fetch('/api/checklist', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                alias: aliasName,
                bucket: bucketName
            })
        });
        
        if (!response.ok) throw new Error('Failed');
        
        const data = await response.json();
        const detailsElement = document.getElementById(`bucket-${bucketName}-details`);
        
        if (detailsElement && data.job_id) {
            // Poll for results
            pollChecklistResult(data.job_id, bucketName);
        }
    } catch (error) {
        const detailsElement = document.getElementById(`bucket-${bucketName}-details`);
        if (detailsElement) {
            detailsElement.innerHTML = '<div class="checklist-section"><span class="checklist-badge failed">Error loading checklist</span></div>';
        }
    }
}

async function pollChecklistResult(jobId, bucketName) {
    const maxAttempts = 30;
    let attempts = 0;
    
    const poll = async () => {
        if (attempts >= maxAttempts) {
            const detailsElement = document.getElementById(`bucket-${bucketName}-details`);
            if (detailsElement) {
                detailsElement.innerHTML = '<div class="checklist-section"><span class="checklist-badge failed">Timeout</span></div>';
            }
            return;
        }
        
        try {
            const response = await fetch(`/api/jobs/${jobId}`);
            if (!response.ok) throw new Error('Failed');
            
            const job = await response.json();
            
            if (job.status === 'completed') {
                displayChecklistDetails(bucketName, job.result);
            } else if (job.status === 'failed') {
                const detailsElement = document.getElementById(`bucket-${bucketName}-details`);
                if (detailsElement) {
                    detailsElement.innerHTML = '<div class="checklist-section"><span class="checklist-badge failed">Error</span></div>';
                }
            } else {
                attempts++;
                setTimeout(poll, 1000);
            }
        } catch (error) {
            const detailsElement = document.getElementById(`bucket-${bucketName}-details`);
            if (detailsElement) {
                detailsElement.innerHTML = '<div class="checklist-section"><span class="checklist-badge failed">Error</span></div>';
            }
        }
    };
    
    poll();
}

function displayChecklistDetails(bucketName, result) {
    const detailsElement = document.getElementById(`bucket-${bucketName}-details`);
    if (!detailsElement) return;
    
    const output = result.output || '';
    let html = '<div class="checklist-grid">';
    
    // Parse the checklist output
    const sections = output.split('===');
    
    // Event Notification Section
    const eventSection = sections.find(s => s.includes('BUCKET EVENT NOTIFICATION'));
    if (eventSection) {
        const hasEvent = eventSection.includes('✓');
        const icon = hasEvent ? '✓' : '⚠️';
        const statusClass = hasEvent ? 'status-success' : 'status-warning';
        
        html += `<div class="checklist-item">
            <div class="checklist-item-header">
                <span class="status-icon ${statusClass}">${icon}</span>
                <span class="checklist-label">Event</span>
            </div>`;
        
        if (hasEvent) {
            const arnMatch = eventSection.match(/ARN:\s*(.+)/);
            const eventsMatch = eventSection.match(/Events:\s*(.+)/);
            
            if (arnMatch) {
                const arn = arnMatch[1].trim();
                html += `<div class="checklist-value" title="${escapeHtml(arn)}">${escapeHtml(arn)}</div>`;
            }
            if (eventsMatch) {
                html += `<div class="checklist-value-small">${escapeHtml(eventsMatch[1].trim())}</div>`;
            }
        } else {
            html += '<div class="checklist-value-muted">Not configured</div>';
        }
        html += '</div>';
    }
    
    // Lifecycle Policy Section
    const lifecycleSection = sections.find(s => s.includes('BUCKET LIFECYCLE POLICY'));
    if (lifecycleSection) {
        const hasLifecycle = lifecycleSection.includes('✓');
        const icon = hasLifecycle ? '✓' : '⚠️';
        const statusClass = hasLifecycle ? 'status-success' : 'status-warning';
        
        html += `<div class="checklist-item">
            <div class="checklist-item-header">
                <span class="status-icon ${statusClass}">${icon}</span>
                <span class="checklist-label">Lifecycle</span>
            </div>`;
        
        if (hasLifecycle) {
            const expirationMatch = lifecycleSection.match(/Expiration:\s*(\d+)\s*days/);
            const noncurrentExpirationMatch = lifecycleSection.match(/Delete Noncurrent Versions After:\s*(\d+)\s*days/);
            const deleteMarkerMatch = lifecycleSection.match(/Delete Expired Object Delete Markers:\s*(.+)/);
            
            let details = [];
            if (expirationMatch) {
                details.push(`Expire: ${expirationMatch[1]}d`);
            }
            if (noncurrentExpirationMatch) {
                details.push(`Old: ${noncurrentExpirationMatch[1]}d`);
            }
            if (deleteMarkerMatch && deleteMarkerMatch[1].trim() === 'Yes') {
                details.push('Delete markers');
            }
            
            if (details.length > 0) {
                html += `<div class="checklist-value">${details.join(' • ')}</div>`;
            } else {
                html += '<div class="checklist-value">Configured</div>';
            }
        } else {
            html += '<div class="checklist-value-muted">Not configured</div>';
        }
        html += '</div>';
    }
    
    html += '</div>';
    detailsElement.innerHTML = html;
}

function renderBucketPagination() {
    const container = document.getElementById('bucketPagination');
    const totalPages = Math.ceil(totalBuckets / bucketsPerPage);
    
    if (totalPages <= 1) {
        container.innerHTML = '';
        return;
    }
    
    let html = '<div class="pagination">';
    
    // Previous button
    html += `<button class="pagination-btn" ${currentBucketPage === 1 ? 'disabled' : ''} onclick="changeBucketPage(${currentBucketPage - 1})">Previous</button>`;
    
    // Page info
    html += `<span class="pagination-info">Page ${currentBucketPage} of ${totalPages}</span>`;
    
    // Next button
    html += `<button class="pagination-btn" ${currentBucketPage === totalPages ? 'disabled' : ''} onclick="changeBucketPage(${currentBucketPage + 1})">Next</button>`;
    
    html += '</div>';
    container.innerHTML = html;
}

function changeBucketPage(page) {
    if (page < 1) return;
    const totalPages = Math.ceil(totalBuckets / bucketsPerPage);
    if (page > totalPages) return;
    
    currentBucketPage = page;
    loadBucketDetails(currentBucketAlias, page);
}
