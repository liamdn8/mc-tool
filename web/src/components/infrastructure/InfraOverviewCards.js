import React from 'react';
import { CheckCircle, XCircle, AlertCircle } from 'lucide-react';

const InfraOverviewCards = ({ result }) => {
    if (!result || !result.summary) return null;

    const { summary } = result;
    const matchPercent = summary.totalComparisons > 0 
        ? ((summary.matchCount / summary.totalComparisons) * 100).toFixed(1)
        : 0;

    const cards = [
        {
            label: 'Total Comparisons',
            value: summary.totalComparisons,
            color: 'var(--primary-color)'
        },
        {
            label: 'Matches',
            value: `${summary.matchCount} (${matchPercent}%)`,
            color: summary.matchCount === summary.totalComparisons ? 'var(--success-color)' : 'var(--text-primary)',
            icon: CheckCircle
        },
        {
            label: 'Mismatches',
            value: summary.mismatchCount,
            color: summary.mismatchCount > 0 ? 'var(--danger-color)' : 'var(--text-muted)',
            icon: XCircle
        },
        {
            label: 'Not Found',
            value: summary.notFoundCount,
            color: summary.notFoundCount > 0 ? 'var(--warning-color)' : 'var(--text-muted)',
            icon: AlertCircle
        }
    ];

    return (
        <div className="stats-grid" style={{ marginBottom: '24px' }}>
            {cards.map(card => {
                const Icon = card.icon;
                return (
                    <div key={card.label} className="stat-card">
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <div className="stat-value" style={{ color: card.color }}>
                                {card.value}
                            </div>
                            {Icon && <Icon size={20} style={{ color: card.color }} />}
                        </div>
                        <div className="stat-label">{card.label}</div>
                    </div>
                );
            })}
        </div>
    );
};

export default InfraOverviewCards;
