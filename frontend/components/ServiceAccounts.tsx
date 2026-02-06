
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus, Trash2, Edit2, Bot, CheckCircle, XCircle, Copy, Loader2 } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useOAuthClients, useDeleteOAuthClient } from '../hooks/useOAuthClients';
import { formatDate } from '../lib/date';

const ServiceAccounts: React.FC = () => {
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const { t } = useLanguage();

  const { data: clientsResponse, isLoading, error } = useOAuthClients();
  const deleteClientMutation = useDeleteOAuthClient();

  // Filter to only show service accounts (clients with client_credentials grant type)
  const accounts = (clientsResponse?.clients || []).filter(
    (client: any) => client.allowed_grant_types?.includes('client_credentials')
  );

  const handleDelete = async (id: string) => {
    if (window.confirm(t('common.confirm_delete'))) {
      try {
        await deleteClientMutation.mutateAsync(id);
      } catch (err: any) {
        console.error('Failed to delete service account:', err);
        alert(err?.message || 'Failed to delete service account');
      }
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
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 text-destructive">
        Failed to load service accounts. Please try again.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('sa.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('sa.desc')}</p>
        </div>
        <Link
          to="/developers/service-accounts/new"
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('sa.create')}
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {accounts.map((account) => (
          <div key={account.id} className="bg-card rounded-xl shadow-sm border border-border p-6 flex flex-col">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-primary/10 text-primary">
                  <Bot size={24} />
                </div>
                <div>
                  <h3 className="font-semibold text-foreground">{account.name}</h3>
                  <div className="flex items-center gap-2 mt-0.5">
                    {account.is_active
                      ? <span className="text-success text-xs font-medium flex items-center"><CheckCircle size={10} className="mr-1"/> {t('users.active')}</span>
                      : <span className="text-muted-foreground text-xs font-medium flex items-center"><XCircle size={10} className="mr-1"/> Disabled</span>
                    }
                  </div>
                </div>
              </div>
            </div>

            <p className="text-sm text-muted-foreground mb-4 line-clamp-2 h-10">
              {account.description || 'No description provided.'}
            </p>

            <div className="bg-muted rounded-lg p-3 mb-6 border border-border">
              <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">Client ID</label>
              <div className="flex items-center justify-between">
                <code className="text-sm font-mono text-foreground truncate mr-2">{account.client_id}</code>
                <button
                  onClick={() => copyToClipboard(account.client_id, account.id)}
                  className="text-muted-foreground hover:text-primary"
                >
                  {copiedId === account.id ? <CheckCircle size={16} className="text-success"/> : <Copy size={16}/>}
                </button>
              </div>
            </div>

            <div className="mt-auto flex items-center justify-between pt-4 border-t border-border">
               <span className="text-xs text-muted-foreground">{account.created_at ? formatDate(account.created_at) : '-'}</span>
               <div className="flex gap-2">
                 <Link
                   to={`/developers/service-accounts/${account.id}`}
                   className="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded transition-colors"
                 >
                   <Edit2 size={18} />
                 </Link>
                 <button
                   onClick={() => handleDelete(account.id)}
                   disabled={deleteClientMutation.isPending}
                   className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded transition-colors disabled:opacity-50"
                 >
                   <Trash2 size={18} />
                 </button>
               </div>
            </div>
          </div>
        ))}
        {accounts.length === 0 && (
           <div className="col-span-full py-12 text-center text-muted-foreground bg-card rounded-xl border border-border border-dashed">
            <Bot size={48} className="mx-auto text-muted-foreground mb-3" />
            <h3 className="text-lg font-medium text-foreground">No Service Accounts</h3>
          </div>
        )}
      </div>
    </div>
  );
};

export default ServiceAccounts;
