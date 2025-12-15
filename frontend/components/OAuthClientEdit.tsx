import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, Plus, X, Copy, Check, Eye, EyeOff } from 'lucide-react';
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

const GRANT_TYPES: { value: GrantType; label: string; description: string }[] = [
  { value: 'authorization_code', label: 'Authorization Code', description: 'For web apps with server-side code' },
  { value: 'client_credentials', label: 'Client Credentials', description: 'For machine-to-machine auth' },
  { value: 'refresh_token', label: 'Refresh Token', description: 'Allow token refresh' },
  { value: 'urn:ietf:params:oauth:grant-type:device_code', label: 'Device Code', description: 'For devices with limited input' },
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
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  // Show success modal with client secret after creation
  if (newClientSecret) {
    return (
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="bg-green-50 border border-green-200 rounded-xl p-6">
          <h2 className="text-xl font-bold text-green-800 mb-2">OAuth Client Created Successfully</h2>
          <p className="text-green-700 mb-4">
            Save this client secret now. You won't be able to see it again.
          </p>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-green-700 mb-1">Client Secret</label>
              <div className="flex items-center gap-2">
                <code className="flex-1 bg-white rounded-lg px-4 py-3 text-sm font-mono border border-green-200 break-all">
                  {showSecret ? newClientSecret : '••••••••••••••••••••••••••••••••'}
                </code>
                <button
                  onClick={() => setShowSecret(!showSecret)}
                  className="p-2 text-green-700 hover:bg-green-100 rounded-lg"
                >
                  {showSecret ? <EyeOff size={20} /> : <Eye size={20} />}
                </button>
                <button
                  onClick={copySecret}
                  className="p-2 text-green-700 hover:bg-green-100 rounded-lg"
                >
                  {copiedSecret ? <Check size={20} /> : <Copy size={20} />}
                </button>
              </div>
            </div>
          </div>
          <div className="mt-6 flex justify-end">
            <button
              onClick={() => navigate('/oauth-clients')}
              className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg font-medium"
            >
              Done
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
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft size={20} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            {isNew ? 'Create OAuth Client' : 'Edit OAuth Client'}
          </h1>
          <p className="text-gray-500 mt-1">
            {isNew ? 'Register a new OAuth 2.0 client application' : 'Update client configuration'}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Info */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Basic Information</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={e => setFormData(prev => ({ ...prev, name: e.target.value }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="My Application"
                required
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <textarea
                value={formData.description}
                onChange={e => setFormData(prev => ({ ...prev, description: e.target.value }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="A brief description of your application"
                rows={2}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Logo URL</label>
              <input
                type="url"
                value={formData.logo_url}
                onChange={e => setFormData(prev => ({ ...prev, logo_url: e.target.value }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="https://example.com/logo.png"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Client Type *</label>
              <select
                value={formData.client_type}
                onChange={e => setFormData(prev => ({ ...prev, client_type: e.target.value as ClientType }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={!isNew}
              >
                <option value="confidential">Confidential (Server-side)</option>
                <option value="public">Public (SPA, Mobile)</option>
              </select>
              <p className="text-xs text-gray-500 mt-1">
                {formData.client_type === 'confidential'
                  ? 'Can securely store client secret'
                  : 'Cannot securely store client secret, requires PKCE'}
              </p>
            </div>
          </div>
        </div>

        {/* Redirect URIs */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Redirect URIs</h2>
          <div className="space-y-3">
            {(formData.redirect_uris || []).map((uri, index) => (
              <div key={index} className="flex items-center gap-2">
                <input
                  type="text"
                  value={uri}
                  readOnly
                  className="flex-1 px-4 py-2 bg-gray-50 border border-gray-200 rounded-lg text-gray-600"
                />
                <button
                  type="button"
                  onClick={() => handleRemoveRedirectUri(uri)}
                  className="p-2 text-red-500 hover:bg-red-50 rounded-lg"
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
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
              <button
                type="button"
                onClick={handleAddRedirectUri}
                className="p-2 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded-lg"
              >
                <Plus size={18} />
              </button>
            </div>
            <p className="text-xs text-gray-500 mt-1">Press Enter or click + to add URI</p>
          </div>
        </div>

        {/* Grant Types */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Grant Types</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {GRANT_TYPES.map(grant => (
              <label
                key={grant.value}
                className={`flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors ${
                  formData.allowed_grant_types?.includes(grant.value)
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <input
                  type="checkbox"
                  checked={formData.allowed_grant_types?.includes(grant.value)}
                  onChange={() => handleGrantTypeToggle(grant.value)}
                  className="mt-1"
                />
                <div>
                  <div className="font-medium text-gray-900">{grant.label}</div>
                  <div className="text-sm text-gray-500">{grant.description}</div>
                </div>
              </label>
            ))}
          </div>
        </div>

        {/* Scopes */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Scopes</h2>
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium text-gray-700 mb-2">Allowed Scopes</h3>
              <div className="flex flex-wrap gap-2">
                {availableScopes.map(scope => (
                  <button
                    key={scope}
                    type="button"
                    onClick={() => handleScopeToggle(scope)}
                    className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                      formData.allowed_scopes?.includes(scope)
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {scope}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <h3 className="text-sm font-medium text-gray-700 mb-2">Default Scopes</h3>
              <p className="text-xs text-gray-500 mb-2">Scopes granted automatically if not specified in request</p>
              <div className="flex flex-wrap gap-2">
                {formData.allowed_scopes?.map(scope => (
                  <button
                    key={scope}
                    type="button"
                    onClick={() => handleScopeToggle(scope, true)}
                    className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
                      formData.default_scopes?.includes(scope)
                        ? 'bg-green-600 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
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
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Token Settings</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Access Token TTL (seconds)</label>
              <input
                type="number"
                value={formData.access_token_ttl}
                onChange={e => setFormData(prev => ({ ...prev, access_token_ttl: parseInt(e.target.value) }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                min={60}
                max={86400}
              />
              <p className="text-xs text-gray-500 mt-1">{Math.floor((formData.access_token_ttl || 900) / 60)} minutes</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Refresh Token TTL (seconds)</label>
              <input
                type="number"
                value={formData.refresh_token_ttl}
                onChange={e => setFormData(prev => ({ ...prev, refresh_token_ttl: parseInt(e.target.value) }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                min={3600}
                max={2592000}
              />
              <p className="text-xs text-gray-500 mt-1">{Math.floor((formData.refresh_token_ttl || 604800) / 86400)} days</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ID Token TTL (seconds)</label>
              <input
                type="number"
                value={formData.id_token_ttl}
                onChange={e => setFormData(prev => ({ ...prev, id_token_ttl: parseInt(e.target.value) }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                min={300}
                max={86400}
              />
              <p className="text-xs text-gray-500 mt-1">{Math.floor((formData.id_token_ttl || 3600) / 60)} minutes</p>
            </div>
          </div>
        </div>

        {/* Security Settings */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Security Settings</h2>
          <div className="space-y-4">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.require_pkce}
                onChange={e => setFormData(prev => ({ ...prev, require_pkce: e.target.checked }))}
                className="w-5 h-5 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <div>
                <div className="font-medium text-gray-900">Require PKCE</div>
                <div className="text-sm text-gray-500">Require Proof Key for Code Exchange (recommended)</div>
              </div>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.require_consent}
                onChange={e => setFormData(prev => ({ ...prev, require_consent: e.target.checked }))}
                className="w-5 h-5 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <div>
                <div className="font-medium text-gray-900">Require User Consent</div>
                <div className="text-sm text-gray-500">Show consent screen to users before granting access</div>
              </div>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.first_party}
                onChange={e => setFormData(prev => ({ ...prev, first_party: e.target.checked }))}
                className="w-5 h-5 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <div>
                <div className="font-medium text-gray-900">First Party Application</div>
                <div className="text-sm text-gray-500">Skip consent for trusted first-party applications</div>
              </div>
            </label>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate('/oauth-clients')}
            className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createClient.isPending || updateClient.isPending}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium disabled:opacity-50"
          >
            <Save size={18} />
            {isNew ? 'Create Client' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default OAuthClientEdit;
