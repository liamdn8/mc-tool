import React, { useState } from 'react';
import { List, Play, CheckCircle } from 'lucide-react';

const ChecklistOperations = () => {
    const [checklistResults, setChecklistResults] = useState(null);
    const [isRunning, setIsRunning] = useState(false);

    const executeChecklist = async () => {
        setIsRunning(true);
        try {
            const response = await fetch('/api/operations/checklist', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' }
            });
            
            const result = await response.json();
            if (response.ok) {
                setChecklistResults(result);
            } else {
                alert(`Configuration checklist failed: ${result.error || 'Unknown error'}`);
            }
        } catch (error) {
            alert(`Configuration checklist failed: ${error.message}`);
        } finally {
            setIsRunning(false);
        }
    };

    const groupChecklist = (items) => {
        const groups = {
            'Environment Variables': [],
            'Event Configuration': [],
            'Bucket Events': [],
            'Object Lifecycle': [],
            'Other': []
        };

        items.forEach(item => {
            if (item.category === 'env' || item.type?.includes('environment')) {
                groups['Environment Variables'].push(item);
            } else if (item.category === 'event' || item.type?.includes('event')) {
                groups['Event Configuration'].push(item);
            } else if (item.category === 'bucket_event' || item.type?.includes('bucket')) {
                groups['Bucket Events'].push(item);
            } else if (item.category === 'lifecycle' || item.type?.includes('lifecycle')) {
                groups['Object Lifecycle'].push(item);
            } else {
                groups['Other'].push(item);
            }
        });

        return groups;
    };

    const renderChecklistResults = () => {
        if (!checklistResults) return null;

        const groups = groupChecklist(checklistResults.items || []);

        return (
            <div style={{ marginTop: '20px' }}>
                <h5 style={{ margin: '0 0 16px 0', color: '#495057' }}>
                    📋 Checklist Results
                </h5>
                
                {/* Summary Section */}
                <div style={{ 
                    marginBottom: '24px', 
                    padding: '16px',
                    backgroundColor: 'white',
                    borderRadius: '6px',
                    border: '1px solid #e9ecef'
                }}>
                    <h6 style={{ margin: '0 0 12px 0', fontWeight: '600' }}>Summary</h6>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))', gap: '12px' }}>
                        <div style={{ textAlign: 'center', padding: '12px', backgroundColor: '#d4edda', borderRadius: '6px', border: '1px solid #c3e6cb' }}>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#28a745' }}>
                                {checklistResults.summary?.passed || 0}
                            </div>
                            <div style={{ fontSize: '12px', color: '#155724' }}>Passed</div>
                        </div>
                        <div style={{ textAlign: 'center', padding: '12px', backgroundColor: '#fff3cd', borderRadius: '6px', border: '1px solid #ffeaa7' }}>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#ffc107' }}>
                                {checklistResults.summary?.warnings || 0}
                            </div>
                            <div style={{ fontSize: '12px', color: '#856404' }}>Warnings</div>
                        </div>
                        <div style={{ textAlign: 'center', padding: '12px', backgroundColor: '#f8d7da', borderRadius: '6px', border: '1px solid #f5c6cb' }}>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#dc3545' }}>
                                {checklistResults.summary?.failed || 0}
                            </div>
                            <div style={{ fontSize: '12px', color: '#721c24' }}>Failed</div>
                        </div>
                        <div style={{ textAlign: 'center', padding: '12px', backgroundColor: '#f8f9fa', borderRadius: '6px', border: '1px solid #e9ecef' }}>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#495057' }}>
                                {checklistResults.summary?.total || 0}
                            </div>
                            <div style={{ fontSize: '12px', color: '#6c757d' }}>Total Checks</div>
                        </div>
                    </div>
                </div>

                {/* Detailed Groups */}
                <div style={{ display: 'grid', gap: '16px' }}>
                    {Object.entries(groups).map(([groupName, items]) => {
                        if (items.length === 0) return null;
                        
                        const passed = items.filter(item => item.status === 'pass').length;
                        const failed = items.filter(item => item.status === 'fail').length;
                        const warnings = items.filter(item => item.status === 'warning').length;
                        
                        return (
                            <div key={groupName} style={{ 
                                border: '1px solid #e9ecef',
                                borderRadius: '8px',
                                backgroundColor: 'white',
                                overflow: 'hidden'
                            }}>
                                <div style={{ 
                                    padding: '16px',
                                    backgroundColor: '#f8f9fa',
                                    borderBottom: '1px solid #e9ecef',
                                    display: 'flex',
                                    justifyContent: 'space-between',
                                    alignItems: 'center'
                                }}>
                                    <h4 style={{ margin: 0, fontSize: '16px', fontWeight: '600' }}>{groupName}</h4>
                                    <div style={{ display: 'flex', gap: '8px' }}>
                                        {passed > 0 && (
                                            <span style={{ 
                                                padding: '4px 8px',
                                                backgroundColor: '#d4edda',
                                                color: '#155724',
                                                borderRadius: '12px',
                                                fontSize: '12px',
                                                fontWeight: '500'
                                            }}>
                                                {passed} OK
                                            </span>
                                        )}
                                        {warnings > 0 && (
                                            <span style={{ 
                                                padding: '4px 8px',
                                                backgroundColor: '#fff3cd',
                                                color: '#856404',
                                                borderRadius: '12px',
                                                fontSize: '12px',
                                                fontWeight: '500'
                                            }}>
                                                {warnings} Warn
                                            </span>
                                        )}
                                        {failed > 0 && (
                                            <span style={{ 
                                                padding: '4px 8px',
                                                backgroundColor: '#f8d7da',
                                                color: '#721c24',
                                                borderRadius: '12px',
                                                fontSize: '12px',
                                                fontWeight: '500'
                                            }}>
                                                {failed} Fail
                                            </span>
                                        )}
                                    </div>
                                </div>
                                
                                <div style={{ padding: '16px' }}>
                                    {items.map((item, index) => (
                                        <div key={index} style={{ 
                                            display: 'flex', 
                                            alignItems: 'flex-start', 
                                            gap: '12px', 
                                            marginBottom: index < items.length - 1 ? '12px' : '0',
                                            padding: '12px',
                                            backgroundColor: item.status === 'pass' ? '#d4edda' : 
                                                           item.status === 'warning' ? '#fff3cd' : '#f8d7da',
                                            borderRadius: '6px',
                                            border: `1px solid ${item.status === 'pass' ? '#c3e6cb' : 
                                                                 item.status === 'warning' ? '#ffeaa7' : '#f5c6cb'}`
                                        }}>
                                            <CheckCircle 
                                                size={20} 
                                                style={{ 
                                                    marginTop: '2px',
                                                    color: item.status === 'pass' ? '#28a745' : 
                                                           item.status === 'warning' ? '#ffc107' : '#dc3545',
                                                    flexShrink: 0
                                                }} 
                                            />
                                            <div style={{ flex: 1 }}>
                                                <div style={{ 
                                                    fontWeight: '500', 
                                                    marginBottom: '4px',
                                                    color: item.status === 'pass' ? '#155724' : 
                                                           item.status === 'warning' ? '#856404' : '#721c24'
                                                }}>
                                                    {item.alias ? `${item.alias}: ` : 'Global: '}{item.name}
                                                </div>
                                                <div style={{ 
                                                    fontSize: '14px', 
                                                    color: item.status === 'pass' ? '#155724' : 
                                                           item.status === 'warning' ? '#856404' : '#721c24',
                                                    marginBottom: item.details ? '8px' : '0'
                                                }}>
                                                    {item.description || item.message}
                                                </div>
                                                {item.details && (
                                                    <div style={{ 
                                                        fontSize: '12px', 
                                                        fontFamily: 'monospace', 
                                                        backgroundColor: 'rgba(0,0,0,0.05)', 
                                                        padding: '8px', 
                                                        borderRadius: '4px',
                                                        wordBreak: 'break-all'
                                                    }}>
                                                        {item.details}
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        );
                    })}
                </div>
            </div>
        );
    };

    return (
        <div>
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">
                        <List size={20} style={{ marginRight: '8px' }} />
                        Configuration Checklist
                    </h3>
                    <p style={{ margin: '8px 0 0 0', fontSize: '14px', color: '#6c757d' }}>
                        Verify environment variables, events, and lifecycle configurations across all sites
                    </p>
                </div>

                <div style={{ padding: '20px' }}>
                    <button 
                        className="btn btn-primary"
                        onClick={executeChecklist}
                        disabled={isRunning}
                        style={{ 
                            width: '100%',
                            padding: '12px',
                            fontSize: '16px',
                            marginBottom: '20px'
                        }}
                    >
                        <Play size={16} />
                        {isRunning ? 'Running Configuration Check...' : 'Run Configuration Checklist'}
                    </button>

                    {renderChecklistResults()}
                </div>
            </div>
        </div>
    );
};

export default ChecklistOperations;