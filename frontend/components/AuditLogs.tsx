
import React from 'react';
import { Clock, Shield, Globe, User } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useAuditLogs } from '../hooks/useAuditLogs';

const AuditLogs: React.FC = () => {
  const { t } = useLanguage();

  // Fetch audit logs with React Query
  const { data, isLoading, error } = useAuditLogs(1, 100);

  const logs = data?.logs || data?.items || [];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive">{t('common.error_loading')}: {(error as Error).message}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">{t('nav.audit_logs')}</h1>

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full text-left text-sm whitespace-nowrap">
            <thead className="uppercase tracking-wider border-b border-border bg-muted">
              <tr>
                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('common.actions')}</th>
                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('users.col_user')}</th>
                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('ip.address')}</th>
                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('common.status')}</th>
                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('audit.col_time')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {logs.map((log: any) => (
                <tr key={log.id} className="hover:bg-accent">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className={`p-2 rounded-lg bg-muted`}>
                        <Shield size={16} className="text-muted-foreground" />
                      </div>
                      <span className="font-medium text-foreground capitalize">{log.action?.replace(/_/g, ' ')}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <User size={16} />
                      {log.user_id || t('audit.system')}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2 text-muted-foreground font-mono text-xs">
                      <Globe size={14} />
                      {log.ip_address || '-'}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize
                      ${log.status === 'success' ? 'bg-success/10 text-success' :
                        log.status === 'blocked' ? 'bg-muted text-muted-foreground' : 'bg-destructive/10 text-destructive'}`}>
                      {log.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-muted-foreground">
                    <div className="flex items-center gap-2">
                      <Clock size={14} />
                      {new Date(log.created_at).toLocaleString()}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="p-4 border-t border-border flex justify-between items-center text-sm text-muted-foreground">
          <span>{t('common.showing')} {logs.length} {t('audit.showing')}</span>
        </div>
      </div>
    </div>
  );
};

export default AuditLogs;
