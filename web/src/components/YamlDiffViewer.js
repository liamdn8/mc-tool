import React, { useState } from 'react';
import { X, Copy, Check, ChevronUp, ChevronDown } from 'lucide-react';

const YamlDiffViewer = ({ baseline, target, resourceName, baselineLabel, targetLabel, onClose }) => {
    const [copied, setCopied] = useState(false);
    const [copiedSide, setCopiedSide] = useState(null);
    const [collapsedHunks, setCollapsedHunks] = useState({});
    const [viewMode, setViewMode] = useState('differences'); // 'full' or 'differences'

    const copyToClipboard = (text, side) => {
        navigator.clipboard.writeText(text);
        setCopiedSide(side);
        setCopied(true);
        setTimeout(() => {
            setCopied(false);
            setCopiedSide(null);
        }, 2000);
    };

    const toggleHunk = (hunkIndex) => {
        setCollapsedHunks(prev => ({
            ...prev,
            [hunkIndex]: !prev[hunkIndex]
        }));
    };

    // Myers diff algorithm - proper LCS-based diff
    const computeDiff = (baselineLines, targetLines) => {
        const n = baselineLines.length;
        const m = targetLines.length;
        const max = n + m;
        const v = {};
        const trace = [];

        // Find the shortest edit script
        for (let d = 0; d <= max; d++) {
            const current = { ...v };
            trace.push(current);

            for (let k = -d; k <= d; k += 2) {
                let x;
                if (k === -d || (k !== d && v[k - 1] < v[k + 1])) {
                    x = v[k + 1] || 0;
                } else {
                    x = (v[k - 1] || 0) + 1;
                }
                let y = x - k;

                while (x < n && y < m && baselineLines[x] === targetLines[y]) {
                    x++;
                    y++;
                }

                v[k] = x;

                if (x >= n && y >= m) {
                    return backtrack(trace, baselineLines, targetLines, d);
                }
            }
        }

        return [];
    };

    const backtrack = (trace, baselineLines, targetLines, d) => {
        let x = baselineLines.length;
        let y = targetLines.length;
        const changes = [];

        for (let i = d; i >= 0; i--) {
            const v = trace[i];
            const k = x - y;

            let prevK;
            if (k === -i || (k !== i && (v[k - 1] || 0) < (v[k + 1] || 0))) {
                prevK = k + 1;
            } else {
                prevK = k - 1;
            }

            const prevX = v[prevK] || 0;
            const prevY = prevX - prevK;

            while (x > prevX && y > prevY) {
                changes.unshift({
                    type: 'normal',
                    baseIdx: x - 1,
                    targetIdx: y - 1,
                    baseLine: baselineLines[x - 1],
                    targetLine: targetLines[y - 1]
                });
                x--;
                y--;
            }

            if (i > 0) {
                if (x > prevX) {
                    changes.unshift({
                        type: 'delete',
                        baseIdx: x - 1,
                        targetIdx: null,
                        baseLine: baselineLines[x - 1],
                        targetLine: ''
                    });
                    x--;
                } else {
                    changes.unshift({
                        type: 'insert',
                        baseIdx: null,
                        targetIdx: y - 1,
                        baseLine: '',
                        targetLine: targetLines[y - 1]
                    });
                    y--;
                }
            }
        }

        return changes;
    };

    // Compute diff hunks (similar to git diff)
    const computeDiffHunks = () => {
        const baselineLines = (baseline || '').split('\n');
        const targetLines = (target || '').split('\n');
        
        // Use Myers diff algorithm
        const changes = computeDiff(baselineLines, targetLines);

        // Group changes into hunks (groups of changes with context)
        const hunks = [];
        const contextLines = 3;
        let currentHunk = null;

        for (let i = 0; i < changes.length; i++) {
            const change = changes[i];
            
            if (change.type !== 'normal') {
                // Start a new hunk if needed
                if (!currentHunk) {
                    const startIdx = Math.max(0, i - contextLines);
                    currentHunk = {
                        startBase: changes[startIdx].baseIdx !== null ? changes[startIdx].baseIdx : 0,
                        startTarget: changes[startIdx].targetIdx !== null ? changes[startIdx].targetIdx : 0,
                        changes: []
                    };
                    
                    // Add context before
                    for (let j = startIdx; j < i; j++) {
                        if (changes[j].type === 'normal') {
                            currentHunk.changes.push(changes[j]);
                        }
                    }
                }
                
                currentHunk.changes.push(change);
            } else if (currentHunk) {
                // Add context after change
                currentHunk.changes.push(change);
                
                // Check if we should close this hunk
                let normalCount = 0;
                for (let j = i; j < changes.length && changes[j].type === 'normal'; j++) {
                    normalCount++;
                }
                
                if (normalCount > contextLines * 2) {
                    // Add remaining context lines
                    for (let j = i + 1; j < Math.min(changes.length, i + contextLines); j++) {
                        if (changes[j].type === 'normal') {
                            currentHunk.changes.push(changes[j]);
                        }
                    }
                    // Close current hunk
                    hunks.push(currentHunk);
                    currentHunk = null;
                    i += contextLines - 1;
                }
            }
        }
        
        // Close last hunk if open
        if (currentHunk) {
            hunks.push(currentHunk);
        }

        return hunks;
    };

    const renderDiffHunks = () => {
        const baselineLines = (baseline || '').split('\n');
        const targetLines = (target || '').split('\n');
        
        // In "full" mode, show all lines
        if (viewMode === 'full') {
            const changes = computeDiff(baselineLines, targetLines);
            
            if (changes.length === 0) {
                return (
                    <div style={{ padding: '40px', textAlign: 'center', color: 'var(--text-secondary)' }}>
                        No content
                    </div>
                );
            }
            
            // Create a single hunk with all changes
            return (
                <div style={{ marginBottom: '16px' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                        <tbody>
                            {changes.map((change, idx) => (
                                <tr
                                    key={idx}
                                    className={`diff-line diff-line-${change.type}`}
                                    style={{
                                        backgroundColor: 
                                            change.type === 'insert' ? '#e6ffec' :
                                            change.type === 'delete' ? '#ffebe9' :
                                            'transparent'
                                    }}
                                >
                                    {/* Baseline side */}
                                    <td
                                        className="diff-gutter"
                                        style={{
                                            width: '40px',
                                            textAlign: 'right',
                                            paddingRight: '12px',
                                            color: 'var(--text-secondary)',
                                            fontSize: '12px',
                                            borderRight: '1px solid var(--border-color)',
                                            userSelect: 'none',
                                            backgroundColor: change.type === 'delete' ? '#ffd7d5' : 'transparent'
                                        }}
                                    >
                                        {change.baseIdx !== null ? change.baseIdx + 1 : ''}
                                    </td>
                                    <td
                                        className="diff-code"
                                        style={{
                                            width: '50%',
                                            padding: '2px 12px',
                                            fontFamily: 'monospace',
                                            fontSize: '13px',
                                            whiteSpace: 'pre',
                                            borderRight: '2px solid var(--border-color)',
                                            color: change.type === 'delete' ? '#b71c1c' : 'var(--text-primary)'
                                        }}
                                    >
                                        {change.type === 'delete' && <span style={{ marginRight: '4px', color: '#d32f2f' }}>-</span>}
                                        {change.baseLine}
                                    </td>
                                    
                                    {/* Target side */}
                                    <td
                                        className="diff-gutter"
                                        style={{
                                            width: '40px',
                                            textAlign: 'right',
                                            paddingRight: '12px',
                                            color: 'var(--text-secondary)',
                                            fontSize: '12px',
                                            borderRight: '1px solid var(--border-color)',
                                            userSelect: 'none',
                                            backgroundColor: change.type === 'insert' ? '#c3e6cb' : 'transparent'
                                        }}
                                    >
                                        {change.targetIdx !== null ? change.targetIdx + 1 : ''}
                                    </td>
                                    <td
                                        className="diff-code"
                                        style={{
                                            width: '50%',
                                            padding: '2px 12px',
                                            fontFamily: 'monospace',
                                            fontSize: '13px',
                                            whiteSpace: 'pre',
                                            color: change.type === 'insert' ? '#1b5e20' : 'var(--text-primary)'
                                        }}
                                    >
                                        {change.type === 'insert' && <span style={{ marginRight: '4px', color: '#388e3c' }}>+</span>}
                                        {change.targetLine}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            );
        }
        
        // In "differences" mode, show only hunks with changes
        const hunks = computeDiffHunks();
        
        if (hunks.length === 0) {
            return (
                <div style={{ padding: '40px', textAlign: 'center', color: 'var(--text-secondary)' }}>
                    No differences found
                </div>
            );
        }

        return hunks.map((hunk, hunkIdx) => {
            const isCollapsed = collapsedHunks[hunkIdx];
            
            return (
                <div key={hunkIdx} style={{ marginBottom: '16px' }}>
                    {/* Hunk Header */}
                    <div
                        onClick={() => toggleHunk(hunkIdx)}
                        style={{
                            padding: '8px 12px',
                            backgroundColor: 'var(--bg-secondary)',
                            borderBottom: '1px solid var(--border-color)',
                            cursor: 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            fontFamily: 'monospace',
                            fontSize: '12px',
                            color: 'var(--text-secondary)',
                            userSelect: 'none'
                        }}
                    >
                        {isCollapsed ? <ChevronDown size={16} /> : <ChevronUp size={16} />}
                        <span>Lines {hunk.startBase + 1} - {hunk.startBase + hunk.changes.length}</span>
                    </div>
                    
                    {/* Hunk Content */}
                    {!isCollapsed && (
                        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                            <tbody>
                                {hunk.changes.map((change, idx) => (
                                    <tr
                                        key={idx}
                                        className={`diff-line diff-line-${change.type}`}
                                        style={{
                                            backgroundColor: 
                                                change.type === 'insert' ? '#e6ffec' :
                                                change.type === 'delete' ? '#ffebe9' :
                                                'transparent'
                                        }}
                                    >
                                        {/* Baseline side */}
                                        <td
                                            className="diff-gutter"
                                            style={{
                                                width: '40px',
                                                textAlign: 'right',
                                                paddingRight: '12px',
                                                color: 'var(--text-secondary)',
                                                fontSize: '12px',
                                                borderRight: '1px solid var(--border-color)',
                                                userSelect: 'none',
                                                backgroundColor: change.type === 'delete' ? '#ffd7d5' : 'transparent'
                                            }}
                                        >
                                            {change.baseIdx !== null ? change.baseIdx + 1 : ''}
                                        </td>
                                        <td
                                            className="diff-code"
                                            style={{
                                                width: '50%',
                                                padding: '2px 12px',
                                                fontFamily: 'monospace',
                                                fontSize: '13px',
                                                whiteSpace: 'pre',
                                                borderRight: '2px solid var(--border-color)',
                                                color: change.type === 'delete' ? '#b71c1c' : 'var(--text-primary)'
                                            }}
                                        >
                                            {change.type === 'delete' && <span style={{ marginRight: '4px', color: '#d32f2f' }}>-</span>}
                                            {change.baseLine}
                                        </td>
                                        
                                        {/* Target side */}
                                        <td
                                            className="diff-gutter"
                                            style={{
                                                width: '40px',
                                                textAlign: 'right',
                                                paddingRight: '12px',
                                                color: 'var(--text-secondary)',
                                                fontSize: '12px',
                                                borderRight: '1px solid var(--border-color)',
                                                userSelect: 'none',
                                                backgroundColor: change.type === 'insert' ? '#c3e6cb' : 'transparent'
                                            }}
                                        >
                                            {change.targetIdx !== null ? change.targetIdx + 1 : ''}
                                        </td>
                                        <td
                                            className="diff-code"
                                            style={{
                                                width: '50%',
                                                padding: '2px 12px',
                                                fontFamily: 'monospace',
                                                fontSize: '13px',
                                                whiteSpace: 'pre',
                                                color: change.type === 'insert' ? '#1b5e20' : 'var(--text-primary)'
                                            }}
                                        >
                                            {change.type === 'insert' && <span style={{ marginRight: '4px', color: '#388e3c' }}>+</span>}
                                            {change.targetLine}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                </div>
            );
        });
    };

    return (
        <div style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: 'rgba(0, 0, 0, 0.75)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000,
            padding: '20px'
        }}>
            <div style={{
                backgroundColor: 'var(--bg-primary)',
                borderRadius: '8px',
                maxWidth: '95vw',
                maxHeight: '90vh',
                width: '1400px',
                display: 'flex',
                flexDirection: 'column',
                boxShadow: '0 10px 40px rgba(0, 0, 0, 0.3)',
                border: '1px solid var(--border-color)'
            }}>
                {/* Header */}
                <div style={{
                    padding: '20px',
                    borderBottom: '1px solid var(--border-color)',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    backgroundColor: 'var(--bg-primary)'
                }}>
                    <div>
                        <h3 style={{ margin: 0, fontSize: '18px', fontWeight: 'bold', color: 'var(--text-primary)' }}>
                            Configuration Diff: {resourceName}
                        </h3>
                        <p style={{ margin: '4px 0 0 0', fontSize: '14px', color: 'var(--text-secondary)' }}>
                            Side-by-side comparison of normalized YAML configurations
                        </p>
                    </div>
                    <button
                        onClick={onClose}
                        className="btn btn-secondary"
                        style={{ padding: '8px' }}
                    >
                        <X size={20} />
                    </button>
                </div>

                {/* Column Headers */}
                <div style={{
                    display: 'grid',
                    gridTemplateColumns: '40px 1fr 40px 1fr',
                    borderBottom: '1px solid var(--border-color)',
                    backgroundColor: 'var(--bg-secondary)',
                    padding: '12px 0'
                }}>
                    <div></div>
                    <div style={{ 
                        padding: '0 12px',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '4px'
                    }}>
                        <div style={{ 
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center'
                        }}>
                            <div>
                                <div style={{ fontWeight: 'bold', color: 'var(--text-primary)', fontSize: '14px' }}>
                                    Baseline
                                </div>
                                {baselineLabel && (
                                    <div style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '2px' }}>
                                        {baselineLabel}
                                    </div>
                                )}
                            </div>
                            <button
                                onClick={() => copyToClipboard(baseline, 'baseline')}
                                className="btn btn-secondary"
                                style={{ padding: '4px 8px', fontSize: '12px' }}
                            >
                                {copied && copiedSide === 'baseline' ? <Check size={14} /> : <Copy size={14} />}
                                <span style={{ marginLeft: '4px' }}>Copy</span>
                            </button>
                        </div>
                    </div>
                    <div></div>
                    <div style={{ 
                        padding: '0 12px',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '4px'
                    }}>
                        <div style={{ 
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center'
                        }}>
                            <div>
                                <div style={{ fontWeight: 'bold', color: 'var(--text-primary)', fontSize: '14px' }}>
                                    Target
                                </div>
                                {targetLabel && (
                                    <div style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '2px' }}>
                                        {targetLabel}
                                    </div>
                                )}
                            </div>
                            <button
                                onClick={() => copyToClipboard(target, 'target')}
                                className="btn btn-secondary"
                                style={{ padding: '4px 8px', fontSize: '12px' }}
                            >
                                {copied && copiedSide === 'target' ? <Check size={14} /> : <Copy size={14} />}
                                <span style={{ marginLeft: '4px' }}>Copy</span>
                            </button>
                        </div>
                    </div>
                </div>

                {/* Toolbar */}
                <div style={{
                    padding: '12px 16px',
                    borderBottom: '1px solid var(--border-color)',
                    backgroundColor: 'var(--bg-primary)',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '12px'
                }}>
                    <span style={{ fontSize: '13px', color: 'var(--text-secondary)', marginRight: '8px' }}>
                        View:
                    </span>
                    <label style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '6px',
                        cursor: 'pointer',
                        fontSize: '13px',
                        color: 'var(--text-primary)',
                        userSelect: 'none'
                    }}>
                        <input
                            type="radio"
                            name="viewMode"
                            value="differences"
                            checked={viewMode === 'differences'}
                            onChange={(e) => setViewMode(e.target.value)}
                            style={{ cursor: 'pointer' }}
                        />
                        Show only differences
                    </label>
                    <label style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '6px',
                        cursor: 'pointer',
                        fontSize: '13px',
                        color: 'var(--text-primary)',
                        userSelect: 'none'
                    }}>
                        <input
                            type="radio"
                            name="viewMode"
                            value="full"
                            checked={viewMode === 'full'}
                            onChange={(e) => setViewMode(e.target.value)}
                            style={{ cursor: 'pointer' }}
                        />
                        Show full
                    </label>
                </div>

                {/* Diff Content */}
                <div style={{
                    flex: 1,
                    overflow: 'auto',
                    backgroundColor: 'var(--bg-primary)',
                    padding: '16px'
                }}>
                    {renderDiffHunks()}
                </div>

                {/* Footer */}
                <div style={{
                    padding: '16px 20px',
                    borderTop: '1px solid var(--border-color)',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    backgroundColor: 'var(--bg-primary)'
                }}>
                    <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                        <div style={{ display: 'flex', gap: '16px', alignItems: 'center' }}>
                            <div>
                                <span style={{ 
                                    display: 'inline-block',
                                    width: '12px',
                                    height: '12px',
                                    backgroundColor: '#e6ffec',
                                    border: '1px solid #388e3c',
                                    marginRight: '6px',
                                    verticalAlign: 'middle'
                                }}></span>
                                Added lines
                            </div>
                            <div>
                                <span style={{ 
                                    display: 'inline-block',
                                    width: '12px',
                                    height: '12px',
                                    backgroundColor: '#ffebe9',
                                    border: '1px solid #d32f2f',
                                    marginRight: '6px',
                                    verticalAlign: 'middle'
                                }}></span>
                                Removed lines
                            </div>
                        </div>
                    </div>
                    <button onClick={onClose} className="btn btn-primary">
                        Close
                    </button>
                </div>
            </div>
        </div>
    );
};

export default YamlDiffViewer;
