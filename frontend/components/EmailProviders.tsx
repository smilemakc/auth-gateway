import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Mail,
  Plus,
  Edit2,
  Trash2,
  Send,
  CheckCircle,
  XCircle,
  Loader2,
  Server,
  Cloud,
  AlertTriangle,
  Eye,
  EyeOff,
} from 'lucide-react';
import { useLanguage } from '../services/i18n';
import {
  useEmailProviders,
  useDeleteEmailProvider,
  useTestEmailProvider,
  EmailProvider,
} from '../hooks/useEmailProviders';

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

interface EmailProvidersProps {
  embedded?: boolean;
}

const EmailProviders: React.FC<EmailProvidersProps> = ({ embedded = false }) => {
  const { t } = useLanguage();
  const { data: providers, isLoading, error } = useEmailProviders();
  const deleteProvider = useDeleteEmailProvider();
  const testProvider = useTestEmailProvider();

  const [testEmail, setTestEmail] = useState('');
  const [testingProviderId, setTestingProviderId] = useState<string | null>(null);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  const handleTest = async (providerId: string) => {
    if (!testEmail) {
      setTestResult({ success: false, message: t('email.enter_email') });
      return;
    }

    setTestingProviderId(providerId);
    setTestResult(null);

    try {
      await testProvider.mutateAsync({ id: providerId, email: testEmail });
      setTestResult({ success: true, message: t('email.test_success') });
    } catch (err) {
      setTestResult({ success: false, message: err instanceof Error ? err.message : t('email.test_failed') });
    } finally {
      setTestingProviderId(null);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteProvider.mutateAsync(id);
      setDeleteConfirm(null);
    } catch (err) {
      console.error('Failed to delete provider:', err);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 text-destructive">
        {t('email.load_error')}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      {!embedded && (
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-foreground">
              {t('email.providers')}
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              {t('email.providers_desc')}
            </p>
          </div>
          <Link
            to="/email/providers/new"
            className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            <Plus size={18} />
            {t('email.add_provider')}
          </Link>
        </div>
      )}
      {embedded && (
        <div className="flex justify-end">
          <Link
            to="/email/providers/new"
            className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            <Plus size={18} />
            {t('email.add_provider')}
          </Link>
        </div>
      )}

      {/* Test Email Input */}
      <div className="bg-card rounded-xl shadow-sm border border-border p-4">
        <div className="flex items-center gap-4">
          <label className="text-sm font-medium text-foreground whitespace-nowrap">
            {t('email.test_email')}:
          </label>
          <input
            type="email"
            value={testEmail}
            onChange={(e) => setTestEmail(e.target.value)}
            placeholder="test@example.com"
            className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary text-sm"
          />
        </div>
        {testResult && (
          <div className={`mt-3 p-3 rounded-lg text-sm ${testResult.success ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'}`}>
            {testResult.message}
          </div>
        )}
      </div>

      {/* Providers List */}
      {(!providers || providers.length === 0) ? (
        <div className="flex flex-col items-center justify-center py-12 bg-card rounded-xl border border-border">
          <div className="p-4 bg-muted rounded-full mb-4">
            <Mail size={48} className="text-muted-foreground" />
          </div>
          <h3 className="text-lg font-semibold text-foreground mb-2">
            {t('email.no_providers')}
          </h3>
          <p className="text-sm text-muted-foreground mb-6 text-center max-w-md">
            {t('email.no_providers_desc')}
          </p>
          <Link
            to="/email/providers/new"
            className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm font-medium transition-colors"
          >
            <Plus size={18} />
            {t('email.add_first_provider')}
          </Link>
        </div>
      ) : (
        <div className="grid gap-4">
          {providers.map((provider: EmailProvider) => (
            <div
              key={provider.id}
              className="bg-card rounded-xl shadow-sm border border-border p-6"
            >
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

                    {/* Provider Details */}
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
                    onClick={() => handleTest(provider.id)}
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
                    onClick={() => setDeleteConfirm(provider.id)}
                    className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                  >
                    <Trash2 size={18} />
                  </button>
                </div>
              </div>

              {/* Delete Confirmation */}
              {deleteConfirm === provider.id && (
                <div className="mt-4 p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
                  <div className="flex items-center gap-3">
                    <AlertTriangle className="text-destructive" size={20} />
                    <p className="text-sm text-foreground">
                      {t('email.delete_confirm')}
                    </p>
                  </div>
                  <div className="flex items-center gap-2 mt-3">
                    <button
                      onClick={() => handleDelete(provider.id)}
                      disabled={deleteProvider.isPending}
                      className="px-3 py-1.5 bg-destructive text-destructive-foreground rounded-lg text-sm font-medium hover:bg-destructive/90 disabled:opacity-50"
                    >
                      {deleteProvider.isPending ? t('email.deleting') : t('common.delete')}
                    </button>
                    <button
                      onClick={() => setDeleteConfirm(null)}
                      className="px-3 py-1.5 bg-muted text-foreground rounded-lg text-sm font-medium hover:bg-muted/80"
                    >
                      {t('common.cancel')}
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default EmailProviders;
