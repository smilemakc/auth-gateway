import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, Plus, X, Copy, Check, Eye, EyeOff, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import {
  useOAuthClientDetail,
  useCreateOAuthClient,
  useUpdateOAuthClient,
  useOAuthScopes,
} from '../hooks/useOAuthClients';
import type {
  CreateOAuthClientRequest,
  UpdateOAuthClientRequest,
  GrantType,
  ClientType,
} from '@auth-gateway/client-sdk';

const GRANT_TYPES: { value: GrantType; labelKey: string; descKey: string }[] = [
  { value: 'authorization_code', labelKey: 'oauth_edit.grant_auth_code', descKey: 'oauth_edit.grant_auth_code_desc' },
  { value: 'client_credentials', labelKey: 'oauth_edit.grant_client_creds', descKey: 'oauth_edit.grant_client_creds_desc' },
  { value: 'refresh_token', labelKey: 'oauth_edit.grant_refresh', descKey: 'oauth_edit.grant_refresh_desc' },
  { value: 'urn:ietf:params:oauth:grant-type:device_code', labelKey: 'oauth_edit.grant_device', descKey: 'oauth_edit.grant_device_desc' },
];

const STANDARD_SCOPES = ['openid', 'profile', 'email', 'address', 'phone', 'offline_access'];

const OAuthClientEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isNew = !id || id === 'new';

  const { data: existingClient, isLoading: loadingClient } = useOAuthClientDetail(id || '');
  const { data: scopesData } = useOAuthScopes();
  const createClient = useCreateOAuthClient();
  const updateClient = useUpdateOAuthClient();

  const [showSecret, setShowSecret] = useState(false);
  const [copiedSecret, setCopiedSecret] = useState(false);
  const [newClientSecret, setNewClientSecret] = useState<string | null>(null);
  const [newRedirectUri, setNewRedirectUri] = useState('');

  const [formData, setFormData] = useState<CreateOAuthClientRequest>({
    name: '',
    description: '',
    logo_url: '',
    client_type: 'confidential',
    redirect_uris: [],
    allowed_grant_types: ['authorization_code', 'refresh_token'],
    allowed_scopes: ['openid', 'profile', 'email'],
    default_scopes: ['openid', 'profile'],
    access_token_ttl: 900,
    refresh_token_ttl: 604800,
    id_token_ttl: 3600,
    require_pkce: true,
    require_consent: true,
    first_party: false,
  });

  useEffect(() => {
    if (existingClient && !isNew) {
      setFormData({
        name: existingClient.name,
        description: existingClient.description || '',
        logo_url: existingClient.logo_url || '',
        client_type: existingClient.client_type,
        redirect_uris: existingClient.redirect_uris || [],
        allowed_grant_types: existingClient.allowed_grant_types || [],
        allowed_scopes: existingClient.allowed_scopes || [],
        default_scopes: existingClient.default_scopes || [],
        access_token_ttl: existingClient.access_token_ttl,
        refresh_token_ttl: existingClient.refresh_token_ttl,
        id_token_ttl: existingClient.id_token_ttl,
        require_pkce: existingClient.require_pkce,
        require_consent: existingClient.require_consent,
        first_party: existingClient.first_party,
      });
    }
  }, [existingClient, isNew]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      if (isNew) {
        const result = await createClient.mutateAsync(formData);
        setNewClientSecret(result.client_secret);
      } else {
        const updateData: UpdateOAuthClientRequest = {
          name: formData.name,
          description: formData.description,
          logo_url: formData.logo_url,
          redirect_uris: formData.redirect_uris,
          allowed_grant_types: formData.allowed_grant_types,
          allowed_scopes: formData.allowed_scopes,
          default_scopes: formData.default_scopes,
          access_token_ttl: formData.access_token_ttl,
          refresh_token_ttl: formData.refresh_token_ttl,
          id_token_ttl: formData.id_token_ttl,
          require_pkce: formData.require_pkce,
          require_consent: formData.require_consent,
          first_party: formData.first_party,
        };
        await updateClient.mutateAsync({ id: id!, data: updateData });
        navigate('/oauth-clients');
      }
    } catch (error) {
      console.error('Failed to save client:', error);
    }
  };

  const handleAddRedirectUri = () => {
    if (newRedirectUri && !(formData.redirect_uris || []).includes(newRedirectUri)) {
      setFormData(prev => ({
        ...prev,
        redirect_uris: [...(prev.redirect_uris || []), newRedirectUri],
      }));
      setNewRedirectUri('');
    }
  };

  const handleRemoveRedirectUri = (uri: string) => {
    setFormData(prev => ({
      ...prev,
      redirect_uris: (prev.redirect_uris || []).filter(u => u !== uri),
    }));
  };

  const handleGrantTypeToggle = (grantType: GrantType) => {
    setFormData(prev => ({
      ...prev,
      allowed_grant_types: prev.allowed_grant_types?.includes(grantType)
        ? prev.allowed_grant_types.filter(g => g !== grantType)
        : [...(prev.allowed_grant_types || []), grantType],
    }));
  };

  const handleScopeToggle = (scope: string, isDefault: boolean = false) => {
    if (isDefault) {
      setFormData(prev => ({
        ...prev,
        default_scopes: prev.default_scopes?.includes(scope)
          ? prev.default_scopes.filter(s => s !== scope)
          : [...(prev.default_scopes || []), scope],
      }));
    } else {
      setFormData(prev => {
        const newAllowedScopes = prev.allowed_scopes?.includes(scope)
          ? prev.allowed_scopes.filter(s => s !== scope)
          : [...(prev.allowed_scopes || []), scope];
        // Remove from default scopes if removed from allowed
        const newDefaultScopes = prev.default_scopes?.filter(s => newAllowedScopes.includes(s)) || [];
        return {
          ...prev,
          allowed_scopes: newAllowedScopes,
          default_scopes: newDefaultScopes,
        };
      });
    }
  };

  const copySecret = () => {
    if (newClientSecret) {
      navigator.clipboard.writeText(newClientSecret);
      setCopiedSecret(true);
      setTimeout(() => setCopiedSecret(false), 2000);
    }
  };

  if (!isNew && loadingClient) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  // Show success modal with client secret after creation
  if (newClientSecret) {
    return (
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="bg-success/10 border border-success rounded-xl p-6">
          <h2 className="text-xl font-bold text-success mb-2">{t('oauth_edit.success_title')}</h2>
          <p className="text-success mb-4">
            {t('oauth_edit.success_desc')}
          </p>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-success mb-1">{t('oauth_edit.client_secret')}</label>
              <div className="flex items-center gap-2">
                <code className="flex-1 bg-card rounded-lg px-4 py-3 text-sm font-mono border border-success break-all">
                  {showSecret ? newClientSecret : '••••••••••••••••••••••••••••••••'}
                </code>
                <button
                  onClick={() => setShowSecret(!showSecret)}
                  className="p-2 text-success hover:bg-success/10 rounded-lg"
                >
                  {showSecret ? <EyeOff size={20} /> : <Eye size={20} />}
                </button>
                <button
                  onClick={copySecret}
                  className="p-2 text-success hover:bg-success/10 rounded-lg"
                >
                  {copiedSecret ? <Check size={20} /> : <Copy size={20} />}
                </button>
              </div>
            </div>
          </div>
          <div className="mt-6 flex justify-end">
            <button
              onClick={() => navigate('/oauth-clients')}
              className="px-4 py-2 bg-success hover:bg-success text-primary-foreground rounded-lg font-medium"
            >
              {t('oauth_edit.done')}
            </button>
          </div>
        </div>
      </div>
    );
  }

  const availableScopes = [
    ...STANDARD_SCOPES,
    ...(scopesData?.scopes?.filter(s => !STANDARD_SCOPES.includes(s.name)).map(s => s.name) || []),
  ];

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/oauth-clients')}
          className="p-2 hover:bg-accent rounded-lg transition-colors"
        >
          <ArrowLeft size={20} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-foreground">
            {isNew ? t('oauth_edit.create_title') : t('oauth_edit.edit_title')}
          </h1>
          <p className="text-muted-foreground mt-1">
            {isNew ? t('oauth_edit.create_desc') : t('oauth_edit.edit_desc')}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Info */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('oauth_edit.basic_info')}</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-muted-foreground mb-1">{t('common.name')} *</label>
              <input
                type="text"
                value={formData.name}
                onChange={e => setFormData(prev => ({ ...prev, name: e.target.value }))}
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
                placeholder="My Application"
                required
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth_edit.description')}</label>
              <textarea
                value={formData.description}
                onChange={e => setFormData(prev => ({ ...prev, description: e.target.value }))}
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
                onChange={e => setFormData(prev => ({ ...prev, logo_url: e.target.value }))}
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
                placeholder="https://example.com/logo.png"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth_edit.client_type')} *</label>
              <select
                value={formData.client_type}
                onChange={e => setFormData(prev => ({ ...prev, client_type: e.target.value as ClientType }))}
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

        {/* Redirect URIs */}
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
                  onClick={() => handleRemoveRedirectUri(uri)}
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
                onChange={e => setNewRedirectUri(e.target.value)}
                onKeyDown={e => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    handleAddRedirectUri();
                  }
                }}
                onBlur={() => {
                  if (newRedirectUri.trim()) {
                    handleAddRedirectUri();
                  }
                }}
                placeholder="https://example.com/callback"
                className="flex-1 px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
              />
              <button
                type="button"
                onClick={handleAddRedirectUri}
                className="p-2 bg-primary/10 text-primary hover:bg-primary/20 rounded-lg"
              >
                <Plus size={18} />
              </button>
            </div>
            <p className="text-xs text-muted-foreground mt-1">{t('oauth_edit.redirect_uris_hint')}</p>
          </div>
        </div>

        {/* Grant Types */}
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

        {/* Scopes */}
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

        {/* Token Settings */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('oauth_edit.token_settings')}</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-muted-foreground mb-1">{t('oauth_edit.access_token_ttl')}</label>
              <input
                type="number"
                value={formData.access_token_ttl}
                onChange={e => setFormData(prev => ({ ...prev, access_token_ttl: parseInt(e.target.value) }))}
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
                onChange={e => setFormData(prev => ({ ...prev, refresh_token_ttl: parseInt(e.target.value) }))}
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
                onChange={e => setFormData(prev => ({ ...prev, id_token_ttl: parseInt(e.target.value) }))}
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent"
                min={300}
                max={86400}
              />
              <p className="text-xs text-muted-foreground mt-1">{Math.floor((formData.id_token_ttl || 3600) / 60)} {t('oauth_edit.minutes')}</p>
            </div>
          </div>
        </div>

        {/* Security Settings */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('oauth_edit.security_settings')}</h2>
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={() => setFormData(prev => ({ ...prev, require_pkce: !prev.require_pkce }))}
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
                onClick={() => setFormData(prev => ({ ...prev, require_consent: !prev.require_consent }))}
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
                onClick={() => setFormData(prev => ({ ...prev, first_party: !prev.first_party }))}
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

        {/* Actions */}
        <div className="flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate('/oauth-clients')}
            className="px-4 py-2 border border-input rounded-lg text-muted-foreground hover:bg-accent"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={createClient.isPending || updateClient.isPending}
            className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg font-medium disabled:opacity-50"
          >
            <Save size={18} />
            {isNew ? t('oauth_edit.create_btn') : t('oauth_edit.save_btn')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default OAuthClientEdit;
