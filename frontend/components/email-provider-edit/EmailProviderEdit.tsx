import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { Save, Loader2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { LoadingSpinner, PageHeader } from '../ui';
import {
  useEmailProvider,
  useCreateEmailProvider,
  useUpdateEmailProvider,
  CreateEmailProviderRequest,
  UpdateEmailProviderRequest,
} from '../../hooks/useEmailProviders';
import { logger } from '@/lib/logger';
import EmailProviderTestSection from './EmailProviderTestSection';
import EmailProviderSMTPFields from './EmailProviderSMTPFields';

type ProviderType = 'smtp' | 'sendgrid' | 'ses' | 'mailgun';

interface FormData {
  name: string;
  type: ProviderType;
  is_active: boolean;
  smtp_host: string;
  smtp_port: number;
  smtp_username: string;
  smtp_password: string;
  smtp_use_tls: boolean;
  sendgrid_api_key: string;
  ses_region: string;
  ses_access_key_id: string;
  ses_secret_access_key: string;
  mailgun_domain: string;
  mailgun_api_key: string;
}

const EmailProviderEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isNew = id === 'new';

  const { data: provider, isLoading } = useEmailProvider(id || '');
  const createProvider = useCreateEmailProvider();
  const updateProvider = useUpdateEmailProvider();

  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState<FormData>({
    name: '',
    type: 'smtp',
    is_active: true,
    smtp_host: '',
    smtp_port: 587,
    smtp_username: '',
    smtp_password: '',
    smtp_use_tls: true,
    sendgrid_api_key: '',
    ses_region: 'us-east-1',
    ses_access_key_id: '',
    ses_secret_access_key: '',
    mailgun_domain: '',
    mailgun_api_key: '',
  });

  useEffect(() => {
    if (provider && !isNew) {
      setFormData({
        name: provider.name || '',
        type: provider.type || 'smtp',
        is_active: provider.is_active ?? true,
        smtp_host: provider.smtp_host || '',
        smtp_port: provider.smtp_port || 587,
        smtp_username: provider.smtp_username || '',
        smtp_password: '',
        smtp_use_tls: provider.smtp_use_tls ?? true,
        sendgrid_api_key: '',
        ses_region: provider.ses_region || 'us-east-1',
        ses_access_key_id: provider.ses_access_key_id || '',
        ses_secret_access_key: '',
        mailgun_domain: provider.mailgun_domain || '',
        mailgun_api_key: '',
      });
    }
  }, [provider, isNew]);

  const handleFormChange = (partial: Partial<FormData>) => {
    setFormData(prev => ({ ...prev, ...partial }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      if (isNew) {
        const createData: CreateEmailProviderRequest = {
          name: formData.name,
          type: formData.type,
          is_active: formData.is_active,
        };

        if (formData.type === 'smtp') {
          createData.smtp_host = formData.smtp_host;
          createData.smtp_port = formData.smtp_port;
          createData.smtp_username = formData.smtp_username;
          createData.smtp_password = formData.smtp_password;
          createData.smtp_use_tls = formData.smtp_use_tls;
        } else if (formData.type === 'sendgrid') {
          createData.sendgrid_api_key = formData.sendgrid_api_key;
        } else if (formData.type === 'ses') {
          createData.ses_region = formData.ses_region;
          createData.ses_access_key_id = formData.ses_access_key_id;
          createData.ses_secret_access_key = formData.ses_secret_access_key;
        } else if (formData.type === 'mailgun') {
          createData.mailgun_domain = formData.mailgun_domain;
          createData.mailgun_api_key = formData.mailgun_api_key;
        }

        await createProvider.mutateAsync(createData);
      } else {
        const updateData: UpdateEmailProviderRequest = {
          name: formData.name,
          is_active: formData.is_active,
        };

        if (formData.type === 'smtp') {
          updateData.smtp_host = formData.smtp_host;
          updateData.smtp_port = formData.smtp_port;
          updateData.smtp_username = formData.smtp_username;
          if (formData.smtp_password) {
            updateData.smtp_password = formData.smtp_password;
          }
          updateData.smtp_use_tls = formData.smtp_use_tls;
        } else if (formData.type === 'sendgrid' && formData.sendgrid_api_key) {
          updateData.sendgrid_api_key = formData.sendgrid_api_key;
        } else if (formData.type === 'ses') {
          updateData.ses_region = formData.ses_region;
          updateData.ses_access_key_id = formData.ses_access_key_id;
          if (formData.ses_secret_access_key) {
            updateData.ses_secret_access_key = formData.ses_secret_access_key;
          }
        } else if (formData.type === 'mailgun') {
          updateData.mailgun_domain = formData.mailgun_domain;
          if (formData.mailgun_api_key) {
            updateData.mailgun_api_key = formData.mailgun_api_key;
          }
        }

        await updateProvider.mutateAsync({ id: id!, data: updateData });
      }

      navigate('/settings/email-providers');
    } catch (err) {
      logger.error('Failed to save provider:', err);
    }
  };

  if (isLoading && !isNew) {
    return <LoadingSpinner />;
  }

  const isPending = createProvider.isPending || updateProvider.isPending;

  return (
    <div className="space-y-6">
      <PageHeader
        title={isNew ? t('email.new_provider') : t('email.edit_provider')}
        subtitle={isNew ? t('email.new_provider_desc') : t('email.edit_provider_desc')}
        onBack={() => navigate('/settings/email-providers')}
      />

      <form onSubmit={handleSubmit} className="space-y-6">
        <EmailProviderTestSection
          name={formData.name}
          type={formData.type}
          isActive={formData.is_active}
          isNew={isNew}
          onNameChange={(value) => handleFormChange({ name: value })}
          onTypeChange={(value) => handleFormChange({ type: value })}
          onActiveToggle={() => handleFormChange({ is_active: !formData.is_active })}
        />

        <EmailProviderSMTPFields
          formData={formData}
          isNew={isNew}
          provider={provider}
          showPassword={showPassword}
          onFormChange={handleFormChange}
          onTogglePassword={() => setShowPassword(prev => !prev)}
        />

        <div className="flex items-center justify-end gap-4">
          <Link
            to="/settings/email-providers"
            className="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
          >
            {t('common.cancel')}
          </Link>
          <button
            type="submit"
            disabled={isPending}
            className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
          >
            {isPending ? (
              <Loader2 size={18} className="animate-spin" />
            ) : (
              <Save size={18} />
            )}
            {isPending ? t('common.saving') : t('common.save')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default EmailProviderEdit;
