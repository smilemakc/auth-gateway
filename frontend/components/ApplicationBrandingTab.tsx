import React, { useState, useEffect } from 'react';
import { Loader2, Save, Palette, Image, Mail, FileText } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useApplicationBranding, useUpdateApplicationBranding } from '../hooks/useApplications';

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
      console.error('Failed to update branding:', error);
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
        {/* Visual Branding */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <Image className="text-primary" size={20} />
            </div>
            <h2 className="text-lg font-semibold text-foreground">{t('brand.visual') || 'Visual Identity'}</h2>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('brand.logo') || 'Logo URL'}
              </label>
              <input
                type="url"
                value={formData.logo_url}
                onChange={e => setFormData(prev => ({ ...prev, logo_url: e.target.value }))}
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="https://example.com/logo.png"
              />
              {formData.logo_url && (
                <div className="mt-2 p-4 bg-muted rounded-lg flex items-center justify-center">
                  <img src={formData.logo_url} alt="Logo preview" className="max-h-16 object-contain" />
                </div>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('brand.favicon') || 'Favicon URL'}
              </label>
              <input
                type="url"
                value={formData.favicon_url}
                onChange={e => setFormData(prev => ({ ...prev, favicon_url: e.target.value }))}
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="https://example.com/favicon.ico"
              />
            </div>
          </div>
        </div>

        {/* Colors */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <Palette className="text-primary" size={20} />
            </div>
            <h2 className="text-lg font-semibold text-foreground">{t('brand.colors') || 'Colors'}</h2>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('brand.primary') || 'Primary Color'}
              </label>
              <div className="flex items-center gap-3">
                <input
                  type="color"
                  value={formData.primary_color}
                  onChange={e => setFormData(prev => ({ ...prev, primary_color: e.target.value }))}
                  className="w-12 h-10 rounded border border-input cursor-pointer"
                />
                <input
                  type="text"
                  value={formData.primary_color}
                  onChange={e => setFormData(prev => ({ ...prev, primary_color: e.target.value }))}
                  className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('brand.secondary') || 'Secondary Color'}
              </label>
              <div className="flex items-center gap-3">
                <input
                  type="color"
                  value={formData.secondary_color}
                  onChange={e => setFormData(prev => ({ ...prev, secondary_color: e.target.value }))}
                  className="w-12 h-10 rounded border border-input cursor-pointer"
                />
                <input
                  type="text"
                  value={formData.secondary_color}
                  onChange={e => setFormData(prev => ({ ...prev, secondary_color: e.target.value }))}
                  className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('brand.bg') || 'Background Color'}
              </label>
              <div className="flex items-center gap-3">
                <input
                  type="color"
                  value={formData.background_color}
                  onChange={e => setFormData(prev => ({ ...prev, background_color: e.target.value }))}
                  className="w-12 h-10 rounded border border-input cursor-pointer"
                />
                <input
                  type="text"
                  value={formData.background_color}
                  onChange={e => setFormData(prev => ({ ...prev, background_color: e.target.value }))}
                  className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
            </div>

            {/* Color Preview */}
            <div className="mt-4 p-4 rounded-lg border border-border" style={{ backgroundColor: formData.background_color }}>
              <div className="text-center">
                <div
                  className="inline-block px-4 py-2 rounded-lg text-white font-medium"
                  style={{ backgroundColor: formData.primary_color }}
                >
                  {t('brand.preview_btn') || 'Primary Button'}
                </div>
                <p className="mt-2 text-sm" style={{ color: formData.secondary_color }}>
                  {t('brand.preview_text') || 'Secondary text color'}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Company Information */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <Mail className="text-primary" size={20} />
            </div>
            <h2 className="text-lg font-semibold text-foreground">{t('brand.company_info') || 'Company Information'}</h2>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('brand.company') || 'Company Name'}
              </label>
              <input
                type="text"
                value={formData.company_name}
                onChange={e => setFormData(prev => ({ ...prev, company_name: e.target.value }))}
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="My Company"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('brand.support_email') || 'Support Email'}
              </label>
              <input
                type="email"
                value={formData.support_email}
                onChange={e => setFormData(prev => ({ ...prev, support_email: e.target.value }))}
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="support@example.com"
              />
            </div>
          </div>
        </div>

        {/* Legal Links */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <FileText className="text-primary" size={20} />
            </div>
            <h2 className="text-lg font-semibold text-foreground">{t('brand.legal') || 'Legal Links'}</h2>
          </div>

          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('brand.terms') || 'Terms of Service URL'}
              </label>
              <input
                type="url"
                value={formData.terms_url}
                onChange={e => setFormData(prev => ({ ...prev, terms_url: e.target.value }))}
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="https://example.com/terms"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                {t('brand.privacy') || 'Privacy Policy URL'}
              </label>
              <input
                type="url"
                value={formData.privacy_url}
                onChange={e => setFormData(prev => ({ ...prev, privacy_url: e.target.value }))}
                className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                placeholder="https://example.com/privacy"
              />
            </div>
          </div>
        </div>

        {/* Custom CSS */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6 lg:col-span-2">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('brand.custom_css') || 'Custom CSS'}</h2>
          <p className="text-sm text-muted-foreground mb-4">
            {t('brand.custom_css_hint') || 'Add custom CSS to override default styles on login pages.'}
          </p>
          <textarea
            value={formData.custom_css}
            onChange={e => setFormData(prev => ({ ...prev, custom_css: e.target.value }))}
            rows={6}
            className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder=".login-form { /* custom styles */ }"
          />
        </div>
      </div>

      {/* Save Button */}
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
          {updateBranding.isPending ? (t('common.saving') || 'Saving...') : (t('common.save') || 'Save Branding')}
        </button>
      </div>
    </form>
  );
};

export default ApplicationBrandingTab;
