// Utility functions for formatting and notifications

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

// Enhanced notification system with auto-reload
export function showNotification(type, message, options = {}) {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <div class="notification-content">
            ${type === 'success' ? '✓' : type === 'error' ? '✗' : 'ℹ'} ${message}
        </div>
    `;
    
    document.body.appendChild(notification);
    
    // Trigger animation
    setTimeout(() => {
        notification.classList.add('show');
    }, 10);
    
    // Auto-remove duration (default 5 seconds)
    const duration = options.duration || 5000;
    
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => {
            if (document.body.contains(notification)) {
                document.body.removeChild(notification);
            }
        }, 300);
    }, duration);
    
    return notification;
}

export function showErrorDialog(title, message) {
    const dialog = document.createElement('div');
    dialog.className = 'modal-overlay error-dialog';
    dialog.innerHTML = `
        <div class="modal error-modal">
            <div class="modal-header error-header">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="8" x2="12" y2="12"></line>
                    <line x1="12" y1="16" x2="12.01" y2="16"></line>
                </svg>
                <h3>${title}</h3>
            </div>
            <div class="modal-body">
                <pre class="error-message">${message}</pre>
            </div>
            <div class="modal-footer">
                <button class="btn-primary" id="closeErrorDialog">OK</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(dialog);
    
    dialog.querySelector('#closeErrorDialog').onclick = () => {
        document.body.removeChild(dialog);
    };
    
    // Close on overlay click
    dialog.onclick = (e) => {
        if (e.target === dialog) {
            document.body.removeChild(dialog);
        }
    };
    
    // Close on Escape key
    const handleEscape = (e) => {
        if (e.key === 'Escape') {
            document.body.removeChild(dialog);
            document.removeEventListener('keydown', handleEscape);
        }
    };
    document.addEventListener('keydown', handleEscape);
}

export function showSiteSelectionDialog(sites, direction) {
    return new Promise((resolve) => {
        const dialog = document.createElement('div');
        dialog.className = 'modal-overlay';
        dialog.innerHTML = `
            <div class="modal">
                <div class="modal-header">
                    <h3>${direction === 'resync-from' ? 'Select Source Site' : 'Select Target Site'}</h3>
                </div>
                <div class="modal-body">
                    <p>${direction === 'resync-from' 
                        ? 'Select the site to pull data FROM:' 
                        : 'Select the site to push data TO:'}</p>
                    <div class="site-selection-list">
                        ${sites.map(site => `
                            <button class="site-selection-item" data-alias="${site.alias}">
                                <div>
                                    <div class="site-name">${site.alias}</div>
                                    <div class="site-url">${site.url}</div>
                                </div>
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <polyline points="9 18 15 12 9 6"></polyline>
                                </svg>
                            </button>
                        `).join('')}
                    </div>
                </div>
                <div class="modal-footer">
                    <button class="btn-secondary" id="cancelSiteSelection">Cancel</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(dialog);
        
        // Add event listeners
        dialog.querySelectorAll('.site-selection-item').forEach(btn => {
            btn.onclick = () => {
                resolve(btn.dataset.alias);
                document.body.removeChild(dialog);
            };
        });
        
        dialog.querySelector('#cancelSiteSelection').onclick = () => {
            resolve(null);
            document.body.removeChild(dialog);
        };
        
        // Close on overlay click
        dialog.onclick = (e) => {
            if (e.target === dialog) {
                resolve(null);
                document.body.removeChild(dialog);
            }
        };
    });
}

// Auto-reload functionality for replication actions
export function autoReloadAfterReplicationAction(delay = 2000) {
    setTimeout(() => {
        showNotification('info', 'Reloading to show updated status...', { duration: 1500 });
        setTimeout(() => {
            window.location.reload();
        }, 1500);
    }, delay);
}

// Enhanced success handler with auto-reload
export function handleReplicationSuccess(message, autoReload = true) {
    showNotification('success', message);
    
    if (autoReload) {
        autoReloadAfterReplicationAction();
    }
}

// Enhanced error handler
export function handleReplicationError(title, error, response) {
    console.error(`${title}:`, error);
    
    if (response && !response.ok) {
        showErrorDialog(title, error.error || error.message || 'Unknown error occurred');
    } else {
        showErrorDialog('Connection Error', 'Failed to connect to the server. Please try again.');
    }
}