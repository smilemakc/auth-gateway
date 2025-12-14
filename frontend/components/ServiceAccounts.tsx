
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
          <h1 className="text-2xl font-bold text-gray-900">{t('sa.title')}</h1>
          <p className="text-gray-500 mt-1">{t('sa.desc')}</p>
        </div>
        <Link 
          to="/developers/service-accounts/new"
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('sa.create')}
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {accounts.map((account) => (
          <div key={account.id} className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 flex flex-col">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-indigo-50 text-indigo-600">
                  <Bot size={24} />
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">{account.name}</h3>
                  <div className="flex items-center gap-2 mt-0.5">
                    {account.is_active
                      ? <span className="text-green-600 text-xs font-medium flex items-center"><CheckCircle size={10} className="mr-1"/> {t('users.active')}</span>
                      : <span className="text-gray-500 text-xs font-medium flex items-center"><XCircle size={10} className="mr-1"/> Disabled</span>
                    }
                  </div>
                </div>
              </div>
            </div>

            <p className="text-sm text-gray-500 mb-4 line-clamp-2 h-10">
              {account.description || 'No description provided.'}
            </p>

            <div className="bg-gray-50 rounded-lg p-3 mb-6 border border-gray-100">
              <label className="text-xs font-semibold text-gray-500 uppercase tracking-wider block mb-1">Client ID</label>
              <div className="flex items-center justify-between">
                <code className="text-sm font-mono text-gray-800 truncate mr-2">{account.client_id}</code>
                <button
                  onClick={() => copyToClipboard(account.client_id, account.id)}
                  className="text-gray-400 hover:text-blue-600"
                >
                  {copiedId === account.id ? <CheckCircle size={16} className="text-green-500"/> : <Copy size={16}/>}
                </button>
              </div>
            </div>

            <div className="mt-auto flex items-center justify-between pt-4 border-t border-gray-100">
               <span className="text-xs text-gray-400">{new Date(account.created_at).toLocaleDateString()}</span>
               <div className="flex gap-2">
                 <Link 
                   to={`/developers/service-accounts/${account.id}`}
                   className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                 >
                   <Edit2 size={18} />
                 </Link>
                 <button 
                   onClick={() => handleDelete(account.id)}
                   className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                 >
                   <Trash2 size={18} />
                 </button>
               </div>
            </div>
          </div>
        ))}
        {accounts.length === 0 && (
           <div className="col-span-full py-12 text-center text-gray-500 bg-white rounded-xl border border-gray-100 border-dashed">
            <Bot size={48} className="mx-auto text-gray-300 mb-3" />
            <h3 className="text-lg font-medium text-gray-900">No Service Accounts</h3>
          </div>
        )}
      </div>
    </div>
  );
};

export default ServiceAccounts;
