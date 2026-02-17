import React, { useState } from 'react';
import { Eye, EyeOff } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface OAuthProviderCredentialsSectionProps {
  clientId: string;
  clientSecret: string;
  callbackUrl: string;
  scopes: string;
  isEditMode: boolean;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const OAuthProviderCredentialsSection: React.FC<OAuthProviderCredentialsSectionProps> = ({
  clientId,
  clientSecret,
  callbackUrl,
  scopes,
  isEditMode,
  onChange,
}) => {
  const { t } = useLanguage();
  const [showSecret, setShowSecret] = useState(false);

  return (
    <>
      <div className="grid grid-cols-1 gap-6">
        <div>
          <label htmlFor="client_id" className="block text-sm font-medium text-muted-foreground mb-1">{t('app_oauth.client_id')}</label>
          <input
            type="text"
            id="client_id"
            name="client_id"
            value={clientId}
            onChange={onChange}
            required
            className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all font-mono text-sm"
            placeholder="e.g. 1234567890-abc..."
          />
        </div>
        <div>
          <label htmlFor="client_secret" className="block text-sm font-medium text-muted-foreground mb-1">{t('app_oauth.client_secret')}</label>
          <div className="relative">
            <input
              type={showSecret ? "text" : "password"}
              id="client_secret"
              name="client_secret"
              value={clientSecret}
              onChange={onChange}
              required={!isEditMode}
              className="w-full pl-4 pr-12 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all font-mono text-sm"
              placeholder={isEditMode ? t('app_oauth.secret_unchanged') : 'e.g. GOCSPX-...'}
            />
            <button
              type="button"
              onClick={() => setShowSecret(!showSecret)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              {showSecret ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
        </div>
      </div>

      <div>
        <label htmlFor="callback_url" className="block text-sm font-medium text-muted-foreground mb-1">{t('app_oauth.callback_url')}</label>
        <input
          type="url"
          id="callback_url"
          name="callback_url"
          value={callbackUrl}
          onChange={onChange}
          required
          className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none font-mono text-sm"
          placeholder="https://your-app.com/auth/callback"
        />
      </div>

      <div>
        <label htmlFor="scopes" className="block text-sm font-medium text-muted-foreground mb-1">
          {t('app_oauth.scopes')} <span className="text-xs font-normal">{t('app_oauth.scopes_hint')}</span>
        </label>
        <input
          type="text"
          id="scopes"
          name="scopes"
          value={scopes}
          onChange={onChange}
          className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none text-sm"
          placeholder="e.g. email, profile, openid"
        />
      </div>
    </>
  );
};

export default OAuthProviderCredentialsSection;
