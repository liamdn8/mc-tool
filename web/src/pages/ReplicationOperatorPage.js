import React from 'react';
import { Link } from 'react-router-dom';
import { GitCompare, Zap, ArrowRight } from 'lucide-react';
import { useI18n } from '../utils/i18n';

const ReplicationOperatorPage = ({ sites, replicationInfo }) => {
    const { t } = useI18n();
    const hasReplication = replicationInfo && replicationInfo.enabled;

    const operationCategories = [
        {
            id: 'compare',
            titleKey: 'operations_compare_title',
            titleFallback: 'Compare Buckets/Paths',
            descriptionKey: 'operations_compare_description',
            descriptionFallback: 'Compare content between two MinIO aliases to identify differences, missing files, and content mismatches',
            featureKeys: [
                'operations_compare_feature_1',
                'operations_compare_feature_2',
                'operations_compare_feature_3',
                'operations_compare_feature_4',
                'operations_compare_feature_5'
            ],
            featureFallbacks: [
                'Compare bucket contents across different sites',
                'Identify missing files in source or target',
                'Detect content differences with detailed reports',
                'Support for path-specific comparisons',
                'Paginated results with customizable page sizes'
            ],
            icon: GitCompare,
            path: '/replication-operator/compare',
            color: '#4f46e5',
            requiresReplication: false
        },
        {
            id: 'resync',
            titleKey: 'replication_resync_title',
            titleFallback: 'Replication Resync',
            descriptionKey: 'replication_resync_description',
            descriptionFallback: 'Manage site replication, sync policies, and validate consistency across replicated sites',
            featureKeys: [
                'replication_resync_feature_1',
                'replication_resync_feature_2',
                'replication_resync_feature_3',
                'replication_resync_feature_4'
            ],
            featureFallbacks: [
                'Sync bucket policies across sites',
                'Sync lifecycle configurations',
                'Validate replication consistency',
                'Health check for replicated sites'
            ],
            icon: Zap,
            path: '/replication-operator/resync',
            color: '#dc2626',
            requiresReplication: true
        }
    ];

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('replication_operator', 'Replication Operator')}</h2>
                <p style={{ margin: '8px 0 0 0', color: 'var(--text-secondary)' }}>
                    {t('replication_operator_intro', 'Manage replication operations including bucket comparison and synchronization')}
                </p>
            </div>

            {/* Guidelines - Moved to top */}
            <div className="card" style={{ marginBottom: '24px' }}>
                <div className="card-header">
                    <h3 className="card-title">{t('replication_operator_guidelines', 'Replication Operator Guidelines')}</h3>
                </div>
                <div style={{ padding: '20px' }}>
                    <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        <li>{t('operations_guideline_health', 'Ensure all sites are healthy before running operations')}</li>
                        <li>{t('operations_guideline_duration', 'Operations may take several minutes to complete')}</li>
                        <li>{t('operations_guideline_logs', 'Check the logs for detailed operation results')}</li>
                        <li>{t('replication_operator_guideline_compare', 'Compare operations work without site replication')}</li>
                        <li>{t('replication_operator_guideline_resync', 'Resync operations require active site replication')}</li>
                    </ul>
                </div>
            </div>

            {!hasReplication && (
                <div className="card" style={{ marginBottom: '24px', backgroundColor: '#fff3cd', border: '1px solid #ffeaa7' }}>
                    <div style={{ padding: '16px' }}>
                        <h4 style={{ margin: '0 0 8px 0', color: '#856404' }}>
                            {t('operations_replication_warning_title', '⚠️ Site Replication Not Configured')}
                        </h4>
                        <p style={{ margin: 0, color: '#856404' }}>
                            {t('replication_operator_warning', 'Some operations require site replication to be configured. Compare operations can work with individual aliases.')}
                        </p>
                    </div>
                </div>
            )}

            <div style={{ display: 'grid', gap: '16px' }}>
                {operationCategories.map(category => {
                    const Icon = category.icon;
                    const isAvailable = !category.requiresReplication || hasReplication;
                    const title = t(category.titleKey, category.titleFallback);
                    const description = t(category.descriptionKey, category.descriptionFallback);
                    const features = category.featureKeys.map((featureKey, index) =>
                        t(featureKey, category.featureFallbacks[index])
                    );
                    
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
                            
                            <div style={{ padding: '16px' }}>
                                <div style={{ display: 'flex', alignItems: 'flex-start', gap: '16px' }}>
                                    <div style={{
                                        width: '48px',
                                        height: '48px',
                                        borderRadius: '10px',
                                        backgroundColor: category.color + '15',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        flexShrink: 0
                                    }}>
                                        <Icon size={24} style={{ color: category.color }} />
                                    </div>
                                    
                                    <div style={{ flex: 1 }}>
                                        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '8px' }}>
                                            <div>
                                                <h3 style={{ margin: '0 0 4px 0', fontSize: '18px', fontWeight: '600' }}>
                                                    {title}
                                                    {category.requiresReplication && !hasReplication && (
                                                        <span style={{ 
                                                            fontSize: '11px', 
                                                            marginLeft: '10px', 
                                                            color: '#856404',
                                                            backgroundColor: '#fff3cd',
                                                            padding: '3px 6px',
                                                            borderRadius: '10px'
                                                        }}>
                                                            {t('operations_requires_replication', 'Requires Site Replication')}
                                                        </span>
                                                    )}
                                                </h3>
                                                <p style={{ margin: '0', color: 'var(--text-secondary)', lineHeight: '1.4', fontSize: '14px' }}>
                                                    {description}
                                                </p>
                                            </div>
                                        </div>
                                        
                                        <div style={{ marginBottom: '12px' }}>
                                            <ul style={{ margin: 0, paddingLeft: '16px', color: 'var(--text-secondary)' }}>
                                                {features.map((feature, index) => (
                                                    <li key={category.featureKeys[index]} style={{ marginBottom: '2px', fontSize: '13px' }}>
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
                                                    gap: '6px',
                                                    padding: '6px 14px',
                                                    fontSize: '14px'
                                                }}
                                            >
                                                {t('operations_open_link', 'Open {title}', { title })}
                                                <ArrowRight size={14} />
                                            </Link>
                                        ) : (
                                            <button 
                                                className="btn btn-secondary"
                                                disabled
                                                style={{ cursor: 'not-allowed', padding: '6px 14px', fontSize: '14px' }}
                                            >
                                                {t('operations_requires_replication', 'Requires Site Replication')}
                                            </button>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export default ReplicationOperatorPage;
