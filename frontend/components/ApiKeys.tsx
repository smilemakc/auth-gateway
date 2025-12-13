
import React, { useState } from 'react';
import { mockApiKeys } from '../services/mockData';
import { Trash2, Ban, Copy, CheckCircle } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const ApiKeys: React.FC = () => {
  const [keys, setKeys] = useState(mockApiKeys);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const { t } = useLanguage();

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleRevoke = (id: string) => {
    if(window.confirm(t('keys.revoke_confirm'))) {
      setKeys(prev => prev.map(k => k.id === id ? { ...k, status: 'revoked' } : k));
    }
  };

  const handleDelete = (id: string) => {
    if(window.confirm(t('common.confirm_delete'))) {
      setKeys(prev => prev.filter(k => k.id !== id));
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h1 className="text-2xl font-bold text-gray-900">{t('keys.title')}</h1>
        <button className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors">
          + {t('keys.generate')}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {keys.map((apiKey) => (
          <div key={apiKey.id} className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 flex flex-col">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <div className={`p-2 rounded-lg ${apiKey.status === 'active' ? 'bg-amber-50 text-amber-600' : 'bg-gray-100 text-gray-500'}`}>
                  <div className="font-mono text-xl font-bold">K</div>
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">{apiKey.name}</h3>
                  <p className="text-sm text-gray-500">{t('keys.owner')}: {apiKey.ownerName}</p>
                </div>
              </div>
              <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                apiKey.status === 'active' 
                  ? 'bg-green-100 text-green-700' 
                  : 'bg-red-100 text-red-700'
              }`}>
                {apiKey.status === 'active' ? t('users.active') : t('keys.revoked')}
              </span>
            </div>
            
            <div className="bg-gray-50 rounded-md p-3 mb-4 flex items-center justify-between group">
              <code className="text-sm text-gray-600 font-mono">
                {apiKey.prefix}****************
              </code>
              <button 
                onClick={() => handleCopy(apiKey.prefix, apiKey.id)}
                className="text-gray-400 hover:text-blue-600 transition-colors"
              >
                {copiedId === apiKey.id ? <CheckCircle size={16} className="text-green-500" /> : <Copy size={16} />}
              </button>
            </div>

            <div className="flex flex-wrap gap-2 mb-6">
              {apiKey.scopes.map(scope => (
                <span key={scope} className="px-2 py-1 bg-blue-50 text-blue-700 text-xs rounded border border-blue-100">
                  {scope}
                </span>
              ))}
            </div>

            <div className="mt-auto pt-4 border-t border-gray-100 flex items-center justify-between text-sm">
               <span className="text-gray-500">{t('common.created')}: {new Date(apiKey.createdAt).toLocaleDateString()}</span>
               
               <div className="flex gap-2">
                 {apiKey.status === 'active' && (
                   <button 
                    onClick={() => handleRevoke(apiKey.id)}
                    className="p-1.5 text-gray-500 hover:text-orange-600 hover:bg-orange-50 rounded transition-colors" title={t('user.revoke')}>
                     <Ban size={18} />
                   </button>
                 )}
                 <button 
                   onClick={() => handleDelete(apiKey.id)}
                   className="p-1.5 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded transition-colors" title={t('common.delete')}>
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
