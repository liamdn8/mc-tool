import React, { useState, useMemo } from 'react';
import { CheckCircle, XCircle, AlertCircle, ChevronRight, ChevronLeft, Search } from 'lucide-react';

const ConfigComparisonTable = ({ configType, configTable, validationResults, calculateTableSeverity, onViewConfig }) => {
    const [statusFilter, setStatusFilter] = useState('all');
    const [bucketFilter, setBucketFilter] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 10;

    if (!configTable || configTable.length === 0) return null;

    const aliases = validationResults.aliases || [];
    const severity = calculateTableSeverity(configTable, validationResults.buckets, aliases);

    const title = configType === 'lifecycle_table' ? 'Bucket Lifecycle Configuration' : 'Bucket Event Notifications';

    // Calculate status counts
    const statusCounts = configTable.reduce((acc, row) => {
        let hasMatch = false;
        let hasMismatch = false;
        let hasNotConfigured = false;
        let hasNotExist = false;

        aliases.forEach(alias => {
            const cell = row[alias];
            if (!cell || cell.status === 'not_exist') {
                hasNotExist = true;
            } else if (cell.status === 'not_configured') {
                hasNotConfigured = true;
            } else if (cell.status === 'match') {
                hasMatch = true;
            } else if (cell.status === 'mismatch') {
                hasMismatch = true;
            }
        });

        // Determine overall row status
        if (hasNotExist) {
            // If any bucket doesn't exist, it's invalid
            acc.invalid++;
        } else if (!hasMatch && !hasMismatch && hasNotConfigured) {
            // All are not configured
            acc.notConfig++;
        } else if (hasMatch && !hasMismatch) {
            // All configured sites match
            acc.valid++;
        } else {
            // Has mismatches
            acc.invalid++;
        }

        return acc;
    }, { valid: 0, invalid: 0, notConfig: 0 });

    const getRowStatus = (row) => {
        let hasMatch = false;
        let hasMismatch = false;
        let hasNotConfigured = false;
        let hasNotExist = false;

        aliases.forEach(alias => {
            const cell = row[alias];
            if (!cell || cell.status === 'not_exist') {
                hasNotExist = true;
            } else if (cell.status === 'not_configured') {
                hasNotConfigured = true;
            } else if (cell.status === 'match') {
                hasMatch = true;
            } else if (cell.status === 'mismatch') {
                hasMismatch = true;
            }
        });

        if (hasNotExist) {
            return { status: 'invalid', label: 'Invalid', icon: XCircle, color: 'var(--danger-color)' };
        } else if (!hasMatch && !hasMismatch && hasNotConfigured) {
            return { status: 'not_config', label: 'Not Config', icon: AlertCircle, color: 'var(--text-secondary)' };
        } else if (hasMatch && !hasMismatch) {
            return { status: 'valid', label: 'Valid', icon: CheckCircle, color: 'var(--success-color)' };
        } else {
            return { status: 'invalid', label: 'Invalid', icon: XCircle, color: 'var(--danger-color)' };
        }
    };

    // Filter and paginate data
    const filteredData = useMemo(() => {
        let filtered = configTable;

        // Apply status filter
        if (statusFilter !== 'all') {
            filtered = filtered.filter(row => {
                const rowStatus = getRowStatus(row);
                return rowStatus.status === statusFilter;
            });
        }

        // Apply bucket name filter
        if (bucketFilter.trim()) {
            const searchLower = bucketFilter.toLowerCase();
            filtered = filtered.filter(row => 
                row.bucket.toLowerCase().includes(searchLower)
            );
        }

        return filtered;
    }, [configTable, statusFilter, bucketFilter, aliases]);

    const totalPages = Math.ceil(filteredData.length / itemsPerPage);
    const paginatedData = filteredData.slice(
        (currentPage - 1) * itemsPerPage,
        currentPage * itemsPerPage
    );

    // Reset to page 1 when filters change
    useMemo(() => {
        setCurrentPage(1);
    }, [statusFilter, bucketFilter]);

    return (
        <div style={{ marginBottom: '24px' }} id={configType}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px', flexWrap: 'wrap' }}>
                <h4 className="card-title" style={{ margin: 0, fontSize: '16px' }}>
                    {title}
                </h4>
                {severity === 'success' && <CheckCircle size={20} style={{ color: 'var(--success-color)' }} />}
                {severity === 'warning' && <AlertCircle size={20} style={{ color: 'var(--warning-color)' }} />}
                {severity === 'danger' && <XCircle size={20} style={{ color: 'var(--danger-color)' }} />}
                
                {/* Status count badges */}
                <div style={{ display: 'flex', gap: '8px', marginLeft: 'auto' }}>
                    {statusCounts.valid > 0 && (
                        <span className="badge badge-success" style={{ fontSize: '12px' }}>
                            Valid: {statusCounts.valid}
                        </span>
                    )}
                    {statusCounts.invalid > 0 && (
                        <span className="badge badge-danger" style={{ fontSize: '12px' }}>
                            Invalid: {statusCounts.invalid}
                        </span>
                    )}
                    {statusCounts.notConfig > 0 && (
                        <span className="badge badge-secondary" style={{ fontSize: '12px' }}>
                            Not Config: {statusCounts.notConfig}
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
                        placeholder="Search bucket name..."
                        value={bucketFilter}
                        onChange={(e) => setBucketFilter(e.target.value)}
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
                    <option value="valid">Valid</option>
                    <option value="invalid">Invalid</option>
                    <option value="not_config">Not Config</option>
                </select>
                <div style={{ 
                    fontSize: '13px', 
                    color: '#6b7280', 
                    display: 'flex', 
                    alignItems: 'center',
                    marginLeft: 'auto'
                }}>
                    Showing {paginatedData.length} of {filteredData.length} buckets
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
                                <th style={{ position: 'sticky', left: 0, backgroundColor: 'var(--bg-secondary)', zIndex: 1 }}>
                                    Bucket
                                </th>
                                {[...aliases].sort().map(alias => (
                                    <th key={alias} style={{ textAlign: 'center', minWidth: '150px' }}>
                                        {alias}
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody>
                            {paginatedData.map((row) => {
                                const rowStatus = getRowStatus(row);
                                const StatusIcon = rowStatus.icon;
                                return (
                                    <tr key={row.bucket}>
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
                                            {row.bucket}
                                        </td>
                                        {[...aliases].sort().map(alias => {
                                            const cell = row[alias];
                                            if (!cell) return <td key={alias} style={{ textAlign: 'center' }}>-</td>;
                                            
                                            return (
                                                <td key={alias} style={{ textAlign: 'center' }}>
                                                    {cell.status === 'not_exist' ? (
                                                        <span className="badge badge-danger">
                                                            <XCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            Not Found
                                                        </span>
                                                    ) : cell.status === 'not_configured' ? (
                                                        <span className="badge badge-secondary">
                                                            <AlertCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            Not Config
                                                        </span>
                                                    ) : (
                                                        <button
                                                            onClick={() => onViewConfig(row.bucket, alias, cell.value)}
                                                            className="badge"
                                                            style={{
                                                                cursor: 'pointer',
                                                                border: 'none',
                                                                backgroundColor: cell.status === 'match' ? 'var(--success-light)' : 'var(--warning-light)',
                                                                color: cell.status === 'match' ? 'var(--success-text)' : 'var(--warning-text)',
                                                                transition: 'var(--transition)'
                                                            }}
                                                        >
                                                            {cell.status === 'match' ? (
                                                                <CheckCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            ) : (
                                                                <AlertCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            )}
                                                            {cell.status === 'match' ? 'Match' : 'Mismatch'}
                                                            <ChevronRight size={14} style={{ marginLeft: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                        </button>
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
                            // Show first page, last page, current page, and pages around current
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

export default ConfigComparisonTable;
