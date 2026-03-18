import React from 'react';
import { Lock, Github } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface BrandingConfig {
  company_name: string;
  logo_url: string;
  theme: {
    primary_color: string;
    secondary_color: string;
    background_color: string;
  };
  loginPageTitle: string;
  loginPageSubtitle: string;
  showSocialLogins: boolean;
}

interface BrandingPreviewProps {
  config: BrandingConfig;
}

export const BrandingPreview: React.FC<BrandingPreviewProps> = ({ config }) => {
  const { t } = useLanguage();

  return (
    <div className="flex-1 bg-muted rounded-xl shadow-inner border border-border overflow-hidden relative flex flex-col">
      <div className="bg-card border-b border-border px-4 py-2 flex items-center justify-between text-xs text-muted-foreground">
         <span>{t('brand.preview')}</span>
         <span className="flex items-center gap-1"><Lock size={10}/> auth.example.com/login</span>
      </div>

      <div
         className="flex-1 flex items-center justify-center p-4 relative"
         style={{ backgroundColor: config.theme.background_color }}
      >
         <div className="w-full max-w-md bg-card rounded-2xl shadow-xl overflow-hidden">
            <div className="p-8">
               <div className="text-center mb-8">
                  {config.logo_url ? (
                     <img src={config.logo_url} alt="Logo" className="h-12 mx-auto mb-4 object-contain" />
                  ) : (
                     <div
                        className="w-12 h-12 rounded-lg flex items-center justify-center mx-auto mb-4 text-primary-foreground font-bold text-xl"
                        style={{ backgroundColor: config.theme.primary_color }}
                     >
                        {config.company_name.charAt(0)}
                     </div>
                  )}
                  <h2 className="text-2xl font-bold text-foreground">{config.loginPageTitle}</h2>
                  <p className="text-muted-foreground mt-2 text-sm">{config.loginPageSubtitle}</p>
               </div>

               <div className="space-y-4">
                  <div>
                     <label className="block text-sm font-medium text-muted-foreground mb-1">Email</label>
                     <input disabled type="email" className="w-full px-4 py-2 border border-input rounded-lg bg-muted" placeholder="user@example.com" />
                  </div>
                  <div>
                     <label className="block text-sm font-medium text-muted-foreground mb-1">Password</label>
                     <input disabled type="password" className="w-full px-4 py-2 border border-input rounded-lg bg-muted" placeholder="••••••••" />
                  </div>
                  <div className="flex justify-end">
                     <span className="text-xs font-medium cursor-pointer" style={{ color: config.theme.secondary_color }}>Forgot password?</span>
                  </div>

                  <button
                     className="w-full py-2.5 px-4 text-white font-medium rounded-lg shadow-sm transition-colors"
                     style={{ backgroundColor: config.theme.primary_color }}
                  >
                     Sign In
                  </button>

                  {config.showSocialLogins && (
                     <>
                        <div className="relative my-6">
                           <div className="absolute inset-0 flex items-center">
                              <div className="w-full border-t border-border"></div>
                           </div>
                           <div className="relative flex justify-center text-sm">
                              <span className="px-2 bg-card text-muted-foreground">Or continue with</span>
                           </div>
                        </div>
                        <div className="grid grid-cols-2 gap-3">
                           <button
                              className="flex items-center justify-center px-4 py-2 rounded-lg bg-card text-sm font-medium text-muted-foreground"
                              style={{ border: `1px solid ${config.theme.secondary_color}30` }}
                           >
                              <Github size={16} className="mr-2" /> GitHub
                           </button>
                           <button
                              className="flex items-center justify-center px-4 py-2 rounded-lg bg-card text-sm font-medium text-muted-foreground"
                              style={{ border: `1px solid ${config.theme.secondary_color}30` }}
                           >
                              <span className="font-bold text-destructive mr-2">G</span> Google
                           </button>
                        </div>
                     </>
                  )}
               </div>
            </div>
            <div className="px-8 py-4 text-center text-sm text-muted-foreground" style={{ borderTop: `1px solid ${config.theme.secondary_color}20`, backgroundColor: `${config.theme.secondary_color}08` }}>
               Don't have an account? <span style={{ color: config.theme.secondary_color }} className="font-medium cursor-pointer">Sign up</span>
            </div>
         </div>

         <div className="absolute bottom-4 text-xs text-muted-foreground">
            &copy; {new Date().getFullYear()} {config.company_name}. All rights reserved.
         </div>
      </div>
    </div>
  );
};
