// Replication page functionality
import { loadReplicationStatus } from '../utils/api.js';

export async function renderReplicationPage() {
    const container = document.getElementById('replicationContent');
    container.innerHTML = '<div class="loading"></div>';
    
    try {
        const { response, data } = await loadReplicationStatus();
        
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