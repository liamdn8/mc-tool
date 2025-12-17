import React from 'react';
import { CheckCircle, XCircle, AlertCircle } from 'lucide-react';

const OverviewCards = ({ validationResults, checkLifecycle, checkEvents, calculateTableSeverity }) => {
    if (!validationResults) return null;

    const existence = validationResults.bucket_existence || {};
    const buckets = validationResults.buckets || [];
    const aliases = validationResults.aliases || [];
    const lifecycleTable = validationResults.lifecycle_table || [];
    const eventsTable = validationResults.events_table || [];

    // Calculate stats
    let totalChecks = 0;
    let existCount = 0;
    Object.values(existence).forEach(aliasMap => {
        Object.values(aliasMap).forEach(exists => {
            totalChecks++;
            if (exists) existCount++;
        });
    });

    const lifecycleSeverity = checkLifecycle ? calculateTableSeverity(lifecycleTable, buckets, aliases) : null;
    const eventsSeverity = checkEvents ? calculateTableSeverity(eventsTable, buckets, aliases) : null;

    const cards = [
        {
            label: 'Total Checks',
            value: totalChecks,
            color: 'var(--primary-color)'
        },
        {
            label: 'Buckets Found',
            value: existCount,
            color: existCount === totalChecks ? 'var(--success-color)' : 'var(--warning-color)'
        },
        {
            label: 'Buckets Missing',
            value: totalChecks - existCount,
            color: (totalChecks - existCount) > 0 ? 'var(--danger-color)' : 'var(--text-muted)'
        }
    ];

    if (lifecycleSeverity) {
        cards.push({
            label: 'Lifecycle Status',
            value: lifecycleSeverity === 'success' ? '✓ All Match' : lifecycleSeverity === 'warning' ? '⚠ Partial' : '✗ Mismatch',
            color: lifecycleSeverity === 'success' ? 'var(--success-color)' : lifecycleSeverity === 'warning' ? 'var(--warning-color)' : 'var(--danger-color)'
        });
    }

    if (eventsSeverity) {
        cards.push({
            label: 'Events Status',
            value: eventsSeverity === 'success' ? '✓ All Match' : eventsSeverity === 'warning' ? '⚠ Partial' : '✗ Mismatch',
            color: eventsSeverity === 'success' ? 'var(--success-color)' : eventsSeverity === 'warning' ? 'var(--warning-color)' : 'var(--danger-color)'
        });
    }

    return (
        <div className="stats-grid" style={{ marginBottom: '24px' }}>
            {cards.map(card => (
                <div key={card.label} className="stat-card">
                    <div className="stat-value" style={{ color: card.color }}>
                        {card.value}
                    </div>
                    <div className="stat-label">{card.label}</div>
                </div>
            ))}
        </div>
    );
};

export default OverviewCards;
