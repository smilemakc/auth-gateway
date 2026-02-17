import React, { useState } from 'react';
import { useLanguage } from '../../services/i18n';
import { useApplication } from '../../services/appContext';
import Sidebar from './Sidebar';
import TopBar from './TopBar';
import { Boxes } from 'lucide-react';

interface LayoutProps {
  children: React.ReactNode;
  onLogout: () => void;
}

const Layout: React.FC<LayoutProps> = ({ children, onLogout }) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const { t } = useLanguage();
  const { currentApplication } = useApplication();

  const toggleSidebar = () => setIsSidebarOpen(prev => !prev);
  const closeSidebar = () => setIsSidebarOpen(false);

  return (
    <div className="flex h-screen bg-background overflow-hidden">
      {isSidebarOpen && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-20 lg:hidden"
          onClick={closeSidebar}
        />
      )}

      <Sidebar
        isOpen={isSidebarOpen}
        onClose={closeSidebar}
        onLogout={onLogout}
      />

      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <TopBar onToggleSidebar={toggleSidebar} />

        <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
          {currentApplication && (
            <div className="mb-4 flex items-center gap-2 rounded-lg border border-primary/20 bg-primary/5 px-4 py-2 text-sm text-primary">
              <Boxes className="h-4 w-4" />
              <span>
                {t('apps.filtering_by')}:{' '}
                <strong>{currentApplication.display_name || currentApplication.name}</strong>
              </span>
            </div>
          )}
          {children}
        </main>
      </div>
    </div>
  );
};

export default Layout;
