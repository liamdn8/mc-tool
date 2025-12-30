import React, { useState, useEffect } from 'react';
import { GitCompare, Play, BarChart3, CheckCircle, AlertTriangle, ChevronUp, ChevronDown, FileText, Folder, FileX } from 'lucide-react';
import ErrorAlert from '../ErrorAlert';
import { apiCall } from '../../utils/api';
import { useContentsPanel } from '../../contexts/ContentsPanelContext';
import CompareNavigation from './CompareNavigation';

const CompareOperations = ({ sites }) => {
    const { setContentsComponent } = useContentsPanel();
    const [compareResults, setCompareResults] = useState(null);
    const [error, setError] = useState(null);
    const [compareFormData, setCompareFormData] = useState({
        sourceAlias: '',
        destAlias: '',
        bucket: '',
        path: '',
        compareVersion: false,
        insecure: false
    });
    const [availableBuckets, setAvailableBuckets] = useState({});
    const [pathSuggestions, setPathSuggestions] = useState([]);
    const [isRunning, setIsRunning] = useState(false);
    const [versioningStatus, setVersioningStatus] = useState({
        bothVersioned: false,
        sourceVersioning: false,
        destVersioning: false,
        checked: false
    });

    // Pagination states
    const [pagination, setPagination] = useState({
        onlyInSource: { page: 1, pageSize: 10 },
        onlyInDest: { page: 1, pageSize: 10 },
        different: { page: 1, pageSize: 10 }
    });

    // Update contents panel when results change
    useEffect(() => {
        if (compareResults) {
            setContentsComponent(
                <CompareNavigation 
                    compareResults={compareResults}
                    compareFormData={compareFormData}
                />
            );
        } else {
            setContentsComponent(null);
        }

        // Cleanup when component unmounts
        return () => {
            setContentsComponent(null);
        };
    }, [compareResults, compareFormData, setContentsComponent]);

    // Fetch buckets when source alias changes
    const fetchBucketsForAlias = async (alias) => {
        if (!alias) {
            setAvailableBuckets(prev => ({ ...prev, [alias]: [] }));
            return;
        }

        try {
            const { response, data: result } = await apiCall(`/api/operations/buckets?alias=${encodeURIComponent(alias)}`);
            if (response.ok) {
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
            const { response, data: result } = await apiCall(`/api/operations/path-suggestions?alias=${encodeURIComponent(alias)}&bucket=${encodeURIComponent(bucket)}`);
            if (response.ok) {
                setPathSuggestions(result.paths || []);
            } else {
                setPathSuggestions([]);
            }
        } catch (error) {
            console.error('Failed to fetch path suggestions:', error);
            setPathSuggestions([]);
        }
    };

    // Check bucket versioning status
    const checkBucketVersioning = async (sourceAlias, destAlias, bucket) => {
        if (!sourceAlias || !destAlias || !bucket) {
            setVersioningStatus({
                bothVersioned: false,
                sourceVersioning: false,
                destVersioning: false,
                checked: false
            });
            return;
        }

        try {
            const { response, data: result } = await apiCall(`/api/operations/bucket-versioning?sourceAlias=${encodeURIComponent(sourceAlias)}&destAlias=${encodeURIComponent(destAlias)}&bucket=${encodeURIComponent(bucket)}`);
            if (response.ok) {
                setVersioningStatus({
                    bothVersioned: result.bothVersioned,
                    sourceVersioning: result.sourceVersioning,
                    destVersioning: result.destVersioning,
                    checked: true
                });
                
                // If versioning is not available on both sides, disable the compareVersion option
                if (!result.bothVersioned && compareFormData.compareVersion) {
                    setCompareFormData(prev => ({ ...prev, compareVersion: false }));
                }
            } else {
                setVersioningStatus({
                    bothVersioned: false,
                    sourceVersioning: false,
                    destVersioning: false,
                    checked: true
                });
            }
        } catch (error) {
            console.error('Failed to check bucket versioning:', error);
            setVersioningStatus({
                bothVersioned: false,
                sourceVersioning: false,
                destVersioning: false,
                checked: true
            });
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
        
        // Check versioning when bucket changes and both aliases are selected
        if (compareFormData.sourceAlias && compareFormData.destAlias && bucket) {
            checkBucketVersioning(compareFormData.sourceAlias, compareFormData.destAlias, bucket);
        }
    };

    const handleDestAliasChange = (alias) => {
        setCompareFormData(prev => ({ ...prev, destAlias: alias }));
        
        // Check versioning when dest alias changes and bucket is selected
        if (compareFormData.sourceAlias && alias && compareFormData.bucket) {
            checkBucketVersioning(compareFormData.sourceAlias, alias, compareFormData.bucket);
        }
    };

    const executeCompare = async () => {
        setIsRunning(true);
        setError(null);
        try {
            const { response, data: result } = await apiCall('/api/operations/compare', {
                method: 'POST',
                body: JSON.stringify({
                    sourceAlias: compareFormData.sourceAlias,
                    destAlias: compareFormData.destAlias,
                    path: compareFormData.bucket + (compareFormData.path ? '/' + compareFormData.path : ''),
                    compareVersion: compareFormData.compareVersion,
                    insecure: compareFormData.insecure
                })
            });
            if (response.ok) {
                setCompareResults(result);
            } else {
                setError(result.error || 'Unknown error');
            }
        } catch (err) {
            setError(err.message || 'Network error occurred');
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
                flexWrap: 'wrap',
                justifyContent: 'space-between', 
                alignItems: 'center',
                gap: '12px',
                padding: '12px 0 0',
                borderTop: '1px solid #e5e7eb'
            }}>
                <span style={{ fontSize: '12px', color: '#4b5563' }}>
                    {((currentPage - 1) * pageSize) + 1}-{Math.min(currentPage * pageSize, totalItems)} of {totalItems}
                </span>
                
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                    <label style={{ fontSize: '12px', color: '#4b5563', display: 'inline-flex', alignItems: 'center', gap: '6px' }}>
                        Rows per page
                        <select
                            value={pageSize}
                            onChange={(e) => updatePagination(category, { pageSize: parseInt(e.target.value), page: 1 })}
                            style={{
                                padding: '6px 10px',
                                borderRadius: '6px',
                                border: '1px solid #d1d5db',
                                fontSize: '12px',
                                backgroundColor: '#ffffff'
                            }}
                        >
                            <option value={10}>10</option>
                            <option value={20}>20</option>
                            <option value={50}>50</option>
                            <option value={100}>100</option>
                        </select>
                    </label>
                    
                    <div style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                        <button
                            onClick={() => updatePagination(category, { page: Math.max(1, currentPage - 1) })}
                            disabled={currentPage === 1}
                            style={{
                                padding: '6px 10px',
                                borderRadius: '6px',
                                border: '1px solid #d1d5db',
                                backgroundColor: currentPage === 1 ? '#f3f4f6' : '#ffffff',
                                color: '#1f2937',
                                fontSize: '12px',
                                cursor: currentPage === 1 ? 'not-allowed' : 'pointer'
                            }}
                        >
                            Prev
                        </button>
                        
                        <span style={{ fontSize: '12px', color: '#4b5563', minWidth: '60px', textAlign: 'center' }}>
                            {currentPage} / {totalPages}
                        </span>
                        
                        <button
                            onClick={() => updatePagination(category, { page: Math.min(totalPages, currentPage + 1) })}
                            disabled={currentPage === totalPages}
                            style={{
                                padding: '6px 10px',
                                borderRadius: '6px',
                                border: '1px solid #d1d5db',
                                backgroundColor: currentPage === totalPages ? '#f3f4f6' : '#ffffff',
                                color: '#1f2937',
                                fontSize: '12px',
                                cursor: currentPage === totalPages ? 'not-allowed' : 'pointer'
                            }}
                        >
                            Next
                        </button>
                    </div>
                </div>
            </div>
        );
    };

    // Render table for items
    const renderTable = (items, category, icon, title, emptyMessage) => {
        const paginated = paginateData(items, category);
        const Icon = icon;
        
        return (
            <div>
                {items.length === 0 ? (
                    <div className="empty-state text-sm">
                        {emptyMessage}
                    </div>
                ) : (
                    <>
                        {/* Pagination Controls - Top */}
                        {paginated.totalPages > 1 && (
                            <div style={{ 
                                display: 'flex', 
                                flexWrap: 'wrap',
                                justifyContent: 'space-between', 
                                alignItems: 'center',
                                gap: '12px',
                                padding: '0 0 12px',
                                borderTop: 'none'
                            }}>
                                <span style={{ fontSize: '12px', color: '#4b5563' }}>
                                    {((paginated.currentPage - 1) * paginated.pageSize) + 1}-{Math.min(paginated.currentPage * paginated.pageSize, paginated.totalItems)} of {paginated.totalItems}
                                </span>
                                
                                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                                    <label style={{ fontSize: '12px', color: '#4b5563', display: 'inline-flex', alignItems: 'center', gap: '6px' }}>
                                        Rows per page
                                        <select
                                            value={paginated.pageSize}
                                            onChange={(e) => updatePagination(category, { pageSize: parseInt(e.target.value), page: 1 })}
                                            style={{
                                                padding: '6px 10px',
                                                borderRadius: '6px',
                                                border: '1px solid #d1d5db',
                                                fontSize: '12px',
                                                backgroundColor: '#ffffff'
                                            }}
                                        >
                                            <option value={10}>10</option>
                                            <option value={20}>20</option>
                                            <option value={50}>50</option>
                                            <option value={100}>100</option>
                                        </select>
                                    </label>
                                    
                                    <div style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                                        <button
                                            onClick={() => updatePagination(category, { page: Math.max(1, paginated.currentPage - 1) })}
                                            disabled={paginated.currentPage === 1}
                                            style={{
                                                padding: '6px 10px',
                                                borderRadius: '6px',
                                                border: '1px solid #d1d5db',
                                                backgroundColor: paginated.currentPage === 1 ? '#f3f4f6' : '#ffffff',
                                                color: '#1f2937',
                                                fontSize: '12px',
                                                cursor: paginated.currentPage === 1 ? 'not-allowed' : 'pointer'
                                            }}
                                        >
                                            Prev
                                        </button>
                                        
                                        <span style={{ fontSize: '12px', color: '#4b5563', minWidth: '60px', textAlign: 'center' }}>
                                            {paginated.currentPage} / {paginated.totalPages}
                                        </span>
                                        
                                        <button
                                            onClick={() => updatePagination(category, { page: Math.min(paginated.totalPages, paginated.currentPage + 1) })}
                                            disabled={paginated.currentPage === paginated.totalPages}
                                            style={{
                                                padding: '6px 10px',
                                                borderRadius: '6px',
                                                border: '1px solid #d1d5db',
                                                backgroundColor: paginated.currentPage === paginated.totalPages ? '#f3f4f6' : '#ffffff',
                                                color: '#1f2937',
                                                fontSize: '12px',
                                                cursor: paginated.currentPage === paginated.totalPages ? 'not-allowed' : 'pointer'
                                            }}
                                        >
                                            Next
                                        </button>
                                    </div>
                                </div>
                            </div>
                        )}

                        <div style={{ border: '1px solid #d1d5db', borderRadius: '6px', overflowX: 'auto' }}>
                            <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '640px' }}>
                                <thead style={{ backgroundColor: '#f9fafb' }}>
                                    <tr>
                                        <th style={{ 
                                            padding: '12px', 
                                            textAlign: 'left', 
                                            fontSize: '12px', 
                                            fontWeight: '400',
                                            color: '#6b7280',
                                            whiteSpace: 'nowrap'
                                        }}>
                                            #
                                        </th>
                                        <th style={{ 
                                            padding: '12px', 
                                            textAlign: 'left', 
                                            fontSize: '12px', 
                                            fontWeight: '400',
                                            color: '#6b7280'
                                        }}>
                                            Path
                                        </th>
                                        {category === 'different' && (
                                            <th style={{ 
                                                padding: '12px', 
                                                textAlign: 'left', 
                                                fontSize: '12px', 
                                                fontWeight: '400',
                                                color: '#6b7280'
                                            }}>
                                                Description
                                            </th>
                                        )}
                                    </tr>
                                </thead>
                                <tbody>
                                    {paginated.data.map((item, index) => {
                                        const globalIndex = (paginated.currentPage - 1) * paginated.pageSize + index + 1;
                                        return (
                                            <tr key={index} style={{ borderTop: '1px solid #f3f4f6' }}>
                                                <td style={{ 
                                                    padding: '12px', 
                                                    textAlign: 'center', 
                                                    fontWeight: '600',
                                                    fontSize: '13px',
                                                    color: '#6b7280',
                                                    whiteSpace: 'nowrap'
                                                }}>
                                                    {globalIndex}
                                                </td>
                                                <td style={{ 
                                                    padding: '12px', 
                                                    fontSize: '12px', 
                                                    color: '#111827',
                                                    wordBreak: 'break-all',
                                                    fontFamily: 'monospace'
                                                }}>
                                                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                                                        <Icon size={14} style={{ flexShrink: 0, color: '#6b7280' }} />
                                                        <span>{typeof item === 'string' ? item : item.path}</span>
                                                    </div>
                                                </td>
                                                {category === 'different' && (
                                                    <td style={{ 
                                                        padding: '12px', 
                                                        fontSize: '12px',
                                                        color: '#4b5563'
                                                    }}>
                                                        {item.description}
                                                    </td>
                                                )}
                                            </tr>
                                        );
                                    })}
                                </tbody>
                            </table>
                        </div>

                        {/* Pagination Controls - Bottom */}
                        {paginated.totalPages > 1 && renderPaginationControls(category, paginated.totalItems, paginated.totalPages, paginated.currentPage, paginated.pageSize)}
                    </>
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
            <div className="card mt-6">
                <div className="card-header">
                    <h3 id="overview" className="card-title" style={{ 
                        paddingBottom: '12px',
                        borderBottom: '2px solid var(--border-color)'
                    }}>
                        Comparison Results
                    </h3>
                    <p className="m-0 mt-2 text-sm text-secondary">
                        Comparing {compareFormData.sourceAlias} → {compareFormData.destAlias}
                        {compareFormData.bucket && ` in bucket "${compareFormData.bucket}"`}
                        {compareFormData.path && ` at path "${compareFormData.path}"`}
                    </p>
                </div>

                <div style={{ padding: '20px' }}>
                    {/* Badges */}
                    <div style={{ 
                        display: 'flex', 
                        flexWrap: 'wrap', 
                        gap: '8px', 
                        marginBottom: '16px' 
                    }}>
                        {compareFormData.compareVersion && (
                            <span className="badge badge-success">Version comparison enabled</span>
                        )}
                        <span className="badge badge-neutral">
                            Source: {compareFormData.sourceAlias}
                        </span>
                        <span className="badge badge-neutral">
                            Destination: {compareFormData.destAlias}
                        </span>
                    </div>

                    {/* Overview Cards */}
                    <div className="stats-grid" style={{ marginBottom: '24px' }}>
                        <div className="stat-card">
                            <div className="stat-value" style={{ color: 'var(--success-color)' }}>
                                {summary.identical || 0}
                            </div>
                            <div className="stat-label">Identical</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-value" style={{ color: 'var(--warning-color)' }}>
                                {summary.different || 0}
                            </div>
                            <div className="stat-label">Different</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-value" style={{ color: 'var(--danger-color)' }}>
                                {summary.missingInSource || 0}
                            </div>
                            <div className="stat-label">Only in Destination</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-value" style={{ color: 'var(--primary-color)' }}>
                                {summary.missingInTarget || 0}
                            </div>
                            <div className="stat-label">Only in Source</div>
                        </div>
                        <div className="stat-card">
                            <div className="stat-value" style={{ color: 'var(--text-muted)' }}>
                                {summary.total || 0}
                            </div>
                            <div className="stat-label">Total Compared</div>
                        </div>
                    </div>

                    {/* Perfect Match Message */}
                    {onlyInSource.length === 0 && onlyInDest.length === 0 && different.length === 0 && (
                        <div id="perfect-match" className="p-6 text-center text-success bg-success-light rounded-lg border mt-6" style={{ borderColor: '#bbf7d0' }}>
                            <div className="flex justify-center mb-4" style={{ fontSize: '48px' }}>
                                <CheckCircle size={48} color="var(--success-color)" />
                            </div>
                            <div className="text-xl font-semibold mb-2">
                                Perfect Match!
                            </div>
                            <div className="text-sm" style={{ color: 'var(--success-color)' }}>
                                All files are identical between the two locations.
                            </div>
                        </div>
                    )}

                    {/* Detailed Tables */}
                    {(onlyInSource.length > 0 || onlyInDest.length > 0 || different.length > 0) && (
                        <>
                            {onlyInSource.length > 0 && (
                                <div id="only-source" className="mt-6" style={{ 
                                    scrollMarginTop: '80px',
                                    paddingTop: '16px',
                                    borderTop: '1px solid var(--border-color)'
                                }}>
                                    <h4 style={{
                                        fontSize: '16px',
                                        fontWeight: '600',
                                        color: 'var(--text-primary)',
                                        marginBottom: '12px',
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: '8px'
                                    }}>
                                        <FileText size={18} style={{ color: 'var(--primary-color)' }} />
                                        Only in Source ({compareResults.sourceAlias})
                                    </h4>
                                    {renderTable(
                                        onlyInSource,
                                        'onlyInSource',
                                        FileText,
                                        'No files found only in source'
                                    )}
                                </div>
                            )}
                            
                            {onlyInDest.length > 0 && (
                                <div id="only-dest" className="mt-6" style={{ 
                                    scrollMarginTop: '80px',
                                    paddingTop: '16px',
                                    borderTop: '1px solid var(--border-color)'
                                }}>
                                    <h4 style={{
                                        fontSize: '16px',
                                        fontWeight: '600',
                                        color: 'var(--text-primary)',
                                        marginBottom: '12px',
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: '8px'
                                    }}>
                                        <FileX size={18} style={{ color: 'var(--danger-color)' }} />
                                        Only in Destination ({compareResults.destAlias})
                                    </h4>
                                    {renderTable(
                                        onlyInDest,
                                        'onlyInDest',
                                        FileX,
                                        'No files found only in destination'
                                    )}
                                </div>
                            )}
                            
                            {different.length > 0 && (
                                <div id="different" className="mt-6" style={{ 
                                    scrollMarginTop: '80px',
                                    paddingTop: '16px',
                                    borderTop: '1px solid var(--border-color)'
                                }}>
                                    <h4 style={{
                                        fontSize: '16px',
                                        fontWeight: '600',
                                        color: 'var(--text-primary)',
                                        marginBottom: '12px',
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: '8px'
                                    }}>
                                        <AlertTriangle size={18} style={{ color: 'var(--warning-color)' }} />
                                        Different Content
                                    </h4>
                                    {renderTable(
                                        different,
                                        'different',
                                        AlertTriangle,
                                        'No differences found'
                                    )}
                                </div>
                            )}
                        </>
                    )}
                </div>
            </div>
        );
    };

    const sourceBuckets = availableBuckets[compareFormData.sourceAlias] || [];
    const isDisabled = isRunning || !compareFormData.sourceAlias || !compareFormData.destAlias || !compareFormData.bucket;

    return (
        <div>
            {error && <ErrorAlert message={error} onClose={() => setError(null)} />}
            
            <div className="card">
                <div className="card-header">
                    <div style={{ display: 'flex', alignItems: 'center' }}>
                        <GitCompare style={{ width: '20px', height: '20px', marginRight: '8px', color: '#2563eb' }} />
                        <h3 className="card-title">Compare Buckets/Paths</h3>
                    </div>
                    <p style={{ margin: '8px 0 0 0', fontSize: '13px', color: 'var(--text-secondary)' }}>
                        Compare content between two MinIO aliases to identify differences
                    </p>
                </div>

                <div style={{ padding: '20px' }}>
                    {/* Compare Configuration Form */}
                    <div style={{ 
                        display: 'grid',
                        gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
                        gap: '16px',
                        marginBottom: '16px'
                    }}>
                        {/* Source Alias */}
                        <div>
                            <label style={{ 
                                display: 'block',
                                fontSize: '13px',
                                color: '#4b5563',
                                marginBottom: '6px',
                                fontWeight: '500'
                            }}>
                                Source Alias
                                <span style={{ color: '#dc2626', marginLeft: '4px' }}>*</span>
                            </label>
                            <select 
                                value={compareFormData.sourceAlias}
                                onChange={(e) => handleSourceAliasChange(e.target.value)}
                                style={{ 
                                    width: '100%',
                                    padding: '10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '14px',
                                    backgroundColor: '#ffffff'
                                }}
                            >
                                <option value="">Select source alias...</option>
                                {sites.map(site => (
                                    <option key={site.name} value={site.name}>{site.name}</option>
                                ))}
                            </select>
                        </div>

                        {/* Destination Alias */}
                        <div>
                            <label style={{ 
                                display: 'block',
                                fontSize: '13px',
                                color: '#4b5563',
                                marginBottom: '6px',
                                fontWeight: '500'
                            }}>
                                Destination Alias
                                <span style={{ color: '#dc2626', marginLeft: '4px' }}>*</span>
                            </label>
                            <select 
                                value={compareFormData.destAlias}
                                onChange={(e) => handleDestAliasChange(e.target.value)}
                                style={{ 
                                    width: '100%',
                                    padding: '10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '14px',
                                    backgroundColor: '#ffffff'
                                }}
                            >
                                <option value="">Select destination alias...</option>
                                {sites.map(site => (
                                    <option key={site.name} value={site.name}>{site.name}</option>
                                ))}
                            </select>
                        </div>

                        {/* Bucket */}
                        <div>
                            <label style={{ 
                                display: 'block',
                                fontSize: '13px',
                                color: '#4b5563',
                                marginBottom: '6px',
                                fontWeight: '500'
                            }}>
                                Bucket
                                <span style={{ color: '#dc2626', marginLeft: '4px' }}>*</span>
                            </label>
                            <select 
                                value={compareFormData.bucket}
                                onChange={(e) => handleBucketChange(e.target.value)}
                                disabled={!compareFormData.sourceAlias}
                                style={{ 
                                    width: '100%',
                                    padding: '10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '14px',
                                    backgroundColor: !compareFormData.sourceAlias ? '#f3f4f6' : '#ffffff',
                                    cursor: !compareFormData.sourceAlias ? 'not-allowed' : 'pointer'
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

                        {/* Path */}
                        <div>
                            <label style={{ 
                                display: 'block',
                                fontSize: '13px',
                                color: '#4b5563',
                                marginBottom: '6px',
                                fontWeight: '500'
                            }}>
                                Path (Optional)
                            </label>
                            <input 
                                type="text"
                                value={compareFormData.path}
                                onChange={(e) => setCompareFormData(prev => ({ ...prev, path: e.target.value }))}
                                placeholder={compareFormData.bucket ? "Enter path or leave empty" : "Select bucket first"}
                                disabled={!compareFormData.bucket}
                                style={{ 
                                    width: '100%',
                                    padding: '10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '14px',
                                    backgroundColor: !compareFormData.bucket ? '#f3f4f6' : '#ffffff',
                                    cursor: !compareFormData.bucket ? 'not-allowed' : 'text'
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

                    {/* Options */}
                    <div style={{ marginBottom: '20px' }}>
                        <label style={{ 
                            display: 'flex',
                            alignItems: 'center',
                            padding: '12px',
                            border: compareFormData.insecure ? '1px solid #f59e0b' : '1px solid #d1d5db',
                            borderRadius: '6px',
                            cursor: 'pointer',
                            backgroundColor: compareFormData.insecure ? '#fffbeb' : '#ffffff',
                            fontSize: '13px',
                            gap: '8px',
                            marginBottom: '8px'
                        }}>
                            <input 
                                type="checkbox"
                                checked={compareFormData.insecure}
                                onChange={(e) => setCompareFormData(prev => ({ ...prev, insecure: e.target.checked }))}
                            />
                            <span style={{ 
                                fontWeight: compareFormData.insecure ? '600' : '400',
                                color: compareFormData.insecure ? '#f59e0b' : '#374151'
                            }}>
                                Skip TLS certificate verification (--insecure)
                            </span>
                        </label>

                        <label style={{ 
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            fontSize: '14px',
                            color: '#374151',
                            cursor: versioningStatus.bothVersioned ? 'pointer' : 'not-allowed',
                            opacity: versioningStatus.bothVersioned ? 1 : 0.6
                        }}>
                            <input 
                                type="checkbox"
                                checked={compareFormData.compareVersion}
                                onChange={(e) => setCompareFormData(prev => ({ ...prev, compareVersion: e.target.checked }))}
                                disabled={!versioningStatus.bothVersioned}
                                style={{ cursor: versioningStatus.bothVersioned ? 'pointer' : 'not-allowed' }}
                            />
                            Compare all versions (require bucket versioning enabled on both aliases)
                        </label>
                    </div>

                    <button 
                        className="btn btn-primary"
                        onClick={executeCompare}
                        disabled={isDisabled}
                        style={{ 
                            width: '100%',
                            display: 'inline-flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            gap: '8px',
                            padding: '12px 18px',
                            fontSize: '15px'
                        }}
                    >
                        {isRunning ? (
                            <>
                                <div className="spinner" style={{ width: '16px', height: '16px' }}></div>
                                Comparing...
                            </>
                        ) : (
                            <>
                                <Play size={16} />
                                Execute Compare
                            </>
                        )}
                    </button>

                    {/* Render Compare Results */}
                    {renderCompareResults()}
                </div>
            </div>
        </div>
    );
};

export default CompareOperations;