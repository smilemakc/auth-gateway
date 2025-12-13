
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getBranding, updateBranding } from '../services/mockData';
import { BrandingConfig } from '../types';
import { ArrowLeft, Save, Palette, Image as ImageIcon, Layout, Lock, Github, Mail } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const Branding: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [config, setConfig] = useState<BrandingConfig>({
    companyName: 'Auth Gateway',
    logoUrl: '',
    faviconUrl: '',
    primaryColor: '#2563EB',
    accentColor: '#1E40AF',
    backgroundColor: '#F3F4F6',
    loginPageTitle: 'Sign in to your account',
    loginPageSubtitle: 'Welcome back! Please enter your details.',
    showSocialLogins: true
  });
  
  const [loading, setLoading] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    const data = getBranding();
    if (data) setConfig(data);
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setConfig(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
  };

  const handleSave = () => {
    setLoading(true);
    setTimeout(() => {
      updateBranding(config);
      setLoading(false);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    }, 800);
  };

  return (
    <div className="h-[calc(100vh-6rem)] flex flex-col">
      <div className="flex items-center justify-between mb-4 flex-shrink-0">
        <div className="flex items-center gap-4">
          <button 
            onClick={() => navigate('/settings')}
            className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h1 className="text-xl font-bold text-gray-900">{t('settings.branding')}</h1>
            <p className="text-xs text-gray-500">{t('settings.branding_desc')}</p>
          </div>
        </div>
        <button
          onClick={handleSave}
          disabled={loading}
          className={`flex items-center gap-2 px-6 py-2 rounded-lg font-medium text-sm transition-colors
            ${saved ? 'bg-green-600 text-white' : 'bg-blue-600 text-white hover:bg-blue-700'}`}
        >
          {loading ? t('common.saving') : saved ? t('common.saved') : (
            <>
              <Save size={18} /> {t('common.save')}
            </>
          )}
        </button>
      </div>

      <div className="flex-1 flex gap-6 min-h-0 overflow-hidden">
        
        {/* Settings Panel */}
        <div className="w-1/3 bg-white rounded-xl shadow-sm border border-gray-100 flex flex-col overflow-y-auto">
          <div className="p-6 space-y-8">
            
            {/* General Info */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
                <Layout size={16} /> General
              </h3>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">{t('brand.company')}</label>
                <input
                  type="text"
                  name="companyName"
                  value={config.companyName}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">{t('brand.logo')}</label>
                <input
                  type="text"
                  name="logoUrl"
                  value={config.logoUrl}
                  onChange={handleChange}
                  placeholder="https://example.com/logo.png"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>

            <hr className="border-gray-100" />

            {/* Colors */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
                <Palette size={16} /> {t('brand.colors')}
              </h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">{t('brand.primary')}</label>
                  <div className="flex gap-2">
                    <input
                      type="color"
                      name="primaryColor"
                      value={config.primaryColor}
                      onChange={handleChange}
                      className="h-9 w-9 p-1 rounded border border-gray-300 cursor-pointer"
                    />
                    <input
                      type="text"
                      name="primaryColor"
                      value={config.primaryColor}
                      onChange={handleChange}
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm font-mono"
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">{t('brand.bg')}</label>
                  <div className="flex gap-2">
                    <input
                      type="color"
                      name="backgroundColor"
                      value={config.backgroundColor}
                      onChange={handleChange}
                      className="h-9 w-9 p-1 rounded border border-gray-300 cursor-pointer"
                    />
                    <input
                      type="text"
                      name="backgroundColor"
                      value={config.backgroundColor}
                      onChange={handleChange}
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-md text-sm font-mono"
                    />
                  </div>
                </div>
              </div>
            </div>

            <hr className="border-gray-100" />

            {/* Content */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
                <ImageIcon size={16} /> {t('brand.content')}
              </h3>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">{t('brand.heading')}</label>
                <input
                  type="text"
                  name="loginPageTitle"
                  value={config.loginPageTitle}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">{t('brand.subtitle')}</label>
                <input
                  type="text"
                  name="loginPageSubtitle"
                  value={config.loginPageSubtitle}
                  onChange={handleChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                />
              </div>
               <div className="pt-2">
                  <label className="flex items-center text-sm text-gray-700">
                    <input
                      type="checkbox"
                      name="showSocialLogins"
                      checked={config.showSocialLogins}
                      onChange={handleChange}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500 mr-2"
                    />
                    {t('brand.socials')}
                  </label>
               </div>
            </div>

          </div>
        </div>

        {/* Live Preview */}
        <div className="flex-1 bg-gray-50 rounded-xl shadow-inner border border-gray-200 overflow-hidden relative flex flex-col">
          <div className="bg-white border-b border-gray-200 px-4 py-2 flex items-center justify-between text-xs text-gray-500">
             <span>{t('brand.preview')}</span>
             <span className="flex items-center gap-1"><Lock size={10}/> auth.example.com/login</span>
          </div>
          
          <div 
             className="flex-1 flex items-center justify-center p-4 relative"
             style={{ backgroundColor: config.backgroundColor }}
          >
             {/* Mock Login Card */}
             <div className="w-full max-w-md bg-white rounded-2xl shadow-xl overflow-hidden">
                <div className="p-8">
                   <div className="text-center mb-8">
                      {config.logoUrl ? (
                         <img src={config.logoUrl} alt="Logo" className="h-12 mx-auto mb-4 object-contain" />
                      ) : (
                         <div 
                            className="w-12 h-12 rounded-lg flex items-center justify-center mx-auto mb-4 text-white font-bold text-xl"
                            style={{ backgroundColor: config.primaryColor }}
                         >
                            {config.companyName.charAt(0)}
                         </div>
                      )}
                      <h2 className="text-2xl font-bold text-gray-900">{config.loginPageTitle}</h2>
                      <p className="text-gray-500 mt-2 text-sm">{config.loginPageSubtitle}</p>
                   </div>

                   <div className="space-y-4">
                      <div>
                         <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                         <input disabled type="email" className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50" placeholder="user@example.com" />
                      </div>
                      <div>
                         <label className="block text-sm font-medium text-gray-700 mb-1">Password</label>
                         <input disabled type="password" className="w-full px-4 py-2 border border-gray-300 rounded-lg bg-gray-50" placeholder="••••••••" />
                      </div>
                      
                      <button 
                         className="w-full py-2 px-4 text-white font-medium rounded-lg shadow-sm"
                         style={{ backgroundColor: config.primaryColor }}
                      >
                         Sign In
                      </button>

                      {config.showSocialLogins && (
                         <>
                            <div className="relative my-6">
                               <div className="absolute inset-0 flex items-center">
                                  <div className="w-full border-t border-gray-200"></div>
                               </div>
                               <div className="relative flex justify-center text-sm">
                                  <span className="px-2 bg-white text-gray-500">Or continue with</span>
                               </div>
                            </div>
                            <div className="grid grid-cols-2 gap-3">
                               <button className="flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 bg-white text-sm font-medium text-gray-700">
                                  <Github size={16} className="mr-2" /> GitHub
                               </button>
                               <button className="flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 bg-white text-sm font-medium text-gray-700">
                                  <span className="font-bold text-red-500 mr-2">G</span> Google
                               </button>
                            </div>
                         </>
                      )}
                   </div>
                </div>
                <div className="bg-gray-50 px-8 py-4 text-center text-sm text-gray-500 border-t border-gray-100">
                   Don't have an account? <span style={{ color: config.primaryColor }} className="font-medium">Sign up</span>
                </div>
             </div>
             
             {/* Footer Mock */}
             <div className="absolute bottom-4 text-xs text-gray-400">
                &copy; {new Date().getFullYear()} {config.companyName}. All rights reserved.
             </div>
          </div>
        </div>

      </div>
    </div>
  );
};

export default Branding;
