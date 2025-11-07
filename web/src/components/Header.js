import React from 'react';
import { Package, RefreshCw } from 'lucide-react';
import { useI18n } from '../utils/i18n';

const Header = ({ onRefresh }) => {
    const { currentLang, setLanguage, t } = useI18n();

    const handleLanguageChange = (e) => {
        setLanguage(e.target.value);
    };

    return (
        <header className="app-header">
            <div className="header-left">
                <Package className="app-logo" size={32} />
                <div className="app-title">
                    <h1>{t('app_title', 'MinIO Site Replication')}</h1>
                    <span className="app-subtitle">{t('app_subtitle', 'Management Console')}</span>
                </div>
            </div>
            <div className="header-right">
                <select 
                    className="language-selector" 
                    value={currentLang} 
                    onChange={handleLanguageChange}
                >
                    <option value="en">🇬🇧 {t('language_english', 'English')}</option>
                    <option value="vi">🇻🇳 {t('language_vietnamese', 'Vietnamese')}</option>
                </select>
                <button 
                    className="btn-icon" 
                    title={t('refresh', 'Refresh')} 
                    aria-label={t('refresh', 'Refresh')}
                    onClick={onRefresh}
                >
                    <RefreshCw size={20} />
                </button>
            </div>
        </header>
    );
};

export default Header;