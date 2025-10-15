import React from 'react';
import { Link } from 'react-router-dom';
import { GitCompare, List, Zap, ArrowRight } from 'lucide-react';
import { useI18n } from '../utils/i18n';

const OperationsPage = ({ sites, replicationInfo }) => {
    const { t } = useI18n();
    const hasReplication = replicationInfo && replicationInfo.enabled;

    const operationCategories = [
        {
            id: 'compare',
            title: 'Compare Buckets/Paths',
            description: 'Compare content between two MinIO aliases to identify differences, missing files, and content mismatches',
            icon: GitCompare,
            path: '/operations/compare',
            color: '#4f46e5',
            features: [
                'Compare bucket contents across different sites',
                'Identify missing files in source or target',
                'Detect content differences with detailed reports',
                'Support for path-specific comparisons',
                'Paginated results with customizable page sizes'
            ],
            requiresReplication: false
        },
        {
            id: 'checklist',
            title: 'Configuration Checklist',
            description: 'Verify environment variables, events, and lifecycle configurations across all sites',
            icon: List,
            path: '/operations/checklist',
            color: '#059669',
            features: [
                'Validate environment variables across sites',
                'Check event notification configurations',
                'Verify bucket lifecycle policies',
                'Grouped results by configuration category',
                'Pass/Warning/Fail status indicators'
            ],
            requiresReplication: false
        },
        {
            id: 'site-operations',
            title: 'Site Replication Operations',
            description: 'Manage site replication, sync policies, and validate consistency across replicated sites',
            icon: Zap,
            path: '/operations/site-operations',
            color: '#dc2626',
            features: [
                'Sync bucket policies across sites',
                'Sync lifecycle configurations',
                'Validate replication consistency',
                'Health check for replicated sites'
            ],
            requiresReplication: true
        }
    ];

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('automated_operations')}</h2>
                <p style={{ margin: '8px 0 0 0', color: 'var(--text-secondary)' }}>
                    Select an operation category to manage your MinIO infrastructure
                </p>
            </div>

            <div style={{ marginBottom: '24px' }}>
                <div className="stats-grid">
                    <div className="stat-card">
                        <div className="stat-value">{sites.filter(s => s.replicationEnabled || true).length}</div>
                        <div className="stat-label">Available Sites</div>
                        <div className="stat-summary">Sites available for operations</div>
                    </div>

                    <div className="stat-card">
                        <div className="stat-value">{operationCategories.length}</div>
                        <div className="stat-label">Operation Categories</div>
                        <div className="stat-summary">Different types of operations</div>
                    </div>

                    <div className="stat-card">
                        <div className="stat-value">
                            {operationCategories.filter(op => !op.requiresReplication || hasReplication).length}
                        </div>
                        <div className="stat-label">Available Operations</div>
                        <div className="stat-summary">Currently accessible</div>
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
            </div>

            {!hasReplication && (
                <div className="card" style={{ marginBottom: '24px', backgroundColor: '#fff3cd', border: '1px solid #ffeaa7' }}>
                    <div style={{ padding: '16px' }}>
                        <h4 style={{ margin: '0 0 8px 0', color: '#856404' }}>⚠️ Site Replication Not Configured</h4>
                        <p style={{ margin: 0, color: '#856404' }}>
                            Some operations require site replication to be configured. 
                            However, compare and checklist operations can work with individual aliases.
                        </p>
                    </div>
                </div>
            )}

            <div style={{ display: 'grid', gap: '24px' }}>
                {operationCategories.map(category => {
                    const Icon = category.icon;
                    const isAvailable = !category.requiresReplication || hasReplication;
                    
                    return (
                        <div key={category.id} className="card" style={{ 
                            opacity: isAvailable ? 1 : 0.6,
                            position: 'relative',
                            overflow: 'hidden'
                        }}>
                            <div style={{ 
                                position: 'absolute',
                                top: 0,
                                left: 0,
                                width: '4px',
                                height: '100%',
                                backgroundColor: category.color
                            }} />
                            
                            <div style={{ padding: '24px' }}>
                                <div style={{ display: 'flex', alignItems: 'flex-start', gap: '20px' }}>
                                    <div style={{
                                        width: '60px',
                                        height: '60px',
                                        borderRadius: '12px',
                                        backgroundColor: category.color + '15',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        flexShrink: 0
                                    }}>
                                        <Icon size={28} style={{ color: category.color }} />
                                    </div>
                                    
                                    <div style={{ flex: 1 }}>
                                        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '12px' }}>
                                            <div>
                                                <h3 style={{ margin: '0 0 8px 0', fontSize: '20px', fontWeight: '600' }}>
                                                    {category.title}
                                                    {category.requiresReplication && !hasReplication && (
                                                        <span style={{ 
                                                            fontSize: '12px', 
                                                            marginLeft: '12px', 
                                                            color: '#856404',
                                                            backgroundColor: '#fff3cd',
                                                            padding: '4px 8px',
                                                            borderRadius: '12px'
                                                        }}>
                                                            Requires Site Replication
                                                        </span>
                                                    )}
                                                </h3>
                                                <p style={{ margin: '0', color: 'var(--text-secondary)', lineHeight: '1.5' }}>
                                                    {category.description}
                                                </p>
                                            </div>
                                        </div>
                                        
                                        <div style={{ marginBottom: '20px' }}>
                                            <h4 style={{ margin: '0 0 12px 0', fontSize: '14px', fontWeight: '600', color: '#6c757d' }}>
                                                Features:
                                            </h4>
                                            <ul style={{ margin: 0, paddingLeft: '20px', color: 'var(--text-secondary)' }}>
                                                {category.features.map((feature, index) => (
                                                    <li key={index} style={{ marginBottom: '4px', fontSize: '14px' }}>
                                                        {feature}
                                                    </li>
                                                ))}
                                            </ul>
                                        </div>
                                        
                                        {isAvailable ? (
                                            <Link 
                                                to={category.path}
                                                className="btn btn-primary"
                                                style={{ 
                                                    textDecoration: 'none',
                                                    display: 'inline-flex',
                                                    alignItems: 'center',
                                                    gap: '8px'
                                                }}
                                            >
                                                Open {category.title}
                                                <ArrowRight size={16} />
                                            </Link>
                                        ) : (
                                            <button 
                                                className="btn btn-secondary"
                                                disabled
                                                style={{ cursor: 'not-allowed' }}
                                            >
                                                Requires Site Replication
                                            </button>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>

            <div className="card" style={{ marginTop: '24px' }}>
                <div className="card-header">
                    <h3 className="card-title">Operation Guidelines</h3>
                </div>
                <div style={{ padding: '20px' }}>
                    <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        <li>Ensure all sites are healthy before running operations</li>
                        <li>Operations may take several minutes to complete</li>
                        <li>Check the logs for detailed operation results</li>
                        <li>Some operations may temporarily affect performance</li>
                        <li>Compare and checklist operations work without site replication</li>
                        <li>Use browser back/forward buttons - each operation has its own URL</li>
                    </ul>
                </div>
            </div>
        </div>
    );
};

export default OperationsPage;