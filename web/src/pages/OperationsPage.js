import React from 'react';
import { Link } from 'react-router-dom';
import { GitCompare, List, Zap, ArrowRight, Activity, AlertTriangle } from 'lucide-react';
import { useI18n } from '../utils/i18n';

const OperationsPage = ({ sites, replicationInfo }) => {
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
            path: '/operations/compare',
            color: '#4f46e5',
            requiresReplication: false
        },
        {
            id: 'trace',
            titleKey: 'operations_trace_title',
            titleFallback: 'Trace Error Analyzer',
            descriptionKey: 'operations_trace_description',
            descriptionFallback: 'Capture mc admin trace output to identify repeated object failures, filter by status code or message, and group by API or client for faster debugging',
            featureKeys: [
                'operations_trace_feature_1',
                'operations_trace_feature_2',
                'operations_trace_feature_3',
                'operations_trace_feature_4'
            ],
            featureFallbacks: [
                'Filter captured errors by HTTP status code or message content',
                'Group repeated failures by API action and impacted clients',
                'Highlight top affected objects with sample error messages',
                'Export raw trace events for follow-up investigation'
            ],
            icon: Activity,
            path: '/operations/trace',
            color: '#2563eb',
            requiresReplication: false
        },
        {
            id: 'checklist',
            titleKey: 'operations_checklist_title',
            titleFallback: 'Configuration Checklist',
            descriptionKey: 'operations_checklist_description',
            descriptionFallback: 'Verify environment variables, events, and lifecycle configurations across all sites',
            featureKeys: [
                'operations_checklist_feature_1',
                'operations_checklist_feature_2',
                'operations_checklist_feature_3',
                'operations_checklist_feature_4',
                'operations_checklist_feature_5'
            ],
            featureFallbacks: [
                'Validate environment variables across sites',
                'Check event notification configurations',
                'Verify bucket lifecycle policies',
                'Grouped results by configuration category',
                'Pass/Warning/Fail status indicators'
            ],
            icon: List,
            path: '/operations/checklist',
            color: '#059669',
            requiresReplication: false
        },
        {
            id: 'site-operations',
            titleKey: 'operations_site_title',
            titleFallback: 'Site Replication Operations',
            descriptionKey: 'operations_site_description',
            descriptionFallback: 'Manage site replication, sync policies, and validate consistency across replicated sites',
            featureKeys: [
                'operations_site_feature_1',
                'operations_site_feature_2',
                'operations_site_feature_3',
                'operations_site_feature_4'
            ],
            featureFallbacks: [
                'Sync bucket policies across sites',
                'Sync lifecycle configurations',
                'Validate replication consistency',
                'Health check for replicated sites'
            ],
            icon: Zap,
            path: '/operations/site-operations',
            color: '#dc2626',
            requiresReplication: true
        }
    ];

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('automated_operations')}</h2>
                <p style={{ margin: '8px 0 0 0', color: 'var(--text-secondary)' }}>
                    {t('operations_intro', 'Select an operation category to manage your MinIO infrastructure')}
                </p>
            </div>

            <div style={{ marginBottom: '24px' }}>
                <div className="stats-grid">
                    <div className="stat-card">
                        <div className="stat-value">{sites.filter(s => s.replicationEnabled || true).length}</div>
                        <div className="stat-label">{t('operations_stats_sites_label', 'Available Sites')}</div>
                        <div className="stat-summary">{t('operations_stats_sites_summary', 'Sites available for operations')}</div>
                    </div>

                    <div className="stat-card">
                        <div className="stat-value">{operationCategories.length}</div>
                        <div className="stat-label">{t('operations_stats_categories_label', 'Operation Categories')}</div>
                        <div className="stat-summary">{t('operations_stats_categories_summary', 'Different types of operations')}</div>
                    </div>

                    <div className="stat-card">
                        <div className="stat-value">
                            {operationCategories.filter(op => !op.requiresReplication || hasReplication).length}
                        </div>
                        <div className="stat-label">{t('operations_stats_available_label', 'Available Operations')}</div>
                        <div className="stat-summary">{t('operations_stats_available_summary', 'Currently accessible')}</div>
                    </div>

                    <div className="stat-card">
                        <div className="stat-value">
                            <span className={`badge ${sites.every(s => s.healthy) ? 'badge-success' : 'badge-warning'}`}>
                                {sites.every(s => s.healthy) ? 'Ready' : 'Issues'}
                            </span>
                        </div>
                        <div className="stat-label">{t('operation_status')}</div>
                        <div className="stat-summary">{t('operations_stats_status_summary', 'System readiness')}</div>
                    </div>
                </div>
            </div>

            {!hasReplication && (
                <div className="card" style={{ marginBottom: '24px', backgroundColor: '#fff3cd', border: '1px solid #ffeaa7' }}>
                    <div style={{ padding: '16px' }}>
                        <h4 style={{ margin: '0 0 8px 0', color: '#856404', display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <AlertTriangle size={18} />
                            <span>{t('operations_replication_warning_title', 'Site Replication Not Configured')}</span>
                        </h4>
                        <p style={{ margin: 0, color: '#856404' }}>
                            {t('operations_replication_warning_desc', 'Some operations require site replication to be configured. However, compare and checklist operations can work with individual aliases.')}
                        </p>
                    </div>
                </div>
            )}

            <div style={{ display: 'grid', gap: '24px' }}>
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
                                                    {title}
                                                    {category.requiresReplication && !hasReplication && (
                                                        <span style={{ 
                                                            fontSize: '12px', 
                                                            marginLeft: '12px', 
                                                            color: '#856404',
                                                            backgroundColor: '#fff3cd',
                                                            padding: '4px 8px',
                                                            borderRadius: '12px'
                                                        }}>
                                                            {t('operations_requires_replication', 'Requires Site Replication')}
                                                        </span>
                                                    )}
                                                </h3>
                                                <p style={{ margin: '0', color: 'var(--text-secondary)', lineHeight: '1.5' }}>
                                                    {description}
                                                </p>
                                            </div>
                                        </div>
                                        
                                        <div style={{ marginBottom: '20px' }}>
                                            <h4 style={{ margin: '0 0 12px 0', fontSize: '14px', fontWeight: '600', color: '#6c757d' }}>
                                                {t('operations_features_label', 'Features:')}
                                            </h4>
                                            <ul style={{ margin: 0, paddingLeft: '20px', color: 'var(--text-secondary)' }}>
                                                {features.map((feature, index) => (
                                                    <li key={category.featureKeys[index]} style={{ marginBottom: '4px', fontSize: '14px' }}>
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
                                                {t('operations_open_link', 'Open {title}', { title })}
                                                <ArrowRight size={16} />
                                            </Link>
                                        ) : (
                                            <button 
                                                className="btn btn-secondary"
                                                disabled
                                                style={{ cursor: 'not-allowed' }}
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

            <div className="card" style={{ marginTop: '24px' }}>
                <div className="card-header">
                    <h3 className="card-title">{t('operations_guidelines_title', 'Operation Guidelines')}</h3>
                </div>
                <div style={{ padding: '20px' }}>
                    <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        <li>{t('operations_guideline_health', 'Ensure all sites are healthy before running operations')}</li>
                        <li>{t('operations_guideline_duration', 'Operations may take several minutes to complete')}</li>
                        <li>{t('operations_guideline_logs', 'Check the logs for detailed operation results')}</li>
                        <li>{t('operations_guideline_performance', 'Some operations may temporarily affect performance')}</li>
                        <li>{t('operations_guideline_independent', 'Compare and checklist operations work without site replication')}</li>
                        <li>{t('operations_guideline_navigation', 'Use browser back/forward buttons - each operation has its own URL')}</li>
                    </ul>
                </div>
            </div>
        </div>
    );
};

export default OperationsPage;