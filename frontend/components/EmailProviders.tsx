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

const getProviderTypeLabel = (type: string) => {
  switch (type) {
    case 'smtp':
      return 'SMTP';
    case 'sendgrid':
      return 'SendGrid';
    case 'ses':
      return 'AWS SES';
    case 'mailgun':
      return 'Mailgun';
    default:
      return type;
  }
};

const EmailProviders: React.FC = () => {
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
      setTestResult({ success: false, message: 'Please enter an email address' });
      return;
    }

    setTestingProviderId(providerId);
    setTestResult(null);

    try {
      await testProvider.mutateAsync({ id: providerId, email: testEmail });
      setTestResult({ success: true, message: 'Test email sent successfully!' });
    } catch (err) {
      setTestResult({ success: false, message: err instanceof Error ? err.message : 'Failed to send test email' });
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
        Failed to load email providers. Please try again.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">
            {t('email.providers') || 'Email Providers'}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t('email.providers_desc') || 'Configure SMTP and email service providers'}
          </p>
        </div>
        <Link
          to="/settings/email-providers/new"
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('email.add_provider') || 'Add Provider'}
        </Link>
      </div>

      {/* Test Email Input */}
      <div className="bg-card rounded-xl shadow-sm border border-border p-4">
        <div className="flex items-center gap-4">
          <label className="text-sm font-medium text-foreground whitespace-nowrap">
            {t('email.test_email') || 'Test Email Address'}:
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
            {t('email.no_providers') || 'No Email Providers'}
          </h3>
          <p className="text-sm text-muted-foreground mb-6 text-center max-w-md">
            {t('email.no_providers_desc') || 'Configure an email provider to enable sending emails from the system.'}
          </p>
          <Link
            to="/settings/email-providers/new"
            className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm font-medium transition-colors"
          >
            <Plus size={18} />
            {t('email.add_first_provider') || 'Add Your First Provider'}
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
                        {provider.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground mt-1">
                      {getProviderTypeLabel(provider.type)}
                    </p>

                    {/* Provider Details */}
                    <div className="mt-3 space-y-1 text-sm">
                      {provider.type === 'smtp' && (
                        <>
                          <p className="text-muted-foreground">
                            <span className="font-medium text-foreground">Host:</span> {provider.smtp_host || 'N/A'}
                          </p>
                          <p className="text-muted-foreground">
                            <span className="font-medium text-foreground">Port:</span> {provider.smtp_port || 'N/A'}
                          </p>
                          <p className="text-muted-foreground">
                            <span className="font-medium text-foreground">Username:</span> {provider.smtp_username || 'N/A'}
                          </p>
                          <p className="text-muted-foreground">
                            <span className="font-medium text-foreground">TLS:</span> {provider.smtp_use_tls ? 'Enabled' : 'Disabled'}
                          </p>
                          <p className="text-muted-foreground">
                            <span className="font-medium text-foreground">Password:</span> {provider.has_smtp_password ? '••••••••' : 'Not set'}
                          </p>
                        </>
                      )}
                      {provider.type === 'sendgrid' && (
                        <p className="text-muted-foreground">
                          <span className="font-medium text-foreground">API Key:</span> {provider.has_sendgrid_api_key ? '••••••••' : 'Not set'}
                        </p>
                      )}
                      {provider.type === 'ses' && (
                        <>
                          <p className="text-muted-foreground">
                            <span className="font-medium text-foreground">Region:</span> {provider.ses_region || 'N/A'}
                          </p>
                          <p className="text-muted-foreground">
                            <span className="font-medium text-foreground">Access Key:</span> {provider.ses_access_key_id || 'N/A'}
                          </p>
                        </>
                      )}
                      {provider.type === 'mailgun' && (
                        <p className="text-muted-foreground">
                          <span className="font-medium text-foreground">Domain:</span> {provider.mailgun_domain || 'N/A'}
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
                    Test
                  </button>
                  <Link
                    to={`/settings/email-providers/${provider.id}`}
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
                      Are you sure you want to delete this provider? This action cannot be undone.
                    </p>
                  </div>
                  <div className="flex items-center gap-2 mt-3">
                    <button
                      onClick={() => handleDelete(provider.id)}
                      disabled={deleteProvider.isPending}
                      className="px-3 py-1.5 bg-destructive text-destructive-foreground rounded-lg text-sm font-medium hover:bg-destructive/90 disabled:opacity-50"
                    >
                      {deleteProvider.isPending ? 'Deleting...' : 'Delete'}
                    </button>
                    <button
                      onClick={() => setDeleteConfirm(null)}
                      className="px-3 py-1.5 bg-muted text-foreground rounded-lg text-sm font-medium hover:bg-muted/80"
                    >
                      Cancel
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
