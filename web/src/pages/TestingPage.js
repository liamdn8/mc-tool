import React from 'react';
import { Link } from 'react-router-dom';
import { FlaskConical, ArrowRight } from 'lucide-react';
import { useI18n } from '../utils/i18n';

const TestingPage = ({ sites }) => {
    const { t } = useI18n();

    const testingCategories = [
        {
            id: 'performance',
            titleKey: 'testing_performance_title',
            titleFallback: 'Performance Testing',
            descriptionKey: 'testing_performance_description',
            descriptionFallback: 'Test upload performance with various configurations to analyze throughput and latency',
            featureKeys: [
                'testing_performance_feature_1',
                'testing_performance_feature_2',
                'testing_performance_feature_3',
                'testing_performance_feature_4',
                'testing_performance_feature_5'
            ],
            featureFallbacks: [
                'Upload mode: parallel or timed rounds',
                'Configurable object sizes (small, medium, large)',
                'Override support for versioning tests',
                'Auto-generated paths with timestamps',
                'Detailed performance metrics and throughput'
            ],
            icon: FlaskConical,
            path: '/testing/performance',
            color: '#7C3AED'
        }
    ];

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('testing', 'Testing')}</h2>
                <p style={{ margin: '8px 0 0 0', color: 'var(--text-secondary)' }}>
                    {t('testing_intro', 'Test and analyze upload performance across your MinIO infrastructure')}
                </p>
            </div>

            {/* Guidelines */}
            <div className="card" style={{ marginBottom: '24px' }}>
                <div className="card-header">
                    <h3 className="card-title">{t('testing_guidelines', 'Testing Guidelines')}</h3>
                </div>
                <div style={{ padding: '20px' }}>
                    <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        <li>{t('testing_guideline_alias', 'Select an alias (site) to run tests against')}</li>
                        <li>{t('testing_guideline_bucket', 'Specify a bucket for test objects')}</li>
                        <li>{t('testing_guideline_mode', 'Choose upload mode: all-at-once (parallel) or timed rounds (interval)')}</li>
                        <li>{t('testing_guideline_size', 'Configure object size and count based on test goals')}</li>
                        <li>{t('testing_guideline_metrics', 'Review metrics: throughput, latency, success rate')}</li>
                    </ul>
                </div>
            </div>

            <div style={{ display: 'grid', gap: '16px' }}>
                {testingCategories.map(category => {
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
                                background: category.color
                            }} />
                            <div style={{ padding: '24px', paddingLeft: '32px' }}>
                                <div style={{ display: 'flex', alignItems: 'flex-start', gap: '16px', marginBottom: '16px' }}>
                                    <div style={{
                                        width: '48px',
                                        height: '48px',
                                        borderRadius: '12px',
                                        background: `${category.color}15`,
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        flexShrink: 0
                                    }}>
                                        <Icon size={24} style={{ color: category.color }} />
                                    </div>
                                    <div style={{ flex: 1 }}>
                                        <h3 style={{ margin: 0, marginBottom: '8px', fontSize: '18px', fontWeight: 600 }}>
                                            {title}
                                        </h3>
                                        <p style={{ margin: 0, color: 'var(--text-secondary)', fontSize: '14px', lineHeight: 1.6 }}>
                                            {description}
                                        </p>
                                    </div>
                                </div>

                                <div style={{ 
                                    background: 'var(--bg-secondary)', 
                                    borderRadius: '8px', 
                                    padding: '16px',
                                    marginBottom: '20px'
                                }}>
                                    <div style={{ 
                                        fontSize: '12px', 
                                        fontWeight: 600, 
                                        color: 'var(--text-secondary)', 
                                        marginBottom: '12px',
                                        textTransform: 'uppercase',
                                        letterSpacing: '0.5px'
                                    }}>
                                        {t('features', 'Features')}
                                    </div>
                                    <ul style={{ margin: 0, paddingLeft: '20px' }}>
                                        {features.map((feature, index) => (
                                            <li key={index} style={{ 
                                                marginBottom: index < features.length - 1 ? '8px' : 0,
                                                color: 'var(--text-secondary)',
                                                fontSize: '13px'
                                            }}>
                                                {feature}
                                            </li>
                                        ))}
                                    </ul>
                                </div>

                                <Link 
                                    to={category.path}
                                    style={{
                                        display: 'inline-flex',
                                        alignItems: 'center',
                                        gap: '8px',
                                        padding: '10px 20px',
                                        background: category.color,
                                        color: 'white',
                                        borderRadius: '6px',
                                        textDecoration: 'none',
                                        fontSize: '14px',
                                        fontWeight: 500,
                                        transition: 'opacity 0.2s',
                                        border: 'none',
                                        cursor: 'pointer'
                                    }}
                                    onMouseOver={(e) => e.currentTarget.style.opacity = '0.9'}
                                    onMouseOut={(e) => e.currentTarget.style.opacity = '1'}
                                >
                                    {t('launch', 'Launch')}
                                    <ArrowRight size={16} />
                                </Link>
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export default TestingPage;
