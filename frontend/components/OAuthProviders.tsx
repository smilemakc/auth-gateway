
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getOAuthProviders, updateOAuthProvider, deleteOAuthProvider } from '../services/mockData';
import { OAuthProviderConfig } from '../types';
import { Plus, Edit2, Trash2, ToggleLeft, ToggleRight, Globe, Github, Send, Instagram } from 'lucide-react';
import { useLanguage } from '../services/i18n';

// Icon mapper
const getProviderIcon = (provider: string) => {
  switch (provider.toLowerCase()) {
    case 'google': return <span className="font-bold text-lg text-red-500">G</span>;
    case 'github': return <Github className="text-gray-900" size={24} />;
    case 'yandex': return <span className="font-bold text-lg text-red-600">Y</span>;
    case 'telegram': return <Send className="text-blue-500" size={24} />;
    case 'instagram': return <Instagram className="text-pink-600" size={24} />;
    default: return <Globe className="text-gray-500" size={24} />;
  }
};

const OAuthProviders: React.FC = () => {
  const [providers, setProviders] = useState<OAuthProviderConfig[]>([]);
  const { t } = useLanguage();

  useEffect(() => {
    setProviders(getOAuthProviders());
  }, []);

  const handleToggle = (id: string, currentStatus: boolean) => {
    const updated = updateOAuthProvider(id, { is_enabled: !currentStatus });
    if (updated) {
      setProviders(prev => prev.map(p => p.id === id ? updated : p));
    }
  };

  const handleDelete = (id: string) => {
    if (window.confirm(t('common.confirm_delete'))) {
      if (deleteOAuthProvider(id)) {
        setProviders(prev => prev.filter(p => p.id !== id));
      }
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{t('oauth.title')}</h1>
          <p className="text-gray-500 mt-1">{t('oauth.manage_desc')}</p>
        </div>
        <Link 
          to="/oauth/new"
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('oauth.add')}
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {providers.map((provider) => (
          <div key={provider.id} className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden flex flex-col">
            <div className="p-6 flex-1">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 rounded-xl bg-gray-50 flex items-center justify-center shadow-sm">
                    {getProviderIcon(provider.provider)}
                  </div>
                  <div>
                    <h3 className="font-semibold text-gray-900 capitalize text-lg">{provider.provider}</h3>
                    <div className="flex items-center gap-2 mt-1">
                      <span className={`w-2 h-2 rounded-full ${provider.is_enabled ? 'bg-green-500' : 'bg-gray-300'}`}></span>
                      <span className="text-xs text-gray-500 font-medium uppercase tracking-wide">
                        {provider.is_enabled ? 'Enabled' : 'Disabled'}
                      </span>
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => handleToggle(provider.id, provider.is_enabled)}
                  className={`transition-colors ${provider.is_enabled ? 'text-green-600 hover:text-green-700' : 'text-gray-300 hover:text-gray-400'}`}
                >
                  {provider.is_enabled ? <ToggleRight size={36} /> : <ToggleLeft size={36} />}
                </button>
              </div>

              <div className="space-y-3 mt-6">
                <div>
                  <label className="text-xs font-semibold text-gray-400 uppercase tracking-wider block mb-1">{t('oauth.client_id')}</label>
                  <code className="block bg-gray-50 rounded px-3 py-2 text-sm text-gray-600 font-mono truncate border border-gray-100">
                    {provider.client_id}
                  </code>
                </div>
                <div>
                  <label className="text-xs font-semibold text-gray-400 uppercase tracking-wider block mb-1">Callback URL</label>
                  <div className="text-xs text-gray-500 truncate" title={provider.redirect_uris[0]}>
                    {provider.redirect_uris[0] || <span className="italic text-gray-400">Not configured</span>}
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-gray-50 px-6 py-4 border-t border-gray-100 flex items-center justify-between">
              <span className="text-xs text-gray-400">
                {new Date(provider.created_at).toLocaleDateString()}
              </span>
              <div className="flex items-center gap-2">
                <Link 
                  to={`/oauth/${provider.id}`}
                  className="p-2 text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                >
                  <Edit2 size={18} />
                </Link>
                <button 
                  onClick={() => handleDelete(provider.id)}
                  className="p-2 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                >
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

export default OAuthProviders;
