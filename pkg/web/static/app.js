// Main application controller
import { getCurrentLang, setCurrentLang, updateI18n } from './js/utils/i18n.js';
import { loadAliases, loadSiteReplicationInfo } from './js/utils/api.js';
import { renderOverviewPage, updateOverviewStats, renderSitesList } from './js/pages/overview.js';
import { renderSitesPage } from './js/pages/sites.js';
import { renderBucketsPage } from './js/pages/buckets.js';
import { renderReplicationPage } from './js/pages/replication.js';
import { renderConsistencyPage } from './js/pages/consistency.js';
import { renderOperationsPage } from './js/pages/operations.js';

// Global app state
let sites = [];
let replicationInfo = null;

// Main App Controller
class MCToolApp {
    constructor() {
        this.currentPage = 'overview';
        this.sites = [];
        this.replicationInfo = null;
    }

    async init() {
        this.initializeEventListeners();
        await this.loadInitialData();
        updateI18n();
    }

    initializeEventListeners() {
        // Language selector
        document.getElementById('languageSelector').addEventListener('change', (e) => {
            setCurrentLang(e.target.value);
        });

        // Navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const page = link.dataset.page;
                this.navigateToPage(page);
            });
        });

        // Refresh button
        document.getElementById('refreshBtn').addEventListener('click', () => {
            this.loadInitialData();
        });

        // Add site button
        document.getElementById('addSiteBtn')?.addEventListener('click', () => {
            // TODO: Show add site modal
            alert('Add Site functionality will be implemented');
        });

        // Modal close
        document.querySelectorAll('.modal-close').forEach(btn => {
            btn.addEventListener('click', () => {
                btn.closest('.modal').classList.remove('active');
            });
        });
    }

    navigateToPage(pageName) {
        // Update active nav link
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        document.querySelector(`[data-page="${pageName}"]`).classList.add('active');

        // Update active page
        document.querySelectorAll('.page').forEach(page => {
            page.classList.remove('active');
        });
        document.getElementById(`${pageName}-page`).classList.add('active');

        this.currentPage = pageName;

        // Load page-specific data
        this.loadPageData(pageName);
    }

    async loadInitialData() {
        try {
            // Load aliases (sites)
            this.sites = await loadAliases();
            
            // Load site replication info
            const replicationData = await loadSiteReplicationInfo();
            this.replicationInfo = replicationData.replicationInfo;
            
            // Update sites data with replication info
            if (replicationData.sites && replicationData.sites.length > 0) {
                this.sites = replicationData.sites;
            }
            
            // Update global references for backward compatibility
            sites = this.sites;
            replicationInfo = this.replicationInfo;
            
            // Update overview stats and render sites list
            updateOverviewStats(this.sites, this.replicationInfo);
            renderSitesList(this.sites);
            
        } catch (error) {
            console.error('Error loading initial data:', error);
        }
    }

    async loadPageData(pageName) {
        switch(pageName) {
            case 'overview':
                await renderOverviewPage();
                break;
            case 'sites':
                await renderSitesPage(this.sites);
                break;
            case 'buckets':
                await renderBucketsPage(this.sites);
                break;
            case 'replication':
                await renderReplicationPage();
                break;
            case 'consistency':
                await renderConsistencyPage();
                break;
            case 'operations':
                await renderOperationsPage();
                break;
        }
    }

    // Public methods for external access
    getCurrentSites() {
        return this.sites;
    }

    getReplicationInfo() {
        return this.replicationInfo;
    }

    async refresh() {
        await this.loadInitialData();
        await this.loadPageData(this.currentPage);
    }
}

// Initialize app
const app = new MCToolApp();

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    app.init();
});

// Export app instance for global access
window.app = app;

// Export for legacy compatibility
window.loadSites = () => app.refresh();
export default app;