import React, { useState, useEffect } from 'react';
import { Server, CheckCircle, XCircle, AlertCircle, Loader, Play } from 'lucide-react';
import { apiCall } from '../utils/api';
import { useI18n } from '../utils/i18n';
import { useContentsPanel } from '../contexts/ContentsPanelContext';
import YamlDiffViewer from '../components/YamlDiffViewer';
import InfraOverviewCards from '../components/infrastructure/InfraOverviewCards';
import ResourceComparisonTable from '../components/infrastructure/ResourceComparisonTable';
import InfraValidationNavigation from '../components/infrastructure/InfraValidationNavigation';

const InfrastructureValidatePage = () => {
    const { t } = useI18n();
    const { setContentsComponent } = useContentsPanel();
    const [vims, setVims] = useState([]);
    const [loading, setLoading] = useState(false);
    const [baseline, setBaseline] = useState('');
    const [baselineNamespace, setBaselineNamespace] = useState('');
    const [baselineNamespaces, setBaselineNamespaces] = useState([]);
    const [loadingBaselineNs, setLoadingBaselineNs] = useState(false);
    const [targets, setTargets] = useState([{ vim: '', namespace: '', namespaces: [] }]);
    const [validating, setValidating] = useState(false);
    const [jobId, setJobId] = useState(null);
    const [result, setResult] = useState(null);
    const [error, setError] = useState(null);
    const [diffViewer, setDiffViewer] = useState(null);

    useEffect(() => {
        console.log('InfrastructureValidatePage mounted, loading VIMs...');
        loadVIMs();
        
        // Cleanup contents panel when component unmounts
        return () => {
            setContentsComponent(null);
        };
    }, [setContentsComponent]);

    useEffect(() => {
        if (baseline) {
            console.log('Baseline changed:', baseline, 'loading namespaces...');
            loadNamespaces(baseline, 'baseline');
        } else {
            setBaselineNamespaces([]);
            setBaselineNamespace('');
        }
    }, [baseline]);

    // Update contents panel when validation results change
    useEffect(() => {
        if (result) {
            setContentsComponent(
                <InfraValidationNavigation 
                    result={result}
                    embedded={true}
                />
            );
        } else {
            setContentsComponent(null);
        }
    }, [result, setContentsComponent]);

    const loadVIMs = async () => {
        console.log('loadVIMs called');
        setLoading(true);
        try {
            console.log('Calling API: /api/validate/infrastructure/vims');
            const { data } = await apiCall('/api/validate/infrastructure/vims');
            console.log('VIMs loaded:', data.vims);
            setVims(data.vims || []);
        } catch (err) {
            console.error('Failed to load VIMs:', err);
            setError('Failed to load VIMs: ' + err.message);
        } finally {
            setLoading(false);
        }
    };

    const loadNamespaces = async (vim, target) => {
        if (target === 'baseline') {
            setLoadingBaselineNs(true);
        }
        
        try {
            const { data } = await apiCall(`/api/validate/infrastructure/namespaces?vim=${vim}`);
            
            if (target === 'baseline') {
                setBaselineNamespaces(data.namespaces || []);
            } else {
                // Update target namespaces
                const targetIndex = parseInt(target);
                const newTargets = [...targets];
                newTargets[targetIndex].namespaces = data.namespaces || [];
                setTargets(newTargets);
            }
        } catch (err) {
            console.error('Failed to load namespaces:', err);
            if (target === 'baseline') {
                setBaselineNamespaces([]);
            }
        } finally {
            if (target === 'baseline') {
                setLoadingBaselineNs(false);
            }
        }
    };

    const addTarget = () => {
        setTargets([...targets, { vim: '', namespace: '', namespaces: [] }]);
    };

    const removeTarget = (index) => {
        setTargets(targets.filter((_, i) => i !== index));
    };

    const updateTarget = (index, field, value) => {
        const newTargets = [...targets];
        newTargets[index][field] = value;
        
        // If VIM changed, load namespaces for that VIM
        if (field === 'vim' && value) {
            newTargets[index].namespace = '';
            newTargets[index].namespaces = [];
            setTargets(newTargets);
            loadNamespaces(value, index.toString());
        } else {
            setTargets(newTargets);
        }
    };

    const handleValidate = async () => {
        // Validation
        if (!baseline || !baselineNamespace) {
            setError('Please select baseline VIM and namespace');
            return;
        }

        const validTargets = targets.filter(t => t.vim && t.namespace);
        if (validTargets.length === 0) {
            setError('Please add at least one target');
            return;
        }

        setError(null);
        setValidating(true);
        setResult(null);

        try {
            const baselineStr = `${baseline}/${baselineNamespace}`;
            const targetStrs = validTargets.map(t => `${t.vim}/${t.namespace}`);

            const { data } = await apiCall('/api/validate/infrastructure', {
                method: 'POST',
                body: JSON.stringify({
                    baseline: baselineStr,
                    targets: targetStrs
                })
            });

            setJobId(data.job_id);
            pollJobStatus(data.job_id);
        } catch (err) {
            setError(err.message || 'Failed to start validation');
            setValidating(false);
        }
    };

    const pollJobStatus = async (jid) => {
        const maxAttempts = 60;
        let attempts = 0;

        const poll = async () => {
            try {
                const { data } = await apiCall(`/api/jobs/${jid}`);
                
                if (data.status === 'completed') {
                    setResult(data.result);
                    setValidating(false);
                } else if (data.status === 'failed') {
                    setError(data.error || 'Validation failed');
                    setValidating(false);
                } else if (attempts < maxAttempts) {
                    attempts++;
                    setTimeout(poll, 1000);
                } else {
                    setError('Validation timeout');
                    setValidating(false);
                }
            } catch (err) {
                setError(err.message || 'Failed to get job status');
                setValidating(false);
            }
        };

        poll();
    };

    const renderResultSummary = () => {
        if (!result || !result.summary) return null;

        const { summary } = result;
        
        // Build resource table data structure
        const resourceTableData = buildResourceTableData(summary.resource_table || [], summary.baseline, summary.targets || []);

        return (
            <div className="card" style={{ marginTop: '20px' }}>
                <div className="card-header">
                    <h3 className="card-title">Validation Results</h3>
                    <div style={{ marginTop: '8px' }}>
                        <strong>Baseline:</strong> {summary.baseline}
                    </div>
                </div>
                <div style={{ padding: '20px' }}>
                    {/* Overview Section */}
                    <div id="overview">
                        <InfraOverviewCards result={result} />

                        {/* Status Alert */}
                        {summary.status === 'success' && (
                            <div className="alert alert-success" style={{ marginBottom: '20px' }}>
                                <CheckCircle size={20} />
                                <span>All configurations match!</span>
                            </div>
                        )}

                        {summary.status === 'drift' && (
                            <div className="alert alert-warning" style={{ marginBottom: '20px' }}>
                                <AlertCircle size={20} />
                                <span>Configuration drift detected!</span>
                            </div>
                        )}
                    </div>

                        {/* Resource Comparison Tables by Type */}
                        {Object.keys(resourceTableData).length > 0 && (
                            <div style={{ marginTop: '20px' }}>
                                <h4 style={{ marginBottom: '16px', fontSize: '16px', fontWeight: 'bold' }}>
                                    Resource Comparison by Type
                                </h4>
                                {Object.entries(resourceTableData).sort().map(([resourceType, resources]) => (
                                    <div key={resourceType} id={`resource-${resourceType}`}>
                                        <ResourceComparisonTable
                                            resourceType={resourceType}
                                            resources={resources}
                                            baseline={summary.baseline}
                                            targets={summary.targets || []}
                                            onViewDiff={handleViewDiff}
                                        />
                                    </div>
                                ))}
                            </div>
                        )}

                        {/* Raw Output */}
                        {result.output && (
                            <details style={{ marginTop: '20px' }}>
                                <summary style={{ cursor: 'pointer', fontWeight: 'bold', marginBottom: '8px' }}>
                                    View Raw Output
                                </summary>
                                <pre style={{
                                    backgroundColor: 'var(--card-bg)',
                                    padding: '12px',
                                    borderRadius: '4px',
                                    overflow: 'auto',
                                    fontSize: '12px',
                                    maxHeight: '400px',
                                    border: '1px solid var(--border)'
                                }}>
                                    {result.output}
                                </pre>
                            </details>
                        )}
                </div>
            </div>
        );
    };

    const handleViewDiff = async (resourceType, resourceName, baselineNs, targetNs) => {
        try {
            const { data } = await apiCall(
                `/api/validate/infrastructure/diff?baseline=${baselineNs}&target=${targetNs}&resource_type=${resourceType}&resource_name=${resourceName}`
            );

            setDiffViewer({
                baseline: data.baseline,
                target: data.target,
                resourceName: `${resourceType}/${resourceName}`,
                baselineLabel: baselineNs,
                targetLabel: targetNs
            });
        } catch (err) {
            console.error('Failed to load diff:', err);
            setError('Failed to load diff: ' + err.message);
        }
    };

    const buildResourceTableData = (resourceTable, baseline, targets) => {
        if (!resourceTable || resourceTable.length === 0) return {};

        console.log('buildResourceTableData called with:', { resourceTable, baseline, targets });

        const allNamespaces = [baseline, ...(targets || [])];
        
        // Group by resource type
        const groupedByType = resourceTable.reduce((acc, row) => {
            const type = row.resource_type || 'Unknown';
            const name = row.resource_name;
            
            if (!acc[type]) {
                acc[type] = [];
            }
            
            // Build row with namespace columns
            const rowData = {
                resource_name: name,
                resource_type: type
            };
            
            // Add status for each namespace - backend already provides this in row[ns]
            allNamespaces.forEach(ns => {
                if (row[ns]) {
                    // Backend already has status in row[ns]
                    rowData[ns] = row[ns];
                } else if (ns === baseline) {
                    // Baseline should always exist
                    rowData[ns] = { status: 'configured' };
                } else {
                    // Default to not_found for missing namespaces
                    rowData[ns] = { status: 'not_found' };
                }
            });
            
            console.log('Built row:', rowData);
            acc[type].push(rowData);
            return acc;
        }, {});

        console.log('Grouped by type:', groupedByType);
        return groupedByType;
    };

    return (
        <div>
            <div className="card-header">
                <h2 className="card-title">
                    <Server size={24} style={{ marginRight: '8px', verticalAlign: 'middle' }} />
                    Infrastructure Validation
                </h2>
                <p style={{ margin: '8px 0 0 0', color: 'var(--text-secondary)' }}>
                    Validate Kubernetes namespace configurations across multiple clusters
                </p>
            </div>

            {/* Guidelines */}
            <div className="card" style={{ marginBottom: '20px' }}>
                <div className="card-header">
                    <h3 className="card-title">How It Works</h3>
                </div>
                <div style={{ padding: '20px' }}>
                    <ul style={{ margin: 0, paddingLeft: '20px' }}>
                        <li>Select a baseline VIM (Virtual Infrastructure) and namespace to compare against</li>
                        <li>Add target VIMs and namespaces for comparison</li>
                        <li>The tool compares Deployments, StatefulSets, ConfigMaps, Secrets, and Services</li>
                        <li>Configuration differences are highlighted in the results</li>
                        <li>Useful for validating infrastructure consistency across environments (prod, staging, dev)</li>
                    </ul>
                </div>
            </div>

            {/* Configuration Form */}
            <div className="card">
                <div className="card-header">
                    <h3 className="card-title">Validation Configuration</h3>
                </div>
                <div style={{ padding: '20px' }}>
                    {/* Baseline Selection */}
                    <div style={{ marginBottom: '20px' }}>
                        <h4 style={{ marginBottom: '12px', fontSize: '14px', fontWeight: 'bold' }}>
                            Baseline (Reference Configuration)
                        </h4>
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                            <div>
                                <label className="form-label">VIM (Virtual Infrastructure)</label>
                                <select
                                    className="form-input"
                                    value={baseline}
                                    onChange={(e) => setBaseline(e.target.value)}
                                    disabled={loading}
                                >
                                    <option value="">Select VIM...</option>
                                    {vims.map(vim => (
                                        <option key={vim.name} value={vim.name}>
                                            {vim.name} - {vim.endpoint}
                                        </option>
                                    ))}
                                </select>
                            </div>
                            <div>
                                <label className="form-label">Namespace</label>
                                <select
                                    className="form-input"
                                    value={baselineNamespace}
                                    onChange={(e) => setBaselineNamespace(e.target.value)}
                                    disabled={!baseline || loadingBaselineNs}
                                >
                                    <option value="">
                                        {!baseline ? 'Select VIM first...' : loadingBaselineNs ? 'Loading...' : 'Select namespace...'}
                                    </option>
                                    {baselineNamespaces.map(ns => (
                                        <option key={ns} value={ns}>{ns}</option>
                                    ))}
                                </select>
                            </div>
                        </div>
                    </div>

                    {/* Targets */}
                    <div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                            <h4 style={{ fontSize: '14px', fontWeight: 'bold', margin: 0 }}>
                                Targets (Environments to Compare)
                            </h4>
                            <button className="btn btn-secondary" onClick={addTarget}>
                                + Add Target
                            </button>
                        </div>

                        {targets.map((target, index) => (
                            <div key={index} style={{
                                display: 'grid',
                                gridTemplateColumns: '1fr 1fr auto',
                                gap: '12px',
                                marginBottom: '12px',
                                padding: '12px',
                                backgroundColor: 'var(--card-bg)',
                                borderRadius: '8px',
                                border: '1px solid var(--border)'
                            }}>
                                <div>
                                    <label className="form-label">VIM (Virtual Infrastructure)</label>
                                    <select
                                        className="form-input"
                                        value={target.vim}
                                        onChange={(e) => updateTarget(index, 'vim', e.target.value)}
                                        disabled={loading}
                                    >
                                        <option value="">Select VIM...</option>
                                        {vims.map(vim => (
                                            <option key={vim.name} value={vim.name}>
                                                {vim.name} - {vim.endpoint}
                                            </option>
                                        ))}
                                    </select>
                                </div>
                                <div>
                                    <label className="form-label">Namespace</label>
                                    <select
                                        className="form-input"
                                        value={target.namespace}
                                        onChange={(e) => updateTarget(index, 'namespace', e.target.value)}
                                        disabled={!target.vim}
                                    >
                                        <option value="">
                                            {!target.vim ? 'Select VIM first...' : 'Select namespace...'}
                                        </option>
                                        {target.namespaces && target.namespaces.map(ns => (
                                            <option key={ns} value={ns}>{ns}</option>
                                        ))}
                                    </select>
                                </div>
                                <div style={{ display: 'flex', alignItems: 'flex-end' }}>
                                    <button
                                        className="btn btn-danger"
                                        onClick={() => removeTarget(index)}
                                        disabled={targets.length === 1}
                                    >
                                        Remove
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>

                    {error && (
                        <div className="alert alert-error" style={{ marginTop: '16px' }}>
                            <XCircle size={20} />
                            <span>{error}</span>
                        </div>
                    )}

                    <div style={{ marginTop: '20px' }}>
                        <button
                            className="btn btn-primary"
                            onClick={handleValidate}
                            disabled={validating || !baseline || !baselineNamespace}
                            style={{ minWidth: '150px' }}
                        >
                            {validating ? (
                                <>
                                    <Loader size={16} className="spinner" style={{ marginRight: '8px' }} />
                                    Validating...
                                </>
                            ) : (
                                <>
                                    <Play size={16} style={{ marginRight: '8px' }} />
                                    Start Validation
                                </>
                            )}
                        </button>
                    </div>
                </div>
            </div>

            {/* Results */}
            {renderResultSummary()}

            {/* Diff Viewer Modal */}
            {diffViewer && (
                <YamlDiffViewer
                    baseline={diffViewer.baseline}
                    target={diffViewer.target}
                    resourceName={diffViewer.resourceName}
                    baselineLabel={diffViewer.baselineLabel}
                    targetLabel={diffViewer.targetLabel}
                    onClose={() => setDiffViewer(null)}
                />
            )}
        </div>
    );
};

export default InfrastructureValidatePage;
