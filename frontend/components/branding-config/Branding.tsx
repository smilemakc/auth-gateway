
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, Palette, Image as ImageIcon, Layout, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { LoadingSpinner } from '../ui';
import { useBranding, useUpdateBranding } from '../../hooks/useSettings';
import { toast } from '../../services/toast';
import { BrandingPreview } from './BrandingPreview';

const Branding: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();

  const { data: brandingData, isLoading: loadingBranding } = useBranding();
  const updateMutation = useUpdateBranding();

  const [config, setConfig] = useState({
    company_name: 'Auth Gateway',
    logo_url: '',
    favicon_url: '',
    theme: {
      primary_color: '#2563EB',
      secondary_color: '#1E40AF',
      background_color: '#F3F4F6'
    },
    loginPageTitle: 'Sign in to your account',
    loginPageSubtitle: 'Welcome back! Please enter your details.',
    showSocialLogins: true,
  });

  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (brandingData) {
      setConfig({
        company_name: brandingData.company_name || 'Auth Gateway',
        logo_url: brandingData.logo_url || '',
        favicon_url: brandingData.favicon_url || '',
        theme: {
          primary_color: brandingData.theme?.primary_color || '#2563EB',
          secondary_color: brandingData.theme?.secondary_color || '#1E40AF',
          background_color: brandingData.theme?.background_color || '#F3F4F6'
        },
        loginPageTitle: brandingData.login_page_title || 'Sign in to your account',
        loginPageSubtitle: brandingData.login_page_subtitle || 'Welcome back! Please enter your details.',
        showSocialLogins: brandingData.show_social_logins ?? true,
      });
    }
  }, [brandingData]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;

    if (name === 'primary_color' || name === 'secondary_color' || name === 'background_color') {
      setConfig(prev => ({
        ...prev,
        theme: {
          ...prev.theme,
          [name]: value
        }
      }));
    } else {
      setConfig(prev => ({
        ...prev,
        [name]: type === 'checkbox' ? checked : value
      }));
    }
  };

  const handleSave = async () => {
    try {
      await updateMutation.mutateAsync({
        company_name: config.company_name,
        logo_url: config.logo_url || undefined,
        favicon_url: config.favicon_url || undefined,
        primary_color: config.theme.primary_color,
        secondary_color: config.theme.secondary_color,
        background_color: config.theme.background_color,
      });
      toast.success(t('common.saved'));
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      toast.error(t('common.error'));
    }
  };

  if (loadingBranding) {
    return <LoadingSpinner />;
  }

  return (
    <div className="h-[calc(100vh-6rem)] flex flex-col">
      <div className="flex items-center justify-between mb-4 flex-shrink-0">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/settings')}
            className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h1 className="text-xl font-bold text-foreground">{t('settings.branding')}</h1>
            <p className="text-xs text-muted-foreground">{t('settings.branding_desc')}</p>
          </div>
        </div>
        <button
          onClick={handleSave}
          disabled={updateMutation.isPending}
          className={`flex items-center gap-2 px-6 py-2 rounded-lg font-medium text-sm transition-colors
            ${saved ? 'bg-green-600 text-primary-foreground' : 'bg-primary text-primary-foreground hover:bg-primary-600'}`}
        >
          {updateMutation.isPending ? t('common.saving') : saved ? t('common.saved') : (
            <>
              <Save size={18} /> {t('common.save')}
            </>
          )}
        </button>
      </div>

      <div className="flex-1 flex gap-6 min-h-0 overflow-hidden">

        {/* Settings Panel */}
        <div className="w-1/3 bg-card rounded-xl shadow-sm border border-border flex flex-col overflow-y-auto">
          <div className="p-6 space-y-8">

            {/* General Info */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                <Layout size={16} /> General
              </h3>
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">{t('brand.company')}</label>
                <input
                  type="text"
                  name="company_name"
                  value={config.company_name}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-input rounded-md text-sm focus:ring-ring focus:border-ring"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">{t('brand.logo')}</label>
                <input
                  type="text"
                  name="logo_url"
                  value={config.logo_url}
                  onChange={handleChange}
                  placeholder="https://example.com/logo.png"
                  className="w-full px-3 py-2 border border-input rounded-md text-sm focus:ring-ring focus:border-ring"
                />
              </div>
            </div>

            <hr className="border-border" />

            {/* Colors */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                <Palette size={16} /> {t('brand.colors')}
              </h3>
              <div className="space-y-3">
                {([
                  ['primary_color', t('brand.primary'), config.theme.primary_color],
                  ['secondary_color', t('brand.secondary'), config.theme.secondary_color],
                  ['background_color', t('brand.bg'), config.theme.background_color],
                ] as const).map(([name, label, value]) => (
                  <div key={name}>
                    <label className="block text-xs font-medium text-muted-foreground mb-1.5">{label}</label>
                    <div className="flex items-center border border-input rounded-lg overflow-hidden bg-background">
                      <label className="relative cursor-pointer shrink-0">
                        <input
                          type="color"
                          name={name}
                          value={value}
                          onChange={handleChange}
                          className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                        />
                        <div className="w-10 h-10 border-r border-input" style={{ backgroundColor: value }} />
                      </label>
                      <input
                        type="text"
                        name={name}
                        value={value}
                        onChange={handleChange}
                        className="flex-1 px-3 py-2 text-sm font-mono bg-transparent border-0 outline-none"
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <hr className="border-border" />

            {/* Content */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                <ImageIcon size={16} /> {t('brand.content')}
              </h3>
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">{t('brand.heading')}</label>
                <input
                  type="text"
                  name="loginPageTitle"
                  value={config.loginPageTitle}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-input rounded-md text-sm"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">{t('brand.subtitle')}</label>
                <input
                  type="text"
                  name="loginPageSubtitle"
                  value={config.loginPageSubtitle}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-input rounded-md text-sm"
                />
              </div>
               <div className="pt-2">
                  <div className="flex items-center gap-3">
                    <button type="button" onClick={() => setConfig(prev => ({ ...prev, showSocialLogins: !prev.showSocialLogins }))}
                      className={`transition-colors ${config.showSocialLogins ? 'text-success' : 'text-muted-foreground'}`}>
                      {config.showSocialLogins ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                    </button>
                    <span className="text-sm text-muted-foreground">{t('brand.socials')}</span>
                  </div>
               </div>
            </div>

          </div>
        </div>

        {/* Live Preview */}
        <BrandingPreview config={config} />

      </div>
    </div>
  );
};

export default Branding;
