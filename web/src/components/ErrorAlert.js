import React, { useEffect } from 'react';
import { X, AlertCircle } from 'lucide-react';

const ErrorAlert = ({ message, onClose, autoClose = true, duration = 8000 }) => {
    useEffect(() => {
        if (autoClose && onClose) {
            const timer = setTimeout(() => {
                onClose();
            }, duration);
            return () => clearTimeout(timer);
        }
    }, [autoClose, duration, onClose]);

    if (!message) return null;

    return (
        <div style={{
            position: 'fixed',
            top: '20px',
            right: '20px',
            maxWidth: '500px',
            backgroundColor: '#fee',
            border: '1px solid #fcc',
            borderRadius: '8px',
            padding: '16px',
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            zIndex: 9999,
            animation: 'slideIn 0.3s ease-out'
        }}>
            <div style={{ display: 'flex', gap: '12px' }}>
                <AlertCircle 
                    size={24} 
                    style={{ 
                        color: '#dc2626', 
                        flexShrink: 0,
                        marginTop: '2px'
                    }} 
                />
                <div style={{ flex: 1 }}>
                    <h4 style={{ 
                        margin: '0 0 8px 0', 
                        fontSize: '16px', 
                        fontWeight: '600',
                        color: '#991b1b'
                    }}>
                        Operation Failed
                    </h4>
                    <div style={{ 
                        fontSize: '14px', 
                        color: '#7f1d1d',
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-word'
                    }}>
                        {message}
                    </div>
                </div>
                {onClose && (
                    <button
                        onClick={onClose}
                        style={{
                            background: 'none',
                            border: 'none',
                            cursor: 'pointer',
                            padding: '4px',
                            color: '#991b1b',
                            flexShrink: 0
                        }}
                        title="Close"
                    >
                        <X size={20} />
                    </button>
                )}
            </div>
            <style>{`
                @keyframes slideIn {
                    from {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                    to {
                        transform: translateX(0);
                        opacity: 1;
                    }
                }
            `}</style>
        </div>
    );
};

export default ErrorAlert;
