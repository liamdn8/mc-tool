import React, { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { I18nProvider } from './utils/i18n';
import Header from './components/Header';
import Sidebar from './components/Sidebar';
import OverviewPage from './pages/OverviewPage';
import SitesPage from './pages/SitesPage';
import OperationsPage from './pages/OperationsPage';
import CompareOperations from './components/operations/CompareOperations';
import ChecklistOperations from './components/operations/ChecklistOperations';
import SiteOperations from './components/operations/SiteOperations';
import { loadAliases, loadSiteReplicationInfo } from './utils/api';

function App() {
    const [sites, setSites] = useState([]);
    const [replicationInfo, setReplicationInfo] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadInitialData();
    }, []);

    const loadInitialData = async () => {
        setLoading(true);
        try {
            // Load aliases (sites)
            const sitesData = await loadAliases();
            setSites(sitesData);
            
            // Load site replication info
            const replicationData = await loadSiteReplicationInfo();
            setReplicationInfo(replicationData.replicationInfo);
            
            // Update sites data with replication info
            if (replicationData.sites && replicationData.sites.length > 0) {
                setSites(replicationData.sites);
            }
        } catch (error) {
            console.error('Error loading initial data:', error);
        } finally {
            setLoading(false);
        }
    };

    const refreshData = async () => {
        try {
            // Load aliases (sites) - non-blocking refresh
            const sitesData = await loadAliases();
            setSites(sitesData);
            
            // Load site replication info
            const replicationData = await loadSiteReplicationInfo();
            setReplicationInfo(replicationData.replicationInfo);
            
            // Update sites data with replication info
            if (replicationData.sites && replicationData.sites.length > 0) {
                setSites(replicationData.sites);
            }
        } catch (error) {
            console.error('Error refreshing data:', error);
        }
    };

    const pageProps = {
        sites,
        replicationInfo,
        onRefresh: refreshData
    };

    return (
        <I18nProvider>
            <Router>
                <div className="app-container">
                    <Header onRefresh={loadInitialData} />
                    <div className="app-layout">
                        <Sidebar />
                        <main className="app-main">
                            {loading ? (
                                <div className="loading">
                                    <div className="spinner"></div>
                                </div>
                            ) : (
                                <Routes>
                                    <Route path="/" element={<Navigate to="/overview" replace />} />
                                    <Route path="/overview" element={<OverviewPage {...pageProps} />} />
                                    <Route path="/sites" element={<SitesPage {...pageProps} />} />
                                    <Route path="/operations" element={<OperationsPage {...pageProps} />} />
                                    <Route path="/operations/compare" element={<CompareOperations sites={sites} />} />
                                    <Route path="/operations/checklist" element={<ChecklistOperations />} />
                                    <Route path="/operations/site-operations" element={<SiteOperations hasReplication={replicationInfo?.enabled} />} />
                                </Routes>
                            )}
                        </main>
                    </div>
                </div>
            </Router>
        </I18nProvider>
    );
}

export default App;