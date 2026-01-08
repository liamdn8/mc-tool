import React, { useState, useEffect } from 'react';
import { CheckCircle, XCircle, AlertCircle, List } from 'lucide-react';

const InfraValidationNavigation = ({ result, embedded = false }) => {
    const [activeSection, setActiveSection] = useState('');

    useEffect(() => {
        const handleScroll = () => {
            if (!result || !result.summary) return;

            const resourceTypes = getResourceTypes();
            const sections = ['overview', ...resourceTypes.map(rt => `resource-${rt}`)];
            const scrollPosition = window.scrollY + 100;

            for (const sectionId of sections) {
                const element = document.getElementById(sectionId);
                if (element) {
                    const { offsetTop, offsetHeight } = element;
                    if (scrollPosition >= offsetTop && scrollPosition < offsetTop + offsetHeight) {
                        setActiveSection(sectionId);
                        break;
                    }
                }
            }
        };

        window.addEventListener('scroll', handleScroll);
        handleScroll();

        return () => window.removeEventListener('scroll', handleScroll);
    }, [result]);

    if (!result || !result.summary) return null;

    const scrollToSection = (sectionId) => {
        const element = document.getElementById(sectionId);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    };

    const getResourceTypes = () => {
        if (!result.summary.resource_table) return [];
        
        const types = new Set();
        result.summary.resource_table.forEach(row => {
            if (row.resource_type) {
                types.add(row.resource_type);
            }
        });
        return Array.from(types).sort();
    };

    const calculateResourceTypeStatus = (resourceType) => {
        if (!result.summary.resource_table) return { icon: AlertCircle, color: '#6b7280' };
        
        const resources = result.summary.resource_table.filter(r => r.resource_type === resourceType);
        if (resources.length === 0) return { icon: AlertCircle, color: '#6b7280' };

        const allNamespaces = [result.summary.baseline, ...(result.summary.targets || [])];
        
        let hasMatch = false;
        let hasMismatch = false;
        let hasNotFound = false;

        resources.forEach(resource => {
            allNamespaces.forEach(ns => {
                const cell = resource[ns];
                if (!cell || cell.status === 'not_found') {
                    hasNotFound = true;
                } else if (cell.status === 'match') {
                    hasMatch = true;
                } else if (cell.status === 'mismatch') {
                    hasMismatch = true;
                }
            });
        });

        if (hasMismatch || hasNotFound) {
            return { icon: XCircle, color: 'var(--danger-color)' };
        } else if (hasMatch) {
            return { icon: CheckCircle, color: 'var(--success-color)' };
        } else {
            return { icon: AlertCircle, color: 'var(--warning-color)' };
        }
    };

    const overallStatus = () => {
        const { summary } = result;
        if (summary.mismatchCount > 0 || summary.notFoundCount > 0) {
            return { icon: XCircle, color: 'var(--danger-color)' };
        } else if (summary.matchCount > 0) {
            return { icon: CheckCircle, color: 'var(--success-color)' };
        }
        return { icon: AlertCircle, color: 'var(--warning-color)' };
    };

    const resourceTypes = getResourceTypes();

    const navItems = [
        {
            id: 'overview',
            label: 'Overview',
            status: overallStatus(),
            show: true
        },
        ...resourceTypes.map(resourceType => ({
            id: `resource-${resourceType}`,
            label: resourceType,
            status: calculateResourceTypeStatus(resourceType),
            show: true
        }))
    ].filter(item => item.show);

    // Embedded style (for sidebar panel)
    if (embedded) {
        return (
            <div style={{
                position: 'sticky',
                top: '20px',
                padding: '16px',
                backgroundColor: 'var(--bg-secondary)',
                borderRadius: '8px',
                maxHeight: 'calc(100vh - 100px)',
                overflowY: 'auto'
            }}>
                {/* <div style={{
                    marginBottom: '12px',
                    paddingBottom: '8px',
                    borderBottom: '1px solid var(--border-color)',
                    fontSize: '14px',
                    fontWeight: 600,
                    color: 'var(--text-primary)'
                }}>
                    Contents
                </div> */}
                <div>
                    {navItems.map(item => {
                        const StatusIcon = item.status.icon;
                        const isActive = activeSection === item.id;
                        
                        return (
                            <div
                                key={item.id}
                                onClick={() => scrollToSection(item.id)}
                                style={{
                                    padding: '10px 12px',
                                    marginBottom: '4px',
                                    borderRadius: '6px',
                                    cursor: 'pointer',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'space-between',
                                    transition: 'background-color 0.2s',
                                    backgroundColor: isActive ? 'var(--hover-bg)' : 'transparent'
                                }}
                                onMouseEnter={(e) => {
                                    e.currentTarget.style.backgroundColor = 'var(--hover-bg)';
                                }}
                                onMouseLeave={(e) => {
                                    if (!isActive) {
                                        e.currentTarget.style.backgroundColor = 'transparent';
                                    }
                                }}
                            >
                                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                    <StatusIcon 
                                        size={16} 
                                        style={{ 
                                            color: item.status.color,
                                            flexShrink: 0
                                        }} 
                                    />
                                    <span style={{ 
                                        fontSize: '13px',
                                        color: 'var(--text-primary)',
                                        fontWeight: isActive ? '500' : '400'
                                    }}>
                                        {item.label}
                                    </span>
                                </div>
                            </div>
                        );
                    })}
                </div>
            </div>
        );
    }

    // Floating style
    return (
        <div style={{
            position: 'fixed',
            left: '20px',
            bottom: '20px',
            width: '220px',
            backgroundColor: 'white',
            border: '1px solid var(--border-color)',
            borderRadius: '8px',
            padding: '12px',
            boxShadow: '0 2px 12px rgba(0, 0, 0, 0.15)',
            maxHeight: '400px',
            overflowY: 'auto',
            zIndex: 1000
        }}>
            <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                marginBottom: '12px',
                paddingBottom: '8px',
                borderBottom: '1px solid var(--border-color)'
            }}>
                <List size={18} style={{ color: 'var(--primary-color)' }} />
                <span style={{ fontSize: '14px', fontWeight: 600, color: 'var(--text-primary)' }}>
                    Contents
                </span>
            </div>
            
            {navItems.map(item => {
                const StatusIcon = item.status.icon;
                const isActive = activeSection === item.id;
                
                return (
                    <button
                        key={item.id}
                        onClick={() => scrollToSection(item.id)}
                        style={{
                            width: '100%',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            padding: '8px 12px',
                            marginBottom: '4px',
                            border: 'none',
                            borderRadius: '6px',
                            backgroundColor: isActive ? 'var(--primary-light)' : 'transparent',
                            color: isActive ? 'var(--primary-color)' : 'var(--text-primary)',
                            cursor: 'pointer',
                            transition: 'all 0.2s',
                            textAlign: 'left',
                            fontSize: '13px',
                            fontWeight: isActive ? 600 : 400
                        }}
                        onMouseEnter={(e) => {
                            if (!isActive) {
                                e.currentTarget.style.backgroundColor = 'var(--bg-secondary)';
                            }
                        }}
                        onMouseLeave={(e) => {
                            if (!isActive) {
                                e.currentTarget.style.backgroundColor = 'transparent';
                            }
                        }}
                    >
                        <StatusIcon 
                            size={16} 
                            style={{ 
                                color: item.status.color,
                                flexShrink: 0
                            }} 
                        />
                        <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                            {item.label}
                        </span>
                    </button>
                );
            })}
        </div>
    );
};

export default InfraValidationNavigation;
