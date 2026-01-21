
import React, { useState } from 'react';
import { Trash2, Ban, Copy, CheckCircle } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useApiKeys, useRevokeApiKey, useDeleteApiKey } from '../hooks/useApiKeys';

const ApiKeys: React.FC = () => {
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const { t } = useLanguage();

  // Fetch API keys with React Query
  const { data, isLoading, error } = useApiKeys(1, 100);
  const revokeApiKeyMutation = useRevokeApiKey();
  const deleteApiKeyMutation = useDeleteApiKey();

  const keys = data?.api_keys || [];

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleRevoke = async (id: string) => {
    if (window.confirm(t('keys.revoke_confirm'))) {
      try {
        await revokeApiKeyMutation.mutateAsync(id);
      } catch (error) {
        console.error('Failed to revoke API key:', error);
        alert('Failed to revoke API key');
      }
    }
  };

  const handleDelete = async (id: string) => {
    if (window.confirm(t('common.confirm_delete'))) {
      try {
        await deleteApiKeyMutation.mutateAsync(id);
      } catch (error) {
        console.error('Failed to delete API key:', error);
        alert('Failed to delete API key');
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive">Error loading API keys: {(error as Error).message}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h1 className="text-2xl font-bold text-foreground">{t('keys.title')}</h1>
        <button className="bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors">
          + {t('keys.generate')}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {keys.map((apiKey) => (
          <div key={apiKey.id} className="bg-card rounded-xl shadow-sm border border-border p-6 flex flex-col">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-lg ${apiKey.is_active ? 'bg-amber-50 text-amber-600' : 'bg-muted text-muted-foreground'}`}>
                  <div className="font-mono text-xl font-bold">K</div>
                </div>
                <div>
                  <h3 className="font-semibold text-foreground">{apiKey.name}</h3>
                  <p className="text-sm text-muted-foreground">{t('keys.owner')}: {apiKey.user_email || apiKey.user_name || 'N/A'}</p>
                </div>
              </div>
              <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                apiKey.is_active
                  ? 'bg-success/10 text-success'
                  : 'bg-destructive/10 text-destructive'
              }`}>
                {apiKey.is_active ? t('users.active') : t('keys.revoked')}
              </span>
            </div>

            <div className="bg-muted rounded-md p-3 mb-4 flex items-center justify-between group">
              <code className="text-sm text-muted-foreground font-mono">
                {apiKey.key_prefix}****************
              </code>
              <button
                onClick={() => handleCopy(apiKey.key_prefix, apiKey.id)}
                className="text-muted-foreground hover:text-primary transition-colors"
              >
                {copiedId === apiKey.id ? <CheckCircle size={16} className="text-success" /> : <Copy size={16} />}
              </button>
            </div>

            <div className="flex flex-wrap gap-2 mb-6">
              {apiKey.scopes.map(scope => (
                <span key={scope} className="px-2 py-1 bg-primary/10 text-primary text-xs rounded border border-border">
                  {scope}
                </span>
              ))}
            </div>

            <div className="mt-auto pt-4 border-t border-border flex items-center justify-between text-sm">
               <span className="text-muted-foreground">{t('common.created')}: {new Date(apiKey.created_at).toLocaleDateString()}</span>

               <div className="flex gap-2">
                 {apiKey.is_active && (
                   <button
                    onClick={() => handleRevoke(apiKey.id)}
                    className="p-1.5 text-muted-foreground hover:text-orange-600 hover:bg-orange-50 rounded transition-colors" title={t('user.revoke')}>
                     <Ban size={18} />
                   </button>
                 )}
                 <button
                   onClick={() => handleDelete(apiKey.id)}
                   className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded transition-colors" title={t('common.delete')}>
                   <Trash2 size={18} />
                 </button>
               </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ApiKeys;
