import React from 'react';
import { Link } from 'react-router-dom';
import { Activity, ArrowRight } from 'lucide-react';
import { useI18n } from '../utils/i18n';

const TracingPage = ({ sites }) => {
    const { t } = useI18n();

    const tracingCategories = [
        {
            id: 'trace-analyzer',
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
            path: '/tracing/analyzer',
            color: '#2563eb'
        }
    ];

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('tracing', 'Tracing')}</h2>
                <p style={{ margin: '8px 0 0 0', color: 'var(--text-secondary)' }}>
                    {t('tracing_intro', 'Trace and analyze MinIO API calls, errors, and performance metrics')}
                </p>
            </div>

            {/* Guidelines - Moved to top */}
            <div className="card" style={{ marginBottom: '24px' }}>
                <div className="card-header">
                    <h3 className="card-title">{t('tracing_guidelines', 'Tracing Guidelines')}</h3>
                </div>
                <div style={{ padding: '20px' }}>
                    <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        <li>{t('tracing_guideline_duration', 'Trace captures run for a specified duration (default 10-15 seconds)')}</li>
                        <li>{t('tracing_guideline_realtime', 'Real-time analysis of API calls and errors as they occur')}</li>
                        <li>{t('tracing_guideline_filter', 'Filter results by status code, API operation, or client')}</li>
                        <li>{t('tracing_guideline_performance', 'Minimal performance impact during trace capture')}</li>
                        <li>{t('tracing_guideline_export', 'Export trace data for external analysis tools')}</li>
                    </ul>
                </div>
            </div>

            <div style={{ display: 'grid', gap: '16px' }}>
                {tracingCategories.map(category => {
                    const Icon = category.icon;
                    const title = t(category.titleKey, category.titleFallback);
                    const description = t(category.descriptionKey, category.descriptionFallback);
                    const features = category.featureKeys.map((featureKey, index) =>
                        t(featureKey, category.featureFallbacks[index])
                    );
                    
                    return (
                        <div key={category.id} className="card" style={{ 
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

export default TracingPage;
