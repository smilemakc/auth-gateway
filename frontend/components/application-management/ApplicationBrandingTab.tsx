import React, { useState, useEffect } from 'react';
import { Loader2, Save } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useApplicationBranding, useUpdateApplicationBranding } from '../../hooks/useApplications';
import { logger } from '@/lib/logger';
import BrandingVisualSection from './BrandingVisualSection';
import BrandingColorsSection from './BrandingColorsSection';
import BrandingCompanySection from './BrandingCompanySection';

interface ApplicationBrandingTabProps {
  applicationId: string;
}

const ApplicationBrandingTab: React.FC<ApplicationBrandingTabProps> = ({ applicationId }) => {
  const { t } = useLanguage();
  const { data: branding, isLoading } = useApplicationBranding(applicationId);
  const updateBranding = useUpdateApplicationBranding();

  const [formData, setFormData] = useState({
    logo_url: '',
    favicon_url: '',
    primary_color: '#3b82f6',
    secondary_color: '#64748b',
    background_color: '#ffffff',
    custom_css: '',
    company_name: '',
    support_email: '',
    terms_url: '',
    privacy_url: '',
  });

  useEffect(() => {
    if (branding) {
      setFormData({
        logo_url: branding.logo_url || '',
        favicon_url: branding.favicon_url || '',
        primary_color: branding.primary_color || '#3b82f6',
        secondary_color: branding.secondary_color || '#64748b',
        background_color: branding.background_color || '#ffffff',
        custom_css: branding.custom_css || '',
        company_name: branding.company_name || '',
        support_email: branding.support_email || '',
        terms_url: branding.terms_url || '',
        privacy_url: branding.privacy_url || '',
      });
    }
  }, [branding]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await updateBranding.mutateAsync({
        applicationId,
        data: {
          logo_url: formData.logo_url || undefined,
          favicon_url: formData.favicon_url || undefined,
          primary_color: formData.primary_color || undefined,
          secondary_color: formData.secondary_color || undefined,
          background_color: formData.background_color || undefined,
          custom_css: formData.custom_css || undefined,
          company_name: formData.company_name || undefined,
          support_email: formData.support_email || undefined,
          terms_url: formData.terms_url || undefined,
          privacy_url: formData.privacy_url || undefined,
        },
      });
    } catch (error) {
      logger.error('Failed to update branding:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <BrandingVisualSection
          logoUrl={formData.logo_url}
          faviconUrl={formData.favicon_url}
          onLogoUrlChange={value => setFormData(prev => ({ ...prev, logo_url: value }))}
          onFaviconUrlChange={value => setFormData(prev => ({ ...prev, favicon_url: value }))}
        />

        <BrandingColorsSection
          primaryColor={formData.primary_color}
          secondaryColor={formData.secondary_color}
          backgroundColor={formData.background_color}
          onPrimaryColorChange={value => setFormData(prev => ({ ...prev, primary_color: value }))}
          onSecondaryColorChange={value => setFormData(prev => ({ ...prev, secondary_color: value }))}
          onBackgroundColorChange={value => setFormData(prev => ({ ...prev, background_color: value }))}
        />

        <BrandingCompanySection
          companyName={formData.company_name}
          supportEmail={formData.support_email}
          termsUrl={formData.terms_url}
          privacyUrl={formData.privacy_url}
          customCss={formData.custom_css}
          onCompanyNameChange={value => setFormData(prev => ({ ...prev, company_name: value }))}
          onSupportEmailChange={value => setFormData(prev => ({ ...prev, support_email: value }))}
          onTermsUrlChange={value => setFormData(prev => ({ ...prev, terms_url: value }))}
          onPrivacyUrlChange={value => setFormData(prev => ({ ...prev, privacy_url: value }))}
          onCustomCssChange={value => setFormData(prev => ({ ...prev, custom_css: value }))}
        />
      </div>

      <div className="flex justify-end">
        <button
          type="submit"
          disabled={updateBranding.isPending}
          className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
        >
          {updateBranding.isPending ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            <Save size={18} />
          )}
          {updateBranding.isPending ? t('common.saving') : t('common.save')}
        </button>
      </div>
    </form>
  );
};

export default ApplicationBrandingTab;
