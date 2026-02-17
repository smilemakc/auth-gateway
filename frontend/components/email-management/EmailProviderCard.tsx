import React from 'react';
import { Link } from 'react-router-dom';
import {
  Edit2,
  Trash2,
  Send,
  Loader2,
  Server,
  Cloud,
  Mail,
  AlertTriangle,
} from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { EmailProvider } from '../../hooks/useEmailProviders';

const getProviderIcon = (type: string) => {
  switch (type) {
    case 'smtp':
      return <Server size={20} />;
    case 'sendgrid':
    case 'ses':
    case 'mailgun':
      return <Cloud size={20} />;
    default:
      return <Mail size={20} />;
  }
};

const getProviderTypeLabel = (type: string, t: (key: string) => string) => {
  switch (type) {
    case 'smtp':
      return t('email.provider_smtp');
    case 'sendgrid':
      return t('email.provider_sendgrid');
    case 'ses':
      return t('email.provider_ses');
    case 'mailgun':
      return t('email.provider_mailgun');
    default:
      return type;
  }
};

interface EmailProviderCardProps {
  provider: EmailProvider;
  testEmail: string;
  testingProviderId: string | null;
  deleteConfirmId: string | null;
  isDeleting: boolean;
  onTest: (providerId: string) => void;
  onDelete: (id: string) => void;
  onDeleteConfirm: (id: string | null) => void;
}

export const EmailProviderCard: React.FC<EmailProviderCardProps> = ({
  provider,
  testEmail,
  testingProviderId,
  deleteConfirmId,
  isDeleting,
  onTest,
  onDelete,
  onDeleteConfirm,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-4">
          <div className={`p-3 rounded-lg ${provider.is_active ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'}`}>
            {getProviderIcon(provider.type)}
          </div>
          <div>
            <div className="flex items-center gap-3">
              <h3 className="text-lg font-semibold text-foreground">{provider.name}</h3>
              <span className={`px-2 py-0.5 text-xs font-medium rounded-full ${
                provider.is_active
                  ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                  : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
              }`}>
                {provider.is_active ? t('common.active') : t('common.inactive')}
              </span>
            </div>
            <p className="text-sm text-muted-foreground mt-1">
              {getProviderTypeLabel(provider.type, t)}
            </p>

            <div className="mt-3 space-y-1 text-sm">
              {provider.type === 'smtp' && (
                <>
                  <p className="text-muted-foreground">
                    <span className="font-medium text-foreground">{t('email.host')}:</span> {provider.smtp_host || 'N/A'}
                  </p>
                  <p className="text-muted-foreground">
                    <span className="font-medium text-foreground">{t('email.port')}:</span> {provider.smtp_port || 'N/A'}
                  </p>
                  <p className="text-muted-foreground">
                    <span className="font-medium text-foreground">{t('email.username')}:</span> {provider.smtp_username || 'N/A'}
                  </p>
                  <p className="text-muted-foreground">
                    <span className="font-medium text-foreground">{t('email.tls')}:</span> {provider.smtp_use_tls ? t('common.enabled') : t('common.disabled')}
                  </p>
                  <p className="text-muted-foreground">
                    <span className="font-medium text-foreground">{t('email.password')}:</span> {provider.has_smtp_password ? '••••••••' : t('common.not_set')}
                  </p>
                </>
              )}
              {provider.type === 'sendgrid' && (
                <p className="text-muted-foreground">
                  <span className="font-medium text-foreground">{t('email.api_key')}:</span> {provider.has_sendgrid_api_key ? '••••••••' : t('common.not_set')}
                </p>
              )}
              {provider.type === 'ses' && (
                <>
                  <p className="text-muted-foreground">
                    <span className="font-medium text-foreground">{t('email.region')}:</span> {provider.ses_region || 'N/A'}
                  </p>
                  <p className="text-muted-foreground">
                    <span className="font-medium text-foreground">{t('email.access_key')}:</span> {provider.ses_access_key_id || 'N/A'}
                  </p>
                </>
              )}
              {provider.type === 'mailgun' && (
                <p className="text-muted-foreground">
                  <span className="font-medium text-foreground">{t('email.domain')}:</span> {provider.mailgun_domain || 'N/A'}
                </p>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => onTest(provider.id)}
            disabled={testingProviderId === provider.id || !testEmail}
            className="flex items-center gap-2 px-3 py-2 text-sm font-medium text-primary hover:bg-primary/10 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {testingProviderId === provider.id ? (
              <Loader2 size={16} className="animate-spin" />
            ) : (
              <Send size={16} />
            )}
            {t('common.test')}
          </button>
          <Link
            to={`/email/providers/${provider.id}`}
            className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
          >
            <Edit2 size={18} />
          </Link>
          <button
            onClick={() => onDeleteConfirm(provider.id)}
            className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
          >
            <Trash2 size={18} />
          </button>
        </div>
      </div>

      {deleteConfirmId === provider.id && (
        <div className="mt-4 p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
          <div className="flex items-center gap-3">
            <AlertTriangle className="text-destructive" size={20} />
            <p className="text-sm text-foreground">
              {t('email.delete_confirm')}
            </p>
          </div>
          <div className="flex items-center gap-2 mt-3">
            <button
              onClick={() => onDelete(provider.id)}
              disabled={isDeleting}
              className="px-3 py-1.5 bg-destructive text-destructive-foreground rounded-lg text-sm font-medium hover:bg-destructive/90 disabled:opacity-50"
            >
              {isDeleting ? t('email.deleting') : t('common.delete')}
            </button>
            <button
              onClick={() => onDeleteConfirm(null)}
              className="px-3 py-1.5 bg-muted text-foreground rounded-lg text-sm font-medium hover:bg-muted/80"
            >
              {t('common.cancel')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
