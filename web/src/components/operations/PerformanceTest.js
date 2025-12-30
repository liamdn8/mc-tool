import React, { useState, useEffect } from 'react';
import { CheckSquare, Play, Loader } from 'lucide-react';

import { apiCall } from '../../utils/api';
import { useContentsPanel } from '../../contexts/ContentsPanelContext';
import TestProgress from '../testing/TestProgress';
import TestSummary from '../testing/TestSummary';
import TestNavigation from '../testing/TestNavigation';

const PerformanceTest = () => {
  const { setContentsComponent } = useContentsPanel();
  const [aliases, setAliases] = useState([]);
  const [aliasesData, setAliasesData] = useState([]); // Store full alias objects
  const [selectedAlias, setSelectedAlias] = useState('');
  const [buckets, setBuckets] = useState([]);
  const [selectedBucket, setSelectedBucket] = useState('');
  const [isLoadingBuckets, setIsLoadingBuckets] = useState(false);
  const [objectSizeType, setObjectSizeType] = useState('small');
  const [objectCount, setObjectCount] = useState(10);
  const [overrideCount, setOverrideCount] = useState(0);
  const [uploadMode, setUploadMode] = useState('all');
  const [uploadInterval, setUploadInterval] = useState('5s');
  const [iterations, setIterations] = useState(5);
  const [parallelism, setParallelism] = useState(1);
  const [insecure, setInsecure] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [testStatus, setTestStatus] = useState(null);
  const [testResult, setTestResult] = useState(null);
  const [error, setError] = useState(null);

  // Update insecure when alias changes
  useEffect(() => {
    if (selectedAlias && aliasesData.length > 0) {
      const aliasObj = aliasesData.find(a => a.name === selectedAlias);
      console.log('Alias selected:', selectedAlias, 'Alias obj:', aliasObj);
      if (aliasObj && aliasObj.insecure !== undefined) {
        console.log('Setting insecure to:', aliasObj.insecure);
        setInsecure(aliasObj.insecure);
      }
    }
  }, [selectedAlias, aliasesData]);

  // Load aliases on mount
  useEffect(() => {
    loadAliases();
  }, []);

  // Load buckets when alias changes
  useEffect(() => {
    if (selectedAlias) {
      loadBuckets(selectedAlias);
    } else {
      setBuckets([]);
      setSelectedBucket('');
    }
  }, [selectedAlias]);

  // Cleanup contents panel on unmount
  useEffect(() => {
    return () => {
      setContentsComponent(null);
    };
  }, [setContentsComponent]);

  // Poll status when test is running
  useEffect(() => {
    let interval;
    if (isRunning) {
      interval = setInterval(async () => {
        try {
          const { response, data } = await apiCall('/api/perftest/status');
          if (response.ok) {
            console.log('Status poll response:', data);
            
            // Backend returns { running: bool, status: {...} }
            // Convert snake_case to camelCase
            const status = data.status || {};
            
            // Convert round_details array
            const roundDetails = (status.round_details || []).map(rd => ({
              round: rd.round,
              objectsUploaded: rd.objects_uploaded || 0,
              successfulUploads: rd.successful_uploads || 0,
              failedUploads: rd.failed_uploads || 0,
              duration: rd.duration || 0,
              startTime: rd.start_time,
              endTime: rd.end_time
            }));
            
            // Convert recent_uploads array
            const recentUploads = (status.recent_uploads || []).map(ru => ({
              objectKey: ru.object_key,
              objectSize: ru.object_size,
              uploadNumber: ru.upload_number,
              roundNumber: ru.round_number,
              startTime: ru.start_time,
              endTime: ru.end_time,
              duration: ru.duration,
              success: ru.success,
              error: ru.error
            }));
            
            const mergedStatus = {
              running: data.running,
              progress: status.progress || 0,
              totalUploads: status.total_uploads || 0,
              completedUploads: status.completed_uploads || 0,
              successfulUploads: status.successful_uploads || 0,
              failedUploads: status.failed_uploads || 0,
              elapsedTime: status.elapsed_time || 0,
              currentPhase: status.current_phase || '',
              currentRound: status.current_round || 0,
              totalRounds: status.total_rounds || 0,
              roundDetails: roundDetails,
              recentUploads: recentUploads,
              _timestamp: Date.now()
            };
            
            console.log('Merged status:', mergedStatus);
            setTestStatus(mergedStatus);
            
            if (!data.running) {
              setIsRunning(false);
              loadResult();
            }
          }
        } catch (err) {
          console.error('Failed to check status:', err);
        }
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [isRunning]);

  // Update contents panel when test results change
  useEffect(() => {
    if (isRunning || testResult) {
      setContentsComponent(
        <TestNavigation 
          isRunning={isRunning}
          testStatus={testStatus}
          testResult={testResult}
          overrideCount={overrideCount}
          embedded={true}
        />
      );
    } else {
      setContentsComponent(null);
    }
  }, [isRunning, testResult, testStatus, overrideCount, setContentsComponent]);

  const loadAliases = async () => {
    try {
      const { response, data } = await apiCall('/api/aliases');
      if (response.ok) {
        const aliasList = data.aliases || [];
        setAliasesData(aliasList); // Store full alias objects
        // Extract alias names from objects (format: [{name: "alias1", ...}, ...])
        const aliasNames = aliasList.map(alias => typeof alias === 'string' ? alias : alias.name);
        setAliases(aliasNames);
        // Don't auto-select - force user to choose
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
        const bucketList = data.buckets || [];
        setBuckets(bucketList);
        if (bucketList.length > 0) {
          setSelectedBucket(bucketList[0]);
        } else {
          setSelectedBucket('');
        }
      } else {
        setBuckets([]);
        setSelectedBucket('');
      }
    } catch (error) {
      console.error('Failed to load buckets:', error);
      setBuckets([]);
      setSelectedBucket('');
    } finally {
      setIsLoadingBuckets(false);
    }
  };

  const loadResult = async () => {
    try {
      const { response, data } = await apiCall('/api/perftest/result');
      if (response.ok && data.success && data.result) {
        setTestResult(data.result);
      }
    } catch (err) {
      console.error('Failed to load result:', err);
    }
  };

  const executeTest = async () => {
    if (!selectedAlias || !selectedBucket) {
      setError('Please select a site and bucket');
      return;
    }

    setIsRunning(true);
    setTestResult(null);
    setError(null);
    
    // Initialize testStatus with all rounds pending
    const totalRounds = uploadMode === 'interval' ? iterations : 1;
    setTestStatus({
      running: true,
      progress: 0,
      totalUploads: objectCount * (overrideCount + 1) * totalRounds,
      completedUploads: 0,
      successfulUploads: 0,
      failedUploads: 0,
      elapsedTime: 0,
      currentPhase: 'starting',
      currentRound: 0,
      totalRounds: totalRounds,
      roundDetails: [],
      recentUploads: []
    });

    try {
      const { response, data } = await apiCall('/api/perftest/start', {
        method: 'POST',
        body: JSON.stringify({
          site_alias: selectedAlias,
          bucket: selectedBucket,
          object_path: '', // Auto-generated
          object_size_type: objectSizeType,
          object_count: objectCount,
          override_count: overrideCount,
          upload_mode: uploadMode,
          upload_interval: uploadInterval,
          iterations: iterations,
          parallelism: parallelism,
          insecure: insecure
        })
      });

      if (!response.ok) {
        setIsRunning(false);
        setError(data.error || 'Failed to start test');
      }
    } catch (err) {
      setIsRunning(false);
      setError(err.message);
    }
  };

  const formatDuration = (ns) => {
    if (!ns) return '0ms';
    const ms = ns / 1000000;
    if (ms < 1000) return `${ms.toFixed(0)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const formatBytes = (bytes) => {
    if (!bytes) return '0 B';
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
  };

  return (
    <div style={{ padding: '20px' }}>
      {/* Configuration Section */}
      <div className="card" style={{ marginBottom: '20px' }}>
        <div className="card-header">
          <h3 className="card-title">Test Configuration</h3>
        </div>
        <div style={{ padding: '20px' }}>
          {/* Target Configuration */}
          <div style={{ marginBottom: '24px' }}>
            <h4 style={{ fontSize: '14px', fontWeight: 600, marginBottom: '12px', color: 'var(--text-secondary)' }}>
              Target Configuration
            </h4>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '16px' }}>
              {/* Site Selection */}
              <div className="form-group">
                <label style={{ display: 'block', marginBottom: '6px', fontWeight: 500 }}>Site</label>
                <select
                  value={selectedAlias}
                  onChange={(e) => setSelectedAlias(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    border: '1px solid var(--border-color)',
                    borderRadius: '6px',
                    fontSize: '14px'
                  }}
                >
                  <option value="">-- Select Alias --</option>
                  {[...aliases].sort().map(alias => (
                    <option key={alias} value={alias}>{alias}</option>
                  ))}
                </select>
              </div>

              {/* Bucket Selection */}
              <div className="form-group">
                <label style={{ display: 'block', marginBottom: '6px', fontWeight: 500 }}>Bucket</label>
                <select
                  value={selectedBucket}
                  onChange={(e) => setSelectedBucket(e.target.value)}
                  disabled={isLoadingBuckets || buckets.length === 0}
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    border: '1px solid var(--border-color)',
                    borderRadius: '6px',
                    fontSize: '14px',
                    backgroundColor: isLoadingBuckets ? 'var(--bg-secondary)' : 'white'
                  }}
                >
                  {isLoadingBuckets ? (
                    <option>Loading buckets...</option>
                  ) : buckets.length === 0 ? (
                    <option>No buckets available</option>
                  ) : (
                    buckets.map(bucket => (
                      <option key={bucket} value={bucket}>{bucket}</option>
                    ))
                  )}
                </select>
              </div>
            </div>

            {/* Insecure - Second Row */}
            <div style={{ marginTop: '16px' }}>
              <label style={{
                display: 'flex',
                alignItems: 'center',
                padding: '12px',
                borderRadius: '6px',
                cursor: 'pointer',
                border: insecure ? `1px solid #FF9800` : '1px solid #d1d5db',
                backgroundColor: insecure ? '#fdfbf8ff' : '#ffffff',
                fontSize: '13px',
                gap: '8px',
                marginBottom: '8px'
              }}>
                <input
                  type="checkbox"
                  checked={insecure}
                  onChange={(e) => setInsecure(e.target.checked)}
                  style={{ cursor: 'pointer' }}
                />
                {/* <span style={{ fontWeight: 600, color: 'rgb(245, 158, 11)' }}>Skip TLS certificate verification (--insecure)</span> */}
                <span>{'Skip TLS certificate verification (--insecure)'}</span>
              </label>
            </div>
          </div>

          {/* Upload Mode Configuration */}
          <div style={{ marginBottom: '24px' }}>
            <h4 style={{ fontSize: '14px', fontWeight: 600, marginBottom: '12px', color: 'var(--text-secondary)' }}>
              Upload Mode
            </h4>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '16px' }}>
              {/* Upload Mode */}
              <div className="form-group">
                <label style={{ display: 'block', marginBottom: '6px', fontWeight: 500 }}>Mode</label>
                <select
                  value={uploadMode}
                  onChange={(e) => setUploadMode(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    border: '1px solid var(--border-color)',
                    borderRadius: '6px',
                    fontSize: '14px'
                  }}
                >
                  <option value="all">All at Once (Parallel)</option>
                  <option value="interval">Timed Rounds (Interval)</option>
                </select>
              </div>

              {/* Mode-specific options */}
              {uploadMode === 'all' ? (
                <>
                  <div className="form-group">
                    <label style={{ display: 'block', marginBottom: '6px', fontWeight: 500 }}>Parallelism</label>
                    <input
                      type="number"
                      value={parallelism}
                      onChange={(e) => setParallelism(parseInt(e.target.value, 10))}
                      min="1"
                      max="50"
                      style={{
                        width: '100%',
                        padding: '8px 12px',
                        border: '1px solid var(--border-color)',
                        borderRadius: '6px',
                        fontSize: '14px'
                      }}
                    />
                  </div>
                </>
              ) : (
                <>
                  <div className="form-group">
                    <label style={{ display: 'block', marginBottom: '6px', fontWeight: 500 }}>Number of Rounds</label>
                    <input
                      type="number"
                      value={iterations}
                      onChange={(e) => setIterations(parseInt(e.target.value, 10))}
                      min="1"
                      max="100"
                      style={{
                        width: '100%',
                        padding: '8px 12px',
                        border: '1px solid var(--border-color)',
                        borderRadius: '6px',
                        fontSize: '14px'
                      }}
                    />
                  </div>
                  <div className="form-group">
                    <label style={{ display: 'block', marginBottom: '6px', fontWeight: 500 }}>Interval Between Rounds</label>
                    <input
                      type="text"
                      value={uploadInterval}
                      onChange={(e) => setUploadInterval(e.target.value)}
                      placeholder="5s, 10s, 1m"
                      style={{
                        width: '100%',
                        padding: '8px 12px',
                        border: '1px solid var(--border-color)',
                        borderRadius: '6px',
                        fontSize: '14px'
                      }}
                    />
                  </div>
                </>
              )}
            </div>
          </div>

          {/* Object Configuration */}
          <div style={{ marginBottom: '24px' }}>
            <h4 style={{ fontSize: '14px', fontWeight: 600, marginBottom: '12px', color: 'var(--text-secondary)' }}>
              Object Configuration
            </h4>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '16px' }}>
              {/* Object Size */}
              <div className="form-group">
                <label style={{ display: 'block', marginBottom: '6px', fontWeight: 500 }}>Object Size</label>
                <select
                  value={objectSizeType}
                  onChange={(e) => setObjectSizeType(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    border: '1px solid var(--border-color)',
                    borderRadius: '6px',
                    fontSize: '14px'
                  }}
                >
                  <option value="small">Small (1-10 KB)</option>
                  <option value="medium">Medium (100-500 KB)</option>
                  <option value="large">Large (1-5 MB)</option>
                </select>
              </div>

              {/* Object Count */}
              <div className="form-group">
                <label style={{ display: 'block', marginBottom: '6px', fontWeight: 500 }}>
                  {uploadMode === 'interval' ? 'Unique Files per Round' : 'Number of Files'}
                  <span style={{ fontWeight: 400, fontSize: '12px', color: 'var(--text-muted)', marginLeft: '6px' }}>
                    {uploadMode === 'interval' ? '(once per round)' : '(total files)'}
                  </span>
                </label>
                <input
                  type="number"
                  value={objectCount}
                  onChange={(e) => setObjectCount(parseInt(e.target.value, 10))}
                  min="1"
                  max={uploadMode === 'interval' ? '100' : '1000'}
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    border: '1px solid var(--border-color)',
                    borderRadius: '6px',
                    fontSize: '14px'
                  }}
                />
              </div>

              {/* Override Count - Only show for interval mode */}
              {uploadMode === 'interval' && (
              <div className="form-group">
                <label style={{ display: 'block', marginBottom: '6px', fontWeight: 500 }}>
                  Repeated Files per Round
                  <span style={{ fontWeight: 400, fontSize: '12px', color: 'var(--text-muted)', marginLeft: '6px' }}>
                    (optional)
                  </span>
                </label>
                <input
                  type="number"
                  value={overrideCount}
                  onChange={(e) => setOverrideCount(parseInt(e.target.value, 10))}
                  min="0"
                  max="10"
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    border: '1px solid var(--border-color)',
                    borderRadius: '6px',
                    fontSize: '14px'
                  }}
                />
                <small style={{ color: 'var(--text-muted)', fontSize: '12px' }}>
                  Same files uploaded in every round with new versions
                </small>
              </div>
              )}
            </div>
          </div>

          {/* Test Execution Description */}
          {(objectCount > 0 || overrideCount > 0) && (
            <div style={{ 
              marginTop: '20px', 
              padding: '16px', 
              backgroundColor: '#eff6ff', 
              border: '1px solid #bfdbfe', 
              borderRadius: '6px'
            }}>
              <div style={{ fontWeight: 600, fontSize: '14px', color: '#1e40af', marginBottom: '8px', display: 'inline-flex', alignItems: 'center', gap: '6px' }}>
                <CheckSquare size={16} style={{ color: '#2563eb' }} />
                <span>Test Execution Plan:</span>
              </div>
              <div style={{ fontSize: '13px', color: '#1e3a8a', lineHeight: '1.6' }}>
                {/* Target Info */}
                <div style={{ marginBottom: '8px', paddingBottom: '8px', borderBottom: '1px solid #bfdbfe' }}>
                  * Target: <strong>{selectedAlias}/{selectedBucket}</strong>
                  <br />* Folder: <strong>mc-test/[timestamp]/</strong>
                </div>
                
                {uploadMode === 'all' ? (
                  <>
                    * Upload <strong>{objectCount} files</strong> in parallel
                    {overrideCount > 0 && (
                      <>
                        <br />* Each file will be uploaded <strong>{overrideCount + 1} times</strong> (1 original + {overrideCount} override{overrideCount > 1 ? 's' : ''})
                      </>
                    )}
                    <br />* <strong>Total uploads: {objectCount * (overrideCount + 1)}</strong>
                  </>
                ) : (
                  <>
                    * Run <strong>{iterations} round{iterations > 1 ? 's' : ''}</strong> with <strong>{uploadInterval}</strong> interval
                    <br />* Each round uploads:
                    <br />  - <strong>{objectCount} unique files</strong> (different in each round)
                    {overrideCount > 0 && (
                      <>
                        <br />  - <strong>{overrideCount} repeated files</strong> (same files, new version each round)
                      </>
                    )}
                    <br />* <strong>Total per round: {objectCount + overrideCount} files</strong>
                    <br />* <strong>Total uploads: {(objectCount + overrideCount) * iterations}</strong>
                  </>
                )}
              </div>
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div style={{
              marginTop: '16px',
              padding: '12px',
              backgroundColor: 'var(--danger-light)',
              border: '1px solid var(--danger-color)',
              borderRadius: '6px',
              color: 'var(--danger-text)',
              fontSize: '14px'
            }}>
              {error}
            </div>
          )}

          {/* Action Button */}
          <div style={{ marginTop: '20px' }}>
            <button
              onClick={executeTest}
              disabled={isRunning}
              style={{
                padding: '10px 24px',
                backgroundColor: isRunning ? 'var(--text-muted)' : 'var(--primary-color)',
                color: 'white',
                border: 'none',
                borderRadius: '6px',
                fontSize: '14px',
                fontWeight: 500,
                cursor: isRunning ? 'not-allowed' : 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '8px'
              }}
            >
              {isRunning ? (
                <>
                  <Loader size={16} className="spin" />
                  Test Running...
                </>
              ) : (
                <>
                  <Play size={16} />
                  Start Test
                </>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Progress Section */}
      {(isRunning || testResult) && (
        <div className="card" style={{ marginTop: '24px' }}>
          <div className="card-header">
            <h3 className="card-title">{isRunning ? 'Test Progress' : 'Test Results'}</h3>
          </div>
          <div style={{ padding: '20px' }}>
            {testStatus && <TestProgress status={testStatus} isComplete={!isRunning} result={testResult} />}
          </div>
        </div>
      )}
    </div>
  );
};

export default PerformanceTest;
