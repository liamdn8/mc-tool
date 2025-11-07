import React, { useState } from 'react';
import { Play } from 'lucide-react';
import { useI18n } from '../utils/i18n';
import { performConsistencyCheck } from '../utils/api';

const ConsistencyPage = ({ sites, replicationInfo }) => {
    const { t } = useI18n();
    const [isRunning, setIsRunning] = useState(false);
    const [results, setResults] = useState(null);
    const [selectedBuckets, setSelectedBuckets] = useState([]);

    const hasReplication = replicationInfo && replicationInfo.enabled;

    const handleRunCheck = async () => {
        setIsRunning(true);
        try {
            const checkResults = await performConsistencyCheck(selectedBuckets);
            setResults(checkResults);
        } catch (error) {
            alert(t('error_running_consistency', `Error running consistency check: ${error.message}`, { error: error.message }));
        } finally {
            setIsRunning(false);
        }
    };

    const replicatedSites = sites.filter(s => s.replicationEnabled);
    const totalBuckets = replicationInfo?.replicationGroup?.sites?.reduce((acc, site) => {
        if (site.buckets) {
            site.buckets.forEach(bucket => acc.add(bucket));
        }
        return acc;
    }, new Set())?.size || 0;

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('consistency_check', 'Consistency Check')}</h2>
            </div>

            {!hasReplication ? (
                <div className="card">
                    <div style={{ textAlign: 'center', padding: '40px' }}>
                        <h3>{t('consistency_no_config_title', 'No Site Replication Configured')}</h3>
                        <p>{t('consistency_no_config_description', 'Consistency checks are only available when site replication is configured.')}</p>
                    </div>
                </div>
            ) : (
                <div>
                    <div className="stats-grid">
                        <div className="stat-card">
                            <div className="stat-value">{replicatedSites.length}</div>
                            <div className="stat-label">{t('consistency_stats_sites_label', 'Sites to Check')}</div>
                            <div className="stat-summary">{t('consistency_stats_sites_summary', 'In replication group')}</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">{totalBuckets}</div>
                            <div className="stat-label">{t('consistency_stats_buckets_label', 'Buckets to Verify')}</div>
                            <div className="stat-summary">{t('consistency_stats_buckets_summary', 'Across all sites')}</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">
                                {results?.totalObjects || 0}
                            </div>
                            <div className="stat-label">{t('consistency_stats_objects_label', 'Objects Checked')}</div>
                            <div className="stat-summary">{t('consistency_stats_objects_summary', 'Last run')}</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">
                                <span className={`badge ${
                                    results ? 
                                        (results.inconsistencies === 0 ? 'badge-success' : 'badge-warning') :
                                        'badge-secondary'
                                }`}>
                                    {results ? 
                                        (results.inconsistencies === 0
                                            ? t('consistency_status_consistent', 'Consistent')
                                            : t('consistency_status_issues', `${results.inconsistencies} Issues`, { count: results.inconsistencies })
                                        ) :
                                        t('consistency_status_not_run', 'Not Run')
                                    }
                                </span>
                            </div>
                            <div className="stat-label">{t('consistency_stats_status_label', 'Status')}</div>
                            <div className="stat-summary">{t('consistency_stats_status_summary', 'Consistency status')}</div>
                        </div>
                    </div>

                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">{t('consistency_run_title', 'Run Consistency Check')}</h3>
                            <button 
                                className="btn btn-primary"
                                onClick={handleRunCheck}
                                disabled={isRunning}
                            >
                                <Play size={16} />
                                {isRunning ? t('consistency_running', 'Running...') : t('run_check', 'Run Check')}
                            </button>
                        </div>

                        <p>
                            {t('consistency_run_description', 'This will verify that all objects and metadata are consistent across all sites in the replication group.')}
                        </p>

                        {results && (
                            <div style={{ marginTop: '20px' }}>
                                <h4>{t('consistency_results_title', 'Last Check Results')}</h4>
                                <div className="table-container">
                                    <table className="table">
                                        <thead>
                                            <tr>
                                                <th>{t('consistency_table_metric', 'Metric')}</th>
                                                <th>{t('consistency_table_value', 'Value')}</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            <tr>
                                                <td>{t('consistency_metric_total_objects', 'Total Objects Checked')}</td>
                                                <td>{results.totalObjects}</td>
                                            </tr>
                                            <tr>
                                                <td>{t('consistency_metric_buckets', 'Buckets Verified')}</td>
                                                <td>{results.bucketsChecked}</td>
                                            </tr>
                                            <tr>
                                                <td>{t('consistency_metric_inconsistencies', 'Inconsistencies Found')}</td>
                                                <td>
                                                    <span className={`badge ${results.inconsistencies === 0 ? 'badge-success' : 'badge-warning'}`}>
                                                        {results.inconsistencies}
                                                    </span>
                                                </td>
                                            </tr>
                                            <tr>
                                                <td>{t('consistency_metric_duration', 'Check Duration')}</td>
                                                <td>{results.duration || '-'}</td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </div>

                                {results.details && results.details.length > 0 && (
                                    <div style={{ marginTop: '20px' }}>
                                        <h4>{t('consistency_issues_title', 'Issues Found')}</h4>
                                        <div className="table-container">
                                            <table className="table">
                                                <thead>
                                                    <tr>
                                                        <th>{t('consistency_issue_bucket', 'Bucket')}</th>
                                                        <th>{t('consistency_issue_object', 'Object')}</th>
                                                        <th>{t('consistency_issue_description', 'Issue')}</th>
                                                        <th>{t('consistency_issue_sites', 'Sites Affected')}</th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    {results.details.map((issue, index) => (
                                                        <tr key={index}>
                                                            <td>{issue.bucket}</td>
                                                            <td>{issue.object}</td>
                                                            <td>{issue.description}</td>
                                                            <td>{issue.sites?.join(', ')}</td>
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
                </div>
            )}
        </div>
    );
};

export default ConsistencyPage;