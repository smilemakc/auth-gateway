import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeft, Save, Loader2, Server, Cloud, Eye, EyeOff, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import {
  useEmailProvider,
  useCreateEmailProvider,
  useUpdateEmailProvider,
  CreateEmailProviderRequest,
  UpdateEmailProviderRequest,
} from '../hooks/useEmailProviders';
import { logger } from '@/lib/logger';

type ProviderType = 'smtp' | 'sendgrid' | 'ses' | 'mailgun';

const EmailProviderEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isNew = id === 'new';

  const { data: provider, isLoading } = useEmailProvider(id || '');
  const createProvider = useCreateEmailProvider();
  const updateProvider = useUpdateEmailProvider();

  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState<{
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
  }>({
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      if (isNew) {
        const createData: CreateEmailProviderRequest = {
          name: formData.name,
          type: formData.type,
          is_active: formData.is_active,
        };

        // Add type-specific fields
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

        // Add type-specific fields
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
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  const isPending = createProvider.isPending || updateProvider.isPending;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link
          to="/settings/email-providers"
          className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
        >
          <ArrowLeft size={20} />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-foreground">
            {isNew ? t('email.new_provider') : t('email.edit_provider')}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            {isNew
              ? t('email.new_provider_desc')
              : t('email.edit_provider_desc')}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Settings */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">
            {t('email.basic_settings')}
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('email.provider_name')} *
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="My SMTP Server"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('email.provider_type')} *
              </label>
              <select
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value as ProviderType })}
                disabled={!isNew}
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
              >
                <option value="smtp">SMTP</option>
                <option value="sendgrid">SendGrid</option>
                <option value="ses">AWS SES</option>
                <option value="mailgun">Mailgun</option>
              </select>
            </div>

            <div className="flex items-center gap-3">
              <button type="button" onClick={() => setFormData(prev => ({ ...prev, is_active: !prev.is_active }))}
                className={`transition-colors ${formData.is_active ? 'text-success' : 'text-muted-foreground'}`}>
                {formData.is_active ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
              </button>
              <span className="text-sm font-medium text-foreground">
                {t('email.provider_active')}
              </span>
            </div>
          </div>
        </div>

        {/* SMTP Settings */}
        {formData.type === 'smtp' && (
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <div className="flex items-center gap-3 mb-4">
              <Server className="text-primary" size={20} />
              <h2 className="text-lg font-semibold text-foreground">
                {t('email.smtp_settings')}
              </h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('email.smtp_host')} *
                </label>
                <input
                  type="text"
                  value={formData.smtp_host}
                  onChange={(e) => setFormData({ ...formData, smtp_host: e.target.value })}
                  required
                  className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="smtp.gmail.com"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('email.smtp_port')} *
                </label>
                <input
                  type="number"
                  value={formData.smtp_port}
                  onChange={(e) => setFormData({ ...formData, smtp_port: parseInt(e.target.value) || 587 })}
                  required
                  className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="587"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('email.smtp_username')}
                </label>
                <input
                  type="text"
                  value={formData.smtp_username}
                  onChange={(e) => setFormData({ ...formData, smtp_username: e.target.value })}
                  className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="user@example.com"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('email.smtp_password')}
                  {!isNew && provider?.has_smtp_password && (
                    <span className="text-muted-foreground font-normal ml-2">(leave blank to keep current)</span>
                  )}
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={formData.smtp_password}
                    onChange={(e) => setFormData({ ...formData, smtp_password: e.target.value })}
                    className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="••••••••"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <button type="button" onClick={() => setFormData(prev => ({ ...prev, smtp_use_tls: !prev.smtp_use_tls }))}
                  className={`transition-colors ${formData.smtp_use_tls ? 'text-success' : 'text-muted-foreground'}`}>
                  {formData.smtp_use_tls ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                </button>
                <span className="text-sm font-medium text-foreground">
                  {t('email.smtp_use_tls')}
                </span>
              </div>
            </div>
          </div>
        )}

        {/* SendGrid Settings */}
        {formData.type === 'sendgrid' && (
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <div className="flex items-center gap-3 mb-4">
              <Cloud className="text-primary" size={20} />
              <h2 className="text-lg font-semibold text-foreground">
                {t('email.sendgrid_settings')}
              </h2>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('email.api_key')} *
                {!isNew && provider?.has_sendgrid_api_key && (
                  <span className="text-muted-foreground font-normal ml-2">(leave blank to keep current)</span>
                )}
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={formData.sendgrid_api_key}
                  onChange={(e) => setFormData({ ...formData, sendgrid_api_key: e.target.value })}
                  required={isNew}
                  className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="SG.xxxxxxxxxxxxx"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* AWS SES Settings */}
        {formData.type === 'ses' && (
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <div className="flex items-center gap-3 mb-4">
              <Cloud className="text-primary" size={20} />
              <h2 className="text-lg font-semibold text-foreground">
                {t('email.ses_settings')}
              </h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('email.ses_region')} *
                </label>
                <select
                  value={formData.ses_region}
                  onChange={(e) => setFormData({ ...formData, ses_region: e.target.value })}
                  className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="us-east-1">US East (N. Virginia)</option>
                  <option value="us-east-2">US East (Ohio)</option>
                  <option value="us-west-2">US West (Oregon)</option>
                  <option value="eu-west-1">EU (Ireland)</option>
                  <option value="eu-central-1">EU (Frankfurt)</option>
                  <option value="ap-southeast-1">Asia Pacific (Singapore)</option>
                  <option value="ap-southeast-2">Asia Pacific (Sydney)</option>
                  <option value="ap-northeast-1">Asia Pacific (Tokyo)</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('email.ses_access_key')} *
                </label>
                <input
                  type="text"
                  value={formData.ses_access_key_id}
                  onChange={(e) => setFormData({ ...formData, ses_access_key_id: e.target.value })}
                  required
                  className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="AKIAIOSFODNN7EXAMPLE"
                />
              </div>

              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('email.ses_secret_key')} *
                  {!isNew && provider?.has_ses_secret_access_key && (
                    <span className="text-muted-foreground font-normal ml-2">(leave blank to keep current)</span>
                  )}
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={formData.ses_secret_access_key}
                    onChange={(e) => setFormData({ ...formData, ses_secret_access_key: e.target.value })}
                    required={isNew}
                    className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Mailgun Settings */}
        {formData.type === 'mailgun' && (
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <div className="flex items-center gap-3 mb-4">
              <Cloud className="text-primary" size={20} />
              <h2 className="text-lg font-semibold text-foreground">
                {t('email.mailgun_settings')}
              </h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('email.mailgun_domain')} *
                </label>
                <input
                  type="text"
                  value={formData.mailgun_domain}
                  onChange={(e) => setFormData({ ...formData, mailgun_domain: e.target.value })}
                  required
                  className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                  placeholder="mg.example.com"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('email.api_key')} *
                  {!isNew && provider?.has_mailgun_api_key && (
                    <span className="text-muted-foreground font-normal ml-2">(leave blank to keep current)</span>
                  )}
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={formData.mailgun_api_key}
                    onChange={(e) => setFormData({ ...formData, mailgun_api_key: e.target.value })}
                    required={isNew}
                    className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="key-xxxxxxxxxxxxxxxxx"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Actions */}
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
