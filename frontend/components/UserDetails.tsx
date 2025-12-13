
import React from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import {
  ArrowLeft,
  Mail,
  Phone,
  Shield,
  Calendar,
  Clock,
  Edit,
  Key,
  Globe,
  CheckCircle,
  XCircle,
  Activity,
  User as UserIcon,
  Smartphone,
  Monitor,
  Tablet,
  LogOut,
  AlertTriangle,
  RotateCcw,
  Send
} from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useUserDetail, useResetUser2FA, useSendPasswordReset } from '../hooks/useUsers';
import { useUserSessions, useRevokeSession } from '../hooks/useSessions';
import { useApiKeys } from '../hooks/useApiKeys';
import { useUserAuditLogs } from '../hooks/useAuditLogs';

const UserDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();

  // Fetch user data with React Query
  const { data: user, isLoading: userLoading, error: userError } = useUserDetail(id!);
  const { data: sessionsData } = useUserSessions(id!, 1, 50);
  const { data: apiKeysData } = useApiKeys(1, 50);
  const { data: logsData } = useUserAuditLogs(id!, 1, 5);

  // Mutations
  const revokeSessionMutation = useRevokeSession();
  const reset2FAMutation = useResetUser2FA();
  const sendPasswordResetMutation = useSendPasswordReset();

  // Extract data from responses
  const sessions = sessionsData?.sessions || sessionsData?.items || [];
  const apiKeys = (apiKeysData?.apiKeys || apiKeysData?.items || []).filter(
    (key: any) => key.userId === id
  );
  const logs = (logsData?.logs || logsData?.items || []).slice(0, 5);
  const oauthAccounts: any[] = []; // TODO: Add OAuth accounts API

  const handleRevokeSession = async (sessionId: string) => {
    if (window.confirm('Are you sure you want to revoke this session?')) {
      try {
        await revokeSessionMutation.mutateAsync(sessionId);
        alert('Session revoked successfully');
      } catch (error) {
        console.error('Failed to revoke session:', error);
        alert('Failed to revoke session');
      }
    }
  };

  const handleReset2FA = async () => {
    if (!user) return;
    if (window.confirm(t('user.reset_2fa_confirm'))) {
      try {
        await reset2FAMutation.mutateAsync(user.id);
        alert('2FA has been reset.');
      } catch (error) {
        console.error('Failed to reset 2FA:', error);
        alert('Failed to reset 2FA');
      }
    }
  };

  const handleResetPassword = async () => {
    if (!user) return;
    try {
      await sendPasswordResetMutation.mutateAsync(user.id);
      alert('Password reset email sent.');
    } catch (error) {
      console.error('Failed to send password reset:', error);
      alert('Failed to send password reset email');
    }
  };

  const getDeviceIcon = (type: string) => {
    switch (type) {
      case 'mobile': return <Smartphone size={18} />;
      case 'tablet': return <Tablet size={18} />;
      case 'desktop': return <Monitor size={18} />;
      default: return <Monitor size={18} />;
    }
  };

  if (userLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (userError || !user) {
    return (
      <div className="p-8 text-center">
        <p className="text-red-600">
          {userError ? `Error loading user: ${(userError as Error).message}` : 'User not found'}
        </p>
        <button
          onClick={() => navigate('/users')}
          className="mt-4 text-blue-600 hover:underline"
        >
          Back to Users
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <button 
            onClick={() => navigate('/users')}
            className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">{user.fullName}</h1>
            <p className="text-gray-500">{t('user.id')}: <span className="font-mono text-sm">{user.id}</span></p>
          </div>
        </div>
        <div className="flex gap-3">
          <Link 
            to={`/users/${user.id}/edit`}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors"
          >
            <Edit size={18} />
            {t('user.edit_profile')}
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Sidebar - Profile Info */}
        <div className="space-y-6">
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="p-6 text-center border-b border-gray-100">
              <img 
                src={user.avatarUrl} 
                alt={user.fullName} 
                className="w-24 h-24 rounded-full mx-auto mb-4 border-4 border-gray-50"
              />
              <h2 className="text-xl font-bold text-gray-900">{user.username}</h2>
              <div className="flex justify-center gap-2 flex-wrap mt-2">
                {user.roles?.map(role => (
                  <span
                    key={role.id}
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize
                      ${role.name === 'admin' ? 'bg-purple-100 text-purple-800' :
                        role.name === 'moderator' ? 'bg-indigo-100 text-indigo-800' : 'bg-gray-100 text-gray-800'}`}>
                    {role.displayName || role.name}
                  </span>
                ))}
                <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium
                  ${user.isActive ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                  {user.isActive ? t('users.active') : t('users.blocked')}
                </span>
              </div>
            </div>
            
            <div className="p-6 space-y-4">
              <div className="flex items-center gap-3 text-gray-600">
                <UserIcon size={18} className="text-gray-400" />
                <span className="text-sm font-medium">{user.username}</span>
              </div>
              <div className="flex items-center gap-3 text-gray-600">
                <Mail size={18} className="text-gray-400" />
                <span className="text-sm">{user.email}</span>
                {user.isEmailVerified && <CheckCircle size={14} className="text-green-500 ml-auto" />}
              </div>
              <div className="flex items-center gap-3 text-gray-600">
                <Phone size={18} className="text-gray-400" />
                <span className="text-sm">{user.phone || '-'}</span>
              </div>
              <div className="pt-4 border-t border-gray-100 space-y-3">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-500 flex items-center gap-2">
                    <Calendar size={16} /> {t('users.col_created')}
                  </span>
                  <span className="text-gray-900">{new Date(user.createdAt).toLocaleDateString()}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-500 flex items-center gap-2">
                    <Clock size={16} /> Login
                  </span>
                  <span className="text-gray-900">
                    {user.lastLogin ? new Date(user.lastLogin).toLocaleDateString() : '-'}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Security Status */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
            <h3 className="font-semibold text-gray-900 mb-4 flex items-center gap-2">
              <Shield size={18} className="text-blue-500" />
              {t('user.security')}
            </h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">Two-Factor Auth</span>
                {user.is2FAEnabled ? (
                  <span className="text-xs font-medium text-green-700 bg-green-50 px-2 py-1 rounded-full">Enabled</span>
                ) : (
                  <span className="text-xs font-medium text-gray-600 bg-gray-100 px-2 py-1 rounded-full">Disabled</span>
                )}
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">{t('user.email_verified')}</span>
                {user.isEmailVerified ? (
                  <span className="text-xs font-medium text-green-700 bg-green-50 px-2 py-1 rounded-full">{t('common.yes')}</span>
                ) : (
                  <span className="text-xs font-medium text-yellow-700 bg-yellow-50 px-2 py-1 rounded-full">{t('common.no')}</span>
                )}
              </div>
            </div>
          </div>

          {/* Danger Zone */}
          <div className="bg-white rounded-xl shadow-sm border border-red-100 p-6">
             <h3 className="font-semibold text-red-700 mb-4 flex items-center gap-2">
                <AlertTriangle size={18} />
                {t('user.danger_zone')}
             </h3>
             <div className="space-y-3">
               <button 
                 onClick={handleReset2FA}
                 disabled={!user.is2FAEnabled}
                 className="w-full flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium text-red-700 bg-red-50 hover:bg-red-100 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed"
               >
                 <RotateCcw size={16} />
                 {t('user.reset_2fa')}
               </button>
               <button 
                 onClick={handleResetPassword}
                 className="w-full flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium text-gray-700 bg-gray-50 hover:bg-gray-100 rounded-lg"
               >
                 <Send size={16} />
                 {t('user.reset_password_email')}
               </button>
             </div>
          </div>
        </div>

        {/* Main Content Area */}
        <div className="lg:col-span-2 space-y-6">

          {/* Active Sessions - NEW */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="p-6 border-b border-gray-100 flex items-center justify-between">
              <h3 className="font-semibold text-gray-900 flex items-center gap-2">
                <Monitor size={18} className="text-blue-500" />
                {t('user.sessions')}
              </h3>
              <span className="text-xs font-medium bg-gray-100 text-gray-600 px-2 py-1 rounded-full">{sessions.length}</span>
            </div>
            <div className="divide-y divide-gray-100">
              {sessions.length > 0 ? (
                sessions.map(session => (
                  <div key={session.id} className="p-4 flex flex-col sm:flex-row sm:items-center justify-between hover:bg-gray-50 gap-4">
                    <div className="flex items-start gap-3">
                      <div className={`p-2 rounded-lg bg-gray-100 text-gray-600`}>
                        {getDeviceIcon(session.deviceType)}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-medium text-gray-900">{session.os} - {session.browser}</p>
                          {session.isCurrent && (
                            <span className="text-[10px] font-bold bg-green-100 text-green-700 px-1.5 py-0.5 rounded uppercase">{t('user.current')}</span>
                          )}
                        </div>
                        <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500 mt-1">
                          <span className="flex items-center gap-1">
                            <Globe size={12} /> {session.ipAddress}
                          </span>
                           <span className="flex items-center gap-1">
                            • {session.location}
                          </span>
                          <span className="flex items-center gap-1">
                            • Active {new Date(session.lastActive).toLocaleDateString()}
                          </span>
                        </div>
                      </div>
                    </div>
                    <button 
                      onClick={() => handleRevokeSession(session.id)}
                      className="text-sm text-red-600 hover:text-red-800 hover:bg-red-50 px-3 py-1.5 rounded-md font-medium transition-colors border border-transparent hover:border-red-100 self-start sm:self-center"
                    >
                      {t('user.revoke')}
                    </button>
                  </div>
                ))
              ) : (
                <div className="p-6 text-sm text-gray-500 italic">{t('user.no_sessions')}</div>
              )}
            </div>
          </div>
          
          {/* Linked Accounts */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="p-6 border-b border-gray-100">
              <h3 className="font-semibold text-gray-900 flex items-center gap-2">
                <Globe size={18} className="text-blue-500" />
                {t('user.linked_accounts')}
              </h3>
            </div>
            <div className="p-6">
              {oauthAccounts.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  {oauthAccounts.map(acc => (
                    <div key={acc.id} className="flex items-center justify-between p-3 border border-gray-200 rounded-lg">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded bg-gray-100 flex items-center justify-center font-bold text-gray-500 uppercase">
                          {acc.provider[0]}
                        </div>
                        <div>
                          <p className="text-sm font-medium capitalize">{acc.provider}</p>
                          <p className="text-xs text-gray-500">Connected {new Date(acc.connectedAt).toLocaleDateString()}</p>
                        </div>
                      </div>
                      <button className="text-xs text-red-600 hover:underline">Unlink</button>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-gray-500 italic">No external accounts linked.</p>
              )}
            </div>
          </div>

          {/* API Keys */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="p-6 border-b border-gray-100 flex justify-between items-center">
              <h3 className="font-semibold text-gray-900 flex items-center gap-2">
                <Key size={18} className="text-blue-500" />
                {t('dash.api_keys')}
              </h3>
              <span className="text-xs font-medium bg-gray-100 text-gray-600 px-2 py-1 rounded-full">{apiKeys.length}</span>
            </div>
            <div className="overflow-x-auto">
              {apiKeys.length > 0 ? (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Prefix</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {apiKeys.map(key => (
                      <tr key={key.id}>
                        <td className="px-6 py-4 text-sm font-medium text-gray-900">{key.name}</td>
                        <td className="px-6 py-4 text-sm font-mono text-gray-600">{key.prefix}...</td>
                        <td className="px-6 py-4">
                          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize
                            ${key.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                            {key.status}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">{new Date(key.createdAt).toLocaleDateString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="p-6 text-sm text-gray-500 italic">{t('user.no_keys')}</div>
              )}
            </div>
          </div>

          {/* Recent Activity */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
            <div className="p-6 border-b border-gray-100">
              <h3 className="font-semibold text-gray-900 flex items-center gap-2">
                <Activity size={18} className="text-blue-500" />
                {t('user.recent_activity')}
              </h3>
            </div>
            <div className="divide-y divide-gray-200">
              {logs.length > 0 ? (
                logs.map(log => (
                  <div key={log.id} className="p-4 flex items-center justify-between hover:bg-gray-50">
                    <div className="flex items-center gap-4">
                      <div className={`p-2 rounded-full ${
                        log.status === 'success' ? 'bg-green-100 text-green-600' : 
                        log.status === 'failed' ? 'bg-red-100 text-red-600' : 'bg-gray-100 text-gray-600'
                      }`}>
                        <Activity size={16} />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-900 capitalize">{log.action.replace(/_/g, ' ')}</p>
                        <p className="text-xs text-gray-500">{new Date(log.timestamp).toLocaleString()}</p>
                      </div>
                    </div>
                    <div className="text-right">
                       <span className={`text-xs font-medium px-2 py-1 rounded-full capitalize ${
                         log.status === 'success' ? 'text-green-700 bg-green-50' : 'text-red-700 bg-red-50'
                       }`}>
                         {log.status}
                       </span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="p-6 text-sm text-gray-500 italic">No recent activity.</div>
              )}
            </div>
          </div>

        </div>
      </div>
    </div>
  );
};

export default UserDetails;