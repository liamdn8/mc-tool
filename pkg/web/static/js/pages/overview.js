// Overview page functionality
import { loadAliases, loadSiteReplicationInfo, loadSiteBucketCount } from '../utils/api.js';
import { formatNumber, formatBytes } from '../utils/helpers.js';

export async function renderOverviewPage() {
    // Load data is handled by main app initialization
    // This function updates the overview stats display
}

export function updateOverviewStats(sites, replicationInfo) {
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

export function renderSitesList(sites) {
    const container = document.getElementById('sitesList');
    
    if (sites.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <i data-lucide="globe" width="48" height="48"></i>
                <h3>No MinIO Aliases Configured</h3>
                <p>Add MinIO aliases using mc command to get started</p>
            </div>
        `;
        // Re-initialize lucide icons for the new content
        if (typeof lucide !== 'undefined') {
            lucide.createIcons();
        }
        return;
    }
    
    container.innerHTML = sites.map(site => {
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
        
        // Health badge
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
        loadSiteBucketCountForSite(site.alias, sites);
    });
}

async function loadSiteBucketCountForSite(alias, sites) {
    try {
        const siteData = await loadSiteBucketCount(alias);
        
        const bucketCountEl = document.getElementById(`site-buckets-${alias}`);
        if (bucketCountEl) {
            bucketCountEl.textContent = siteData.bucketCount;
        }
        
        // Update site object with data
        const siteIndex = sites.findIndex(s => s.alias === alias);
        if (siteIndex !== -1) {
            sites[siteIndex].totalObjects = siteData.totalObjects;
            sites[siteIndex].totalSize = siteData.totalSize;
        }
        
        const objectCountEl = document.getElementById(`site-objects-${alias}`);
        if (objectCountEl) {
            objectCountEl.textContent = formatNumber(siteData.totalObjects);
        }
    } catch (error) {
        console.error(`Error loading bucket count for ${alias}:`, error);
        const bucketCountEl = document.getElementById(`site-buckets-${alias}`);
        if (bucketCountEl) bucketCountEl.textContent = '-';
        
        const objectCountEl = document.getElementById(`site-objects-${alias}`);
        if (objectCountEl) objectCountEl.textContent = '-';
    }
}

// Export for global access
window.viewSiteDetails = function(alias) {
    // Navigate to buckets page and filter by site
    window.app.navigateToPage('buckets');
    // TODO: Filter buckets by selected site
};