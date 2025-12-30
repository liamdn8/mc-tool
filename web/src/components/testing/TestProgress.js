import React, { useState } from 'react';
import { CheckCircle2, XCircle, Clock, Upload, AlertCircle, Loader, Copy, Check } from 'lucide-react';

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

const TestProgress = ({ status, isComplete, result }) => {
    const [copiedCommand, setCopiedCommand] = useState(null);
    
    if (!status && !result) {
        console.log('TestProgress: No status or result provided');
        return null;
    }
    
    const copyToClipboard = (text, key) => {
        navigator.clipboard.writeText(text).then(() => {
            setCopiedCommand(key);
            setTimeout(() => setCopiedCommand(null), 2000);
        });
    };

    // Use result data if test is complete, otherwise use status
    const displayData = isComplete && result ? {
        running: false,
        progress: 100,
        currentPhase: 'completed',
        completedUploads: result.summary.total_uploads,
        totalUploads: result.summary.total_uploads,
        successfulUploads: result.summary.successful_uploads,
        failedUploads: result.summary.failed_uploads,
        elapsedTime: result.total_duration,
        currentRound: 0,
        totalRounds: 0,
        roundDetails: [],
        dataUploaded: result.summary.total_data_uploaded
    } : status;

    console.log('TestProgress render:', {
        isComplete,
        currentRound: displayData.currentRound,
        totalRounds: displayData.totalRounds,
        roundDetailsCount: displayData.roundDetails?.length || 0,
        progress: displayData.progress
    });

    const { 
        running, 
        progress, 
        currentPhase,
        completedUploads,
        totalUploads,
        successfulUploads,
        failedUploads,
        currentRound,
        totalRounds,
        elapsedTime,
        roundDetails,
        recentUploads,
        dataUploaded
    } = displayData;

    // Extract round groups from result if complete
    // Also track all uploads per object (not just overrides)
    const roundGroups = {};
    const objectUploadTracker = {}; // Track all uploads per object
    
    // Debug: log result structure
    if (isComplete && result) {
        console.log('TestProgress result:', {
            config: result.config,
            configKeys: result.config ? Object.keys(result.config) : [],
            firstObject: result.object_results?.[0]
        });
    }
    
    // Determine if we should show override tracking
    const testConfig = result?.config || status?.config;
    const willHaveOverrides = testConfig && testConfig.OverrideCount > 0;
    const expectedObjectKeys = [];
    
    // If we know the test will have overrides, initialize expected objects
    if (willHaveOverrides && testConfig) {
        const objectPath = testConfig.ObjectPath || '';
        const objectCount = testConfig.OverrideCount || 0; // Use OverrideCount, not ObjectCount!
        // Backend generates override objects: {objectPath}override-obj-{timestamp}-{paddedNum}.bin
        // These are the objects that get uploaded multiple times (creating versions)
        // Note: ObjectCount creates unique objects per round (round-N-obj-...)
        // OverrideCount creates objects uploaded in EVERY round (override-obj-...)
        const pathParts = objectPath.replace(/\/+$/, '').split('/');
        const timestamp = pathParts[pathParts.length - 1];
        
        console.log('Override tracking init:', { objectPath, overrideCount: objectCount, timestamp });
        
        // Generate expected object keys for override objects
        for (let i = 1; i <= objectCount; i++) {
            const paddedNum = String(i).padStart(4, '0');
            const expectedKey = `${objectPath}override-obj-${timestamp}-${paddedNum}.bin`;
            expectedObjectKeys.push(expectedKey);
            console.log('Expected override object:', expectedKey);
            // Initialize tracker for each expected object
            if (!objectUploadTracker[expectedKey]) {
                objectUploadTracker[expectedKey] = {
                    rounds: [],
                    totalVersions: 0,
                    expectedVersions: testConfig.Iterations || testConfig.OverrideCount + 1,
                    alias: testConfig.SiteAlias || '',
                    bucket: testConfig.Bucket || ''
                };
            }
        }
    }
    
    if (isComplete && result && result.object_results) {
        console.log('Processing', result.object_results.length, 'object results');
        result.object_results.forEach(obj => {
            if (obj.round_number > 0) {
                console.log('Object:', obj.object_key, 'Upload:', obj.upload_number, 'Round:', obj.round_number);
                // Track round statistics
                if (!roundGroups[obj.round_number]) {
                    roundGroups[obj.round_number] = {
                        round: obj.round_number,
                        objectsUploaded: 0,
                        successfulUploads: 0,
                        failedUploads: 0,
                        duration: 0,
                        dataUploaded: 0,
                        overrideFiles: []
                    };
                }
                roundGroups[obj.round_number].objectsUploaded++;
                if (obj.success) {
                    roundGroups[obj.round_number].successfulUploads++;
                } else {
                    roundGroups[obj.round_number].failedUploads++;
                }
                roundGroups[obj.round_number].duration += obj.duration;
                roundGroups[obj.round_number].dataUploaded += obj.object_size || 1024;
                
                // Track ALL uploads for each object (including first upload)
                const objectKey = obj.object_key;
                if (!objectUploadTracker[objectKey]) {
                    objectUploadTracker[objectKey] = {
                        rounds: [],
                        totalVersions: 0,
                        expectedVersions: objectKey.includes('override-obj-') 
                            ? (testConfig?.Iterations || testConfig?.OverrideCount + 1) 
                            : undefined,
                        alias: result.config?.SiteAlias || '',
                        bucket: result.config?.Bucket || ''
                    };
                }
                objectUploadTracker[objectKey].rounds.push(obj.round_number);
                objectUploadTracker[objectKey].totalVersions++;
                
                // Mark as override file if uploaded more than once
                if (obj.upload_number > 1) {
                    roundGroups[obj.round_number].overrideFiles.push({
                        key: obj.object_key,
                        uploadNumber: obj.upload_number,
                        alias: result.config?.SiteAlias || '',
                        bucket: result.config?.Bucket || ''
                    });
                }
            }
        });
    }
    
    // Also process status data for ongoing tests
    if (!isComplete && status && status.roundDetails) {
        status.roundDetails.forEach(round => {
            if (!roundGroups[round.round]) {
                roundGroups[round.round] = round;
            }
        });
    }
    
    // Update tracker from recent uploads during ongoing test
    if (!isComplete && status && status.recentUploads) {
        status.recentUploads.forEach(obj => {
            const objectKey = obj.object_key;
            if (objectUploadTracker[objectKey]) {
                // Update existing tracker
                if (!objectUploadTracker[objectKey].rounds.includes(obj.round_number)) {
                    objectUploadTracker[objectKey].rounds.push(obj.round_number);
                }
                objectUploadTracker[objectKey].totalVersions = Math.max(
                    objectUploadTracker[objectKey].totalVersions,
                    obj.upload_number
                );
            }
        });
    }
    
    const completedRounds = Object.values(roundGroups);
    const displayTotalRounds = isComplete ? completedRounds.length : totalRounds;
    const displayRoundDetails = isComplete ? completedRounds : roundDetails;

    return (
        <div className="space-y-4">
            {/* Test Summary - Always on top, updates in real-time */}
            <div className="bg-white border border-gray-300 rounded-lg p-4" style={{ marginBottom: '24px' }} id="test-summary">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Test Summary</h3>
                
                {/* Stats Grid */}
                <div className="stats-grid" style={{ marginBottom: '24px' }}>
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--primary-color)' }}>
                            {totalUploads || 0}
                        </div>
                        <div className="stat-label">Total Uploads</div>
                    </div>
                    
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--success-color)' }}>
                            {successfulUploads || 0}
                        </div>
                        <div className="stat-label">Successful</div>
                    </div>
                    
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--danger-color)' }}>
                            {failedUploads || 0}
                        </div>
                        <div className="stat-label">Failed</div>
                    </div>
                    
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--primary-color)' }}>
                            {formatDuration(elapsedTime)}
                        </div>
                        <div className="stat-label">Total Duration</div>
                    </div>
                    
                    <div className="stat-card">
                        <div className="stat-value" style={{ color: 'var(--primary-color)' }}>
                            {formatBytes(dataUploaded || (successfulUploads || 0) * 1024)}
                        </div>
                        <div className="stat-label">Data Uploaded</div>
                    </div>
                </div>

                {/* Progress Bar */}
                <div style={{ marginBottom: '16px' }}>
                    <div className="flex justify-between text-sm text-gray-600 mb-1">
                        <span>Overall Progress</span>
                        <span>{progress}%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                        <div 
                            className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                            style={{ width: `${progress}%` }}
                        />
                    </div>
                </div>

                {/* Current Round Info */}
                {!isComplete && displayTotalRounds > 1 && (
                    <div className="p-3 bg-gray-50 border border-gray-200 rounded-md" style={{ marginBottom: '12px' }}>
                        <div className="text-sm text-gray-700">
                            <span className="font-semibold">Current Round:</span> {currentRound} / {displayTotalRounds}
                        </div>
                    </div>
                )}

                {/* Current Phase */}
                <div className="flex items-center space-x-2 text-sm text-gray-600">
                    {running && (
                        <>
                            <div className="animate-spin h-4 w-4 border-2 border-blue-600 border-t-transparent rounded-full"></div>
                            <span className="capitalize">{currentPhase || 'Processing'}...</span>
                        </>
                    )}
                    {!running && progress === 100 && (
                        <span className="text-green-600 font-semibold">✓ Completed</span>
                    )}
                </div>
            </div>

            {/* Round Details - Table Format (for interval mode) */}
            {displayTotalRounds > 1 && (
                <div className="bg-white border border-gray-300 rounded-lg p-4" style={{ marginBottom: '24px' }} id="round-by-round-status">
                    <h3 className="text-lg font-semibold text-gray-900 mb-4">Round-by-Round Status</h3>
                    
                    <div className="table-container">
                        <div style={{ overflowX: 'auto' }}>
                        <table className="table">
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
                                        Data Uploaded
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
                                {(isComplete ? completedRounds : Array.from({ length: displayTotalRounds }, (_, idx) => {
                                    const roundNum = idx + 1;
                                    return displayRoundDetails?.find(r => r.round === roundNum) || { round: roundNum };
                                })).map((roundDetail, idx) => {
                                    const roundNum = isComplete ? roundDetail.round : (idx + 1);
                                    const isRoundComplete = isComplete || (!!roundDetail && roundDetail.duration > 0);
                                    const isRunning = !isComplete && currentRound === roundNum && !isRoundComplete;
                                    const isPending = !isComplete && roundNum > currentRound;
                                    const hasFailed = roundDetail && roundDetail.failedUploads > 0;
                                    const hasOverrides = roundDetail && roundDetail.overrideFiles && roundDetail.overrideFiles.length > 0;
                                    
                                    // Debug log for first render and when state changes
                                    if (idx === 0) {
                                        console.log(`Table render - Round ${roundNum}:`, {
                                            isPending,
                                            isRunning,
                                            isComplete,
                                            currentRound,
                                            roundDetail
                                        });
                                    }
                                    
                                    return (
                                        <tr 
                                            key={roundNum}
                                            style={{ 
                                                borderBottom: '1px solid rgb(229, 231, 235)',
                                                backgroundColor: isRunning ? 'rgb(239, 246, 255)' : 'white',
                                                transition: 'background-color 0.3s ease'
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
                                                    {isComplete && !hasFailed && <CheckCircle2 size={16} className="text-green-600" />}
                                                    {isComplete && hasFailed && <XCircle size={16} className="text-red-600" />}
                                                    {isRunning && <Loader size={16} className="text-blue-600 spin" />}
                                                    {isPending && <Clock size={16} className="text-gray-400" />}
                                                    <span>Round {roundNum}</span>
                                                </div>
                                            </td>
                                            
                                            {/* Status tags column */}
                                            <td style={{ padding: '12px 16px' }}>
                                                {isPending && (
                                                    <span className="badge" style={{ 
                                                        display: 'inline-flex',
                                                        alignItems: 'center',
                                                        padding: '4px 12px',
                                                        backgroundColor: 'var(--bg-secondary)',
                                                        color: 'var(--text-muted)',
                                                        border: '1px solid var(--border-color)',
                                                        borderRadius: '4px',
                                                        fontSize: '13px',
                                                        fontWeight: 500
                                                    }}>
                                                        <Clock size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                        Pending
                                                    </span>
                                                )}
                                                
                                                {isRunning && (
                                                    <span className="badge" style={{ 
                                                        display: 'inline-flex',
                                                        alignItems: 'center',
                                                        padding: '4px 12px',
                                                        backgroundColor: '#dbeafe',
                                                        color: '#1e40af',
                                                        // border: '1px solid #93c5fd',
                                                        borderRadius: '4px',
                                                        fontSize: '13px',
                                                        fontWeight: 500
                                                    }}>
                                                        <Loader size={14} className="spin" style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                        Running...
                                                    </span>
                                                )}
                                                
                                                {isRoundComplete && (
                                                    <div>
                                                        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px', alignItems: 'center' }}>
                                                            <span className="badge" style={{ 
                                                                display: 'inline-flex',
                                                                alignItems: 'center',
                                                                padding: '4px 12px',
                                                                backgroundColor: '#dbeafe',
                                                                color: '#1e40af',
                                                                // border: '1px solid #93c5fd',
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
                                                    </div>
                                                )}
                                            </td>
                                            
                                            {/* Data Uploaded column */}
                                            <td style={{ 
                                                padding: '12px 16px',
                                                textAlign: 'right',
                                                fontSize: '14px',
                                                color: 'rgb(107, 114, 128)'
                                            }}>
                                                {isRoundComplete && roundDetail.dataUploaded ? formatBytes(roundDetail.dataUploaded) : '-'}
                                            </td>
                                            
                                            {/* Duration column */}
                                            <td style={{ 
                                                padding: '12px 16px',
                                                textAlign: 'right',
                                                fontSize: '14px',
                                                color: 'rgb(107, 114, 128)'
                                            }}>
                                                {isRoundComplete ? formatDuration(roundDetail.duration) : '-'}
                                            </td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                        </div>
                    </div>
                </div>
            )}

            {/* Overridden Files - Show list of override objects */}
            {willHaveOverrides && Object.keys(objectUploadTracker).length > 0 && (
                <div className="bg-white border border-gray-300 rounded-lg p-4" id="overridden-files">
                    <h3 className="text-lg font-semibold text-gray-900 mb-4">
                        {isComplete ? 'Overridden Files' : 'Override Tracking'}
                    </h3>
                    
                    <div style={{ marginBottom: '12px', padding: '12px', backgroundColor: isComplete ? '#fef3c7' : '#dbeafe', border: `1px solid ${isComplete ? '#fbbf24' : '#3b82f6'}`, borderRadius: '6px' }}>
                        <div style={{ fontSize: '14px', color: isComplete ? '#78350f' : '#1e40af', marginBottom: '8px' }}>
                            {isComplete 
                                ? '📝 These files were uploaded multiple times across different rounds, creating multiple versions.'
                                : '🔄 Tracking file versions as they are uploaded. Each file will be overridden multiple times.'}
                        </div>
                        <div style={{ fontSize: '13px', color: '#6b7280', marginTop: '8px' }}>
                            Debug: Tracking {Object.keys(objectUploadTracker).length} objects
                        </div>
                    </div>

                    <div className="table-container">
                        <div style={{ overflowX: 'auto' }}>
                        <table className="table">
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
                                        width: '100px'
                                    }}>
                                        Alias
                                    </th>
                                    <th style={{ 
                                        padding: '12px 16px', 
                                        textAlign: 'left',
                                        fontWeight: 600,
                                        fontSize: '14px',
                                        color: 'rgb(55, 65, 81)',
                                        width: '120px'
                                    }}>
                                        Bucket
                                    </th>
                                    <th style={{ 
                                        padding: '12px 16px', 
                                        textAlign: 'left',
                                        fontWeight: 600,
                                        fontSize: '14px',
                                        color: 'rgb(55, 65, 81)'
                                    }}>
                                        Object Path
                                    </th>
                                    <th style={{ 
                                        padding: '12px 16px', 
                                        textAlign: 'center',
                                        fontWeight: 600,
                                        fontSize: '14px',
                                        color: 'rgb(55, 65, 81)',
                                        width: '120px'
                                    }}>
                                        Upload Count
                                    </th>
                                    <th style={{ 
                                        padding: '12px 16px', 
                                        textAlign: 'center',
                                        fontWeight: 600,
                                        fontSize: '14px',
                                        color: 'rgb(55, 65, 81)',
                                        width: '80px'
                                    }}>
                                        Actions
                                    </th>
                                </tr>
                            </thead>
                            <tbody>
                                {Object.keys(objectUploadTracker)
                                    .filter(key => {
                                        const objData = objectUploadTracker[key];
                                        // Only show objects that were actually uploaded multiple times (override objects)
                                        // This means totalVersions > 1 OR it's an expected override object (override-obj-)
                                        return key.includes('override-obj-') && objData.totalVersions > 0;
                                    })
                                    .sort()
                                    .map((key, idx) => {
                                        const objData = objectUploadTracker[key];
                                        const mcCommand = `mc ls --versions ${objData.alias}/${objData.bucket}/${key}`;
                                        const isCopied = copiedCommand === key;
                                        
                                        return (
                                            <tr 
                                                key={idx}
                                                style={{ 
                                                    borderBottom: '1px solid rgb(229, 231, 235)',
                                                    backgroundColor: 'white'
                                                }}
                                            >
                                                <td style={{ 
                                                    padding: '12px 16px',
                                                    fontSize: '13px',
                                                    fontWeight: 500,
                                                    color: 'rgb(55, 65, 81)'
                                                }}>
                                                    <span style={{
                                                        padding: '2px 8px',
                                                        backgroundColor: '#dbeafe',
                                                        color: '#1e40af',
                                                        borderRadius: '4px',
                                                        fontSize: '12px',
                                                        fontFamily: 'monospace'
                                                    }}>
                                                        {objData.alias}
                                                    </span>
                                                </td>
                                                <td style={{ 
                                                    padding: '12px 16px',
                                                    fontSize: '13px',
                                                    fontWeight: 500,
                                                    color: 'rgb(55, 65, 81)'
                                                }}>
                                                    <span style={{
                                                        padding: '2px 8px',
                                                        backgroundColor: '#f3f4f6',
                                                        color: '#374151',
                                                        borderRadius: '4px',
                                                        fontSize: '12px',
                                                        fontFamily: 'monospace'
                                                    }}>
                                                        {objData.bucket}
                                                    </span>
                                                </td>
                                                <td style={{ 
                                                    padding: '12px 16px',
                                                    fontSize: '13px',
                                                    fontFamily: 'monospace',
                                                    color: 'rgb(55, 65, 81)',
                                                    maxWidth: '300px',
                                                    overflow: 'hidden',
                                                    textOverflow: 'ellipsis',
                                                    whiteSpace: 'nowrap'
                                                }}
                                                title={key}
                                                >
                                                    {key}
                                                </td>
                                                <td style={{ 
                                                    padding: '12px 16px',
                                                    textAlign: 'center',
                                                    fontSize: '14px',
                                                    fontWeight: 600,
                                                    color: '#1e40af'
                                                }}>
                                                    {objData.totalVersions}
                                                </td>
                                                <td style={{ 
                                                    padding: '12px 16px',
                                                    textAlign: 'center'
                                                }}>
                                                    <button
                                                        onClick={() => copyToClipboard(mcCommand, key)}
                                                        title="Copy mc ls --versions command"
                                                        style={{
                                                            padding: '8px 10px',
                                                            backgroundColor: isCopied ? '#E8F5E8' : 'white',
                                                            color: isCopied ? '#4CAF50' : '#2563eb',
                                                            border: '1.5px solid',
                                                            borderColor: isCopied ? '#4CAF50' : '#3b82f6',
                                                            borderRadius: '6px',
                                                            cursor: 'pointer',
                                                            display: 'inline-flex',
                                                            alignItems: 'center',
                                                            justifyContent: 'center',
                                                            transition: 'all 0.2s ease',
                                                            fontSize: '13px',
                                                            fontWeight: 500,
                                                            boxShadow: '0 1px 2px rgba(0, 0, 0, 0.05)'
                                                        }}
                                                        onMouseOver={(e) => {
                                                            if (!isCopied) {
                                                                e.currentTarget.style.backgroundColor = '#eff6ff';
                                                                e.currentTarget.style.borderColor = '#2563eb';
                                                            }
                                                        }}
                                                        onMouseOut={(e) => {
                                                            if (!isCopied) {
                                                                e.currentTarget.style.backgroundColor = 'white';
                                                                e.currentTarget.style.borderColor = '#3b82f6';
                                                            }
                                                        }}
                                                    >
                                                        {isCopied ? (
                                                            <Check size={16} />
                                                        ) : (
                                                            <Copy size={16} />
                                                        )}
                                                    </button>
                                                </td>
                                            </tr>
                                        );
                                    })}
                            </tbody>
                        </table>
                        </div>
                    </div>
                    
                    <div style={{ marginTop: '12px', padding: '10px', backgroundColor: '#f0f9ff', border: '1px solid #bae6fd', borderRadius: '6px' }}>
                        <div style={{ fontSize: '13px', color: '#0c4a6e', display: 'flex', alignItems: 'center', gap: '6px' }}>
                            <Copy size={14} />
                            <span><strong>Tip:</strong> Click the copy button to get the <code style={{ 
                                backgroundColor: '#e0f2fe', 
                                padding: '2px 6px', 
                                borderRadius: '3px',
                                fontFamily: 'monospace',
                                fontSize: '12px'
                            }}>mc ls --versions</code> command for each object.</span>
                        </div>
                    </div>
                </div>
            )}

        </div>
    );
};

export default TestProgress;
