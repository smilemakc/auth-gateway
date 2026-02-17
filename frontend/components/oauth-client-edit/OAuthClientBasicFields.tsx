import React from 'react';
import { Plus, X } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { CreateOAuthClientRequest, ClientType } from '@auth-gateway/client-sdk';

interface OAuthClientBasicFieldsProps {
  formData: CreateOAuthClientRequest;
  isNew: boolean;
  newRedirectUri: string;
  onFormChange: (updater: (prev: CreateOAuthClientRequest) => CreateOAuthClientRequest) => void;
  onNewRedirectUriChange: (value: string) => void;
  onAddRedirectUri: () => void;
  onRemoveRedirectUri: (uri: string) => void;
}

const OAuthClientBasicFields: React.FC<OAuthClientBasicFieldsProps> = ({
  formData,
  isNew,
  newRedirectUri,
  onFormChange,
  onNewRedirectUriChange,
  onAddRedirectUri,
  onRemoveRedirectUri,
}) => {
  const { t } = useLanguage();

  return (
    <>
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('oauth_edit.basic_info')}</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-muted-foreground mb-1">{t('common.name')} *</label>
            <input
              type="text"
              value={formData.name}
              onChange={e => onFormChange(prev => ({ ...prev, name: e.target.value }))}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
              placeholder="My Application"
              required
            />
          </div>
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth_edit.description')}</label>
            <textarea
              value={formData.description}
              onChange={e => onFormChange(prev => ({ ...prev, description: e.target.value }))}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
              placeholder="A brief description of your application"
              rows={2}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth_edit.logo_url')}</label>
            <input
              type="url"
              value={formData.logo_url}
              onChange={e => onFormChange(prev => ({ ...prev, logo_url: e.target.value }))}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
              placeholder="https://example.com/logo.png"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth_edit.client_type')} *</label>
            <select
              value={formData.client_type}
              onChange={e => onFormChange(prev => ({ ...prev, client_type: e.target.value as ClientType }))}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
              disabled={!isNew}
            >
              <option value="confidential">{t('oauth_edit.confidential')}</option>
              <option value="public">{t('oauth_edit.public')}</option>
            </select>
            <p className="text-xs text-muted-foreground mt-1">
              {formData.client_type === 'confidential'
                ? t('oauth_edit.confidential_desc')
                : t('oauth_edit.public_desc')}
            </p>
          </div>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('oauth_edit.redirect_uris')}</h2>
        <div className="space-y-3">
          {(formData.redirect_uris || []).map((uri, index) => (
            <div key={index} className="flex items-center gap-2">
              <input
                type="text"
                value={uri}
                readOnly
                className="flex-1 px-4 py-2 bg-muted border border-border rounded-lg text-muted-foreground"
              />
              <button
                type="button"
                onClick={() => onRemoveRedirectUri(uri)}
                className="p-2 text-destructive hover:bg-destructive/10 rounded-lg"
              >
                <X size={18} />
              </button>
            </div>
          ))}
          <div className="flex items-center gap-2">
            <input
              type="url"
              value={newRedirectUri}
              onChange={e => onNewRedirectUriChange(e.target.value)}
              onKeyDown={e => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  onAddRedirectUri();
                }
              }}
              onBlur={() => {
                if (newRedirectUri.trim()) {
                  onAddRedirectUri();
                }
              }}
              placeholder="https://example.com/callback"
              className="flex-1 px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
            />
            <button
              type="button"
              onClick={onAddRedirectUri}
              className="p-2 bg-primary/10 text-primary hover:bg-primary/20 rounded-lg"
            >
              <Plus size={18} />
            </button>
          </div>
          <p className="text-xs text-muted-foreground mt-1">{t('oauth_edit.redirect_uris_hint')}</p>
        </div>
      </div>
    </>
  );
};

export default OAuthClientBasicFields;
