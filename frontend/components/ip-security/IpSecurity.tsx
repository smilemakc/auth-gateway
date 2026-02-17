
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, ShieldAlert, ShieldCheck, Plus, Trash2, Search, Info, Loader2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useWhitelistFilters, useBlacklistFilters, useCreateIpFilter, useDeleteIpFilter } from '../../hooks/useIpFilters';
import { formatDate } from '../../lib/date';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';
import { IpAddRuleModal } from './IpAddRuleModal';

const IpSecurity: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [activeTab, setActiveTab] = useState<'blacklist' | 'whitelist'>('blacklist');
  const [showAddModal, setShowAddModal] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

  const { data: blacklistResponse, isLoading: blacklistLoading } = useBlacklistFilters();
  const { data: whitelistResponse, isLoading: whitelistLoading } = useWhitelistFilters();
  const createFilterMutation = useCreateIpFilter();
  const deleteFilterMutation = useDeleteIpFilter();

  const rules = activeTab === 'blacklist'
    ? (blacklistResponse?.filters || [])
    : (whitelistResponse?.filters || []);

  const isLoading = activeTab === 'blacklist' ? blacklistLoading : whitelistLoading;

  const handleAdd = async (ip: string, description: string) => {
    try {
      await createFilterMutation.mutateAsync({
        type: activeTab,
        ip_address: ip,
        description,
      });
      setShowAddModal(false);
    } catch (err) {
      logger.error('Failed to add IP rule:', err);
    }
  };

  const handleDelete = async (id: string) => {
    const ok = await confirm({
      description: t('common.confirm_delete'),
      title: t('confirm.delete_title'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await deleteFilterMutation.mutateAsync(id);
      } catch (err) {
        logger.error('Failed to delete IP rule:', err);
      }
    }
  };

  const filteredRules = rules.filter(r =>
    r.ip_address.includes(searchTerm) ||
    r.description?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/settings')}
            className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-foreground">{t('ip.title')}</h1>
            <p className="text-muted-foreground mt-1">{t('settings.ip_desc')}</p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="bg-card rounded-xl shadow-sm border border-border p-1 flex">
        <button
          onClick={() => setActiveTab('blacklist')}
          className={`flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-medium transition-all ${
            activeTab === 'blacklist'
              ? 'bg-destructive/10 text-destructive shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <ShieldAlert size={18} />
          {t('ip.blacklist')}
        </button>
        <button
          onClick={() => setActiveTab('whitelist')}
          className={`flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-medium transition-all ${
            activeTab === 'whitelist'
              ? 'bg-success/10 text-success shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <ShieldCheck size={18} />
          {t('ip.whitelist')}
        </button>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        {/* Actions Bar */}
        <div className="p-4 border-b border-border flex flex-col sm:flex-row gap-4 justify-between items-center">
          <div className="relative w-full sm:w-64">
             <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" size={18} />
             <input
               type="text"
               placeholder={t('common.search')}
               value={searchTerm}
               onChange={(e) => setSearchTerm(e.target.value)}
               className="w-full pl-9 pr-4 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
             />
          </div>
          <button
            onClick={() => setShowAddModal(true)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-primary-foreground transition-colors
              ${activeTab === 'blacklist' ? 'bg-destructive hover:bg-destructive/90' : 'bg-success hover:bg-success/90'}`}
          >
            <Plus size={18} />
            {activeTab === 'blacklist' ? t('ip.add_block') : t('ip.add_allow')}
          </button>
        </div>

        {/* List */}
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-8 h-8 animate-spin text-primary" />
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border">
              <thead className="bg-muted">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('ip.address')}</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Description</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('ip.added_by')}</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('common.created')}</th>
                  <th className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
                </tr>
              </thead>
              <tbody className="bg-card divide-y divide-border">
                {filteredRules.length > 0 ? (
                  filteredRules.map((rule) => (
                    <tr key={rule.id} className="hover:bg-accent">
                      <td className="px-6 py-4 whitespace-nowrap font-mono text-sm text-foreground">
                        {rule.ip_address}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                        {rule.description || <span className="text-muted-foreground italic">{t('ip.no_description')}</span>}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                        {rule.created_by || '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                        {rule.created_at ? formatDate(rule.created_at) : '-'}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <button
                          onClick={() => handleDelete(rule.id)}
                          disabled={deleteFilterMutation.isPending}
                          className="text-muted-foreground hover:text-destructive transition-colors disabled:opacity-50"
                        >
                          <Trash2 size={18} />
                        </button>
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={5} className="px-6 py-12 text-center text-muted-foreground">
                      <div className="flex flex-col items-center justify-center gap-2">
                         {activeTab === 'blacklist' ? <ShieldAlert size={32} className="text-muted-foreground" /> : <ShieldCheck size={32} className="text-muted-foreground" />}
                         <p>{t('ip.no_rules')}</p>
                      </div>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Info Card */}
      <div className="bg-primary/10 rounded-lg p-4 border border-primary/20 flex gap-3">
        <Info className="text-primary flex-shrink-0" size={20} />
        <div className="text-sm text-primary">
          <p className="font-semibold mb-1">{t('ip.how_it_works')}</p>
          <p>
            {t('ip.how_it_works_desc')}
          </p>
        </div>
      </div>

      {/* Add Modal */}
      {showAddModal && (
        <IpAddRuleModal
          activeTab={activeTab}
          isCreating={createFilterMutation.isPending}
          onAdd={handleAdd}
          onClose={() => setShowAddModal(false)}
        />
      )}
    </div>
  );
};

export default IpSecurity;
