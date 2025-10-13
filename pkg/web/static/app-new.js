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
        resync_from: "Resync Từ",
        resync_to: "Resync Đến",
        remove: "Xóa",
    }
};

let currentLang = 'en';
let sites = [];
let replicationInfo = null;

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initializeEventListeners();
    loadInitialData();
    updateI18n();
});

function initializeEventListeners() {
    // Language selector
    document.getElementById('languageSelector').addEventListener('change', (e) => {
        currentLang = e.target.value;
        updateI18n();
    });

    // Navigation
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const page = link.dataset.page;
            navigateToPage(page);
        });
    });

    // Refresh button
    document.getElementById('refreshBtn').addEventListener('click', () => {
        loadInitialData();
    });

    // Add site button
    document.getElementById('addSiteBtn')?.addEventListener('click', () => {
        // TODO: Show add site modal
        alert('Add Site functionality will be implemented');
    });

    // Operation buttons
    document.querySelectorAll('[data-operation]').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const operation = e.currentTarget.dataset.operation;
            executeOperation(operation);
        });
    });

    // Modal close
    document.querySelectorAll('.modal-close').forEach(btn => {
        btn.addEventListener('click', () => {
            btn.closest('.modal').classList.remove('active');
        });
    });
}

function navigateToPage(pageName) {
    // Update active nav link
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
    });
    document.querySelector(`[data-page="${pageName}"]`).classList.add('active');

    // Update active page
    document.querySelectorAll('.page').forEach(page => {
        page.classList.remove('active');
    });
    document.getElementById(`${pageName}-page`).classList.add('active');

    // Load page-specific data
    loadPageData(pageName);
}

async function loadInitialData() {
    try {
        // Load aliases (sites)
        await loadAliases();
        
        // Load site replication info
        await loadSiteReplicationInfo();
        
        // Update overview stats
        updateOverviewStats();
    } catch (error) {
        console.error('Error loading initial data:', error);
    }
}

async function loadAliases() {
    try {
        const response = await fetch('/api/aliases');
        const data = await response.json();
        sites = data.aliases || [];
        
        // Render sites list
        renderSitesList();
    } catch (error) {
        console.error('Error loading aliases:', error);
        sites = [];
    }
}

async function loadSiteReplicationInfo() {
    try {
        // Use new site replication API
        const response = await fetch('/api/replication/info');
        const data = await response.json();
        
        replicationInfo = {
            enabled: data.enabled || false,
            sites: data.totalAliases || 0,
            syncedBuckets: 0,
            totalObjects: 0,
            health: 'healthy',
            replicationGroup: data.replicationGroup || null
        };
        
        // Update sites list from aliases
        if (data.aliases && data.aliases.length > 0) {
            sites = data.aliases.map(aliasData => ({
                alias: aliasData.alias,
                url: aliasData.url || aliasData.endpoint, // Backend returns 'url' field
                healthy: aliasData.healthy === true,
                replicationEnabled: aliasData.replicationEnabled === true,
                replicationStatus: aliasData.replicationStatus || 'not_configured',
                siteName: aliasData.siteName || '',
                deploymentID: aliasData.deploymentID || '',
                serverCount: aliasData.serverCount || 0
            }));
        }
    } catch (error) {
        console.error('Error loading site replication info:', error);
    }
}

