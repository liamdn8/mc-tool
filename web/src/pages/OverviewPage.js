import React from 'react';
import { useI18n } from '../utils/i18n';
import { formatNumber } from '../utils/helpers';

const OverviewPage = ({ sites, replicationInfo }) => {
    const { t } = useI18n();

    // Calculate stats
    const replicatedSites = sites.filter(s => s.replicationEnabled).length;
    const healthySites = sites.filter(s => s.healthy).length;
    
    // Get bucket count from replication info
    let totalBuckets = 0;
    if (replicationInfo && replicationInfo.replicationGroup && replicationInfo.replicationGroup.sites) {
        const bucketSet = new Set();
        replicationInfo.replicationGroup.sites.forEach(site => {
            if (site.buckets) {
                site.buckets.forEach(bucket => bucketSet.add(bucket));
            }
        });
        totalBuckets = bucketSet.size;
    }

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('site_replication_overview')}</h2>
            </div>

            <div className="stats-grid">
                <div className="stat-card">
                    <div className="stat-value">{formatNumber(sites.length)}</div>
                    <div className="stat-label">{t('total_sites')}</div>
                    <div className="stat-summary">
                        {replicatedSites > 0 
                            ? `${replicatedSites} in replication group`
                            : 'No replication configured'
                        }
                    </div>
                </div>

                <div className="stat-card">
                    <div className="stat-value">{formatNumber(totalBuckets)}</div>
                    <div className="stat-label">{t('synced_buckets')}</div>
                    <div className="stat-summary">
                        {totalBuckets > 0 
                            ? `Across ${replicatedSites} sites`
                            : 'No buckets synced'
                        }
                    </div>
                </div>

                <div className="stat-card">
                    <div className="stat-value">{formatNumber(replicationInfo?.totalObjects || 0)}</div>
                    <div className="stat-label">{t('total_objects')}</div>
                    <div className="stat-summary">
                        In all synced buckets
                    </div>
                </div>

                <div className="stat-card">
                    <div className="stat-value">
                        <span className={`badge ${healthySites === sites.length ? 'badge-success' : 'badge-warning'}`}>
                            {healthySites === sites.length ? t('healthy') : 'Issues'}
                        </span>
                    </div>
                    <div className="stat-label">{t('replication_health')}</div>
                    <div className="stat-summary">
                        {healthySites}/{sites.length} sites healthy
                    </div>
                </div>
            </div>

            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">{t('configured_aliases')}</h3>
                    <p className="card-subtitle">{t('manage_sites')}</p>
                </div>

                <div className="table-container">
                    <table className="table">
                        <thead>
                            <tr>
                                <th>{t('alias')}</th>
                                <th>{t('endpoint')}</th>
                                <th>{t('status')}</th>
                                <th>Replication</th>
                            </tr>
                        </thead>
                        <tbody>
                            {sites.map(site => (
                                <tr key={site.name}>
                                    <td>{site.name}</td>
                                    <td>{site.url}</td>
                                    <td>
                                        <span className={`badge ${site.healthy ? 'badge-success' : 'badge-danger'}`}>
                                            {site.healthy ? t('healthy') : 'Unhealthy'}
                                        </span>
                                    </td>
                                    <td>
                                        <span className={`badge ${site.replicationEnabled ? 'badge-success' : 'badge-secondary'}`}>
                                            {site.replicationEnabled ? t('replication_enabled') : t('replication_disabled')}
                                        </span>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

export default OverviewPage;