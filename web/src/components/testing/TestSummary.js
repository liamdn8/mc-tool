import React from 'react';
import { CheckCircle2, XCircle, Clock, Upload, AlertCircle, Loader } from 'lucide-react';

const formatDuration = (ms) => {
    if (!ms) return '0s';
    const seconds = Math.floor(ms / 1000000000); // ns to seconds
    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${minutes}m ${secs}s`;
};

const formatBytes = (bytes) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
};

const TestSummary = ({ result }) => {
    if (!result) return null;

    const { summary, total_duration, errors } = result;

    // Extract round-by-round details from object_results if available
    const roundGroups = {};
    if (result.object_results) {
        result.object_results.forEach(obj => {
            if (obj.round_number > 0) {
                if (!roundGroups[obj.round_number]) {
                    roundGroups[obj.round_number] = {
                        objectsUploaded: 0,
                        successfulUploads: 0,
                        failedUploads: 0,
                        totalDuration: 0
                    };
                }
                roundGroups[obj.round_number].objectsUploaded++;
                if (obj.success) {
                    roundGroups[obj.round_number].successfulUploads++;
                } else {
                    roundGroups[obj.round_number].failedUploads++;
                }
                roundGroups[obj.round_number].totalDuration += obj.duration;
            }
        });
    }

    const totalRounds = Object.keys(roundGroups).length;
    const hasRounds = totalRounds > 0;

    return (
        <div className="space-y-4">
            {/* Test Summary - Always on top, same as TestProgress */}
            <div className="bg-white border border-gray-300 rounded-lg p-4" style={{ marginBottom: '24px' }}>
                <h3 className="text-lg font-semibold text-gray-900 mb-4" style={{ marginBottom: '24px' }}>Test Summary</h3>
                
                <div className="stats-grid" style={{ marginBottom: 0 }}>
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--primary-color)' }}>
                            {summary.total_uploads}
                        </div>
                        <div className="stat-label">Total Uploads</div>
                    </div>
                    
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--success-color)' }}>
                            {summary.successful_uploads}
                        </div>
                        <div className="stat-label">Successful</div>
                    </div>
                    
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--danger-color)' }}>
                            {summary.failed_uploads}
                        </div>
                        <div className="stat-label">Failed</div>
                    </div>
                    
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--primary-color)' }}>
                            {formatDuration(total_duration)}
                        </div>
                        <div className="stat-label">Total Duration</div>
                    </div>
                    
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--primary-color)' }}>
                            {formatBytes(summary.total_data_uploaded)}
                        </div>
                        <div className="stat-label">Data Uploaded</div>
                    </div>
                </div>
            </div>

            {/* Round Details - Table Format (for interval mode) */}
            {hasRounds && totalRounds > 1 && (
                <div className="bg-white border border-gray-300 rounded-lg p-4">
                    <h3 className="text-lg font-semibold text-gray-900 mb-4">Round-by-Round Status</h3>
                    
                    <div style={{ overflowX: 'auto' }}>
                        <table style={{ 
                            width: '100%', 
                            borderCollapse: 'separate',
                            borderSpacing: 0
                        }}>
                            <thead>
                                <tr style={{ 
                                    backgroundColor: 'rgb(249, 250, 251)',
                                    borderBottom: '2px solid rgb(229, 231, 235)'
                                }}>
                                    <th style={{ 
                                        padding: '12px 16px', 
                                        textAlign: 'left',
                                        fontWeight: 600,
                                        fontSize: '14px',
                                        color: 'rgb(55, 65, 81)',
                                        width: '120px'
                                    }}>
                                        Round
                                    </th>
                                    <th style={{ 
                                        padding: '12px 16px', 
                                        textAlign: 'left',
                                        fontWeight: 600,
                                        fontSize: '14px',
                                        color: 'rgb(55, 65, 81)'
                                    }}>
                                        Status
                                    </th>
                                    <th style={{ 
                                        padding: '12px 16px', 
                                        textAlign: 'right',
                                        fontWeight: 600,
                                        fontSize: '14px',
                                        color: 'rgb(55, 65, 81)',
                                        width: '140px'
                                    }}>
                                        Duration
                                    </th>
                                </tr>
                            </thead>
                            <tbody>
                                {Object.keys(roundGroups).sort((a, b) => parseInt(a) - parseInt(b)).map((roundNum) => {
                                    const roundDetail = roundGroups[roundNum];
                                    const hasFailed = roundDetail.failedUploads > 0;
                                    
                                    return (
                                        <tr 
                                            key={roundNum}
                                            style={{ 
                                                borderBottom: '1px solid rgb(229, 231, 235)',
                                                backgroundColor: 'white'
                                            }}
                                        >
                                            {/* Round column */}
                                            <td style={{ 
                                                padding: '12px 16px',
                                                fontWeight: 500,
                                                fontSize: '14px',
                                                color: 'rgb(55, 65, 81)'
                                            }}>
                                                <div className="flex items-center gap-2">
                                                    {hasFailed ? (
                                                        <XCircle size={16} className="text-red-600" />
                                                    ) : (
                                                        <CheckCircle2 size={16} className="text-green-600" />
                                                    )}
                                                    <span>Round {roundNum}</span>
                                                </div>
                                            </td>
                                            
                                            {/* Status tags column */}
                                            <td style={{ padding: '12px 16px' }}>
                                                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px', alignItems: 'center' }}>
                                                    <span className="badge" style={{ 
                                                        display: 'inline-flex',
                                                        alignItems: 'center',
                                                        padding: '4px 12px',
                                                        backgroundColor: '#dbeafe',
                                                        color: '#1e40af',
                                                        border: '1px solid #93c5fd',
                                                        borderRadius: '4px',
                                                        fontSize: '13px',
                                                        fontWeight: 500
                                                    }}>
                                                        <Upload size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                        {roundDetail.objectsUploaded}
                                                    </span>
                                                    
                                                    <span className="badge badge-success" style={{ 
                                                        display: 'inline-flex',
                                                        alignItems: 'center',
                                                        padding: '4px 12px',
                                                        borderRadius: '4px',
                                                        fontSize: '13px',
                                                        fontWeight: 500
                                                    }}>
                                                        <CheckCircle2 size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                        {roundDetail.successfulUploads}
                                                    </span>
                                                    
                                                    {roundDetail.failedUploads > 0 && (
                                                        <span className="badge badge-danger" style={{ 
                                                            display: 'inline-flex',
                                                            alignItems: 'center',
                                                            padding: '4px 12px',
                                                            borderRadius: '4px',
                                                            fontSize: '13px',
                                                            fontWeight: 500
                                                        }}>
                                                            <XCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                            {roundDetail.failedUploads}
                                                        </span>
                                                    )}
                                                </div>
                                            </td>
                                            
                                            {/* Duration column */}
                                            <td style={{ 
                                                padding: '12px 16px',
                                                textAlign: 'right',
                                                fontSize: '14px',
                                                color: 'rgb(107, 114, 128)'
                                            }}>
                                                {formatDuration(roundDetail.totalDuration / roundDetail.objectsUploaded)}/obj
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    </div>
                </div>
            )}

            {/* Errors */}
            {errors && errors.length > 0 && (
                <div className="bg-white border border-red-300 rounded-lg p-4">
                    <h3 className="text-lg font-semibold text-red-900 mb-3">Errors</h3>
                    <ul className="space-y-2">
                        {errors.map((err, idx) => (
                            <li key={idx} className="text-sm text-red-700 bg-red-50 p-3 rounded border border-red-200">
                                {err}
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    );
};

export default TestSummary;
