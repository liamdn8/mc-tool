import React, { useState, useEffect, useMemo } from 'react';
import { CheckSquare, Play, CheckCircle, AlertCircle, XCircle, Loader, ChevronDown, ChevronRight } from 'lucide-react';
import { apiCall } from '../../utils/api';
import { useContentsPanel } from '../../contexts/ContentsPanelContext';
import ConfigModal from '../validation/ConfigModal';
import OverviewCards from '../validation/OverviewCards';
import BucketExistenceTable from '../validation/BucketExistenceTable';
import ConfigComparisonTable from '../validation/ConfigComparisonTable';
import ValidationNavigation from '../validation/ValidationNavigation';

const ValidateOperations = () => {
    const { setContentsComponent } = useContentsPanel();
    const [aliases, setAliases] = useState([]);
    const [selectedAliases, setSelectedAliases] = useState([]);
    const [buckets, setBuckets] = useState([]);
    const [selectedBuckets, setSelectedBuckets] = useState([]);
    const [bucketSearch, setBucketSearch] = useState('');
    const [checkLifecycle, setCheckLifecycle] = useState(true);
    const [checkEvents, setCheckEvents] = useState(true);
    const [validationResults, setValidationResults] = useState(null);
    const [isRunning, setIsRunning] = useState(false);
    const [isLoadingBuckets, setIsLoadingBuckets] = useState(false);
    const [modalConfig, setModalConfig] = useState(null);

    // Update contents panel when validation results change
    useEffect(() => {
        if (validationResults) {
            setContentsComponent(
                <ValidationNavigation 
                    validationResults={validationResults}
                    checkLifecycle={checkLifecycle}
                    checkEvents={checkEvents}
                    embedded={true}
                />
            );
        } else {
            setContentsComponent(null);
        }

        // Cleanup when component unmounts
        return () => {
            setContentsComponent(null);
        };
    }, [validationResults, checkLifecycle, checkEvents, setContentsComponent]);

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
            setSelectedBuckets([]);
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

    const handleBucketToggle = (bucketName) => {
        setSelectedBuckets(prev => {
            if (prev.includes(bucketName)) {
                return prev.filter(b => b !== bucketName);
            } else {
                return [...prev, bucketName];
            }
        });
    };

    const handleSelectAllBuckets = () => {
        if (selectedBuckets.length === buckets.length) {
            setSelectedBuckets([]);
        } else {
            setSelectedBuckets([...buckets]);
        }
    };

    // Filter buckets based on search
    const filteredBuckets = useMemo(() => {
        if (!bucketSearch.trim()) return buckets;
        const searchLower = bucketSearch.toLowerCase();
        return buckets.filter(bucket => bucket.toLowerCase().includes(searchLower));
    }, [buckets, bucketSearch]);

    const executeValidation = async () => {
        if (selectedAliases.length === 0) {
            alert('Please select at least one alias');
            return;
        }

        if (selectedBuckets.length === 0) {
            alert('Please select at least one bucket');
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
                    buckets: selectedBuckets,
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
    const calculateTableSeverity = (table, buckets, aliases) => {
        if (!table || table.length === 0) return 'info';
        
        let totalCells = 0;
        let matchCells = 0;
        
        table.forEach(row => {
            aliases.forEach(alias => {
                const cell = row[alias];
                if (cell && cell.status !== 'not_exist') {
                    totalCells++;
                    if (cell.status === 'match') matchCells++;
                }
            });
        });
        
        if (totalCells === 0) return 'info';
        if (matchCells === totalCells) return 'success';
        if (matchCells >= totalCells / 2) return 'warning';
        return 'danger';
    };

    const openConfigModal = (bucket, alias, config, configType) => {
        // Always use the first selected alias as the reference site
        let referenceConfig = null;
        let referenceAlias = null;
        
        if (validationResults && configType) {
            const table = validationResults[configType];
            if (table) {
                const row = table.find(r => r.bucket === bucket);
                if (row) {
                    // Get the first alias from validation results (this is the first selected alias)
                    const firstAlias = validationResults.aliases && validationResults.aliases.length > 0 
                        ? validationResults.aliases[0] 
                        : null;
                    
                    if (firstAlias) {
                        const cell = row[firstAlias];
                        if (cell && cell.status !== 'not_exist' && cell.status !== 'not_configured' && cell.value) {
                            referenceConfig = cell.value;
                            referenceAlias = firstAlias;
                        }
                    }
                }
            }
        }
        
        setModalConfig({ 
            bucket, 
            alias, 
            value: config,
            referenceValue: referenceConfig,
            referenceAlias: referenceAlias
        });
    };

    const closeModal = () => {
        setModalConfig(null);
    };

    const renderValidationResults = () => {
        if (!validationResults) return null;

        return (
            <div style={{ marginTop: '24px' }}>
                <h3 className="card-title" style={{ 
                    marginBottom: '16px', 
                    paddingBottom: '12px', 
                    borderBottom: '2px solid var(--border-color)' 
                }}>
                    Validation Results
                </h3>
                
                <OverviewCards 
                    validationResults={validationResults}
                    checkLifecycle={checkLifecycle}
                    checkEvents={checkEvents}
                    calculateTableSeverity={calculateTableSeverity}
                />
                
                <BucketExistenceTable validationResults={validationResults} />
                
                {validationResults.lifecycle_table && (
                    <ConfigComparisonTable 
                        configType="lifecycle_table"
                        configTable={validationResults.lifecycle_table}
                        validationResults={validationResults}
                        calculateTableSeverity={calculateTableSeverity}
                        onViewConfig={(bucket, alias, config) => openConfigModal(bucket, alias, config, 'lifecycle_table')}
                    />
                )}
                
                {validationResults.events_table && (
                    <ConfigComparisonTable 
                        configType="events_table"
                        configTable={validationResults.events_table}
                        validationResults={validationResults}
                        calculateTableSeverity={calculateTableSeverity}
                        onViewConfig={(bucket, alias, config) => openConfigModal(bucket, alias, config, 'events_table')}
                    />
                )}

                {validationResults.error && (
                    <div className="alert alert-danger">
                        <XCircle size={20} style={{ marginRight: '12px', display: 'inline-block', verticalAlign: 'middle' }} />
                        <div style={{ display: 'inline-block' }}>
                            <strong>Validation Error</strong>
                            <div>{validationResults.error}</div>
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
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '8px' }}>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '13px', 
                                color: '#4b5563'
                            }}>
                                Select Buckets
                                <span style={{ color: '#dc2626', marginLeft: '4px' }}>*</span>
                            </label>
                            <button
                                type="button"
                                onClick={handleSelectAllBuckets}
                                style={{
                                    padding: '6px 12px',
                                    fontSize: '12px',
                                    backgroundColor: 'white',
                                    color: '#2563eb',
                                    border: '1px solid #2563eb',
                                    borderRadius: '4px',
                                    cursor: 'pointer',
                                    fontWeight: 500
                                }}
                            >
                                {selectedBuckets.length === buckets.length ? 'Deselect All' : 'Select All'}
                            </button>
                        </div>
                        <input
                            type="text"
                            placeholder="Search buckets..."
                            value={bucketSearch}
                            onChange={(e) => setBucketSearch(e.target.value)}
                            disabled={buckets.length === 0 || selectedAliases.length === 0}
                            style={{
                                width: '100%',
                                padding: '8px 12px',
                                fontSize: '14px',
                                border: '1px solid #d1d5db',
                                borderRadius: '6px',
                                marginBottom: '8px',
                                backgroundColor: (buckets.length === 0 || selectedAliases.length === 0) ? '#f3f4f6' : 'white'
                            }}
                        />
                        {isLoadingBuckets ? (
                            <div style={{ padding: '14px', textAlign: 'center', color: '#6b7280', backgroundColor: '#f9fafb', borderRadius: '6px', border: '1px solid #e5e7eb' }}>
                                <Loader className="spin" size={18} style={{ display: 'inline-block', marginRight: '8px' }} />
                                Loading buckets...
                            </div>
                        ) : (
                            <div style={{
                                maxHeight: '200px',
                                overflowY: 'auto',
                                border: '1px solid #d1d5db',
                                borderRadius: '6px',
                                backgroundColor: selectedAliases.length === 0 ? '#f3f4f6' : 'white',
                                padding: '8px'
                            }}>
                                {buckets.length === 0 ? (
                                    <div style={{ padding: '12px', textAlign: 'center', color: '#6b7280' }}>
                                        {selectedAliases.length === 0 ? 'Select aliases first' : 'No buckets found'}
                                    </div>
                                ) : filteredBuckets.length === 0 ? (
                                    <div style={{ padding: '12px', textAlign: 'center', color: '#6b7280' }}>
                                        No buckets match "{bucketSearch}"
                                    </div>
                                ) : (
                                    filteredBuckets.map(bucket => (
                                        <label
                                            key={bucket}
                                            style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                padding: '8px',
                                                cursor: selectedAliases.length === 0 ? 'not-allowed' : 'pointer',
                                                borderRadius: '4px',
                                                backgroundColor: selectedBuckets.includes(bucket) ? '#eff6ff' : 'transparent',
                                                opacity: selectedAliases.length === 0 ? 0.5 : 1
                                            }}
                                            onMouseEnter={(e) => {
                                                if (selectedAliases.length > 0 && !selectedBuckets.includes(bucket)) {
                                                    e.currentTarget.style.backgroundColor = '#f9fafb';
                                                }
                                            }}
                                            onMouseLeave={(e) => {
                                                if (!selectedBuckets.includes(bucket)) {
                                                    e.currentTarget.style.backgroundColor = 'transparent';
                                                }
                                            }}
                                        >
                                            <input
                                                type="checkbox"
                                                checked={selectedBuckets.includes(bucket)}
                                                onChange={() => handleBucketToggle(bucket)}
                                                disabled={selectedAliases.length === 0}
                                                style={{ marginRight: '8px', cursor: selectedAliases.length === 0 ? 'not-allowed' : 'pointer' }}
                                            />
                                            <span style={{ fontSize: '14px', color: '#111827', fontFamily: 'monospace' }}>
                                                {bucket}
                                            </span>
                                        </label>
                                    ))
                                )}
                            </div>
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
                        disabled={isRunning || selectedAliases.length === 0 || selectedBuckets.length === 0}
                        className="btn btn-primary"
                        style={{ 
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '8px',
                            padding: '12px 18px',
                            fontSize: '15px',
                            width: '100%',
                            justifyContent: 'center',
                            opacity: (isRunning || selectedAliases.length === 0 || selectedBuckets.length === 0) ? 0.5 : 1,
                            cursor: (isRunning || selectedAliases.length === 0 || selectedBuckets.length === 0) ? 'not-allowed' : 'pointer'
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
            
            {modalConfig && (
                <ConfigModal 
                    config={modalConfig} 
                    onClose={closeModal} 
                />
            )}
        </div>
    );
};

export default ValidateOperations;