import React, { useState, useEffect } from 'react';
import { useI18n } from '../utils/i18n';
import { loadReplicationStatus } from '../utils/api';
import { getBadgeClass, getStatusText, formatDate } from '../utils/helpers';

const ReplicationPage = ({ sites, replicationInfo, onRefresh }) => {
    const { t } = useI18n();
    const [replicationStatus, setReplicationStatus] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadReplicationStatusData();
    }, []);

    const loadReplicationStatusData = async () => {
        setLoading(true);
        try {
            const statusData = await loadReplicationStatus();
            setReplicationStatus(statusData);
        } catch (error) {
            console.error('Error loading replication status:', error);
        } finally {
            setLoading(false);
        }
    };

    const hasReplication = replicationInfo && replicationInfo.enabled;

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('replication_status')}</h2>
                <button className="btn btn-secondary" onClick={loadReplicationStatusData}>
                    {t('refresh')}
                </button>
            </div>

            {!hasReplication ? (
                <div className="card">
                    <div style={{ textAlign: 'center', padding: '40px' }}>
                        <h3>No Site Replication Configured</h3>
                        <p>Go to the Sites page to set up site replication.</p>
                    </div>
                </div>
            ) : (
                <div>
                    <div className="stats-grid">
                        <div className="stat-card">
                            <div className="stat-value">{sites.filter(s => s.replicationEnabled).length}</div>
                            <div className="stat-label">Sites in Replication</div>
                            <div className="stat-summary">Active replication group</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">
                                <span className={`badge ${getBadgeClass(replicationInfo.health)}`}>
                                    {getStatusText(replicationInfo.health)}
                                </span>
                            </div>
                            <div className="stat-label">Overall Status</div>
                            <div className="stat-summary">Replication health</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">{replicationInfo.syncedBuckets || 0}</div>
                            <div className="stat-label">Synced Buckets</div>
                            <div className="stat-summary">Across all sites</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">{replicationInfo.totalObjects || 0}</div>
                            <div className="stat-label">Total Objects</div>
                            <div className="stat-summary">In replication</div>
                        </div>
                    </div>

                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Site Replication Details</h3>
                        </div>

                        {loading ? (
                            <div className="loading">
                                <div className="spinner"></div>
                            </div>
                        ) : (
                            <div className="table-container">
                                <table className="table">
                                    <thead>
                                        <tr>
                                            <th>Site</th>
                                            <th>Deployment ID</th>
                                            <th>Status</th>
                                            <th>Last Sync</th>
                                            <th>Buckets</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {sites.filter(site => site.replicationEnabled).map(site => (
                                            <tr key={site.name}>
                                                <td>
                                                    <div>
                                                        <strong>{site.name}</strong>
                                                        <br />
                                                        <small style={{ color: 'var(--text-muted)' }}>{site.url}</small>
                                                    </div>
                                                </td>
                                                <td>
                                                    <code style={{ fontSize: '12px' }}>
                                                        {site.deploymentID || '-'}
                                                    </code>
                                                </td>
                                                <td>
                                                    <span className={`badge ${getBadgeClass(site.replicationStatus)}`}>
                                                        {getStatusText(site.replicationStatus)}
                                                    </span>
                                                </td>
                                                <td>
                                                    {formatDate(site.lastSync)}
                                                </td>
                                                <td>
                                                    {replicationStatus?.sites?.[site.name]?.bucketCount || 0}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                    </div>

                    {replicationStatus?.replicationGroup?.buckets && (
                        <div className="card">
                            <div className="card-header">
                                <h3 className="card-title">Replicated Buckets</h3>
                            </div>
                            <div className="table-container">
                                <table className="table">
                                    <thead>
                                        <tr>
                                            <th>Bucket Name</th>
                                            <th>Sites</th>
                                            <th>Sync Status</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {replicationStatus.replicationGroup.buckets.map(bucket => (
                                            <tr key={bucket.name}>
                                                <td>{bucket.name}</td>
                                                <td>{bucket.sites?.length || 0} sites</td>
                                                <td>
                                                    <span className={`badge ${getBadgeClass(bucket.status)}`}>
                                                        {getStatusText(bucket.status)}
                                                    </span>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

export default ReplicationPage;