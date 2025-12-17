import React from 'react';
import { FileText, AlertTriangle, CheckCircle } from 'lucide-react';

const CompareNavigation = ({ compareResults, compareFormData, embedded = false }) => {
    if (!compareResults) return null;

    const summary = compareResults.summary || {};
    const onlyInSource = compareResults.onlyInSource || [];
    const onlyInDest = compareResults.onlyInDest || [];
    const different = compareResults.different || [];

    const sections = [];

    // Overview section
    // sections.push({
    //     id: 'overview',
    //     label: 'Overview',
    //     icon: FileText,
    //     count: null
    // });

    // Add sections for each category if they have items
    if (onlyInSource.length > 0) {
        sections.push({
            id: 'only-source',
            label: 'Only in Source',
            icon: AlertTriangle,
            count: onlyInSource.length,
            color: 'var(--primary-color)'
        });
    }

    if (onlyInDest.length > 0) {
        sections.push({
            id: 'only-dest',
            label: 'Only in Destination',
            icon: AlertTriangle,
            count: onlyInDest.length,
            color: 'var(--danger-color)'
        });
    }

    if (different.length > 0) {
        sections.push({
            id: 'different',
            label: 'Different Content',
            icon: AlertTriangle,
            count: different.length,
            color: 'var(--warning-color)'
        });
    }

    // Perfect match section
    if (onlyInSource.length === 0 && onlyInDest.length === 0 && different.length === 0) {
        sections.push({
            id: 'perfect-match',
            label: 'Perfect Match',
            icon: CheckCircle,
            count: null,
            color: 'var(--success-color)'
        });
    }

    const scrollToSection = (sectionId) => {
        const element = document.getElementById(sectionId);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    };

    const containerStyle = embedded ? {
        padding: '16px',
        backgroundColor: 'var(--bg-secondary)',
        borderRadius: '8px'
    } : {
        position: 'sticky',
        top: '20px',
        padding: '16px',
        backgroundColor: 'var(--bg-secondary)',
        borderRadius: '8px',
        maxHeight: 'calc(100vh - 100px)',
        overflowY: 'auto'
    };

    return (
        <div style={containerStyle}>
            {/* <h3 style={{ 
                fontSize: '14px', 
                fontWeight: '600', 
                marginBottom: '16px',
                color: 'var(--text-primary)'
            }}>
                Contents
            </h3> */}
            
            {/* <div style={{ marginBottom: '16px', paddingBottom: '16px', borderBottom: '1px solid var(--border-color)' }}>
                <div style={{ fontSize: '12px', color: 'var(--text-muted)', marginBottom: '4px' }}>
                    Comparing
                </div>
                <div style={{ fontSize: '13px', fontWeight: '500', color: 'var(--text-primary)' }}>
                    {compareFormData.sourceAlias} → {compareFormData.destAlias}
                </div>
                {compareFormData.bucket && (
                    <div style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '2px' }}>
                        {compareFormData.bucket}{compareFormData.path && `/${compareFormData.path}`}
                    </div>
                )}
            </div> */}
            
            <div>
                {sections.map(section => {
                    const Icon = section.icon;
                    return (
                        <div
                            key={section.id}
                            onClick={() => scrollToSection(section.id)}
                            style={{
                                padding: '10px 12px',
                                marginBottom: '4px',
                                borderRadius: '6px',
                                cursor: 'pointer',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'space-between',
                                transition: 'background-color 0.2s',
                                backgroundColor: 'transparent'
                            }}
                            onMouseEnter={(e) => {
                                e.currentTarget.style.backgroundColor = 'var(--hover-bg)';
                            }}
                            onMouseLeave={(e) => {
                                e.currentTarget.style.backgroundColor = 'transparent';
                            }}
                        >
                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                <Icon 
                                    size={16} 
                                    style={{ color: section.color || 'var(--text-muted)' }} 
                                />
                                <span style={{ 
                                    fontSize: '13px',
                                    color: 'var(--text-primary)'
                                }}>
                                    {section.label}
                                </span>
                            </div>
                            {section.count !== null && (
                                <span style={{
                                    fontSize: '12px',
                                    fontWeight: '500',
                                    color: section.color || 'var(--text-muted)',
                                    backgroundColor: 'var(--bg-primary)',
                                    padding: '2px 8px',
                                    borderRadius: '12px'
                                }}>
                                    {section.count}
                                </span>
                            )}
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export default CompareNavigation;
