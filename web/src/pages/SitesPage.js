import React, { useState, useEffect } from 'react';
import { Plus, Trash2, RefreshCw, Download, Upload, X } from 'lucide-react';
import { useI18n } from '../utils/i18n';
import { getBadgeClass, getStatusText } from '../utils/helpers';
import { 
    loadAliases, 
    loadSiteReplicationInfo, 
    addSitesToReplication, 
    loadReplicationStatus, 
    resyncReplication,
    removeSiteFromReplication,
    removeBulkSitesFromReplication
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
            await addSiteReplication(selectedAliases);
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

        setIsAddingToCluster(true);
        try {
            await addSitesToReplication(selectedSitesToAdd);
            setSelectedSitesToAdd([]);
            onRefresh();
            alert('Sites added to cluster successfully');
        } catch (error) {
            alert(`Error adding sites to cluster: ${error.message}`);
        } finally {
            setIsAddingToCluster(false);
        }
    };

    const handleRemoveSiteFromCluster = async (alias) => {
        if (window.confirm(`Are you sure you want to remove ${alias} from the replication cluster?`)) {
            try {
                await removeSiteFromReplication(alias);
                onRefresh();
                alert(`${alias} removed from cluster successfully`);
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

        if (window.confirm(`Remove ${selectedSitesToRemove.length} sites from replication cluster?`)) {
            try {
                await removeBulkSitesFromReplication(selectedSitesToRemove);
                setSelectedSitesToRemove([]);
                onRefresh();
                alert('Selected sites removed from cluster successfully');
            } catch (error) {
                alert(`Error removing sites from cluster: ${error.message}`);
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

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('sites')}</h2>
            </div>

            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">{t('site_replication_config')}</h3>
                    {hasReplication && (
                        <span className="badge badge-success">✓ Configured</span>
                    )}
                </div>

                {!hasReplication ? (
                    <div>
                        <p className="card-subtitle">{t('setup_replication_desc')}</p>
                        
                        <div className="form-group">
                            <label className="form-label">{t('select_aliases')}</label>
                            <div style={{ marginBottom: '16px' }}>
                                {sites.map(site => (
                                    <label key={site.alias} style={{ display: 'block', marginBottom: '8px' }}>
                                        <input 
                                            type="checkbox"
                                            checked={selectedAliases.includes(site.alias)}
                                            onChange={() => handleAliasToggle(site.alias)}
                                            style={{ marginRight: '8px' }}
                                        />
                                        {site.alias} ({site.url})
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
                                <div>
                                    <div style={{ display: 'grid', gap: '8px', marginBottom: '16px' }}>
                                        {sites.filter(s => !s.replicationEnabled).map(site => (
                                            <label key={site.alias} style={{ 
                                                display: 'flex', 
                                                alignItems: 'center', 
                                                padding: '8px 12px', 
                                                border: '1px solid var(--border-color)', 
                                                borderRadius: '6px',
                                                backgroundColor: selectedSitesToAdd.includes(site.alias) ? 'var(--primary-light)' : 'transparent',
                                                cursor: 'pointer',
                                                gap: '12px'
                                            }}>
                                                <input 
                                                    type="checkbox"
                                                    checked={selectedSitesToAdd.includes(site.alias)}
                                                    onChange={() => handleSiteToAddToggle(site.alias)}
                                                    style={{ margin: 0, flexShrink: 0 }}
                                                />
                                                <div style={{ flex: 1, minWidth: 0 }}>
                                                    <div style={{ fontWeight: 'bold', marginBottom: '2px' }}>{site.alias}</div>
                                                    <div style={{ 
                                                        fontSize: '0.875rem', 
                                                        color: 'var(--text-muted)',
                                                        overflow: 'hidden',
                                                        textOverflow: 'ellipsis',
                                                        whiteSpace: 'nowrap',
                                                        marginBottom: '2px'
                                                    }}>{site.url}</div>
                                                    {site.deploymentID && (
                                                        <div style={{ 
                                                            fontSize: '0.75rem', 
                                                            color: 'var(--text-muted)',
                                                            fontFamily: 'monospace',
                                                            overflow: 'hidden',
                                                            textOverflow: 'ellipsis',
                                                            whiteSpace: 'nowrap'
                                                        }}>
                                                            ID: {site.deploymentID}
                                                        </div>
                                                    )}
                                                </div>
                                                <span className={`badge ${site.healthy ? 'badge-success' : 'badge-danger'}`} style={{ flexShrink: 0 }}>
                                                    {site.healthy ? '● Healthy' : '● Unhealthy'}
                                                </span>
                                            </label>
                                        ))}
                                    </div>
                                    <button 
                                        className="btn btn-primary"
                                        onClick={handleAddToCluster}
                                        disabled={selectedSitesToAdd.length === 0 || isAddingToCluster}
                                    >
                                        <Plus size={16} />
                                        {isAddingToCluster ? 'Adding...' : t('add_to_cluster')}
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

                            <div style={{ display: 'grid', gap: '12px' }}>
                                {sites.filter(site => site.replicationEnabled).map(site => (
                                    <div key={site.alias} style={{ 
                                        display: 'flex', 
                                        alignItems: 'center', 
                                        padding: '12px 16px', 
                                        border: '1px solid var(--border-color)', 
                                        borderRadius: '8px',
                                        backgroundColor: 'var(--card-bg)',
                                        gap: '12px'
                                    }}>
                                        <input 
                                            type="checkbox"
                                            checked={selectedSitesToRemove.includes(site.alias)}
                                            onChange={() => handleSiteToRemoveToggle(site.alias)}
                                            style={{ flexShrink: 0 }}
                                        />
                                        <div style={{ flex: 1, minWidth: 0 }}>
                                            <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>{site.alias}</div>
                                            <div style={{ 
                                                fontSize: '0.875rem', 
                                                color: 'var(--text-muted)',
                                                overflow: 'hidden',
                                                textOverflow: 'ellipsis',
                                                whiteSpace: 'nowrap',
                                                marginBottom: '2px'
                                            }}>{site.url}</div>
                                            {site.deploymentID && (
                                                <div style={{ 
                                                    fontSize: '0.75rem', 
                                                    color: 'var(--text-muted)',
                                                    fontFamily: 'monospace',
                                                    overflow: 'hidden',
                                                    textOverflow: 'ellipsis',
                                                    whiteSpace: 'nowrap'
                                                }}>
                                                    ID: {site.deploymentID}
                                                </div>
                                            )}
                                        </div>
                                        <span className={`badge ${getBadgeClass(site.replicationStatus)}`} style={{ flexShrink: 0 }}>
                                            ✓ Active
                                        </span>
                                        <div style={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
                                            <button 
                                                className="btn-icon"
                                                onClick={() => handleQuickResync(site.alias, 'from')}
                                                title="Resync FROM this site"
                                                style={{ padding: '6px' }}
                                            >
                                                <Download size={14} />
                                            </button>
                                            <button 
                                                className="btn-icon"
                                                onClick={() => handleQuickResync(site.alias, 'to')}
                                                title="Resync TO this site"
                                                style={{ padding: '6px' }}
                                            >
                                                <Upload size={14} />
                                            </button>
                                            <button 
                                                className="btn-icon"
                                                style={{ color: 'var(--danger-color)', padding: '6px' }}
                                                onClick={() => handleRemoveSiteFromCluster(site.alias)}
                                                title="Remove from cluster"
                                            >
                                                <X size={14} />
                                            </button>
                                        </div>
                                    </div>
                                ))}
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
                                        <option key={site.alias} value={site.alias}>
                                            {site.alias} ({site.url})
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
                                    {sites.filter(s => s.replicationEnabled && s.alias !== resyncFromSite).map(site => (
                                        <option key={site.alias} value={site.alias}>
                                            {site.alias} ({site.url})
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