
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
          <h1 className="text-2xl font-bold text-gray-900">{t('hooks.title')}</h1>
        </div>
        <Link 
          to="/developers/webhooks/new"
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('hooks.add')}
        </Link>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('hooks.url')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('hooks.events')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('hooks.secret')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('common.status')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Triggered</th>
                <th scope="col" className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {webhooks.map((webhook) => (
                <tr key={webhook.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex flex-col">
                      <span className="font-medium text-gray-900 font-mono text-sm truncate max-w-xs" title={webhook.url}>{webhook.url}</span>
                      <span className="text-xs text-gray-500 mt-0.5">{webhook.description}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex flex-wrap gap-1 max-w-xs">
                       {webhook.events.slice(0, 2).map(e => (
                         <span key={e} className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-50 text-blue-700 border border-blue-100">
                           {e}
                         </span>
                       ))}
                       {webhook.events.length > 2 && (
                         <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600">
                           +{webhook.events.length - 2}
                         </span>
                       )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      <code className="text-xs font-mono bg-gray-100 px-2 py-1 rounded border border-gray-200 text-gray-600 w-24 truncate">
                        {revealedSecrets.includes(webhook.id) ? webhook.secret : 'whsec_••••••••'}
                      </code>
                      <button 
                        onClick={() => toggleSecret(webhook.id)}
                        className="text-gray-400 hover:text-gray-600"
                      >
                        {revealedSecrets.includes(webhook.id) ? <EyeOff size={14} /> : <Eye size={14} />}
                      </button>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                       {webhook.isActive 
                        ? <span className="flex items-center text-green-700 text-xs font-medium bg-green-50 px-2 py-1 rounded-full"><CheckCircle size={12} className="mr-1"/> {t('users.active')}</span>
                        : <span className="flex items-center text-gray-600 text-xs font-medium bg-gray-100 px-2 py-1 rounded-full"><XCircle size={12} className="mr-1"/> Disabled</span>
                       }
                       {webhook.failureCount > 0 && (
                         <span className="flex items-center text-red-700 text-xs font-medium bg-red-50 px-2 py-1 rounded-full" title={`${webhook.failureCount} recent failures`}>
                            <Activity size={12} className="mr-1"/> {webhook.failureCount} failed
                         </span>
                       )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {webhook.lastTriggeredAt ? new Date(webhook.lastTriggeredAt).toLocaleString() : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <Link 
                        to={`/developers/webhooks/${webhook.id}`}
                        className="p-1 text-gray-400 hover:text-blue-600 rounded-md hover:bg-gray-100"
                      >
                        <Edit2 size={18} />
                      </Link>
                      <button 
                        onClick={() => handleDelete(webhook.id)}
                        className="p-1 text-gray-400 hover:text-red-600 rounded-md hover:bg-gray-100"
                      >
                        <Trash2 size={18} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {webhooks.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center text-gray-500">
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
