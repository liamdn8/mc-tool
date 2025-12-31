import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { 
    LayoutDashboard, 
    Globe, 
    Circle,
    GitCompare,
    List,
    Terminal,
    Zap,
    Activity,
    CheckSquare,
    RefreshCw,
    FlaskConical
} from 'lucide-react';
import { useI18n } from '../utils/i18n';
import { useContentsPanel } from '../contexts/ContentsPanelContext';
import PanelGroup, { Panel } from './PanelGroup';

const Sidebar = () => {
    const { t } = useI18n();
    const location = useLocation();
    const { contentsComponent } = useContentsPanel();

    const navItems = [
        { 
            id: 'overview', 
            path: '/overview',
            icon: LayoutDashboard, 
            label: t('overview') 
        },
        { 
            id: 'sites', 
            path: '/sites',
            icon: Globe, 
            label: t('sites') 
        },
        { 
            id: 'replication-operator', 
            path: '/replication-operator',
            icon: RefreshCw, 
            label: t('replication_operator', 'Replication Operator'),
            subItems: [
                {
                    id: 'compare',
                    path: '/replication-operator/compare',
                    icon: GitCompare,
                    label: t('operations_compare', 'Compare Buckets')
                }
            ]
        },
        { 
            id: 'tracing', 
            path: '/tracing',
            icon: Activity, 
            label: t('tracing', 'Tracing'),
            subItems: [
                {
                    id: 'trace-analyzer',
                    path: '/tracing/analyzer',
                    icon: Activity,
                    label: t('trace_error_analyzer', 'API Analyzer')
                },
                {
                    id: 'profiler',
                    path: '/tracing/profiler',
                    icon: Activity,
                    label: t('profiler', 'Profiler')
                }
            ]
        },
        { 
            id: 'validate', 
            path: '/validate',
            icon: CheckSquare, 
            label: t('validate', 'Validate'),
            subItems: [
                {
                    id: 'config-validate',
                    path: '/validate/configuration',
                    icon: List,
                    label: t('configuration_validate', 'Configuration Validation')
                }
            ]
        },
        {
            id: 'testing',
            path: '/testing',
            icon: FlaskConical,
            label: t('testing', 'Testing'),
            subItems: [
                {
                    id: 'testing',
                    path: '/testing/performance',
                    icon: FlaskConical,
                    label: t('performance_test', 'Performance Test')
                }
            ]
        },
        {
            id: 'terminal',
            path: '/terminal',
            icon: Terminal,
            label: t('terminal', 'Terminal')
        }
    ];

    const isActive = (path) => location.pathname === path;
    const isParentActive = (item) => {
        if (item.subItems) {
            return item.subItems.some(sub => location.pathname === sub.path) || location.pathname === item.path;
        }
        return location.pathname === item.path;
    };

    return (
        <aside className="app-sidebar">
            <PanelGroup direction="vertical">
                <Panel 
                    id="sidebar-navigation" 
                    title="Navigation"
                    collapsible={false}
                    size={contentsComponent ? 65 : 100}
                >
                    <nav className="sidebar-nav">
                        {navItems.map(item => {
                            const Icon = item.icon;
                            const hasSubItems = item.subItems && item.subItems.length > 0;
                            const isExpanded = isParentActive(item);
                            
                            return (
                                <div key={item.id}>
                                    <Link
                                        to={item.path}
                                        className={`nav-link ${isParentActive(item) ? 'active' : ''}`}
                                    >
                                        <Icon size={20} />
                                        <span>{item.label}</span>
                                    </Link>
                                    
                                    {hasSubItems && isExpanded && (
                                        <div className="sub-nav">
                                            {item.subItems.map(subItem => {
                                                const SubIcon = subItem.icon;
                                                return (
                                                    <Link
                                                        key={subItem.id}
                                                        to={subItem.path}
                                                        className={`nav-link sub-nav-link ${isActive(subItem.path) ? 'active' : ''}`}
                                                    >
                                                        <SubIcon size={16} />
                                                        <span>{subItem.label}</span>
                                                    </Link>
                                                );
                                            })}
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </nav>
                    <div className="sidebar-footer">
                        <div className="mc-status">
                            <Circle className="status-indicator" size={8} />
                            <span>{t('mc_tool_running', 'mc-tool running')}</span>
                        </div>
                    </div>
                </Panel>

                {contentsComponent && (
                    <Panel 
                        id="sidebar-contents" 
                        title="Contents"
                        collapsible={true}
                        size={35}
                    >
                        {contentsComponent}
                    </Panel>
                )}
            </PanelGroup>
        </aside>
    );
};

export default Sidebar;
