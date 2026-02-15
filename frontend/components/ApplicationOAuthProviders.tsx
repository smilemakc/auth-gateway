
import React from 'react';
import { Link } from 'react-router-dom';
import { Plus, Edit2, Trash2, Globe, Github, Send, Instagram, Loader2, CheckCircle2, XCircle } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useApplicationOAuthProviders, useDeleteApplicationOAuthProvider } from '../hooks/useApplicationOAuthProviders';
import { formatDate } from '../lib/date';
import { confirm } from '../services/confirm';
import { logger } from '@/lib/logger';

interface ApplicationOAuthProvidersProps {
  applicationId: string;
}

const getProviderIcon = (provider: string) => {
  switch (provider.toLowerCase()) {
    case 'google': return <span className="font-bold text-lg text-red-500">G</span>;
    case 'github': return <Github className="text-gray-900" size={24} />;
    case 'yandex': return <span className="font-bold text-lg text-red-600">Y</span>;
    case 'telegram': return <Send className="text-blue-500" size={24} />;
    case 'instagram': return <Instagram className="text-pink-600" size={24} />;
    default: return <Globe className="text-gray-500" size={24} />;
  }
};

const maskClientId = (clientId: string) => {
  if (!clientId || clientId.length <= 8) return clientId;
  return `${clientId.substring(0, 4)}...${clientId.substring(clientId.length - 4)}`;
};

const ApplicationOAuthProviders: React.FC<ApplicationOAuthProvidersProps> = ({ applicationId }) => {
  const { t } = useLanguage();
  const { data: providersResponse, isLoading, error } = useApplicationOAuthProviders(applicationId);
  const deleteProviderMutation = useDeleteApplicationOAuthProvider();

  const providers = Array.isArray(providersResponse) ? providersResponse : [];

  const handleDelete = async (id: string) => {
    const ok = await confirm({
      title: t('confirm.delete_title'),
      description: t('common.confirm_delete'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await deleteProviderMutation.mutateAsync({ appId: applicationId, id });
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
        {t('app_oauth.load_error')}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold text-foreground">{t('app_oauth.title')}</h2>
          <p className="text-muted-foreground mt-1">{t('app_oauth.desc')}</p>
        </div>
        <Link
          to={`/applications/${applicationId}/oauth/new`}
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('app_oauth.add_provider')}
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
                      {provider.is_active ? (
                        <>
                          <CheckCircle2 className="w-4 h-4 text-green-500" />
                          <span className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                            {t('common.active')}
                          </span>
                        </>
                      ) : (
                        <>
                          <XCircle className="w-4 h-4 text-gray-300" />
                          <span className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                            {t('common.inactive')}
                          </span>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              <div className="space-y-3 mt-6">
                <div>
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">{t('app_oauth.client_id')}</label>
                  <code className="block bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono truncate border border-border">
                    {maskClientId(provider.client_id)}
                  </code>
                </div>
                <div>
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">{t('app_oauth.callback_url')}</label>
                  <div className="text-xs text-muted-foreground truncate" title={provider.callback_url}>
                    {provider.callback_url || <span className="italic text-muted-foreground">{t('app_oauth.not_configured')}</span>}
                  </div>
                </div>
                {provider.scopes && provider.scopes.length > 0 && (
                  <div>
                    <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">{t('app_oauth.scopes')}</label>
                    <div className="flex flex-wrap gap-1 mt-1">
                      {provider.scopes.map((scope, idx) => (
                        <span key={idx} className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded">
                          {scope}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>

            <div className="bg-muted px-6 py-4 border-t border-border flex items-center justify-between">
              <span className="text-xs text-muted-foreground">
                {provider.created_at ? formatDate(provider.created_at) : '-'}
              </span>
              <div className="flex items-center gap-2">
                <Link
                  to={`/applications/${applicationId}/oauth/${provider.id}`}
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
            <p className="text-muted-foreground">{t('app_oauth.no_providers')}</p>
            <Link
              to={`/applications/${applicationId}/oauth/new`}
              className="mt-4 inline-block text-primary hover:underline text-sm font-medium"
            >
              {t('app_oauth.add_first')}
            </Link>
          </div>
        )}
      </div>
    </div>
  );
};

export default ApplicationOAuthProviders;
