
import React, { useState } from 'react';
import { Clock, Shield, Globe, User, ShieldAlert } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useAuditLogs } from '../hooks/useAuditLogs';
import { formatDateTime } from '../lib/date';

const AuditLogs: React.FC = () => {
  const { t } = useLanguage();

  const [page, setPage] = useState(1);
  const pageSize = 50;
  const [actionFilter, setActionFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');

  // Fetch audit logs with React Query
  const { data, isLoading, error } = useAuditLogs(page, pageSize, {
    action: actionFilter || undefined,
    status: (statusFilter as 'success' | 'failure') || undefined,
  });

  const logs = data?.logs || [];
  const total = data?.total || 0;
  const totalPages = Math.ceil(total / pageSize);

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

      <div className="flex flex-wrap gap-3 mb-4">
        <select
          value={actionFilter}
          onChange={(e) => { setActionFilter(e.target.value); setPage(1); }}
          className="border border-input rounded-lg px-3 py-2 text-sm bg-card text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        >
          <option value="">{t('audit.filter_action')}</option>
          <option value="signin">Signin</option>
          <option value="signup">Signup</option>
          <option value="logout">Logout</option>
          <option value="refresh_token">Refresh Token</option>
          <option value="password_reset">Password Reset</option>
          <option value="password_change">Password Change</option>
          <option value="2fa_enable">2FA Enable</option>
          <option value="2fa_disable">2FA Disable</option>
        </select>
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
          className="border border-input rounded-lg px-3 py-2 text-sm bg-card text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        >
          <option value="">{t('audit.filter_status')}</option>
          <option value="success">Success</option>
          <option value="failure">Failure</option>
        </select>
      </div>

      {logs.length === 0 && !isLoading && (
        <div className="text-center py-12 bg-card rounded-xl border border-border">
          <ShieldAlert className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-semibold text-foreground">{t('audit.no_logs')}</h3>
          <p className="mt-1 text-sm text-muted-foreground">{t('audit.no_logs_desc')}</p>
        </div>
      )}

      {logs.length > 0 && (
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
                      {log.user_email || log.user_id || t('audit.system')}
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
                      {formatDateTime(log.created_at)}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-between bg-card px-4 py-3 rounded-lg border border-border mt-4">
          <div className="text-sm text-muted-foreground">
            {t('common.showing')} <span className="font-medium">{(page - 1) * pageSize + 1}</span> {t('common.to')}{' '}
            <span className="font-medium">{Math.min(page * pageSize, total)}</span> {t('common.of')}{' '}
            <span className="font-medium">{total}</span>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.previous')}
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.next')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default AuditLogs;
