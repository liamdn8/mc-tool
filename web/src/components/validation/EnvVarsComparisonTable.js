import React, { useState, useMemo, useEffect } from 'react';
import { CheckCircle, XCircle, AlertCircle, Search, Info } from 'lucide-react';

const EnvVarsComparisonTable = ({ envVars, aliases }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [filterStatus, setFilterStatus] = useState('all');
    const [currentPage, setCurrentPage] = useState(1);
    const [pageSize, setPageSize] = useState(20);

    if (!envVars || envVars.length === 0) return null;

    // Collect all unique environment variable keys across all aliases
    const allEnvKeys = new Set();
    const envDataByAlias = {};

    envVars.forEach(env => {
        envDataByAlias[env.alias] = {
            version: env.version,
            commitID: env.commitID,
            status: env.status,
            error: env.error,
            vars: env.filteredVars || {}
        };
        
        if (env.filteredVars) {
            Object.keys(env.filteredVars).forEach(key => allEnvKeys.add(key));
        }
    });

    const sortedEnvKeys = Array.from(allEnvKeys).sort();

    const getVarStatus = (varName) => {
        const values = new Set();
        const nonRedactedValues = new Set();
        let hasError = false;
        let notConfiguredCount = 0;
        let allRedacted = true;
        
        envVars.forEach(env => {
            if (env.status === 'error') {
                hasError = true;
            } else if (env.filteredVars && env.filteredVars[varName]) {
                const value = env.filteredVars[varName];
                values.add(value);
                
                // Check if value is redacted
                if (value !== '*** EXISTS, REDACTED ***') {
                    nonRedactedValues.add(value);
                    allRedacted = false;
                }
            } else {
                notConfiguredCount++;
                allRedacted = false;
            }
        });

        // If all configured values are REDACTED, treat as info (valid)
        if (allRedacted && values.size > 0) return 'redacted';
        
        if (hasError) return 'warning';
        if (values.size === 0) return 'not_config';
        
        // Compare only non-redacted values
        if (nonRedactedValues.size === 1 && notConfiguredCount === 0) return 'valid';
        if (nonRedactedValues.size === 0 && notConfiguredCount === 0) return 'valid'; // All redacted
        
        return 'invalid';
    };

    // Filter and search
    const filteredVars = useMemo(() => {
        return sortedEnvKeys.filter(varName => {
            const matchesSearch = varName.toLowerCase().includes(searchTerm.toLowerCase());
            if (!matchesSearch) return false;

            const status = getVarStatus(varName);
            if (filterStatus === 'all') return true;
            return status === filterStatus;
        });
    }, [sortedEnvKeys, searchTerm, filterStatus]);

    // Reset to first page when search/filter changes
    useEffect(() => {
        setCurrentPage(1);
    }, [searchTerm, filterStatus]);

    // Pagination
    const paginatedVars = useMemo(() => {
        const startIndex = (currentPage - 1) * pageSize;
        const endIndex = startIndex + pageSize;
        return filteredVars.slice(startIndex, endIndex);
    }, [filteredVars, currentPage, pageSize]);

    const totalPages = useMemo(() => {
        return Math.ceil(filteredVars.length / pageSize);
    }, [filteredVars.length, pageSize]);

    const handlePageChange = (newPage) => {
        setCurrentPage(Math.max(1, Math.min(newPage, totalPages)));
    };

    // Calculate summary
    const summary = useMemo(() => {
        let valid = 0, invalid = 0, notConfig = 0, redacted = 0;
        sortedEnvKeys.forEach(varName => {
            const status = getVarStatus(varName);
            if (status === 'valid') valid++;
            else if (status === 'redacted') {
                redacted++;
                valid++; // Count redacted as valid for overall status
            }
            else if (status === 'invalid') invalid++;
            else if (status === 'not_config') notConfig++;
        });
        return { valid, invalid, notConfig, redacted };
    }, [sortedEnvKeys]);

    const getStatusIcon = (status) => {
        switch (status) {
            case 'valid':
                return <CheckCircle size={20} title="Valid" style={{ color: 'var(--success-color)' }} />;
            case 'redacted':
                return <Info size={20} title="Redacted" style={{ color: '#2563eb' }} />;
            case 'invalid':
                return <XCircle size={20} title="Invalid" style={{ color: 'var(--danger-color)' }} />;
            case 'not_config':
                return <AlertCircle size={20} title="Not Config" style={{ color: 'var(--text-secondary)' }} />;
            default:
                return <AlertCircle size={20} title="Warning" style={{ color: 'var(--warning-color)' }} />;
        }
    };

    const getSeverityIcon = () => {
        if (summary.invalid > 0) return <XCircle size={20} style={{ color: 'var(--danger-color)' }} />;
        if (summary.notConfig > 0) return <AlertCircle size={20} style={{ color: 'var(--text-secondary)' }} />;
        return <CheckCircle size={20} style={{ color: 'var(--success-color)' }} />;
    };

    return (
        <div id="env_vars" style={{ marginBottom: '24px' }}>
            {/* Header with summary */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px', flexWrap: 'wrap' }}>
                <h4 className="card-title" style={{ margin: 0, fontSize: '16px' }}>
                    MinIO Environment Variables
                </h4>
                {getSeverityIcon()}
                <div style={{ display: 'flex', gap: '8px', marginLeft: 'auto' }}>
                    {summary.valid > 0 && (
                        <span className="badge badge-success" style={{ fontSize: '12px' }}>
                            Valid: {summary.valid}
                        </span>
                    )}
                    {summary.invalid > 0 && (
                        <span className="badge badge-danger" style={{ fontSize: '12px' }}>
                            Invalid: {summary.invalid}
                        </span>
                    )}
                    {summary.notConfig > 0 && (
                        <span className="badge badge-secondary" style={{ fontSize: '12px' }}>
                            Not Config: {summary.notConfig}
                        </span>
                    )}
                </div>
            </div>

            {/* Search and Filter */}
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
                        placeholder="Search variable name..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
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
                    value={filterStatus}
                    onChange={(e) => setFilterStatus(e.target.value)}
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
                    <option value="redacted">Redacted</option>
                    <option value="invalid">Invalid</option>
                    <option value="not_config">Not Config</option>
                </select>
            </div>

            {/* Pagination Controls - Top */}
            {totalPages > 1 && (
                <div style={{ 
                    display: 'flex', 
                    flexWrap: 'wrap',
                    justifyContent: 'space-between', 
                    alignItems: 'center',
                    gap: '12px',
                    padding: '12px 0',
                    borderTop: 'none'
                }}>
                    <span style={{ fontSize: '12px', color: '#4b5563' }}>
                        {((currentPage - 1) * pageSize) + 1}-{Math.min(currentPage * pageSize, filteredVars.length)} of {filteredVars.length}
                    </span>
                    
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                        <label style={{ fontSize: '12px', color: '#4b5563', display: 'inline-flex', alignItems: 'center', gap: '6px' }}>
                            Rows per page
                            <select
                                value={pageSize}
                                onChange={(e) => {
                                    setPageSize(Number(e.target.value));
                                    setCurrentPage(1);
                                }}
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
                                onClick={() => handlePageChange(currentPage - 1)}
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
                                onClick={() => handlePageChange(currentPage + 1)}
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
            )}

            {/* Environment Variables Table */}
            <div className="table-container">
                <div style={{ overflowX: 'auto' }}>
                    <table className="table">
                        <thead>
                            <tr>
                                <th style={{ textAlign: 'center', minWidth: '80px' }}>Status</th>
                                <th style={{ position: 'sticky', left: 0, backgroundColor: 'var(--bg-secondary)', zIndex: 1 }}>
                                    Variable Name
                                </th>
                                {aliases.map(alias => (
                                    <th key={alias} style={{ textAlign: 'center', minWidth: '150px' }}>
                                        {alias}
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody>
                            {paginatedVars.map((varName) => {
                                const varStatus = getVarStatus(varName);
                                
                                return (
                                    <tr key={varName}>
                                        <td style={{ textAlign: 'center' }}>
                                            {getStatusIcon(varStatus)}
                                        </td>
                                        <td style={{ 
                                            position: 'sticky',
                                            left: 0,
                                            backgroundColor: 'var(--bg-primary)',
                                            zIndex: 1,
                                            fontFamily: 'monospace',
                                            fontWeight: 500
                                        }}>
                                            {varName}
                                        </td>
                                        {aliases.map(alias => {
                                            const aliasData = envDataByAlias[alias];
                                            const value = aliasData?.vars?.[varName];
                                            const hasValue = value !== undefined;
                                            
                                            return (
                                                <td key={alias} style={{ textAlign: 'center' }}>
                                                    {aliasData?.status === 'error' ? (
                                                        <span className="badge badge-danger">
                                                            <AlertCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            Error
                                                        </span>
                                                    ) : hasValue ? (
                                                        <div style={{ 
                                                            fontFamily: 'monospace',
                                                            fontSize: '12px',
                                                            maxWidth: '200px',
                                                            overflow: 'hidden',
                                                            textOverflow: 'ellipsis',
                                                            whiteSpace: 'nowrap',
                                                            margin: '0 auto',
                                                            color: 'var(--text-primary)'
                                                        }} title={value}>
                                                            {value}
                                                        </div>
                                                    ) : (
                                                        <span className="badge badge-secondary">
                                                            <AlertCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            Not Config
                                                        </span>
                                                    )}
                                                </td>
                                            );
                                        })}
                                    </tr>
                                );
                            })}
                            {paginatedVars.length === 0 && (
                                <tr>
                                    <td colSpan={2 + aliases.length} style={{ 
                                        padding: '20px',
                                        textAlign: 'center',
                                        color: '#6b7280',
                                        fontStyle: 'italic'
                                    }}>
                                        {searchTerm || filterStatus !== 'all' 
                                            ? 'No variables match your search criteria'
                                            : 'No custom environment variables found (PORT and SERVICE vars are filtered out)'}
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Pagination Controls - Bottom */}
            {totalPages > 1 && (
                <div style={{ 
                    display: 'flex', 
                    flexWrap: 'wrap',
                    justifyContent: 'space-between', 
                    alignItems: 'center',
                    gap: '12px',
                    padding: '12px 0 0',
                    borderTop: 'none'
                }}>
                    <span style={{ fontSize: '12px', color: '#4b5563' }}>
                        {((currentPage - 1) * pageSize) + 1}-{Math.min(currentPage * pageSize, filteredVars.length)} of {filteredVars.length}
                    </span>
                    
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                        <label style={{ fontSize: '12px', color: '#4b5563', display: 'inline-flex', alignItems: 'center', gap: '6px' }}>
                            Rows per page
                            <select
                                value={pageSize}
                                onChange={(e) => {
                                    setPageSize(Number(e.target.value));
                                    setCurrentPage(1);
                                }}
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
                                onClick={() => handlePageChange(currentPage - 1)}
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
                                onClick={() => handlePageChange(currentPage + 1)}
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
            )}

            {/* Note about REDACTED values */}
            <div style={{ 
                marginTop: '16px',
                padding: '12px 16px',
                backgroundColor: '#eff6ff',
                border: '1px solid #bfdbfe',
                borderRadius: '6px',
                display: 'flex',
                alignItems: 'flex-start',
                gap: '8px'
            }}>
                <Info size={16} style={{ color: '#2563eb', marginTop: '2px', flexShrink: 0 }} />
                <div style={{ fontSize: '13px', color: '#1e40af' }}>
                    <strong>Note:</strong> Environment variables with value <code style={{ 
                        backgroundColor: '#dbeafe',
                        padding: '2px 6px',
                        borderRadius: '3px',
                        fontFamily: 'monospace',
                        fontSize: '12px'
                    }}>*** EXISTS, REDACTED ***</code> are sensitive credentials that MinIO has hidden for security reasons. 
                    These variables are not compared across aliases and are marked as <Info size={14} style={{ display: 'inline-block', verticalAlign: 'middle', margin: '0 2px' }} /> <strong>Redacted</strong> status, 
                    which counts as valid in the overall validation result.
                </div>
            </div>
        </div>
    );
};

export default EnvVarsComparisonTable;
