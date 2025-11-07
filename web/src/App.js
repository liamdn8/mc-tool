import React, { useState, useEffect, useRef } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { I18nProvider } from './utils/i18n';
import Header from './components/Header';
import Sidebar from './components/Sidebar';
import OverviewPage from './pages/OverviewPage';
import SitesPage from './pages/SitesPage';
import ReplicationOperatorPage from './pages/ReplicationOperatorPage';
import TracingPage from './pages/TracingPage';
import ChecklistPage from './pages/ChecklistPage';
import CompareOperations from './components/operations/CompareOperations';
import ChecklistOperations from './components/operations/ChecklistOperations';
import SiteOperations from './components/operations/SiteOperations';
import TraceOperations from './components/operations/TraceOperations';
import TerminalPage from './pages/TerminalPage';
import { loadAliases, loadSiteReplicationInfo, checkAliasHealth } from './utils/api';

function App() {
    const [sites, setSites] = useState([]);
    const [replicationInfo, setReplicationInfo] = useState(null);
    const [loading, setLoading] = useState(true);
    const [checkingReplication, setCheckingReplication] = useState(true);
    const replicationCheckRef = useRef({ loaded: false, timeout: null });

    useEffect(() => {
        loadInitialData();
    }, []);

    const loadInitialData = async () => {
        setLoading(true);
        setCheckingReplication(true);
        replicationCheckRef.current = { loaded: false, timeout: null };
        
        try {
            // Load aliases immediately (with "checking" status)
            const sitesData = await loadAliases();
            setSites(sitesData);
            setLoading(false); // Show UI immediately
            
            // Check health for each alias in parallel (non-blocking)
            // Will load replication info after first healthy site is found
            checkAliasesHealthAsync(sitesData);
        } catch (error) {
            console.error('Error loading initial data:', error);
            setLoading(false);
            setCheckingReplication(false);
        }
    };

    const checkAliasesHealthAsync = async (sitesData) => {
        // Set a timeout to stop checking if no healthy sites found
        replicationCheckRef.current.timeout = setTimeout(() => {
            if (!replicationCheckRef.current.loaded) {
                console.log('Timeout: No healthy sites found within 5 seconds, stopping replication check');
                setCheckingReplication(false);
            }
        }, 5000);
        
        // Check health for each alias in parallel, update status as each completes
        sitesData.forEach(site => {
            checkAliasHealth(site.name)
                .then(healthData => {
                    setSites(prevSites => 
                        prevSites.map(s => 
                            s.name === healthData.alias 
                                ? { ...s, healthy: healthData.healthy, status: healthData.status }
                                : s
                        )
                    );
                    
                    // Only load replication info if site is healthy (skip timeout/error/unhealthy)
                    if (healthData.healthy && healthData.status === 'healthy' && !replicationCheckRef.current.loaded) {
                        replicationCheckRef.current.loaded = true;
                        if (replicationCheckRef.current.timeout) {
                            clearTimeout(replicationCheckRef.current.timeout);
                        }
                        console.log(`First healthy site found: ${healthData.alias}, loading replication info...`);
                        
                        loadSiteReplicationInfo().then(replicationData => {
                            console.log('Replication info loaded:', replicationData);
                            setReplicationInfo(replicationData.replicationInfo);
                            setCheckingReplication(false);
                            
                            // Update sites data with latest replication info
                            if (replicationData.sites && replicationData.sites.length > 0) {
                                setSites(prevSites => {
                                    return replicationData.sites.map(site => {
                                        const existingSite = prevSites.find(s => s.name === site.name);
                                        return {
                                            ...site,
                                            healthy: existingSite?.healthy ?? false,
                                            status: existingSite?.status ?? 'checking'
                                        };
                                    });
                                });
                            }
                        }).catch(error => {
                            console.error('Error loading replication info:', error);
                            setCheckingReplication(false);
                        });
                    } else if (healthData.status === 'timeout' || healthData.status === 'error' || healthData.status === 'unhealthy') {
                        // If site is unhealthy/timeout, log and skip it for replication check
                        console.log(`Skipping ${healthData.alias} for replication check - status: ${healthData.status}`);
                    }
                })
                .catch(error => {
                    console.error(`Error checking health for ${site.name}:`, error);
                    setSites(prevSites => 
                        prevSites.map(s => 
                            s.name === site.name 
                                ? { ...s, healthy: false, status: 'error' }
                                : s
                        )
                    );
                    console.log(`Skipping ${site.name} for replication check - health check failed`);
                });
        });
    };

    const refreshData = async () => {
        try {
            // Load aliases (sites) - non-blocking refresh
            const sitesData = await loadAliases();
            setSites(sitesData);
            
            // Check health async
            checkAliasesHealthAsync(sitesData);
            
            // Load site replication info
            const replicationData = await loadSiteReplicationInfo();
            setReplicationInfo(replicationData.replicationInfo);
            
            // Update sites data with replication info
            if (replicationData.sites && replicationData.sites.length > 0) {
                setSites(prevSites => {
                    return replicationData.sites.map(site => {
                        const existingSite = prevSites.find(s => s.name === site.name);
                        return {
                            ...site,
                            healthy: existingSite?.healthy ?? false,
                            status: existingSite?.status ?? 'checking'
                        };
                    });
                });
            }
        } catch (error) {
            console.error('Error refreshing data:', error);
        }
    };

    const pageProps = {
        sites,
        replicationInfo,
        checkingReplication,
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
                                    
                                    {/* Replication Operator Routes */}
                                    <Route path="/replication-operator" element={<ReplicationOperatorPage {...pageProps} />} />
                                    <Route path="/replication-operator/compare" element={<CompareOperations sites={sites} />} />
                                    <Route path="/replication-operator/resync" element={<SiteOperations hasReplication={replicationInfo?.enabled} />} />
                                    
                                    {/* Tracing Routes */}
                                    <Route path="/tracing" element={<TracingPage sites={sites} />} />
                                    <Route path="/tracing/analyzer" element={<TraceOperations sites={sites} />} />
                                    
                                    {/* Checklist Routes */}
                                    <Route path="/checklist" element={<ChecklistPage sites={sites} />} />
                                    <Route path="/checklist/configuration" element={<ChecklistOperations />} />
                                    
                                    {/* Terminal */}
                                    <Route path="/terminal" element={<TerminalPage />} />
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