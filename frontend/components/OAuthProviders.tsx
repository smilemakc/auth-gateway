
import React from 'react';
import { Link } from 'react-router-dom';
import { Plus, Edit2, Trash2, ToggleLeft, ToggleRight, Globe, Github, Send, Instagram, Loader2, Copy } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useOAuthProviders, useDeleteOAuthProvider, useToggleOAuthProvider } from '../hooks/useOAuth';
import { formatDate } from '../lib/date';
import { confirm } from '../services/confirm';
import { logger } from '@/lib/logger';

// Icon mapper
const getProviderIcon = (provider: string) => {
  switch (provider.toLowerCase()) {
    case 'google': return <span className="font-bold text-lg text-red-500">G</span>;
    case 'github': return <Github className="text-gray-900" size={24} />;
    case 'yandex': return <span className="font-bold text-lg text-red-600">Y</span>;
    case 'telegram': return <Send className="text-blue-500" size={24} />;
    case 'instagram': return <Instagram className="text-pink-600" size={24} />;
    case 'onec': return <span className="font-bold text-lg text-yellow-600">1C</span>;
    default: return <Globe className="text-gray-500" size={24} />;
  }
};

const OAuthProviders: React.FC = () => {
  const { t } = useLanguage();
  const { data: providersResponse, isLoading, error } = useOAuthProviders();
  const deleteProviderMutation = useDeleteOAuthProvider();
  const toggleProviderMutation = useToggleOAuthProvider();

  const providers = Array.isArray(providersResponse) ? providersResponse : [];

  const handleToggle = async (id: string, currentActive: boolean) => {
    try {
      await toggleProviderMutation.mutateAsync({ id, enabled: !currentActive });
    } catch (err) {
      logger.error('Failed to toggle provider:', err);
    }
  };

  const handleDelete = async (id: string) => {
    const ok = await confirm({
      title: t('confirm.delete_title'),
      description: t('common.confirm_delete'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await deleteProviderMutation.mutateAsync(id);
      } catch (err) {
        logger.error('Failed to delete provider:', err);
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 text-destructive">
        {t('oauth.load_error')}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('oauth.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('oauth.manage_desc')}</p>
        </div>
        <Link
          to="/oauth/new"
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('oauth.add')}
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {providers.map((provider) => (
          <div key={provider.id} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col">
            <div className="p-6 flex-1">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center shadow-sm">
                    {getProviderIcon(provider.provider)}
                  </div>
                  <div>
                    <h3 className="font-semibold text-foreground capitalize text-lg">{provider.provider}</h3>
                    <div className="flex items-center gap-2 mt-1">
                      <span className={`w-2 h-2 rounded-full ${provider.is_active ? 'bg-green-500' : 'bg-gray-300'}`}></span>
                      <span className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                        {provider.is_active ? t('common.enabled') : t('common.disabled')}
                      </span>
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => handleToggle(provider.id, provider.is_active)}
                  disabled={toggleProviderMutation.isPending}
                  className={`transition-colors ${provider.is_active ? 'text-success hover:text-success' : 'text-muted-foreground hover:text-muted-foreground'} disabled:opacity-50`}
                >
                  {provider.is_active ? <ToggleRight size={36} /> : <ToggleLeft size={36} />}
                </button>
              </div>

              <div className="space-y-3 mt-6">
                <div>
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">{t('oauth.client_id')}</label>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono truncate border border-border">
                      {provider.client_id
                        ? `${provider.client_id.slice(0, 4)}${'*'.repeat(Math.max(0, provider.client_id.length - 8))}${provider.client_id.slice(-4)}`
                        : '—'}
                    </code>
                    <button
                      onClick={() => navigator.clipboard.writeText(provider.client_id)}
                      className="p-2 text-muted-foreground hover:text-foreground transition-colors rounded-lg hover:bg-muted"
                      title={t('common.copy') || 'Copy'}
                    >
                      <Copy size={14} />
                    </button>
                  </div>
                </div>
                <div>
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">{t('oauth.callback_url')}</label>
                  <div className="text-xs text-muted-foreground truncate" title={provider.callback_url}>
                    {provider.callback_url || <span className="italic text-muted-foreground">{t('oauth.not_configured')}</span>}
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-muted px-6 py-4 border-t border-border flex items-center justify-between">
              <span className="text-xs text-muted-foreground">
                {provider.created_at ? formatDate(provider.created_at) : '-'}
              </span>
              <div className="flex items-center gap-2">
                <Link
                  to={`/oauth/${provider.id}`}
                  className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                >
                  <Edit2 size={18} />
                </Link>
                <button
                  onClick={() => handleDelete(provider.id)}
                  disabled={deleteProviderMutation.isPending}
                  className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors disabled:opacity-50"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
          </div>
        ))}

        {providers.length === 0 && (
          <div className="col-span-full text-center py-12 bg-card rounded-xl border border-border">
            <Globe size={48} className="mx-auto mb-4 text-muted-foreground opacity-50" />
            <p className="text-muted-foreground">{t('oauth.no_providers')}</p>
            <Link
              to="/oauth/new"
              className="mt-4 inline-block text-primary hover:underline text-sm font-medium"
            >
              {t('oauth.add_first')}
            </Link>
          </div>
        )}
      </div>
    </div>
  );
};

export default OAuthProviders;
