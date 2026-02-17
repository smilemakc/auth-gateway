import React from 'react';
import { ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { CreateOAuthClientRequest, GrantType } from '@auth-gateway/client-sdk';

const GRANT_TYPES: { value: GrantType; labelKey: string; descKey: string }[] = [
  { value: 'authorization_code', labelKey: 'oauth_edit.grant_auth_code', descKey: 'oauth_edit.grant_auth_code_desc' },
  { value: 'client_credentials', labelKey: 'oauth_edit.grant_client_creds', descKey: 'oauth_edit.grant_client_creds_desc' },
  { value: 'refresh_token', labelKey: 'oauth_edit.grant_refresh', descKey: 'oauth_edit.grant_refresh_desc' },
  { value: 'urn:ietf:params:oauth:grant-type:device_code', labelKey: 'oauth_edit.grant_device', descKey: 'oauth_edit.grant_device_desc' },
];

interface OAuthClientScopeSelectorProps {
  formData: CreateOAuthClientRequest;
  availableScopes: string[];
  onFormChange: (updater: (prev: CreateOAuthClientRequest) => CreateOAuthClientRequest) => void;
}

const OAuthClientScopeSelector: React.FC<OAuthClientScopeSelectorProps> = ({
  formData,
  availableScopes,
  onFormChange,
}) => {
  const { t } = useLanguage();

  const handleGrantTypeToggle = (grantType: GrantType) => {
    onFormChange(prev => ({
      ...prev,
      allowed_grant_types: prev.allowed_grant_types?.includes(grantType)
        ? prev.allowed_grant_types.filter(g => g !== grantType)
        : [...(prev.allowed_grant_types || []), grantType],
    }));
  };

  const handleScopeToggle = (scope: string, isDefault: boolean = false) => {
    if (isDefault) {
      onFormChange(prev => ({
        ...prev,
        default_scopes: prev.default_scopes?.includes(scope)
          ? prev.default_scopes.filter(s => s !== scope)
          : [...(prev.default_scopes || []), scope],
      }));
    } else {
      onFormChange(prev => {
        const newAllowedScopes = prev.allowed_scopes?.includes(scope)
          ? prev.allowed_scopes.filter(s => s !== scope)
          : [...(prev.allowed_scopes || []), scope];
        const newDefaultScopes = prev.default_scopes?.filter(s => newAllowedScopes.includes(s)) || [];
        return {
          ...prev,
          allowed_scopes: newAllowedScopes,
          default_scopes: newDefaultScopes,
        };
      });
    }
  };

  return (
    <>
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('oauth_edit.grant_types')}</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {GRANT_TYPES.map(grant => (
            <label
              key={grant.value}
              className={`flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors ${
                formData.allowed_grant_types?.includes(grant.value)
                  ? 'border-primary bg-primary/10'
                  : 'border-border hover:border-input'
              }`}
            >
              <input
                type="checkbox"
                checked={formData.allowed_grant_types?.includes(grant.value)}
                onChange={() => handleGrantTypeToggle(grant.value)}
                className="mt-1"
              />
              <div>
                <div className="font-medium text-foreground">{t(grant.labelKey)}</div>
                <div className="text-sm text-muted-foreground">{t(grant.descKey)}</div>
              </div>
            </label>
          ))}
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('oauth_edit.scopes')}</h2>
        <div className="space-y-4">
          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-2">{t('oauth_edit.allowed_scopes')}</h3>
            <div className="flex flex-wrap gap-2">
              {availableScopes.map(scope => (
                <button
                  key={scope}
                  type="button"
                  onClick={() => handleScopeToggle(scope)}
                  className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                    formData.allowed_scopes?.includes(scope)
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted text-muted-foreground hover:bg-accent'
                  }`}
                >
                  {scope}
                </button>
              ))}
            </div>
          </div>
          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-2">{t('oauth_edit.default_scopes')}</h3>
            <p className="text-xs text-muted-foreground mb-2">{t('oauth_edit.default_scopes_desc')}</p>
            <div className="flex flex-wrap gap-2">
              {formData.allowed_scopes?.map(scope => (
                <button
                  key={scope}
                  type="button"
                  onClick={() => handleScopeToggle(scope, true)}
                  className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                    formData.default_scopes?.includes(scope)
                      ? 'bg-success text-primary-foreground'
                      : 'bg-muted text-muted-foreground hover:bg-accent'
                  }`}
                >
                  {scope}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('oauth_edit.token_settings')}</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth_edit.access_token_ttl')}</label>
            <input
              type="number"
              value={formData.access_token_ttl}
              onChange={e => onFormChange(prev => ({ ...prev, access_token_ttl: parseInt(e.target.value) }))}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
              min={60}
              max={86400}
            />
            <p className="text-xs text-muted-foreground mt-1">{Math.floor((formData.access_token_ttl || 900) / 60)} {t('oauth_edit.minutes')}</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth_edit.refresh_token_ttl')}</label>
            <input
              type="number"
              value={formData.refresh_token_ttl}
              onChange={e => onFormChange(prev => ({ ...prev, refresh_token_ttl: parseInt(e.target.value) }))}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
              min={3600}
              max={2592000}
            />
            <p className="text-xs text-muted-foreground mt-1">{Math.floor((formData.refresh_token_ttl || 604800) / 86400)} {t('oauth_edit.days')}</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth_edit.id_token_ttl')}</label>
            <input
              type="number"
              value={formData.id_token_ttl}
              onChange={e => onFormChange(prev => ({ ...prev, id_token_ttl: parseInt(e.target.value) }))}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
              min={300}
              max={86400}
            />
            <p className="text-xs text-muted-foreground mt-1">{Math.floor((formData.id_token_ttl || 3600) / 60)} {t('oauth_edit.minutes')}</p>
          </div>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('oauth_edit.security_settings')}</h2>
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => onFormChange(prev => ({ ...prev, require_pkce: !prev.require_pkce }))}
              className={`transition-colors ${formData.require_pkce ? 'text-success' : 'text-muted-foreground'}`}
            >
              {formData.require_pkce ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
            </button>
            <div>
              <div className="font-medium text-foreground">{t('oauth_edit.require_pkce')}</div>
              <div className="text-sm text-muted-foreground">{t('oauth_edit.require_pkce_desc')}</div>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => onFormChange(prev => ({ ...prev, require_consent: !prev.require_consent }))}
              className={`transition-colors ${formData.require_consent ? 'text-success' : 'text-muted-foreground'}`}
            >
              {formData.require_consent ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
            </button>
            <div>
              <div className="font-medium text-foreground">{t('oauth_edit.require_consent')}</div>
              <div className="text-sm text-muted-foreground">{t('oauth_edit.require_consent_desc')}</div>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => onFormChange(prev => ({ ...prev, first_party: !prev.first_party }))}
              className={`transition-colors ${formData.first_party ? 'text-success' : 'text-muted-foreground'}`}
            >
              {formData.first_party ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
            </button>
            <div>
              <div className="font-medium text-foreground">{t('oauth_edit.first_party')}</div>
              <div className="text-sm text-muted-foreground">{t('oauth_edit.first_party_desc')}</div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default OAuthClientScopeSelector;
