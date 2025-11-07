import React, { useState, useEffect } from 'react';
import { Plus, Trash2, RefreshCw } from 'lucide-react';
import { useI18n } from '../utils/i18n';
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
            alert(t('sites_select_two_aliases', 'Please select at least 2 aliases'));
            return;
        }

        setIsAddingReplication(true);
        try {
            await addSitesToReplication(selectedAliases);
            setSelectedAliases([]);
            onRefresh();
        } catch (error) {
            alert(
                t('error_setting_up_replication', 'Error setting up replication: {error}', {
                    error: error.message
                })
            );
        } finally {
            setIsAddingReplication(false);
        }
    };

    const handleResyncReplication = async () => {
        if (!resyncFromSite || !resyncToSite) {
            alert(t('resync_select_sites', 'Please select both source and target sites'));
            return;
        }

        try {
            await resyncSiteReplication(resyncFromSite, resyncToSite);
            setShowResyncModal(false);
            setResyncFromSite('');
            setResyncToSite('');
            onRefresh();
            alert(t('resync_started_success', 'Resync operation started successfully'));
        } catch (error) {
            alert(
                t('error_starting_resync', 'Error starting resync: {error}', {
                    error: error.message
                })
            );
        }
    };

    const handleAddToCluster = async () => {
        if (selectedSitesToAdd.length === 0) {
            alert(t('sites_select_one_to_add', 'Please select at least one site to add'));
            return;
        }

        try {
            const splitBrainStatus = await checkSplitBrainStatus();
            if (splitBrainStatus.splitBrainDetected) {
                alert(
                    t(
                        'split_brain_block_add',
                        '⚠️ SPLIT BRAIN DETECTED - Cannot add sites!\n\n{count} separate clusters found.\nPlease resolve the split brain scenario first.\n\nCheck the warning above for detailed instructions.',
                        { count: splitBrainStatus.clusterCount }
                    )
                );
                return;
            }
        } catch (error) {
            console.error('Error checking split brain status:', error);
        }

        const siteCount = selectedSitesToAdd.length;
        const confirmMessage = siteCount === 1
            ? t('confirm_add_single_site', 'Add site "{alias}" to replication cluster using smart detection?', {
                alias: selectedSitesToAdd[0]
            })
            : t(
                'confirm_add_multiple_sites',
                'Add {count} sites ({sites}) to replication cluster using smart detection?',
                {
                    count: siteCount,
                    sites: selectedSitesToAdd.join(', ')
                }
            );

        if (!window.confirm(confirmMessage)) {
            return;
        }

        setIsAddingToCluster(true);
        try {
            const result = await addSitesToReplicationSmart(selectedSitesToAdd);
            setSelectedSitesToAdd([]);

            const lines = [
                t('smart_cluster_result_title', 'Smart cluster operation for {count} site(s) completed:', {
                    count: siteCount
                }),
                ''
            ];

            if (result.data) {
                const action = result.data.action || result.data.operation;
                if (action) {
                    lines.push(
                        t('smart_cluster_result_operation', 'Operation: {value}', {
                            value: action
                        })
                    );
                }
                if (result.data.clustersFound !== undefined) {
                    lines.push(
                        t('smart_cluster_result_clusters', 'Clusters detected: {value}', {
                            value: result.data.clustersFound
                        })
                    );
                }
                if (result.data.newAliases && result.data.newAliases.length > 0) {
                    lines.push(
                        t('smart_cluster_result_sites_added', '✅ Sites added: {value}', {
                            value: result.data.newAliases.join(', ')
                        })
                    );
                }
                if (result.data.alreadyInCluster && result.data.alreadyInCluster.length > 0) {
                    lines.push(
                        t('smart_cluster_result_already_in_cluster', '⚠️ Already in cluster: {value}', {
                            value: result.data.alreadyInCluster.join(', ')
                        })
                    );
                }
                if (result.data.existingCluster && result.data.existingCluster.sites) {
                    lines.push(
                        t('smart_cluster_result_total_sites', 'Total sites in cluster: {value}', {
                            value: result.data.existingCluster.sites.length
                        })
                    );
                }
                if (result.data.message) {
                    lines.push('');
                    lines.push(result.data.message);
                }
                if (result.data.warnings && result.data.warnings.length > 0) {
                    lines.push('');
                    lines.push(t('smart_cluster_result_warnings', 'Warnings:'));
                    lines.push(...result.data.warnings);
                }
            }

            setTimeout(() => {
                onRefresh();
            }, 500);
            alert(lines.join('\n'));
        } catch (error) {
            console.error('Error adding sites to cluster:', error);
            alert(
                t('error_adding_sites_cluster', 'Error adding sites to cluster: {error}', {
                    error: error.message
                })
            );
        } finally {
            setIsAddingToCluster(false);
        }
    };

    const handleRemoveSiteFromCluster = async (alias) => {
        if (window.confirm(t('confirm_remove_site', 'Are you sure you want to remove {alias} from the replication cluster?', { alias }))) {
            try {
                const result = await removeIndividualSiteFromReplicationSmart(alias);

                const lines = [
                    t('site_removal_result_title', 'Site "{alias}" removal completed:', { alias }),
                    ''
                ];

                if (result.results && result.results.length > 0) {
                    const siteResult = result.results[0];
                    if (siteResult.success) {
                        lines.push(t('site_removed_success', '✅ Successfully removed from cluster'));
                        if (siteResult.message) {
                            lines.push(siteResult.message);
                        }
                    } else {
                        lines.push(
                            t('site_removed_failure', '❌ Failed to remove: {error}', {
                                error: siteResult.error || t('unknown_error', 'Unknown error')
                            })
                        );
                    }
                }

                setTimeout(() => {
                    onRefresh();
                }, 500);
                alert(lines.join('\n'));
            } catch (error) {
                alert(
                    t('error_removing_site', 'Error removing {alias} from cluster: {error}', {
                        alias,
                        error: error.message
                    })
                );
            }
        }
    };

    const handleBulkRemoveFromCluster = async () => {
        if (selectedSitesToRemove.length === 0) {
            alert(t('sites_select_to_remove', 'Please select sites to remove'));
            return;
        }

        if (window.confirm(
            t('confirm_bulk_remove_sites', 'Remove {count} sites from replication cluster using smart removal?', {
                count: selectedSitesToRemove.length
            })
        )) {
            try {
                const result = await removeBulkSitesFromReplicationSmart(selectedSitesToRemove);
                const sitesRemoved = selectedSitesToRemove.length;
                setSelectedSitesToRemove([]);

                const lines = [
                    t('bulk_removal_result_title', 'Bulk removal of {count} sites completed:', {
                        count: sitesRemoved
                    }),
                    ''
                ];

                if (result.results && result.results.length > 0) {
                    const successful = result.results.filter(r => r.success);
                    const failed = result.results.filter(r => !r.success);

                    if (successful.length > 0) {
                        lines.push(
                            t('bulk_removal_success', '✅ Successfully removed: {sites}', {
                                sites: successful.map(r => r.alias).join(', ')
                            })
                        );
                    }
                    if (failed.length > 0) {
                        lines.push(
                            t('bulk_removal_failure', '❌ Failed to remove: {sites}', {
                                sites: failed.map(r => r.alias).join(', ')
                            })
                        );
                        lines.push('');
                        lines.push(t('bulk_removal_errors', 'Errors:'));
                        lines.push(
                            ...failed.map(r => `- ${r.alias}: ${r.error}`)
                        );
                    }
                }

                setTimeout(() => {
                    onRefresh();
                }, 500);
                alert(lines.join('\n'));
            } catch (error) {
                alert(
                    t('error_removing_sites', 'Error removing sites from cluster: {error}', {
                        error: error.message
                    })
                );
            }
        }
    };

    const handleAddSingleSiteToCluster = async (alias) => {
        try {
            const splitBrainStatus = await checkSplitBrainStatus();
            if (splitBrainStatus.splitBrainDetected) {
                alert(
                    t(
                        'split_brain_block_add',
                        '⚠️ SPLIT BRAIN DETECTED - Cannot add sites!\n\n{count} separate clusters found.\nPlease resolve the split brain scenario first.\n\nCheck the warning above for detailed instructions.',
                        { count: splitBrainStatus.clusterCount }
                    )
                );
                return;
            }
        } catch (error) {
            console.error('Error checking split brain status:', error);
        }

        if (window.confirm(
            t('confirm_add_single_site', 'Add site "{alias}" to replication cluster using smart detection?', { alias })
        )) {
            try {
                const result = await addSitesToReplicationSmart([alias]);

                const lines = [
                    t('smart_single_result_title', 'Smart add operation for "{alias}" completed:', { alias }),
                    ''
                ];

                if (result.data) {
                    if (result.data.action) {
                        lines.push(
                            t('smart_cluster_result_operation', 'Operation: {value}', {
                                value: result.data.action
                            })
                        );
                    }
                    if (result.data.clustersFound !== undefined) {
                        lines.push(
                            t('smart_cluster_result_clusters', 'Clusters detected: {value}', {
                                value: result.data.clustersFound
                            })
                        );
                    }
                    if (result.data.alreadyInCluster && result.data.alreadyInCluster.includes(alias)) {
                        lines.push(
                            t('smart_single_result_already', '⚠️ "{alias}" was already in the cluster', {
                                alias
                            })
                        );
                    } else if (result.data.newAliases && result.data.newAliases.includes(alias)) {
                        lines.push(
                            t('smart_single_result_added', '✅ "{alias}" successfully added to cluster', {
                                alias
                            })
                        );
                    }
                    if (result.data.existingCluster) {
                        const totalSites = result.data.existingCluster.sites ? result.data.existingCluster.sites.length : 'unknown';
                        lines.push(
                            t('smart_cluster_result_total_sites', 'Total sites in cluster: {value}', {
                                value: totalSites
                            })
                        );
                    }
                    if (result.data.warnings && result.data.warnings.length > 0) {
                        lines.push('');
                        lines.push(t('smart_cluster_result_warnings', 'Warnings:'));
                        lines.push(...result.data.warnings);
                    }
                }

                setTimeout(() => {
                    onRefresh();
                }, 500);
                alert(lines.join('\n'));
            } catch (error) {
                console.error('Error adding single site:', error);
                alert(
                    t('error_adding_site_cluster', 'Error adding {alias} to cluster: {error}', {
                        alias,
                        error: error.message
                    })
                );
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
                        <div className="stat-label">{t('sites_stats_in_cluster_label', 'Sites in Cluster')}</div>
                        <div className="stat-summary">{t('sites_stats_in_cluster_summary', 'Active replication sites')}</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-value">{healthySites}</div>
                        <div className="stat-label">{t('sites_stats_healthy_label', 'Healthy Sites')}</div>
                        <div className="stat-summary">{t('sites_stats_healthy_summary', 'Sites responding normally')}</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-value">{availableSites}</div>
                        <div className="stat-label">{t('sites_stats_available_label', 'Available to Add')}</div>
                        <div className="stat-summary">{t('sites_stats_available_summary', 'Sites ready for replication')}</div>
                    </div>
                    <div className="stat-card">
                        <div className="stat-value">{replicationInfo?.sites?.length || 0}</div>
                        <div className="stat-label">{t('sites_stats_endpoints_label', 'Total Endpoints')}</div>
                        <div className="stat-summary">{t('sites_stats_endpoints_summary', 'Configured replication endpoints')}</div>
                    </div>
                </div>
            )}

            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">{t('site_replication_config')}</h3>
                    {hasReplication && (
                        <span className="badge badge-success">{t('badge_configured', '✓ Configured')}</span>
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
                            {isAddingReplication ? t('setup_in_progress', 'Setting up...') : t('add_sites')}
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
                                    {t('resync_button', 'Resync')}
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
                                    <strong>{t('smart_add_title', '🧠 Smart Add Feature:')}</strong> {t('smart_add_description', 'Automatically detects existing clusters and intelligently:')}
                                    <ul style={{ margin: '8px 0 0 20px', padding: 0 }}>
                                        <li>{t('smart_add_create_cluster', 'Creates new cluster if no clusters exist')}</li>
                                        <li>{t('smart_add_use_existing', 'Adds to existing cluster if one cluster found')}</li>
                                        <li>{t('smart_add_prevent_split', 'Prevents split-brain scenarios with multiple clusters')}</li>
                                        <li>{t('smart_add_filter', 'Filters out sites already in clusters')}</li>
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
                                                            title={t('tooltip_select_all_available', 'Select all available sites')}
                                                    />
                                                </th>
                                                    <th style={{ width: '200px' }}>{t('column_site_name', 'Site Name')}</th>
                                                    <th style={{ width: '250px' }}>{t('column_endpoint', 'Endpoint')}</th>
                                                    <th style={{ width: '100px' }}>{t('column_health', 'Health')}</th>
                                                    <th style={{ width: '100px' }}>{t('column_status', 'Status')}</th>
                                                    <th style={{ width: '120px' }}>{t('column_actions', 'Actions')}</th>
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
                                                            {site.healthy ? t('badge_healthy_icon', '● Healthy') : t('badge_unhealthy_icon', '● Unhealthy')}
                                                        </span>
                                                    </td>
                                                    <td>
                                                        <span className="badge badge-warning">{t('status_available', 'Available')}</span>
                                                    </td>
                                                    <td>
                                                        <button 
                                                            className="btn btn-primary btn-sm"
                                                            onClick={() => handleAddSingleSiteToCluster(site.name)}
                                                            title={t('tooltip_smart_add_single', 'Smart add this site to replication cluster with automatic cluster detection')}
                                                        >
                                                            <Plus size={14} />
                                                            {t('button_smart_add', 'Smart Add')}
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
                                            t('select_sites_to_add', 'Select sites to add to replication cluster')
                                        ) : selectedSitesToAdd.length === 1 ? (
                                            t('one_site_selected', '1 site selected: {site}', { site: selectedSitesToAdd[0] })
                                        ) : (
                                            t('multiple_sites_selected', '{count} sites selected: {sites}', {
                                                count: selectedSitesToAdd.length,
                                                sites: selectedSitesToAdd.join(', ')
                                            })
                                        )}
                                    </div>
                                    <button 
                                        className="btn btn-primary"
                                        onClick={handleAddToCluster}
                                        disabled={selectedSitesToAdd.length === 0 || isAddingToCluster}
                                        title={selectedSitesToAdd.length === 0 
                                            ? t('tooltip_select_site_before_add', 'Select at least one site')
                                            : t('tooltip_smart_add_bulk', 'Add {count} site(s) using smart detection', {
                                                count: selectedSitesToAdd.length
                                            })}
                                    >
                                        <Plus size={16} />
                                        {isAddingToCluster
                                            ? t('smart_add_in_progress', 'Adding...')
                                            : selectedSitesToAdd.length === 0
                                                ? t('add_selected_to_cluster', 'Add selected sites')
                                                : selectedSitesToAdd.length === 1
                                                    ? t('button_smart_add_single', 'Smart Add "{alias}"', { alias: selectedSitesToAdd[0] })
                                                    : t('button_smart_add_multiple', 'Smart Add {count} Sites', { count: selectedSitesToAdd.length })}
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
                                                    title={t('tooltip_select_all_sites', 'Select all sites')}
                                                />
                                            </th>
                                            <th style={{ width: '200px' }}>{t('column_site_name', 'Site Name')}</th>
                                            <th style={{ width: '250px' }}>{t('column_endpoint', 'Endpoint')}</th>
                                            <th style={{ width: '100px' }}>{t('column_health', 'Health')}</th>
                                            <th style={{ width: '100px' }}>{t('column_status', 'Status')}</th>
                                            <th style={{ width: '120px' }}>{t('column_actions', 'Actions')}</th>
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
                                                        {site.healthy ? t('badge_healthy_icon', '● Healthy') : t('badge_unhealthy_icon', '● Unhealthy')}
                                                    </span>
                                                </td>
                                                <td>
                                                    <span className="badge badge-success">{t('status_active', '✓ Active')}</span>
                                                </td>
                                                <td>
                                                    <div className="action-buttons">
                                                        <button 
                                                            className="btn-danger-icon"
                                                            onClick={() => handleRemoveSiteFromCluster(site.name)}
                                                            title={t('tooltip_remove_site', 'Remove this site from replication cluster')}
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
                            <h3 className="modal-title">{t('modal_resync_title', 'Resync Site Replication')}</h3>
                            <button 
                                className="modal-close"
                                onClick={() => setShowResyncModal(false)}
                            >
                                ×
                            </button>
                        </div>
                        <div style={{ marginBottom: '20px' }}>
                            <p style={{ marginBottom: '16px' }}>
                                {t('modal_resync_description', 'Select source and target sites for replication resync. This will copy data from source to target site.')}
                            </p>
                            
                            <div className="form-group">
                                <label className="form-label">{t('modal_resync_source_label', 'Source Site (copy from)')}</label>
                                <select 
                                    className="form-input"
                                    value={resyncFromSite}
                                    onChange={(e) => setResyncFromSite(e.target.value)}
                                >
                                    <option value="">{t('modal_resync_source_placeholder', 'Select source site...')}</option>
                                    {sites.filter(s => s.replicationEnabled).map(site => (
                                        <option key={site.name} value={site.name}>
                                            {site.name} ({site.url})
                                        </option>
                                    ))}
                                </select>
                            </div>
                            
                            <div className="form-group">
                                <label className="form-label">{t('modal_resync_target_label', 'Target Site (copy to)')}</label>
                                <select 
                                    className="form-input"
                                    value={resyncToSite}
                                    onChange={(e) => setResyncToSite(e.target.value)}
                                >
                                    <option value="">{t('modal_resync_target_placeholder', 'Select target site...')}</option>
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
                                {t('cancel', 'Cancel')}
                            </button>
                            <button 
                                className="btn btn-primary"
                                onClick={handleResyncReplication}
                                disabled={!resyncFromSite || !resyncToSite}
                            >
                                <RefreshCw size={16} />
                                {t('modal_resync_start', 'Start Resync')}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default SitesPage;