
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getServiceAccounts, deleteServiceAccount } from '../services/mockData';
import { ServiceAccount } from '../types';
import { Plus, Trash2, Edit2, Bot, CheckCircle, XCircle, Copy } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const ServiceAccounts: React.FC = () => {
  const [accounts, setAccounts] = useState<ServiceAccount[]>([]);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const { t } = useLanguage();

  useEffect(() => {
    setAccounts(getServiceAccounts());
  }, []);

  const handleDelete = (id: string) => {
    if (window.confirm(t('common.confirm_delete'))) {
      if (deleteServiceAccount(id)) {
        setAccounts(prev => prev.filter(a => a.id !== id));
      }
    }
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

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
               <span className="text-xs text-muted-foreground">{new Date(account.created_at).toLocaleDateString()}</span>
               <div className="flex gap-2">
                 <Link
                   to={`/developers/service-accounts/${account.id}`}
                   className="p-1.5 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded transition-colors"
                 >
                   <Edit2 size={18} />
                 </Link>
                 <button
                   onClick={() => handleDelete(account.id)}
                   className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded transition-colors"
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
