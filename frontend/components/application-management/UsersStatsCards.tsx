import React from 'react';
import { Ban, UserCheck, Users } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface UsersStatsCardsProps {
  total: number;
  activeCount: number;
  bannedCount: number;
}

const UsersStatsCards: React.FC<UsersStatsCardsProps> = ({ total, activeCount, bannedCount }) => {
  const { t } = useLanguage();

  return (
    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <div className="bg-card rounded-xl shadow-sm border border-border p-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
            <Users className="text-primary" size={20} />
          </div>
          <div>
            <p className="text-2xl font-bold text-foreground">{total}</p>
            <p className="text-sm text-muted-foreground">{t('apps.total_users')}</p>
          </div>
        </div>
      </div>
      <div className="bg-card rounded-xl shadow-sm border border-border p-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-success/10 flex items-center justify-center">
            <UserCheck className="text-success" size={20} />
          </div>
          <div>
            <p className="text-2xl font-bold text-foreground">{activeCount}</p>
            <p className="text-sm text-muted-foreground">{t('apps.active_users')}</p>
          </div>
        </div>
      </div>
      <div className="bg-card rounded-xl shadow-sm border border-border p-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-destructive/10 flex items-center justify-center">
            <Ban className="text-destructive" size={20} />
          </div>
          <div>
            <p className="text-2xl font-bold text-foreground">{bannedCount}</p>
            <p className="text-sm text-muted-foreground">{t('apps.banned_users')}</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default UsersStatsCards;
