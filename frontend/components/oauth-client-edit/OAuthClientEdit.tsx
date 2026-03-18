import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Save } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { LoadingSpinner, PageHeader } from '../ui';
import {
  useOAuthClientDetail,
  useCreateOAuthClient,
  useUpdateOAuthClient,
  useOAuthScopes,
} from '../../hooks/useOAuthClients';
import type {
  CreateOAuthClientRequest,
  UpdateOAuthClientRequest,
} from '@auth-gateway/client-sdk';
import { logger } from '@/lib/logger';
import OAuthClientBasicFields from './OAuthClientBasicFields';
import OAuthClientScopeSelector from './OAuthClientScopeSelector';
import OAuthClientSecretSection from './OAuthClientSecretSection';

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
      logger.error('Failed to save client:', error);
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

  if (!isNew && loadingClient) {
    return <LoadingSpinner />;
  }

  if (newClientSecret) {
    return (
      <OAuthClientSecretSection
        clientSecret={newClientSecret}
        onDone={() => navigate('/oauth-clients')}
      />
    );
  }

  const availableScopes = [
    ...STANDARD_SCOPES,
    ...(scopesData?.scopes?.filter(s => !STANDARD_SCOPES.includes(s.name)).map(s => s.name) || []),
  ];

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <PageHeader
        title={isNew ? t('oauth_edit.create_title') : t('oauth_edit.edit_title')}
        subtitle={isNew ? t('oauth_edit.create_desc') : t('oauth_edit.edit_desc')}
        onBack={() => navigate('/oauth-clients')}
      />

      <form onSubmit={handleSubmit} className="space-y-6">
        <OAuthClientBasicFields
          formData={formData}
          isNew={isNew}
          newRedirectUri={newRedirectUri}
          onFormChange={setFormData}
          onNewRedirectUriChange={setNewRedirectUri}
          onAddRedirectUri={handleAddRedirectUri}
          onRemoveRedirectUri={handleRemoveRedirectUri}
        />

        <OAuthClientScopeSelector
          formData={formData}
          availableScopes={availableScopes}
          onFormChange={setFormData}
        />

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
