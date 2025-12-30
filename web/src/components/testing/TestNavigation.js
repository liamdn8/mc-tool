import React from 'react';

const TestNavigation = ({ isRunning, testStatus, testResult, overrideCount, embedded = false }) => {
    const scrollToSection = (sectionId) => {
        const element = document.getElementById(sectionId);
        if (element) {
            element.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    };

    const showRoundByRound = (isRunning && testStatus?.totalRounds > 1) || (!isRunning && testResult);
    const showOverriddenFiles = !isRunning && testResult && overrideCount > 0;

    return (
        <div style={{ 
            position: embedded ? 'static' : 'sticky', 
            top: embedded ? 'auto' : '20px',
            padding: '16px',
            backgroundColor: 'var(--bg-secondary)',
            borderRadius: '8px',
            maxHeight: embedded ? 'none' : 'calc(100vh - 100px)',
            overflowY: 'auto'
        }}>
            <div>
                <div 
                    onClick={() => scrollToSection('test-summary')}
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
                    onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'rgba(59, 130, 246, 0.1)'}
                    onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                >
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <span style={{
                            width: '20px',
                            height: '20px',
                            borderRadius: '50%',
                            backgroundColor: 'var(--primary-color)',
                            color: 'white',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: '11px',
                            fontWeight: 600,
                            flexShrink: 0
                        }}>1</span>
                        <span style={{ fontSize: '13px', color: 'var(--text-primary)', fontWeight: 400 }}>
                            Test Summary
                        </span>
                    </div>
                </div>

                {showRoundByRound && (
                    <div 
                        onClick={() => scrollToSection('round-by-round-status')}
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
                        onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'rgba(59, 130, 246, 0.1)'}
                        onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                    >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <span style={{
                                width: '20px',
                                height: '20px',
                                borderRadius: '50%',
                                backgroundColor: 'var(--primary-color)',
                                color: 'white',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                fontSize: '11px',
                                fontWeight: 600,
                                flexShrink: 0
                            }}>2</span>
                            <span style={{ fontSize: '13px', color: 'var(--text-primary)', fontWeight: 400 }}>
                                Round-by-Round Status
                            </span>
                        </div>
                    </div>
                )}

                {showOverriddenFiles && (
                    <div 
                        onClick={() => scrollToSection('overridden-files')}
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
                        onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'rgba(59, 130, 246, 0.1)'}
                        onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                    >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <span style={{
                                width: '20px',
                                height: '20px',
                                borderRadius: '50%',
                                backgroundColor: 'var(--primary-color)',
                                color: 'white',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                fontSize: '11px',
                                fontWeight: 600,
                                flexShrink: 0
                            }}>3</span>
                            <span style={{ fontSize: '13px', color: 'var(--text-primary)', fontWeight: 400 }}>
                                Overridden Files
                            </span>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default TestNavigation;
