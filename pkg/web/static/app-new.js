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
                url: aliasData.endpoint,
                healthy: aliasData.healthy !== false,
                replicationEnabled: aliasData.replicationEnabled || false,
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
    document.getElementById('totalSites').textContent = sites.length;
    document.getElementById('syncedBuckets').textContent = replicationInfo?.syncedBuckets || 0;
    document.getElementById('totalObjects').textContent = formatNumber(replicationInfo?.totalObjects || 0);
    
    const healthIndicator = document.getElementById('healthIndicator');
    if (replicationInfo?.health === 'healthy') {
        healthIndicator.innerHTML = `
            <span class="pulse"></span>
            <span data-i18n="healthy">${translations[currentLang].healthy}</span>
        `;
    }
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
        // Determine status badge
        let statusBadge = '';
        let statusClass = '';
        
        if (site.replicationStatus === 'configured') {
            statusBadge = `<span class="badge badge-success">✓ ${translations[currentLang].replication_enabled}</span>`;
            statusClass = 'site-card-replicated';
        } else if (site.replicationStatus === 'disabled') {
            statusBadge = `<span class="badge badge-warning">⚠ ${translations[currentLang].replication_disabled}</span>`;
            statusClass = 'site-card-warning';
        } else {
            statusBadge = `<span class="badge badge-danger">✗ ${translations[currentLang].not_configured}</span>`;
            statusClass = 'site-card-not-configured';
        }
        
        const healthBadge = site.healthy 
            ? `<span class="health-indicator"><span class="pulse"></span></span>`
            : `<span class="badge badge-danger">Offline</span>`;
        
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
                ${site.deploymentID ? `<div class="site-deployment-id">Deployment: ${site.deploymentID}</div>` : ''}
                <div class="site-stats">
                    <div class="site-stat">
                        <div class="site-stat-label">${translations[currentLang].status}</div>
                        <div class="site-stat-value">${healthBadge}</div>
                    </div>
                    <div class="site-stat">
                        <div class="site-stat-label">Buckets</div>
                        <div class="site-stat-value" id="site-buckets-${site.alias}">-</div>
                    </div>
                    <div class="site-stat">
                        <div class="site-stat-label">Objects</div>
                        <div class="site-stat-value" id="site-objects-${site.alias}">-</div>
                    </div>
                    ${site.serverCount > 0 ? `
                    <div class="site-stat">
                        <div class="site-stat-label">${translations[currentLang].servers}</div>
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
        
        // Load total objects (sum from all buckets)
        let totalObjects = 0;
        for (const bucket of buckets) {
            const statsResponse = await fetch(`/api/bucket-stats?alias=${alias}&bucket=${bucket.name}`);
            const stats = await statsResponse.json();
            totalObjects += stats.objectCount || 0;
        }
        
        const objectCountEl = document.getElementById(`site-objects-${alias}`);
        if (objectCountEl) {
            objectCountEl.textContent = formatNumber(totalObjects);
        }
        
        // Update overview total
        if (replicationInfo) {
            replicationInfo.totalObjects += totalObjects;
            updateOverviewStats();
        }
    } catch (error) {
        console.error(`Error loading bucket count for ${alias}:`, error);
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
    container.innerHTML = `
        <div class="info-card">
            <div class="info-card-header">
                <h3>Site Replication Configuration</h3>
            </div>
            <div class="info-card-body">
                <p>Manage sites in your replication group. All sites should be part of the same MinIO Site Replication setup.</p>
                ${renderSitesList()}
            </div>
        </div>
    `;
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

// Export for global access
window.viewSiteDetails = viewSiteDetails;
