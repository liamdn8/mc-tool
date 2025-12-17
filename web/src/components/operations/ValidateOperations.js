import React, { useState, useEffect, useMemo } from 'react';
import { CheckSquare, Play, CheckCircle, AlertCircle, XCircle, Loader, ChevronDown, ChevronRight } from 'lucide-react';
import { apiCall } from '../../utils/api';

const ValidateOperations = () => {
    const [aliases, setAliases] = useState([]);
    const [selectedAliases, setSelectedAliases] = useState([]);
    const [buckets, setBuckets] = useState([]);
    const [selectedBucket, setSelectedBucket] = useState('');
    const [checkLifecycle, setCheckLifecycle] = useState(true);
    const [checkEvents, setCheckEvents] = useState(true);
    const [validationResults, setValidationResults] = useState(null);
    const [isRunning, setIsRunning] = useState(false);
    const [isLoadingBuckets, setIsLoadingBuckets] = useState(false);
    const [expandedRows, setExpandedRows] = useState({});

    // Load aliases on mount
    useEffect(() => {
        loadAliases();
    }, []);

    // Load buckets when first alias is selected
    useEffect(() => {
        if (selectedAliases.length > 0) {
            loadBuckets(selectedAliases[0]);
        } else {
            setBuckets([]);
            setSelectedBucket('');
        }
    }, [selectedAliases]);

    const loadAliases = async () => {
        try {
            const { response, data } = await apiCall('/api/aliases');
            if (response.ok) {
                setAliases(data.aliases || []);
            }
        } catch (error) {
            console.error('Failed to load aliases:', error);
        }
    };

    const loadBuckets = async (alias) => {
        setIsLoadingBuckets(true);
        try {
            const { response, data } = await apiCall(`/api/operations/buckets?alias=${encodeURIComponent(alias)}`);
            if (response.ok) {
                setBuckets(data.buckets || []);
            } else {
                setBuckets([]);
            }
        } catch (error) {
            console.error('Failed to load buckets:', error);
            setBuckets([]);
        } finally {
            setIsLoadingBuckets(false);
        }
    };

    const handleAliasToggle = (aliasName) => {
        setSelectedAliases(prev => {
            if (prev.includes(aliasName)) {
                return prev.filter(a => a !== aliasName);
            } else {
                return [...prev, aliasName];
            }
        });
    };

    const handleSelectAllAliases = () => {
        if (selectedAliases.length === aliases.length) {
            setSelectedAliases([]);
        } else {
            setSelectedAliases(aliases.map(a => a.name));
        }
    };

    const executeValidation = async () => {
        if (selectedAliases.length === 0) {
            alert('Please select at least one alias');
            return;
        }

        if (!selectedBucket) {
            alert('Please select a bucket');
            return;
        }

        if (!checkLifecycle && !checkEvents) {
            alert('Please select at least one configuration type to validate');
            return;
        }

        setIsRunning(true);
        setValidationResults(null);

        try {
            const { response, data: result } = await apiCall('/api/operations/validate-bucket-config', {
                method: 'POST',
                body: JSON.stringify({
                    aliases: selectedAliases,
                    bucket: selectedBucket,
                    check_lifecycle: checkLifecycle,
                    check_events: checkEvents
                })
            });
            if (response.ok) {
                setValidationResults(result);
            } else {
                alert(`Validation failed: ${result.error || 'Unknown error'}`);
            }
        } catch (error) {
            alert(`Validation failed: ${error.message}`);
        } finally {
            setIsRunning(false);
        }
    };

    // Calculate severity based on configuration consistency
    const calculateSeverity = (comparisons, referenceConfigured) => {
        if (!comparisons || comparisons.length === 0) {
            return referenceConfigured ? 'success' : 'info';
        }

        const totalAliases = comparisons.length;
        const matchCount = comparisons.filter(c => c.status === 'match').length;
        const allNotConfigured = !referenceConfigured && comparisons.every(c => !c.configured);
        
        // All same (all configured and match, or all not configured)
        if (allNotConfigured || matchCount === totalAliases) {
            return 'success';
        }
        
        // Half or more match
        if (matchCount >= totalAliases / 2) {
            return 'warning';
        }
        
        // Less than half match
        return 'danger';
    };

    const renderSummaryCards = () => {
        if (!validationResults) return null;

        const { bucket_existence, missing_buckets, lifecycle, events } = validationResults;
        const totalAliases = selectedAliases.length;
        const foundCount = totalAliases - (missing_buckets?.length || 0);

        const cards = [
            {
                label: 'Selected Aliases',
                value: totalAliases,
                tone: '#2563eb'
            },
            {
                label: 'Bucket Found',
                value: foundCount,
                tone: foundCount === totalAliases ? '#059669' : '#dc2626'
            },
            {
                label: 'Bucket Missing',
                value: missing_buckets?.length || 0,
                tone: missing_buckets?.length > 0 ? '#dc2626' : '#6b7280'
            }
        ];

        if (checkLifecycle && lifecycle) {
            const severity = calculateSeverity(lifecycle.comparisons, lifecycle.reference_configured);
            cards.push({
                label: 'Lifecycle Status',
                value: severity === 'success' ? '✓' : severity === 'warning' ? '⚠' : '✗',
                tone: severity === 'success' ? '#059669' : severity === 'warning' ? '#f59e0b' : '#dc2626'
            });
        }

        if (checkEvents && events) {
            const severity = calculateSeverity(events.comparisons, events.reference_configured);
            cards.push({
                label: 'Events Status',
                value: severity === 'success' ? '✓' : severity === 'warning' ? '⚠' : '✗',
                tone: severity === 'success' ? '#059669' : severity === 'warning' ? '#f59e0b' : '#dc2626'
            });
        }

        return (
            <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))',
                gap: '16px',
                marginBottom: '20px'
            }}>
                {cards.map(card => (
                    <div key={card.label} style={{
                        border: '1px solid #e5e7eb',
                        borderRadius: '8px',
                        padding: '16px',
                        backgroundColor: 'white'
                    }}>
                        <div style={{
                            fontSize: '26px',
                            fontWeight: 600,
                            color: card.tone
                        }}>{card.value}</div>
                        <div style={{
                            marginTop: '4px',
                            fontSize: '13px',
                            color: '#6b7280'
                        }}>{card.label}</div>
                    </div>
                ))}
            </div>
        );
    };

    const renderBucketExistenceTable = () => {
        if (!validationResults || !validationResults.bucket_existence) return null;

        const existence = validationResults.bucket_existence;
        const missingBuckets = validationResults.missing_buckets || [];
        const severity = missingBuckets.length === 0 ? 'success' : 
                        missingBuckets.length >= selectedAliases.length / 2 ? 'danger' : 'warning';

        return (
            <div style={{ marginBottom: '24px' }}>
                <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '8px',
                    marginBottom: '12px' 
                }}>
                    <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#1f2937', margin: 0 }}>
                        Bucket Existence
                    </h4>
                    {severity === 'success' && <CheckCircle size={20} style={{ color: '#059669' }} />}
                    {severity === 'warning' && <AlertCircle size={20} style={{ color: '#f59e0b' }} />}
                    {severity === 'danger' && <XCircle size={20} style={{ color: '#dc2626' }} />}
                </div>

                <div style={{ border: '1px solid #e5e7eb', borderRadius: '6px', overflow: 'hidden' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                        <thead style={{ backgroundColor: '#f9fafb' }}>
                            <tr>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '13px', fontWeight: 600, color: '#6b7280' }}>
                                    Alias
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '13px', fontWeight: 600, color: '#6b7280' }}>
                                    Bucket: {validationResults.bucket}
                                </th>
                                <th style={{ textAlign: 'center', padding: '12px', fontSize: '13px', fontWeight: 600, color: '#6b7280', width: '100px' }}>
                                    Status
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {Object.entries(existence).map(([alias, exists], index) => (
                                <tr key={alias} style={{ 
                                    borderTop: index > 0 ? '1px solid #f3f4f6' : 'none',
                                    backgroundColor: index % 2 === 0 ? 'white' : '#f9fafb'
                                }}>
                                    <td style={{ padding: '12px', fontFamily: 'monospace', fontSize: '13px', fontWeight: 500 }}>
                                        {alias}
                                    </td>
                                    <td style={{ padding: '12px', fontSize: '13px', color: '#4b5563' }}>
                                        {exists ? 'Bucket exists' : 'Bucket not found'}
                                    </td>
                                    <td style={{ padding: '12px', textAlign: 'center' }}>
                                        {exists ? (
                                            <span style={{
                                                display: 'inline-flex',
                                                alignItems: 'center',
                                                gap: '4px',
                                                padding: '4px 10px',
                                                backgroundColor: '#d1fae5',
                                                color: '#065f46',
                                                borderRadius: '12px',
                                                fontSize: '12px',
                                                fontWeight: 500
                                            }}>
                                                <CheckCircle size={14} />
                                                Found
                                            </span>
                                        ) : (
                                            <span style={{
                                                display: 'inline-flex',
                                                alignItems: 'center',
                                                gap: '4px',
                                                padding: '4px 10px',
                                                backgroundColor: '#fee2e2',
                                                color: '#991b1b',
                                                borderRadius: '12px',
                                                fontSize: '12px',
                                                fontWeight: 500
                                            }}>
                                                <XCircle size={14} />
                                                Missing
                                            </span>
                                        )}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        );
    };

    const renderConfigTable = (configType, configData) => {
        if (!configData) return null;

        const { comparisons, reference_configured } = configData;
        if (!comparisons || comparisons.length === 0) return null;

        // Find reference row
        const referenceRow = comparisons.find(c => c.is_reference);
        const comparisonRows = comparisons.filter(c => !c.is_reference);
        
        const severity = calculateSeverity(comparisons, reference_configured);

        const toggleRow = (alias) => {
            setExpandedRows(prev => ({
                ...prev,
                [`${configType}-${alias}`]: !prev[`${configType}-${alias}`]
            }));
        };

        const formatConfig = (configRaw) => {
            if (!configRaw) return 'No configuration';
            try {
                const parsed = JSON.parse(configRaw);
                return JSON.stringify(parsed, null, 2);
            } catch {
                return configRaw;
            }
        };

        return (
            <div style={{ marginBottom: '24px' }}>
                <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '8px',
                    marginBottom: '12px' 
                }}>
                    <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#1f2937', margin: 0, textTransform: 'capitalize' }}>
                        {configType} Configuration
                    </h4>
                    {severity === 'success' && <CheckCircle size={20} style={{ color: '#059669' }} />}
                    {severity === 'warning' && <AlertCircle size={20} style={{ color: '#f59e0b' }} />}
                    {severity === 'danger' && <XCircle size={20} style={{ color: '#dc2626' }} />}
                </div>

                <div style={{ border: '1px solid #e5e7eb', borderRadius: '6px', overflow: 'hidden' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                        <thead style={{ backgroundColor: '#f9fafb' }}>
                            <tr>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '13px', fontWeight: 600, color: '#6b7280', width: '40px' }}>
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '13px', fontWeight: 600, color: '#6b7280' }}>
                                    Alias
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '13px', fontWeight: 600, color: '#6b7280' }}>
                                    Configuration Status
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '13px', fontWeight: 600, color: '#6b7280' }}>
                                    Comparison Result
                                </th>
                                <th style={{ textAlign: 'center', padding: '12px', fontSize: '13px', fontWeight: 600, color: '#6b7280', width: '120px' }}>
                                    Match Status
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {/* Reference row */}
                            {referenceRow && (
                                <>
                                    <tr style={{ backgroundColor: '#eff6ff', borderBottom: '2px solid #2563eb' }}>
                                        <td style={{ padding: '12px' }}>
                                            <button
                                                onClick={() => toggleRow(referenceRow.alias)}
                                                style={{
                                                    background: 'none',
                                                    border: 'none',
                                                    cursor: 'pointer',
                                                    padding: '4px',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    color: '#2563eb'
                                                }}
                                            >
                                                {expandedRows[`${configType}-${referenceRow.alias}`] ? 
                                                    <ChevronDown size={16} /> : <ChevronRight size={16} />}
                                            </button>
                                        </td>
                                        <td style={{ padding: '12px', fontFamily: 'monospace', fontSize: '13px', fontWeight: 600 }}>
                                            {referenceRow.alias} <span style={{ fontSize: '11px', color: '#6b7280', fontWeight: 'normal' }}>(Reference)</span>
                                        </td>
                                        <td style={{ padding: '12px', fontSize: '13px' }}>
                                            {referenceRow.configured ? (
                                                <span style={{ color: '#059669', fontWeight: 500 }}>
                                                    {referenceRow.config_summary || 'Configured'}
                                                </span>
                                            ) : (
                                                <span style={{ color: '#6b7280' }}>Not configured</span>
                                            )}
                                        </td>
                                        <td style={{ padding: '12px', fontSize: '13px', color: '#6b7280' }}>
                                            -
                                        </td>
                                        <td style={{ padding: '12px', textAlign: 'center' }}>
                                            <span style={{
                                                display: 'inline-flex',
                                                alignItems: 'center',
                                                gap: '4px',
                                                padding: '4px 10px',
                                                backgroundColor: '#dbeafe',
                                                color: '#1e40af',
                                                borderRadius: '12px',
                                                fontSize: '12px',
                                                fontWeight: 500
                                            }}>
                                                Reference
                                            </span>
                                        </td>
                                    </tr>
                                    {expandedRows[`${configType}-${referenceRow.alias}`] && (
                                        <tr style={{ backgroundColor: '#f0f9ff' }}>
                                            <td colSpan="5" style={{ padding: '12px' }}>
                                                <div style={{ 
                                                    backgroundColor: '#1e293b', 
                                                    color: '#e2e8f0', 
                                                    padding: '12px', 
                                                    borderRadius: '4px',
                                                    fontSize: '12px',
                                                    fontFamily: 'monospace',
                                                    overflowX: 'auto',
                                                    maxHeight: '300px',
                                                    overflowY: 'auto'
                                                }}>
                                                    <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                                                        {formatConfig(referenceRow.config_raw)}
                                                    </pre>
                                                </div>
                                            </td>
                                        </tr>
                                    )}
                                </>
                            )}
                            
                            {/* Comparison rows */}
                            {comparisonRows.map((comparison, index) => {
                                const isMatch = comparison.status === 'match';
                                const isError = comparison.status === 'error';
                                const isExpanded = expandedRows[`${configType}-${comparison.alias}`];
                                
                                return (
                                    <React.Fragment key={comparison.alias}>
                                        <tr style={{ 
                                            borderTop: '1px solid #f3f4f6',
                                            backgroundColor: index % 2 === 0 ? 'white' : '#f9fafb'
                                        }}>
                                            <td style={{ padding: '12px' }}>
                                                <button
                                                    onClick={() => toggleRow(comparison.alias)}
                                                    style={{
                                                        background: 'none',
                                                        border: 'none',
                                                        cursor: 'pointer',
                                                        padding: '4px',
                                                        display: 'flex',
                                                        alignItems: 'center',
                                                        color: '#6b7280'
                                                    }}
                                                >
                                                    {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                                                </button>
                                            </td>
                                            <td style={{ padding: '12px', fontFamily: 'monospace', fontSize: '13px', fontWeight: 500 }}>
                                                {comparison.alias}
                                            </td>
                                            <td style={{ padding: '12px', fontSize: '13px' }}>
                                                {comparison.configured ? (
                                                    <span style={{ color: '#059669', fontWeight: 500 }}>
                                                        {comparison.config_summary || 'Configured'}
                                                    </span>
                                                ) : (
                                                    <span style={{ color: '#6b7280' }}>Not configured</span>
                                                )}
                                            </td>
                                            <td style={{ padding: '12px', fontSize: '13px', color: '#4b5563' }}>
                                                {comparison.message || comparison.error || '-'}
                                            </td>
                                            <td style={{ padding: '12px', textAlign: 'center' }}>
                                                {isError ? (
                                                    <span style={{
                                                        display: 'inline-flex',
                                                        alignItems: 'center',
                                                        gap: '4px',
                                                        padding: '4px 10px',
                                                        backgroundColor: '#fee2e2',
                                                        color: '#991b1b',
                                                        borderRadius: '12px',
                                                        fontSize: '12px',
                                                        fontWeight: 500
                                                    }}>
                                                        <XCircle size={14} />
                                                        Error
                                                    </span>
                                                ) : isMatch ? (
                                                    <span style={{
                                                        display: 'inline-flex',
                                                        alignItems: 'center',
                                                        gap: '4px',
                                                        padding: '4px 10px',
                                                        backgroundColor: '#d1fae5',
                                                        color: '#065f46',
                                                        borderRadius: '12px',
                                                        fontSize: '12px',
                                                        fontWeight: 500
                                                    }}>
                                                        <CheckCircle size={14} />
                                                        Match
                                                    </span>
                                                ) : (
                                                    <span style={{
                                                        display: 'inline-flex',
                                                        alignItems: 'center',
                                                        gap: '4px',
                                                        padding: '4px 10px',
                                                        backgroundColor: '#fef3c7',
                                                        color: '#92400e',
                                                        borderRadius: '12px',
                                                        fontSize: '12px',
                                                        fontWeight: 500
                                                    }}>
                                                        <AlertCircle size={14} />
                                                        Mismatch
                                                    </span>
                                                )}
                                            </td>
                                        </tr>
                                        {isExpanded && (
                                            <tr style={{ backgroundColor: index % 2 === 0 ? '#fafafa' : '#f5f5f5' }}>
                                                <td colSpan="5" style={{ padding: '12px' }}>
                                                    <div style={{ 
                                                        backgroundColor: '#1e293b', 
                                                        color: '#e2e8f0', 
                                                        padding: '12px', 
                                                        borderRadius: '4px',
                                                        fontSize: '12px',
                                                        fontFamily: 'monospace',
                                                        overflowX: 'auto',
                                                        maxHeight: '300px',
                                                        overflowY: 'auto'
                                                    }}>
                                                        <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                                                            {formatConfig(comparison.config_raw)}
                                                        </pre>
                                                    </div>
                                                </td>
                                            </tr>
                                        )}
                                    </React.Fragment>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </div>
        );
    };

    const renderValidationResults = () => {
        if (!validationResults) return null;

        return (
            <div style={{ marginTop: '24px' }}>
                <h3 style={{ 
                    fontSize: '18px', 
                    fontWeight: 600, 
                    color: '#111827',
                    marginBottom: '16px',
                    paddingBottom: '12px',
                    borderBottom: '2px solid #e5e7eb'
                }}>
                    Validation Results
                </h3>
                
                {renderSummaryCards()}
                {renderBucketExistenceTable()}
                
                {validationResults.lifecycle && renderConfigTable('lifecycle', validationResults.lifecycle)}
                {validationResults.events && renderConfigTable('events', validationResults.events)}

                {validationResults.error && (
                    <div style={{
                        padding: '16px',
                        backgroundColor: '#fee2e2',
                        border: '1px solid #fecaca',
                        borderRadius: '8px',
                        color: '#991b1b',
                        display: 'flex',
                        alignItems: 'start',
                        gap: '12px'
                    }}>
                        <XCircle size={20} style={{ flexShrink: 0, marginTop: '2px' }} />
                        <div>
                            <strong style={{ display: 'block', marginBottom: '4px' }}>Validation Error</strong>
                            {validationResults.error}
                        </div>
                    </div>
                )}
            </div>
        );
    };

    return (
        <div>
            <div className="card">
                <div className="card-header">
                    <div style={{ display: 'flex', alignItems: 'center' }}>
                        <CheckSquare style={{ width: '20px', height: '20px', marginRight: '8px', color: '#2563eb' }} />
                        <h3 className="card-title">Configuration Validation</h3>
                    </div>
                    <p style={{ margin: '8px 0 0 0', fontSize: '13px', color: 'var(--text-secondary)' }}>
                        Validate bucket lifecycle and event configurations across multiple aliases
                    </p>
                </div>

                <div style={{ padding: '20px' }}>
                    {/* Alias Selection */}
                    <div style={{ marginBottom: '20px' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '8px' }}>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '13px', 
                                color: '#4b5563'
                            }}>
                                Select Aliases
                                <span style={{ color: '#dc2626', marginLeft: '4px' }}>*</span>
                            </label>
                            <button
                                type="button"
                                onClick={handleSelectAllAliases}
                                style={{
                                    padding: '6px 12px',
                                    fontSize: '12px',
                                    fontWeight: 500,
                                    color: '#2563eb',
                                    backgroundColor: 'white',
                                    border: '1px solid #2563eb',
                                    borderRadius: '4px',
                                    cursor: 'pointer',
                                    transition: 'all 0.2s'
                                }}
                            >
                                {selectedAliases.length === aliases.length ? 'Deselect All' : 'Select All'}
                            </button>
                        </div>
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', marginTop: '8px' }}>
                            {aliases.map(alias => {
                                const isSelected = selectedAliases.includes(alias.name);
                                return (
                                    <label
                                        key={alias.name}
                                        style={{ 
                                            display: 'inline-flex',
                                            alignItems: 'center',
                                            padding: '8px 16px',
                                            border: isSelected ? '1px solid #2563eb' : '1px solid #d1d5db',
                                            borderRadius: '6px',
                                            cursor: 'pointer',
                                            backgroundColor: isSelected ? '#eff6ff' : '#ffffff',
                                            transition: '0.2s',
                                            fontSize: '13px'
                                        }}
                                    >
                                        <input
                                            type="checkbox"
                                            checked={isSelected}
                                            onChange={() => handleAliasToggle(alias.name)}
                                            style={{ marginRight: '8px' }}
                                        />
                                        <span style={{ fontWeight: isSelected ? 600 : 400, color: isSelected ? '#2563eb' : '#374151' }}>
                                            {alias.name}
                                        </span>
                                    </label>
                                );
                            })}
                        </div>
                    </div>

                    {/* Bucket Selection */}
                    <div style={{ marginBottom: '20px' }}>
                        <label style={{ 
                            display: 'block', 
                            fontSize: '13px', 
                            color: '#4b5563', 
                            marginBottom: '6px' 
                        }}>
                            Select Bucket
                            <span style={{ color: '#dc2626', marginLeft: '4px' }}>*</span>
                        </label>
                        {isLoadingBuckets ? (
                            <div style={{ padding: '14px', textAlign: 'center', color: '#6b7280', backgroundColor: '#f9fafb', borderRadius: '6px', border: '1px solid #e5e7eb' }}>
                                <Loader className="spin" size={18} style={{ display: 'inline-block', marginRight: '8px' }} />
                                Loading buckets...
                            </div>
                        ) : (
                            <select
                                value={selectedBucket}
                                onChange={(e) => setSelectedBucket(e.target.value)}
                                disabled={selectedAliases.length === 0}
                                required
                                style={{
                                    width: '100%',
                                    padding: '10px',
                                    fontSize: '14px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: '6px',
                                    backgroundColor: selectedAliases.length === 0 ? '#f3f4f6' : 'white',
                                    color: '#111827',
                                    cursor: selectedAliases.length === 0 ? 'not-allowed' : 'pointer'
                                }}
                            >
                                <option value="">Select bucket...</option>
                                {buckets.map(bucket => (
                                    <option key={bucket} value={bucket}>{bucket}</option>
                                ))}
                            </select>
                        )}
                    </div>

                    {/* Configuration Types */}
                    <div style={{ margin: '0 0 20px 0' }}>
                        <label style={{ 
                            display: 'block', 
                            fontSize: '13px', 
                            color: '#4b5563', 
                            marginBottom: '6px' 
                        }}>
                            Configuration Types
                            <span style={{ color: '#dc2626', marginLeft: '4px' }}>*</span>
                        </label>
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', marginTop: '8px' }}>
                            <label style={{ 
                                display: 'inline-flex',
                                alignItems: 'center',
                                padding: '8px 16px',
                                border: checkLifecycle ? '1px solid #2563eb' : '1px solid #d1d5db',
                                borderRadius: '6px',
                                cursor: 'pointer',
                                backgroundColor: checkLifecycle ? '#eff6ff' : '#ffffff',
                                transition: '0.2s',
                                fontSize: '13px'
                            }}>
                                <input
                                    type="checkbox"
                                    checked={checkLifecycle}
                                    onChange={(e) => setCheckLifecycle(e.target.checked)}
                                    style={{ marginRight: '8px' }}
                                />
                                <CheckSquare size={16} style={{ marginRight: '6px', color: checkLifecycle ? '#2563eb' : '#6b7280' }} />
                                <span style={{ fontWeight: checkLifecycle ? 600 : 400, color: checkLifecycle ? '#2563eb' : '#374151' }}>
                                    Bucket Lifecycle
                                </span>
                            </label>
                            <label style={{ 
                                display: 'inline-flex',
                                alignItems: 'center',
                                padding: '8px 16px',
                                border: checkEvents ? '1px solid #2563eb' : '1px solid #d1d5db',
                                borderRadius: '6px',
                                cursor: 'pointer',
                                backgroundColor: checkEvents ? '#eff6ff' : '#ffffff',
                                transition: '0.2s',
                                fontSize: '13px'
                            }}>
                                <input
                                    type="checkbox"
                                    checked={checkEvents}
                                    onChange={(e) => setCheckEvents(e.target.checked)}
                                    style={{ marginRight: '8px' }}
                                />
                                <AlertCircle size={16} style={{ marginRight: '6px', color: checkEvents ? '#2563eb' : '#6b7280' }} />
                                <span style={{ fontWeight: checkEvents ? 600 : 400, color: checkEvents ? '#2563eb' : '#374151' }}>
                                    Event Notifications
                                </span>
                            </label>
                        </div>
                    </div>

                    {/* Run Button */}
                    <button 
                        type="submit"
                        onClick={executeValidation}
                        disabled={isRunning || selectedAliases.length === 0 || !selectedBucket}
                        className="btn btn-primary"
                        style={{ 
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '8px',
                            padding: '12px 18px',
                            fontSize: '15px',
                            width: '100%',
                            justifyContent: 'center',
                            opacity: (isRunning || selectedAliases.length === 0 || !selectedBucket) ? 0.5 : 1,
                            cursor: (isRunning || selectedAliases.length === 0 || !selectedBucket) ? 'not-allowed' : 'pointer'
                        }}
                    >
                        {isRunning ? (
                            <>
                                <Loader className="spin" size={16} />
                                Running Validation...
                            </>
                        ) : (
                            <>
                                <Play size={16} />
                                Validate Configuration
                            </>
                        )}
                    </button>
                </div>
            </div>

            {renderValidationResults()}
        </div>
    );
};

export default ValidateOperations;