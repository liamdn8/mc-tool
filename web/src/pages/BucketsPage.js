import React, { useState, useEffect } from 'react';
import { useI18n } from '../utils/i18n';
import { loadBuckets } from '../utils/api';
import { formatNumber, formatBytes } from '../utils/helpers';

const replicationStatusKey = {
    fully_replicated: 'status_fully_replicated',
    partial_replication: 'status_partial_replication',
    not_configured: 'status_not_replicated',
};

const BucketsPage = ({ sites, replicationInfo }) => {
    const { t } = useI18n();
    const [buckets, setBuckets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedSite, setSelectedSite] = useState('');

    useEffect(() => {
        if (sites.length > 0 && !selectedSite) {
            setSelectedSite(sites[0].alias);
        }
    }, [sites]);

    useEffect(() => {
        if (selectedSite) {
            loadBucketsData();
        }
    }, [selectedSite]);

    const loadBucketsData = async () => {
        if (!selectedSite) return;
        
        setLoading(true);
        try {
            const bucketsData = await loadBuckets(selectedSite);
            setBuckets(bucketsData);
        } catch (error) {
            console.error('Error loading buckets:', error);
            setBuckets([]);
        } finally {
            setLoading(false);
        }
    };

    const getBucketReplicationStatus = (bucketName) => {
        if (!replicationInfo?.replicationGroup?.sites) {
            return 'not_configured';
        }

        const sitesWithBucket = replicationInfo.replicationGroup.sites.filter(site => 
            site.buckets && site.buckets.includes(bucketName)
        );
        
        if (sitesWithBucket.length === 0) return 'not_configured';
        if (sitesWithBucket.length === replicationInfo.replicationGroup.sites.length) return 'fully_replicated';
        return 'partial_replication';
    };

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('buckets_overview', 'Buckets Overview')}</h2>
            </div>

            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">{t('buckets_select_site', 'Select Site')}</h3>
                    <select 
                        className="form-input" 
                        value={selectedSite} 
                        onChange={(e) => setSelectedSite(e.target.value)}
                        style={{ width: '200px' }}
                    >
                        {sites.map(site => (
                            <option key={site.alias} value={site.alias}>
                                {site.alias}
                            </option>
                        ))}
                    </select>
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
                                    <th>{t('buckets_table_name', 'Bucket Name')}</th>
                                    <th>{t('buckets_table_objects', 'Objects')}</th>
                                    <th>{t('buckets_table_size', 'Size')}</th>
                                    <th>{t('buckets_table_status', 'Replication Status')}</th>
                                    <th>{t('buckets_table_created', 'Created')}</th>
                                </tr>
                            </thead>
                            <tbody>
                                {buckets.length === 0 ? (
                                    <tr>
                                        <td colSpan="5" style={{ textAlign: 'center', padding: '40px' }}>
                                            {t('buckets_empty', `No buckets found in ${selectedSite}`, { site: selectedSite })}
                                        </td>
                                    </tr>
                                ) : (
                                    buckets.map(bucket => {
                                        const replicationStatus = getBucketReplicationStatus(bucket.name);
                                        return (
                                            <tr key={bucket.name}>
                                                <td>{bucket.name}</td>
                                                <td>{formatNumber(bucket.objectCount || 0)}</td>
                                                <td>{formatBytes(bucket.size || 0)}</td>
                                                <td>
                                                    <span className={`badge ${
                                                        replicationStatus === 'fully_replicated' ? 'badge-success' :
                                                        replicationStatus === 'partial_replication' ? 'badge-warning' :
                                                        'badge-secondary'
                                                    }`}>
                                                        {t(replicationStatusKey[replicationStatus] || 'status_unknown', replicationStatus || '-')}
                                                    </span>
                                                </td>
                                                <td>{bucket.creationDate ? new Date(bucket.creationDate).toLocaleDateString() : '-'}</td>
                                            </tr>
                                        );
                                    })
                                )}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        </div>
    );
};

export default BucketsPage;