import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Plus, Edit, Trash2, TestTube, RefreshCw, FileText, CheckCircle, XCircle, Clock } from 'lucide-react';
import type { LDAPConfig } from '@auth-gateway/client-sdk';
import { useLDAPConfigs, useDeleteLDAPConfig, useSyncLDAP, useTestLDAPConnection } from '../hooks/useLDAP';
import { useLanguage } from '../services/i18n';
import { formatRelative } from '../lib/date';
import { toast } from '../services/toast';
import { confirm } from '../services/confirm';
import { logger } from '@/lib/logger';

const LDAPConfigs: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);

  const { data, isLoading, error } = useLDAPConfigs();
  const deleteConfig = useDeleteLDAPConfig();
  const syncLDAP = useSyncLDAP();
  const testConnection = useTestLDAPConnection();

  const handleDelete = async (id: string, name: string) => {
    const ok = await confirm({
      title: t('confirm.delete_title'),
      description: `${t('ldap.delete_confirm')} "${name}"?`,
      variant: 'danger',
    });
    if (ok) {
      try {
        await deleteConfig.mutateAsync(id);
      } catch (error) {
        logger.error('Failed to delete LDAP config:', error);
        toast.error(t('common.failed_to_load'));
      }
    }
  };

  const handleTestConnection = async (config: LDAPConfig) => {
    setTestingId(config.id);
    try {
      const result = await testConnection.mutateAsync({
        server: config.server,
        port: config.port,
        use_tls: config.use_tls,
        use_ssl: config.use_ssl,
        insecure: config.insecure,
        bind_dn: config.bind_dn,
        bind_password: '', // Password not available, test will use stored password
        base_dn: config.base_dn,
      });
      if (result.success) {
        toast.success(`${t('ldap.connection_success')}\nUsers: ${result.user_count || 0}\nGroups: ${result.group_count || 0}`);
      } else {
        toast.error(`${t('ldap.connection_failed')}: ${result.error || result.message}`);
      }
    } catch (error) {
      toast.error(`${t('ldap.connection_failed')}: ${(error as Error).message}`);
    } finally {
      setTestingId(null);
    }
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    try {
      await syncLDAP.mutateAsync({ id, data: { sync_users: true, sync_groups: true } });
      toast.success(t('ldap.sync_started'));
    } catch (error) {
      logger.error('Failed to start sync:', error);
      toast.error(t('common.failed_to_load'));
    } finally {
      setSyncingId(null);
    }
  };

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
        <p className="text-destructive">{t('ldap.error_loading')}: {(error as Error).message}</p>
      </div>
    );
  }

  const configs = data?.configs || [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('ldap.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('ldap.desc')}</p>
        </div>
        <button
          onClick={() => navigate('/ldap/new')}
          className="bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
        >
          <Plus size={18} />
          {t('ldap.create')}
        </button>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('ldap.col_name_server')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('ldap.col_port')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('ldap.col_base_dn')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('ldap.col_status')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('ldap.col_last_sync')}
                </th>
                <th scope="col" className="relative px-6 py-3">
                  <span className="sr-only">{t('ldap.col_actions')}</span>
                </th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {configs.map((config) => (
                <tr key={config.id} className="hover:bg-accent transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-foreground">{config.server}</div>
                    <div className="text-sm text-muted-foreground">{config.base_dn}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">{config.port}</td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-muted-foreground max-w-xs truncate">{config.base_dn}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      {config.is_active ? (
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-success/10 text-success">
                          <CheckCircle size={12} />
                          {t('ldap.active')}
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-muted text-muted-foreground">
                          <XCircle size={12} />
                          {t('ldap.inactive')}
                        </span>
                      )}
                      {config.sync_enabled && (
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary">
                          <Clock size={12} />
                          {t('ldap.auto_sync')}
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {config.last_sync_at ? formatRelative(config.last_sync_at) : t('common.never')}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-1">
                      <button
                        onClick={() => handleTestConnection(config)}
                        disabled={testingId === config.id}
                        className="p-1.5 text-muted-foreground hover:text-primary rounded-md hover:bg-accent disabled:opacity-50"
                        title={t('ldap.test_connection')}
                      >
                        <TestTube size={16} />
                      </button>
                      <button
                        onClick={() => handleSync(config.id)}
                        disabled={syncingId === config.id}
                        className="p-1.5 text-muted-foreground hover:text-success rounded-md hover:bg-accent disabled:opacity-50"
                        title={t('ldap.sync_now')}
                      >
                        <RefreshCw size={16} className={syncingId === config.id ? 'animate-spin' : ''} />
                      </button>
                      <Link
                        to={`/ldap/${config.id}/logs`}
                        className="p-1.5 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                        title={t('ldap.view_logs')}
                      >
                        <FileText size={16} />
                      </Link>
                      <Link
                        to={`/ldap/${config.id}`}
                        className="p-1.5 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                        title={t('common.edit')}
                      >
                        <Edit size={16} />
                      </Link>
                      <button
                        onClick={() => handleDelete(config.id, config.server)}
                        className="p-1.5 text-muted-foreground hover:text-destructive rounded-md hover:bg-accent"
                        title={t('common.delete')}
                        disabled={deleteConfig.isPending}
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {configs.length === 0 && (
            <div className="p-12 text-center text-muted-foreground">{t('ldap.no_configs')}</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default LDAPConfigs;

