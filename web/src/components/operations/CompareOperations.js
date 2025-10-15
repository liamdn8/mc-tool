import React, { useState } from 'react';
import { GitCompare, Play } from 'lucide-react';

const CompareOperations = ({ sites }) => {
    const [compareResults, setCompareResults] = useState(null);
    const [compareFormData, setCompareFormData] = useState({
        sourceAlias: '',
        destAlias: '',
        bucket: '',
        path: ''
    });
    const [availableBuckets, setAvailableBuckets] = useState({});
    const [pathSuggestions, setPathSuggestions] = useState([]);
    const [isRunning, setIsRunning] = useState(false);

    // Pagination states
    const [pagination, setPagination] = useState({
        onlyInSource: { page: 1, pageSize: 10 },
        onlyInDest: { page: 1, pageSize: 10 },
        different: { page: 1, pageSize: 10 }
    });

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

    const executeCompare = async () => {
        setIsRunning(true);
        try {
            const response = await fetch('/api/operations/compare', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    sourceAlias: compareFormData.sourceAlias,
                    destAlias: compareFormData.destAlias,
                    path: compareFormData.bucket + (compareFormData.path ? '/' + compareFormData.path : '')
                })
            });
            
            const result = await response.json();
            if (response.ok) {
                setCompareResults(result);
            } else {
                alert(`Compare failed: ${result.error || 'Unknown error'}`);
            }
        } catch (error) {
            alert(`Compare failed: ${error.message}`);
        } finally {
            setIsRunning(false);
        }
    };

    // Pagination helper function
    const paginateData = (data, category) => {
        const { page, pageSize } = pagination[category];
        const startIndex = (page - 1) * pageSize;
        const endIndex = startIndex + pageSize;
        return {
            data: data.slice(startIndex, endIndex),
            totalItems: data.length,
            totalPages: Math.ceil(data.length / pageSize),
            currentPage: page,
            pageSize: pageSize
        };
    };

    // Update pagination
    const updatePagination = (category, updates) => {
        setPagination(prev => ({
            ...prev,
            [category]: { ...prev[category], ...updates }
        }));
    };

    // Render pagination controls
    const renderPaginationControls = (category, totalItems, totalPages, currentPage, pageSize) => {
        if (totalItems <= 10) return null;

        return (
            <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'center', 
                marginTop: '12px',
                padding: '8px 0',
                borderTop: '1px solid #e9ecef'
            }}>
                <div style={{ fontSize: '14px', color: '#6c757d' }}>
                    Showing {Math.min((currentPage - 1) * pageSize + 1, totalItems)} to {Math.min(currentPage * pageSize, totalItems)} of {totalItems} items
                </div>
                
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <select 
                        value={pageSize} 
                        onChange={(e) => updatePagination(category, { pageSize: parseInt(e.target.value), page: 1 })}
                        style={{ 
                            padding: '4px 8px', 
                            border: '1px solid #ccc', 
                            borderRadius: '4px',
                            fontSize: '14px'
                        }}
                    >
                        <option value={10}>10</option>
                        <option value={25}>25</option>
                        <option value={50}>50</option>
                        <option value={100}>100</option>
                    </select>
                    
                    <button 
                        onClick={() => updatePagination(category, { page: Math.max(1, currentPage - 1) })}
                        disabled={currentPage === 1}
                        style={{ 
                            padding: '4px 8px', 
                            border: '1px solid #ccc', 
                            borderRadius: '4px',
                            backgroundColor: currentPage === 1 ? '#f5f5f5' : 'white',
                            cursor: currentPage === 1 ? 'not-allowed' : 'pointer'
                        }}
                    >
                        Previous
                    </button>
                    
                    <span style={{ fontSize: '14px' }}>
                        Page {currentPage} of {totalPages}
                    </span>
                    
                    <button 
                        onClick={() => updatePagination(category, { page: Math.min(totalPages, currentPage + 1) })}
                        disabled={currentPage === totalPages}
                        style={{ 
                            padding: '4px 8px', 
                            border: '1px solid #ccc', 
                            borderRadius: '4px',
                            backgroundColor: currentPage === totalPages ? '#f5f5f5' : 'white',
                            cursor: currentPage === totalPages ? 'not-allowed' : 'pointer'
                        }}
                    >
                        Next
                    </button>
                </div>
            </div>
        );
    };

    // Render table for items
    const renderTable = (items, category, title, emptyMessage) => {
        const paginated = paginateData(items, category);
        
        return (
            <div style={{ marginBottom: '24px' }}>
                <h4 style={{ 
                    margin: '0 0 12px 0', 
                    fontSize: '16px', 
                    fontWeight: '600',
                    color: '#495057',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px'
                }}>
                    {title}
                    <span style={{ 
                        fontSize: '14px', 
                        fontWeight: 'normal', 
                        color: '#6c757d',
                        backgroundColor: '#e9ecef',
                        padding: '2px 8px',
                        borderRadius: '12px'
                    }}>
                        {items.length}
                    </span>
                </h4>
                
                {items.length === 0 ? (
                    <div style={{ 
                        padding: '20px', 
                        textAlign: 'center', 
                        color: '#6c757d',
                        fontStyle: 'italic',
                        backgroundColor: '#f8f9fa',
                        borderRadius: '4px',
                        border: '1px solid #e9ecef'
                    }}>
                        {emptyMessage}
                    </div>
                ) : (
                    <div style={{ 
                        border: '1px solid #e9ecef',
                        borderRadius: '4px',
                        overflow: 'hidden'
                    }}>
                        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                            <thead>
                                <tr style={{ backgroundColor: '#f8f9fa' }}>
                                    <th style={{ 
                                        padding: '12px', 
                                        textAlign: 'left', 
                                        borderBottom: '1px solid #e9ecef',
                                        fontWeight: '600',
                                        fontSize: '14px'
                                    }}>
                                        Path
                                    </th>
                                    {category === 'different' && (
                                        <th style={{ 
                                            padding: '12px', 
                                            textAlign: 'left', 
                                            borderBottom: '1px solid #e9ecef',
                                            fontWeight: '600',
                                            fontSize: '14px'
                                        }}>
                                            Description
                                        </th>
                                    )}
                                </tr>
                            </thead>
                            <tbody>
                                {paginated.data.map((item, index) => (
                                    <tr key={index} style={{ 
                                        borderBottom: index < paginated.data.length - 1 ? '1px solid #f1f3f4' : 'none'
                                    }}>
                                        <td style={{ 
                                            padding: '12px', 
                                            fontFamily: 'monospace',
                                            fontSize: '13px',
                                            wordBreak: 'break-all'
                                        }}>
                                            {typeof item === 'string' ? item : item.path}
                                        </td>
                                        {category === 'different' && (
                                            <td style={{ 
                                                padding: '12px', 
                                                fontSize: '13px',
                                                color: '#6c757d'
                                            }}>
                                                {item.description}
                                            </td>
                                        )}
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                        
                        {renderPaginationControls(category, paginated.totalItems, paginated.totalPages, paginated.currentPage, paginated.pageSize)}
                    </div>
                )}
            </div>
        );
    };

    const renderCompareResults = () => {
        if (!compareResults) return null;

        const summary = compareResults.summary || {};
        const onlyInSource = compareResults.onlyInSource || [];
        const onlyInDest = compareResults.onlyInDest || [];
        const different = compareResults.different || [];
        
        return (
            <div style={{ marginTop: '20px', padding: '16px', border: '1px solid #e9ecef', borderRadius: '8px', backgroundColor: '#fafafa' }}>
                <h5 style={{ margin: '0 0 16px 0', color: '#495057' }}>
                    📊 Comparison Results
                </h5>
                
                {/* Summary Section First */}
                <div style={{ 
                    marginBottom: '24px', 
                    padding: '16px',
                    backgroundColor: 'white',
                    borderRadius: '6px',
                    border: '1px solid #e9ecef'
                }}>
                    <h6 style={{ margin: '0 0 12px 0', fontWeight: '600' }}>Summary</h6>
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))', gap: '12px' }}>
                        <div style={{ textAlign: 'center', padding: '8px', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#28a745' }}>{summary.identical || 0}</div>
                            <div style={{ fontSize: '12px', color: '#6c757d' }}>Identical</div>
                        </div>
                        <div style={{ textAlign: 'center', padding: '8px', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#ffc107' }}>{summary.different || 0}</div>
                            <div style={{ fontSize: '12px', color: '#6c757d' }}>Different</div>
                        </div>
                        <div style={{ textAlign: 'center', padding: '8px', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#dc3545' }}>{summary.missingInSource || 0}</div>
                            <div style={{ fontSize: '12px', color: '#6c757d' }}>Missing in Source</div>
                        </div>
                        <div style={{ textAlign: 'center', padding: '8px', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#007bff' }}>{summary.missingInTarget || 0}</div>
                            <div style={{ fontSize: '12px', color: '#6c757d' }}>Missing in Target</div>
                        </div>
                        <div style={{ textAlign: 'center', padding: '8px', backgroundColor: '#f8f9fa', borderRadius: '4px' }}>
                            <div style={{ fontSize: '24px', fontWeight: 'bold', color: '#6c757d' }}>{summary.total || 0}</div>
                            <div style={{ fontSize: '12px', color: '#6c757d' }}>Total Compared</div>
                        </div>
                    </div>
                </div>

                {/* Detailed Tables */}
                <div style={{ backgroundColor: 'white', padding: '16px', borderRadius: '6px', border: '1px solid #e9ecef' }}>
                    {renderTable(
                        onlyInSource,
                        'onlyInSource',
                        '📤 Only in Source (' + compareResults.sourceAlias + ')',
                        'No files found only in source'
                    )}
                    
                    {renderTable(
                        onlyInDest,
                        'onlyInDest',
                        '📥 Only in Destination (' + compareResults.destAlias + ')',
                        'No files found only in destination'
                    )}
                    
                    {renderTable(
                        different,
                        'different',
                        '⚠️ Different Content',
                        'No differences found'
                    )}
                    
                    {onlyInSource.length === 0 && onlyInDest.length === 0 && different.length === 0 && (
                        <div style={{ 
                            padding: '40px', 
                            textAlign: 'center', 
                            color: '#28a745',
                            backgroundColor: '#d4edda',
                            borderRadius: '6px',
                            border: '1px solid #c3e6cb'
                        }}>
                            <div style={{ fontSize: '48px', marginBottom: '16px' }}>✅</div>
                            <div style={{ fontSize: '18px', fontWeight: '600', marginBottom: '8px' }}>Perfect Match!</div>
                            <div style={{ fontSize: '14px' }}>All files are identical between the two locations.</div>
                        </div>
                    )}
                </div>
            </div>
        );
    };

    const sourceBuckets = availableBuckets[compareFormData.sourceAlias] || [];
    const isDisabled = isRunning || !compareFormData.sourceAlias || !compareFormData.destAlias || !compareFormData.bucket;

    return (
        <div>
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">
                        <GitCompare size={20} style={{ marginRight: '8px' }} />
                        Compare Buckets/Paths
                    </h3>
                    <p style={{ margin: '8px 0 0 0', fontSize: '14px', color: '#6c757d' }}>
                        Compare content between two MinIO aliases to identify differences
                    </p>
                </div>

                <div style={{ padding: '20px' }}>
                    {/* Compare Configuration Form */}
                    <div style={{ padding: '16px', border: '1px solid var(--border-color)', borderRadius: '8px', backgroundColor: '#f8f9fa' }}>
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
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', marginBottom: '16px' }}>
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
                            </div>
                            <div>
                                <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: '500' }}>
                                    Path (Optional):
                                </label>
                                <input 
                                    type="text"
                                    value={compareFormData.path}
                                    onChange={(e) => setCompareFormData(prev => ({ ...prev, path: e.target.value }))}
                                    placeholder={compareFormData.bucket ? "Enter path or leave empty for entire bucket" : "Select bucket first"}
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
                        </div>

                        <button 
                            className="btn btn-primary"
                            onClick={executeCompare}
                            disabled={isDisabled}
                            style={{ width: '100%' }}
                        >
                            <Play size={16} />
                            {isRunning ? 'Comparing...' : 'Execute Compare'}
                        </button>
                    </div>

                    {/* Render Compare Results inside the form */}
                    {renderCompareResults()}
                </div>
            </div>
        </div>
    );
};

export default CompareOperations;