
import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus, Trash2, Edit2, Activity, CheckCircle, XCircle, Eye, EyeOff, Loader2, Network } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useWebhooks, useDeleteWebhook } from '../hooks/useWebhooks';
import { formatRelative } from '../lib/date';
import { confirm } from '../services/confirm';
import { logger } from '@/lib/logger';

const Webhooks: React.FC = () => {
  const [revealedSecrets, setRevealedSecrets] = useState<string[]>([]);
  const { t } = useLanguage();

  const { data: webhooksResponse, isLoading, error } = useWebhooks();
  const deleteWebhookMutation = useDeleteWebhook();

  const webhooks = webhooksResponse?.webhooks || [];

  const handleDelete = async (id: string) => {
    const ok = await confirm({
      description: t('common.confirm_delete'),
      title: t('confirm.delete_title'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await deleteWebhookMutation.mutateAsync(id);
      } catch (err) {
        logger.error('Failed to delete webhook:', err);
      }
    }
  };

  const toggleSecret = (id: string) => {
    if (revealedSecrets.includes(id)) {
      setRevealedSecrets(prev => prev.filter(i => i !== id));
    } else {
      setRevealedSecrets(prev => [...prev, id]);
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
        {t('hooks.load_error')}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('hooks.title')}</h1>
        </div>
        <Link
          to="/developers/webhooks/new"
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('hooks.add')}
        </Link>
      </div>

      {webhooks.length === 0 && !isLoading && (
        <div className="text-center py-12 bg-card rounded-xl border border-border">
          <Network className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-semibold text-foreground">{t('hooks.no_webhooks_title')}</h3>
          <p className="mt-1 text-sm text-muted-foreground">{t('hooks.no_webhooks_desc')}</p>
          <div className="mt-6">
            <Link to="/developers/webhooks/new" className="inline-flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary/90">
              {t('hooks.add_endpoint')}
            </Link>
          </div>
        </div>
      )}

      {webhooks.length > 0 && (
        <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border">
              <thead className="bg-muted">
                <tr>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('hooks.url')}</th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('hooks.events')}</th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('hooks.secret')}</th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('common.status')}</th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('hooks.last_triggered')}</th>
                  <th scope="col" className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
                </tr>
              </thead>
              <tbody className="bg-card divide-y divide-border">
                {webhooks.map((webhook) => (
                  <tr key={webhook.id} className="hover:bg-accent transition-colors">
                    <td className="px-6 py-4">
                      <div className="flex flex-col">
                        <span className="font-medium text-foreground font-mono text-sm truncate max-w-xs" title={webhook.url}>{webhook.url}</span>
                        <span className="text-xs text-muted-foreground mt-0.5">{webhook.description}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1 max-w-xs">
                         {webhook.events?.slice(0, 2).map(e => (
                           <span key={e} className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary border border-primary/20">
                             {e}
                           </span>
                         ))}
                         {(webhook.events?.length || 0) > 2 && (
                           <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-muted text-muted-foreground">
                             +{(webhook.events?.length || 0) - 2}
                           </span>
                         )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <code className="text-xs font-mono bg-muted px-2 py-1 rounded border border-border text-muted-foreground w-24 truncate">
                          {revealedSecrets.includes(webhook.id) ? (webhook.secret || 'N/A') : 'whsec_••••••••'}
                        </code>
                        <button
                          onClick={() => toggleSecret(webhook.id)}
                          className="text-muted-foreground hover:text-muted-foreground"
                        >
                          {revealedSecrets.includes(webhook.id) ? <EyeOff size={14} /> : <Eye size={14} />}
                        </button>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                         {webhook.is_active
                          ? <span className="flex items-center text-success text-xs font-medium bg-success/10 px-2 py-1 rounded-full"><CheckCircle size={12} className="mr-1"/> {t('users.active')}</span>
                          : <span className="flex items-center text-muted-foreground text-xs font-medium bg-muted px-2 py-1 rounded-full"><XCircle size={12} className="mr-1"/> {t('hooks.disabled')}</span>
                         }
                         {(webhook.failure_count || 0) > 0 && (
                           <span className="flex items-center text-destructive text-xs font-medium bg-destructive/10 px-2 py-1 rounded-full" title={`${webhook.failure_count} recent failures`}>
                              <Activity size={12} className="mr-1"/> {webhook.failure_count} {t('hooks.recent_failures')}
                           </span>
                         )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                      {webhook.last_triggered_at ? formatRelative(webhook.last_triggered_at) : '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex justify-end gap-2">
                        <Link
                          to={`/developers/webhooks/${webhook.id}`}
                          className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                          title={t('common.edit')}
                        >
                          <Edit2 size={18} />
                        </Link>
                        <button
                          onClick={() => handleDelete(webhook.id)}
                          disabled={deleteWebhookMutation.isPending}
                          className="p-1 text-muted-foreground hover:text-destructive rounded-md hover:bg-accent disabled:opacity-50"
                          title={t('common.delete')}
                        >
                          <Trash2 size={18} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};

export default Webhooks;