function updateOverviewStats() {
    // Count sites with replication enabled
    const replicatedSites = sites.filter(s => s.replicationEnabled).length;
    const healthySites = sites.filter(s => s.healthy).length;
    
    document.getElementById('totalSites').textContent = sites.length;
    
    // Update sites summary
    const sitesSummary = document.getElementById('sitesSummary');
    if (sitesSummary) {
        if (replicatedSites > 0) {
            sitesSummary.textContent = `${replicatedSites} in replication group`;
        } else {
            sitesSummary.textContent = 'No replication configured';
        }
    }
    
    // Get bucket count from replication info
    let totalBuckets = 0;
    if (replicationInfo && replicationInfo.replicationGroup && replicationInfo.replicationGroup.sites) {
        // Count unique buckets across all sites
        const bucketSet = new Set();
        replicationInfo.replicationGroup.sites.forEach(site => {
            if (site.buckets) {
                site.buckets.forEach(bucket => bucketSet.add(bucket));
            }
        });
        totalBuckets = bucketSet.size;
    }
    
    document.getElementById('syncedBuckets').textContent = totalBuckets;
    
    // Update buckets summary
    const bucketsSummary = document.getElementById('bucketsSummary');
    if (bucketsSummary) {
        if (totalBuckets > 0) {
            bucketsSummary.textContent = `Across ${replicatedSites} sites`;
        } else {
            bucketsSummary.textContent = 'No buckets synced';
        }
    }
    
    // Calculate total objects and size from sites
    let totalObjects = 0;
    let totalSize = 0;
    sites.forEach(site => {
        if (site.totalObjects) totalObjects += site.totalObjects;
        if (site.totalSize) totalSize += site.totalSize;
    });
    
    document.getElementById('totalObjects').textContent = formatNumber(totalObjects);
    
    // Update total size
    const totalSizeEl = document.getElementById('totalSize');
    if (totalSizeEl) {
        totalSizeEl.textContent = formatBytes(totalSize);
    }
    
    // Update health indicator
    const healthIndicator = document.getElementById('healthIndicator');
    const healthSummary = document.getElementById('healthSummary');
    
    if (sites.length === 0) {
        healthIndicator.innerHTML = `
            <span class="status-offline"></span>
            <span data-i18n="not_configured">Not Configured</span>
        `;
        if (healthSummary) healthSummary.textContent = 'No sites configured';
    } else if (healthySites === sites.length) {
        healthIndicator.innerHTML = `
            <span class="pulse"></span>
            <span data-i18n="healthy">Healthy</span>
        `;
        if (healthSummary) healthSummary.textContent = `All ${sites.length} sites online`;
    } else if (healthySites > 0) {
        healthIndicator.innerHTML = `
            <span class="status-warning"></span>
            <span>Degraded</span>
        `;
        if (healthSummary) healthSummary.textContent = `${healthySites}/${sites.length} sites online`;
    } else {
        healthIndicator.innerHTML = `
            <span class="status-offline"></span>
            <span>Offline</span>
        `;
        if (healthSummary) healthSummary.textContent = 'All sites offline';
    }
    
    // Update group status badge
    const groupStatus = document.getElementById('groupStatus');
    if (groupStatus) {
        if (replicatedSites >= 2) {
            groupStatus.className = 'badge badge-success';
            groupStatus.textContent = 'Active';
        } else {
            groupStatus.className = 'badge badge-warning';
            groupStatus.textContent = 'Not Configured';
        }
    }
    
    // Update replication details
    const replicationDetails = document.getElementById('replicationDetails');
    if (replicationDetails && replicationInfo && replicationInfo.replicationGroup) {
        replicationDetails.style.display = 'block';
        
        const serviceAccount = document.getElementById('serviceAccount');
        if (serviceAccount) {
            serviceAccount.textContent = replicationInfo.replicationGroup.serviceAccountAccessKey || '-';
        }
        
        const sitesInGroup = document.getElementById('sitesInGroup');
        if (sitesInGroup && replicationInfo.replicationGroup.sites) {
            const siteNames = replicationInfo.replicationGroup.sites.map(s => s.name).join(', ');
            sitesInGroup.textContent = siteNames;
        }
    } else if (replicationDetails) {
        replicationDetails.style.display = 'none';
    }
}

// Format bytes to human readable
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
}

