
import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useLanguage } from '../services/i18n';
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
  Search
} from 'lucide-react';

interface LayoutProps {
  children: React.ReactNode;
  onLogout: () => void;
}

const Layout: React.FC<LayoutProps> = ({ children, onLogout }) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const location = useLocation();
  const { t, language, setLanguage } = useLanguage();

  const navItems = [
    { path: '/', label: t('nav.dashboard'), icon: LayoutDashboard },
    { path: '/users', label: t('nav.users'), icon: Users },
    { path: '/sessions', label: 'Sessions', icon: Key },
    { path: '/api-keys', label: t('nav.api_keys'), icon: Key },
    { path: '/oauth', label: t('nav.oauth'), icon: Globe },
    { path: '/roles', label: 'Roles', icon: ShieldAlert },
    { path: '/permissions', label: 'Permissions', icon: ShieldAlert },
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
    <div className="flex h-screen bg-gray-100 overflow-hidden">
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
          fixed inset-y-0 left-0 z-30 w-64 bg-slate-900 text-white transform transition-transform duration-200 ease-in-out
          lg:translate-x-0 lg:static lg:inset-0
          ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'}
        `}
      >
        <div className="flex items-center justify-between h-16 px-6 bg-slate-950 border-b border-slate-800">
          <span className="text-xl font-bold tracking-wider text-blue-400">Auth Gateway</span>
          <button onClick={toggleSidebar} className="lg:hidden text-gray-400 hover:text-white">
            <X size={24} />
          </button>
        </div>

        <div className="p-4 overflow-y-auto max-h-[calc(100vh-8rem)]">
          <div className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-4">
            {t('nav.menu')}
          </div>
          <nav className="space-y-1 mb-8">
            {navItems.map((item) => {
              const Icon = item.icon;
              // Check for exact match or sub-route match for active state
              const isActive = location.pathname === item.path || (item.path !== '/' && location.pathname.startsWith(item.path) && !location.pathname.startsWith('/developers'));
              return (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setIsSidebarOpen(false)}
                  className={`
                    flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors
                    ${isActive
                      ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/50'
                      : 'text-slate-300 hover:bg-slate-800 hover:text-white'}
                  `}
                >
                  <Icon className="mr-3 h-5 w-5" />
                  {item.label}
                </Link>
              );
            })}
          </nav>

          <div className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-4">
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
                      ? 'bg-blue-600 text-white shadow-lg shadow-blue-900/50'
                      : 'text-slate-300 hover:bg-slate-800 hover:text-white'}
                  `}
                >
                  <Icon className="mr-3 h-5 w-5" />
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </div>

        <div className="absolute bottom-0 w-full p-4 border-t border-slate-800 bg-slate-900">
          <button
            onClick={onLogout}
            className="flex items-center w-full px-4 py-3 text-sm font-medium text-red-400 hover:bg-slate-800 hover:text-red-300 rounded-lg transition-colors"
          >
            <LogOut className="mr-3 h-5 w-5" />
            {t('nav.logout')}
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Header */}
        <header className="bg-white shadow-sm h-16 flex items-center justify-between px-6 z-10">
          <button onClick={toggleSidebar} className="lg:hidden text-gray-500 hover:text-gray-700">
            <Menu size={24} />
          </button>

          <div className="flex-1 px-4">
            {/* Breadcrumb or Search could go here */}
          </div>

          <div className="flex items-center gap-4">
            {/* Language Switcher */}
            <div className="flex items-center border rounded-md overflow-hidden border-gray-200">
              <button
                onClick={() => setLanguage('en')}
                className={`px-3 py-1.5 text-xs font-medium transition-colors ${language === 'en' ? 'bg-blue-100 text-blue-700' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
              >
                EN
              </button>
              <button
                onClick={() => setLanguage('ru')}
                className={`px-3 py-1.5 text-xs font-medium transition-colors ${language === 'ru' ? 'bg-blue-100 text-blue-700' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
              >
                RU
              </button>
            </div>

            <button className="relative p-2 text-gray-400 hover:text-gray-600 transition-colors">
              <Bell size={20} />
              <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full border-2 border-white"></span>
            </button>
            <div className="flex items-center gap-3">
              <img
                src="https://picsum.photos/id/64/100/100"
                alt="Profile"
                className="h-8 w-8 rounded-full border border-gray-200"
              />
              <div className="hidden md:block">
                <p className="text-sm font-medium text-gray-700">Admin User</p>
                <p className="text-xs text-gray-500">Super Administrator</p>
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