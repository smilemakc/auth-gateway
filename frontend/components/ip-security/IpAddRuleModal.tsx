import React, { useState } from 'react';
import { ShieldAlert, ShieldCheck, Loader2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface IpAddRuleModalProps {
  activeTab: 'blacklist' | 'whitelist';
  isCreating: boolean;
  onAdd: (ip: string, description: string) => Promise<void>;
  onClose: () => void;
}

export const IpAddRuleModal: React.FC<IpAddRuleModalProps> = ({
  activeTab,
  isCreating,
  onAdd,
  onClose,
}) => {
  const { t } = useLanguage();
  const [newIp, setNewIp] = useState('');
  const [newDescription, setNewDescription] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newIp) return;
    await onAdd(newIp, newDescription);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50">
      <div className="bg-card rounded-xl shadow-xl max-w-md w-full overflow-hidden animate-in fade-in zoom-in duration-200">
        <div className={`px-6 py-4 border-b flex items-center gap-3 ${activeTab === 'blacklist' ? 'bg-destructive/10 border-destructive/20' : 'bg-success/10 border-success/20'}`}>
           {activeTab === 'blacklist' ? <ShieldAlert className="text-destructive" /> : <ShieldCheck className="text-success" />}
           <h3 className={`font-semibold ${activeTab === 'blacklist' ? 'text-destructive' : 'text-success'}`}>
             {t('ip.add_to')} {activeTab === 'blacklist' ? t('ip.blacklist') : t('ip.whitelist')}
           </h3>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('ip.address')}</label>
            <input
              type="text"
              value={newIp}
              onChange={(e) => setNewIp(e.target.value)}
              placeholder={t('ip.example_ip')}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('common.description')}</label>
            <input
              type="text"
              value={newDescription}
              onChange={(e) => setNewDescription(e.target.value)}
              placeholder={t('ip.example_desc')}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
            />
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 text-foreground bg-muted hover:bg-accent rounded-lg font-medium"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={isCreating}
              className={`flex-1 px-4 py-2 text-primary-foreground rounded-lg font-medium disabled:opacity-50
                ${activeTab === 'blacklist' ? 'bg-destructive hover:bg-destructive/90' : 'bg-success hover:bg-success/90'}`}
            >
              {isCreating ? (
                <Loader2 size={16} className="mx-auto animate-spin" />
              ) : (
                t('common.create')
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
