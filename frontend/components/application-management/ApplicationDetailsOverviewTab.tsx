import React, { useState } from 'react';
import { Copy, Check, ExternalLink } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { formatDateTime } from '../../lib/date';
import type { Application } from '../../types';
import ApplicationSecretSection from '../ApplicationSecretSection';

interface ApplicationDetailsOverviewTabProps {
  application: Application;
  initialSecret?: string;
}

const ApplicationDetailsOverviewTab: React.FC<ApplicationDetailsOverviewTabProps> = ({
  application,
  initialSecret,
}) => {
  const { t } = useLanguage();
  const [copiedField, setCopiedField] = useState<string | null>(null);

  const copyToClipboard = (text: string, field: string) => {
    navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
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

      {application.allowed_auth_methods && application.allowed_auth_methods.length > 0 && (
        <div className="bg-card rounded-xl shadow-sm border border-border p-6 lg:col-span-2">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('apps.auth_methods.title')}</h2>
          <div className="flex flex-wrap gap-2">
            {application.allowed_auth_methods.map(method => (
              <span key={method} className="px-2 py-1 bg-primary/10 text-primary rounded-md text-xs font-medium">
                {t(`apps.auth_methods.${method}`)}
              </span>
            ))}
          </div>
        </div>
      )}

      <div className="lg:col-span-2">
        <ApplicationSecretSection application={application} initialSecret={initialSecret} />
      </div>

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
  );
};

export default ApplicationDetailsOverviewTab;
