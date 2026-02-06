import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, CheckCircle, XCircle, Clock, AlertCircle } from 'lucide-react';
import { useLDAPSyncLogs } from '../hooks/useLDAP';
import { formatDateTime } from '../lib/date';

const LDAPSyncLogs: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { data, isLoading, error } = useLDAPSyncLogs(id || '');

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
        <p className="text-destructive">Error loading sync logs: {(error as Error).message}</p>
        <Link to="/ldap" className="text-primary hover:underline mt-4 inline-block">
          Back to LDAP Configurations
        </Link>
      </div>
    );
  }

  const logs = data?.logs || [];

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="text-success" size={20} />;
      case 'failed':
        return <XCircle className="text-destructive" size={20} />;
      case 'partial':
        return <AlertCircle className="text-warning" size={20} />;
      default:
        return <Clock className="text-muted-foreground" size={20} />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'success':
        return 'bg-success/10 text-success';
      case 'failed':
        return 'bg-destructive/10 text-destructive';
      case 'partial':
        return 'bg-warning/10 text-warning';
      default:
        return 'bg-muted text-foreground';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          to="/ldap"
          className="text-muted-foreground hover:text-foreground flex items-center gap-2"
        >
          <ArrowLeft size={20} />
          Back
        </Link>
        <h1 className="text-2xl font-bold text-foreground">LDAP Sync Logs</h1>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Date
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Users
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Groups
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Duration
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Error
                </th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {logs.map((log) => (
                <tr key={log.id} className="hover:bg-accent transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-foreground">
                      {formatDateTime(log.started_at)}
                    </div>
                    {log.completed_at && (
                      <div className="text-xs text-muted-foreground">
                        Completed: {formatDateTime(log.completed_at)}
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
                    <div className="text-sm text-foreground">
                      <div>Synced: {log.users_synced}</div>
                      <div className="text-xs text-muted-foreground">
                        Created: {log.users_created} | Updated: {log.users_updated} | Deleted: {log.users_deleted}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-foreground">
                      <div>Synced: {log.groups_synced}</div>
                      <div className="text-xs text-muted-foreground">
                        Created: {log.groups_created} | Updated: {log.groups_updated}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {log.duration_ms ? `${(log.duration_ms / 1000).toFixed(2)}s` : '-'}
                  </td>
                  <td className="px-6 py-4">
                    {log.error_message ? (
                      <div className="text-sm text-destructive max-w-xs truncate" title={log.error_message}>
                        {log.error_message}
                      </div>
                    ) : (
                      <span className="text-sm text-muted-foreground">-</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {logs.length === 0 && (
            <div className="p-12 text-center text-muted-foreground">No sync logs found.</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default LDAPSyncLogs;

