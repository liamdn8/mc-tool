import React, { useState, useMemo, useEffect } from 'react';
import { Activity, Play, Filter, AlertCircle, BarChart3 } from 'lucide-react';
import { runTraceCapture } from '../../utils/api';
import { useI18n } from '../../utils/i18n';

const MAX_TABLE_ROWS = 100;
const PAGE_SIZE_OPTIONS = [20, 50, 100];

const TIME_FIELD_CANDIDATES = ['time', 'Time', 'timestamp', 'Timestamp', 'eventTime', 'ts', 'trace.time', 'req.time', 'loggedAt'];
const API_FIELD_CANDIDATES = ['api', 'API', 'action', 'Action', 'method', 'operation', 'req.api', 'req.method', 'req.action', 'trace.api'];
const STATUS_FIELD_CANDIDATES = ['statusCode', 'status', 'resp.status', 'resp.statusCode', 'trace.statusCode'];
const CLIENT_FIELD_CANDIDATES = ['client', 'clientHost', 'clientAddr', 'remoteHost', 'remotehost', 'remoteAddr', 'host', 'Host', 'source', 'req.client', 'req.remoteHost', 'req.remoteAddr', 'trace.remoteHost', 'trace.client', 'trace.remoteAddr'];
const OBJECT_FIELD_CANDIDATES = ['object', 'Object', 'objectName', 'objectKey', 'key', 'resource', 'resourceName', 'bucket', 'bucketName', 'path', 'req.path', 'req.object', 'req.resource', 'trace.path'];
const ERROR_FIELD_CANDIDATES = ['error', 'Error', 'message', 'msg', 'err', 'statusMsg', 'resp.error', 'trace.error'];

const isPlainObject = (value) => value !== null && typeof value === 'object' && !Array.isArray(value);

const getCandidateValue = (entry, candidates) => {
    if (!isPlainObject(entry)) {
        return undefined;
    }

    for (const candidate of candidates) {
        const parts = candidate.split('.');
        let current = entry;
        let matched = true;

        for (const part of parts) {
            if (current && typeof current === 'object' && !Array.isArray(current) && Object.prototype.hasOwnProperty.call(current, part)) {
                current = current[part];
            } else {
                matched = false;
                break;
            }
        }

        if (!matched) {
            continue;
        }

        if (current !== undefined && current !== null) {
            if (typeof current === 'string' && current.trim() === '') {
                continue;
            }
            return current;
        }
    }

    return undefined;
};

const normalizeValue = (value) => {
    if (value === null || value === undefined) {
        return '';
    }
    if (typeof value === 'string') {
        return value.trim();
    }
    if (typeof value === 'number' || typeof value === 'bigint') {
        if (!Number.isFinite(Number(value))) {
            return '';
        }
        return String(value);
    }
    if (typeof value === 'boolean') {
        return value ? 'true' : 'false';
    }
    return '';
};

const formatStatusField = (value) => {
    const text = normalizeValue(value);
    if (!text) {
        return '';
    }

    const numeric = Number(text);
    if (!Number.isNaN(numeric)) {
        return String(Math.trunc(numeric));
    }

    return text;
};

const formatTimeField = (value) => {
    const text = normalizeValue(value);
    if (!text) {
        return '';
    }

    const unixPattern = /^[0-9]{10}$/;
    const unixMilliPattern = /^[0-9]{13}$/;

    if (unixPattern.test(text)) {
        const candidate = new Date(Number(text) * 1000);
        if (!Number.isNaN(candidate.getTime())) {
            return candidate.toLocaleString();
        }
    }

    if (unixMilliPattern.test(text)) {
        const candidate = new Date(Number(text));
        if (!Number.isNaN(candidate.getTime())) {
            return candidate.toLocaleString();
        }
    }

    const parsed = Date.parse(text);
    if (!Number.isNaN(parsed)) {
        const candidate = new Date(parsed);
        if (!Number.isNaN(candidate.getTime())) {
            return candidate.toLocaleString();
        }
    }

    return text;
};

const buildTableRows = (events) => {
    if (!Array.isArray(events)) {
        return [];
    }

    const rows = [];

    for (let index = 0; index < events.length && rows.length < MAX_TABLE_ROWS; index += 1) {
        const event = events[index];
        if (!isPlainObject(event)) {
            continue;
        }

        const timeValue = getCandidateValue(event, TIME_FIELD_CANDIDATES);
        const apiValue = getCandidateValue(event, API_FIELD_CANDIDATES);
        const statusValue = getCandidateValue(event, STATUS_FIELD_CANDIDATES);
        const clientValue = getCandidateValue(event, CLIENT_FIELD_CANDIDATES);
        const objectValue = getCandidateValue(event, OBJECT_FIELD_CANDIDATES);
        const errorValue = getCandidateValue(event, ERROR_FIELD_CANDIDATES);

        const row = {
            id: index,
            time: formatTimeField(timeValue),
            api: normalizeValue(apiValue),
            status: formatStatusField(statusValue),
            client: normalizeValue(clientValue),
            object: normalizeValue(objectValue),
            error: normalizeValue(errorValue),
        };

        if (row.time || row.api || row.status || row.client || row.object || row.error) {
            rows.push(row);
        }
    }

    return rows;
};

