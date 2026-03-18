import React from 'react';
import { Mail, FileText } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface BrandingCompanySectionProps {
  companyName: string;
  supportEmail: string;
  termsUrl: string;
  privacyUrl: string;
  customCss: string;
  onCompanyNameChange: (value: string) => void;
  onSupportEmailChange: (value: string) => void;
  onTermsUrlChange: (value: string) => void;
  onPrivacyUrlChange: (value: string) => void;
  onCustomCssChange: (value: string) => void;
}

const BrandingCompanySection: React.FC<BrandingCompanySectionProps> = ({
  companyName,
  supportEmail,
  termsUrl,
  privacyUrl,
  customCss,
  onCompanyNameChange,
  onSupportEmailChange,
  onTermsUrlChange,
  onPrivacyUrlChange,
  onCustomCssChange,
}) => {
  const { t } = useLanguage();

  return (
    <>
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
            <Mail className="text-primary" size={20} />
          </div>
          <h2 className="text-lg font-semibold text-foreground">{t('brand.company_info')}</h2>
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              {t('brand.company')}
            </label>
            <input
              type="text"
              value={companyName}
              onChange={e => onCompanyNameChange(e.target.value)}
              className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
              placeholder="My Company"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              {t('brand.support_email')}
            </label>
            <input
              type="email"
              value={supportEmail}
              onChange={e => onSupportEmailChange(e.target.value)}
              className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
              placeholder="support@example.com"
            />
          </div>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
            <FileText className="text-primary" size={20} />
          </div>
          <h2 className="text-lg font-semibold text-foreground">{t('brand.legal')}</h2>
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              {t('brand.terms')}
            </label>
            <input
              type="url"
              value={termsUrl}
              onChange={e => onTermsUrlChange(e.target.value)}
              className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
              placeholder="https://example.com/terms"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              {t('brand.privacy')}
            </label>
            <input
              type="url"
              value={privacyUrl}
              onChange={e => onPrivacyUrlChange(e.target.value)}
              className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
              placeholder="https://example.com/privacy"
            />
          </div>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6 lg:col-span-2">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('brand.custom_css')}</h2>
        <p className="text-sm text-muted-foreground mb-4">
          {t('brand.custom_css_hint')}
        </p>
        <textarea
          value={customCss}
          onChange={e => onCustomCssChange(e.target.value)}
          rows={6}
          className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          placeholder=".login-form { /* custom styles */ }"
        />
      </div>
    </>
  );
};

export default BrandingCompanySection;
