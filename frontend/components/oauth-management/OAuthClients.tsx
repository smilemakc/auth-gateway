import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus, Copy, Check, Shield } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import {
  useOAuthClients,
  useDeleteOAuthClient,
  useActivateOAuthClient,
  useDeactivateOAuthClient,
  useRotateOAuthClientSecret,
} from '../../hooks/useOAuthClients';
import type { OAuthClient } from '@auth-gateway/client-sdk';
import { confirm } from '../../services/confirm';
import { OAuthClientCard } from './OAuthClientCard';

const OAuthClients: React.FC = () => {
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [newSecret, setNewSecret] = useState<{ clientId: string; secret: string } | null>(null);
  const pageSize = 20;
  const { t } = useLanguage();

  const isActive = statusFilter === 'active' ? true : statusFilter === 'inactive' ? false : undefined;
  const { data, isLoading, error } = useOAuthClients(page, pageSize, isActive);
  const deleteClient = useDeleteOAuthClient();
  const activateClient = useActivateOAuthClient();
  const deactivateClient = useDeactivateOAuthClient();
  const rotateSecret = useRotateOAuthClientSecret();

  const handleToggle = async (client: OAuthClient) => {
    if (client.is_active) {
      await deactivateClient.mutateAsync(client.id);
    } else {
      await activateClient.mutateAsync(client.id);
    }
  };

  const handleDelete = async (id: string) => {
    const ok = await confirm({
      title: t('confirm.delete_title'),
      description: t('common.confirm_delete'),
      variant: 'danger'
    });
    if (ok) {
      await deleteClient.mutateAsync(id);
    }
  };

  const handleRotateSecret = async (clientId: string) => {
    const ok = await confirm({
      description: t('oauth_clients.rotate_confirm'),
      variant: 'danger'
    });
    if (ok) {
      const result = await rotateSecret.mutateAsync(clientId);
      setNewSecret({ clientId, secret: result.client_secret });
    }
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 border border-destructive rounded-lg p-4 text-destructive">
        {t('oauth_clients.load_error')}
      </div>
    );
  }

  const clients = data?.clients || [];
  const total = data?.total || 0;
  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('oauth_clients.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('oauth_clients.desc')}</p>
        </div>
        <Link
          to="/oauth-clients/new"
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('oauth_clients.add')}
        </Link>
      </div>

      <div className="flex flex-wrap gap-3">
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          className="border border-input rounded-lg px-3 py-2 text-sm bg-card text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        >
          <option value="">{t('oauth_clients.filter_status')}</option>
          <option value="active">{t('common.active')}</option>
          <option value="inactive">{t('common.inactive')}</option>
        </select>
      </div>

      {/* New Secret Alert */}
      {newSecret && (
        <div className="bg-primary/10 border border-primary rounded-lg p-4">
          <h3 className="font-semibold text-primary mb-2">{t('oauth_clients.new_secret')}</h3>
          <p className="text-sm text-primary mb-3">
            {t('oauth_clients.new_secret_desc')}
          </p>
          <div className="flex items-center gap-2">
            <code className="flex-1 bg-primary/20 rounded px-3 py-2 text-sm font-mono text-primary break-all">
              {newSecret.secret}
            </code>
            <button
              onClick={() => copyToClipboard(newSecret.secret, 'new-secret')}
              className="p-2 text-primary hover:bg-primary/20 rounded"
            >
              {copiedId === 'new-secret' ? <Check size={18} /> : <Copy size={18} />}
            </button>
          </div>
          <button
            onClick={() => setNewSecret(null)}
            className="mt-3 text-sm text-primary hover:text-primary"
          >
            {t('common.dismiss')}
          </button>
        </div>
      )}

      {/* Clients Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {clients.map((client) => (
          <OAuthClientCard
            key={client.id}
            client={client}
            copiedId={copiedId}
            isToggling={activateClient.isPending || deactivateClient.isPending}
            isDeleting={deleteClient.isPending}
            isRotating={rotateSecret.isPending}
            onToggle={handleToggle}
            onDelete={handleDelete}
            onRotateSecret={handleRotateSecret}
            onCopy={copyToClipboard}
          />
        ))}
      </div>

      {clients.length === 0 && (
        <div className="text-center py-12">
          <Shield className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-medium text-foreground">{t('oauth_clients.no_clients')}</h3>
          <p className="mt-1 text-sm text-muted-foreground">{t('oauth_clients.no_clients_desc')}</p>
          <div className="mt-6">
            <Link
              to="/oauth-clients/new"
              className="inline-flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium"
            >
              <Plus size={18} />
              {t('oauth_clients.add')}
            </Link>
          </div>
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between bg-card px-4 py-3 rounded-lg border border-border">
          <div className="text-sm text-foreground">
            {t('common.showing')} <span className="font-medium">{(page - 1) * pageSize + 1}</span> {t('common.to')}{' '}
            <span className="font-medium">{Math.min(page * pageSize, total)}</span> {t('common.of')}{' '}
            <span className="font-medium">{total}</span> {t('common.results')}
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.previous')}
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.next')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default OAuthClients;
