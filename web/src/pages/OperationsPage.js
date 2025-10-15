import React, { useState } from 'react';
import { Play, Settings, CheckCircle, Activity, GitCompare, List, AlertCircle, FileText } from 'lucide-react';
import { useI18n } from '../utils/i18n';

const OperationsPage = ({ sites, replicationInfo }) => {
    const { t } = useI18n();
    const [runningOperations, setRunningOperations] = useState(new Set());
    const [compareResults, setCompareResults] = useState(null);
    const [checklistResults, setChecklistResults] = useState(null);
    const [compareFormData, setCompareFormData] = useState({
        sourceAlias: '',
        destAlias: '',
        bucket: '',
        path: ''
    });
    const [availableBuckets, setAvailableBuckets] = useState({});
    const [pathSuggestions, setPathSuggestions] = useState([]);

    const hasReplication = replicationInfo && replicationInfo.enabled;

    // Fetch buckets when source alias changes
    const fetchBucketsForAlias = async (alias) => {
        if (!alias) {
            setAvailableBuckets(prev => ({ ...prev, [alias]: [] }));
            return;
        }

        try {
            const response = await fetch(`/api/operations/buckets?alias=${encodeURIComponent(alias)}`);
            if (response.ok) {
                const result = await response.json();
                setAvailableBuckets(prev => ({ ...prev, [alias]: result.buckets || [] }));
            } else {
                setAvailableBuckets(prev => ({ ...prev, [alias]: [] }));
            }
        } catch (error) {
            console.error('Failed to fetch buckets:', error);
            setAvailableBuckets(prev => ({ ...prev, [alias]: [] }));
        }
    };

    // Fetch path suggestions when bucket changes
    const fetchPathSuggestions = async (alias, bucket) => {
        if (!alias || !bucket) {
            setPathSuggestions([]);
            return;
        }

        try {
            const response = await fetch(`/api/operations/path-suggestions?alias=${encodeURIComponent(alias)}&bucket=${encodeURIComponent(bucket)}`);
            if (response.ok) {
                const result = await response.json();
                setPathSuggestions(result.paths || []);
            } else {
                setPathSuggestions([]);
            }
        } catch (error) {
            console.error('Failed to fetch path suggestions:', error);
            setPathSuggestions([]);
        }
    };

    // Handle source alias change
    const handleSourceAliasChange = (alias) => {
        setCompareFormData(prev => ({ 
            ...prev, 
            sourceAlias: alias,
            bucket: '',
            path: ''
        }));
        fetchBucketsForAlias(alias);
        setPathSuggestions([]);
    };

    // Handle bucket change
    const handleBucketChange = (bucket) => {
        setCompareFormData(prev => ({ 
            ...prev, 
            bucket: bucket,
            path: ''
        }));
        if (compareFormData.sourceAlias && bucket) {
            fetchPathSuggestions(compareFormData.sourceAlias, bucket);
        } else {
            setPathSuggestions([]);
        }
    };

    const operations = [
        {
            id: 'sync_policies',
            title: t('sync_bucket_policies'),
            description: t('sync_bucket_policies_desc'),
            icon: Settings,
            endpoint: '/api/operations/sync-policies'
        },
        {
            id: 'sync_lifecycle',
            title: t('sync_lifecycle'),
            description: t('sync_lifecycle_desc'),
            icon: Activity,
            endpoint: '/api/operations/sync-lifecycle'
        },
        {
            id: 'validate_consistency',
            title: t('validate_consistency'),
            description: t('validate_consistency_desc'),
            icon: CheckCircle,
            endpoint: '/api/operations/validate-consistency'
        },
        {
            id: 'health_check',
            title: t('health_check'),
            description: t('health_check_desc'),
            icon: Activity,
            endpoint: '/api/operations/health-check'
        },
        {
            id: 'compare_buckets',
            title: 'Compare Buckets/Paths',
            description: 'Compare content between two aliases to identify differences',
            icon: GitCompare,
            endpoint: '/api/operations/compare',
            isCustom: true
        },
        {
            id: 'config_checklist',
            title: 'Configuration Checklist',
            description: 'Verify environment variables, events, and lifecycle configurations',
            icon: List,
            endpoint: '/api/operations/checklist',
            isCustom: true
        }
    ];

    const executeOperation = async (operation) => {
        setRunningOperations(prev => new Set([...prev, operation.id]));
        
        try {
            let response;
            
            if (operation.id === 'compare_buckets') {
                response = await fetch(operation.endpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        sourceAlias: compareFormData.sourceAlias,
                        destAlias: compareFormData.destAlias,
                        path: compareFormData.bucket + (compareFormData.path ? '/' + compareFormData.path : '')
                    })
                });
            } else if (operation.id === 'config_checklist') {
                response = await fetch(operation.endpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                });
            } else {
                response = await fetch(operation.endpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                });
            }
            
            const result = await response.json();
            
            if (response.ok) {
                if (operation.id === 'compare_buckets') {
                    setCompareResults(result);
                } else if (operation.id === 'config_checklist') {
                    setChecklistResults(result);
                } else {
                    alert(`${operation.title} completed successfully`);
                }
            } else {
                alert(`${operation.title} failed: ${result.error || 'Unknown error'}`);
            }
        } catch (error) {
            alert(`${operation.title} failed: ${error.message}`);
        } finally {
            setRunningOperations(prev => {
                const newSet = new Set(prev);
                newSet.delete(operation.id);
                return newSet;
            });
        }
    };

    const renderCompareForm = (operation) => {
        if (operation.id !== 'compare_buckets') return null;
        
        const sourceBuckets = availableBuckets[compareFormData.sourceAlias] || [];
        
        return (
            <div style={{ marginTop: '16px', padding: '16px', border: '1px solid var(--border-color)', borderRadius: '8px' }}>
                <h5 style={{ margin: '0 0 16px 0' }}>Compare Configuration</h5>
                
                {/* Alias Selection Row */}
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', marginBottom: '16px' }}>
                    <div>
                        <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: '500' }}>
                            Source Alias:
                        </label>
                        <select 
                            value={compareFormData.sourceAlias}
                            onChange={(e) => handleSourceAliasChange(e.target.value)}
                            style={{ 
                                width: '100%', 
                                padding: '8px', 
                                border: '1px solid var(--border-color)', 
                                borderRadius: '4px' 
                            }}
                        >
                            <option value="">Select source alias...</option>
                            {sites.map(site => (
                                <option key={site.name} value={site.name}>{site.name}</option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: '500' }}>
                            Destination Alias:
                        </label>
                        <select 
                            value={compareFormData.destAlias}
                            onChange={(e) => setCompareFormData(prev => ({ ...prev, destAlias: e.target.value }))}
                            style={{ 
                                width: '100%', 
                                padding: '8px', 
                                border: '1px solid var(--border-color)', 
                                borderRadius: '4px' 
                            }}
                        >
                            <option value="">Select destination alias...</option>
                            {sites.map(site => (
                                <option key={site.name} value={site.name}>{site.name}</option>
                            ))}
                        </select>
                    </div>
                </div>

                {/* Bucket and Path Selection Row */}
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                    <div>
                        <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: '500' }}>
                            Bucket:
                        </label>
                        <select 
                            value={compareFormData.bucket}
                            onChange={(e) => handleBucketChange(e.target.value)}
                            disabled={!compareFormData.sourceAlias}
                            style={{ 
                                width: '100%', 
                                padding: '8px', 
                                border: '1px solid var(--border-color)', 
                                borderRadius: '4px',
                                backgroundColor: !compareFormData.sourceAlias ? '#f5f5f5' : 'white'
                            }}
                        >
                            <option value="">
                                {!compareFormData.sourceAlias ? 'Select source alias first' : 'Select bucket...'}
                            </option>
                            {sourceBuckets.map(bucket => (
                                <option key={bucket} value={bucket}>{bucket}</option>
                            ))}
                        </select>
                        {compareFormData.sourceAlias && sourceBuckets.length === 0 && (
                            <div style={{ fontSize: '12px', color: '#6c757d', marginTop: '4px' }}>
                                No buckets found in source alias
                            </div>
                        )}
                    </div>
                    <div>
                        <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: '500' }}>
                            Path (Optional):
                        </label>
                        <div style={{ position: 'relative' }}>
                            <input 
                                type="text"
                                value={compareFormData.path}
                                onChange={(e) => setCompareFormData(prev => ({ ...prev, path: e.target.value }))}
                                placeholder={compareFormData.bucket ? "Select from suggestions or type path..." : "Select bucket first"}
                                disabled={!compareFormData.bucket}
                                style={{ 
                                    width: '100%', 
                                    padding: '8px', 
                                    border: '1px solid var(--border-color)', 
                                    borderRadius: '4px',
                                    backgroundColor: !compareFormData.bucket ? '#f5f5f5' : 'white'
                                }}
                                list="path-suggestions"
                            />
                            <datalist id="path-suggestions">
                                {pathSuggestions.map(path => (
                                    <option key={path} value={path} />
                                ))}
                            </datalist>
                        </div>
                        {compareFormData.bucket && pathSuggestions.length > 0 && (
                            <div style={{ fontSize: '12px', color: '#6c757d', marginTop: '4px' }}>
                                {pathSuggestions.length} path suggestion(s) available
                            </div>
                        )}
                    </div>
                </div>
            </div>
        );
    };

    const renderCompareResults = () => {
        if (!compareResults) return null;

        const sourceAlias = compareResults.sourceAlias;
        const destAlias = compareResults.destAlias;
        const path = compareResults.path;
        const summary = compareResults.summary || {};
        
        // Display comparison info like command line
        const compareTitle = `${sourceAlias}${path ? '/' + path : ''} ↔ ${destAlias}${path ? '/' + path : ''}`;

        return (
            <div className="card" style={{ marginTop: '20px' }}>
                <div className="card-header">
                    <h3 className="card-title">
                        <GitCompare size={20} style={{ marginRight: '8px' }} />
                        Comparison Results
                    </h3>
                    <p style={{ margin: '8px 0 0 0', fontSize: '14px', color: '#6c757d' }}>
                        {compareTitle}
                    </p>
                </div>
                
                <div style={{ padding: '0 20px 20px' }}>
                    {/* Results Section - Similar to command line output */}
                    <div style={{ 
                        fontFamily: 'monospace', 
                        fontSize: '14px', 
                        backgroundColor: '#f8f9fa', 
                        padding: '16px', 
                        borderRadius: '4px',
                        border: '1px solid #e9ecef'
                    }}>
                        <div style={{ fontWeight: 'bold', marginBottom: '12px', color: '#495057' }}>
                            Comparison Results:
                        </div>
                        <div style={{ borderBottom: '1px solid #dee2e6', marginBottom: '12px' }}></div>
                        
                        {/* Files only in source (missing in target) */}
                        {compareResults.onlyInSource && compareResults.onlyInSource.length > 0 && (
                            <div style={{ marginBottom: '8px' }}>
                                {compareResults.onlyInSource.map((item, index) => (
                                    <div key={index} style={{ color: '#28a745', marginBottom: '2px' }}>
                                        + {item} - Missing in target
                                    </div>
                                ))}
                            </div>
                        )}
                        
                        {/* Files only in destination (missing in source) */}
                        {compareResults.onlyInDest && compareResults.onlyInDest.length > 0 && (
                            <div style={{ marginBottom: '8px' }}>
                                {compareResults.onlyInDest.map((item, index) => (
                                    <div key={index} style={{ color: '#dc3545', marginBottom: '2px' }}>
                                        - {item} - Missing in source
                                    </div>
                                ))}
                            </div>
                        )}
                        
                        {/* Files with different content */}
                        {compareResults.different && compareResults.different.length > 0 && (
                            <div style={{ marginBottom: '8px' }}>
                                {compareResults.different.map((diff, index) => (
                                    <div key={index} style={{ color: '#ffc107', marginBottom: '2px' }}>
                                        ⚠ {diff.path} - {diff.description || 'Content differs'}
                                    </div>
                                ))}
                            </div>
                        )}
                        
                        {/* Show message if no differences */}
                        {(!compareResults.onlyInSource || compareResults.onlyInSource.length === 0) &&
                         (!compareResults.onlyInDest || compareResults.onlyInDest.length === 0) &&
                         (!compareResults.different || compareResults.different.length === 0) && (
                            <div style={{ color: '#6c757d', fontStyle: 'italic' }}>
                                No differences found. Content is identical.
                            </div>
                        )}
                    </div>
                    
                    {/* Summary Section - Like command line */}
                    <div style={{ 
                        marginTop: '16px', 
                        fontFamily: 'monospace', 
                        fontSize: '14px',
                        backgroundColor: '#f8f9fa',
                        padding: '16px',
                        borderRadius: '4px',
                        border: '1px solid #e9ecef'
                    }}>
                        <div style={{ fontWeight: 'bold', marginBottom: '8px', color: '#495057' }}>
                            Summary:
                        </div>
                        <div style={{ color: '#6c757d' }}>
                            <div>  Identical: {summary.identical || 0}</div>
                            <div>  Different: {summary.different || 0}</div>
                            <div>  Missing in source: {summary.missingInSource || 0}</div>
                            <div>  Missing in target: {summary.missingInTarget || 0}</div>
                            <div>  Total compared: {summary.total || 0}</div>
                        </div>
                    </div>
                    
                    {/* Additional info */}
                    <div style={{ marginTop: '12px', fontSize: '12px', color: '#6c757d' }}>
                        <strong>Legend:</strong> 
                        <span style={{ color: '#28a745', marginLeft: '8px' }}>+ Missing in target</span>
                        <span style={{ color: '#dc3545', marginLeft: '8px' }}>- Missing in source</span>
                        <span style={{ color: '#ffc107', marginLeft: '8px' }}>⚠ Different content</span>
                    </div>
                </div>
            </div>
        );
    };

    const renderChecklistResults = () => {
        if (!checklistResults) return null;

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

        const groups = groupChecklist(checklistResults.items || []);

        return (
            <div className="card" style={{ marginTop: '20px' }}>
                <div className="card-header">
                    <h3 className="card-title">
                        <List size={20} style={{ marginRight: '8px' }} />
                        Configuration Checklist Results
                    </h3>
                </div>
                
                <div style={{ display: 'grid', gap: '16px' }}>
                    {Object.entries(groups).map(([groupName, items]) => {
                        if (items.length === 0) return null;
                        
                        const passed = items.filter(item => item.status === 'pass').length;
                        const failed = items.filter(item => item.status === 'fail').length;
                        const warnings = items.filter(item => item.status === 'warning').length;
                        
                        return (
                            <div key={groupName} className="card" style={{ margin: 0 }}>
                                <div className="card-header" style={{ padding: '12px 16px' }}>
                                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                        <h4 style={{ margin: 0 }}>{groupName}</h4>
                                        <div style={{ display: 'flex', gap: '8px' }}>
                                            {passed > 0 && <span className="badge badge-success">{passed} OK</span>}
                                            {warnings > 0 && <span className="badge badge-warning">{warnings} Warn</span>}
                                            {failed > 0 && <span className="badge badge-danger">{failed} Fail</span>}
                                        </div>
                                    </div>
                                </div>
                                <div style={{ padding: '0 16px 16px' }}>
                                    {items.map((item, index) => (
                                        <div key={index} style={{ 
                                            display: 'flex', 
                                            alignItems: 'flex-start', 
                                            gap: '12px', 
                                            marginBottom: '8px',
                                            padding: '8px',
                                            backgroundColor: item.status === 'pass' ? '#d4edda' : 
                                                           item.status === 'warning' ? '#fff3cd' : '#f8d7da',
                                            borderRadius: '4px',
                                            border: `1px solid ${item.status === 'pass' ? '#c3e6cb' : 
                                                                 item.status === 'warning' ? '#ffeaa7' : '#f5c6cb'}`
                                        }}>
                                            <CheckCircle 
                                                size={16} 
                                                style={{ 
                                                    marginTop: '2px',
                                                    color: item.status === 'pass' ? '#28a745' : 
                                                           item.status === 'warning' ? '#ffc107' : '#dc3545'
                                                }} 
                                            />
                                            <div style={{ flex: 1 }}>
                                                <div style={{ fontWeight: '500', marginBottom: '2px' }}>
                                                    {item.alias || 'Global'}: {item.name}
                                                </div>
                                                <div style={{ fontSize: '14px', color: '#6c757d' }}>
                                                    {item.description || item.message}
                                                </div>
                                                {item.details && (
                                                    <div style={{ fontSize: '12px', marginTop: '4px', fontFamily: 'monospace', background: '#f8f9fa', padding: '4px', borderRadius: '2px' }}>
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
                    
                    {/* Summary */}
                    <div className="stats-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
                        <div className="stat-card">
                            <div className="stat-value">{checklistResults.summary?.total || 0}</div>
                            <div className="stat-label">Total Checks</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-value" style={{ color: '#28a745' }}>
                                {checklistResults.summary?.passed || 0}
                            </div>
                            <div className="stat-label">Passed</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-value" style={{ color: '#ffc107' }}>
                                {checklistResults.summary?.warnings || 0}
                            </div>
                            <div className="stat-label">Warnings</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-value" style={{ color: '#dc3545' }}>
                                {checklistResults.summary?.failed || 0}
                            </div>
                            <div className="stat-label">Failed</div>
                        </div>
                    </div>
                </div>
            </div>
        );
    };

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">{t('automated_operations')}</h2>
            </div>

            <div>
                <div className="stats-grid">
                    <div className="stat-card">
                        <div className="stat-value">{sites.filter(s => s.replicationEnabled || true).length}</div>
                        <div className="stat-label">Available Sites</div>
                        <div className="stat-summary">Sites available for operations</div>
                    </div>

                    <div className="stat-card">
                        <div className="stat-value">{operations.length}</div>
                        <div className="stat-label">Available Operations</div>
                        <div className="stat-summary">Management and analysis tasks</div>
                    </div>

                    <div className="stat-card">
                        <div className="stat-value">{runningOperations.size}</div>
                        <div className="stat-label">Running Operations</div>
                        <div className="stat-summary">Currently executing</div>
                    </div>

                    <div className="stat-card">
                        <div className="stat-value">
                            <span className={`badge ${sites.every(s => s.healthy) ? 'badge-success' : 'badge-warning'}`}>
                                {sites.every(s => s.healthy) ? 'Ready' : 'Issues'}
                            </span>
                        </div>
                        <div className="stat-label">{t('operation_status')}</div>
                        <div className="stat-summary">System readiness</div>
                    </div>
                </div>

                {!hasReplication && (
                    <div className="card" style={{ marginBottom: '20px', backgroundColor: '#fff3cd', border: '1px solid #ffeaa7' }}>
                        <div style={{ padding: '16px' }}>
                            <h4 style={{ margin: '0 0 8px 0', color: '#856404' }}>⚠️ Site Replication Not Configured</h4>
                            <p style={{ margin: 0, color: '#856404' }}>
                                Some operations (sync policies, sync lifecycle) require site replication to be configured. 
                                However, compare and checklist operations can work with individual aliases.
                            </p>
                        </div>
                    </div>
                )}

                <div className="card">
                    <div className="card-header">
                        <h3 className="card-title">Available Operations</h3>
                    </div>

                    <div style={{ display: 'grid', gap: '16px' }}>
                        {operations.map(operation => {
                            const Icon = operation.icon;
                            const isRunning = runningOperations.has(operation.id);
                            const isCompareOperation = operation.id === 'compare_buckets';
                            const isChecklistOperation = operation.id === 'config_checklist';
                            const isReplicationOperation = ['sync_policies', 'sync_lifecycle', 'validate_consistency'].includes(operation.id);
                            const isDisabled = isRunning || 
                                (isCompareOperation && (!compareFormData.sourceAlias || !compareFormData.destAlias || !compareFormData.bucket)) ||
                                (isReplicationOperation && !hasReplication);
                            
                            return (
                                <div key={operation.id} className="card" style={{ 
                                    margin: 0, 
                                    padding: '20px',
                                    opacity: (isReplicationOperation && !hasReplication) ? 0.6 : 1
                                }}>
                                    <div style={{ display: 'flex', alignItems: 'flex-start', gap: '16px' }}>
                                        <Icon size={24} style={{ color: 'var(--primary-color)', flexShrink: 0, marginTop: '4px' }} />
                                        <div style={{ flex: 1 }}>
                                            <h4 style={{ margin: '0 0 8px 0', fontSize: '16px' }}>
                                                {operation.title}
                                                {isReplicationOperation && !hasReplication && (
                                                    <span style={{ fontSize: '12px', marginLeft: '8px', color: '#856404' }}>
                                                        (Requires Site Replication)
                                                    </span>
                                                )}
                                            </h4>
                                            <p style={{ margin: '0 0 16px 0', color: 'var(--text-secondary)' }}>
                                                {operation.description}
                                            </p>
                                            
                                            {renderCompareForm(operation)}
                                            
                                            <button 
                                                className="btn btn-primary"
                                                onClick={() => executeOperation(operation)}
                                                disabled={isDisabled}
                                                style={{ marginTop: isCompareOperation ? '16px' : '0' }}
                                            >
                                                <Play size={16} />
                                                {isRunning ? 'Running...' : t('execute')}
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                </div>

                {renderCompareResults()}
                {renderChecklistResults()}

                <div className="card">
                    <div className="card-header">
                        <h3 className="card-title">Operation Guidelines</h3>
                    </div>
                    <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        <li>Ensure all sites are healthy before running operations</li>
                        <li>Operations may take several minutes to complete</li>
                        <li>Check the logs for detailed operation results</li>
                        <li>Some operations may temporarily affect performance</li>
                        <li>Compare and checklist operations work without site replication</li>
                    </ul>
                </div>
            </div>
        </div>
    );
};

export default OperationsPage;