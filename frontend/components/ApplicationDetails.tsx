import React, { useState } from 'react';
import { useParams, Link, useSearchParams } from 'react-router-dom';
import { ArrowLeft, Edit2, Boxes, Users, Palette, Settings, Copy, Check, ExternalLink, Mail, Globe, Bot } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useApplicationDetail, useApplicationBranding } from '../hooks/useApplications';
import ApplicationBrandingTab from './ApplicationBrandingTab';
import ApplicationUsersTab from './ApplicationUsersTab';
import ApplicationEmailTemplatesTab from './ApplicationEmailTemplatesTab';
import ApplicationOAuthProviders from './ApplicationOAuthProviders';
import TelegramBots from './TelegramBots';
import { formatDateTime } from '../lib/date';

type Tab = 'overview' | 'branding' | 'users' | 'templates' | 'oauth' | 'telegram';

const ApplicationDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { t } = useLanguage();
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = (searchParams.get('tab') as Tab) || 'overview';
  const setActiveTab = (tab: Tab) => {
    setSearchParams(tab === 'overview' ? {} : { tab });
  };
  const [copiedField, setCopiedField] = useState<string | null>(null);

  const { data: application, isLoading, error } = useApplicationDetail(id || '');
  const { data: branding } = useApplicationBranding(id || '');

  const copyToClipboard = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
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
    { id: 'branding' as Tab, label: t('apps.tab_branding'), icon: Palette },
    { id: 'users' as Tab, label: t('apps.tab_users'), icon: Users },
    { id: 'templates' as Tab, label: t('apps.tab_templates'), icon: Mail },
    { id: 'oauth' as Tab, label: t('apps.tab_oauth'), icon: Globe },
    { id: 'telegram' as Tab, label: t('apps.tab_telegram'), icon: Bot },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link
            to="/applications"
            className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
          >
            <ArrowLeft size={20} />
          </Link>
          <div className="flex items-center gap-4">
            <div className="w-14 h-14 rounded-xl bg-muted flex items-center justify-center shadow-sm">
              {branding?.logo_url ? (
                <img src={branding.logo_url} alt={application.display_name} className="w-10 h-10 object-contain" />
              ) : (
                <Boxes className="text-primary" size={28} />
              )}
            </div>
            <div>
              <h1 className="text-2xl font-bold text-foreground">{application.display_name}</h1>
              <div className="flex items-center gap-2 mt-1">
                <span className={`w-2 h-2 rounded-full ${application.is_active ? 'bg-success' : 'bg-muted-foreground'}`}></span>
                <span className="text-sm text-muted-foreground">
                  {application.is_active ? t('common.active') : t('common.inactive')}
                </span>
                {application.is_system && (
                  <>
                    <span className="text-muted-foreground">•</span>
                    <span className="text-sm text-warning">{t('apps.system')}</span>
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
        <Link
          to={`/applications/${id}/edit`}
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Edit2 size={18} />
          {t('common.edit')}
        </Link>
      </div>

      {/* Tabs */}
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

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Basic Information */}
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">{t('apps.basic_info')}</h2>
            <dl className="space-y-4">
              <div>
                <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                  {t('apps.name_slug')}
                </dt>
                <dd className="flex items-center gap-2">
                  <code className="flex-1 bg-muted rounded px-3 py-2 text-sm text-foreground font-mono border border-border">
                    {application.name}
                  </code>
                  <button
                    onClick={() => copyToClipboard(application.name, 'name')}
                    className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-accent rounded"
                  >
                    {copiedField === 'name' ? <Check size={14} /> : <Copy size={14} />}
                  </button>
                </dd>
              </div>

              <div>
                <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                  {t('apps.app_id')}
                </dt>
                <dd className="flex items-center gap-2">
                  <code className="flex-1 bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono border border-border truncate">
                    {application.id}
                  </code>
                  <button
                    onClick={() => copyToClipboard(application.id, 'id')}
                    className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-accent rounded"
                  >
                    {copiedField === 'id' ? <Check size={14} /> : <Copy size={14} />}
                  </button>
                </dd>
              </div>

              {application.description && (
                <div>
                  <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                    {t('apps.description')}
                  </dt>
                  <dd className="text-sm text-foreground">{application.description}</dd>
                </div>
              )}

              {application.homepage_url && (
                <div>
                  <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                    {t('apps.homepage')}
                  </dt>
                  <dd>
                    <a
                      href={application.homepage_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
                    >
                      {application.homepage_url}
                      <ExternalLink size={14} />
                    </a>
                  </dd>
                </div>
              )}

              <div>
                <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                  {t('common.created')}
                </dt>
                <dd className="text-sm text-foreground">
                  {formatDateTime(application.created_at)}
                </dd>
              </div>

              <div>
                <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                  {t('common.updated')}
                </dt>
                <dd className="text-sm text-foreground">
                  {formatDateTime(application.updated_at)}
                </dd>
              </div>
            </dl>
          </div>

          {/* Callback URLs */}
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">{t('apps.callback_urls')}</h2>
            {application.callback_urls && application.callback_urls.length > 0 ? (
              <ul className="space-y-2">
                {application.callback_urls.map((url, index) => (
                  <li key={index} className="flex items-center gap-2">
                    <code className="flex-1 bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono border border-border truncate">
                      {url}
                    </code>
                    <button
                      onClick={() => copyToClipboard(url, `url-${index}`)}
                      className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-accent rounded"
                    >
                      {copiedField === `url-${index}` ? <Check size={14} /> : <Copy size={14} />}
                    </button>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-muted-foreground italic">{t('apps.no_callbacks')}</p>
            )}
          </div>

          {/* Quick Stats */}
          <div className="bg-card rounded-xl shadow-sm border border-border p-6 lg:col-span-2">
            <h2 className="text-lg font-semibold text-foreground mb-4">{t('apps.integration')}</h2>
            <div className="bg-muted rounded-lg p-4">
              <p className="text-sm text-muted-foreground mb-3">
                {t('apps.integration_hint')}
              </p>
              <div className="flex items-center gap-2">
                <code className="flex-1 bg-background rounded px-3 py-2 text-sm text-foreground font-mono border border-border">
                  X-Application-ID: {application.id}
                </code>
                <button
                  onClick={() => copyToClipboard(`X-Application-ID: ${application.id}`, 'header')}
                  className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded"
                >
                  {copiedField === 'header' ? <Check size={16} /> : <Copy size={16} />}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'branding' && <ApplicationBrandingTab applicationId={id!} />}
      {activeTab === 'users' && <ApplicationUsersTab applicationId={id!} />}
      {activeTab === 'templates' && <ApplicationEmailTemplatesTab applicationId={id!} />}
      {activeTab === 'oauth' && <ApplicationOAuthProviders applicationId={id!} />}
      {activeTab === 'telegram' && <TelegramBots applicationId={id!} />}
    </div>
  );
};

export default ApplicationDetails;
