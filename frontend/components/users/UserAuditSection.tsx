import React from 'react';
import { Activity } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { formatRelative } from '../../lib/date';

interface AuditLog {
  id: string;
  action: string;
  status: string;
  created_at: string;
}

interface UserAuditSectionProps {
  logs: AuditLog[];
}

const UserAuditSection: React.FC<UserAuditSectionProps> = ({ logs }) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="p-6 border-b border-border">
        <h3 className="font-semibold text-foreground flex items-center gap-2">
          <Activity size={18} className="text-primary" />
          {t('user.recent_activity')}
        </h3>
      </div>
      <div className="divide-y divide-border">
        {logs.length > 0 ? (
          logs.map(log => (
            <div key={log.id} className="p-4 flex items-center justify-between hover:bg-accent">
              <div className="flex items-center gap-4">
                <div className={`p-2 rounded-full ${
                  log.status === 'success' ? 'bg-success/20 text-success' :
                  log.status === 'failure' ? 'bg-destructive/20 text-destructive' : 'bg-muted text-muted-foreground'
                }`}>
                  <Activity size={16} />
                </div>
                <div>
                  <p className="text-sm font-medium text-foreground capitalize">{log.action.replace(/_/g, ' ')}</p>
                  <p className="text-xs text-muted-foreground">{formatRelative(log.created_at)}</p>
                </div>
              </div>
              <div className="text-right">
                 <span className={`text-xs font-medium px-2 py-1 rounded-full capitalize ${
                   log.status === 'success' ? 'text-success bg-success/10' : 'text-destructive bg-destructive/10'
                 }`}>
                   {log.status}
                 </span>
              </div>
            </div>
          ))
        ) : (
          <div className="p-6 text-sm text-muted-foreground italic">{t('user.no_activity')}</div>
        )}
      </div>
    </div>
  );
};

export default UserAuditSection;
