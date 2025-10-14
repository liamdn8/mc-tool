// Helper utility functions for React components

export function formatNumber(num) {
    return new Intl.NumberFormat().format(num);
}

export function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
}

export function getBadgeClass(status) {
    switch (status) {
        case 'healthy':
        case 'configured':
        case 'fully_replicated':
        case 'completed':
        case 'success':
            return 'badge-success';
        case 'warning':
        case 'partial_replication':
        case 'pending':
            return 'badge-warning';
        case 'unhealthy':
        case 'not_configured':
        case 'failed':
        case 'error':
            return 'badge-danger';
        default:
            return 'badge-secondary';
    }
}

export function getStatusText(status) {
    switch (status) {
        case 'healthy':
            return 'Healthy';
        case 'unhealthy':
            return 'Unhealthy';
        case 'configured':
            return 'Configured';
        case 'not_configured':
            return 'Not Configured';
        case 'fully_replicated':
            return 'Fully Replicated';
        case 'partial_replication':
            return 'Partial Replication';
        case 'completed':
            return 'Completed';
        case 'pending':
            return 'Pending';
        case 'failed':
            return 'Failed';
        case 'success':
            return 'Success';
        default:
            return status;
    }
}

export function formatDate(dateString) {
    if (!dateString) return '-';
    try {
        return new Date(dateString).toLocaleString();
    } catch (error) {
        return '-';
    }
}

export function showNotification(type, message, duration = 5000) {
    // Simple notification - in a real app you might want to use a toast library
    alert(`${type.toUpperCase()}: ${message}`);
}