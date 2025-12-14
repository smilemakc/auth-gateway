
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
        <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <p className="text-red-600">Error loading audit logs: {(error as Error).message}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">{t('nav.audit_logs')}</h1>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full text-left text-sm whitespace-nowrap">
            <thead className="uppercase tracking-wider border-b border-gray-200 bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-4 font-semibold text-gray-700">{t('common.actions')}</th>
                <th scope="col" className="px-6 py-4 font-semibold text-gray-700">{t('users.col_user')}</th>
                <th scope="col" className="px-6 py-4 font-semibold text-gray-700">{t('ip.address')}</th>
                <th scope="col" className="px-6 py-4 font-semibold text-gray-700">{t('common.status')}</th>
                <th scope="col" className="px-6 py-4 font-semibold text-gray-700">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {logs.map((log: any) => (
                <tr key={log.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className={`p-2 rounded-lg bg-gray-100`}>
                        <Shield size={16} className="text-gray-600" />
                      </div>
                      <span className="font-medium text-gray-900 capitalize">{log.action?.replace(/_/g, ' ')}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2 text-gray-600">
                      <User size={16} />
                      {log.user_id || 'System'}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2 text-gray-500 font-mono text-xs">
                      <Globe size={14} />
                      {log.ip_address || '-'}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize
                      ${log.status === 'success' ? 'bg-green-100 text-green-800' :
                        log.status === 'blocked' ? 'bg-gray-100 text-gray-800' : 'bg-red-100 text-red-800'}`}>
                      {log.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-500">
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
        <div className="p-4 border-t border-gray-200 flex justify-between items-center text-sm text-gray-500">
          <span>Showing {logs.length} entries</span>
        </div>
      </div>
    </div>
  );
};

export default AuditLogs;