function renderSitesList() {
    const container = document.getElementById('sitesList');
    
    if (sites.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="2" y1="12" x2="22" y2="12"></line>
                    <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path>
                </svg>
                <h3>No MinIO Aliases Configured</h3>
                <p>Add MinIO aliases using mc command to get started</p>
            </div>
        `;
        return;
    }
    
    container.innerHTML = sites.map(site => {
        console.log(site);
        // Determine status badge
        // Determine status badge based on replication status
        let statusBadge = '';
        let statusClass = '';
        
        if (site.replicationEnabled === true && site.replicationStatus === 'configured') {
            statusBadge = `<span class="badge badge-success">✓ Replicated</span>`;
            statusClass = 'site-card-replicated';
        } else if (site.replicationStatus === 'disabled') {
            statusBadge = `<span class="badge badge-warning">⚠ Disabled</span>`;
            statusClass = 'site-card-warning';
        } else {
            statusBadge = `<span class="badge badge-info">○ Standalone</span>`;
            statusClass = 'site-card-not-configured';
        }
        
        // Health badge - giống Replication Status
        const healthBadge = site.healthy === true
            ? `<span class="badge badge-success">✓ Healthy</span>`
            : `<span class="badge badge-danger">✗ Unhealthy</span>`;

        // Replication status indicator
        const replicationStatusBadge = site.replicationEnabled === true
            ? `<span class="badge badge-success">✓ Active</span>`
            : `<span class="badge badge-secondary">○ Not Replicated</span>`;
        
        return `
            <div class="site-card ${statusClass}" onclick="viewSiteDetails('${site.alias}')">
                <div class="site-card-header">
                    <div>
                        <div class="site-name">${site.alias}</div>
                        ${site.siteName ? `<div class="site-label">Site: ${site.siteName}</div>` : ''}
                    </div>
                    ${statusBadge}
                </div>
                <div class="site-url">${site.url}</div>
                ${site.deploymentID ? `<div class="site-deployment-id" title="${site.deploymentID}">📋 ${site.deploymentID.substring(0, 24)}...</div>` : ''}
                <div class="site-stats">
                    <div class="site-stat">
                        <div class="site-stat-label">Health</div>
                        <div class="site-stat-value">${healthBadge}</div>
                    </div>
                    <div class="site-stat">
                        <div class="site-stat-label">Buckets</div>
                        <div class="site-stat-value" id="site-buckets-${site.alias}">
                            <span class="loading-spinner-small"></span>
                        </div>
                    </div>
                    <div class="site-stat">
                        <div class="site-stat-label">Replication</div>
                        <div class="site-stat-value">${replicationStatusBadge}</div>
                    </div>
                    ${site.serverCount > 0 ? `
                    <div class="site-stat">
                        <div class="site-stat-label">Servers</div>
                        <div class="site-stat-value">${site.serverCount}</div>
                    </div>
                    ` : ''}
                </div>
            </div>
        `;
    }).join('');
    
    // Load bucket counts for each site
    sites.forEach(site => {
        loadSiteBucketCount(site.alias);
    });
}

async function loadSiteBucketCount(alias) {
    try {
        const response = await fetch(`/api/buckets?alias=${alias}`);
        const data = await response.json();
        const buckets = data.buckets || [];
        
        const bucketCountEl = document.getElementById(`site-buckets-${alias}`);
        if (bucketCountEl) {
            bucketCountEl.textContent = buckets.length;
        }
        
        // Load admin info to get total objects and size
        const infoResponse = await fetch(`/api/alias-health?alias=${alias}`);
        if (infoResponse.ok) {
            const infoData = await infoResponse.json();
            
            // Update site object with data
            const siteIndex = sites.findIndex(s => s.alias === alias);
            if (siteIndex !== -1) {
                sites[siteIndex].totalObjects = infoData.objectCount || 0;
                sites[siteIndex].totalSize = infoData.totalSize || 0;
            }
            
            const objectCountEl = document.getElementById(`site-objects-${alias}`);
            if (objectCountEl) {
                objectCountEl.textContent = formatNumber(infoData.objectCount || 0);
            }
            
            // Update overview stats after loading data
            updateOverviewStats();
        } else {
            // Fallback: calculate from buckets
            let totalObjects = 0;
            for (const bucket of buckets) {
                const statsResponse = await fetch(`/api/bucket-stats?alias=${alias}&bucket=${bucket.name}`);
                if (statsResponse.ok) {
                    const stats = await statsResponse.json();
                    totalObjects += stats.objectCount || 0;
                }
            }
            
            const objectCountEl = document.getElementById(`site-objects-${alias}`);
            if (objectCountEl) {
                objectCountEl.textContent = formatNumber(totalObjects);
            }
            
            // Update site object
            const siteIndex = sites.findIndex(s => s.alias === alias);
            if (siteIndex !== -1) {
                sites[siteIndex].totalObjects = totalObjects;
            }
            
            updateOverviewStats();
        }
    } catch (error) {
        console.error(`Error loading bucket count for ${alias}:`, error);
        const bucketCountEl = document.getElementById(`site-buckets-${alias}`);
        if (bucketCountEl) bucketCountEl.textContent = '-';
        
        const objectCountEl = document.getElementById(`site-objects-${alias}`);
        if (objectCountEl) objectCountEl.textContent = '-';
    }
}

function viewSiteDetails(alias) {
    // Navigate to buckets page and filter by site
    navigateToPage('buckets');
    // TODO: Filter buckets by selected site
}

async function loadPageData(pageName) {
    switch(pageName) {
        case 'overview':
            // Already loaded in loadInitialData
            break;
        case 'sites':
            renderSitesPage();
            break;
        case 'buckets':
            renderBucketsPage();
            break;
        case 'replication':
            renderReplicationPage();
            break;
        case 'consistency':
            renderConsistencyPage();
            break;
        case 'operations':
            // Static page, no dynamic data needed
            break;
    }
}

function renderSitesPage() {
    const container = document.getElementById('sitesContent');
    
    // Get replication info to check if already configured
    fetch('/api/replication/info')
        .then(res => res.json())
        .then(data => {
            const isConfigured = data.enabled === true;
            
            container.innerHTML = `
                <div class="info-card">
                    <div class="info-card-header">
                        <h3 data-i18n="site_replication_config">Site Replication Configuration</h3>
                        ${isConfigured ? 
                            '<span class="badge badge-success">✓ Configured</span>' : 
                            '<span class="badge badge-warning">Not Configured</span>'
                        }
                    </div>
                    <div class="info-card-body">
                        ${!isConfigured ? `
                            <div class="replication-setup">
                                <h4 data-i18n="setup_replication">Setup Site Replication</h4>
                                <p data-i18n="setup_replication_desc">Select aliases in order and click "Add Sites" to create site replication cluster.</p>
                                
                                <div class="alias-selection">
                                    <h5 data-i18n="select_aliases">Select Aliases (minimum 2):</h5>
                                    <div id="aliasCheckboxes" class="alias-checkboxes">
                                        ${sites.map((site, index) => `
                                            <label class="alias-checkbox-label">
                                                <input type="checkbox" class="alias-checkbox" value="${site.alias}" data-index="${index}">
                                                <span>${site.alias} (${site.url})</span>
                                            </label>
                                        `).join('')}
                                    </div>
                                    
                                    <div class="selected-order" id="selectedOrder">
                                        <h5 data-i18n="selected_order">Selected Order:</h5>
                                        <div id="selectedAliasesList" class="selected-aliases-list">
                                            <em data-i18n="no_selection">No aliases selected</em>
                                        </div>
                                    </div>
                                    
                                    <button class="btn-primary" id="addSitesBtn" disabled>
                                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                            <line x1="12" y1="5" x2="12" y2="19"></line>
                                            <line x1="5" y1="12" x2="19" y2="12"></line>
                                        </svg>
                                        <span data-i18n="add_sites">Add Sites to Replication</span>
                                    </button>
                                </div>
                            </div>
                        ` : `
                            <div class="replication-management">
                                <h4 data-i18n="manage_replication">Manage Site Replication</h4>
                                <p data-i18n="manage_replication_desc">Manage sites in your replication cluster.</p>
                                
                                <div class="sites-grid">
                                    ${sites.map(site => `
                                        <div class="site-management-card">
                                            <div class="site-management-header">
                                                <div>
                                                    <div class="site-name">${site.alias}</div>
                                                    <div class="site-url">${site.url}</div>
                                                </div>
                                                ${site.replicationStatus === 'configured' ? 
                                                    '<span class="badge badge-success">✓ Active</span>' : 
                                                    '<span class="badge badge-warning">Inactive</span>'
                                                }
                                            </div>
                                            
                                            <div class="site-management-actions">
                                                <button class="btn-icon" onclick="resyncSite('${site.alias}', 'resync-from')" 
                                                        title="Resync FROM this site (pull data)">
                                                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                        <polyline points="23 4 23 10 17 10"></polyline>
                                                        <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 20l4-4 4 4"></path>
                                                    </svg>
                                                    <span data-i18n="resync_from">Resync From</span>
                                                </button>
                                                
                                                <button class="btn-icon" onclick="resyncSite('${site.alias}', 'resync-to')" 
                                                        title="Resync TO this site (push data)">
                                                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                        <polyline points="17 1 21 5 17 9"></polyline>
                                                        <path d="M3 11V9a4 4 0 0 1 4-4h14M1 20l4-4 4 4"></path>
                                                    </svg>
                                                    <span data-i18n="resync_to">Resync To</span>
                                                </button>
                                                
                                                <button class="btn-danger-icon" onclick="removeSite('${site.alias}')" 
                                                        title="Remove from replication cluster">
                                                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                                        <polyline points="3 6 5 6 21 6"></polyline>
                                                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                                                    </svg>
                                                    <span data-i18n="remove">Remove</span>
                                                </button>
                                            </div>
                                        </div>
                                    `).join('')}
                                </div>
                            </div>
                        `}
                    </div>
                </div>
            `;
            
            // Setup event listeners if not configured
            if (!isConfigured) {
                setupAliasSelection();
            }
        })
        .catch(error => {
            console.error('Error loading replication info:', error);
            container.innerHTML = '<div class="error">Error loading replication information</div>';
        });
}

async function renderBucketsPage() {
    const container = document.getElementById('bucketsContent');
    container.innerHTML = '<div class="loading"></div>';
    
    try {
        // Load buckets from all sites
        const allBuckets = [];
        
        for (const site of sites) {
            const response = await fetch(`/api/buckets?alias=${site.alias}`);
            const data = await response.json();
            const buckets = data.buckets || [];
            
            buckets.forEach(bucket => {
                bucket.site = site.alias;
                allBuckets.push(bucket);
            });
        }
        
        // Group buckets by name (to show replication status)
        const bucketGroups = {};
        allBuckets.forEach(bucket => {
            if (!bucketGroups[bucket.name]) {
                bucketGroups[bucket.name] = [];
            }
            bucketGroups[bucket.name].push(bucket);
        });
        
        container.innerHTML = `
            <div class="buckets-table">
                ${Object.keys(bucketGroups).map(bucketName => {
                    const buckets = bucketGroups[bucketName];
                    const replicatedCount = buckets.length;
                    const isFullyReplicated = replicatedCount === sites.length;
                    
                    return `
                        <div class="bucket-row">
                            <div class="bucket-name">${bucketName}</div>
                            <div class="bucket-replication">
                                <span class="badge ${isFullyReplicated ? 'badge-success' : 'badge-warning'}">
                                    ${replicatedCount}/${sites.length} sites
                                </span>
                            </div>
                            <div class="bucket-sites">
                                ${buckets.map(b => `<span class="site-tag">${b.site}</span>`).join('')}
                            </div>
                        </div>
                    `;
                }).join('')}
            </div>
        `;
    } catch (error) {
        console.error('Error rendering buckets page:', error);
        container.innerHTML = '<div class="error">Error loading buckets</div>';
    }
}

async function renderReplicationPage() {
    const container = document.getElementById('replicationContent');
    container.innerHTML = '<div class="loading"></div>';
    
    try {
        const response = await fetch('/api/replication/status');
        const data = await response.json();
        
        let statusHtml = `
            <div class="info-card">
                <div class="info-card-header">
                    <h3>Site Replication Status</h3>
                    <span class="badge badge-${data.status === 'healthy' ? 'success' : 'warning'}">${data.status}</span>
                </div>
                <div class="info-card-body">
                    <div class="replication-sites-grid">
        `;
        
        if (data.sites) {
            for (const [siteName, siteData] of Object.entries(data.sites)) {
                statusHtml += `
                    <div class="site-status-card">
                        <h4>${siteName}</h4>
                        <div class="site-status-stats">
                            <div class="stat-row">
                                <span>Buckets:</span>
                                <strong>${siteData.replicatedBuckets}</strong>
                            </div>
                            <div class="stat-row">
                                <span>Pending:</span>
                                <strong>${siteData.pendingObjects}</strong>
                            </div>
                            <div class="stat-row">
                                <span>Failed:</span>
                                <strong class="${siteData.failedObjects > 0 ? 'text-danger' : ''}">${siteData.failedObjects}</strong>
                            </div>
                            <div class="stat-row">
                                <span>Status:</span>
                                <span class="badge badge-${siteData.healthy ? 'success' : 'danger'}">
                                    ${siteData.healthy ? 'Healthy' : 'Unhealthy'}
                                </span>
                            </div>
                        </div>
                    </div>
                `;
            }
        }
        
        statusHtml += `
                    </div>
                </div>
            </div>
        `;
        
        container.innerHTML = statusHtml;
    } catch (error) {
        console.error('Error loading replication status:', error);
        container.innerHTML = `
            <div class="info-card">
                <div class="info-card-header">
                    <h3>Site Replication Status</h3>
                </div>
                <div class="info-card-body">
                    <div class="empty-state">
                        <h3>Unable to load replication status</h3>
                        <p>${error.message}</p>
                    </div>
                </div>
            </div>
        `;
    }
}

async function renderConsistencyPage() {
    const container = document.getElementById('consistencyContent');
    
    // Add run button handler
    const runBtn = document.getElementById('runConsistencyBtn');
    if (runBtn) {
        runBtn.onclick = async () => {
            await runConsistencyCheck();
        };
    }
    
    container.innerHTML = `
        <div class="info-card">
            <div class="info-card-header">
                <h3>Configuration Consistency</h3>
            </div>
            <div class="info-card-body">
                <p>Check if bucket configurations are consistent across all sites in the replication group.</p>
                <div class="consistency-results" id="consistencyResults">
                    <div class="empty-state">
                        <h3>No checks run yet</h3>
                        <p>Click "Run Check" to validate configuration consistency</p>
                    </div>
                </div>
            </div>
        </div>
    `;
}

async function runConsistencyCheck() {
    const resultsContainer = document.getElementById('consistencyResults');
    resultsContainer.innerHTML = '<div class="loading"></div>';
    
    try {
        const response = await fetch('/api/replication/compare');
        const data = await response.json();
        
        if (!data.buckets || Object.keys(data.buckets).length === 0) {
            resultsContainer.innerHTML = `
                <div class="empty-state">
                    <h3>${data.message || 'No buckets found'}</h3>
                </div>
            `;
            return;
        }
        
        let html = '<div class="consistency-table">';
        
        for (const [bucketName, bucketData] of Object.entries(data.buckets)) {
            const allConsistent = bucketData.policy.consistent && 
                                 bucketData.lifecycle.consistent && 
                                 bucketData.versioning.consistent;
            
            html += `
                <div class="consistency-bucket">
                    <div class="bucket-header">
                        <h4>${bucketName}</h4>
                        <span class="badge badge-${allConsistent ? 'success' : 'warning'}">
                            ${allConsistent ? '✓ Consistent' : '⚠ Inconsistent'}
                        </span>
                    </div>
                    <div class="bucket-details">
                        <div class="detail-row">
                            <span class="detail-label">Exists on:</span>
                            <span>${bucketData.existsOn.join(', ')}</span>
                        </div>
                        <div class="detail-row">
                            <span class="detail-label">Policy:</span>
                            <span class="badge badge-${bucketData.policy.consistent ? 'success' : 'warning'}">
                                ${bucketData.policy.consistent ? 'Consistent' : 'Inconsistent'}
                            </span>
                        </div>
                        <div class="detail-row">
                            <span class="detail-label">Lifecycle:</span>
                            <span class="badge badge-${bucketData.lifecycle.consistent ? 'success' : 'warning'}">
                                ${bucketData.lifecycle.consistent ? 'Consistent' : 'Inconsistent'}
                            </span>
                        </div>
                        <div class="detail-row">
                            <span class="detail-label">Versioning:</span>
                            <span class="badge badge-${bucketData.versioning.consistent ? 'success' : 'warning'}">
                                ${bucketData.versioning.consistent ? 'Consistent' : 'Inconsistent'}
                            </span>
                        </div>
                    </div>
                </div>
            `;
        }
        
        html += '</div>';
        resultsContainer.innerHTML = html;
    } catch (error) {
        console.error('Error running consistency check:', error);
        resultsContainer.innerHTML = `
            <div class="error-message">
                <h3>Check Failed</h3>
                <p>${error.message}</p>
            </div>
        `;
    }
}

async function executeOperation(operation) {
    const modal = document.getElementById('jobModal');
    const jobStatus = document.getElementById('jobStatus');
    
    modal.classList.add('active');
    jobStatus.innerHTML = '<div class="loading">Executing operation...</div>';
    
    try {
        // TODO: Implement actual operations
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        jobStatus.innerHTML = `
            <div class="success-message">
                <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                <h3>Operation Completed Successfully</h3>
                <p>Operation: ${operation}</p>
            </div>
        `;
    } catch (error) {
        jobStatus.innerHTML = `
            <div class="error-message">
                <h3>Operation Failed</h3>
                <p>${error.message}</p>
            </div>
        `;
    }
}

function updateI18n() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.dataset.i18n;
        if (translations[currentLang][key]) {
            el.textContent = translations[currentLang][key];
        }
    });
}

function formatNumber(num) {
    return new Intl.NumberFormat().format(num);
}

// Site Replication Management Functions

let selectedAliases = [];

function setupAliasSelection() {
    const checkboxes = document.querySelectorAll('.alias-checkbox');
    const addBtn = document.getElementById('addSitesBtn');
    const selectedList = document.getElementById('selectedAliasesList');
    
    checkboxes.forEach(checkbox => {
        checkbox.addEventListener('change', (e) => {
            if (e.target.checked) {
                selectedAliases.push(e.target.value);
            } else {
                selectedAliases = selectedAliases.filter(a => a !== e.target.value);
            }
            
            updateSelectedList();
            addBtn.disabled = selectedAliases.length < 2;
        });
    });
    
    if (addBtn) {
        addBtn.onclick = async () => {
            await addSitesToReplication();
        };
    }
}

function updateSelectedList() {
    const selectedList = document.getElementById('selectedAliasesList');
    
    if (selectedAliases.length === 0) {
        selectedList.innerHTML = '<em data-i18n="no_selection">No aliases selected</em>';
    } else {
        selectedList.innerHTML = selectedAliases.map((alias, index) => `
            <div class="selected-alias-item">
                <span class="alias-order">${index + 1}</span>
                <span class="alias-name">${alias}</span>
            </div>
        `).join('');
    }
}

async function addSitesToReplication() {
    const addBtn = document.getElementById('addSitesBtn');
    const originalText = addBtn.innerHTML;
    
    addBtn.disabled = true;
    addBtn.innerHTML = '<span class="loading-spinner"></span> Adding...';
    
    try {
        const response = await fetch('/api/replication/add', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                aliases: selectedAliases
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showNotification('success', data.message || 'Sites added to replication successfully');
            selectedAliases = [];
            // Reload the page after a delay
            setTimeout(() => {
                loadSites();
                renderSitesPage();
            }, 2000);
        } else {
            // Show detailed error in modal instead of simple notification
            showErrorDialog('Site Replication Setup Failed', data.error || 'Failed to add sites to replication');
            addBtn.disabled = false;
            addBtn.innerHTML = originalText;
        }
    } catch (error) {
        console.error('Error adding sites:', error);
        showErrorDialog('Connection Error', 'Failed to connect to the server. Please try again.');
        addBtn.disabled = false;
        addBtn.innerHTML = originalText;
    }
}

async function removeSite(alias) {
    const warningMsg = `⚠️ WARNING: Remove Site Replication Configuration

This will COMPLETELY REMOVE the entire site replication configuration from all sites in the group.

❌ What will happen:
• Site replication will be DISABLED on ALL sites
• All sites will need to be re-added to recreate the replication group
• Existing data will remain, but new changes won't sync

📝 Note: MinIO does not support removing individual sites from a replication group. The "remove" operation removes the entire replication configuration.

Are you absolutely sure you want to proceed?`;

    if (!confirm(warningMsg)) {
        return;
    }
    
    try {
        const response = await fetch('/api/replication/remove', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                alias: alias
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showNotification('success', data.message || 'Site replication configuration removed successfully');
            
            // Show note if provided
            if (data.note) {
                setTimeout(() => {
                    showNotification('info', data.note);
                }, 2000);
            }
            
            // Reload the page after a delay
            setTimeout(() => {
                loadSites();
                renderSitesPage();
                loadInitialData(); // Refresh overview
            }, 3000);
        } else {
            showNotification('error', data.error || 'Failed to remove site replication configuration');
        }
    } catch (error) {
        console.error('Error removing site replication:', error);
        showNotification('error', 'Error removing site replication configuration');
    }
}

async function resyncSite(alias, direction) {
    // Show dialog to select target site
    const otherSites = sites.filter(s => s.alias !== alias);
    
    if (otherSites.length === 0) {
        showNotification('error', 'No other sites available for resync');
        return;
    }
    
    const targetAlias = await showSiteSelectionDialog(otherSites, direction);
    
    if (!targetAlias) {
        return; // User cancelled
    }
    
    const directionText = direction === 'resync-from' ? 'FROM' : 'TO';
    const confirmMsg = direction === 'resync-from' 
        ? `Resync FROM "${alias}" TO "${targetAlias}"?\n\nThis will pull data from ${alias} to ${targetAlias}.`
        : `Resync FROM "${targetAlias}" TO "${alias}"?\n\nThis will push data from ${targetAlias} to ${alias}.`;
    
    if (!confirm(confirmMsg)) {
        return;
    }
    
    try {
        showNotification('info', `Starting resync ${directionText} ${alias}...`);
        
        const response = await fetch('/api/replication/resync', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                source_alias: alias,
                target_alias: targetAlias,
                direction: direction
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showNotification('success', data.message || 'Resync started successfully');
        } else {
            showNotification('error', data.error || 'Failed to start resync');
        }
    } catch (error) {
        console.error('Error starting resync:', error);
        showNotification('error', 'Error starting resync operation');
    }
}

function showSiteSelectionDialog(sites, direction) {
    return new Promise((resolve) => {
        const dialog = document.createElement('div');
        dialog.className = 'modal-overlay';
        dialog.innerHTML = `
            <div class="modal">
                <div class="modal-header">
                    <h3>${direction === 'resync-from' ? 'Select Source Site' : 'Select Target Site'}</h3>
                </div>
                <div class="modal-body">
                    <p>${direction === 'resync-from' 
                        ? 'Select the site to pull data FROM:' 
                        : 'Select the site to push data TO:'}</p>
                    <div class="site-selection-list">
                        ${sites.map(site => `
                            <button class="site-selection-item" data-alias="${site.alias}">
                                <div>
                                    <div class="site-name">${site.alias}</div>
                                    <div class="site-url">${site.url}</div>
                                </div>
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <polyline points="9 18 15 12 9 6"></polyline>
                                </svg>
                            </button>
                        `).join('')}
                    </div>
                </div>
                <div class="modal-footer">
                    <button class="btn-secondary" id="cancelSiteSelection">Cancel</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(dialog);
        
        // Add event listeners
        dialog.querySelectorAll('.site-selection-item').forEach(btn => {
            btn.onclick = () => {
                resolve(btn.dataset.alias);
                document.body.removeChild(dialog);
            };
        });
        
        dialog.querySelector('#cancelSiteSelection').onclick = () => {
            resolve(null);
            document.body.removeChild(dialog);
        };
        
        // Close on overlay click
        dialog.onclick = (e) => {
            if (e.target === dialog) {
                resolve(null);
                document.body.removeChild(dialog);
            }
        };
    });
}

