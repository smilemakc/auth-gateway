import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus, Edit2, Trash2, ToggleLeft, ToggleRight, Key, Copy, Check, Shield, Globe } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import {
  useOAuthClients,
  useDeleteOAuthClient,
  useActivateOAuthClient,
  useDeactivateOAuthClient,
  useRotateOAuthClientSecret,
} from '../hooks/useOAuthClients';
import type { OAuthClient } from '@auth-gateway/client-sdk';

const OAuthClients: React.FC = () => {
  const [page, setPage] = useState(1);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [newSecret, setNewSecret] = useState<{ clientId: string; secret: string } | null>(null);
  const pageSize = 20;
  const { t } = useLanguage();

  const { data, isLoading, error } = useOAuthClients(page, pageSize);
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
    if (window.confirm(t('common.confirm_delete'))) {
      await deleteClient.mutateAsync(id);
    }
  };

  const handleRotateSecret = async (clientId: string) => {
    if (window.confirm('Are you sure you want to rotate the client secret? The old secret will stop working immediately.')) {
      const result = await rotateSecret.mutateAsync(clientId);
      setNewSecret({ clientId, secret: result.client_secret });
    }
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const getClientTypeIcon = (clientType: string) => {
    if (clientType === 'confidential') {
      return <Shield className="text-success" size={20} />;
    }
    return <Globe className="text-primary" size={20} />;
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
        Failed to load OAuth clients. Please try again.
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
          <h1 className="text-2xl font-bold text-foreground">OAuth Clients</h1>
          <p className="text-muted-foreground mt-1">Manage OAuth 2.0 / OIDC client applications</p>
        </div>
        <Link
          to="/oauth-clients/new"
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          Add Client
        </Link>
      </div>

      {/* New Secret Alert */}
      {newSecret && (
        <div className="bg-primary/10 border border-primary rounded-lg p-4">
          <h3 className="font-semibold text-primary mb-2">New Client Secret Generated</h3>
          <p className="text-sm text-primary mb-3">
            Copy this secret now. You won't be able to see it again.
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
            Dismiss
          </button>
        </div>
      )}

      {/* Clients Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {clients.map((client) => (
          <div key={client.id} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col">
            <div className="p-6 flex-1">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center shadow-sm">
                    {getClientTypeIcon(client.client_type)}
                  </div>
                  <div>
                    <h3 className="font-semibold text-foreground text-lg">{client.name}</h3>
                    <div className="flex items-center gap-2 mt-1">
                      <span className={`w-2 h-2 rounded-full ${client.is_active ? 'bg-success' : 'bg-muted-foreground'}`}></span>
                      <span className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                        {client.is_active ? 'Active' : 'Inactive'}
                      </span>
                      <span className="text-xs text-muted-foreground">|</span>
                      <span className="text-xs text-muted-foreground capitalize">{client.client_type}</span>
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => handleToggle(client)}
                  className={`transition-colors ${client.is_active ? 'text-success hover:text-success' : 'text-muted-foreground hover:text-muted-foreground'}`}
                  disabled={activateClient.isPending || deactivateClient.isPending}
                >
                  {client.is_active ? <ToggleRight size={36} /> : <ToggleLeft size={36} />}
                </button>
              </div>

              {client.description && (
                <p className="text-sm text-muted-foreground mb-4 line-clamp-2">{client.description}</p>
              )}

              <div className="space-y-3">
                <div>
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">Client ID</label>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono truncate border border-border">
                      {client.client_id}
                    </code>
                    <button
                      onClick={() => copyToClipboard(client.client_id, `client-${client.id}`)}
                      className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-accent rounded"
                    >
                      {copiedId === `client-${client.id}` ? <Check size={14} /> : <Copy size={14} />}
                    </button>
                  </div>
                </div>
                <div>
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">Redirect URIs</label>
                  <div className="text-xs text-muted-foreground truncate" title={client.redirect_uris.join(', ')}>
                    {client.redirect_uris.length > 0 ? client.redirect_uris[0] : <span className="italic text-muted-foreground">None configured</span>}
                    {client.redirect_uris.length > 1 && (
                      <span className="text-muted-foreground"> (+{client.redirect_uris.length - 1} more)</span>
                    )}
                  </div>
                </div>
                <div>
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">Grant Types</label>
                  <div className="flex flex-wrap gap-1">
                    {client.allowed_grant_types.slice(0, 2).map((grant) => (
                      <span key={grant} className="px-2 py-0.5 bg-primary/10 text-primary text-xs rounded">
                        {grant.replace('urn:ietf:params:oauth:grant-type:', '')}
                      </span>
                    ))}
                    {client.allowed_grant_types.length > 2 && (
                      <span className="px-2 py-0.5 bg-muted text-muted-foreground text-xs rounded">
                        +{client.allowed_grant_types.length - 2}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-muted px-6 py-4 border-t border-border flex items-center justify-between">
              <span className="text-xs text-muted-foreground">
                {new Date(client.created_at).toLocaleDateString()}
              </span>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => handleRotateSecret(client.id)}
                  className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                  title="Rotate Secret"
                  disabled={rotateSecret.isPending}
                >
                  <Key size={18} />
                </button>
                <Link
                  to={`/oauth-clients/${client.id}`}
                  className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                >
                  <Edit2 size={18} />
                </Link>
                <button
                  onClick={() => handleDelete(client.id)}
                  className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                  disabled={deleteClient.isPending}
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {clients.length === 0 && (
        <div className="text-center py-12">
          <Shield className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-medium text-foreground">No OAuth clients</h3>
          <p className="mt-1 text-sm text-muted-foreground">Get started by creating a new OAuth client.</p>
          <div className="mt-6">
            <Link
              to="/oauth-clients/new"
              className="inline-flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium"
            >
              <Plus size={18} />
              Add Client
            </Link>
          </div>
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between bg-card px-4 py-3 rounded-lg border border-border">
          <div className="text-sm text-foreground">
            Showing <span className="font-medium">{(page - 1) * pageSize + 1}</span> to{' '}
            <span className="font-medium">{Math.min(page * pageSize, total)}</span> of{' '}
            <span className="font-medium">{total}</span> results
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              Previous
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default OAuthClients;
