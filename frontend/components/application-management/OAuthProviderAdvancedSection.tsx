import React from 'react';
import { ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface OAuthProviderAdvancedSectionProps {
  authUrl: string;
  tokenUrl: string;
  userInfoUrl: string;
  isActive: boolean;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onToggleActive: () => void;
}

const OAuthProviderAdvancedSection: React.FC<OAuthProviderAdvancedSectionProps> = ({
  authUrl,
  tokenUrl,
  userInfoUrl,
  isActive,
  onChange,
  onToggleActive,
}) => {
  const { t } = useLanguage();

  return (
    <>
      <div className="space-y-4 pt-4 border-t border-border">
        <h3 className="text-sm font-semibold text-foreground">{t('app_oauth.advanced')}</h3>
        <div className="grid grid-cols-1 gap-4">
          <div>
            <label htmlFor="auth_url" className="block text-sm font-medium text-muted-foreground mb-1">
              {t('app_oauth.auth_url')} <span className="text-xs font-normal">{t('app_oauth.optional')}</span>
            </label>
            <input
              type="url"
              id="auth_url"
              name="auth_url"
              value={authUrl}
              onChange={onChange}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none font-mono text-sm"
              placeholder="https://accounts.provider.com/oauth/authorize"
            />
          </div>
          <div>
            <label htmlFor="token_url" className="block text-sm font-medium text-muted-foreground mb-1">
              {t('app_oauth.token_url')} <span className="text-xs font-normal">{t('app_oauth.optional')}</span>
            </label>
            <input
              type="url"
              id="token_url"
              name="token_url"
              value={tokenUrl}
              onChange={onChange}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none font-mono text-sm"
              placeholder="https://oauth.provider.com/token"
            />
          </div>
          <div>
            <label htmlFor="user_info_url" className="block text-sm font-medium text-muted-foreground mb-1">
              {t('app_oauth.user_info_url')} <span className="text-xs font-normal">{t('app_oauth.optional')}</span>
            </label>
            <input
              type="url"
              id="user_info_url"
              name="user_info_url"
              value={userInfoUrl}
              onChange={onChange}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none font-mono text-sm"
              placeholder="https://api.provider.com/userinfo"
            />
          </div>
        </div>
      </div>

      <div className="pt-6 border-t border-border">
         <div className="flex items-center gap-3">
           <button
             type="button"
             onClick={onToggleActive}
             className={`transition-colors ${isActive ? 'text-success' : 'text-muted-foreground'}`}
           >
             {isActive ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
           </button>
           <div>
             <span className="font-medium text-foreground block">{t('app_oauth.enable_provider')}</span>
             <p className="text-xs text-muted-foreground">{t('app_oauth.enable_provider_desc')}</p>
           </div>
         </div>
      </div>
    </>
  );
};

export default OAuthProviderAdvancedSection;