function showNotification(type, message) {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <div class="notification-content">
            ${type === 'success' ? '✓' : type === 'error' ? '✗' : 'ℹ'} ${message}
        </div>
    `;
    
    document.body.appendChild(notification);
    
    // Trigger animation
    setTimeout(() => {
        notification.classList.add('show');
    }, 10);
    
    // Remove after 5 seconds
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 5000);
}

function showErrorDialog(title, message) {
    const dialog = document.createElement('div');
    dialog.className = 'modal-overlay error-dialog';
    dialog.innerHTML = `
        <div class="modal error-modal">
            <div class="modal-header error-header">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                </svg>
                <h3>${title}</h3>
            </div>
            <div class="modal-body">
                <pre class="error-message">${message}</pre>
            </div>
            <div class="modal-footer">
                <button class="btn-primary" id="closeErrorDialog">OK</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(dialog);
    
    dialog.querySelector('#closeErrorDialog').onclick = () => {
        document.body.removeChild(dialog);
    };
    
    // Close on overlay click
    dialog.onclick = (e) => {
        if (e.target === dialog) {
            document.body.removeChild(dialog);
        }
    };
    
    // Close on Escape key
    const handleEscape = (e) => {
        if (e.key === 'Escape') {
            document.body.removeChild(dialog);
            document.removeEventListener('keydown', handleEscape);
        }
    };
    document.addEventListener('keydown', handleEscape);
}

// Export for global access
window.viewSiteDetails = viewSiteDetails;
window.removeSite = removeSite;
window.resyncSite = resyncSite;

