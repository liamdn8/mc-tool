// Consistency page functionality
import { runConsistencyCheck } from '../utils/api.js';

export async function renderConsistencyPage() {
    const container = document.getElementById('consistencyContent');
    
    // Add run button handler
    const runBtn = document.getElementById('runConsistencyBtn');
    if (runBtn) {
        runBtn.onclick = async () => {
            await runConsistencyCheckHandler();
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

async function runConsistencyCheckHandler() {
    const resultsContainer = document.getElementById('consistencyResults');
    resultsContainer.innerHTML = '<div class="loading"></div>';
    
    try {
        const { response, data } = await runConsistencyCheck();
        
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