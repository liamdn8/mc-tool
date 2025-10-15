import React, { useState, useEffect } from 'react';
import { Plus, Trash2, RefreshCw, Download, Upload, X } from 'lucide-react';
import { useI18n } from '../utils/i18n';
import { getBadgeClass, getStatusText } from '../utils/helpers';
import SplitBrainWarning from '../components/SplitBrainWarning';
import { 
    loadAliases, 
    loadSiteReplicationInfo, 
    addSitesToReplication,
    addSitesToReplicationSmart, 
    loadReplicationStatus, 
    resyncReplication,
    removeSiteFromReplication,
    removeBulkSitesFromReplication,
    removeIndividualSiteFromReplication,
    removeSiteFromReplicationSmart,
    removeBulkSitesFromReplicationSmart,
    removeIndividualSiteFromReplicationSmart,
    checkSplitBrainStatus
} from '../utils/api';

const SitesPage = ({ sites, replicationInfo, onRefresh }) => {
    const { t } = useI18n();
    const [selectedAliases, setSelectedAliases] = useState([]);
    const [selectedSitesToAdd, setSelectedSitesToAdd] = useState([]);
    const [selectedSitesToRemove, setSelectedSitesToRemove] = useState([]);
    const [isAddingReplication, setIsAddingReplication] = useState(false);
    const [isAddingToCluster, setIsAddingToCluster] = useState(false);
    const [showResyncModal, setShowResyncModal] = useState(false);
    const [resyncFromSite, setResyncFromSite] = useState('');
    const [resyncToSite, setResyncToSite] = useState('');

    const hasReplication = replicationInfo && replicationInfo.enabled;

    const handleAliasToggle = (alias) => {
        setSelectedAliases(prev => {
            if (prev.includes(alias)) {
                return prev.filter(a => a !== alias);
            } else {
                return [...prev, alias];
            }
        });
    };

    const handleAddReplication = async () => {
        if (selectedAliases.length < 2) {
            alert('Please select at least 2 aliases');
            return;
        }

        setIsAddingReplication(true);
        try {
            await addSitesToReplication(selectedAliases);
            setSelectedAliases([]);
            onRefresh();
        } catch (error) {
            alert(`Error setting up replication: ${error.message}`);
        } finally {
            setIsAddingReplication(false);
        }
    };

    const handleResyncReplication = async () => {
        if (!resyncFromSite || !resyncToSite) {
            alert('Please select both source and target sites');
            return;
        }

        try {
            await resyncSiteReplication(resyncFromSite, resyncToSite);
            setShowResyncModal(false);
            setResyncFromSite('');
            setResyncToSite('');
            onRefresh();
            alert('Resync operation started successfully');
        } catch (error) {
            alert(`Error starting resync: ${error.message}`);
        }
    };

    const handleAddToCluster = async () => {
        if (selectedSitesToAdd.length === 0) {
            alert('Please select at least one site to add');
            return;
        }

        // Check for split brain first
        try {
            const splitBrainStatus = await checkSplitBrainStatus();
            if (splitBrainStatus.splitBrainDetected) {
                let errorMsg = '⚠️ SPLIT BRAIN DETECTED - Cannot add sites!\n\n';
                errorMsg += `${splitBrainStatus.clusterCount} separate clusters found.\n`;
                errorMsg += 'Please resolve the split brain scenario first.\n\n';
                errorMsg += 'Check the warning above for detailed instructions.';
                alert(errorMsg);
                return;
            }
        } catch (error) {
            console.error('Error checking split brain status:', error);
            // Continue with operation if check fails
        }

        const siteCount = selectedSitesToAdd.length;
        const confirmMessage = siteCount === 1 
            ? `Add site "${selectedSitesToAdd[0]}" to replication cluster using smart detection?`
            : `Add ${siteCount} sites (${selectedSitesToAdd.join(', ')}) to replication cluster using smart detection?`;

        if (!window.confirm(confirmMessage)) {
            return;
        }

        console.log('Adding sites to cluster (smart):', selectedSitesToAdd);
        setIsAddingToCluster(true);
        try {
            const result = await addSitesToReplicationSmart(selectedSitesToAdd);
            setSelectedSitesToAdd([]);
            
            // Show detailed result to user
            let message = `Smart cluster operation for ${siteCount} site${siteCount > 1 ? 's' : ''} completed:\n\n`;
            if (result.data) {
                message += `Operation: ${result.data.action || result.data.operation}\n`;
                if (result.data.clustersFound !== undefined) {
                    message += `Clusters detected: ${result.data.clustersFound}\n`;
                }
                
                // Handle new sites added
                if (result.data.newAliases && result.data.newAliases.length > 0) {
                    message += `✅ Sites added: ${result.data.newAliases.join(', ')}\n`;
                }
                
                // Handle sites already in cluster
                if (result.data.alreadyInCluster && result.data.alreadyInCluster.length > 0) {
                    message += `⚠️ Already in cluster: ${result.data.alreadyInCluster.join(', ')}\n`;
                }
                
                // Show cluster info
                if (result.data.existingCluster && result.data.existingCluster.sites) {
                    message += `Total sites in cluster: ${result.data.existingCluster.sites.length}\n`;
                }
                
                // Handle result message
                if (result.data.message) {
                    message += `\n${result.data.message}\n`;
                }
                
                // Show warnings
                if (result.data.warnings && result.data.warnings.length > 0) {
                    message += `\nWarnings:\n${result.data.warnings.join('\n')}`;
                }
            }
            
            // Add delay for backend operation to complete
            setTimeout(() => {
                onRefresh();
            }, 500);
            alert(message);
        } catch (error) {
            console.error('Error adding sites to cluster:', error);
            alert(`Error adding sites to cluster: ${error.message}`);
        } finally {
            setIsAddingToCluster(false);
        }
    };

    const handleRemoveSiteFromCluster = async (alias) => {
        if (window.confirm(`Are you sure you want to remove ${alias} from the replication cluster?`)) {
            try {
                const result = await removeIndividualSiteFromReplicationSmart(alias);
                
                // Show detailed result
                let message = `Site "${alias}" removal completed:\n\n`;
                if (result.results && result.results.length > 0) {
                    const siteResult = result.results[0];
                    if (siteResult.success) {
                        message += `✅ Successfully removed from cluster\n`;
                        if (siteResult.message) {
                            message += `${siteResult.message}\n`;
                        }
                    } else {
                        message += `❌ Failed to remove: ${siteResult.error || 'Unknown error'}`;
                    }
                }
                
                // Add a small delay to ensure backend operation completes
                setTimeout(() => {
                    onRefresh();
                }, 500);
                alert(message);
            } catch (error) {
                alert(`Error removing ${alias} from cluster: ${error.message}`);
            }
        }
    };

    const handleBulkRemoveFromCluster = async () => {
        if (selectedSitesToRemove.length === 0) {
            alert('Please select sites to remove');
            return;
        }

        if (window.confirm(`Remove ${selectedSitesToRemove.length} sites from replication cluster using smart removal?`)) {
            try {
                const result = await removeBulkSitesFromReplicationSmart(selectedSitesToRemove);
                setSelectedSitesToRemove([]);
                
                // Show detailed result
                let message = `Bulk removal of ${selectedSitesToRemove.length} sites completed:\n\n`;
                if (result.results && result.results.length > 0) {
                    const successful = result.results.filter(r => r.success);
                    const failed = result.results.filter(r => !r.success);
                    
                    if (successful.length > 0) {
                        message += `✅ Successfully removed: ${successful.map(r => r.alias).join(', ')}\n`;
                    }
                    if (failed.length > 0) {
                        message += `❌ Failed to remove: ${failed.map(r => r.alias).join(', ')}\n`;
                        message += `\nErrors:\n${failed.map(r => `- ${r.alias}: ${r.error}`).join('\n')}`;
                    }
                }
                
                // Add delay for backend operation to complete
                setTimeout(() => {
                    onRefresh();
                }, 500);
                alert(message);
            } catch (error) {
                alert(`Error removing sites from cluster: ${error.message}`);
            }
        }
    };

    const handleAddSingleSiteToCluster = async (alias) => {
        // Check for split brain first
        try {
            const splitBrainStatus = await checkSplitBrainStatus();
            if (splitBrainStatus.splitBrainDetected) {
                let errorMsg = '⚠️ SPLIT BRAIN DETECTED - Cannot add sites!\n\n';
                errorMsg += `${splitBrainStatus.clusterCount} separate clusters found.\n`;
                errorMsg += 'Please resolve the split brain scenario first.\n\n';
                errorMsg += 'Check the warning above for detailed instructions.';
                alert(errorMsg);
                return;
            }
        } catch (error) {
            console.error('Error checking split brain status:', error);
            // Continue with operation if check fails
        }

        if (window.confirm(`Add site "${alias}" to replication cluster using smart detection?`)) {
            console.log('Adding single site to cluster (smart):', alias);
            try {
                const result = await addSitesToReplicationSmart([alias]);
                
                // Show detailed result to user for single site add
                let message = `Smart add operation for "${alias}" completed:\n\n`;
                if (result.data) {
                    message += `Operation: ${result.data.action}\n`;
                    if (result.data.clustersFound !== undefined) {
                        message += `Clusters detected: ${result.data.clustersFound}\n`;
                    }
                    if (result.data.alreadyInCluster && result.data.alreadyInCluster.includes(alias)) {
                        message += `⚠️ "${alias}" was already in the cluster\n`;
                    } else if (result.data.newAliases && result.data.newAliases.includes(alias)) {
                        message += `✅ "${alias}" successfully added to cluster\n`;
                    }
                    if (result.data.existingCluster) {
                        message += `Total sites in cluster: ${result.data.existingCluster.sites ? result.data.existingCluster.sites.length : 'unknown'}\n`;
                    }
                    if (result.data.warnings && result.data.warnings.length > 0) {
                        message += `\nWarnings:\n${result.data.warnings.join('\n')}`;
                    }
                }
                
                // Add delay for backend operation to complete
                setTimeout(() => {
                    onRefresh();
                }, 500);
                alert(message);
            } catch (error) {
                console.error('Error adding single site:', error);
                alert(`Error adding ${alias} to cluster: ${error.message}`);
            }
        }
    };

    const handleSiteToAddToggle = (alias) => {
        setSelectedSitesToAdd(prev => 
            prev.includes(alias) 
                ? prev.filter(a => a !== alias)
                : [...prev, alias]
        );
    };

    const handleSiteToRemoveToggle = (alias) => {
        setSelectedSitesToRemove(prev => 
            prev.includes(alias) 
                ? prev.filter(a => a !== alias)
                : [...prev, alias]
        );
    };

    const handleQuickResync = async (alias, direction) => {
        const replicatedSites = sites.filter(s => s.replicationEnabled && s.alias !== alias);
        if (replicatedSites.length === 0) {
            alert('No other sites available for resync');
            return;
        }

        const otherSite = replicatedSites[0].alias;
        const fromSite = direction === 'from' ? alias : otherSite;
        const toSite = direction === 'from' ? otherSite : alias;

        try {
            await resyncSiteReplication(fromSite, toSite);
            onRefresh();
            alert(`Resync operation started: ${fromSite} → ${toSite}`);
        } catch (error) {
            alert(`Error starting resync: ${error.message}`);
        }
    };

    // Calculate summary statistics
    const totalSites = sites.length;
    const configuredSites = sites.filter(site => site.replicationStatus === 'configured').length;
    const healthySites = sites.filter(site => site.healthy).length;
    const availableSites = totalSites - configuredSites;

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('sites')}</h2>
            </div>

            {/* Summary Statistics Section */}
            {hasReplication && (
                <div className="stats-summary">
                    <div className="stat-card">
                        <div className="stat-value">{configuredSites}</div>
                        <div className="stat-label">Sites in Cluster</div>
                        <div className="stat-summary">Active replication sites</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-value">{healthySites}</div>
                        <div className="stat-label">Healthy Sites</div>
                        <div className="stat-summary">Sites responding normally</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-value">{availableSites}</div>
                        <div className="stat-label">Available to Add</div>
                        <div className="stat-summary">Sites ready for replication</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-value">{replicationInfo?.sites?.length || 0}</div>
                        <div className="stat-label">Total Endpoints</div>
                        <div className="stat-summary">Configured replication endpoints</div>
                    </div>
                </div>
            )}

            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">{t('site_replication_config')}</h3>
                    {hasReplication && (
                        <span className="badge badge-success">✓ Configured</span>
                    )}
                </div>

                {/* Split Brain Warning Component */}
                <SplitBrainWarning onRefresh={onRefresh} />

                {!hasReplication ? (
                    <div>
                        <p className="card-subtitle">{t('setup_replication_desc')}</p>
                        
                        <div className="form-group">
                            <label className="form-label">{t('select_aliases')}</label>
                            <div style={{ marginBottom: '16px' }}>
                                {sites.map(site => (
                                    <label key={site.name} style={{ display: 'block', marginBottom: '8px' }}>
                                        <input
                                            type="checkbox"
                                            checked={selectedAliases.includes(site.name)}
                                            onChange={() => handleAliasToggle(site.name)}
                                            style={{ marginRight: '8px' }}
                                        />
                                        {site.name} ({site.url})
                                    </label>
                                ))}
                            </div>
                        </div>

                        {selectedAliases.length > 0 && (
                            <div className="form-group">
                                <label className="form-label">{t('selected_order')}</label>
                                <div style={{ padding: '8px', background: 'var(--bg-secondary)', borderRadius: '4px' }}>
                                    {selectedAliases.length === 0 ? (
                                        <span style={{ color: 'var(--text-muted)' }}>{t('no_selection')}</span>
                                    ) : (
                                        selectedAliases.join(' → ')
                                    )}
                                </div>
                            </div>
                        )}

                        <button 
                            className="btn btn-primary"
                            onClick={handleAddReplication}
                            disabled={selectedAliases.length < 2 || isAddingReplication}
                        >
                            <Plus size={16} />
                            {isAddingReplication ? 'Setting up...' : t('add_sites')}
                        </button>
                    </div>
                ) : (
                    <div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
                            <p className="card-subtitle">{t('manage_replication_desc')}</p>
                            <div style={{ display: 'flex', gap: '12px' }}>
                                <button 
                                    className="btn btn-primary"
                                    onClick={() => setShowResyncModal(true)}
                                >
                                    <RefreshCw size={16} />
                                    Resync
                                </button>
                            </div>
                        </div>

                        {/* Add Sites to Existing Cluster */}
                        {sites.filter(s => !s.replicationEnabled).length > 0 && (
                            <div className="card" style={{ marginBottom: '24px' }}>
                                <div className="card-header">
                                    <h4 className="card-title">{t('add_sites_to_cluster')}</h4>
                                </div>
                                
                                {/* Smart Add Info Box */}
                                <div style={{ 
                                    padding: '12px 16px', 
                                    background: '#e8f4fd', 
                                    border: '1px solid #b8daff',
                                    borderRadius: '4px',
                                    margin: '16px',
                                    fontSize: '0.875rem'
                                }}>
                                    <strong>🧠 Smart Add Feature:</strong> Automatically detects existing clusters and intelligently:
                                    <ul style={{ margin: '8px 0 0 20px', padding: 0 }}>
                                        <li>Creates new cluster if no clusters exist</li>
                                        <li>Adds to existing cluster if one cluster found</li>
                                        <li>Prevents split-brain scenarios with multiple clusters</li>
                                        <li>Filters out sites already in clusters</li>
                                    </ul>
                                </div>
                                
                                <div className="table-container">
                                    <table className="table">
                                        <thead>
                                            <tr>
                                                <th style={{ width: '40px' }}>
                                                    <input 
                                                        type="checkbox" 
                                                        onChange={(e) => {
                                                            const availableSites = sites.filter(s => !s.replicationEnabled);
                                                            if (e.target.checked) {
                                                                setSelectedSitesToAdd(availableSites.map(site => site.name));
                                                            } else {
                                                                setSelectedSitesToAdd([]);
                                                            }
                                                        }}
                                                        title="Select all available sites"
                                                    />
                                                </th>
                                                <th style={{ width: '200px' }}>Site Name</th>
                                                <th style={{ width: '250px' }}>Endpoint</th>
                                                <th style={{ width: '100px' }}>Health</th>
                                                <th style={{ width: '100px' }}>Status</th>
                                                <th style={{ width: '120px' }}>Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {sites.filter(s => !s.replicationEnabled).map(site => (
                                                <tr key={site.name}>
                                                    <td>
                                                        <input 
                                                            type="checkbox"
                                                            checked={selectedSitesToAdd.includes(site.name)}
                                                            onChange={() => handleSiteToAddToggle(site.name)}
                                                        />
                                                    </td>
                                                    <td>
                                                        <div className="site-name" style={{ fontWeight: 'bold' }}>{site.name}</div>
                                                    </td>
                                                    <td>
                                                        <div className="site-url" style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>{site.url}</div>
                                                    </td>
                                                    <td>
                                                        <span className={`badge ${site.healthy ? 'badge-success' : 'badge-danger'}`}>
                                                            {site.healthy ? '● Healthy' : '● Unhealthy'}
                                                        </span>
                                                    </td>
                                                    <td>
                                                        <span className="badge badge-warning">Available</span>
                                                    </td>
                                                    <td>
                                                        <button 
                                                            className="btn btn-primary btn-sm"
                                                            onClick={() => handleAddSingleSiteToCluster(site.name)}
                                                            title="Smart add this site to replication cluster with automatic cluster detection"
                                                        >
                                                            <Plus size={14} />
                                                            Smart Add
                                                        </button>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                                
                                <div className="add-sites-actions">
                                    <div className="selection-info" style={{ marginBottom: '10px', fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                                        {selectedSitesToAdd.length === 0 ? (
                                            "Select sites to add to replication cluster"
                                        ) : selectedSitesToAdd.length === 1 ? (
                                            `1 site selected: ${selectedSitesToAdd[0]}`
                                        ) : (
                                            `${selectedSitesToAdd.length} sites selected: ${selectedSitesToAdd.join(', ')}`
                                        )}
                                    </div>
                                    <button 
                                        className="btn btn-primary"
                                        onClick={handleAddToCluster}
                                        disabled={selectedSitesToAdd.length === 0 || isAddingToCluster}
                                        title={selectedSitesToAdd.length === 0 ? "Select at least one site" : `Add ${selectedSitesToAdd.length} site${selectedSitesToAdd.length > 1 ? 's' : ''} using smart detection`}
                                    >
                                        <Plus size={16} />
                                        {isAddingToCluster ? 'Adding...' : 
                                         selectedSitesToAdd.length === 0 ? t('add_selected_to_cluster') :
                                         selectedSitesToAdd.length === 1 ? `Smart Add "${selectedSitesToAdd[0]}"` :
                                         `Smart Add ${selectedSitesToAdd.length} Sites`}
                                    </button>
                                </div>
                            </div>
                        )}

                        {/* Current Cluster Sites */}
                        <div className="card">
                            <div className="card-header">
                                <h4 className="card-title">{t('current_cluster')}</h4>
                                {selectedSitesToRemove.length > 0 && (
                                    <button 
                                        className="btn btn-secondary"
                                        style={{ color: 'var(--danger-color)', borderColor: 'var(--danger-color)' }}
                                        onClick={handleBulkRemoveFromCluster}
                                    >
                                        <Trash2 size={16} />
                                        {t('remove_selected')} ({selectedSitesToRemove.length})
                                    </button>
                                )}
                            </div>

                            <div className="table-container">
                                <table className="table">
                                    <thead>
                                        <tr>
                                            <th style={{ width: '40px' }}>
                                                <input 
                                                    type="checkbox" 
                                                    onChange={(e) => {
                                                        const clusterSites = sites.filter(site => site.replicationEnabled);
                                                        if (e.target.checked) {
                                                            setSelectedSitesToRemove(clusterSites.map(site => site.name));
                                                        } else {
                                                            setSelectedSitesToRemove([]);
                                                        }
                                                    }}
                                                    title="Select all sites"
                                                />
                                            </th>
                                            <th style={{ width: '200px' }}>Site Name</th>
                                            <th style={{ width: '250px' }}>Endpoint</th>
                                            <th style={{ width: '100px' }}>Health</th>
                                            <th style={{ width: '100px' }}>Status</th>
                                            <th style={{ width: '120px' }}>Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {sites.filter(site => site.replicationEnabled).map(site => (
                                            <tr key={site.name}>
                                                <td>
                                                    <input 
                                                        type="checkbox"
                                                        checked={selectedSitesToRemove.includes(site.name)}
                                                        onChange={(e) => {
                                                            if (e.target.checked) {
                                                                setSelectedSitesToRemove(prev => [...prev, site.name]);
                                                            } else {
                                                                setSelectedSitesToRemove(prev => prev.filter(a => a !== site.name));
                                                            }
                                                        }}
                                                    />
                                                </td>
                                                <td>
                                                    <div>
                                                        <div className="site-name" style={{ fontWeight: 'bold', marginBottom: '2px' }}>{site.name}</div>
                                                        {site.deploymentID && (
                                                            <div style={{ 
                                                                fontSize: '0.75rem', 
                                                                color: 'var(--text-muted)',
                                                                fontFamily: 'monospace'
                                                            }}>
                                                                ID: {site.deploymentID}
                                                            </div>
                                                        )}
                                                    </div>
                                                </td>
                                                <td>
                                                    <div className="site-url" style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>{site.url}</div>
                                                </td>
                                                <td>
                                                    <span className={`badge ${site.healthy ? 'badge-success' : 'badge-danger'}`}>
                                                        {site.healthy ? '● Healthy' : '● Unhealthy'}
                                                    </span>
                                                </td>
                                                <td>
                                                    <span className="badge badge-success">✓ Active</span>
                                                </td>
                                                <td>
                                                    <div className="action-buttons">
                                                        <button 
                                                            className="btn-icon"
                                                            onClick={() => handleResyncSite(site.name, 'from')}
                                                            title="Resync FROM this site (pull data)"
                                                        >
                                                            <Download size={16} />
                                                        </button>
                                                        
                                                        <button 
                                                            className="btn-icon"
                                                            onClick={() => handleResyncSite(site.name, 'to')}
                                                            title="Resync TO this site (push data)"
                                                        >
                                                            <Upload size={16} />
                                                        </button>
                                                        
                                                        <button 
                                                            className="btn-danger-icon"
                                                            onClick={() => handleRemoveSiteFromCluster(site.name)}
                                                            title="Remove this site from replication cluster"
                                                        >
                                                            <Trash2 size={16} />
                                                        </button>
                                                    </div>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* Resync Modal */}
            {showResyncModal && (
                <div className="modal active">
                    <div className="modal-content">
                        <div className="modal-header">
                            <h3 className="modal-title">Resync Site Replication</h3>
                            <button 
                                className="modal-close"
                                onClick={() => setShowResyncModal(false)}
                            >
                                ×
                            </button>
                        </div>
                        <div style={{ marginBottom: '20px' }}>
                            <p style={{ marginBottom: '16px' }}>
                                Select source and target sites for replication resync. 
                                This will copy data from source to target site.
                            </p>
                            
                            <div className="form-group">
                                <label className="form-label">Source Site (copy from)</label>
                                <select 
                                    className="form-input"
                                    value={resyncFromSite}
                                    onChange={(e) => setResyncFromSite(e.target.value)}
                                >
                                    <option value="">Select source site...</option>
                                    {sites.filter(s => s.replicationEnabled).map(site => (
                                        <option key={site.name} value={site.name}>
                                            {site.name} ({site.url})
                                        </option>
                                    ))}
                                </select>
                            </div>
                            
                            <div className="form-group">
                                <label className="form-label">Target Site (copy to)</label>
                                <select 
                                    className="form-input"
                                    value={resyncToSite}
                                    onChange={(e) => setResyncToSite(e.target.value)}
                                >
                                    <option value="">Select target site...</option>
                                    {sites.filter(s => s.replicationEnabled && s.name !== resyncFromSite).map(site => (
                                        <option key={site.name} value={site.name}>
                                            {site.name} ({site.url})
                                        </option>
                                    ))}
                                </select>
                            </div>
                        </div>
                        <div style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
                            <button 
                                className="btn btn-secondary"
                                onClick={() => setShowResyncModal(false)}
                            >
                                Cancel
                            </button>
                            <button 
                                className="btn btn-primary"
                                onClick={handleResyncReplication}
                                disabled={!resyncFromSite || !resyncToSite}
                            >
                                <RefreshCw size={16} />
                                Start Resync
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default SitesPage;