import React, { useState } from 'react';
import { Settings, Activity, CheckCircle, Play } from 'lucide-react';
import { useI18n } from '../../utils/i18n';

const SiteOperations = ({ hasReplication }) => {
    const { t } = useI18n();
    const [runningOperations, setRunningOperations] = useState(new Set());

    const replicationOperations = [
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
                headers: { 'Content-Type': 'application/json' }
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
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Site Replication Operations</h3>
                    <p style={{ margin: '8px 0 0 0', fontSize: '14px', color: '#6c757d' }}>
                        Operations that require site replication to be configured
                    </p>
                </div>

                <div style={{ padding: '20px' }}>
                    {!hasReplication && (
                        <div style={{ 
                            marginBottom: '20px', 
                            padding: '16px',
                            backgroundColor: '#fff3cd', 
                            border: '1px solid #ffeaa7',
                            borderRadius: '8px'
                        }}>
                            <h4 style={{ margin: '0 0 8px 0', color: '#856404' }}>⚠️ Site Replication Not Configured</h4>
                            <p style={{ margin: 0, color: '#856404' }}>
                                These operations require site replication to be configured. 
                                Please set up site replication before using these features.
                            </p>
                        </div>
                    )}

                    <div style={{ display: 'grid', gap: '16px' }}>
                        {replicationOperations.map(operation => {
                            const Icon = operation.icon;
                            const isRunning = runningOperations.has(operation.id);
                            const isDisabled = isRunning || !hasReplication;
                            
                            return (
                                <div key={operation.id} style={{ 
                                    padding: '20px',
                                    border: '1px solid #e9ecef',
                                    borderRadius: '8px',
                                    backgroundColor: 'white',
                                    opacity: !hasReplication ? 0.6 : 1
                                }}>
                                    <div style={{ display: 'flex', alignItems: 'flex-start', gap: '16px' }}>
                                        <Icon size={24} style={{ 
                                            color: 'var(--primary-color)', 
                                            flexShrink: 0, 
                                            marginTop: '4px' 
                                        }} />
                                        <div style={{ flex: 1 }}>
                                            <h4 style={{ margin: '0 0 8px 0', fontSize: '16px' }}>
                                                {operation.title}
                                                {!hasReplication && (
                                                    <span style={{ 
                                                        fontSize: '12px', 
                                                        marginLeft: '8px', 
                                                        color: '#856404',
                                                        backgroundColor: '#fff3cd',
                                                        padding: '2px 8px',
                                                        borderRadius: '12px'
                                                    }}>
                                                        Requires Site Replication
                                                    </span>
                                                )}
                                            </h4>
                                            <p style={{ 
                                                margin: '0 0 16px 0', 
                                                color: 'var(--text-secondary)' 
                                            }}>
                                                {operation.description}
                                            </p>
                                            
                                            <button 
                                                className="btn btn-primary"
                                                onClick={() => executeOperation(operation)}
                                                disabled={isDisabled}
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
            </div>
        </div>
    );
};

export default SiteOperations;