const formatScalar = (value) => {
    if (value === null || value === undefined) {
        return 'null';
    }
    if (typeof value === 'string') {
        if (value === '') {
            return '""';
        }
        if (/[:\-?{}\[\],&*#\n\r\t]|^\s|\s$/.test(value)) {
            return JSON.stringify(value);
        }
        return value;
    }
    if (typeof value === 'number' || typeof value === 'bigint') {
        if (!Number.isFinite(Number(value))) {
            return JSON.stringify(value);
        }
        return String(value);
    }
    if (typeof value === 'boolean') {
        return value ? 'true' : 'false';
    }
    if (Array.isArray(value) || isPlainObject(value)) {
        return JSON.stringify(value);
    }
    return JSON.stringify(value);
};

const formatYaml = (value, indent) => {
    const indentation = '  '.repeat(indent);

    if (Array.isArray(value)) {
        if (value.length === 0) {
            return `${indentation}[]`;
        }

        return value
            .map((item) => {
                if (isPlainObject(item) || Array.isArray(item)) {
                    const nested = formatYaml(item, indent + 1);
                    return `${indentation}-\n${nested}`;
                }
                return `${indentation}- ${formatScalar(item)}`;
            })
            .join('\n');
    }

    if (isPlainObject(value)) {
        const entries = Object.entries(value);
        if (entries.length === 0) {
            return `${indentation}{}`;
        }

        return entries
            .map(([key, val]) => {
                if (isPlainObject(val) || Array.isArray(val)) {
                    const nested = formatYaml(val, indent + 1);
                    return `${indentation}${key}:\n${nested}`;
                }
                return `${indentation}${key}: ${formatScalar(val)}`;
            })
            .join('\n');
    }

    return `${indentation}${formatScalar(value)}`;
};

const convertToYaml = (value) => {
    if (value === undefined) {
        return '';
    }

    if (Array.isArray(value) && value.length === 0) {
        return '[]';
    }

    if (isPlainObject(value) && Object.keys(value).length === 0) {
        return '{}';
    }

    return formatYaml(value, 0);
};

const createBadgeStyle = (backgroundColor, color) => ({
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: '999px',
    padding: '2px 8px',
    fontSize: '11px',
    fontWeight: 600,
    backgroundColor,
    color,
    minWidth: '32px'
});

const TraceOperations = ({ sites = [] }) => {
    const { t } = useI18n();
    const [form, setForm] = useState({
        alias: '',
        duration: '10s',
        statusInput: '',
        errorInput: '',
        groupByAPI: true,
        groupByClient: false,
        groupByVersions: false
    });
    const [results, setResults] = useState(null);
    const [rawViewMode, setRawViewMode] = useState('table');
    const [tableFilters, setTableFilters] = useState({
        api: '',
        status: '',
        client: ''
    });
    const [tablePagination, setTablePagination] = useState({
        page: 0,
        pageSize: 20
    });
    const [isRunning, setIsRunning] = useState(false);
    const [errorMessage, setErrorMessage] = useState('');
    const [copyStatus, setCopyStatus] = useState('idle');

    useEffect(() => {
        if (!form.alias && sites.length > 0) {
            const preferred = sites[0].name || sites[0].alias || '';
            setForm(prev => ({ ...prev, alias: preferred }));
        }
    }, [sites]);

    const numberFormatter = useMemo(() => new Intl.NumberFormat(), []);

    useEffect(() => {
        setRawViewMode('table');
        setTableFilters({ api: '', status: '', client: '' });
        setTablePagination({ page: 0, pageSize: 20 });
        setCopyStatus('idle');
    }, [results]);

    const parsedRawEvents = useMemo(() => {
        if (!results || !Array.isArray(results.rawEvents)) {
            return [];
        }

        return results.rawEvents.map((entry) => {
            if (typeof entry !== 'string') {
                return entry;
            }
            try {
                return JSON.parse(entry);
            } catch {
                return entry;
            }
        });
    }, [results]);

    const jsonRawContent = useMemo(() => {
        if (!parsedRawEvents.length) {
            return '[]';
        }
        try {
            return JSON.stringify(parsedRawEvents, null, 2);
        } catch {
            return '[]';
        }
    }, [parsedRawEvents]);

    const yamlRawContent = useMemo(() => {
        if (!parsedRawEvents.length) {
            return '[]';
        }
        return convertToYaml(parsedRawEvents);
    }, [parsedRawEvents]);

    const tableRows = useMemo(() => buildTableRows(parsedRawEvents), [parsedRawEvents]);

    const filteredRows = useMemo(() => {
        if (!tableRows.length) {
            return [];
        }

        const apiTerm = tableFilters.api.trim().toLowerCase();
        const statusTerm = tableFilters.status.trim().toLowerCase();
        const clientTerm = tableFilters.client.trim().toLowerCase();

        return tableRows.filter((row) => {
            const apiText = (row.api || '').toLowerCase();
            const statusText = (row.status || '').toLowerCase();
            const clientText = (row.client || '').toLowerCase();

            const apiMatch = !apiTerm || apiText.includes(apiTerm);
            const statusMatch = !statusTerm || statusText.includes(statusTerm);
            const clientMatch = !clientTerm || clientText.includes(clientTerm);

            return apiMatch && statusMatch && clientMatch;
        });
    }, [tableRows, tableFilters]);

    useEffect(() => {
        setTablePagination(prev => {
            const totalPages = Math.max(1, Math.ceil((filteredRows.length || 1) / prev.pageSize));
            const nextPage = Math.min(prev.page, totalPages - 1);
            if (nextPage !== prev.page) {
                return { ...prev, page: nextPage };
            }
            return prev;
        });
    }, [filteredRows.length]);

    const filterOptions = useMemo(() => {
        const apiSet = new Set();
        const statusSet = new Set();
        const clientSet = new Set();

        tableRows.forEach((row) => {
            if (row.api) {
                apiSet.add(row.api);
            }
            if (row.status) {
                statusSet.add(row.status);
            }
            if (row.client) {
                clientSet.add(row.client);
            }
        });

        return {
            api: Array.from(apiSet).sort((a, b) => a.localeCompare(b)),
            status: Array.from(statusSet).sort((a, b) => a.localeCompare(b)),
            client: Array.from(clientSet).sort((a, b) => a.localeCompare(b))
        };
    }, [tableRows]);

    const paginatedRows = useMemo(() => {
        const start = tablePagination.page * tablePagination.pageSize;
        const end = start + tablePagination.pageSize;
        return filteredRows.slice(start, end);
    }, [filteredRows, tablePagination]);

    const totalPages = useMemo(() => {
        if (!filteredRows.length) {
            return 1;
        }
        return Math.max(1, Math.ceil(filteredRows.length / tablePagination.pageSize));
    }, [filteredRows.length, tablePagination.pageSize]);

    useEffect(() => {
        if (copyStatus === 'success' || copyStatus === 'error') {
            const timer = setTimeout(() => setCopyStatus('idle'), 2000);
            return () => clearTimeout(timer);
        }
        return undefined;
    }, [copyStatus]);

    const handleChange = (field, value) => {
        setForm(prev => ({ ...prev, [field]: value }));
    };

    const handleTableFilterChange = (field, value) => {
        setTableFilters(prev => ({ ...prev, [field]: value }));
        setTablePagination(prev => ({ ...prev, page: 0 }));
    };

    const handlePageSizeChange = (value) => {
        const parsed = Number(value);
        const pageSize = Number.isNaN(parsed) ? 20 : Math.max(1, Math.min(parsed, 100));
        setTablePagination({ page: 0, pageSize });
    };

    const handlePageChange = (direction) => {
        setTablePagination(prev => {
            const nextPage = Math.min(
                Math.max(prev.page + direction, 0),
                Math.max(0, totalPages - 1)
            );
            if (nextPage === prev.page) {
                return prev;
            }
            return { ...prev, page: nextPage };
        });
    };

    const handleCopyRaw = async () => {
        if (!parsedRawEvents.length) {
            return;
        }
        const textToCopy = jsonRawContent || '';
        setCopyStatus('copying');
        try {
            if (navigator.clipboard && navigator.clipboard.writeText) {
                await navigator.clipboard.writeText(textToCopy);
            } else {
                const temporary = document.createElement('textarea');
                temporary.value = textToCopy;
                temporary.style.position = 'fixed';
                temporary.style.opacity = '0';
                document.body.appendChild(temporary);
                temporary.focus();
                temporary.select();
                document.execCommand('copy');
                document.body.removeChild(temporary);
            }
            setCopyStatus('success');
        } catch (error) {
            console.warn('Failed to copy trace payload', error);
            setCopyStatus('error');
        }
    };

    const parseStatusCodes = (input) => {
        if (!input) return [];
        return input
            .split(/[\s,]+/)
            .map(part => part.trim())
            .filter(part => part !== '')
            .map(part => parseInt(part, 10))
            .filter(code => !Number.isNaN(code));
    };

    const parseErrorFilters = (input) => {
        if (!input) return [];
        return input
            .split(/[\n,]+/)
            .map(part => part.trim())
            .filter(part => part !== '');
    };

    const handleRunTrace = async () => {
        if (!form.alias) {
            setErrorMessage(t('trace_operations_missing_alias', 'Please select an alias before running a trace.'));
            return;
        }

        setIsRunning(true);
        setErrorMessage('');

        const payload = {
            alias: form.alias,
            duration: form.duration,
            statusCodes: parseStatusCodes(form.statusInput),
            errorContains: parseErrorFilters(form.errorInput),
            groupByApi: form.groupByAPI,
            groupByClient: form.groupByClient,
            groupByVersions: form.groupByVersions
        };

        try {
            const data = await runTraceCapture(payload);
            setResults(data);
        } catch (error) {
            setErrorMessage(error.message || t('trace_operations_generic_error', 'Trace capture failed. Please try again.'));
        } finally {
            setIsRunning(false);
        }
    };

    const renderSummaryCards = () => {
        if (!results || !results.summary) return null;
        const summary = results.summary;

        const cards = [
            {
                label: t('trace_summary_events', 'Events Captured'),
                value: numberFormatter.format(summary.totalEvents || 0),
                tone: '#2563eb'
            },
            {
                label: t('trace_summary_errors', 'Distinct Errors'),
                value: numberFormatter.format(summary.distinctErrors || 0),
                tone: '#9333ea'
            },
            {
                label: t('trace_summary_objects', 'Objects With Errors'),
                value: numberFormatter.format(summary.objectsWithErrors || 0),
                tone: '#dc2626'
            },
            {
                label: t('trace_summary_duration', 'Capture Window'),
                value: summary.captureWindow || '-',
                tone: '#059669'
            }
        ];

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

    const renderObjectBadgeList = (objects = [], limit = 10) => {
        if (!Array.isArray(objects) || objects.length === 0) {
            return null;
        }

        const normalized = objects
            .filter((obj) => obj && typeof obj === 'object')
            .map((obj) => {
                const name = obj.name || obj.object || obj.path || obj.key || obj.id;
                const rawCount = obj.count ?? obj.errorCount ?? obj.events ?? obj.total ?? obj.value ?? 0;
                const count = Number.isFinite(Number(rawCount)) ? Number(rawCount) : 0;
                if (!name) {
                    return null;
                }
                return { name, count };
            })
            .filter(Boolean);

        if (!normalized.length) {
            return null;
        }

        const limited = normalized.slice(0, limit);

        return (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                {limited.map(({ name, count }, index) => {
                    // Split object name and version if present
                    const versionMatch = name.match(/^(.+?)\s*\(version:\s*(.+?)\)\s*$/);
                    const objectName = versionMatch ? versionMatch[1] : name;
                    const versionId = versionMatch ? versionMatch[2] : null;

                    return (
                        <div
                            key={`${name}-${index}`}
                            style={{
                                display: 'flex',
                                alignItems: 'flex-start',
                                gap: '12px'
                            }}
                        >
                            <span style={createBadgeStyle('#ede9fe', '#5b21b6')}>
                                {numberFormatter.format(count || 0)}
                            </span>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '2px', minWidth: 0, flex: 1 }}>
                                <span
                                    style={{
                                        fontFamily: 'monospace',
                                        fontSize: '12px',
                                        color: '#111827',
                                        wordBreak: 'break-all'
                                    }}
                                >
                                    {objectName}
                                </span>
                                {versionId && (
                                    <span
                                        style={{
                                            fontFamily: 'monospace',
                                            fontSize: '11px',
                                            color: '#6b7280',
                                            wordBreak: 'break-all',
                                            fontStyle: 'italic'
                                        }}
                                    >
                                        version: {versionId}
                                    </span>
                                )}
                            </div>
                        </div>
                    );
                })}
                {normalized.length > limit && (
                    <div style={{ fontSize: '12px', color: '#6b7280' }}>
                        {t('trace_more_indicator', '+{count} more', { count: normalized.length - limit })}
                    </div>
                )}
            </div>
        );
    };

    const renderErrorBadgeList = (errors = {}, limit = 10) => {
        if (!errors || (typeof errors !== 'object' && !Array.isArray(errors))) {
            return null;
        }

        let entries = [];

        if (Array.isArray(errors)) {
            entries = errors
                .map((item) => {
                    if (typeof item === 'string') {
                        return { name: item, count: 1 };
                    }
                    if (isPlainObject(item)) {
                        const name = item.name || item.message || item.error || item.code || item.id;
                        const rawCount = item.count ?? item.events ?? item.total ?? item.value ?? 1;
                        if (!name) {
                            return null;
                        }
                        const count = Number.isFinite(Number(rawCount)) ? Number(rawCount) : 0;
                        return { name, count };
                    }
                    return null;
                })
                .filter((entry) => entry && entry.count > 0 && entry.name);
        } else {
            entries = Object.entries(errors)
                .map(([name, value]) => {
                    const count = Number.isFinite(Number(value)) ? Number(value) : 0;
                    return { name, count };
                })
                .filter((entry) => entry.name && entry.count > 0);
        }

        if (!entries.length) {
            return null;
        }

        entries.sort((a, b) => b.count - a.count || a.name.localeCompare(b.name));
        const limited = entries.slice(0, limit);

        return (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                {limited.map(({ name, count }, index) => (
                    <div
                        key={`${name}-${index}`}
                        style={{
                            display: 'flex',
                            alignItems: 'flex-start',
                            // justifyContent: 'space-between',
                            gap: '12px'
                        }}
                    >
                        <span style={createBadgeStyle('#fee2e2', '#b91c1c')}>
                            {numberFormatter.format(count || 0)}
                        </span>
                        <span
                            style={{
                                fontFamily: 'monospace',
                                fontSize: '12px',
                                color: '#111827',
                                wordBreak: 'break-word'
                            }}
                        >
                            {name}
                        </span>
                    </div>
                ))}
                {entries.length > limit && (
                    <div style={{ fontSize: '12px', color: '#6b7280' }}>
                        {t('trace_more_indicator', '+{count} more', { count: entries.length - limit })}
                    </div>
                )}
            </div>
        );
    };

    const renderObjectsTable = () => {
        const items = (results?.objects || []).slice(0, 10);
        if (items.length === 0) {
            return (
                <div style={{
                    padding: '16px',
                    borderRadius: '6px',
                    backgroundColor: '#f3f4f6',
                    color: '#6b7280',
                    textAlign: 'center'
                }}>
                    {t('trace_objects_empty', 'No objects captured during this trace window.')}
                </div>
            );
        }

        return (
            <div style={{ marginTop: '24px' }}>
                <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#1f2937', marginBottom: '12px' }}>
                    {t('trace_objects_section_title', 'Top objects with errors')}
                </h4>
                <div style={{ border: '1px solid #e5e7eb', borderRadius: '6px', overflow: 'hidden' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                    <thead style={{ backgroundColor: '#f9fafb' }}>
                        <tr>
                            <th style={{ textAlign: 'left', padding: '12px', fontSize: '13px', color: '#6b7280' }}>
                                {t('trace_objects_header_object', 'Object')}
                            </th>
                            <th style={{ textAlign: 'left', padding: '12px', fontSize: '13px', color: '#6b7280' }}>
                                {t('trace_objects_header_events', 'Events')}
                            </th>
                            <th style={{ textAlign: 'left', padding: '12px', fontSize: '13px', color: '#6b7280' }}>
                                {t('trace_objects_header_errors', 'Error Summary')}
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        {items.map(item => {
                            const errorSource = item.errorCounts && Object.keys(item.errorCounts || {}).length > 0
                                ? item.errorCounts
                                : (Array.isArray(item.sampleErrors) && item.sampleErrors.length > 0 ? item.sampleErrors : []);
                            const errorBadges = renderErrorBadgeList(errorSource, 10);
                            return (
                                <tr key={item.name} style={{ borderTop: '1px solid #f3f4f6' }}>
                                    <td style={{ padding: '12px', fontFamily: 'monospace', fontSize: '13px', wordBreak: 'break-all' }}>
                                        {item.name}
                                    </td>
                                    <td style={{ padding: '12px', fontSize: '13px' }}>
                                        <span style={createBadgeStyle('#dbeafe', '#1d4ed8')}>
                                            {numberFormatter.format(item.count || 0)}
                                        </span>
                                    </td>
                                    <td style={{ padding: '12px', fontSize: '13px', color: '#4b5563' }}>
                                        {errorBadges || t('trace_no_error_counts', 'No error summary available.')}
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                    </table>
                </div>
            </div>
        );
    };

    const renderErrorPatterns = () => {
        const items = results?.errors || [];
        if (!items.length) return null;

        return (
            <div style={{ marginTop: '24px' }}>
                <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#1f2937', marginBottom: '12px' }}>
                    {t('trace_error_patterns_title', 'Top Error Patterns')}
                </h4>
                <div style={{ border: '1px solid #e5e7eb', borderRadius: '6px', overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '640px' }}>
                        <thead style={{ backgroundColor: '#f9fafb' }}>
                            <tr>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280' }}>
                                    {t('trace_error_table_column_message', 'Error message')}
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', whiteSpace: 'nowrap' }}>
                                    {t('trace_error_table_column_events', 'Events')}
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280' }}>
                                    {t('trace_error_table_column_objects', 'Top objects')}
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {items.slice(0, 10).map((item) => {
                                const objectList = renderObjectBadgeList(item.objects || [], 10);
                                return (
                                    <tr key={item.message} style={{ borderTop: '1px solid #f3f4f6' }}>
                                        <td style={{ padding: '12px', fontSize: '13px', color: '#111827', wordBreak: 'break-word' }}>{item.message}</td>
                                        <td style={{ padding: '12px', fontSize: '13px', color: '#111827', whiteSpace: 'nowrap' }}>
                                            <span style={createBadgeStyle('#dbeafe', '#1d4ed8')}>
                                                {numberFormatter.format(item.count || 0)}
                                            </span>
                                        </td>
                                        <td style={{ padding: '12px', fontSize: '12px', color: '#4b5563' }}>
                                            {objectList || t('trace_error_table_no_objects', 'No object details available.')}
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </div>
        );
    };

    const renderGroupedSection = (items = [], title) => {
        if (!items || items.length === 0) return null;

        const columnWidths = ['20%', '10%', '12%', '22%', '36%'];
        return (
            <div style={{ marginTop: '24px' }}>
                <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#1f2937', marginBottom: '12px' }}>{title}</h4>
                <div style={{ border: '1px solid #e5e7eb', borderRadius: '6px', overflowX: 'auto' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '760px', tableLayout: 'fixed' }}>
                        <thead style={{ backgroundColor: '#f9fafb' }}>
                            <tr>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', width: columnWidths[0] }}>
                                    {t('trace_group_table_column_group', 'Group')}
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', width: columnWidths[1] }}>
                                    {t('trace_group_table_column_events', 'Events')}
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', width: columnWidths[2] }}>
                                    {t('trace_group_table_column_objects', 'Objects impacted')}
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', width: columnWidths[3] }}>
                                    {t('trace_group_table_column_errors', 'Top errors')}
                                </th>
                                <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', width: columnWidths[4] }}>
                                    {t('trace_group_table_column_top_objects', 'Top objects')}
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {items.slice(0, 10).map((item) => {
                                const objectCount = item.objectCount ?? (item.objects ? item.objects.length : 0);
                                const objectList = renderObjectBadgeList(item.objects || [], 10);
                                const errorList = renderErrorBadgeList(item.errorCounts || item.errors || item.sampleErrors || [], 10);
                                return (
                                    <tr key={item.name} style={{ borderTop: '1px solid #f3f4f6' }}>
                                        <td style={{ padding: '12px', fontSize: '13px', color: '#111827', wordBreak: 'break-word', width: columnWidths[0] }}>{item.name}</td>
                                        <td style={{ padding: '12px', fontSize: '13px', color: '#111827', width: columnWidths[1] }}>
                                            <span style={createBadgeStyle('#dbeafe', '#1d4ed8')}>
                                                {numberFormatter.format(item.count || 0)}
                                            </span>
                                        </td>
                                        <td style={{ padding: '12px', fontSize: '13px', color: '#111827', width: columnWidths[2] }}>
                                            <span style={createBadgeStyle('#fef3c7', '#92400e')}>
                                                {numberFormatter.format(objectCount || 0)}
                                            </span>
                                        </td>
                                        <td style={{ padding: '12px', fontSize: '12px', color: '#4b5563', width: columnWidths[3] }}>
                                            {errorList || t('trace_no_error_counts', 'No error summary available.')}
                                        </td>
                                        <td style={{ padding: '12px', fontSize: '12px', color: '#4b5563', width: columnWidths[4] }}>
                                            {objectList || t('trace_group_table_no_objects', 'No object breakdown available.')}
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </div>
        );
    };

    const renderRawEvents = () => {
        if (!results || !Array.isArray(results.rawEvents) || results.rawEvents.length === 0) {
            return null;
        }

        const rawContent = results.rawEvents.join('\n');
        const viewOptions = [
            { id: 'table', label: t('trace_raw_view_option_table', 'Table') },
            { id: 'json', label: t('trace_raw_view_option_json', 'JSON') },
            { id: 'yaml', label: t('trace_raw_view_option_yaml', 'YAML') },
            { id: 'raw', label: t('trace_raw_view_option_raw', 'Raw') }
        ];

        const renderViewContent = () => {
            const { api: apiOptions, status: statusOptions, client: clientOptions } = filterOptions;
            if (rawViewMode === 'table') {
                if (!tableRows.length) {
                    return (
                        <div
                            style={{
                                padding: '16px',
                                borderRadius: '6px',
                                border: '1px solid #e5e7eb',
                                backgroundColor: '#f9fafb',
                                color: '#6b7280'
                            }}
                        >
                            {t('trace_table_empty', 'No structured events available for tabular view.')}
                        </div>
                    );
                }

                const hasFilteredResults = filteredRows.length > 0;
                const currentStartIndex = hasFilteredResults
                    ? tablePagination.page * tablePagination.pageSize + 1
                    : 0;
                const currentEndIndex = hasFilteredResults
                    ? Math.min(currentStartIndex + tablePagination.pageSize - 1, filteredRows.length)
                    : 0;

                const renderPaginationControls = (position) => (
                    <div
                        style={{
                            display: 'flex',
                            flexWrap: 'wrap',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                            gap: '12px',
                            padding: position === 'bottom' ? '12px' : '0 0 12px 0',
                            borderTop: position === 'bottom' ? '1px solid #d1d5db' : 'none'
                        }}
                    >
                        <span style={{ fontSize: '12px', color: '#4b5563' }}>
                            {t('trace_table_pagination_summary', '{start}-{end} of {total}', {
                                start: numberFormatter.format(currentStartIndex),
                                end: numberFormatter.format(currentEndIndex),
                                total: numberFormatter.format(filteredRows.length)
                            })}
                        </span>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                            <label style={{ fontSize: '12px', color: '#4b5563', display: 'inline-flex', alignItems: 'center', gap: '6px' }}>
                                {t('trace_table_page_size_label', 'Rows per page')}
                                <select
                                    value={tablePagination.pageSize}
                                    onChange={(e) => handlePageSizeChange(e.target.value)}
                                    style={{
                                        padding: '6px 10px',
                                        borderRadius: '6px',
                                        border: '1px solid #d1d5db',
                                        fontSize: '12px',
                                        backgroundColor: '#ffffff'
                                    }}
                                >
                                    {PAGE_SIZE_OPTIONS.map((size) => (
                                        <option key={`page-size-${size}`} value={size}>{size}</option>
                                    ))}
                                </select>
                            </label>
                            <div style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                                <button
                                    type="button"
                                    onClick={() => handlePageChange(-1)}
                                    disabled={tablePagination.page === 0}
                                    style={{
                                        padding: '6px 10px',
                                        borderRadius: '6px',
                                        border: '1px solid #d1d5db',
                                        backgroundColor: tablePagination.page === 0 ? '#f3f4f6' : '#ffffff',
                                        color: '#1f2937',
                                        fontSize: '12px',
                                        cursor: tablePagination.page === 0 ? 'not-allowed' : 'pointer'
                                    }}
                                >
                                    {t('trace_table_page_prev', 'Prev')}
                                </button>
                                <span style={{ fontSize: '12px', color: '#4b5563', minWidth: '60px', textAlign: 'center' }}>
                                    {numberFormatter.format(tablePagination.page + 1)} / {numberFormatter.format(totalPages)}
                                </span>
                                <button
                                    type="button"
                                    onClick={() => handlePageChange(1)}
                                    disabled={tablePagination.page >= totalPages - 1}
                                    style={{
                                        padding: '6px 10px',
                                        borderRadius: '6px',
                                        border: '1px solid #d1d5db',
                                        backgroundColor: tablePagination.page >= totalPages - 1 ? '#f3f4f6' : '#ffffff',
                                        color: '#1f2937',
                                        fontSize: '12px',
                                        cursor: tablePagination.page >= totalPages - 1 ? 'not-allowed' : 'pointer'
                                    }}
                                >
                                    {t('trace_table_page_next', 'Next')}
                                </button>
                            </div>
                        </div>
                    </div>
                );

                return (
                    <div>
                        <div
                            style={{
                                display: 'flex',
                                flexWrap: 'wrap',
                                gap: '12px',
                                marginBottom: '12px'
                            }}
                        >
                            <select
                                value={tableFilters.api}
                                onChange={(e) => handleTableFilterChange('api', e.target.value)}
                                style={{
                                    flex: '1 1 160px',
                                    minWidth: '140px',
                                    padding: '8px 10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '12px',
                                    backgroundColor: '#ffffff'
                                }}
                            >
                                <option value="">{t('trace_table_filter_api_select', 'All APIs')}</option>
                                {apiOptions.map((value) => (
                                    <option key={`api-${value}`} value={value}>{value}</option>
                                ))}
                            </select>
                            <select
                                value={tableFilters.status}
                                onChange={(e) => handleTableFilterChange('status', e.target.value)}
                                style={{
                                    flex: '1 1 140px',
                                    minWidth: '120px',
                                    padding: '8px 10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '12px',
                                    backgroundColor: '#ffffff'
                                }}
                            >
                                <option value="">{t('trace_table_filter_status_select', 'All statuses')}</option>
                                {statusOptions.map((value) => (
                                    <option key={`status-${value}`} value={value}>{value}</option>
                                ))}
                            </select>
                            <select
                                value={tableFilters.client}
                                onChange={(e) => handleTableFilterChange('client', e.target.value)}
                                style={{
                                    flex: '1 1 200px',
                                    minWidth: '160px',
                                    padding: '8px 10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '12px',
                                    backgroundColor: '#ffffff'
                                }}
                            >
                                <option value="">{t('trace_table_filter_client_select', 'All clients')}</option>
                                {clientOptions.map((value) => (
                                    <option key={`client-${value}`} value={value}>{value}</option>
                                ))}
                            </select>
                        </div>
                        {hasFilteredResults ? (
                            <>
                                {renderPaginationControls('top')}
                                <div
                                    style={{
                                        border: '1px solid #d1d5db',
                                        borderRadius: '6px',
                                        overflowX: 'auto'
                                    }}
                                >
                                <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: '640px' }}>
                                    <thead style={{ backgroundColor: '#f9fafb' }}>
                                        <tr>
                                            <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', whiteSpace: 'nowrap' }}>
                                                {t('trace_table_column_time', 'Time')}
                                            </th>
                                            <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', whiteSpace: 'nowrap' }}>
                                                {t('trace_table_column_api', 'API')}
                                            </th>
                                            <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', whiteSpace: 'nowrap' }}>
                                                {t('trace_table_column_status', 'Status')}
                                            </th>
                                            <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280', whiteSpace: 'nowrap' }}>
                                                {t('trace_table_column_client', 'Client')}
                                            </th>
                                            <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280' }}>
                                                {t('trace_table_column_object', 'Object')}
                                            </th>
                                            <th style={{ textAlign: 'left', padding: '12px', fontSize: '12px', color: '#6b7280' }}>
                                                {t('trace_table_column_error', 'Error')}
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {paginatedRows.map((row) => (
                                            <tr key={row.id} style={{ borderTop: '1px solid #f3f4f6' }}>
                                                <td style={{ padding: '12px', fontSize: '12px', color: '#111827', whiteSpace: 'nowrap' }}>{row.time || '--'}</td>
                                                <td style={{ padding: '12px', fontSize: '12px', color: '#111827', whiteSpace: 'nowrap' }}>{row.api || '--'}</td>
                                                <td style={{ padding: '12px', fontSize: '12px', color: '#111827', whiteSpace: 'nowrap' }}>{row.status || '--'}</td>
                                                <td style={{ padding: '12px', fontSize: '12px', color: '#4b5563', whiteSpace: 'nowrap' }}>{row.client || '--'}</td>
                                                <td style={{ padding: '12px', fontSize: '12px', color: '#4b5563', wordBreak: 'break-all' }}>{row.object || '--'}</td>
                                                <td style={{ padding: '12px', fontSize: '12px', color: '#4b5563', wordBreak: 'break-word' }}>{row.error || '--'}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                                    {renderPaginationControls('bottom')}
                                </div>
                            </>
                        ) : (
                            <div
                                style={{
                                    padding: '16px',
                                    borderRadius: '6px',
                                    border: '1px solid #e5e7eb',
                                    backgroundColor: '#fef2f2',
                                    color: '#b91c1c'
                                }}
                            >
                                {t('trace_table_filter_empty', 'No events match the current filters.')}
                            </div>
                        )}
                    </div>
                );
            }

            const content = rawViewMode === 'json' ? jsonRawContent : rawViewMode === 'yaml' ? yamlRawContent : rawContent;

            return (
                <textarea
                    readOnly
                    value={content}
                    style={{
                        width: '100%',
                        height: '220px',
                        fontFamily: 'monospace',
                        fontSize: '12px',
                        padding: '12px',
                        borderRadius: '6px',
                        border: '1px solid #d1d5db',
                        backgroundColor: '#111827',
                        color: '#f9fafb',
                        whiteSpace: 'pre'
                    }}
                />
            );
        };

        return (
            <div style={{ marginTop: '24px' }}>
                <h4 style={{ fontSize: '16px', fontWeight: 600, color: '#1f2937', marginBottom: '8px' }}>
                    {t('trace_raw_events_title', 'Raw Event Sample')} ({results.rawEvents.length}/{results.rawEventCount})
                </h4>
                <p style={{ fontSize: '12px', color: '#6b7280', marginBottom: '12px' }}>
                    {t('trace_raw_events_hint', 'Preview of the first captured events. Use the CLI to export the complete trace if needed.')}
                </p>
                <div
                    style={{
                        display: 'flex',
                        flexWrap: 'wrap',
                        alignItems: 'center',
                        gap: '8px',
                        marginBottom: '12px'
                    }}
                >
                    <span style={{ fontSize: '12px', color: '#6b7280' }}>{t('trace_raw_view_mode', 'View mode')}:</span>
                    {viewOptions.map((option) => {
                        const isActive = rawViewMode === option.id;
                        return (
                            <button
                                key={option.id}
                                type="button"
                                onClick={() => setRawViewMode(option.id)}
                                style={{
                                    padding: '6px 12px',
                                    borderRadius: '16px',
                                    border: `1px solid ${isActive ? '#1f2937' : '#d1d5db'}`,
                                    backgroundColor: isActive ? '#1f2937' : '#ffffff',
                                    color: isActive ? '#f9fafb' : '#1f2937',
                                    fontSize: '12px',
                                    cursor: 'pointer'
                                }}
                            >
                                {option.label}
                            </button>
                        );
                    })}
                    <button
                        type="button"
                        onClick={handleCopyRaw}
                        disabled={!parsedRawEvents.length || copyStatus === 'copying'}
                        style={{
                            padding: '6px 12px',
                            borderRadius: '6px',
                            border: '1px solid #d1d5db',
                            backgroundColor: '#ffffff',
                            color: '#1f2937',
                            fontSize: '12px',
                            cursor: !parsedRawEvents.length || copyStatus === 'copying' ? 'not-allowed' : 'pointer'
                        }}
                    >
                        {copyStatus === 'copying'
                            ? t('trace_raw_copy_working', 'Copying...')
                            : copyStatus === 'success'
                                ? t('trace_raw_copy_success', 'Copied!')
                                : copyStatus === 'error'
                                    ? t('trace_raw_copy_error', 'Copy failed')
                                    : t('trace_raw_copy_button', 'Copy raw data')}
                    </button>
                </div>
                {renderViewContent()}
            </div>
        );
    };

    const filtersSummary = () => {
        if (!results || !results.filters) return null;
        const { statusCodes = [], errorContains = [], groupByAPI, groupByClient, groupByVersions, duration } = results.filters;

        return (
            <div style={{
                display: 'flex',
                flexWrap: 'wrap',
                gap: '8px',
                marginBottom: '16px'
            }}>
                <span className="badge badge-neutral">{t('trace_filter_duration', 'Duration: {value}', { value: duration })}</span>
                {statusCodes.length > 0 && (
                    <span className="badge badge-info">{t('trace_filter_status', 'Status: {value}', { value: statusCodes.join(', ') })}</span>
                )}
                {errorContains.length > 0 && (
                    <span className="badge badge-info">{t('trace_filter_error', 'Error contains: {value}', { value: errorContains.join('; ') })}</span>
                )}
                {groupByAPI && (
                    <span className="badge badge-success">{t('trace_filter_group_api', 'Grouped by API')}</span>
                )}
                {groupByClient && (
                    <span className="badge badge-success">{t('trace_filter_group_client', 'Grouped by client')}</span>
                )}
                {groupByVersions && (
                    <span className="badge badge-success">{t('trace_filter_group_versions', 'Grouped by versions')}</span>
                )}
            </div>
        );
    };

    return (
        <div>
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <Activity size={20} />
                        {t('trace_operations_title', 'Trace Error Analyzer')}
                    </h3>
                    <p style={{ margin: '8px 0 0 0', fontSize: '14px', color: '#6b7280' }}>
                        {t('trace_operations_description', 'Capture mc admin trace output directly from the browser to pinpoint recurring failures, filter by status or message, and highlight the APIs or clients causing the most issues.')}
                    </p>
                </div>

                <div style={{ padding: '20px' }}>
                    {errorMessage && (
                        <div style={{
                            marginBottom: '16px',
                            border: '1px solid #f87171',
                            backgroundColor: '#fee2e2',
                            color: '#b91c1c',
                            padding: '12px',
                            borderRadius: '6px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px'
                        }}>
                            <AlertCircle size={18} />
                            <span>{errorMessage}</span>
                        </div>
                    )}

                    <div style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
                        gap: '16px',
                        marginBottom: '20px'
                    }}>
                        <div>
                            <label style={{ display: 'block', fontSize: '13px', color: '#4b5563', marginBottom: '6px' }}>
                                {t('trace_form_alias', 'Alias')}
                            </label>
                            <select
                                value={form.alias || ''}
                                onChange={(e) => handleChange('alias', e.target.value)}
                                style={{
                                    width: '100%',
                                    padding: '10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '14px'
                                }}
                            >
                                <option value="">{t('trace_form_alias_placeholder', 'Select alias...')}</option>
                                {sites.map(site => {
                                    const optionLabel = site.name || site.alias || site.url || 'alias';
                                    const isDisabled = !site.healthy || site.status === 'checking' || site.status === 'timeout' || site.status === 'error';
                                    const statusSuffix = isDisabled 
                                        ? site.status === 'checking' 
                                            ? ' (Checking...)' 
                                            : site.status === 'timeout'
                                                ? ' (Timeout)'
                                                : ' (Unhealthy)'
                                        : '';
                                    return (
                                        <option 
                                            key={optionLabel} 
                                            value={optionLabel}
                                            disabled={isDisabled}
                                        >
                                            {optionLabel}{statusSuffix}
                                        </option>
                                    );
                                })}
                            </select>
                        </div>

                        <div>
                            <label style={{ display: 'block', fontSize: '13px', color: '#4b5563', marginBottom: '6px' }}>
                                {t('trace_form_duration', 'Duration')}
                            </label>
                            <select
                                value={form.duration}
                                onChange={(e) => handleChange('duration', e.target.value)}
                                style={{
                                    width: '100%',
                                    padding: '10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '14px'
                                }}
                            >
                                {['5s', '10s', '15s', '20s', '30s', '45s', '60s', '90s', '120s', '300s'].map(option => (
                                    <option key={option} value={option}>{option}</option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label style={{ display: 'block', fontSize: '13px', color: '#4b5563', marginBottom: '6px' }}>
                                {t('trace_form_status', 'HTTP Status Filters')}
                            </label>
                            <input
                                type="text"
                                value={form.statusInput}
                                onChange={(e) => handleChange('statusInput', e.target.value)}
                                placeholder={t('trace_form_status_placeholder', 'Example: 404, 409, 500')}
                                style={{
                                    width: '100%',
                                    padding: '10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '14px'
                                }}
                            />
                        </div>

                        <div>
                            <label style={{ display: 'block', fontSize: '13px', color: '#4b5563', marginBottom: '6px' }}>
                                {t('trace_form_error_contains', 'Error Message Filters')}
                            </label>
                            <input
                                value={form.errorInput}
                                onChange={(e) => handleChange('errorInput', e.target.value)}
                                rows={1}
                                placeholder={t('trace_form_error_placeholder', 'One substring per line, e.g. AccessDenied')}
                                style={{
                                    width: '100%',
                                    padding: '10px',
                                    borderRadius: '6px',
                                    border: '1px solid #d1d5db',
                                    fontSize: '14px'
                                }}
                            />
                        </div>
                    </div>

                    <div style={{
                        display: 'flex',
                        flexWrap: 'wrap',
                        gap: '16px',
                        alignItems: 'center',
                        marginBottom: '20px'
                    }}>
                        <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '14px', color: '#374151' }}>
                            <input
                                type="checkbox"
                                checked={form.groupByAPI}
                                onChange={(e) => handleChange('groupByAPI', e.target.checked)}
                            />
                            {t('trace_form_group_api', 'Group by API action')}
                        </label>
                        <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '14px', color: '#374151' }}>
                            <input
                                type="checkbox"
                                checked={form.groupByClient}
                                onChange={(e) => handleChange('groupByClient', e.target.checked)}
                            />
                            {t('trace_form_group_client', 'Group by client host')}
                        </label>
                        <label style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '14px', color: '#374151' }}>
                            <input
                                type="checkbox"
                                checked={form.groupByVersions}
                                onChange={(e) => handleChange('groupByVersions', e.target.checked)}
                            />
                            {t('trace_form_group_versions', 'Group by object versions')}
                        </label>
                        <span style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px', color: '#6b7280' }}>
                            <Filter size={14} />
                            {t('trace_form_tip', 'Trace runtime matches the selected duration. Longer captures may take additional time to return results.')}
                        </span>
                    </div>

                    <button
                        className="btn btn-primary"
                        onClick={handleRunTrace}
                        disabled={!form.alias || isRunning}
                        style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: '8px',
                            padding: '12px 18px',
                            fontSize: '15px'
                        }}
                    >
                        <Play size={16} />
                        {isRunning ? t('trace_form_running', 'Capturing trace...') : t('trace_form_run', 'Run trace capture')}
                    </button>
                </div>
            </div>

            {results && (
                <div className="card" style={{ marginTop: '24px' }}>
                    <div className="card-header">
                        <h3 className="card-title" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <BarChart3 size={18} />
                            {t('trace_results_title', 'Trace Summary')}
                        </h3>
                        <p style={{ margin: '8px 0 0 0', fontSize: '13px', color: '#6b7280' }}>
                            {t('trace_results_subtitle', 'Captured on {alias} between {start} and {end}', {
                                alias: results.summary?.alias || form.alias,
                                start: results.summary?.startedAt,
                                end: results.summary?.completedAt
                            })}
                        </p>
                    </div>
                    <div style={{ padding: '20px' }}>
                        {filtersSummary()}
                        {renderSummaryCards()}
                        {renderObjectsTable()}
                        {renderErrorPatterns()}
                        {form.groupByAPI && renderGroupedSection(results.apiStats, t('trace_api_section_title', 'API Hotspots'), t('trace_api_section_empty', 'No API groups available.'))}
                        {form.groupByClient && renderGroupedSection(results.clientStats, t('trace_client_section_title', 'Client Hotspots'), t('trace_client_section_empty', 'No client groups available.'))}
                        {renderRawEvents()}
                    </div>
                </div>
            )}
        </div>
    );
};

export default TraceOperations;
