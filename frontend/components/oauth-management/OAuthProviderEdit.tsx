
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Save, HelpCircle, Loader2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useOAuthProviderDetail, useCreateOAuthProvider, useUpdateOAuthProvider, useDeleteOAuthProvider } from '../../hooks/useOAuth';
import { LoadingSpinner, PageHeader } from '../ui';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';
import { OAuthProviderFormFields } from './OAuthProviderFormFields';

const OAuthProviderEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isEditMode = id && id !== 'new';
  const isNewMode = !id || id === 'new';

  const [showSecret, setShowSecret] = useState(false);
  const [formData, setFormData] = useState({
    provider: 'google',
    client_id: '',
    client_secret: '',
    callback_url: '',
    scopes: '',
    is_active: true
  });

  const { data: existingProvider, isLoading: loadingProvider } = useOAuthProviderDetail(isEditMode ? id! : '');
  const createMutation = useCreateOAuthProvider();
  const updateMutation = useUpdateOAuthProvider();
  const deleteMutation = useDeleteOAuthProvider();

  useEffect(() => {
    if (isEditMode && existingProvider) {
      setFormData({
        provider: existingProvider.provider || 'google',
        client_id: existingProvider.client_id || '',
        client_secret: '',
        callback_url: existingProvider.callback_url || '',
        scopes: existingProvider.scopes?.join(', ') || '',
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

    const scopes = formData.scopes
      ? formData.scopes.split(',').map(s => s.trim()).filter(Boolean)
      : undefined;

    try {
      if (isNewMode) {
        await createMutation.mutateAsync({
          provider: formData.provider as any,
          client_id: formData.client_id,
          client_secret: formData.client_secret,
          callback_url: formData.callback_url,
          scopes,
        });
      } else if (id) {
        await updateMutation.mutateAsync({
          id,
          data: {
            client_id: formData.client_id,
            ...(formData.client_secret ? { client_secret: formData.client_secret } : {}),
            callback_url: formData.callback_url,
            scopes,
            is_active: formData.is_active,
          }
        });
      }
      navigate('/oauth');
    } catch (err) {
      logger.error('Failed to save provider:', err);
    }
  };

  const handleDelete = async () => {
    if (isEditMode && id) {
      const ok = await confirm({
        title: t('confirm.delete_title'),
        description: t('common.confirm_delete'),
        variant: 'danger'
      });
      if (ok) {
        try {
          await deleteMutation.mutateAsync(id);
          navigate('/oauth');
        } catch (err) {
          logger.error('Failed to delete provider:', err);
        }
      }
    }
  };

  const isLoading = createMutation.isPending || updateMutation.isPending;

  if (isEditMode && loadingProvider) {
    return <LoadingSpinner />;
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <PageHeader
        title={isEditMode ? t('oauth.configure') : t('oauth.add')}
        onBack={() => navigate('/oauth')}
      />

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="p-6 border-b border-border bg-muted flex items-start gap-3">
          <HelpCircle className="text-primary mt-0.5" size={20} />
          <div className="text-sm text-muted-foreground">
            <p className="font-medium text-foreground mb-1">Getting Started</p>
            <p>To configure this provider, you need to create an OAuth application in the provider's developer console.</p>
          </div>
        </div>

        <OAuthProviderFormFields
          formData={formData}
          isEditMode={!!isEditMode}
          isNewMode={!!isNewMode}
          showSecret={showSecret}
          onToggleShowSecret={() => setShowSecret(!showSecret)}
          onChange={handleChange}
          onToggleActive={() => setFormData(prev => ({ ...prev, is_active: !prev.is_active }))}
        />

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
              onClick={() => navigate('/oauth')}
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

export default OAuthProviderEdit;
