import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Plus, Edit, Trash2, TestTube, RefreshCw, FileText, CheckCircle, XCircle, Clock } from 'lucide-react';
import type { LDAPConfig } from '@auth-gateway/client-sdk';
import { useLDAPConfigs, useDeleteLDAPConfig, useSyncLDAP, useTestLDAPConnection } from '../hooks/useLDAP';

const LDAPConfigs: React.FC = () => {
  const navigate = useNavigate();
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);

  const { data, isLoading, error } = useLDAPConfigs();
  const deleteConfig = useDeleteLDAPConfig();
  const syncLDAP = useSyncLDAP();
  const testConnection = useTestLDAPConnection();

  const handleDelete = async (id: string, name: string) => {
    if (window.confirm(`Are you sure you want to delete LDAP configuration "${name}"?`)) {
      try {
        await deleteConfig.mutateAsync(id);
      } catch (error) {
        console.error('Failed to delete LDAP config:', error);
        alert('Failed to delete LDAP configuration');
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
        alert(`Connection successful!\nUsers: ${result.user_count || 0}\nGroups: ${result.group_count || 0}`);
      } else {
        alert(`Connection failed: ${result.error || result.message}`);
      }
    } catch (error) {
      alert(`Connection test failed: ${(error as Error).message}`);
    } finally {
      setTestingId(null);
    }
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    try {
      await syncLDAP.mutateAsync({ id, data: { sync_users: true, sync_groups: true } });
      alert('Synchronization started successfully');
    } catch (error) {
      console.error('Failed to start sync:', error);
      alert('Failed to start synchronization');
    } finally {
      setSyncingId(null);
    }
  };

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
        <p className="text-red-600">Error loading LDAP configurations: {(error as Error).message}</p>
      </div>
    );
  }

  const configs = data?.configs || [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">LDAP Configurations</h1>
          <p className="text-gray-500 mt-1">Manage LDAP/Active Directory integrations</p>
        </div>
        <button
          onClick={() => navigate('/ldap/new')}
          className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
        >
          <Plus size={18} />
          Create Configuration
        </button>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Name / Server
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Port
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Base DN
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Last Sync
                </th>
                <th scope="col" className="relative px-6 py-3">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {configs.map((config) => (
                <tr key={config.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-gray-900">{config.server}</div>
                    <div className="text-sm text-gray-500">{config.base_dn}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{config.port}</td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-gray-500 max-w-xs truncate">{config.base_dn}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      {config.is_active ? (
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                          <CheckCircle size={12} />
                          Active
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                          <XCircle size={12} />
                          Inactive
                        </span>
                      )}
                      {config.sync_enabled && (
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          <Clock size={12} />
                          Auto-sync
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {config.last_sync_at ? new Date(config.last_sync_at).toLocaleString() : 'Never'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-1">
                      <button
                        onClick={() => handleTestConnection(config)}
                        disabled={testingId === config.id}
                        className="p-1.5 text-gray-400 hover:text-blue-600 rounded-md hover:bg-gray-100 disabled:opacity-50"
                        title="Test Connection"
                      >
                        <TestTube size={16} />
                      </button>
                      <button
                        onClick={() => handleSync(config.id)}
                        disabled={syncingId === config.id}
                        className="p-1.5 text-gray-400 hover:text-green-600 rounded-md hover:bg-gray-100 disabled:opacity-50"
                        title="Sync Now"
                      >
                        <RefreshCw size={16} className={syncingId === config.id ? 'animate-spin' : ''} />
                      </button>
                      <Link
                        to={`/ldap/${config.id}/logs`}
                        className="p-1.5 text-gray-400 hover:text-purple-600 rounded-md hover:bg-gray-100"
                        title="View Sync Logs"
                      >
                        <FileText size={16} />
                      </Link>
                      <Link
                        to={`/ldap/${config.id}`}
                        className="p-1.5 text-gray-400 hover:text-blue-600 rounded-md hover:bg-gray-100"
                        title="Edit"
                      >
                        <Edit size={16} />
                      </Link>
                      <button
                        onClick={() => handleDelete(config.id, config.server)}
                        className="p-1.5 text-gray-400 hover:text-red-600 rounded-md hover:bg-gray-100"
                        title="Delete"
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
            <div className="p-12 text-center text-gray-500">No LDAP configurations found.</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default LDAPConfigs;

