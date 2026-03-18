import React from 'react';
import { Link } from 'react-router-dom';
import { useLanguage } from '../../services/i18n';
import { useTheme } from '../../lib/theme';
import { useApplication } from '../../services/appContext';
import Breadcrumb from '../Breadcrumb';
import {
  Menu,
  Bell,
  Sun,
  Moon,
  Monitor,
} from 'lucide-react';

interface TopBarProps {
  onToggleSidebar: () => void;
}

const TopBar: React.FC<TopBarProps> = ({ onToggleSidebar }) => {
  const { t, language, setLanguage } = useLanguage();
  const { mode, setMode } = useTheme();
  const { currentApplicationId, applications, setCurrentApplicationId } = useApplication();

  return (
    <header className="bg-card shadow-sm h-16 flex items-center justify-between px-6 z-10">
      <button onClick={onToggleSidebar} className="lg:hidden text-muted-foreground hover:text-foreground">
        <Menu size={24} />
      </button>

      <div className="flex-1 px-4 hidden sm:block">
        <Breadcrumb />
      </div>

      <div className="flex items-center gap-4">
        {applications.length > 0 && (
          <div className="hidden md:flex items-center gap-2">
            <select
              value={currentApplicationId ?? ''}
              onChange={(e) => setCurrentApplicationId(e.target.value || null)}
              className="text-sm border rounded-md px-2 py-1.5 bg-card border-border text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 max-w-[200px]"
            >
              <option value="">{t('apps.all_applications')}</option>
              {applications.map((app) => (
                <option key={app.id} value={app.id}>
                  {app.display_name || app.name}
                </option>
              ))}
            </select>
            <Link to="/applications" className="text-xs text-primary hover:underline whitespace-nowrap">
              {t('apps.manage')}
            </Link>
          </div>
        )}

        <div className="flex items-center border rounded-md overflow-hidden border-border bg-card">
          <button
            onClick={() => setMode('light')}
            className={`p-1.5 transition-colors ${mode === 'light' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent'}`}
            title={t('layout.light_mode')}
          >
            <Sun size={16} />
          </button>
          <button
            onClick={() => setMode('system')}
            className={`p-1.5 transition-colors ${mode === 'system' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent'}`}
            title={t('layout.system_mode')}
          >
            <Monitor size={16} />
          </button>
          <button
            onClick={() => setMode('dark')}
            className={`p-1.5 transition-colors ${mode === 'dark' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent'}`}
            title={t('layout.dark_mode')}
          >
            <Moon size={16} />
          </button>
        </div>

        <div className="hidden sm:flex items-center border rounded-md overflow-hidden border-border bg-card">
          <button
            onClick={() => setLanguage('en')}
            className={`px-3 py-1.5 text-xs font-medium transition-colors ${language === 'en' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent'}`}
          >
            EN
          </button>
          <button
            onClick={() => setLanguage('ru')}
            className={`px-3 py-1.5 text-xs font-medium transition-colors ${language === 'ru' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent'}`}
          >
            RU
          </button>
        </div>

        <button className="relative p-2 text-muted-foreground hover:text-foreground transition-colors">
          <Bell size={20} />
          <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-destructive rounded-full border-2 border-card"></span>
        </button>
        <div className="flex items-center gap-3">
          <img
            src="https://picsum.photos/id/64/100/100"
            alt="Profile"
            className="h-8 w-8 rounded-full border border-border"
          />
          <div className="hidden md:block">
            <p className="text-sm font-medium text-foreground">{t('layout.admin_user')}</p>
            <p className="text-xs text-muted-foreground">{t('layout.super_admin')}</p>
          </div>
        </div>
      </div>
    </header>
  );
};

export default TopBar;
