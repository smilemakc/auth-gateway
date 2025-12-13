
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getIpRules, createIpRule, deleteIpRule } from '../services/mockData';
import { IpRule } from '../types';
import { ArrowLeft, ShieldAlert, ShieldCheck, Plus, Trash2, Search, Info } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const IpSecurity: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [activeTab, setActiveTab] = useState<'blacklist' | 'whitelist'>('blacklist');
  const [rules, setRules] = useState<IpRule[]>([]);
  const [showAddModal, setShowAddModal] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

  // Form state
  const [newIp, setNewIp] = useState('');
  const [newDescription, setNewDescription] = useState('');

  useEffect(() => {
    setRules(getIpRules(activeTab));
  }, [activeTab]);

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newIp) return;

    createIpRule({
      type: activeTab,
      ipAddress: newIp,
      description: newDescription
    });

    setRules(getIpRules(activeTab));
    setShowAddModal(false);
    setNewIp('');
    setNewDescription('');
  };

  const handleDelete = (id: string) => {
    if (window.confirm(t('common.confirm_delete'))) {
      deleteIpRule(id);
      setRules(getIpRules(activeTab));
    }
  };

  const filteredRules = rules.filter(r => 
    r.ipAddress.includes(searchTerm) || 
    r.description?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <button 
            onClick={() => navigate('/settings')}
            className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{t('ip.title')}</h1>
            <p className="text-gray-500 mt-1">{t('settings.ip_desc')}</p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-1 flex">
        <button
          onClick={() => setActiveTab('blacklist')}
          className={`flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-medium transition-all ${
            activeTab === 'blacklist' 
              ? 'bg-red-50 text-red-700 shadow-sm' 
              : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          <ShieldAlert size={18} />
          {t('ip.blacklist')}
        </button>
        <button
          onClick={() => setActiveTab('whitelist')}
          className={`flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-medium transition-all ${
            activeTab === 'whitelist' 
              ? 'bg-green-50 text-green-700 shadow-sm' 
              : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          <ShieldCheck size={18} />
          {t('ip.whitelist')}
        </button>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        {/* Actions Bar */}
        <div className="p-4 border-b border-gray-100 flex flex-col sm:flex-row gap-4 justify-between items-center">
          <div className="relative w-full sm:w-64">
             <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={18} />
             <input 
               type="text" 
               placeholder={t('common.search')}
               value={searchTerm}
               onChange={(e) => setSearchTerm(e.target.value)}
               className="w-full pl-9 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
             />
          </div>
          <button 
            onClick={() => setShowAddModal(true)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors
              ${activeTab === 'blacklist' ? 'bg-red-600 hover:bg-red-700' : 'bg-green-600 hover:bg-green-700'}`}
          >
            <Plus size={18} />
            {activeTab === 'blacklist' ? t('ip.add_block') : t('ip.add_allow')}
          </button>
        </div>

        {/* List */}
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('ip.address')}</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('ip.added_by')}</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{t('common.created')}</th>
                <th className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredRules.length > 0 ? (
                filteredRules.map((rule) => (
                  <tr key={rule.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap font-mono text-sm text-gray-900">
                      {rule.ipAddress}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {rule.description || <span className="text-gray-400 italic">No description</span>}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {rule.createdBy}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(rule.createdAt).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button 
                        onClick={() => handleDelete(rule.id)}
                        className="text-gray-400 hover:text-red-600 transition-colors"
                      >
                        <Trash2 size={18} />
                      </button>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-gray-500">
                    <div className="flex flex-col items-center justify-center gap-2">
                       {activeTab === 'blacklist' ? <ShieldAlert size={32} className="text-gray-300" /> : <ShieldCheck size={32} className="text-gray-300" />}
                       <p>No rules found</p>
                    </div>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Info Card */}
      <div className="bg-blue-50 rounded-lg p-4 border border-blue-100 flex gap-3">
        <Info className="text-blue-600 flex-shrink-0" size={20} />
        <div className="text-sm text-blue-800">
          <p className="font-semibold mb-1">How IP Filtering Works</p>
          <p>
            The <strong>Whitelist</strong> takes precedence. If a whitelist exists, only IPs in that list (or matching CIDR ranges) can access the system. 
            The <strong>Blacklist</strong> is checked next. Any IP in the blacklist is blocked, even if no whitelist is defined.
          </p>
        </div>
      </div>

      {/* Add Modal */}
      {showAddModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full overflow-hidden animate-in fade-in zoom-in duration-200">
            <div className={`px-6 py-4 border-b flex items-center gap-3 ${activeTab === 'blacklist' ? 'bg-red-50 border-red-100' : 'bg-green-50 border-green-100'}`}>
               {activeTab === 'blacklist' ? <ShieldAlert className="text-red-600" /> : <ShieldCheck className="text-green-600" />}
               <h3 className={`font-semibold ${activeTab === 'blacklist' ? 'text-red-900' : 'text-green-900'}`}>
                 Add to {activeTab === 'blacklist' ? 'Blacklist' : 'Whitelist'}
               </h3>
            </div>
            
            <form onSubmit={handleAdd} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">{t('ip.address')}</label>
                <input 
                  type="text" 
                  value={newIp}
                  onChange={(e) => setNewIp(e.target.value)}
                  placeholder="e.g. 192.168.1.1 or 10.0.0.0/24"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
                  required
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                <input 
                  type="text" 
                  value={newDescription}
                  onChange={(e) => setNewDescription(e.target.value)}
                  placeholder="e.g. Malicious botnet / Office VPN"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
                />
              </div>

              <div className="flex gap-3 pt-4">
                <button 
                  type="button" 
                  onClick={() => setShowAddModal(false)}
                  className="flex-1 px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg font-medium"
                >
                  {t('common.cancel')}
                </button>
                <button 
                  type="submit"
                  className={`flex-1 px-4 py-2 text-white rounded-lg font-medium
                    ${activeTab === 'blacklist' ? 'bg-red-600 hover:bg-red-700' : 'bg-green-600 hover:bg-green-700'}`}
                >
                  {t('common.create')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default IpSecurity;
