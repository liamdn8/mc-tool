import React, { useState, useMemo } from 'react';
import { CheckCircle, XCircle, AlertCircle, ChevronRight, ChevronLeft, Search, Eye } from 'lucide-react';

const ResourceComparisonTable = ({ resourceType, resources, baseline, targets, onViewDiff }) => {
    const [statusFilter, setStatusFilter] = useState('all');
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 10;

    if (!resources || resources.length === 0) return null;

    const allNamespaces = [baseline, ...(targets || [])];

    // Calculate status counts
    const statusCounts = resources.reduce((acc, row) => {
        let hasMatch = false;
        let hasMismatch = false;
        let hasNotFound = false;
        let hasExtra = false;

        allNamespaces.forEach(ns => {
            const cell = row[ns];
            if (!cell || cell.status === 'not_found') {
                hasNotFound = true;
            } else if (cell.status === 'match') {
                hasMatch = true;
            } else if (cell.status === 'mismatch') {
                hasMismatch = true;
            } else if (cell.status === 'extra') {
                hasExtra = true;
            }
        });

        if (hasExtra) {
            acc.extra++;
        } else if (hasNotFound) {
            acc.notFound++;
        } else if (hasMismatch) {
            acc.mismatch++;
        } else if (hasMatch) {
            acc.match++;
        }

        return acc;
    }, { match: 0, mismatch: 0, notFound: 0, extra: 0 });

    const getRowStatus = (row) => {
        let hasMatch = false;
        let hasMismatch = false;
        let hasNotFound = false;
        let hasExtra = false;

        allNamespaces.forEach(ns => {
            const cell = row[ns];
            if (!cell || cell.status === 'not_found') {
                hasNotFound = true;
            } else if (cell.status === 'match') {
                hasMatch = true;
            } else if (cell.status === 'mismatch') {
                hasMismatch = true;
            } else if (cell.status === 'extra') {
                hasExtra = true;
            }
        });

        if (hasExtra) {
            return { status: 'extra', label: 'Extra (Not in Baseline)', icon: AlertCircle, color: 'var(--info-color)' };
        } else if (hasNotFound) {
            return { status: 'not_found', label: 'Not Found', icon: AlertCircle, color: 'var(--warning-color)' };
        } else if (hasMismatch) {
            return { status: 'mismatch', label: 'Mismatch', icon: XCircle, color: 'var(--danger-color)' };
        } else {
            return { status: 'match', label: 'Match', icon: CheckCircle, color: 'var(--success-color)' };
        }
    };

    // Filter and paginate data
    const filteredData = useMemo(() => {
        let filtered = resources;

        // Apply status filter
        if (statusFilter !== 'all') {
            filtered = filtered.filter(row => {
                const rowStatus = getRowStatus(row);
                return rowStatus.status === statusFilter;
            });
        }

        // Apply search filter
        if (searchQuery.trim()) {
            const searchLower = searchQuery.toLowerCase();
            filtered = filtered.filter(row => 
                row.resource_name.toLowerCase().includes(searchLower)
            );
        }

        return filtered;
    }, [resources, statusFilter, searchQuery]);

    const totalPages = Math.ceil(filteredData.length / itemsPerPage);
    const paginatedData = filteredData.slice(
        (currentPage - 1) * itemsPerPage,
        currentPage * itemsPerPage
    );

    // Reset to page 1 when filters change
    useMemo(() => {
        setCurrentPage(1);
    }, [statusFilter, searchQuery]);

    const severity = statusCounts.mismatch > 0 || statusCounts.notFound > 0 
        ? 'danger' 
        : statusCounts.match > 0 
            ? 'success' 
            : 'secondary';

    return (
        <div style={{ marginBottom: '24px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px', flexWrap: 'wrap' }}>
                <h4 className="card-title" style={{ margin: 0, fontSize: '16px' }}>
                    {resourceType}
                </h4>
                {severity === 'success' && <CheckCircle size={20} style={{ color: 'var(--success-color)' }} />}
                {severity === 'danger' && <XCircle size={20} style={{ color: 'var(--danger-color)' }} />}
                
                {/* Status count badges */}
                <div style={{ display: 'flex', gap: '8px', marginLeft: 'auto' }}>
                    {statusCounts.match > 0 && (
                        <span className="badge badge-success" style={{ fontSize: '12px' }}>
                            Match: {statusCounts.match}
                        </span>
                    )}
                    {statusCounts.mismatch > 0 && (
                        <span className="badge badge-danger" style={{ fontSize: '12px' }}>
                            Mismatch: {statusCounts.mismatch}
                        </span>
                    )}
                    {statusCounts.notFound > 0 && (
                        <span className="badge badge-warning" style={{ fontSize: '12px' }}>
                            Not Found: {statusCounts.notFound}
                        </span>
                    )}
                    {statusCounts.extra > 0 && (
                        <span className="badge badge-info" style={{ fontSize: '12px' }}>
                            Extra: {statusCounts.extra}
                        </span>
                    )}
                </div>
            </div>

            {/* Filters */}
            <div style={{ display: 'flex', gap: '12px', marginBottom: '16px', flexWrap: 'wrap' }}>
                <div style={{ position: 'relative', flex: '1 1 250px' }}>
                    <Search size={16} style={{ 
                        position: 'absolute', 
                        left: '12px', 
                        top: '50%', 
                        transform: 'translateY(-50%)', 
                        color: '#6b7280' 
                    }} />
                    <input
                        type="text"
                        placeholder="Search resource name..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        style={{
                            width: '100%',
                            padding: '8px 12px 8px 36px',
                            fontSize: '14px',
                            border: '1px solid #d1d5db',
                            borderRadius: '6px'
                        }}
                    />
                </div>
                <select
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value)}
                    style={{
                        padding: '8px 12px',
                        fontSize: '14px',
                        border: '1px solid #d1d5db',
                        borderRadius: '6px',
                        minWidth: '150px'
                    }}
                >
                    <option value="all">All Status</option>
                    <option value="match">Match</option>
                    <option value="mismatch">Mismatch</option>
                    <option value="not_found">Not Found</option>
                </select>
                <div style={{ 
                    fontSize: '13px', 
                    color: '#6b7280', 
                    display: 'flex', 
                    alignItems: 'center',
                    marginLeft: 'auto'
                }}>
                    Showing {paginatedData.length} of {filteredData.length} resources
                </div>
            </div>

            <div className="table-container">
                <div style={{ overflowX: 'auto' }}>
                    <table className="table">
                        <thead>
                            <tr>
                                <th style={{ textAlign: 'center', minWidth: '80px' }}>
                                    Status
                                </th>
                                <th style={{ position: 'sticky', left: 0, backgroundColor: 'var(--bg-secondary)', zIndex: 1, minWidth: '200px' }}>
                                    Resource Name
                                </th>
                                {allNamespaces.map(ns => (
                                    <th key={ns} style={{ textAlign: 'center', minWidth: '150px' }}>
                                        {ns}
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody>
                            {paginatedData.map((row) => {
                                const rowStatus = getRowStatus(row);
                                const StatusIcon = rowStatus.icon;
                                return (
                                    <tr key={row.resource_name}>
                                        <td style={{ textAlign: 'center' }}>
                                            <StatusIcon 
                                                size={20} 
                                                style={{ color: rowStatus.color }}
                                                title={rowStatus.label}
                                            />
                                        </td>
                                        <td style={{ 
                                            position: 'sticky', 
                                            left: 0, 
                                            backgroundColor: 'var(--bg-primary)', 
                                            zIndex: 1,
                                            fontFamily: 'monospace',
                                            fontWeight: 500
                                        }}>
                                            {row.resource_name}
                                        </td>
                                        {allNamespaces.map(ns => {
                                            const cell = row[ns];
                                            if (!cell) return <td key={ns} style={{ textAlign: 'center' }}>-</td>;
                                            
                                            const isBaseline = ns === baseline;
                                            
                                            return (
                                                <td key={ns} style={{ textAlign: 'center' }}>
                                                    {cell.status === 'extra' ? (
                                                        isBaseline ? (
                                                            <span className="badge badge-warning">
                                                                <AlertCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                                Not Found
                                                            </span>
                                                        ) : (
                                                            <button
                                                                onClick={() => onViewDiff(resourceType, row.resource_name, baseline, ns)}
                                                                className="badge"
                                                                style={{
                                                                    cursor: 'pointer',
                                                                    border: 'none',
                                                                    backgroundColor: 'var(--warning-light)',
                                                                    color: 'var(--warning-text)',
                                                                    transition: 'var(--transition)'
                                                                }}
                                                            >
                                                                <AlertCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                                Extra
                                                                <Eye size={14} style={{ marginLeft: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            </button>
                                                        )
                                                    ) : cell.status === 'not_found' ? (
                                                        <span className="badge badge-warning">
                                                            <AlertCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            Not Found
                                                        </span>
                                                    ) : cell.status === 'match' ? (
                                                        <button
                                                            onClick={() => onViewDiff(resourceType, row.resource_name, baseline, ns)}
                                                            className="badge badge-success"
                                                            style={{
                                                                cursor: 'pointer',
                                                                border: 'none',
                                                                transition: 'var(--transition)'
                                                            }}
                                                        >
                                                            <CheckCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            Match
                                                            <Eye size={14} style={{ marginLeft: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                        </button>
                                                    ) : cell.status === 'mismatch' ? (
                                                        isBaseline ? (
                                                            <span className="badge badge-success">
                                                                <CheckCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                                Configured
                                                            </span>
                                                        ) : (
                                                            <button
                                                                onClick={() => onViewDiff(resourceType, row.resource_name, baseline, ns)}
                                                                className="badge"
                                                                style={{
                                                                    cursor: 'pointer',
                                                                    border: 'none',
                                                                    backgroundColor: 'var(--warning-light)',
                                                                    color: 'var(--warning-text)',
                                                                    transition: 'var(--transition)'
                                                                }}
                                                            >
                                                                <XCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                                Mismatch
                                                                <Eye size={14} style={{ marginLeft: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            </button>
                                                        )
                                                    ) : cell.status === '-' ? (
                                                        <span className="badge badge-success">
                                                            <CheckCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            Configured
                                                        </span>
                                                    ) : (
                                                        <span className="badge badge-secondary">-</span>
                                                    )}
                                                </td>
                                            );
                                        })}
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
                <div style={{ 
                    display: 'flex', 
                    justifyContent: 'center', 
                    alignItems: 'center', 
                    gap: '8px',
                    marginTop: '16px'
                }}>
                    <button
                        onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                        disabled={currentPage === 1}
                        className="btn"
                        style={{
                            padding: '6px 12px',
                            fontSize: '14px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '4px',
                            opacity: currentPage === 1 ? 0.5 : 1,
                            cursor: currentPage === 1 ? 'not-allowed' : 'pointer'
                        }}
                    >
                        <ChevronLeft size={16} />
                        Previous
                    </button>
                    
                    <div style={{ display: 'flex', gap: '4px' }}>
                        {[...Array(totalPages)].map((_, i) => {
                            const pageNum = i + 1;
                            if (
                                pageNum === 1 ||
                                pageNum === totalPages ||
                                (pageNum >= currentPage - 1 && pageNum <= currentPage + 1)
                            ) {
                                return (
                                    <button
                                        key={pageNum}
                                        onClick={() => setCurrentPage(pageNum)}
                                        className="btn"
                                        style={{
                                            padding: '6px 12px',
                                            fontSize: '14px',
                                            minWidth: '40px',
                                            backgroundColor: currentPage === pageNum ? 'var(--primary-color)' : 'white',
                                            color: currentPage === pageNum ? 'white' : 'var(--text-primary)',
                                            border: currentPage === pageNum ? '1px solid var(--primary-color)' : '1px solid #d1d5db'
                                        }}
                                    >
                                        {pageNum}
                                    </button>
                                );
                            } else if (
                                pageNum === currentPage - 2 ||
                                pageNum === currentPage + 2
                            ) {
                                return <span key={pageNum} style={{ padding: '6px 4px' }}>...</span>;
                            }
                            return null;
                        })}
                    </div>

                    <button
                        onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                        disabled={currentPage === totalPages}
                        className="btn"
                        style={{
                            padding: '6px 12px',
                            fontSize: '14px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '4px',
                            opacity: currentPage === totalPages ? 0.5 : 1,
                            cursor: currentPage === totalPages ? 'not-allowed' : 'pointer'
                        }}
                    >
                        Next
                        <ChevronRight size={16} />
                    </button>
                </div>
            )}
        </div>
    );
};

export default ResourceComparisonTable;
