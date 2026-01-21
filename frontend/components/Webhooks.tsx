
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getWebhooks, deleteWebhook } from '../services/mockData';
import { WebhookEndpoint } from '../types';
import { Plus, Trash2, Edit2, Activity, CheckCircle, XCircle, Eye, EyeOff } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const Webhooks: React.FC = () => {
  const [webhooks, setWebhooks] = useState<WebhookEndpoint[]>([]);
  const [revealedSecrets, setRevealedSecrets] = useState<string[]>([]);
  const { t } = useLanguage();

  useEffect(() => {
    setWebhooks(getWebhooks());
  }, []);

  const handleDelete = (id: string) => {
    if (window.confirm(t('common.confirm_delete'))) {
      if (deleteWebhook(id)) {
        setWebhooks(prev => prev.filter(w => w.id !== id));
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

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('hooks.url')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('hooks.events')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('hooks.secret')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('common.status')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Last Triggered</th>
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
                       {webhook.events.slice(0, 2).map(e => (
                         <span key={e} className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary border border-primary/20">
                           {e}
                         </span>
                       ))}
                       {webhook.events.length > 2 && (
                         <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-muted text-muted-foreground">
                           +{webhook.events.length - 2}
                         </span>
                       )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      <code className="text-xs font-mono bg-muted px-2 py-1 rounded border border-border text-muted-foreground w-24 truncate">
                        {revealedSecrets.includes(webhook.id) ? webhook.secret_key : 'whsec_••••••••'}
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
                        : <span className="flex items-center text-muted-foreground text-xs font-medium bg-muted px-2 py-1 rounded-full"><XCircle size={12} className="mr-1"/> Disabled</span>
                       }
                       {webhook.failure_count > 0 && (
                         <span className="flex items-center text-destructive text-xs font-medium bg-destructive/10 px-2 py-1 rounded-full" title={`${webhook.failure_count} recent failures`}>
                            <Activity size={12} className="mr-1"/> {webhook.failure_count} failed
                         </span>
                       )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {webhook.last_triggered_at ? new Date(webhook.last_triggered_at).toLocaleString() : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <Link 
                        to={`/developers/webhooks/${webhook.id}`}
                        className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                      >
                        <Edit2 size={18} />
                      </Link>
                      <button 
                        onClick={() => handleDelete(webhook.id)}
                        className="p-1 text-muted-foreground hover:text-destructive rounded-md hover:bg-accent"
                      >
                        <Trash2 size={18} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {webhooks.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center text-muted-foreground">
                    No webhooks configured.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default Webhooks;
