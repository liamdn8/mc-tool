import React, { useState, useEffect } from 'react';
import { useI18n } from '../utils/i18n';
import { loadReplicationStatus } from '../utils/api';
import { getBadgeClass, formatDate } from '../utils/helpers';

const statusKeyMap = {
    healthy: 'status_healthy',
    unhealthy: 'status_unhealthy',
    configured: 'status_configured',
    not_configured: 'status_not_configured',
    fully_replicated: 'status_fully_replicated',
    partial_replication: 'status_partial_replication',
    completed: 'status_completed',
    pending: 'status_pending',
    failed: 'status_failed',
    success: 'status_success',
};

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
                <h2 className="card-title">{t('replication_status', 'Replication Status')}</h2>
                <button className="btn btn-secondary" onClick={loadReplicationStatusData}>
                    {t('refresh', 'Refresh')}
                </button>
            </div>

            {!hasReplication ? (
                <div className="card">
                    <div style={{ textAlign: 'center', padding: '40px' }}>
                        <h3>{t('replication_no_config_title', 'No Site Replication Configured')}</h3>
                        <p>{t('replication_no_config_description', 'Go to the Sites page to set up site replication.')}</p>
                    </div>
                </div>
            ) : (
                <div>
                    <div className="stats-grid">
                        <div className="stat-card">
                            <div className="stat-value">{sites.filter(s => s.replicationEnabled).length}</div>
                            <div className="stat-label">{t('replication_stats_sites_label', 'Sites in Replication')}</div>
                            <div className="stat-summary">{t('replication_stats_sites_summary', 'Active replication group')}</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">
                                <span className={`badge ${getBadgeClass(replicationInfo.health)}`}>
                                    {t(statusKeyMap[replicationInfo.health] || 'status_unknown', replicationInfo.health || '-')}
                                </span>
                            </div>
                            <div className="stat-label">{t('replication_stats_overall_label', 'Overall Status')}</div>
                            <div className="stat-summary">{t('replication_stats_overall_summary', 'Replication health')}</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">{replicationInfo.syncedBuckets || 0}</div>
                            <div className="stat-label">{t('replication_stats_buckets_label', 'Synced Buckets')}</div>
                            <div className="stat-summary">{t('replication_stats_buckets_summary', 'Across all sites')}</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">{replicationInfo.totalObjects || 0}</div>
                            <div className="stat-label">{t('replication_stats_objects_label', 'Total Objects')}</div>
                            <div className="stat-summary">{t('replication_stats_objects_summary', 'In replication')}</div>
                        </div>
                    </div>

                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">{t('replication_details_title', 'Site Replication Details')}</h3>
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
                                            <th>{t('replication_table_site', 'Site')}</th>
                                            <th>{t('replication_table_deployment', 'Deployment ID')}</th>
                                            <th>{t('replication_table_status', 'Status')}</th>
                                            <th>{t('replication_table_last_sync', 'Last Sync')}</th>
                                            <th>{t('replication_table_buckets', 'Buckets')}</th>
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
                                                        {t(statusKeyMap[site.replicationStatus] || 'status_unknown', site.replicationStatus || '-')}
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
                                <h3 className="card-title">{t('replication_bucket_section_title', 'Replicated Buckets')}</h3>
                            </div>
                            <div className="table-container">
                                <table className="table">
                                    <thead>
                                        <tr>
                                            <th>{t('replication_bucket_name', 'Bucket Name')}</th>
                                            <th>{t('replication_bucket_sites', 'Sites')}</th>
                                            <th>{t('replication_bucket_status', 'Sync Status')}</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {replicationStatus.replicationGroup.buckets.map(bucket => (
                                            <tr key={bucket.name}>
                                                <td>{bucket.name}</td>
                                                <td>
                                                    {t(
                                                        'replication_bucket_sites_value',
                                                        `${bucket.sites?.length || 0} sites`,
                                                        { count: bucket.sites?.length || 0 }
                                                    )}
                                                </td>
                                                <td>
                                                    <span className={`badge ${getBadgeClass(bucket.status)}`}>
                                                        {t(statusKeyMap[bucket.status] || 'status_unknown', bucket.status || '-')}
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