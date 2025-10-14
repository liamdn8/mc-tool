import React, { useState } from 'react';
import { Play, Settings, CheckCircle, Activity } from 'lucide-react';
import { useI18n } from '../utils/i18n';

const OperationsPage = ({ sites, replicationInfo }) => {
    const { t } = useI18n();
    const [runningOperations, setRunningOperations] = useState(new Set());

    const hasReplication = replicationInfo && replicationInfo.enabled;

    const operations = [
        {
            id: 'sync_policies',
            title: t('sync_bucket_policies'),
            description: t('sync_bucket_policies_desc'),
            icon: Settings,
            endpoint: '/api/operations/sync-policies'
        },
        {
            id: 'sync_lifecycle',
            title: t('sync_lifecycle'),
            description: t('sync_lifecycle_desc'),
            icon: Activity,
            endpoint: '/api/operations/sync-lifecycle'
        },
        {
            id: 'validate_consistency',
            title: t('validate_consistency'),
            description: t('validate_consistency_desc'),
            icon: CheckCircle,
            endpoint: '/api/operations/validate-consistency'
        },
        {
            id: 'health_check',
            title: t('health_check'),
            description: t('health_check_desc'),
            icon: Activity,
            endpoint: '/api/operations/health-check'
        }
    ];

    const executeOperation = async (operation) => {
        setRunningOperations(prev => new Set([...prev, operation.id]));
        
        try {
            const response = await fetch(operation.endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                }
            });
            
            const result = await response.json();
            
            if (response.ok) {
                alert(`${operation.title} completed successfully`);
            } else {
                alert(`${operation.title} failed: ${result.error || 'Unknown error'}`);
            }
        } catch (error) {
            alert(`${operation.title} failed: ${error.message}`);
        } finally {
            setRunningOperations(prev => {
                const newSet = new Set(prev);
                newSet.delete(operation.id);
                return newSet;
            });
        }
    };

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('automated_operations')}</h2>
            </div>

            {!hasReplication ? (
                <div className="card">
                    <div style={{ textAlign: 'center', padding: '40px' }}>
                        <h3>No Site Replication Configured</h3>
                        <p>Operations are only available when site replication is configured.</p>
                    </div>
                </div>
            ) : (
                <div>
                    <div className="stats-grid">
                        <div className="stat-card">
                            <div className="stat-value">{sites.filter(s => s.replicationEnabled).length}</div>
                            <div className="stat-label">Target Sites</div>
                            <div className="stat-summary">Operations will run on these sites</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">{operations.length}</div>
                            <div className="stat-label">Available Operations</div>
                            <div className="stat-summary">Automated maintenance tasks</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">{runningOperations.size}</div>
                            <div className="stat-label">Running Operations</div>
                            <div className="stat-summary">Currently executing</div>
                        </div>

                        <div className="stat-card">
                            <div className="stat-value">
                                <span className={`badge ${sites.every(s => s.healthy) ? 'badge-success' : 'badge-warning'}`}>
                                    {sites.every(s => s.healthy) ? 'Ready' : 'Issues'}
                                </span>
                            </div>
                            <div className="stat-label">{t('operation_status')}</div>
                            <div className="stat-summary">System readiness</div>
                        </div>
                    </div>

                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Available Operations</h3>
                        </div>

                        <div style={{ display: 'grid', gap: '16px' }}>
                            {operations.map(operation => {
                                const Icon = operation.icon;
                                const isRunning = runningOperations.has(operation.id);
                                
                                return (
                                    <div key={operation.id} className="card" style={{ margin: 0, padding: '20px' }}>
                                        <div style={{ display: 'flex', alignItems: 'flex-start', gap: '16px' }}>
                                            <Icon size={24} style={{ color: 'var(--primary-color)', flexShrink: 0, marginTop: '4px' }} />
                                            <div style={{ flex: 1 }}>
                                                <h4 style={{ margin: '0 0 8px 0', fontSize: '16px' }}>{operation.title}</h4>
                                                <p style={{ margin: '0 0 16px 0', color: 'var(--text-secondary)' }}>
                                                    {operation.description}
                                                </p>
                                                <button 
                                                    className="btn btn-primary"
                                                    onClick={() => executeOperation(operation)}
                                                    disabled={isRunning}
                                                >
                                                    <Play size={16} />
                                                    {isRunning ? 'Running...' : t('execute')}
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    </div>

                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Operation Guidelines</h3>
                        </div>
                        <ul style={{ margin: 0, paddingLeft: '20px' }}>
                            <li>Ensure all sites are healthy before running operations</li>
                            <li>Operations may take several minutes to complete</li>
                            <li>Check the logs for detailed operation results</li>
                            <li>Some operations may temporarily affect performance</li>
                        </ul>
                    </div>
                </div>
            )}
        </div>
    );
};

export default OperationsPage;