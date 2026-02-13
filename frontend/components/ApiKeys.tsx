
import React, { useState } from 'react';
import { Trash2, Ban, Copy, CheckCircle, X, Plus, Key } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useApiKeys, useRevokeApiKey, useDeleteApiKey, useCreateApiKey } from '../hooks/useApiKeys';
import { formatDate } from '../lib/date';
import { toast } from '../services/toast';
import { confirm } from '../services/confirm';

const AVAILABLE_SCOPES = [
  'users:read',
  'users:write',
  'users:sync',
  'users:import',
  'profile:read',
  'profile:write',
  'token:validate',
  'token:introspect',
  'auth:login',
  'auth:register',
  'auth:otp',
  'email:send',
  'oauth:read',
  'exchange:manage',
  'admin:all',
  'all',
];

const ApiKeys: React.FC = () => {
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyScopes, setNewKeyScopes] = useState<string[]>(['token:validate']);
  const [newKeyExpiresIn, setNewKeyExpiresIn] = useState<number | undefined>(undefined);
  const [generatedKey, setGeneratedKey] = useState<string | null>(null);
  const { t } = useLanguage();

  // Fetch API keys with React Query
  const { data, isLoading, error } = useApiKeys(1, 100);
  const revokeApiKeyMutation = useRevokeApiKey();
  const deleteApiKeyMutation = useDeleteApiKey();
  const createApiKeyMutation = useCreateApiKey();

  const keys = data?.api_keys || [];

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleRevoke = async (id: string) => {
    const ok = await confirm({
      description: t('keys.revoke_confirm'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await revokeApiKeyMutation.mutateAsync(id);
      } catch (error) {
        console.error('Failed to revoke API key:', error);
        toast.error('Failed to revoke API key');
      }
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
        await deleteApiKeyMutation.mutateAsync(id);
      } catch (error) {
        console.error('Failed to delete API key:', error);
        toast.error('Failed to delete API key');
      }
    }
  };

  const handleCreateKey = async () => {
    if (!newKeyName.trim()) {
      toast.warning(t('keys.name_required'));
      return;
    }
    if (newKeyScopes.length === 0) {
      toast.warning(t('keys.scopes_required'));
      return;
    }

    try {
      const expiresAt = newKeyExpiresIn
        ? new Date(Date.now() + newKeyExpiresIn * 24 * 60 * 60 * 1000).toISOString()
        : undefined;
      const result = await createApiKeyMutation.mutateAsync({
        name: newKeyName,
        scopes: newKeyScopes,
        expires_at: expiresAt,
      });
      // The API returns the full key only once - show it to the user
      setGeneratedKey(result.plain_key);
    } catch (error) {
      console.error('Failed to create API key:', error);
      toast.error('Failed to create API key: ' + (error as Error).message);
    }
  };

  const handleCloseModal = () => {
    setShowCreateModal(false);
    setNewKeyName('');
    setNewKeyScopes(['token:validate']);
    setNewKeyExpiresIn(undefined);
    setGeneratedKey(null);
  };

  const toggleScope = (scope: string) => {
    setNewKeyScopes(prev =>
      prev.includes(scope)
        ? prev.filter(s => s !== scope)
        : [...prev, scope]
    );
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
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('keys.generate')}
        </button>
      </div>

      {/* Create API Key Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
          <div className="bg-card rounded-xl shadow-xl border border-border w-full max-w-md">
            <div className="flex items-center justify-between p-6 border-b border-border">
              <h2 className="text-lg font-semibold text-foreground">
                {generatedKey ? t('keys.key_generated') : t('keys.generate')}
              </h2>
              <button
                onClick={handleCloseModal}
                className="text-muted-foreground hover:text-foreground transition-colors"
              >
                <X size={20} />
              </button>
            </div>

            <div className="p-6 space-y-4">
              {generatedKey ? (
                <>
                  <div className="bg-warning/10 border border-warning/20 rounded-lg p-4">
                    <p className="text-sm text-warning font-medium mb-2">
                      {t('keys.copy_warning')}
                    </p>
                    <div className="flex items-center gap-2 bg-card rounded border border-border p-2">
                      <code className="flex-1 text-sm font-mono text-foreground break-all">
                        {generatedKey}
                      </code>
                      <button
                        onClick={() => {
                          navigator.clipboard.writeText(generatedKey);
                          setCopiedId('new-key');
                          setTimeout(() => setCopiedId(null), 2000);
                        }}
                        className="p-2 text-muted-foreground hover:text-primary transition-colors"
                      >
                        {copiedId === 'new-key' ? <CheckCircle size={18} className="text-success" /> : <Copy size={18} />}
                      </button>
                    </div>
                  </div>
                  <button
                    onClick={handleCloseModal}
                    className="w-full bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg font-medium transition-colors"
                  >
                    {t('common.done')}
                  </button>
                </>
              ) : (
                <>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                      {t('keys.name')}
                    </label>
                    <input
                      type="text"
                      value={newKeyName}
                      onChange={(e) => setNewKeyName(e.target.value)}
                      placeholder={t('keys.name_placeholder')}
                      className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                      {t('keys.scopes')}
                    </label>
                    <div className="flex flex-wrap gap-2">
                      {AVAILABLE_SCOPES.map(scope => (
                        <button
                          key={scope}
                          onClick={() => toggleScope(scope)}
                          className={`px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors ${
                            newKeyScopes.includes(scope)
                              ? 'bg-primary text-primary-foreground border-primary'
                              : 'bg-background text-muted-foreground border-input hover:border-primary'
                          }`}
                        >
                          {scope}
                        </button>
                      ))}
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">
                      {t('keys.expires_in')}
                    </label>
                    <select
                      value={newKeyExpiresIn || ''}
                      onChange={(e) => setNewKeyExpiresIn(e.target.value ? parseInt(e.target.value) : undefined)}
                      className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                    >
                      <option value="">{t('keys.never')}</option>
                      <option value="30">30 {t('common.days')}</option>
                      <option value="90">90 {t('common.days')}</option>
                      <option value="180">180 {t('common.days')}</option>
                      <option value="365">365 {t('common.days')}</option>
                    </select>
                  </div>

                  <div className="flex gap-3 pt-2">
                    <button
                      onClick={handleCloseModal}
                      className="flex-1 px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
                    >
                      {t('common.cancel')}
                    </button>
                    <button
                      onClick={handleCreateKey}
                      disabled={createApiKeyMutation.isPending}
                      className="flex-1 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg font-medium transition-colors disabled:opacity-50"
                    >
                      {createApiKeyMutation.isPending ? t('common.creating') : t('keys.generate')}
                    </button>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {keys.length === 0 && !isLoading && (
        <div className="text-center py-12 bg-card rounded-xl border border-border">
          <Key className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-semibold text-foreground">{t('keys.no_keys')}</h3>
          <p className="mt-1 text-sm text-muted-foreground">{t('keys.no_keys_desc')}</p>
        </div>
      )}

      {keys.length > 0 && (
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
                 <span className="text-muted-foreground">{t('common.created')}: {formatDate(apiKey.created_at)}</span>

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
      )}
    </div>
  );
};

export default ApiKeys;
