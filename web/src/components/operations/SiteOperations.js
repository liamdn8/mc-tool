import React, {
    useCallback,
    useEffect,
    useMemo,
    useRef,
    useState,
} from 'react';
import {
    Play,
    Square,
    Terminal,
    Eraser,
    AlertTriangle,
} from 'lucide-react';
import { useI18n } from '../../utils/i18n';
import {
    fetchReplicationResyncStatus,
    loadReplicationResyncOptions,
    startReplicationResyncOperation,
} from '../../utils/api';

const POLLING_DELAY_MS = 1000;
const MAX_HISTORY_LINES = 400;
const ANSI_ESCAPE_PATTERN = /\u001b\[[0-9;?]*[ -\/]*[@-~]/gi;
const COMPLETION_MESSAGE = '# Resync completed \u2713';

const stripAnsi = (value = '') => value.replace(ANSI_ESCAPE_PATTERN, '');

const hasCompletedStatus = (status, lines) => {
    const candidates = [
        status?.status,
        status?.Status,
        status?.state,
        status?.State,
    ]
        .filter((value) => typeof value === 'string' && value.trim().length > 0)
        .map((value) => value.toLowerCase());

    if (candidates.some((value) => value.startsWith('complete'))) {
        return true;
    }

    if (Array.isArray(lines)) {
        return lines.some((line) => {
            const lower = (line || '').toLowerCase();
            return lower.includes('status') && lower.includes('completed');
        });
    }

    return false;
};

const buildTerminalFrame = (status) => {
    if (!status) {
        return ['Waiting for status data...'];
    }

    const asLines = (text) =>
        text
            .replace(/\r/g, '')
            .split('\n')
            .map((line) => stripAnsi(line).replace(/\s+$/, ''));

    if (typeof status.prettyOutput === 'string' && status.prettyOutput.trim().length > 0) {
        return asLines(status.prettyOutput);
    }

    if (typeof status.summaryText === 'string' && status.summaryText.trim().length > 0) {
        return asLines(status.summaryText);
    }

    if (typeof status.rawOutput === 'string' && status.rawOutput.trim().length > 0) {
        return asLines(status.rawOutput);
    }

    if (Array.isArray(status.reports) && status.reports.length > 0) {
        return status.reports.map((entry) => JSON.stringify(entry));
    }

    return ['No status information returned'];
};

const SiteOperations = ({ hasReplication }) => {
    const { t } = useI18n();

    const [loadingOptions, setLoadingOptions] = useState(true);
    const [optionsError, setOptionsError] = useState(null);
    const [resyncOptions, setResyncOptions] = useState({ aliases: [], clusters: [] });

    const [sourceAlias, setSourceAlias] = useState('');
    const [targetAlias, setTargetAlias] = useState('');

    const [isStarting, setIsStarting] = useState(false);
    const [startError, setStartError] = useState(null);
    const [statusError, setStatusError] = useState(null);

    const [isWatching, setIsWatching] = useState(false);

    const [terminalHistory, setTerminalHistory] = useState([]);
    const [terminalFrame, setTerminalFrame] = useState([]);
    const terminalFrameRef = useRef([]);
    const terminalViewportRef = useRef(null);
    const watchingPairRef = useRef({ source: '', target: '' });
    const pollTimeoutRef = useRef(null);
    const isWatchingRef = useRef(false);
    const fetchInFlightRef = useRef(false);
    const completionAnnouncedRef = useRef(false);
    const clearTokenRef = useRef(0);

    const [lastUpdated, setLastUpdated] = useState(null);

    const aliasMap = useMemo(() => {
        const map = new Map();
        (resyncOptions.aliases || []).forEach((alias) => {
            map.set(alias.alias, alias);
        });
        return map;
    }, [resyncOptions]);

    const availableSources = useMemo(
        () => (resyncOptions.aliases || []).filter((item) => item.canResync),
        [resyncOptions]
    );

    const availableTargets = useMemo(() => {
        if (!sourceAlias) {
            return [];
        }
        const source = aliasMap.get(sourceAlias);
        if (!source || !source.clusterId) {
            return [];
        }
        return (resyncOptions.aliases || []).filter(
            (item) =>
                item.alias !== sourceAlias &&
                item.clusterId === source.clusterId &&
                item.canResync
        );
    }, [aliasMap, resyncOptions, sourceAlias]);

    const sourceDetails = useMemo(
        () => (sourceAlias ? aliasMap.get(sourceAlias) || null : null),
        [aliasMap, sourceAlias]
    );

    const targetDetails = useMemo(
        () => (targetAlias ? aliasMap.get(targetAlias) || null : null),
        [aliasMap, targetAlias]
    );

    const combinedTerminalLines = useMemo(
        () => [...terminalHistory, ...terminalFrame],
        [terminalFrame, terminalHistory]
    );

    const clearPolling = useCallback(() => {
        if (pollTimeoutRef.current) {
            clearTimeout(pollTimeoutRef.current);
            pollTimeoutRef.current = null;
        }
    }, []);

    const appendHistory = useCallback((entries) => {
        const normalized = Array.isArray(entries) ? entries : [entries];
        setTerminalHistory((prev) => {
            const next = [...prev, ...normalized];
            return next.slice(-MAX_HISTORY_LINES);
        });
    }, []);

    const clearTerminal = useCallback(() => {
        clearTokenRef.current += 1;
        completionAnnouncedRef.current = false;
        setTerminalHistory([]);
        setTerminalFrame([]);
        terminalFrameRef.current = [];
        setStatusError(null);
        setLastUpdated(null);
    }, []);

    const stopWatching = useCallback(
        (reason = null) => {
            clearPolling();
            isWatchingRef.current = false;
            fetchInFlightRef.current = false;
            watchingPairRef.current = { source: '', target: '' };
            completionAnnouncedRef.current = false;
            setIsWatching(false);
            if (terminalFrameRef.current.length > 0) {
                appendHistory(terminalFrameRef.current);
            }
            if (reason) {
                appendHistory(reason);
            }
            setTerminalFrame([]);
        },
        [appendHistory, clearPolling]
    );

    const fetchStatusFrame = useCallback(async () => {
        if (!sourceAlias || !targetAlias || !isWatchingRef.current) {
            return;
        }
        if (fetchInFlightRef.current) {
            return;
        }

        const clearToken = clearTokenRef.current;
        fetchInFlightRef.current = true;
        try {
            setStatusError(null);
            const response = await fetchReplicationResyncStatus(sourceAlias, targetAlias);
            const status = response?.status;
            let lines = buildTerminalFrame(status);

            if (!isWatchingRef.current) {
                return;
            }

            if (clearToken !== clearTokenRef.current) {
                return;
            }

            if (hasCompletedStatus(status, lines)) {
                completionAnnouncedRef.current = true;
            }

            if (completionAnnouncedRef.current) {
                lines = [...lines, '', COMPLETION_MESSAGE];
            }

            setTerminalFrame(lines);
            setLastUpdated(new Date());
        } catch (error) {
            if (!isWatchingRef.current) {
                return;
            }
            const message = error?.message || 'Failed to fetch resync status';
            setStatusError(message);
            stopWatching('[error] ' + message);
        } finally {
            fetchInFlightRef.current = false;
            if (isWatchingRef.current) {
                clearPolling();
                pollTimeoutRef.current = setTimeout(() => {
                    fetchStatusFrame();
                }, POLLING_DELAY_MS);
            }
        }
    }, [clearPolling, sourceAlias, stopWatching, targetAlias]);

    const startWatching = useCallback(() => {
        if (!sourceAlias || !targetAlias) {
            return;
        }

    clearPolling();
        isWatchingRef.current = true;
        fetchInFlightRef.current = false;
    watchingPairRef.current = { source: sourceAlias, target: targetAlias };
    completionAnnouncedRef.current = false;
        setStatusError(null);
        setLastUpdated(null);
        setTerminalFrame([]);
        appendHistory([
            `$ mc admin replicate resync status ${sourceAlias} ${targetAlias}`,
            '',
        ]);

        setIsWatching(true);

        fetchStatusFrame();
    }, [appendHistory, clearPolling, fetchStatusFrame, sourceAlias, targetAlias]);

    const handleStartResync = useCallback(async () => {
        if (!sourceAlias || !targetAlias || isStarting) {
            return;
        }

        setIsStarting(true);
        setStartError(null);

        try {
            const result = await startReplicationResyncOperation(sourceAlias, targetAlias);
            if (result?.message) {
                appendHistory(`# ${result.message}`);
            }
            startWatching();
        } catch (error) {
            const message = error?.message || 'Failed to start replication resync';
            setStartError(message);
            appendHistory('[error] ' + message);
        } finally {
            setIsStarting(false);
        }
    }, [appendHistory, isStarting, sourceAlias, startWatching, targetAlias]);

    const handleWatchOnly = useCallback(() => {
        if (!sourceAlias || !targetAlias) {
            return;
        }
        startWatching();
    }, [sourceAlias, startWatching, targetAlias]);

    useEffect(() => {
        let mounted = true;
        const loadOptions = async () => {
            setLoadingOptions(true);
            setOptionsError(null);
            try {
                const options = await loadReplicationResyncOptions();
                if (mounted) {
                    setResyncOptions(options || { aliases: [], clusters: [] });
                }
            } catch (error) {
                if (mounted) {
                    const message = error?.message || 'Failed to load resync options';
                    setOptionsError(message);
                }
            } finally {
                if (mounted) {
                    setLoadingOptions(false);
                }
            }
        };

        loadOptions();
        return () => {
            mounted = false;
        };
    }, []);

    useEffect(() => {
        terminalFrameRef.current = terminalFrame;
    }, [terminalFrame]);

    useEffect(() => {
        isWatchingRef.current = isWatching;
        if (!isWatching) {
            clearPolling();
        }
    }, [clearPolling, isWatching]);

    useEffect(() => {
        const viewport = terminalViewportRef.current;
        if (!viewport) {
            return;
        }
        viewport.scrollTop = viewport.scrollHeight;
    }, [combinedTerminalLines]);

    useEffect(() => {
        return () => {
            isWatchingRef.current = false;
            clearPolling();
        };
    }, [clearPolling]);

    useEffect(() => {
        if (!isWatching) {
            return;
        }
        const viewport = terminalViewportRef.current;
        if (viewport) {
            viewport.focus({ preventScroll: true });
        }
    }, [isWatching]);

    useEffect(() => {
        if (!isWatching) {
            return;
        }

        const handleKeyDown = (event) => {
            if (event.ctrlKey && (event.key === 'c' || event.key === 'C')) {
                event.preventDefault();
                stopWatching('^C');
            }
        };

        // Global listener so the user can press Ctrl+C while focused elsewhere.
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [isWatching, stopWatching]);

    useEffect(() => {
        if (!targetAlias) {
            return;
        }
        if (!availableTargets.some((item) => item.alias === targetAlias)) {
            setTargetAlias('');
        }
    }, [availableTargets, targetAlias]);

    useEffect(() => {
        if (!sourceAlias) {
            setTargetAlias('');
        }
    }, [sourceAlias]);

    useEffect(() => {
        if (!isWatching) {
            return;
        }
        const pair = watchingPairRef.current;
        if (pair.source !== sourceAlias || pair.target !== targetAlias) {
            stopWatching('# Selection changed, stop watching');
        }
    }, [isWatching, sourceAlias, stopWatching, targetAlias]);

    const renderAliasCard = (label, details) => {
        const meta = [];
        if (details?.url) {
            meta.push({ label: 'URL', value: details.url });
        }
        const siteInfo = details?.site || null;
        if (siteInfo) {
            if (typeof siteInfo.name === 'string' && siteInfo.name.trim().length > 0) {
                meta.push({ label: 'Site', value: siteInfo.name });
            }
            if (typeof siteInfo.endpoint === 'string' && siteInfo.endpoint.trim().length > 0) {
                meta.push({ label: 'Endpoint', value: siteInfo.endpoint });
            }
            if (typeof siteInfo.status === 'string' && siteInfo.status.trim().length > 0) {
                meta.push({ label: 'Status', value: siteInfo.status });
            }
            if (typeof siteInfo.state === 'string' && siteInfo.state.trim().length > 0) {
                meta.push({ label: 'State', value: siteInfo.state });
            }
            if (typeof siteInfo.syncStatus === 'string' && siteInfo.syncStatus.trim().length > 0) {
                meta.push({ label: 'Sync', value: siteInfo.syncStatus });
            }
        }
        if (typeof details?.clusterName === 'string' && details.clusterName.trim().length > 0) {
            meta.push({ label: 'Cluster', value: details.clusterName });
        }
        if (typeof details?.clusterId === 'string' && details.clusterId.trim().length > 0) {
            meta.push({ label: 'Cluster ID', value: details.clusterId });
        }
        if (typeof details?.peerCount === 'number' && Number.isFinite(details.peerCount)) {
            meta.push({ label: 'Peer Count', value: String(details.peerCount) });
        }

        return (
            <div
                style={{
                    border: '1px solid #e5e7eb',
                    borderRadius: '6px',
                    padding: '12px',
                    backgroundColor: '#ffffff',
                }}
            >
                <div
                    style={{
                        fontSize: '11px',
                        color: '#6b7280',
                        textTransform: 'uppercase',
                        letterSpacing: '0.05em',
                        marginBottom: '6px',
                    }}
                >
                    {label}
                </div>
                <div style={{ fontSize: '14px', fontWeight: 600, color: '#111827' }}>
                    {details ? details.alias : t('resync_alias_unselected', 'Not selected')}
                </div>
                {details && meta.length > 0 && (
                    <div style={{ marginTop: '6px', fontSize: '12px', color: '#4b5563', display: 'grid', gap: '4px' }}>
                        {meta.map(({ label: metaLabel, value }) => (
                            <div key={`${details.alias}-${metaLabel}`}>
                                {metaLabel}: <span style={{ color: '#1f2937' }}>{value}</span>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        );
    };

    const disableActions = !hasReplication || !sourceAlias || !targetAlias;

    return (
        <div className="card">
            <div className="card-header">
                <h3 className="card-title">
                    {t('resync_terminal_title', 'Replication Resync Status')}
                </h3>
                <p style={{ margin: '8px 0 0 0', fontSize: '14px', color: '#6c757d' }}>
                    {t(
                        'resync_terminal_description',
                        'Watch live output from `mc admin replicate resync status` and manage active resync sessions.'
                    )}
                </p>
            </div>

            <div style={{ padding: '20px', display: 'grid', gap: '20px' }}>
                {!hasReplication && (
                    <div
                        style={{
                            border: '1px solid #ffeaa7',
                            backgroundColor: '#fff9e6',
                            padding: '16px',
                            borderRadius: '8px',
                            color: '#7c5400',
                            fontSize: '14px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px'
                        }}
                    >
                        <AlertTriangle size={16} />
                        <span>{t('resync_requires_replication', 'Configure site replication to enable resync operations.')}</span>
                    </div>
                )}

                {optionsError && (
                    <div
                        style={{
                            border: '1px solid #f5c6cb',
                            backgroundColor: '#f8d7da',
                            color: '#721c24',
                            padding: '12px 16px',
                            borderRadius: '6px',
                            fontSize: '14px',
                        }}
                    >
                        {optionsError}
                    </div>
                )}

                <div
                    style={{
                        display: 'grid',
                        gap: '16px',
                        gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
                        alignItems: 'end',
                    }}
                >
                    <div>
                        <label style={{ display: 'block', fontSize: '13px', color: '#6c757d', marginBottom: '4px' }}>
                            {t('resync_from', 'Resync From')}
                        </label>
                        <select
                            value={sourceAlias}
                            onChange={(event) => setSourceAlias(event.target.value)}
                            disabled={loadingOptions || !hasReplication}
                            style={{
                                width: '100%',
                                padding: '8px 12px',
                                borderRadius: '6px',
                                border: '1px solid #ced4da',
                                fontSize: '14px',
                            }}
                        >
                            <option value="">
                                {loadingOptions
                                    ? t('loading_resync_options', 'Loading options...')
                                    : t('select_source_alias', 'Select source alias')}
                            </option>
                            {availableSources.map((alias) => (
                                <option key={alias.alias} value={alias.alias}>
                                    {alias.alias}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div>
                        <label style={{ display: 'block', fontSize: '13px', color: '#6c757d', marginBottom: '4px' }}>
                            {t('resync_to', 'Resync To')}
                        </label>
                        <select
                            value={targetAlias}
                            onChange={(event) => setTargetAlias(event.target.value)}
                            disabled={loadingOptions || !hasReplication || !sourceAlias}
                            style={{
                                width: '100%',
                                padding: '8px 12px',
                                borderRadius: '6px',
                                border: '1px solid #ced4da',
                                fontSize: '14px',
                            }}
                        >
                            <option value="">
                                {!sourceAlias
                                    ? t('select_source_first', 'Select a source alias first')
                                    : availableTargets.length === 0
                                    ? t('no_target_available', 'No peer targets available')
                                    : t('select_target_alias', 'Select target alias')}
                            </option>
                            {availableTargets.map((alias) => (
                                <option key={alias.alias} value={alias.alias}>
                                    {alias.alias}
                                </option>
                            ))}
                        </select>
                    </div>

                    <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
                        <button
                            className="btn btn-primary"
                            onClick={handleStartResync}
                            disabled={disableActions || isStarting}
                            style={{ display: 'inline-flex', alignItems: 'center', gap: '6px' }}
                        >
                            <Play size={16} />
                            {isStarting
                                ? t('starting_resync', 'Starting...')
                                : t('trigger_resync', 'Trigger Resync')}
                        </button>

                        <button
                            className="btn btn-secondary"
                            onClick={handleWatchOnly}
                            disabled={disableActions || isWatching}
                            style={{ display: 'inline-flex', alignItems: 'center', gap: '6px' }}
                        >
                            <Terminal size={16} />
                            {t('watch_status', 'Watch Status')}
                        </button>

                        <button
                            className="btn btn-outline-secondary"
                            onClick={() => stopWatching('# Stream stopped')}
                            disabled={!isWatching}
                            style={{ display: 'inline-flex', alignItems: 'center', gap: '6px' }}
                        >
                            <Square size={16} />
                            {t('stop_watching', 'Stop')}
                        </button>

                        <button
                            className="btn btn-light"
                            onClick={clearTerminal}
                            style={{ display: 'inline-flex', alignItems: 'center', gap: '6px' }}
                        >
                            <Eraser size={16} />
                            {t('clear_console', 'Clear')}
                        </button>
                    </div>
                </div>

                {(startError || statusError) && (
                    <div
                        style={{
                            border: '1px solid #f5c6cb',
                            backgroundColor: '#f8d7da',
                            color: '#721c24',
                            padding: '12px 16px',
                            borderRadius: '6px',
                            fontSize: '14px',
                        }}
                    >
                        {startError || statusError}
                    </div>
                )}

                <div
                    style={{
                        display: 'grid',
                        gap: '20px',
                        gridTemplateColumns: 'minmax(0, 1.75fr) minmax(0, 1fr)',
                        alignItems: 'stretch',
                    }}
                >
                    <div
                        style={{
                            border: '1px solid #1f2937',
                            backgroundColor: '#0b0c10',
                            borderRadius: '8px',
                            padding: '12px',
                            minHeight: 0,
                            height: '100%',
                            display: 'flex',
                            flexDirection: 'column',
                        }}
                    >
                        <div
                            style={{
                                display: 'flex',
                                justifyContent: 'space-between',
                                alignItems: 'center',
                                marginBottom: '8px',
                                color: '#9ca3af',
                                fontSize: '12px',
                            }}
                        >
                            <span>
                                {isWatching
                                    ? t('stream_live', 'Streaming live output...')
                                    : t('stream_idle', 'Stream idle')}
                            </span>
                            {lastUpdated && (
                                <span>
                                    {t('last_updated', 'Last updated')}:&nbsp;
                                    {lastUpdated.toLocaleTimeString()}
                                </span>
                            )}
                        </div>

                        <div
                            ref={terminalViewportRef}
                            tabIndex={-1}
                            style={{
                                flex: 1,
                                overflowY: 'auto',
                                fontFamily: 'Menlo, Consolas, monospace',
                                fontSize: '13px',
                                color: '#e5e7eb',
                                backgroundColor: 'rgba(255,255,255,0.02)',
                                padding: '8px',
                                borderRadius: '4px',
                                lineHeight: 1.35,
                                outline: 'none',
                                whiteSpace: 'pre-wrap',
                                maxHeight: '460px',
                            }}
                        >
                            {combinedTerminalLines.map((line, index) => (
                                <div key={`${index}-${line}`}>{line || ' '}</div>
                            ))}
                        </div>
                    </div>

                    <div
                        style={{
                            border: '1px solid #e5e7eb',
                            borderRadius: '8px',
                            padding: '16px',
                            backgroundColor: '#f8fafc',
                            minHeight: 0,
                            display: 'flex',
                            flexDirection: 'column',
                            gap: '12px',
                            height: '100%',
                        }}
                    >
                        <h4 style={{ margin: '0', fontSize: '15px', color: '#1f2937' }}>
                            {t('resync_session_details', 'Session Details')}
                        </h4>
                        <div
                            style={{
                                display: 'grid',
                                gap: '12px',
                                gridAutoRows: 'minmax(0, auto)',
                            }}
                        >
                            {renderAliasCard(t('resync_from', 'Resync From'), sourceDetails)}
                            {renderAliasCard(t('resync_to', 'Resync To'), targetDetails)}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default SiteOperations;