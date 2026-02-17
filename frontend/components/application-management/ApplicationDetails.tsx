import React from 'react';
import { useParams, useSearchParams, useLocation } from 'react-router-dom';
import { Users, Palette, Settings, Mail, Globe, Bot, Plug } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { LoadingSpinner } from '../ui';
import { useApplicationDetail, useApplicationBranding } from '../../hooks/useApplications';
import ApplicationDetailsHeader from './ApplicationDetailsHeader';
import ApplicationDetailsOverviewTab from './ApplicationDetailsOverviewTab';
import ApplicationBrandingTab from './ApplicationBrandingTab';
import ApplicationUsersTab from './ApplicationUsersTab';
import ApplicationEmailTemplatesTab from './ApplicationEmailTemplatesTab';
import ApplicationOAuthProviders from '../ApplicationOAuthProviders';
import TelegramBots from '../TelegramBots';
import ApplicationIntegrationTab from '../ApplicationIntegrationTab';

type Tab = 'overview' | 'integration' | 'branding' | 'users' | 'templates' | 'oauth' | 'telegram';

const ApplicationDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const location = useLocation();
  const { t } = useLanguage();
  const [searchParams, setSearchParams] = useSearchParams();
  const createdSecret = (location.state as { secret?: string } | null)?.secret;
  const activeTab = (searchParams.get('tab') as Tab) || 'overview';
  const setActiveTab = (tab: Tab) => {
    setSearchParams(tab === 'overview' ? {} : { tab });
  };

  const { data: application, isLoading, error } = useApplicationDetail(id || '');
  const { data: branding } = useApplicationBranding(id || '');

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (error || !application) {
    return (
      <div className="bg-destructive/10 border border-destructive rounded-lg p-4 text-destructive">
        {t('apps.not_found')}
      </div>
    );
  }

  const tabs = [
    { id: 'overview' as Tab, label: t('apps.tab_overview'), icon: Settings },
    { id: 'integration' as Tab, label: t('apps.tabs.integration'), icon: Plug },
    { id: 'branding' as Tab, label: t('apps.tab_branding'), icon: Palette },
    { id: 'users' as Tab, label: t('apps.tab_users'), icon: Users },
    { id: 'templates' as Tab, label: t('apps.tab_templates'), icon: Mail },
    { id: 'oauth' as Tab, label: t('apps.tab_oauth'), icon: Globe },
    { id: 'telegram' as Tab, label: t('apps.tab_telegram'), icon: Bot },
  ];

  return (
    <div className="space-y-6">
      <ApplicationDetailsHeader
        applicationId={application.id}
        displayName={application.display_name}
        isActive={application.is_active}
        isSystem={application.is_system}
        logoUrl={branding?.logo_url}
      />

      <div className="border-b border-border">
        <nav className="flex gap-6">
          {tabs.map(tab => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-1 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === tab.id
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                }`}
              >
                <Icon size={18} />
                {tab.label}
              </button>
            );
          })}
        </nav>
      </div>

      {activeTab === 'overview' && (
        <ApplicationDetailsOverviewTab application={application} initialSecret={createdSecret} />
      )}
      {activeTab === 'integration' && <ApplicationIntegrationTab application={application} />}
      {activeTab === 'branding' && <ApplicationBrandingTab applicationId={id!} />}
      {activeTab === 'users' && <ApplicationUsersTab applicationId={id!} />}
      {activeTab === 'templates' && <ApplicationEmailTemplatesTab applicationId={id!} />}
      {activeTab === 'oauth' && <ApplicationOAuthProviders applicationId={id!} />}
      {activeTab === 'telegram' && <TelegramBots applicationId={id!} />}
    </div>
  );
};

export default ApplicationDetails;
