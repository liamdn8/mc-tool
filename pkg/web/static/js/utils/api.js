// API utility functions

export async function apiCall(endpoint, options = {}) {
    try {
        const response = await fetch(endpoint, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            }
        });
        
        const data = await response.json();
        return { response, data };
    } catch (error) {
        console.error(`API call to ${endpoint} failed:`, error);
        throw error;
    }
}

export async function loadAliases() {
    try {
        const { response, data } = await apiCall('/api/aliases');
        return data.aliases || [];
    } catch (error) {
        console.error('Error loading aliases:', error);
        return [];
    }
}

export async function loadSiteReplicationInfo() {
    try {
        const { response, data } = await apiCall('/api/replication/info');
        
        const replicationInfo = {
            enabled: data.enabled || false,
            sites: data.totalAliases || 0,
            syncedBuckets: 0,
            totalObjects: 0,
            health: 'healthy',
            replicationGroup: data.replicationGroup || null
        };
        
        let sites = [];
        if (data.aliases && data.aliases.length > 0) {
            sites = data.aliases.map(aliasData => ({
                alias: aliasData.alias,
                url: aliasData.url || aliasData.endpoint,
                healthy: aliasData.healthy === true,
                replicationEnabled: aliasData.replicationEnabled === true,
                replicationStatus: aliasData.replicationStatus || 'not_configured',
                siteName: aliasData.siteName || '',
                deploymentID: aliasData.deploymentID || '',
                serverCount: aliasData.serverCount || 0
            }));
        }
        
        return { replicationInfo, sites };
    } catch (error) {
        console.error('Error loading site replication info:', error);
        return { 
            replicationInfo: {
                enabled: false,
                sites: 0,
                syncedBuckets: 0,
                totalObjects: 0,
                health: 'unknown',
                replicationGroup: null
            }, 
            sites: [] 
        };
    }
}

export async function loadSiteBucketCount(alias) {
    try {
        const { response, data } = await apiCall(`/api/buckets?alias=${alias}`);
        const buckets = data.buckets || [];
        
        // Load admin info to get total objects and size
        try {
            const { response: infoResponse, data: infoData } = await apiCall(`/api/alias-health?alias=${alias}`);
            return {
                bucketCount: buckets.length,
                totalObjects: infoData.objectCount || 0,
                totalSize: infoData.totalSize || 0
            };
        } catch (error) {
            // Fallback: calculate from buckets
            let totalObjects = 0;
            for (const bucket of buckets) {
                try {
                    const { response: statsResponse, data: stats } = await apiCall(`/api/bucket-stats?alias=${alias}&bucket=${bucket.name}`);
                    totalObjects += stats.objectCount || 0;
                } catch (err) {
                    console.error(`Error loading stats for ${bucket.name}:`, err);
                }
            }
            
            return {
                bucketCount: buckets.length,
                totalObjects: totalObjects,
                totalSize: 0
            };
        }
    } catch (error) {
        console.error(`Error loading bucket count for ${alias}:`, error);
        return {
            bucketCount: 0,
            totalObjects: 0,
            totalSize: 0
        };
    }
}

export async function addSitesToReplication(aliases) {
    const { response, data } = await apiCall('/api/replication/add', {
        method: 'POST',
        body: JSON.stringify({ aliases })
    });
    
    return { response, data };
}

export async function removeSiteFromReplication(alias) {
    const { response, data } = await apiCall('/api/replication/remove', {
        method: 'POST',
        body: JSON.stringify({ alias })
    });
    
    return { response, data };
}

export async function loadBucketsForAllSites(sites) {
    const allBuckets = [];
    
    for (const site of sites) {
        try {
            const { response, data } = await apiCall(`/api/buckets?alias=${site.alias}`);
            const buckets = data.buckets || [];
            
            buckets.forEach(bucket => {
                bucket.site = site.alias;
                allBuckets.push(bucket);
            });
        } catch (error) {
            console.error(`Error loading buckets for ${site.alias}:`, error);
        }
    }
    
    return allBuckets;
}

export async function loadReplicationStatus() {
    const { response, data } = await apiCall('/api/replication/status');
    return { response, data };
}

export async function runConsistencyCheck() {
    const { response, data } = await apiCall('/api/replication/compare');
    return { response, data };
}