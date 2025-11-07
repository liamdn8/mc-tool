import React from 'react';
import { Link } from 'react-router-dom';
import { List, ArrowRight } from 'lucide-react';
import { useI18n } from '../utils/i18n';

const ChecklistPage = ({ sites }) => {
    const { t } = useI18n();

    const checklistCategories = [
        {
            id: 'configuration',
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
            path: '/checklist/configuration',
            color: '#059669'
        }
    ];

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('checklist', 'Checklist')}</h2>
                <p style={{ margin: '8px 0 0 0', color: 'var(--text-secondary)' }}>
                    {t('checklist_intro', 'Verify system configurations and ensure compliance across your MinIO infrastructure')}
                </p>
            </div>

            {/* Guidelines - Moved to top */}
            <div className="card" style={{ marginBottom: '24px' }}>
                <div className="card-header">
                    <h3 className="card-title">{t('checklist_guidelines', 'Checklist Guidelines')}</h3>
                </div>
                <div style={{ padding: '20px' }}>
                    <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        <li>{t('checklist_guideline_comprehensive', 'Comprehensive verification across all configured sites')}</li>
                        <li>{t('checklist_guideline_categories', 'Results organized by configuration categories')}</li>
                        <li>{t('checklist_guideline_status', 'Clear Pass/Warning/Fail indicators for each check')}</li>
                        <li>{t('checklist_guideline_details', 'Detailed explanations for failed or warning checks')}</li>
                        <li>{t('checklist_guideline_export', 'Export results for compliance documentation')}</li>
                    </ul>
                </div>
            </div>

            <div style={{ display: 'grid', gap: '16px' }}>
                {checklistCategories.map(category => {
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

export default ChecklistPage;
