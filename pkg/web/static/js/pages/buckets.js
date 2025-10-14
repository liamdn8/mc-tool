// Buckets page functionality
import { loadBucketsForAllSites } from '../utils/api.js';

export async function renderBucketsPage(sites) {
    const container = document.getElementById('bucketsContent');
    container.innerHTML = '<div class="loading"></div>';
    
    try {
        // Load buckets from all sites
        const allBuckets = await loadBucketsForAllSites(sites);
        
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