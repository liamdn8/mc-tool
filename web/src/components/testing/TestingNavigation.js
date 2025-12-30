import React from 'react';
import { CheckCircle, XCircle, Activity } from 'lucide-react';

const TestingNavigation = ({ testResult, embedded = false }) => {
    if (!testResult) return null;

    const scrollToSection = (sectionId) => {
        const element = document.getElementById(sectionId);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    };

    const summary = testResult.summary || {};
    const hasErrors = summary.failed_uploads > 0 || (testResult.errors && testResult.errors.length > 0);

    return (
        <div style={{
            position: 'sticky',
            top: '20px',
            backgroundColor: 'white',
            border: '1px solid var(--border-color)',
            borderRadius: '8px',
            padding: '16px',
            maxHeight: 'calc(100vh - 120px)',
            overflowY: 'auto'
        }}>
            <div style={{ marginBottom: '16px' }}>
                <h3 style={{ 
                    fontSize: '14px', 
                    fontWeight: 600, 
                    margin: 0,
                    marginBottom: '12px'
                }}>
                    Test Results Navigation
                </h3>
            </div>

            {/* Quick Stats */}
            <div style={{
                padding: '12px',
                backgroundColor: 'var(--bg-secondary)',
                borderRadius: '6px',
                marginBottom: '16px'
            }}>
                <div style={{ fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '8px' }}>
                    Quick Stats
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '13px' }}>
                        <span>Total Uploads:</span>
                        <strong>{summary.total_uploads}</strong>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '13px' }}>
                        <span style={{ color: 'var(--success-color)' }}>Successful:</span>
                        <strong style={{ color: 'var(--success-color)' }}>{summary.successful_uploads}</strong>
                    </div>
                    {summary.failed_uploads > 0 && (
                        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '13px' }}>
                            <span style={{ color: 'var(--danger-color)' }}>Failed:</span>
                            <strong style={{ color: 'var(--danger-color)' }}>{summary.failed_uploads}</strong>
                        </div>
                    )}
                </div>
            </div>

            {/* Navigation Links */}
            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <button
                    onClick={() => scrollToSection('summary_stats')}
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                        padding: '8px 12px',
                        border: '1px solid var(--border-color)',
                        borderRadius: '6px',
                        backgroundColor: 'white',
                        cursor: 'pointer',
                        fontSize: '13px',
                        textAlign: 'left',
                        transition: 'background-color 0.2s'
                    }}
                    onMouseOver={(e) => e.currentTarget.style.backgroundColor = 'var(--bg-secondary)'}
                    onMouseOut={(e) => e.currentTarget.style.backgroundColor = 'white'}
                >
                    <Activity size={16} style={{ color: 'var(--primary-color)' }} />
                    <span>Summary Statistics</span>
                </button>

                {summary.overridden_objects > 0 && (
                    <button
                        onClick={() => scrollToSection('override_details')}
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            padding: '8px 12px',
                            border: '1px solid var(--border-color)',
                            borderRadius: '6px',
                            backgroundColor: 'white',
                            cursor: 'pointer',
                            fontSize: '13px',
                            textAlign: 'left',
                            transition: 'background-color 0.2s'
                        }}
                        onMouseOver={(e) => e.currentTarget.style.backgroundColor = 'var(--bg-secondary)'}
                        onMouseOut={(e) => e.currentTarget.style.backgroundColor = 'white'}
                    >
                        <CheckCircle size={16} style={{ color: 'var(--primary-color)' }} />
                        <span>Override Details</span>
                    </button>
                )}

                {hasErrors && (
                    <button
                        onClick={() => scrollToSection('error_details')}
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            padding: '8px 12px',
                            border: '1px solid var(--danger-color)',
                            borderRadius: '6px',
                            backgroundColor: 'var(--danger-light)',
                            cursor: 'pointer',
                            fontSize: '13px',
                            textAlign: 'left',
                            transition: 'background-color 0.2s'
                        }}
                        onMouseOver={(e) => e.currentTarget.style.backgroundColor = 'var(--danger-light)'}
                        onMouseOut={(e) => e.currentTarget.style.backgroundColor = 'var(--danger-light)'}
                    >
                        <XCircle size={16} style={{ color: 'var(--danger-color)' }} />
                        <span style={{ color: 'var(--danger-color)' }}>Errors</span>
                    </button>
                )}
            </div>
        </div>
    );
};

export default TestingNavigation;
