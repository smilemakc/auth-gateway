import React from 'react';
import { Eye, EyeOff, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface OAuthProviderFormData {
  provider: string;
  client_id: string;
  client_secret: string;
  callback_url: string;
  scopes: string;
  is_active: boolean;
}

interface OAuthProviderFormFieldsProps {
  formData: OAuthProviderFormData;
  isEditMode: boolean;
  isNewMode: boolean;
  showSecret: boolean;
  onToggleShowSecret: () => void;
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => void;
  onToggleActive: () => void;
}

export const OAuthProviderFormFields: React.FC<OAuthProviderFormFieldsProps> = ({
  formData,
  isEditMode,
  isNewMode,
  showSecret,
  onToggleShowSecret,
  onChange,
  onToggleActive,
}) => {
  const { t } = useLanguage();

  return (
    <div className="p-6 space-y-8">
      {/* Provider Selection */}
      <div>
        <label className="block text-sm font-medium text-muted-foreground mb-2">Provider</label>
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-3">
          {['google', 'github', 'yandex', 'telegram', 'instagram', 'onec'].map(p => (
            <label key={p} className={`
              cursor-pointer border rounded-lg p-3 text-center transition-all hover:bg-accent
              ${formData.provider === p ? 'ring-2 ring-ring border-transparent bg-primary/10' : 'border-border'}
            `}>
              <input
                type="radio"
                name="provider"
                value={p}
                checked={formData.provider === p}
                onChange={onChange}
                className="sr-only"
                disabled={isEditMode}
              />
              <span className="capitalize font-medium block text-foreground">{p}</span>
            </label>
          ))}
        </div>
      </div>

      {/* Credentials */}
      <div className="grid grid-cols-1 gap-6">
        <div>
          <label htmlFor="client_id" className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth.client_id')}</label>
          <input
            type="text"
            id="client_id"
            name="client_id"
            value={formData.client_id}
            onChange={onChange}
            required
            className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all font-mono text-sm"
            placeholder="e.g. 1234567890-abc..."
          />
        </div>
        <div>
          <label htmlFor="client_secret" className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth.client_secret')}</label>
          <div className="relative">
            <input
              type={showSecret ? "text" : "password"}
              id="client_secret"
              name="client_secret"
              value={formData.client_secret}
              onChange={onChange}
              required={isNewMode}
              className="w-full pl-4 pr-12 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all font-mono text-sm"
              placeholder={isEditMode ? '(unchanged)' : 'e.g. GOCSPX-...'}
            />
            <button
              type="button"
              onClick={onToggleShowSecret}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              {showSecret ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
        </div>
      </div>

      {/* Callback URL */}
      <div>
        <label htmlFor="callback_url" className="block text-sm font-medium text-muted-foreground mb-1">Callback URL</label>
        <input
          type="url"
          id="callback_url"
          name="callback_url"
          value={formData.callback_url}
          onChange={onChange}
          required
          className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none font-mono text-sm"
          placeholder="https://your-app.com/auth/callback"
        />
      </div>

      {/* Scopes */}
      <div>
        <label htmlFor="scopes" className="block text-sm font-medium text-muted-foreground mb-1">Scopes</label>
        <input
          type="text"
          id="scopes"
          name="scopes"
          value={formData.scopes}
          onChange={onChange}
          className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none font-mono text-sm"
          placeholder="openid, email, profile"
        />
        <p className="text-xs text-muted-foreground mt-1">Comma-separated list of OAuth scopes</p>
      </div>

      {/* Status */}
      <div className="pt-6 border-t border-border">
         <div className="flex items-center gap-3">
           <button
             type="button"
             onClick={onToggleActive}
             className={`transition-colors ${formData.is_active ? 'text-success' : 'text-muted-foreground'}`}
           >
             {formData.is_active ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
           </button>
           <div>
             <span className="font-medium text-foreground block">{t('oauth.enable')}</span>
           </div>
         </div>
      </div>
    </div>
  );
};
