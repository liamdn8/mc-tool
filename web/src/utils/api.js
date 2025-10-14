// API utility functions for React app

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
        
        try {
            const { response: infoResponse, data: infoData } = await apiCall(`/api/alias-health?alias=${alias}`);
            return {
                bucketCount: buckets.length,
                totalObjects: infoData.objectCount || 0,
                totalSize: infoData.totalSize || 0
            };
        } catch (error) {
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
                totalObjects,
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

export async function loadBuckets(alias) {
    try {
        const { response, data } = await apiCall(`/api/buckets?alias=${alias}`);
        return data.buckets || [];
    } catch (error) {
        console.error(`Error loading buckets for ${alias}:`, error);
        return [];
    }
}

export async function addSiteReplication(aliases) {
    try {
        const { response, data } = await apiCall('/api/replication/add', {
            method: 'POST',
            body: JSON.stringify({ aliases })
        });
        return data;
    } catch (error) {
        console.error('Error adding site replication:', error);
        throw error;
    }
}

export async function removeSiteReplication() {
    try {
        // Get first alias to use for removal command
        const aliases = await loadAliases();
        if (aliases.length === 0) {
            throw new Error('No aliases available');
        }
        
        const { response, data } = await apiCall('/api/replication/remove', {
            method: 'POST',
            body: JSON.stringify({ alias: aliases[0].alias })
        });
        return data;
    } catch (error) {
        console.error('Error removing site replication:', error);
        throw error;
    }
}

export async function loadReplicationStatus() {
    try {
        const { response, data } = await apiCall('/api/replication/status');
        return data;
    } catch (error) {
        console.error('Error loading replication status:', error);
        return null;
    }
}

export async function loadConsistencyCheck() {
    try {
        const { response, data } = await apiCall('/api/consistency/check');
        return data;
    } catch (error) {
        console.error('Error loading consistency check:', error);
        return null;
    }
}

export async function performConsistencyCheck(buckets = []) {
    try {
        const { response, data } = await apiCall('/api/consistency/run', {
            method: 'POST',
            body: JSON.stringify({ buckets })
        });
        return data;
    } catch (error) {
        console.error('Error performing consistency check:', error);
        throw error;
    }
}

export async function resyncSiteReplication(fromSite, toSite) {
    try {
        const { response, data } = await apiCall('/api/replication/resync', {
            method: 'POST',
            body: JSON.stringify({ fromSite, toSite })
        });
        return data;
    } catch (error) {
        console.error('Error resyncing site replication:', error);
        throw error;
    }
}

export async function addSitesToReplication(aliases) {
    try {
        const { response, data } = await apiCall('/api/replication/add', {
            method: 'POST',
            body: JSON.stringify({ aliases })
        });
        return data;
    } catch (error) {
        console.error('Error adding sites to replication:', error);
        throw error;
    }
}

export async function removeSiteFromReplication(alias) {
    try {
        const { response, data } = await apiCall('/api/replication/remove-site', {
            method: 'POST',
            body: JSON.stringify({ alias })
        });
        return data;
    } catch (error) {
        console.error('Error removing site from replication:', error);
        throw error;
    }
}

export async function removeBulkSitesFromReplication(aliases) {
    try {
        const { response, data } = await apiCall('/api/replication/remove-site', {
            method: 'POST',
            body: JSON.stringify({ aliases })
        });
        return data;
    } catch (error) {
        console.error('Error removing sites from replication:', error);
        throw error;
    }
}