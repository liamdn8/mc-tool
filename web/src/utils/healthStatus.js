// Helper functions for site health status display

export const getHealthBadgeClass = (status) => {
    switch (status) {
        case 'healthy':
            return 'badge-success';
        case 'checking':
            return 'badge-checking';
        case 'timeout':
        case 'unhealthy':
            return 'badge-danger';
        case 'error':
            return 'badge-warning';
        default:
            return 'badge-info';
    }
};

export const getHealthBadgeText = (status, t) => {
    switch (status) {
        case 'healthy':
            return t('status_healthy', '● Healthy');
        case 'checking':
            return t('status_checking', '◌ Checking...');
        case 'timeout':
            return t('status_timeout', '● Timeout');
        case 'unhealthy':
            return t('status_unhealthy', '● Unhealthy');
        case 'error':
            return t('status_error', '● Error');
        default:
            return t('status_unknown', '● Unknown');
    }
};

export const getHealthIcon = (status) => {
    switch (status) {
        case 'healthy':
            return '●'; // Filled circle
        case 'checking':
            return '◌'; // Empty circle
        case 'timeout':
        case 'unhealthy':
        case 'error':
            return '●';
        default:
            return '○';
    }
};
