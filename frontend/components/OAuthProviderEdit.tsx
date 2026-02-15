
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, HelpCircle, Eye, EyeOff, Loader2, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useOAuthProviderDetail, useCreateOAuthProvider, useUpdateOAuthProvider, useDeleteOAuthProvider } from '../hooks/useOAuth';
import { confirm } from '../services/confirm';
import { logger } from '@/lib/logger';

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
          onClick={() => navigate('/oauth')}
          className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <h1 className="text-2xl font-bold text-foreground">{isEditMode ? t('oauth.configure') : t('oauth.add')}</h1>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="p-6 border-b border-border bg-muted flex items-start gap-3">
          <HelpCircle className="text-primary mt-0.5" size={20} />
          <div className="text-sm text-muted-foreground">
            <p className="font-medium text-foreground mb-1">Getting Started</p>
            <p>To configure this provider, you need to create an OAuth application in the provider's developer console.</p>
          </div>
        </div>

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
                    onChange={handleChange}
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
                onChange={handleChange}
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
                  onChange={handleChange}
                  required={isNewMode}
                  className="w-full pl-4 pr-12 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all font-mono text-sm"
                  placeholder={isEditMode ? '(unchanged)' : 'e.g. GOCSPX-...'}
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

          {/* Callback URL */}
          <div>
            <label htmlFor="callback_url" className="block text-sm font-medium text-muted-foreground mb-1">Callback URL</label>
            <input
              type="url"
              id="callback_url"
              name="callback_url"
              value={formData.callback_url}
              onChange={handleChange}
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
              onChange={handleChange}
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
                 onClick={() => setFormData(prev => ({ ...prev, is_active: !prev.is_active }))}
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
