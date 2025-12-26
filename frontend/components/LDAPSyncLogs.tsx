import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, CheckCircle, XCircle, Clock, AlertCircle } from 'lucide-react';
import { useLDAPSyncLogs } from '../hooks/useLDAP';

const LDAPSyncLogs: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { data, isLoading, error } = useLDAPSyncLogs(id || '');

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
        <p className="text-red-600">Error loading sync logs: {(error as Error).message}</p>
        <Link to="/ldap" className="text-blue-600 hover:underline mt-4 inline-block">
          Back to LDAP Configurations
        </Link>
      </div>
    );
  }

  const logs = data?.logs || [];

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="text-green-600" size={20} />;
      case 'failed':
        return <XCircle className="text-red-600" size={20} />;
      case 'partial':
        return <AlertCircle className="text-yellow-600" size={20} />;
      default:
        return <Clock className="text-gray-600" size={20} />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'bg-green-100 text-green-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      case 'partial':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          to="/ldap"
          className="text-gray-500 hover:text-gray-700 flex items-center gap-2"
        >
          <ArrowLeft size={20} />
          Back
        </Link>
        <h1 className="text-2xl font-bold text-gray-900">LDAP Sync Logs</h1>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Date
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Users
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Groups
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Duration
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Error
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {logs.map((log) => (
                <tr key={log.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">
                      {new Date(log.started_at).toLocaleString()}
                    </div>
                    {log.completed_at && (
                      <div className="text-xs text-gray-500">
                        Completed: {new Date(log.completed_at).toLocaleString()}
                      </div>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(log.status)}
                      <span
                        className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(
                          log.status
                        )}`}
                      >
                        {log.status}
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">
                      <div>Synced: {log.users_synced}</div>
                      <div className="text-xs text-gray-500">
                        Created: {log.users_created} | Updated: {log.users_updated} | Deleted: {log.users_deleted}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-gray-900">
                      <div>Synced: {log.groups_synced}</div>
                      <div className="text-xs text-gray-500">
                        Created: {log.groups_created} | Updated: {log.groups_updated}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {log.duration_ms ? `${(log.duration_ms / 1000).toFixed(2)}s` : '-'}
                  </td>
                  <td className="px-6 py-4">
                    {log.error_message ? (
                      <div className="text-sm text-red-600 max-w-xs truncate" title={log.error_message}>
                        {log.error_message}
                      </div>
                    ) : (
                      <span className="text-sm text-gray-400">-</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {logs.length === 0 && (
            <div className="p-12 text-center text-gray-500">No sync logs found.</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default LDAPSyncLogs;

