// Sites page functionality - Site Replication Management
import { loadSiteReplicationInfo, addSitesToReplication, removeSiteFromReplication, removeIndividualSiteFromReplication } from '../utils/api.js';
import { showNotification, showErrorDialog, showSiteSelectionDialog, handleReplicationSuccess, handleReplicationError } from '../utils/helpers.js';

let selectedAliases = [];

export async function renderSitesPage(sites) {
    const container = document.getElementById('sitesContent');
    
    try {
        // Get replication info to check if already configured
        const { replicationInfo } = await loadSiteReplicationInfo();
        const isConfigured = replicationInfo.enabled === true;
        
        // Calculate summary statistics
        const totalSites = sites.length;
        const configuredSites = sites.filter(site => site.replicationStatus === 'configured').length;
        const healthySites = sites.filter(site => site.healthy).length;
        const availableSites = totalSites - configuredSites;
        
        container.innerHTML = `
            ${isConfigured ? `
                <!-- Summary Statistics Section -->
                <div class="stats-summary">
                    <div class="stat-card">
                        <div class="stat-value">${configuredSites}</div>
                        <div class="stat-label">Sites in Cluster</div>
                        <div class="stat-summary">Active replication sites</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${healthySites}</div>
                        <div class="stat-label">Healthy Sites</div>
                        <div class="stat-summary">Sites responding normally</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${availableSites}</div>
                        <div class="stat-label">Available to Add</div>
                        <div class="stat-summary">Sites ready for replication</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${replicationInfo.sites ? replicationInfo.sites.length : 0}</div>
                        <div class="stat-label">Total Endpoints</div>
                        <div class="stat-summary">Configured replication endpoints</div>
                    </div>
                </div>
            ` : ''}
            
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
                                    <i data-lucide="plus" width="16" height="16"></i>
                                    <span data-i18n="add_sites">Add Sites to Replication</span>
                                </button>
                            </div>
                        </div>
                    ` : `
                        <div class="replication-management">
                            <h4 data-i18n="manage_replication">Manage Site Replication</h4>
                            <p data-i18n="manage_replication_desc">Manage sites in your replication cluster.</p>
                            
                            <!-- Add Sites to Existing Cluster -->
                            <div class="add-sites-section">
                                <h5 data-i18n="add_sites_to_cluster">Add Sites to Existing Cluster</h5>
                                
                                <div class="table-container">
                                    <table class="table">
                                        <thead>
                                            <tr>
                                                <th style="width: 40px;">
                                                    <input type="checkbox" id="selectAllAvailable" title="Select all available sites">
                                                </th>
                                                <th>Site Name</th>
                                                <th>Endpoint</th>
                                                <th>Health</th>
                                                <th>Status</th>
                                                <th>Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            ${sites.filter(site => site.replicationStatus !== 'configured').map(site => `
                                                <tr>
                                                    <td>
                                                        <input type="checkbox" class="add-site-checkbox" value="${site.alias}">
                                                    </td>
                                                    <td>
                                                        <div class="site-name">${site.alias}</div>
                                                    </td>
                                                    <td>
                                                        <div class="site-url">${site.url}</div>
                                                    </td>
                                                    <td>
                                                        <span class="badge badge-${site.healthy ? 'success' : 'danger'}">
                                                            ${site.healthy ? '● Healthy' : '● Unhealthy'}
                                                        </span>
                                                    </td>
                                                    <td>
                                                        <span class="badge badge-warning">Available</span>
                                                    </td>
                                                    <td>
                                                        <button class="btn-primary btn-sm" onclick="addSingleSite('${site.alias}')" 
                                                                title="Add this site to replication cluster">
                                                            <i data-lucide="plus" width="14" height="14"></i>
                                                            Add
                                                        </button>
                                                    </td>
                                                </tr>
                                            `).join('')}
                                        </tbody>
                                    </table>
                                </div>
                                
                                <div class="add-sites-actions">
                                    <button class="btn-primary" id="addToClusterBtn" disabled>
                                        <i data-lucide="plus" width="16" height="16"></i>
                                        <span data-i18n="add_selected_to_cluster">Add Selected to Cluster</span>
                                    </button>
                                </div>
                            </div>

                            <!-- Current Cluster Sites -->
                            <div class="cluster-sites-section">
                                <div class="cluster-header">
                                    <h5 data-i18n="current_cluster">Current Cluster Sites</h5>
                                    <button class="btn-danger" id="removeSelectedBtn" disabled>
                                        <i data-lucide="trash-2" width="16" height="16"></i>
                                        <span data-i18n="remove_selected">Remove Selected</span>
                                    </button>
                                </div>
                            
                                <div class="table-container">
                                    <table class="table">
                                        <thead>
                                            <tr>
                                                <th style="width: 40px;">
                                                    <input type="checkbox" id="selectAllSites" title="Select all sites">
                                                </th>
                                                <th>Site Name</th>
                                                <th>Endpoint</th>
                                                <th>Status</th>
                                                <th>Health</th>
                                                <th>Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            ${sites.filter(site => site.replicationStatus === 'configured').map(site => `
                                                <tr>
                                                    <td>
                                                        <input type="checkbox" class="remove-site-checkbox" value="${site.alias}">
                                                    </td>
                                                    <td>
                                                        <div class="site-info">
                                                            <div class="site-name">${site.alias}</div>
                                                        </div>
                                                    </td>
                                                    <td>
                                                        <div class="site-url">${site.url}</div>
                                                    </td>
                                                    <td>
                                                        <span class="badge badge-success">✓ Active</span>
                                                    </td>
                                                    <td>
                                                        <span class="badge badge-${site.healthy ? 'success' : 'danger'}">
                                                            ${site.healthy ? '● Healthy' : '● Unhealthy'}
                                                        </span>
                                                    </td>
                                                    <td>
                                                        <div class="action-buttons">
                                                            <button class="btn-icon" onclick="resyncSite('${site.alias}', 'resync-from')" 
                                                                    title="Resync FROM this site (pull data)">
                                                                <i data-lucide="download" width="16" height="16"></i>
                                                            </button>
                                                            
                                                            <button class="btn-icon" onclick="resyncSite('${site.alias}', 'resync-to')" 
                                                                    title="Resync TO this site (push data)">
                                                                <i data-lucide="upload" width="16" height="16"></i>
                                                            </button>
                                                            
                                                            <button class="btn-danger-icon" onclick="removeIndividualSite('${site.alias}')" 
                                                                    title="Remove this site from replication cluster">
                                                                <i data-lucide="trash-2" width="16" height="16"></i>
                                                            </button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            `).join('')}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    `}
                </div>
            </div>
        `;
        
        // Setup event listeners
        if (!isConfigured) {
            setupAliasSelection();
        } else {
            setupReplicationManagement();
        }
        
        // Re-initialize lucide icons for the new content
        if (typeof lucide !== 'undefined') {
            lucide.createIcons();
        }
    } catch (error) {
        console.error('Error loading replication info:', error);
        container.innerHTML = '<div class="error">Error loading replication information</div>';
    }
}

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
            await addSitesToReplicationHandler();
        };
    }
}

function setupReplicationManagement() {
    // Setup Add to Cluster functionality
    const addSiteCheckboxes = document.querySelectorAll('.add-site-checkbox');
    const addToClusterBtn = document.getElementById('addToClusterBtn');
    const removeSiteCheckboxes = document.querySelectorAll('.remove-site-checkbox');
    const removeSelectedBtn = document.getElementById('removeSelectedBtn');
    
    let selectedToAdd = [];
    let selectedToRemove = [];
    
    // Add sites to cluster
    addSiteCheckboxes.forEach(checkbox => {
        checkbox.addEventListener('change', (e) => {
            if (e.target.checked) {
                selectedToAdd.push(e.target.value);
            } else {
                selectedToAdd = selectedToAdd.filter(a => a !== e.target.value);
            }
            
            if (addToClusterBtn) {
                addToClusterBtn.disabled = selectedToAdd.length === 0;
            }
        });
    });
    
    if (addToClusterBtn) {
        addToClusterBtn.onclick = async () => {
            await addSitesToExistingCluster(selectedToAdd);
            selectedToAdd = [];
        };
    }
    
    // Remove sites from cluster
    removeSiteCheckboxes.forEach(checkbox => {
        checkbox.addEventListener('change', (e) => {
            if (e.target.checked) {
                selectedToRemove.push(e.target.value);
            } else {
                selectedToRemove = selectedToRemove.filter(a => a !== e.target.value);
            }
            
            if (removeSelectedBtn) {
                removeSelectedBtn.disabled = selectedToRemove.length === 0;
            }
        });
    });
    
    if (removeSelectedBtn) {
        removeSelectedBtn.onclick = async () => {
            await removeMultipleSites(selectedToRemove);
            selectedToRemove = [];
        };
    }
    
    // Setup select all checkboxes
    const selectAllSites = document.getElementById('selectAllSites');
    const selectAllAvailable = document.getElementById('selectAllAvailable');
    
    if (selectAllSites) {
        selectAllSites.addEventListener('change', (e) => {
            const checkboxes = document.querySelectorAll('.remove-site-checkbox');
            checkboxes.forEach(checkbox => {
                checkbox.checked = e.target.checked;
                const event = new Event('change', { bubbles: true });
                checkbox.dispatchEvent(event);
            });
        });
    }
    
    if (selectAllAvailable) {
        selectAllAvailable.addEventListener('change', (e) => {
            const checkboxes = document.querySelectorAll('.add-site-checkbox');
            checkboxes.forEach(checkbox => {
                checkbox.checked = e.target.checked;
                const event = new Event('change', { bubbles: true });
                checkbox.dispatchEvent(event);
            });
        });
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

async function addSitesToReplicationHandler() {
    const addBtn = document.getElementById('addSitesBtn');
    const originalText = addBtn.innerHTML;
    
    addBtn.disabled = true;
    addBtn.innerHTML = '<span class="loading-spinner"></span> Adding...';
    
    try {
        const { response, data } = await addSitesToReplication(selectedAliases);
        
        // Check both response status and data.success/error
        if (response.ok && data.success !== false && !data.error) {
            handleReplicationSuccess(data.message || 'Sites added to replication successfully');
            selectedAliases = [];
        } else {
            handleReplicationError('Site Replication Setup Failed', data, response);
        }
    } catch (error) {
        handleReplicationError('Site Replication Setup Failed', error, null);
    } finally {
        addBtn.disabled = false;
        addBtn.innerHTML = originalText;
    }
}

async function addSitesToExistingCluster(newSites) {
    const addBtn = document.getElementById('addToClusterBtn');
    const originalText = addBtn.innerHTML;
    
    addBtn.disabled = true;
    addBtn.innerHTML = '<span class="loading-spinner"></span> Adding...';
    
    try {
        // Get current site data
        const { sites } = await loadSiteReplicationInfo();
        
        // Get current configured sites
        const currentConfiguredSites = sites
            .filter(site => site.replicationStatus === 'configured')
            .map(site => site.alias);
        
        // Combine current sites with new sites
        const allSites = [...currentConfiguredSites, ...newSites];
        
        const { response, data } = await addSitesToReplication(allSites);
        
        if (response.ok && data.success !== false && !data.error) {
            handleReplicationSuccess(`Successfully added ${newSites.join(', ')} to replication cluster`);
        } else {
            handleReplicationError('Add Sites Failed', data, response);
        }
    } catch (error) {
        handleReplicationError('Add Sites Failed', error, null);
    } finally {
        addBtn.disabled = false;
        addBtn.innerHTML = originalText;
    }
}

async function removeMultipleSites(sitesToRemove) {
    const removeBtn = document.getElementById('removeSelectedBtn');
    const originalText = removeBtn.innerHTML;
    
    const warningMsg = `⚠️ WARNING: Remove Multiple Sites from Replication

This will remove the following sites from replication cluster:
${sitesToRemove.map(site => `• ${site}`).join('\n')}

❌ What will happen:
• Selected sites will be removed from replication
• If only 2 sites remain, entire replication config will be disabled
• Existing data will remain, but sync will stop

Are you sure you want to proceed?`;

    if (!confirm(warningMsg)) {
        return;
    }
    
    removeBtn.disabled = true;
    removeBtn.innerHTML = '<span class="loading-spinner"></span> Removing...';
    
    try {
        // Remove sites one by one using the new individual site removal API
        for (const siteAlias of sitesToRemove) {
            const { response, data } = await removeIndividualSiteFromReplication(siteAlias);
            
            if (!response.ok || data.error) {
                throw new Error(`Failed to remove ${siteAlias}: ${data.error || 'Unknown error'}`);
            }
        }
        
        handleReplicationSuccess(`Successfully removed ${sitesToRemove.length} sites from replication`);
        
    } catch (error) {
        handleReplicationError('Remove Sites Failed', error, null);
    } finally {
        removeBtn.disabled = false;
        removeBtn.innerHTML = originalText;
    }
}

// Export for global access
window.removeIndividualSite = async function(alias) {
    const warningMsg = `⚠️ WARNING: Remove Individual Site from Replication

This will remove ONLY the site "${alias}" from the replication cluster.

✅ What will happen:
• Site "${alias}" will be removed from replication
• Other sites in the cluster will continue replicating
• Existing data on "${alias}" will remain

Are you sure you want to remove "${alias}" from replication?`;

    if (!confirm(warningMsg)) {
        return;
    }
    
    try {
        const { response, data } = await removeIndividualSiteFromReplication(alias);
        
        if (response.ok) {
            let message = data.message || `Site "${alias}" removed from replication successfully`;
            handleReplicationSuccess(message);
        } else {
            handleReplicationError('Remove Site Failed', data, response);
        }
    } catch (error) {
        handleReplicationError('Remove Site Failed', error, null);
    }
};

window.removeSite = async function(alias) {
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
        const { response, data } = await removeSiteFromReplication(alias);
        
        if (response.ok) {
            let message = data.message || 'Site replication configuration removed successfully';
            if (data.note) {
                message += `\n\n${data.note}`;
            }
            handleReplicationSuccess(message);
        } else {
            handleReplicationError('Remove Site Failed', data, response);
        }
    } catch (error) {
        handleReplicationError('Remove Site Failed', error, null);
    }
};

window.resyncSite = async function(alias, direction) {
    // Get current site data
    const { sites } = await loadSiteReplicationInfo();
    
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
};

// Add single site to replication cluster
window.addSingleSite = async function(alias) {
    const warningMsg = `Add site "${alias}" to replication cluster?

This will add the site "${alias}" to the existing replication cluster.

Are you sure you want to proceed?`;

    if (!confirm(warningMsg)) {
        return;
    }
    
    try {
        await addSitesToExistingCluster([alias]);
    } catch (error) {
        console.error('Error adding single site:', error);
        showNotification('error', 'Error adding site to cluster');
    }
};