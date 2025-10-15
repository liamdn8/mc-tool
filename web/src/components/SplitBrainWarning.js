import React, { useState, useEffect } from 'react';
import { AlertTriangle, X, RefreshCw } from 'lucide-react';
import { checkSplitBrainStatus } from '../utils/api';

const SplitBrainWarning = ({ onRefresh }) => {
    const [splitBrainData, setSplitBrainData] = useState(null);
    const [isVisible, setIsVisible] = useState(false);
    const [isLoading, setIsLoading] = useState(true);

    const checkSplitBrain = async () => {
        try {
            setIsLoading(true);
            const data = await checkSplitBrainStatus();
            setSplitBrainData(data);
            setIsVisible(data.splitBrainDetected);
        } catch (error) {
            console.error('Error checking split brain status:', error);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        checkSplitBrain();
    }, []);

    if (isLoading || !isVisible || !splitBrainData) {
        return null;
    }

    const handleDismiss = () => {
        setIsVisible(false);
    };

    const handleRefresh = () => {
        checkSplitBrain();
        if (onRefresh) {
            onRefresh();
        }
    };

    return (
        <div style={{
            background: 'linear-gradient(135deg, #ff6b6b, #ee5a24)',
            color: 'white',
            padding: '16px',
            borderRadius: '8px',
            margin: '16px 0',
            border: '2px solid #c0392b',
            boxShadow: '0 4px 12px rgba(255, 107, 107, 0.3)'
        }}>
            <div style={{
                display: 'flex',
                alignItems: 'flex-start',
                justifyContent: 'space-between'
            }}>
                <div style={{ flex: 1 }}>
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        marginBottom: '12px'
                    }}>
                        <AlertTriangle size={24} style={{ marginRight: '8px' }} />
                        <h3 style={{ margin: 0, fontSize: '1.2rem', fontWeight: 'bold' }}>
                            ⚠️ SPLIT BRAIN DETECTED
                        </h3>
                    </div>

                    <div style={{ marginBottom: '16px' }}>
                        <p style={{ margin: '0 0 8px 0', fontSize: '1rem' }}>
                            <strong>{splitBrainData.clusterCount} separate replication clusters detected!</strong>
                        </p>
                        <p style={{ margin: '0 0 12px 0', fontSize: '0.9rem', opacity: 0.9 }}>
                            This configuration can cause data inconsistency and conflicts.
                        </p>
                    </div>

                    {/* Warnings */}
                    {splitBrainData.warnings && splitBrainData.warnings.length > 0 && (
                        <div style={{ marginBottom: '16px' }}>
                            <h4 style={{ margin: '0 0 8px 0', fontSize: '1rem' }}>Issues:</h4>
                            <ul style={{ margin: 0, paddingLeft: '20px' }}>
                                {splitBrainData.warnings.map((warning, index) => (
                                    <li key={index} style={{ 
                                        marginBottom: '4px', 
                                        fontSize: '0.875rem',
                                        lineHeight: '1.4'
                                    }}>
                                        {warning}
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}

                    {/* Recommendations */}
                    {splitBrainData.recommendations && splitBrainData.recommendations.length > 0 && (
                        <div style={{ marginBottom: '16px' }}>
                            <h4 style={{ margin: '0 0 8px 0', fontSize: '1rem' }}>Recommended Actions:</h4>
                            <ol style={{ margin: 0, paddingLeft: '20px' }}>
                                {splitBrainData.recommendations.map((rec, index) => (
                                    <li key={index} style={{ 
                                        marginBottom: '4px', 
                                        fontSize: '0.875rem',
                                        lineHeight: '1.4'
                                    }}>
                                        {rec}
                                    </li>
                                ))}
                            </ol>
                        </div>
                    )}

                    {/* Action Buttons */}
                    <div style={{
                        display: 'flex',
                        gap: '12px',
                        marginTop: '16px'
                    }}>
                        <button
                            onClick={handleRefresh}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: 'rgba(255, 255, 255, 0.2)',
                                color: 'white',
                                border: '1px solid rgba(255, 255, 255, 0.3)',
                                borderRadius: '4px',
                                cursor: 'pointer',
                                fontSize: '0.875rem',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '4px'
                            }}
                        >
                            <RefreshCw size={14} />
                            Recheck
                        </button>
                    </div>
                </div>

                <button
                    onClick={handleDismiss}
                    style={{
                        background: 'none',
                        border: 'none',
                        color: 'white',
                        cursor: 'pointer',
                        padding: '4px',
                        marginLeft: '16px'
                    }}
                    title="Dismiss warning"
                >
                    <X size={20} />
                </button>
            </div>
        </div>
    );
};

export default SplitBrainWarning;