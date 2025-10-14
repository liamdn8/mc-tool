import React, { useState, useEffect } from 'react';
import { I18nProvider } from './utils/i18n';
import Header from './components/Header';
import Sidebar from './components/Sidebar';
import OverviewPage from './pages/OverviewPage';
import SitesPage from './pages/SitesPage';
import BucketsPage from './pages/BucketsPage';
import ReplicationPage from './pages/ReplicationPage';
import ConsistencyPage from './pages/ConsistencyPage';
import OperationsPage from './pages/OperationsPage';
import { loadAliases, loadSiteReplicationInfo } from './utils/api';

function App() {
    const [currentPage, setCurrentPage] = useState('overview');
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

    const renderCurrentPage = () => {
        const pageProps = {
            sites,
            replicationInfo,
            onRefresh: loadInitialData
        };

        switch (currentPage) {
            case 'overview':
                return <OverviewPage {...pageProps} />;
            case 'sites':
                return <SitesPage {...pageProps} />;
            case 'buckets':
                return <BucketsPage {...pageProps} />;
            case 'replication':
                return <ReplicationPage {...pageProps} />;
            case 'consistency':
                return <ConsistencyPage {...pageProps} />;
            case 'operations':
                return <OperationsPage {...pageProps} />;
            default:
                return <OverviewPage {...pageProps} />;
        }
    };

    return (
        <I18nProvider>
            <div className="app-container">
                <Header onRefresh={loadInitialData} />
                <div className="app-layout">
                    <Sidebar 
                        currentPage={currentPage} 
                        onPageChange={setCurrentPage} 
                    />
                    <main className="app-main">
                        {loading ? (
                            <div className="loading">
                                <div className="spinner"></div>
                            </div>
                        ) : (
                            renderCurrentPage()
                        )}
                    </main>
                </div>
            </div>
        </I18nProvider>
    );
}

export default App;