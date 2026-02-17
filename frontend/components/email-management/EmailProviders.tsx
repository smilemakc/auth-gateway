import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Mail, Plus, Loader2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import {
  useEmailProviders,
  useDeleteEmailProvider,
  useTestEmailProvider,
  EmailProvider,
} from '../../hooks/useEmailProviders';
import { logger } from '@/lib/logger';
import { EmailProviderCard } from './EmailProviderCard';

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
      logger.error('Failed to delete provider:', err);
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
            <EmailProviderCard
              key={provider.id}
              provider={provider}
              testEmail={testEmail}
              testingProviderId={testingProviderId}
              deleteConfirmId={deleteConfirm}
              isDeleting={deleteProvider.isPending}
              onTest={handleTest}
              onDelete={handleDelete}
              onDeleteConfirm={setDeleteConfirm}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default EmailProviders;
