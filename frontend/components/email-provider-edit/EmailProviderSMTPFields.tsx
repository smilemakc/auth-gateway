import React from 'react';
import { Server, Cloud, Eye, EyeOff, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { EmailProvider } from '../../hooks/useEmailProviders';

type ProviderType = 'smtp' | 'sendgrid' | 'ses' | 'mailgun';

interface EmailProviderFormData {
  type: ProviderType;
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

interface EmailProviderSMTPFieldsProps {
  formData: EmailProviderFormData;
  isNew: boolean;
  provider: EmailProvider | undefined;
  showPassword: boolean;
  onFormChange: (data: Partial<EmailProviderFormData>) => void;
  onTogglePassword: () => void;
}

const EmailProviderSMTPFields: React.FC<EmailProviderSMTPFieldsProps> = ({
  formData,
  isNew,
  provider,
  showPassword,
  onFormChange,
  onTogglePassword,
}) => {
  const { t } = useLanguage();

  if (formData.type === 'smtp') {
    return (
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
              onChange={(e) => onFormChange({ smtp_host: e.target.value })}
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
              onChange={(e) => onFormChange({ smtp_port: parseInt(e.target.value) || 587 })}
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
              onChange={(e) => onFormChange({ smtp_username: e.target.value })}
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
                onChange={(e) => onFormChange({ smtp_password: e.target.value })}
                className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022"
              />
              <button
                type="button"
                onClick={onTogglePassword}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <button type="button" onClick={() => onFormChange({ smtp_use_tls: !formData.smtp_use_tls })}
              className={`transition-colors ${formData.smtp_use_tls ? 'text-success' : 'text-muted-foreground'}`}>
              {formData.smtp_use_tls ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
            </button>
            <span className="text-sm font-medium text-foreground">
              {t('email.smtp_use_tls')}
            </span>
          </div>
        </div>
      </div>
    );
  }

  if (formData.type === 'sendgrid') {
    return (
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
              onChange={(e) => onFormChange({ sendgrid_api_key: e.target.value })}
              required={isNew}
              className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
              placeholder="SG.xxxxxxxxxxxxx"
            />
            <button
              type="button"
              onClick={onTogglePassword}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (formData.type === 'ses') {
    return (
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
              onChange={(e) => onFormChange({ ses_region: e.target.value })}
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
              onChange={(e) => onFormChange({ ses_access_key_id: e.target.value })}
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
                onChange={(e) => onFormChange({ ses_secret_access_key: e.target.value })}
                required={isNew}
                className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
              />
              <button
                type="button"
                onClick={onTogglePassword}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (formData.type === 'mailgun') {
    return (
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
              onChange={(e) => onFormChange({ mailgun_domain: e.target.value })}
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
                onChange={(e) => onFormChange({ mailgun_api_key: e.target.value })}
                required={isNew}
                className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="key-xxxxxxxxxxxxxxxxx"
              />
              <button
                type="button"
                onClick={onTogglePassword}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return null;
};

export default EmailProviderSMTPFields;
