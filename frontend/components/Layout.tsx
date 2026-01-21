
import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useLanguage } from '../services/i18n';
import { useTheme } from '../lib/theme';
import {
  LayoutDashboard,
  Users,
  Key,
  ShieldAlert,
  Settings,
  LogOut,
  Menu,
  X,
  Bell,
  Globe,
  Network,
  Bot,
  Search,
  Shield,
  FolderTree,
  Server,
  FileSpreadsheet,
  Sun,
  Moon,
  Monitor
} from 'lucide-react';

interface LayoutProps {
  children: React.ReactNode;
  onLogout: () => void;
}

const Layout: React.FC<LayoutProps> = ({ children, onLogout }) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const location = useLocation();
  const { t, language, setLanguage } = useLanguage();
  const { mode, setMode, isDark } = useTheme();

  const navItems = [
    { path: '/', label: t('nav.dashboard'), icon: LayoutDashboard },
    { path: '/users', label: t('nav.users'), icon: Users },
    { path: '/bulk', label: 'Bulk Operations', icon: FileSpreadsheet },
    { path: '/groups', label: 'Groups', icon: FolderTree },
    { path: '/sessions', label: 'Sessions', icon: Key },
    { path: '/api-keys', label: t('nav.api_keys'), icon: Key },
    { path: '/oauth', label: t('nav.oauth'), icon: Globe },
    { path: '/oauth-clients', label: 'OAuth Clients', icon: Shield },
    { path: '/settings/access-control', label: 'Access Control', icon: ShieldAlert },
    { path: '/ldap', label: 'LDAP', icon: Server },
    { path: '/saml', label: 'SAML', icon: Shield },
    { path: '/ip-security', label: 'IP Security', icon: ShieldAlert },
    { path: '/audit-logs', label: t('nav.audit_logs'), icon: ShieldAlert },
    { path: '/settings', label: t('nav.settings'), icon: Settings },
  ];

  const devItems = [
    { path: '/developers/webhooks', label: t('nav.webhooks'), icon: Network },
    { path: '/developers/service-accounts', label: t('nav.service_accounts'), icon: Bot },
    { path: '/developers/token-inspector', label: t('nav.token_inspector'), icon: Search },
  ];

  const toggleSidebar = () => setIsSidebarOpen(!isSidebarOpen);

  return (
    <div className="flex h-screen bg-background overflow-hidden">
      {/* Mobile Sidebar Overlay */}
      {isSidebarOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-20 lg:hidden"
          onClick={() => setIsSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`
          fixed inset-y-0 left-0 z-30 w-64 bg-sidebar text-sidebar-foreground transform transition-transform duration-200 ease-in-out
          lg:translate-x-0 lg:static lg:inset-0
          ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'}
        `}
      >
        <div className="flex items-center justify-between h-16 px-6 bg-sidebar border-b border-sidebar-border">
          <span className="text-xl font-bold tracking-wider text-sidebar-accent">Auth Gateway</span>
          <button onClick={toggleSidebar} className="lg:hidden text-sidebar-muted hover:text-sidebar-foreground">
            <X size={24} />
          </button>
        </div>

        <div className="p-4 overflow-y-auto max-h-[calc(100vh-8rem)]">
          <div className="text-xs font-semibold text-sidebar-muted uppercase tracking-wider mb-4">
            {t('nav.menu')}
          </div>
          <nav className="space-y-1 mb-8">
            {navItems.map((item) => {
              const Icon = item.icon;
              // Check for exact match or sub-route match for active state
              // Special case: /settings should only match exactly, not /settings/access-control
              const isActive = item.path === '/settings'
                ? location.pathname === '/settings'
                : location.pathname === item.path || (item.path !== '/' && location.pathname.startsWith(item.path) && !location.pathname.startsWith('/developers'));
              return (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setIsSidebarOpen(false)}
                  className={`
                    flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors
                    ${isActive
                      ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/30'
                      : 'text-sidebar-foreground/80 hover:bg-sidebar-muted hover:text-sidebar-foreground'}
                  `}
                >
                  <Icon className="mr-3 h-5 w-5" />
                  {item.label}
                </Link>
              );
            })}
          </nav>

          <div className="text-xs font-semibold text-sidebar-muted uppercase tracking-wider mb-4">
            {t('nav.developers')}
          </div>
          <nav className="space-y-1">
            {devItems.map((item) => {
              const Icon = item.icon;
              const isActive = location.pathname.startsWith(item.path);
              return (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setIsSidebarOpen(false)}
                  className={`
                    flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors
                    ${isActive
                      ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/30'
                      : 'text-sidebar-foreground/80 hover:bg-sidebar-muted hover:text-sidebar-foreground'}
                  `}
                >
                  <Icon className="mr-3 h-5 w-5" />
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </div>

        <div className="absolute bottom-0 w-full p-4 border-t border-sidebar-border bg-sidebar">
          <button
            onClick={onLogout}
            className="flex items-center w-full px-4 py-3 text-sm font-medium text-destructive hover:bg-sidebar-muted hover:text-destructive/80 rounded-lg transition-colors"
          >
            <LogOut className="mr-3 h-5 w-5" />
            {t('nav.logout')}
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Header */}
        <header className="bg-card shadow-sm h-16 flex items-center justify-between px-6 z-10">
          <button onClick={toggleSidebar} className="lg:hidden text-muted-foreground hover:text-foreground">
            <Menu size={24} />
          </button>

          <div className="flex-1 px-4">
            {/* Breadcrumb or Search could go here */}
          </div>

          <div className="flex items-center gap-4">
            {/* Theme Switcher */}
            <div className="flex items-center border rounded-md overflow-hidden border-border bg-card">
              <button
                onClick={() => setMode('light')}
                className={`p-1.5 transition-colors ${mode === 'light' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent'}`}
                title="Light mode"
              >
                <Sun size={16} />
              </button>
              <button
                onClick={() => setMode('system')}
                className={`p-1.5 transition-colors ${mode === 'system' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent'}`}
                title="System preference"
              >
                <Monitor size={16} />
              </button>
              <button
                onClick={() => setMode('dark')}
                className={`p-1.5 transition-colors ${mode === 'dark' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-accent'}`}
                title="Dark mode"
              >
                <Moon size={16} />
              </button>
            </div>

            {/* Language Switcher */}
            <div className="flex items-center border rounded-md overflow-hidden border-border bg-card">
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
                <p className="text-sm font-medium text-foreground">Admin User</p>
                <p className="text-xs text-muted-foreground">Super Administrator</p>
              </div>
            </div>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
          {children}
        </main>
      </div>
    </div>
  );
};

export default Layout;