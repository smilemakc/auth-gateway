import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, HelpCircle, Loader2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import {
  useApplicationOAuthProviderDetail,
  useCreateApplicationOAuthProvider,
  useUpdateApplicationOAuthProvider,
  useDeleteApplicationOAuthProvider
} from '../../hooks/useApplicationOAuthProviders';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';
import OAuthProviderSelectionSection from './OAuthProviderSelectionSection';
import OAuthProviderCredentialsSection from './OAuthProviderCredentialsSection';
import OAuthProviderAdvancedSection from './OAuthProviderAdvancedSection';

const ApplicationOAuthProviderEdit: React.FC = () => {
  const { applicationId, providerId } = useParams<{ applicationId: string; providerId: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isEditMode = providerId && providerId !== 'new';
  const isNewMode = !providerId || providerId === 'new';

  const [formData, setFormData] = useState({
    provider: 'google',
    client_id: '',
    client_secret: '',
    callback_url: '',
    scopes: '',
    auth_url: '',
    token_url: '',
    user_info_url: '',
    is_active: true
  });

  const { data: existingProvider, isLoading: loadingProvider } = useApplicationOAuthProviderDetail(
    applicationId || '',
    isEditMode ? providerId! : ''
  );
  const createMutation = useCreateApplicationOAuthProvider();
  const updateMutation = useUpdateApplicationOAuthProvider();
  const deleteMutation = useDeleteApplicationOAuthProvider();

  useEffect(() => {
    if (isEditMode && existingProvider) {
      setFormData({
        provider: existingProvider.provider || 'google',
        client_id: existingProvider.client_id || '',
        client_secret: existingProvider.client_secret || '',
        callback_url: existingProvider.callback_url || '',
        scopes: existingProvider.scopes?.join(', ') || '',
        auth_url: existingProvider.auth_url || '',
        token_url: existingProvider.token_url || '',
        user_info_url: existingProvider.user_info_url || '',
        is_active: existingProvider.is_active ?? true
      });
    }
  }, [existingProvider, isEditMode]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target;
    if (type === 'checkbox') {
      const checked = (e.target as HTMLInputElement).checked;
      setFormData(prev => ({ ...prev, [name]: checked }));
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!applicationId) {
      logger.error('Application ID is required');
      return;
    }

    const scopesArray = formData.scopes
      .split(',')
      .map(s => s.trim())
      .filter(s => s.length > 0);

    try {
      if (isNewMode) {
        await createMutation.mutateAsync({
          appId: applicationId,
          data: {
            provider: formData.provider,
            client_id: formData.client_id,
            client_secret: formData.client_secret,
            callback_url: formData.callback_url,
            scopes: scopesArray.length > 0 ? scopesArray : undefined,
            auth_url: formData.auth_url || undefined,
            token_url: formData.token_url || undefined,
            user_info_url: formData.user_info_url || undefined,
            is_active: formData.is_active
          }
        });
      } else if (providerId) {
        await updateMutation.mutateAsync({
          appId: applicationId,
          id: providerId,
          data: {
            client_id: formData.client_id,
            client_secret: formData.client_secret || undefined,
            callback_url: formData.callback_url,
            scopes: scopesArray.length > 0 ? scopesArray : undefined,
            auth_url: formData.auth_url || undefined,
            token_url: formData.token_url || undefined,
            user_info_url: formData.user_info_url || undefined,
            is_active: formData.is_active
          }
        });
      }
      navigate(`/applications/${applicationId}`);
    } catch (err) {
      logger.error('Failed to save provider:', err);
    }
  };

  const handleDelete = async () => {
    if (isEditMode && applicationId && providerId) {
      const ok = await confirm({
        title: t('confirm.delete_title'),
        description: t('common.confirm_delete'),
        variant: 'danger'
      });
      if (ok) {
        try {
          await deleteMutation.mutateAsync({ appId: applicationId, id: providerId });
          navigate(`/applications/${applicationId}`);
        } catch (err) {
          logger.error('Failed to delete provider:', err);
        }
      }
    }
  };

  const isLoading = createMutation.isPending || updateMutation.isPending;

  if (isEditMode && loadingProvider) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate(`/applications/${applicationId}`)}
          className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <h1 className="text-2xl font-bold text-foreground">
          {isEditMode ? t('app_oauth.edit_title') : t('app_oauth.add_title')}
        </h1>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="p-6 border-b border-border bg-muted flex items-start gap-3">
          <HelpCircle className="text-primary mt-0.5" size={20} />
          <div className="text-sm text-muted-foreground">
            <p className="font-medium text-foreground mb-1">{t('app_oauth.getting_started')}</p>
            <p>{t('app_oauth.getting_started_desc')}</p>
          </div>
        </div>

        <div className="p-6 space-y-8">
          <OAuthProviderSelectionSection
            selectedProvider={formData.provider}
            isEditMode={!!isEditMode}
            onChange={handleChange}
          />

          <OAuthProviderCredentialsSection
            clientId={formData.client_id}
            clientSecret={formData.client_secret}
            callbackUrl={formData.callback_url}
            scopes={formData.scopes}
            isEditMode={!!isEditMode}
            onChange={handleChange}
          />

          <OAuthProviderAdvancedSection
            authUrl={formData.auth_url}
            tokenUrl={formData.token_url}
            userInfoUrl={formData.user_info_url}
            isActive={formData.is_active}
            onChange={handleChange}
            onToggleActive={() => setFormData(prev => ({ ...prev, is_active: !prev.is_active }))}
          />
        </div>

        <div className="px-6 py-4 bg-muted border-t border-border flex items-center justify-between">
          <div>
            {isEditMode && (
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
                className="text-destructive hover:text-destructive text-sm font-medium px-2 py-1 rounded hover:bg-destructive/10 transition-colors disabled:opacity-50"
              >
                {t('common.delete')}
              </button>
            )}
          </div>
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => navigate(`/applications/${applicationId}`)}
              className="px-4 py-2 text-sm font-medium text-muted-foreground bg-card border border-input rounded-lg hover:bg-accent focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className={`flex items-center px-4 py-2 text-sm font-medium text-primary-foreground bg-primary border border-transparent rounded-lg hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring
                ${isLoading ? 'opacity-70 cursor-not-allowed' : ''}`}
            >
              {isLoading ? (
                <Loader2 size={16} className="mr-2 animate-spin" />
              ) : (
                <Save size={16} className="mr-2" />
              )}
              {isEditMode ? t('common.save') : t('common.create')}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
};

export default ApplicationOAuthProviderEdit;
