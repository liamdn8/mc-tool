import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { 
    LayoutDashboard, 
    Globe, 
    Settings,
    Circle,
    GitCompare,
    List,
    ShieldAlert,
    Terminal,
    Zap,
    Activity
} from 'lucide-react';
import { useI18n } from '../utils/i18n';

const Sidebar = () => {
    const { t } = useI18n();
    const location = useLocation();

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
            id: 'operations', 
            path: '/operations',
            icon: Settings, 
            label: t('operations'),
            subItems: [
                {
                    id: 'compare',
                    path: '/operations/compare',
                    icon: GitCompare,
                    label: t('operations_compare', 'Compare Buckets')
                },
                {
                    id: 'trace',
                    path: '/operations/trace',
                    icon: Activity,
                    label: t('operations_trace', 'Trace Error Analyzer')
                },
                {
                    id: 'checklist',
                    path: '/operations/checklist',
                    icon: List,
                    label: t('operations_checklist', 'Configuration Checklist')
                },
                {
                    id: 'operator-utilities',
                    path: '/operations/operator-utilities',
                    icon: ShieldAlert,
                    label: t('operations_operator_utilities', 'Operator Utilities')
                },
                {
                    id: 'site-operations',
                    path: '/operations/site-operations',
                    icon: Zap,
                    label: t('operations_site_operations', 'Site Operations')
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
        </aside>
    );
};

export default Sidebar;
