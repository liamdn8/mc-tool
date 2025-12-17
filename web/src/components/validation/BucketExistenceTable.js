import React from 'react';
import { CheckCircle, XCircle, AlertCircle } from 'lucide-react';

const BucketExistenceTable = ({ validationResults }) => {
    if (!validationResults || !validationResults.bucket_existence) return null;

    const existence = validationResults.bucket_existence;
    const buckets = validationResults.buckets || [];
    const aliases = validationResults.aliases || [];

    // Calculate severity
    let totalChecks = 0;
    let foundCount = 0;
    Object.values(existence).forEach(aliasMap => {
        Object.values(aliasMap).forEach(exists => {
            totalChecks++;
            if (exists) foundCount++;
        });
    });

    const severity = foundCount === totalChecks ? 'success' : 
                    foundCount < totalChecks / 2 ? 'danger' : 'warning';

    return (
        <div style={{ marginBottom: '24px' }} id="bucket_existence">
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
                <h4 className="card-title" style={{ margin: 0, fontSize: '16px' }}>
                    Bucket Existence
                </h4>
                {severity === 'success' && <CheckCircle size={20} style={{ color: 'var(--success-color)' }} />}
                {severity === 'warning' && <AlertCircle size={20} style={{ color: 'var(--warning-color)' }} />}
                {severity === 'danger' && <XCircle size={20} style={{ color: 'var(--danger-color)' }} />}
            </div>

            <div className="table-container">
                <div style={{ overflowX: 'auto' }}>
                    <table className="table">
                        <thead>
                            <tr>
                                <th style={{ position: 'sticky', left: 0, backgroundColor: 'var(--bg-secondary)', zIndex: 1 }}>
                                    Bucket
                                </th>
                                {aliases.map(alias => (
                                    <th key={alias} style={{ textAlign: 'center', minWidth: '120px' }}>
                                        {alias}
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody>
                            {buckets.map((bucket) => (
                                <tr key={bucket}>
                                    <td style={{ 
                                        position: 'sticky', 
                                        left: 0, 
                                        backgroundColor: 'var(--bg-primary)', 
                                        zIndex: 1,
                                        fontFamily: 'monospace',
                                        fontWeight: 500
                                    }}>
                                        {bucket}
                                    </td>
                                    {aliases.map(alias => {
                                        const exists = existence[bucket] && existence[bucket][alias];
                                        return (
                                            <td key={alias} style={{ textAlign: 'center' }}>
                                                {exists ? (
                                                    <span className="badge badge-success">
                                                        <CheckCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                        Exists
                                                    </span>
                                                ) : (
                                                    <span className="badge badge-danger">
                                                        <XCircle size={14} style={{ marginRight: '4px', display: 'inline-block', verticalAlign: 'middle' }} />
                                                        Missing
                                                    </span>
                                                )}
                                            </td>
                                        );
                                    })}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

export default BucketExistenceTable;
