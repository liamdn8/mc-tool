import React, { useState, useEffect } from 'react';
import { AlertTriangle, X, RefreshCw } from 'lucide-react';
import { checkSplitBrainStatus } from '../utils/api';
import { useI18n } from '../utils/i18n';

const SplitBrainWarning = ({ onRefresh }) => {
    const { t } = useI18n();
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
        <div className="p-4 rounded-lg my-4 border-2" style={{
            background: 'linear-gradient(135deg, #ff6b6b, #ee5a24)',
            borderColor: '#c0392b',
            boxShadow: '0 4px 12px rgba(255, 107, 107, 0.3)',
            color: 'white'
        }}>
            <div className="flex items-start justify-between">
                <div className="flex-1">
                    <div className="flex items-center mb-3">
                        <AlertTriangle size={24} className="mr-2" />
                        <h3 className="m-0 text-xl font-bold">
                            {t('split_brain_title', '⚠️ SPLIT BRAIN DETECTED')}
                        </h3>
                    </div>

                    <div className="mb-4">
                        <p className="m-0 mb-2 text-base">
                            <strong>
                                {t(
                                    'split_brain_clusters_detected',
                                    '{count} separate replication clusters detected!',
                                    { count: splitBrainData.clusterCount }
                                )}
                            </strong>
                        </p>
                        <p className="m-0 mb-3 text-sm" style={{ opacity: 0.9 }}>
                            {t('split_brain_risk', 'This configuration can cause data inconsistency and conflicts.')}
                        </p>
                    </div>

                    {/* Warnings */}
                    {splitBrainData.warnings && splitBrainData.warnings.length > 0 && (
                        <div className="mb-4">
                            <h4 className="m-0 mb-2 text-base">{t('split_brain_issues', 'Issues:')}</h4>
                            <ul className="m-0" style={{ paddingLeft: '20px' }}>
                                {splitBrainData.warnings.map((warning, index) => (
                                    <li key={index} className="mb-1 text-sm" style={{ lineHeight: '1.4' }}>
                                        {warning}
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}

                    {/* Recommendations */}
                    {splitBrainData.recommendations && splitBrainData.recommendations.length > 0 && (
                        <div className="mb-4">
                            <h4 className="m-0 mb-2 text-base">{t('split_brain_recommendations', 'Recommended Actions:')}</h4>
                            <ol className="m-0" style={{ paddingLeft: '20px' }}>
                                {splitBrainData.recommendations.map((rec, index) => (
                                    <li key={index} className="mb-1 text-sm" style={{ lineHeight: '1.4' }}>
                                        {rec}
                                    </li>
                                ))}
                            </ol>
                        </div>
                    )}

                    {/* Action Buttons */}
                    <div className="flex gap-3 mt-4">
                        <button
                            onClick={handleRefresh}
                            className="btn btn-sm flex items-center gap-1 text-white"
                            style={{
                                backgroundColor: 'rgba(255, 255, 255, 0.2)',
                                border: '1px solid rgba(255, 255, 255, 0.3)'
                            }}
                        >
                            <RefreshCw size={14} />
                            {t('split_brain_recheck', 'Recheck')}
                        </button>
                    </div>
                </div>

                <button
                    onClick={handleDismiss}
                    className="btn-icon text-white p-1"
                    style={{
                        background: 'none',
                        border: 'none',
                        marginLeft: '16px'
                    }}
                    title={t('split_brain_dismiss', 'Dismiss warning')}
                >
                    <X size={20} />
                </button>
            </div>
        </div>
    );
};

export default SplitBrainWarning;