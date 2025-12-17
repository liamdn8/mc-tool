import React, { useState, useEffect, useMemo } from 'react';
import { Activity, Play, Copy, CheckCircle, Terminal, BarChart3, ChevronUp, ChevronDown, ChevronLeft, ChevronRight, Flame, HardDrive, Lock, Shield, Workflow, TrendingUp } from 'lucide-react';
import { runProfileCapture } from '../../utils/api';
import { useI18n } from '../../utils/i18n';
import ErrorAlert from '../ErrorAlert';

const DURATION_OPTIONS = ['5s', '10s', '15s', '30s', '45s', '1m', '2m', '3m', '5m'];
const PROFILE_TYPE_OPTIONS = [
    { value: 'cpu', label: 'CPU', icon: Flame },
    { value: 'mem', label: 'Memory', icon: HardDrive },
    { value: 'block', label: 'Block', icon: Lock },
    { value: 'mutex', label: 'Mutex', icon: Shield },
    { value: 'goroutines', label: 'Goroutines', icon: Workflow }
];

const ProfileOperations = ({ sites = [] }) => {
    const { t } = useI18n();
    const [form, setForm] = useState({
        alias: '',
        duration: '5s',
        profileTypes: ['cpu', 'mem'],
        insecure: false
    });
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState(null);
    const [error, setError] = useState(null);
    const [copiedCommand, setCopiedCommand] = useState(null);
    const [activeTab, setActiveTab] = useState('cpu');
    const [currentPage, setCurrentPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [sortConfig, setSortConfig] = useState({ key: 'rank', direction: 'asc' });

    useEffect(() => {
        if (!form.alias && sites.length > 0) {
            setForm(prev => ({ ...prev, alias: sites[0].name || sites[0].alias || '' }));
        }
    }, [sites]);

    const handleProfileTypeToggle = (type) => {
        setForm(prev => {
            const types = prev.profileTypes.includes(type)
                ? prev.profileTypes.filter(t => t !== type)
                : [...prev.profileTypes, type];
            return { ...prev, profileTypes: types };
        });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!form.alias || form.profileTypes.length === 0) return;

        setLoading(true);
        setError(null);
        setResult(null);

        try {
            const response = await runProfileCapture({
                alias: form.alias.trim(),
                duration: form.duration,
                profileType: form.profileTypes.join(','),
                insecure: form.insecure
            });
            setResult(response);

            // Set active tab to first available profile type
            if (response.analysis) {
                const firstType = Object.keys(response.analysis)[0];
                if (firstType) {
                    setActiveTab(firstType);
                }
            }
        } catch (err) {
            setError(err.message || 'Profile capture failed');
        } finally {
            setLoading(false);
        }
    };

    const formatValue = (value, unit) => {
        if (unit === 'bytes') {
            if (value >= 1024 * 1024 * 1024) {
                return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`;
            }
            if (value >= 1024 * 1024) {
                return `${(value / (1024 * 1024)).toFixed(2)} MB`;
            }
            if (value >= 1024) {
                return `${(value / 1024).toFixed(2)} KB`;
            }
            return `${value} B`;
        }
        return `${value}${unit}`;
    };

    const copyCommand = (command) => {
        navigator.clipboard.writeText(command);
        setCopiedCommand(command);
        setTimeout(() => setCopiedCommand(null), 2000);
    };

    const currentAnalysis = useMemo(() => {
        if (!result || !result.analysis || !activeTab) return null;
        return result.analysis[activeTab];
    }, [result, activeTab]);

    const sortedFunctions = useMemo(() => {
        if (!currentAnalysis || !currentAnalysis.functions) return [];
        
        const sorted = [...currentAnalysis.functions];
        sorted.sort((a, b) => {
            let aVal = a[sortConfig.key];
            let bVal = b[sortConfig.key];
            
            // For string comparison (function name)
            if (typeof aVal === 'string') {
                return sortConfig.direction === 'asc' 
                    ? aVal.localeCompare(bVal)
                    : bVal.localeCompare(aVal);
            }
            
            // For numeric comparison
            if (sortConfig.direction === 'asc') {
                return aVal - bVal;
            }
            return bVal - aVal;
        });
        
        return sorted;
    }, [currentAnalysis, sortConfig]);

    const paginatedFunctions = useMemo(() => {
        const startIndex = (currentPage - 1) * pageSize;
        const endIndex = startIndex + pageSize;
        return sortedFunctions.slice(startIndex, endIndex);
    }, [sortedFunctions, currentPage, pageSize]);

    const totalPages = useMemo(() => {
        return Math.ceil(sortedFunctions.length / pageSize);
    }, [sortedFunctions.length, pageSize]);

    const handleSort = (key) => {
        setSortConfig(prev => ({
            key,
            direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
        }));
        setCurrentPage(1); // Reset to first page on sort
    };

    const handlePageChange = (newPage) => {
        setCurrentPage(Math.max(1, Math.min(newPage, totalPages)));
    };

    const renderBarChart = (analysis) => {
        if (!analysis || !analysis.functions || analysis.functions.length === 0) return null;

        const top10 = analysis.functions.slice(0, 10);
        const maxValue = Math.max(...top10.map(f => f.cumPct));

        return (
            <div className="p-5 bg-gray-50 rounded-lg border">
                <h4 className="mb-4 text-sm font-semibold text-primary">
                    {t('profile_top_functions', 'Top 10 Functions by Cumulative %')}
                </h4>
                <div className="flex flex-col gap-3">
                    {top10.map((item, idx) => (
                        <div key={idx} className="flex items-center gap-3">
                            <div className="text-xs font-semibold text-secondary text-right" style={{ minWidth: '30px' }}>
                                #{item.rank}
                            </div>
                            <div className="flex-1 min-w-0">
                                <div className="flex items-center mb-1 text-xs gap-2">
                                    <code className="flex-1 overflow-hidden text-primary" style={{ 
                                        textOverflow: 'ellipsis',
                                        whiteSpace: 'nowrap',
                                        fontSize: '11px'
                                    }}>
                                        {item.name}
                                    </code>
                                    <span className="font-semibold text-info text-right" style={{ minWidth: '60px' }}>
                                        {item.cumPct.toFixed(2)}%
                                    </span>
                                </div>
                                <div className="bg-secondary rounded-sm overflow-hidden" style={{ height: '6px' }}>
                                    <div style={{ 
                                        height: '100%',
                                        width: `${(item.cumPct / maxValue) * 100}%`,
                                        backgroundColor: `hsl(${220 - idx * 8}, 70%, ${50 + idx * 2}%)`,
                                        transition: 'width 0.3s ease'
                                    }} />
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        );
    };

    const renderProfileTable = (analysis) => {
        if (!analysis || !analysis.functions || analysis.functions.length === 0) {
            return (
                <div className="empty-state">
                    No profile data available
                </div>
            );
        }

        const SortableHeader = ({ column, label, align = 'left' }) => (
            <th 
                onClick={() => handleSort(column)}
                className="p-3 text-xs font-normal text-secondary"
                style={{ 
                    textAlign: align,
                    cursor: 'pointer',
                    userSelect: 'none',
                    whiteSpace: 'nowrap'
                }}
            >
                <div style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                    {label}
                    {sortConfig.key === column && (
                        sortConfig.direction === 'asc' 
                            ? <ChevronUp style={{ width: '14px', height: '14px' }} />
                            : <ChevronDown style={{ width: '14px', height: '14px' }} />
                    )}
                </div>
            </th>
        );

        return (
            <>
                {/* Pagination Controls - Top */}
                {totalPages > 1 && (
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
                            {((currentPage - 1) * pageSize) + 1}-{Math.min(currentPage * pageSize, sortedFunctions.length)} of {sortedFunctions.length}
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

                <div style={{ border: '1px solid #d1d5db', borderRadius: '6px', overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '640px' }}>
                        <thead style={{ backgroundColor: '#f9fafb' }}>
                            <tr>
                                <SortableHeader column="rank" label={t('profile_field_rank', 'Rank')} align="center" />
                                <SortableHeader column="name" label={t('profile_field_function', 'Function')} align="left" />
                                <th 
                                    onClick={() => handleSort('flat')}
                                    style={{ 
                                        padding: '12px', 
                                        textAlign: 'left', 
                                        fontSize: '12px', 
                                        fontWeight: '400',
                                        color: '#6b7280',
                                        cursor: 'pointer',
                                        userSelect: 'none',
                                        whiteSpace: 'nowrap'
                                    }}
                                    title={t('profile_help_flat', 'Time/memory spent in this function only (excluding functions it calls)')}
                                >
                                    <div style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                                        {t('profile_field_flat', 'Flat')}
                                        {sortConfig.key === 'flat' && (
                                            sortConfig.direction === 'asc' 
                                                ? <ChevronUp style={{ width: '14px', height: '14px' }} />
                                                : <ChevronDown style={{ width: '14px', height: '14px' }} />
                                        )}
                                    </div>
                                </th>
                                <SortableHeader column="flatPct" label={t('profile_field_flat_pct', 'Flat %')} align="left" />
                                <th 
                                    onClick={() => handleSort('cum')}
                                    style={{ 
                                        padding: '12px', 
                                        textAlign: 'left', 
                                        fontSize: '12px', 
                                        fontWeight: '400',
                                        color: '#6b7280',
                                        cursor: 'pointer',
                                        userSelect: 'none',
                                        whiteSpace: 'nowrap'
                                    }}
                                    title={t('profile_help_cumulative', 'Total time/memory including all functions called by this function')}
                                >
                                    <div style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                                        {t('profile_field_cumulative', 'Cumulative')}
                                        {sortConfig.key === 'cum' && (
                                            sortConfig.direction === 'asc' 
                                                ? <ChevronUp style={{ width: '14px', height: '14px' }} />
                                                : <ChevronDown style={{ width: '14px', height: '14px' }} />
                                        )}
                                    </div>
                                </th>
                                <SortableHeader column="cumPct" label={t('profile_field_cum_pct', 'Cum %')} align="left" />
                            </tr>
                        </thead>
                        <tbody>
                            {paginatedFunctions.map((item, idx) => (
                                <tr key={idx} style={{ borderTop: '1px solid #f3f4f6' }}>
                                    <td style={{ 
                                        padding: '12px', 
                                        textAlign: 'center', 
                                        fontWeight: '600',
                                        fontSize: '13px',
                                        color: '#6b7280',
                                        whiteSpace: 'nowrap'
                                    }}>
                                        #{item.rank}
                                    </td>
                                    <td style={{ padding: '12px', fontSize: '12px', color: '#4b5563', wordBreak: 'break-all' }}>
                                        <code style={{ 
                                            fontSize: '12px', 
                                            color: '#111827',
                                            fontFamily: 'monospace'
                                        }}>
                                            {item.name}
                                        </code>
                                    </td>
                                    <td style={{ 
                                        padding: '12px', 
                                        fontSize: '13px',
                                        color: '#111827',
                                        whiteSpace: 'nowrap'
                                    }}>
                                        {formatValue(item.flat, analysis.unit)}
                                    </td>
                                    <td style={{ 
                                        padding: '12px', 
                                        fontSize: '13px',
                                        color: '#111827',
                                        whiteSpace: 'nowrap'
                                    }}>
                                        {item.flatPct.toFixed(2)}%
                                    </td>
                                    <td style={{ 
                                        padding: '12px', 
                                        fontSize: '13px',
                                        color: '#111827',
                                        whiteSpace: 'nowrap'
                                    }}>
                                        {formatValue(item.cum, analysis.unit)}
                                    </td>
                                    <td style={{ 
                                        padding: '12px', 
                                        fontSize: '13px',
                                        color: '#111827',
                                        whiteSpace: 'nowrap'
                                    }}>
                                        {item.cumPct.toFixed(2)}%
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>

                    {/* Pagination Controls - Bottom (inside table border) */}
                    {totalPages > 1 && (
                        <div style={{ 
                            display: 'flex', 
                            flexWrap: 'wrap',
                            justifyContent: 'space-between', 
                            alignItems: 'center',
                            gap: '12px',
                            padding: '12px',
                            borderTop: '1px solid #d1d5db'
                        }}>
                            <span style={{ fontSize: '12px', color: '#4b5563' }}>
                                {((currentPage - 1) * pageSize) + 1}-{Math.min(currentPage * pageSize, sortedFunctions.length)} of {sortedFunctions.length}
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
                </div>
            </>
        );
    };

    return (
        <div>
            {/* Form Card */}
            <div className="card">
                <div className="card-header">
                    <div style={{ display: 'flex', alignItems: 'center' }}>
                        <Activity style={{ width: '20px', height: '20px', marginRight: '8px', color: '#2563eb' }} />
                        <h3 className="card-title">
                            {t('operations_profile_title', 'MinIO Performance Profiling')}
                        </h3>
                    </div>
                    <p style={{ margin: '8px 0 0 0', fontSize: '13px', color: 'var(--text-secondary)' }}>
                        {t('operations_profile_description', 'Capture and analyze MinIO performance profiles with automatic pprof analysis')}
                    </p>
                </div>

                <form onSubmit={handleSubmit} style={{ padding: '20px' }}>
                    {/* Single Row Layout */}
                    <div style={{ 
                        display: 'grid', 
                        gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
                        gap: '16px',
                        marginBottom: '16px'
                    }}>
                        <div>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '13px', 
                                color: '#4b5563', 
                                marginBottom: '6px' 
                            }}>
                                {t('trace_filter_alias', 'Alias')}
                                <span style={{ color: '#dc2626', marginLeft: '4px' }}>*</span>
                            </label>
                            <select
                                value={form.alias}
                                onChange={(e) => setForm(prev => ({ ...prev, alias: e.target.value }))}
                                style={{ 
                                    width: '100%', 
                                    padding: '10px', 
                                    borderRadius: '6px', 
                                    border: '1px solid #d1d5db', 
                                    fontSize: '14px' 
                                }}
                                required
                            >
                                <option value="">Select alias...</option>
                                {sites.map(site => (
                                    <option key={site.alias || site.name} value={site.alias || site.name}>
                                        {site.alias || site.name}
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label style={{ 
                                display: 'block', 
                                fontSize: '13px', 
                                color: '#4b5563', 
                                marginBottom: '6px' 
                            }}>
                                {t('trace_filter_duration', 'Duration')}
                            </label>
                            <select
                                value={form.duration}
                                onChange={(e) => setForm(prev => ({ ...prev, duration: e.target.value }))}
                                style={{ 
                                    width: '100%', 
                                    padding: '10px', 
                                    borderRadius: '6px', 
                                    border: '1px solid #d1d5db', 
                                    fontSize: '14px' 
                                }}
                            >
                                {DURATION_OPTIONS.map(dur => (
                                    <option key={dur} value={dur}>{dur}</option>
                                ))}
                            </select>
                        </div>
                    </div>

                    {/* Profile Types */}
                    <div style={{ margin: '0 0 20px 0' }}>
                        <label style={{ 
                            display: 'block', 
                            fontSize: '13px', 
                            color: '#4b5563', 
                            marginBottom: '6px' 
                        }}>
                            {t('profile_types', 'Profile Types')}
                            <span style={{ color: '#dc2626', marginLeft: '4px' }}>*</span>
                        </label>
                        <div style={{ 
                            display: 'flex', 
                            flexWrap: 'wrap', 
                            gap: '8px',
                            marginTop: '8px'
                        }}>
                            {PROFILE_TYPE_OPTIONS.map(type => {
                                const isSelected = form.profileTypes.includes(type.value);
                                const IconComponent = type.icon;
                                return (
                                    <label
                                        key={type.value}
                                        style={{
                                            display: 'inline-flex',
                                            alignItems: 'center',
                                            padding: '8px 16px',
                                            border: `1px solid ${isSelected ? '#2563eb' : '#d1d5db'}`,
                                            borderRadius: '6px',
                                            cursor: 'pointer',
                                            backgroundColor: isSelected ? '#eff6ff' : '#ffffff',
                                            transition: 'all 0.2s',
                                            fontSize: '13px'
                                        }}
                                    >
                                        <input
                                            type="checkbox"
                                            checked={isSelected}
                                            onChange={() => handleProfileTypeToggle(type.value)}
                                            style={{ marginRight: '8px' }}
                                        />
                                        <IconComponent 
                                            size={16} 
                                            style={{ 
                                                marginRight: '6px',
                                                color: isSelected ? '#2563eb' : '#6b7280'
                                            }} 
                                        />
                                        <span style={{ 
                                            fontWeight: isSelected ? '600' : '400',
                                            color: isSelected ? '#2563eb' : '#374151'
                                        }}>
                                            {type.label}
                                        </span>
                                    </label>
                                );
                            })}
                        </div>
                    </div>

                    <div style={{ margin: '0 0 20px 0' }}>
                        <label style={{
                            display: 'flex',
                            alignItems: 'center',
                            padding: '12px',
                            border: '1px solid #d1d5db',
                            borderRadius: '6px',
                            cursor: 'pointer',
                            border: form.insecure ? `1px solid #FF9800` : '1px solid #d1d5db',
                            backgroundColor: form.insecure ? '#fdfbf8ff' : '#ffffff',
                            fontSize: '13px',
                            gap: '8px'
                        }}>
                            <input
                                type="checkbox"
                                checked={form.insecure}
                                onChange={(e) => setForm(prev => ({ ...prev, insecure: e.target.checked }))}
                            />
                            <span>{t('insecure_option', 'Skip TLS certificate verification (--insecure)')}</span>
                        </label>
                    </div>

                    <button
                        type="submit"
                        disabled={loading || !form.alias || form.profileTypes.length === 0}
                        className="btn btn-primary"
                        style={{ 
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '8px',
                            padding: '12px 18px',
                            fontSize: '15px',
                            width: '100%',
                            justifyContent: 'center'
                        }}
                    >
                        {loading ? (
                            <>
                                <div className="spinner" style={{ width: '16px', height: '16px' }}></div>
                                {t('trace_running', 'Capturing profile...')}
                            </>
                        ) : (
                            <>
                                <Play style={{ width: '16px', height: '16px' }} />
                                {t('profile_start', 'Start Profile & Analyze')}
                            </>
                        )}
                    </button>
                </form>
            </div>

            {error && <ErrorAlert message={error} onClose={() => setError(null)} />}

            {result && result.success && (
                <>
                    <div style={{ padding: '20px', marginTop: '24px', backgroundColor: 'white', borderRadius: '8px', border: '1px solid #e5e7eb' }}>
                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', marginBottom: '16px' }}>
                            <span className="badge badge-neutral">Profile Captured</span>
                            <span className="badge badge-neutral">Alias: {result.alias}</span>
                        </div>



                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))', gap: '16px' }}>
                            {currentAnalysis && (
                                <>
                                    <div style={{ border: '1px solid #e5e7eb', borderRadius: '8px', padding: '16px', backgroundColor: 'white' }}>
                                        <div style={{ fontSize: '26px', fontWeight: '600', color: '#2563eb' }}>
                                            {currentAnalysis.functions?.length || 0}
                                        </div>
                                        <div style={{ marginTop: '4px', fontSize: '13px', color: '#6b7280' }}>{t('profile_functions_analyzed', 'Functions Analyzed')}</div>
                                    </div>

                                    <div style={{ border: '1px solid #e5e7eb', borderRadius: '8px', padding: '16px', backgroundColor: 'white' }}>
                                        <div style={{ fontSize: '26px', fontWeight: '600', color: '#9333ea' }}>
                                            {result.duration || 0}
                                        </div>
                                        <div style={{ marginTop: '4px', fontSize: '13px', color: '#6b7280' }}>Duration</div>
                                    </div>

                                    {result.files && (
                                        <div style={{ border: '1px solid #e5e7eb', borderRadius: '8px', padding: '16px', backgroundColor: 'white' }}>
                                            <div style={{ fontSize: '26px', fontWeight: '600', color: '#059669' }}>
                                                {(result.files.cpu?.length || 0) + (result.files.mem?.length || 0) + (result.files.other?.length || 0)}
                                            </div>
                                            <div style={{ marginTop: '4px', fontSize: '13px', color: '#6b7280' }}>{t('profile_files', 'Profile Files')}</div>
                                        </div>
                                    )}

                                    <div style={{ border: '1px solid #e5e7eb', borderRadius: '8px', padding: '16px', backgroundColor: 'white' }}>
                                        <div style={{ fontSize: '26px', fontWeight: '600', color: '#dc2626' }}>
                                            {Object.keys(result.analysis || {}).length}
                                        </div>
                                        <div style={{ marginTop: '4px', fontSize: '13px', color: '#6b7280' }}>{t('profile_types_count', 'Profile Types')}</div>
                                    </div>
                                </>
                            )}
                        </div>
                    </div>

                    {/* Explanation Card */}
                    <div className="card" style={{ position: 'relative', overflow: 'hidden', marginTop: '24px' }}>
                        <div style={{ 
                            position: 'absolute',
                            top: 0,
                            left: 0,
                            width: '4px',
                            height: '100%',
                            backgroundColor: '#9333ea'
                        }} />
                        <div style={{ padding: '16px' }}>
                            <div style={{ display: 'flex', alignItems: 'flex-start', gap: '16px' }}>
                                <div style={{
                                    width: '48px',
                                    height: '48px',
                                    borderRadius: '10px',
                                    backgroundColor: 'rgba(147, 51, 234, 0.082)',
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    flexShrink: 0
                                }}>
                                    <Activity style={{ color: '#9333ea' }} size={24} />
                                </div>
                                <div style={{ flex: 1 }}>
                                    <h3 style={{ margin: '0 0 4px 0', fontSize: '18px', fontWeight: '600' }}>
                                        {t('profile_explanation_title', 'Understanding Profile Metrics')}
                                    </h3>
                                    <ul style={{ margin: '8px 0 0 0', paddingLeft: '16px', color: 'var(--text-secondary)' }}>
                                        <li style={{ marginBottom: '2px', fontSize: '13px' }}>
                                            {t('profile_explanation_flat', 'Flat = function\'s own execution time, excluding child functions.')}
                                        </li>
                                        <li style={{ marginBottom: '2px', fontSize: '13px' }}>
                                            {t('profile_explanation_cumulative', 'Cumulative = total time, including all descendants.')}
                                        </li>
                                        <li style={{ marginBottom: '2px', fontSize: '13px' }}>
                                            {t('profile_explanation_use_flat', 'Flat helps find functions consuming CPU directly.')}
                                        </li>
                                        <li style={{ marginBottom: '2px', fontSize: '13px' }}>
                                            {t('profile_explanation_use_cumulative', 'Cumulative helps find call chains causing overload.')}
                                        </li>
                                    </ul>
                                </div>
                            </div>
                        </div>
                    </div>

                    {result.analysis && Object.keys(result.analysis).length > 0 && (
                        <div className="card" style={{ marginTop: '24px' }}>
                            <div className="card-header">
                                <div style={{ display: 'flex', alignItems: 'center' }}>
                                    <BarChart3 style={{ width: '20px', height: '20px', marginRight: '8px', color: '#2563eb' }} />
                                    <h3 className="card-title">Profile Analysis</h3>
                                </div>
                            </div>

                            <div style={{ 
                                borderBottom: '2px solid #e5e7eb',
                                padding: '0 20px',
                                display: 'flex',
                                gap: '4px',
                                flexWrap: 'wrap'
                            }}>
                                {Object.keys(result.analysis).map(type => {
                                    const typeConfig = PROFILE_TYPE_OPTIONS.find(opt => opt.value === type) || { icon: TrendingUp, label: type };
                                    const IconComponent = typeConfig.icon;
                                    return (
                                        <button
                                            key={type}
                                            onClick={() => {
                                                setActiveTab(type);
                                                setCurrentPage(1);
                                            }}
                                            style={{
                                                padding: '12px 20px',
                                                border: 'none',
                                                background: 'none',
                                                borderBottom: activeTab === type ? '2px solid #2563eb' : '2px solid transparent',
                                                color: activeTab === type ? '#2563eb' : '#6b7280',
                                                fontWeight: activeTab === type ? '600' : '400',
                                                cursor: 'pointer',
                                                fontSize: '13px',
                                                transition: 'all 0.2s',
                                                marginBottom: '-2px',
                                                display: 'inline-flex',
                                                alignItems: 'center',
                                                gap: '6px'
                                            }}
                                        >
                                            <IconComponent size={16} />
                                            <span>{typeConfig.label}</span>
                                        </button>
                                    );
                                })}
                            </div>

                            <div style={{ padding: '20px' }}>
                                <div style={{ marginBottom: '24px' }}>
                                    {renderBarChart(currentAnalysis)}
                                </div>

                                <div>
                                    <h4 style={{ fontSize: '16px', fontWeight: '600', color: '#1f2937', marginBottom: '8px' }}>
                                        {t('profile_table_title', 'Profile Functions ({count} total)').replace('{count}', sortedFunctions.length)}
                                    </h4>
                                    <p style={{ fontSize: '12px', color: '#6b7280', marginBottom: '12px' }}>
                                        {t('profile_table_help', 'Click column headers to sort. Flat = time/memory in this function only. Cumulative = including called functions.')}
                                    </p>
                                    {renderProfileTable(currentAnalysis)}
                                </div>
                            </div>
                        </div>
                    )}

                    {result.commands && Object.keys(result.commands).length > 0 && (
                        <div className="card" style={{ marginTop: '24px' }}>
                            <div className="card-header">
                                <div style={{ display: 'flex', alignItems: 'center' }}>
                                    <Terminal style={{ width: '20px', height: '20px', marginRight: '8px', color: '#059669' }} />
                                    <h3 className="card-title">{t('profile_commands_title', 'Advanced Analysis Commands')}</h3>
                                </div>
                                <p style={{ margin: '8px 0 0 0', fontSize: '13px', color: 'var(--text-secondary)' }}>
                                    {t('profile_commands_description', 'For deeper analysis, copy and run these pprof commands')}
                                </p>
                            </div>

                            <div style={{ padding: '20px' }}>
                                {Object.entries(result.commands).map(([type, commands]) => {
                                    if (!commands || commands.length === 0) return null;
                                    const typeConfig = PROFILE_TYPE_OPTIONS.find(opt => opt.value === type) || { icon: TrendingUp, label: type };
                                    const IconComponent = typeConfig.icon;

                                    return (
                                        <div key={type} style={{ marginBottom: '20px' }}>
                                            <h4 style={{ 
                                                fontSize: '13px', 
                                                fontWeight: '600',
                                                marginBottom: '10px',
                                                color: '#374151',
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: '6px'
                                            }}>
                                                <IconComponent size={16} />
                                                <span>{typeConfig.label} Profile</span>
                                            </h4>
                                            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                                {commands.slice(0, 3).map((cmd, idx) => (
                                                    <div 
                                                        key={idx}
                                                        style={{
                                                            backgroundColor: '#f9fafb',
                                                            border: '1px solid #e5e7eb',
                                                            borderRadius: '6px',
                                                            padding: '10px 12px'
                                                        }}
                                                    >
                                                        <div style={{ 
                                                            display: 'flex', 
                                                            alignItems: 'center',
                                                            justifyContent: 'space-between',
                                                            marginBottom: '6px'
                                                        }}>
                                                            <span style={{ 
                                                                fontSize: '11px',
                                                                fontWeight: '500',
                                                                color: '#6b7280'
                                                            }}>
                                                                {cmd.label}
                                                            </span>
                                                            <button
                                                                onClick={() => copyCommand(cmd.command)}
                                                                style={{
                                                                    border: 'none',
                                                                    background: 'none',
                                                                    cursor: 'pointer',
                                                                    padding: '2px',
                                                                    display: 'flex',
                                                                    alignItems: 'center',
                                                                    color: copiedCommand === cmd.command ? '#059669' : '#6b7280'
                                                                }}
                                                            >
                                                                {copiedCommand === cmd.command ? (
                                                                    <CheckCircle style={{ width: '14px', height: '14px' }} />
                                                                ) : (
                                                                    <Copy style={{ width: '14px', height: '14px' }} />
                                                                )}
                                                            </button>
                                                        </div>
                                                        <code style={{
                                                            fontSize: '10px',
                                                            color: '#374151',
                                                            display: 'block',
                                                            overflowX: 'auto',
                                                            whiteSpace: 'nowrap',
                                                            fontFamily: 'monospace'
                                                        }}>
                                                            {cmd.command}
                                                        </code>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    )}
                </>
            )}
        </div>
    );
};

export default ProfileOperations;
