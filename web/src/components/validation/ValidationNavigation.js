import React, { useState, useEffect } from 'react';
import { CheckCircle, XCircle, AlertCircle, List } from 'lucide-react';

const ValidationNavigation = ({ validationResults, checkLifecycle, checkEvents, embedded = false }) => {
    const [activeSection, setActiveSection] = useState('');

    useEffect(() => {
        const handleScroll = () => {
            const sections = ['bucket_existence', 'lifecycle_table', 'events_table'];
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
    }, []);

    if (!validationResults) return null;

    const scrollToSection = (sectionId) => {
        const element = document.getElementById(sectionId);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    };

    const calculateSectionStatus = (table, buckets, aliases) => {
        if (!table || table.length === 0) return { icon: AlertCircle, color: '#6b7280' };
        
        let hasInvalid = false;
        let allValid = true;

        table.forEach(row => {
            let rowHasMatch = false;
            let rowHasMismatch = false;
            let rowHasNotExist = false;

            aliases.forEach(alias => {
                const cell = row[alias];
                if (!cell || cell.status === 'not_exist') {
                    rowHasNotExist = true;
                } else if (cell.status === 'match') {
                    rowHasMatch = true;
                } else if (cell.status === 'mismatch') {
                    rowHasMismatch = true;
                }
            });

            if (rowHasNotExist || rowHasMismatch) {
                hasInvalid = true;
                allValid = false;
            } else if (!rowHasMatch) {
                allValid = false;
            }
        });

        if (allValid) {
            return { icon: CheckCircle, color: 'var(--success-color)' };
        } else if (hasInvalid) {
            return { icon: XCircle, color: 'var(--danger-color)' };
        } else {
            return { icon: AlertCircle, color: 'var(--warning-color)' };
        }
    };

    const bucketExistenceStatus = () => {
        if (!validationResults.bucket_existence) return { icon: AlertCircle, color: '#6b7280' };
        
        const buckets = validationResults.buckets || [];
        const aliases = validationResults.aliases || [];
        let allExist = true;

        buckets.forEach(bucket => {
            const bucketData = validationResults.bucket_existence[bucket];
            if (bucketData) {
                aliases.forEach(alias => {
                    if (!bucketData[alias]) {
                        allExist = false;
                    }
                });
            }
        });

        return allExist 
            ? { icon: CheckCircle, color: 'var(--success-color)' }
            : { icon: XCircle, color: 'var(--danger-color)' };
    };

    const navItems = [
        {
            id: 'bucket_existence',
            label: 'Bucket Existence',
            status: bucketExistenceStatus(),
            show: true
        },
        {
            id: 'lifecycle_table',
            label: 'Lifecycle Config',
            status: calculateSectionStatus(
                validationResults.lifecycle_table,
                validationResults.buckets,
                validationResults.aliases
            ),
            show: checkLifecycle && validationResults.lifecycle_table
        },
        {
            id: 'events_table',
            label: 'Event Notifications',
            status: calculateSectionStatus(
                validationResults.events_table,
                validationResults.buckets,
                validationResults.aliases
            ),
            show: checkEvents && validationResults.events_table
        }
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

    // Floating style (legacy, kept for compatibility)
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

export default ValidationNavigation;
