import React from 'react';
import { 
    LayoutDashboard, 
    Globe, 
    Settings,
    Circle
} from 'lucide-react';
import { useI18n } from '../utils/i18n';

const Sidebar = ({ currentPage, onPageChange }) => {
    const { t } = useI18n();

    const navItems = [
        { 
            id: 'overview', 
            icon: LayoutDashboard, 
            label: t('overview') 
        },
        { 
            id: 'sites', 
            icon: Globe, 
            label: t('sites') 
        },
        { 
            id: 'operations', 
            icon: Settings, 
            label: t('operations') 
        }
    ];

    return (
        <aside className="app-sidebar">
            <nav className="sidebar-nav">
                {navItems.map(item => {
                    const Icon = item.icon;
                    return (
                        <div
                            key={item.id}
                            className={`nav-link ${currentPage === item.id ? 'active' : ''}`}
                            onClick={() => onPageChange(item.id)}
                        >
                            <Icon size={20} />
                            <span>{item.label}</span>
                        </div>
                    );
                })}
            </nav>
            <div className="sidebar-footer">
                <div className="mc-status">
                    <Circle className="status-indicator" size={8} />
                    <span>mc-tool running</span>
                </div>
            </div>
        </aside>
    );
};

export default Sidebar;