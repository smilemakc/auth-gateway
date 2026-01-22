
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
import { useUserDetail, useReset2FA, useSendPasswordReset, useUserOAuthAccounts } from '../hooks/useUsers';
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
  const { data: oauthAccountsData } = useUserOAuthAccounts(id!);

  // Mutations
  const revokeSessionMutation = useRevokeSession();
  const reset2FAMutation = useReset2FA();
  const sendPasswordResetMutation = useSendPasswordReset();

  // Extract data from responses
  const sessions = sessionsData?.sessions || [];
  const apiKeys = (apiKeysData?.api_keys || []).filter(
    (key: any) => key.user_id === id
  );
  const logs = (logsData?.logs || logsData?.items || []).slice(0, 5);
  const oauthAccounts = oauthAccountsData || [];

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
    if (confirm('Are you sure you want to reset 2FA for this user? They will need to set up 2FA again.')) {
      try {
        await reset2FAMutation.mutateAsync(id!);
        alert('2FA has been reset successfully');
      } catch (error) {
        alert('Failed to reset 2FA: ' + (error as Error).message);
      }
    }
  };

  const handleSendPasswordReset = async () => {
    if (confirm('Are you sure you want to send a password reset email to this user?')) {
      try {
        const result = await sendPasswordResetMutation.mutateAsync(id!);
        alert(`Password reset email sent to ${result.email}`);
      } catch (error) {
        alert('Failed to send password reset: ' + (error as Error).message);
      }
    }
  };

  const getDeviceIcon = (userAgent: string) => {
    const ua = userAgent.toLowerCase();
    if (ua.includes('mobile') || ua.includes('android') || ua.includes('iphone')) {
      return <Smartphone size={18} />;
    }
    if (ua.includes('tablet') || ua.includes('ipad')) {
      return <Tablet size={18} />;
    }
    return <Monitor size={18} />;
  };

  const parseUserAgent = (userAgent: string) => {
    // Simple UA parsing - extract browser and OS info
    const browsers = ['Chrome', 'Firefox', 'Safari', 'Edge', 'Opera'];
    const oses = ['Windows', 'Mac OS', 'Linux', 'Android', 'iOS'];

    let browser = 'Unknown Browser';
    let os = 'Unknown OS';

    for (const b of browsers) {
      if (userAgent.includes(b)) {
        browser = b;
        break;
      }
    }

    for (const o of oses) {
      if (userAgent.includes(o)) {
        os = o;
        break;
      }
    }

    return { browser, os };
  };

  if (userLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (userError || !user) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive">
          {userError ? `Error loading user: ${(userError as Error).message}` : 'User not found'}
        </p>
        <button
          onClick={() => navigate('/users')}
          className="mt-4 text-primary hover:underline"
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
            className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-foreground">{user.full_name}</h1>
            <p className="text-muted-foreground">{t('user.id')}: <span className="font-mono text-sm">{user.id}</span></p>
          </div>
        </div>
        <div className="flex gap-3">
          <Link
            to={`/users/${user.id}/edit`}
            className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg font-medium transition-colors"
          >
            <Edit size={18} />
            {t('user.edit_profile')}
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Sidebar - Profile Info */}
        <div className="space-y-6">
          <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="p-6 text-center border-b border-border">
              <img
                src={user.profile_picture_url}
                alt={user.full_name}
                className="w-24 h-24 rounded-full mx-auto mb-4 border-4 border-muted"
              />
              <h2 className="text-xl font-bold text-foreground">{user.username}</h2>
              <div className="flex justify-center gap-2 flex-wrap mt-2">
                {user.roles?.map(role => (
                  <span
                    key={role.id}
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize
                      ${role.name === 'admin' ? 'bg-purple-100 text-purple-800' :
                        role.name === 'moderator' ? 'bg-indigo-100 text-indigo-800' : 'bg-muted text-foreground'}`}>
                    {role.display_name || role.name}
                  </span>
                ))}
                <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium
                  ${user.is_active ? 'bg-success/20 text-success' : 'bg-destructive/20 text-destructive'}`}>
                  {user.is_active ? t('users.active') : t('users.blocked')}
                </span>
              </div>
            </div>
            
            <div className="p-6 space-y-4">
              <div className="flex items-center gap-3 text-muted-foreground">
                <UserIcon size={18} className="text-muted-foreground" />
                <span className="text-sm font-medium">{user.username}</span>
              </div>
              <div className="flex items-center gap-3 text-muted-foreground">
                <Mail size={18} className="text-muted-foreground" />
                <span className="text-sm">{user.email}</span>
                {user.email_verified && <CheckCircle size={14} className="text-green-500 ml-auto" />}
              </div>
              <div className="flex items-center gap-3 text-muted-foreground">
                <Phone size={18} className="text-muted-foreground" />
                <span className="text-sm">{user.phone || '-'}</span>
              </div>
              <div className="pt-4 border-t border-border space-y-3">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground flex items-center gap-2">
                    <Calendar size={16} /> {t('users.col_created')}
                  </span>
                  <span className="text-foreground">{new Date(user.created_at).toLocaleDateString()}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground flex items-center gap-2">
                    <Clock size={16} /> Login
                  </span>
                  <span className="text-foreground">
                    {user.last_login ? new Date(user.last_login).toLocaleDateString() : '-'}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Security Status */}
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <Shield size={18} className="text-primary" />
              {t('user.security')}
            </h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Two-Factor Auth</span>
                {user.totp_enabled ? (
                  <span className="text-xs font-medium text-success bg-success/10 px-2 py-1 rounded-full">Enabled</span>
                ) : (
                  <span className="text-xs font-medium text-muted-foreground bg-muted px-2 py-1 rounded-full">Disabled</span>
                )}
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">{t('user.email_verified')}</span>
                {user.email_verified ? (
                  <span className="text-xs font-medium text-success bg-success/10 px-2 py-1 rounded-full">{t('common.yes')}</span>
                ) : (
                  <span className="text-xs font-medium text-yellow-700 bg-yellow-50 px-2 py-1 rounded-full">{t('common.no')}</span>
                )}
              </div>
            </div>
          </div>

          {/* Danger Zone */}
          <div className="bg-card rounded-xl shadow-sm border border-destructive/20 p-6">
             <h3 className="font-semibold text-destructive mb-4 flex items-center gap-2">
                <AlertTriangle size={18} />
                {t('user.danger_zone')}
             </h3>
             <div className="space-y-3">
               {user.totp_enabled && (
                 <button
                   onClick={handleReset2FA}
                   disabled={reset2FAMutation.isPending}
                   className="w-full flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium text-destructive bg-destructive/10 hover:bg-destructive/20 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                 >
                   <RotateCcw size={16} />
                   {reset2FAMutation.isPending ? 'Resetting...' : t('user.reset_2fa')}
                 </button>
               )}
               <button
                 onClick={handleSendPasswordReset}
                 disabled={sendPasswordResetMutation.isPending}
                 className="w-full flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
               >
                 <Send size={16} />
                 {sendPasswordResetMutation.isPending ? 'Sending...' : t('user.reset_password_email')}
               </button>
             </div>
          </div>
        </div>

        {/* Main Content Area */}
        <div className="lg:col-span-2 space-y-6">

          {/* Active Sessions - NEW */}
          <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="p-6 border-b border-border flex items-center justify-between">
              <h3 className="font-semibold text-foreground flex items-center gap-2">
                <Monitor size={18} className="text-primary" />
                {t('user.sessions')}
              </h3>
              <span className="text-xs font-medium bg-muted text-muted-foreground px-2 py-1 rounded-full">{sessions.length}</span>
            </div>
            <div className="divide-y divide-border">
              {sessions.length > 0 ? (
                sessions.map(session => {
                  const { browser, os } = parseUserAgent(session.user_agent);
                  return (
                    <div key={session.id} className="p-4 flex flex-col sm:flex-row sm:items-center justify-between hover:bg-accent gap-4">
                      <div className="flex items-start gap-3">
                        <div className={`p-2 rounded-lg bg-muted text-muted-foreground`}>
                          {getDeviceIcon(session.user_agent)}
                        </div>
                        <div>
                          <div className="flex items-center gap-2">
                            <p className="text-sm font-medium text-foreground">{os} - {browser}</p>
                            {session.is_current && (
                              <span className="text-[10px] font-bold bg-success/20 text-success px-1.5 py-0.5 rounded uppercase">{t('user.current')}</span>
                            )}
                          </div>
                          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground mt-1">
                            <span className="flex items-center gap-1">
                              <Globe size={12} /> {session.ip_address}
                            </span>
                            <span className="flex items-center gap-1">
                              • Active {new Date(session.last_active_at).toLocaleDateString()}
                            </span>
                          </div>
                        </div>
                      </div>
                      <button
                        onClick={() => handleRevokeSession(session.id)}
                        className="text-sm text-destructive hover:text-destructive hover:bg-destructive/10 px-3 py-1.5 rounded-md font-medium transition-colors border border-transparent hover:border-destructive/20 self-start sm:self-center"
                      >
                        {t('user.revoke')}
                      </button>
                    </div>
                  );
                })
              ) : (
                <div className="p-6 text-sm text-muted-foreground italic">{t('user.no_sessions')}</div>
              )}
            </div>
          </div>
          
          {/* Linked Accounts */}
          <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="p-6 border-b border-border">
              <h3 className="font-semibold text-foreground flex items-center gap-2">
                <Globe size={18} className="text-primary" />
                {t('user.linked_accounts')}
              </h3>
            </div>
            <div className="p-6">
              {oauthAccounts.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  {oauthAccounts.map((acc: any) => (
                    <div key={acc.id} className="flex items-center justify-between p-3 border border-border rounded-lg">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded bg-muted flex items-center justify-center font-bold text-muted-foreground uppercase">
                          {acc.provider?.[0] || '?'}
                        </div>
                        <div>
                          <p className="text-sm font-medium capitalize">{acc.provider}</p>
                          <p className="text-xs text-muted-foreground">Connected {new Date(acc.created_at).toLocaleDateString()}</p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground italic">No external accounts linked.</p>
              )}
            </div>
          </div>

          {/* API Keys */}
          <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="p-6 border-b border-border flex justify-between items-center">
              <h3 className="font-semibold text-foreground flex items-center gap-2">
                <Key size={18} className="text-primary" />
                {t('dash.api_keys')}
              </h3>
              <span className="text-xs font-medium bg-muted text-muted-foreground px-2 py-1 rounded-full">{apiKeys.length}</span>
            </div>
            <div className="overflow-x-auto">
              {apiKeys.length > 0 ? (
                <table className="min-w-full divide-y divide-border">
                  <thead className="bg-muted">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase">Name</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase">Prefix</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase">Status</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase">Created</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {apiKeys.map(key => (
                      <tr key={key.id}>
                        <td className="px-6 py-4 text-sm font-medium text-foreground">{key.name}</td>
                        <td className="px-6 py-4 text-sm font-mono text-muted-foreground">{key.key_prefix}...</td>
                        <td className="px-6 py-4">
                          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize
                            ${key.is_active ? 'bg-success/20 text-success' : 'bg-destructive/20 text-destructive'}`}>
                            {key.is_active ? 'active' : 'inactive'}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-sm text-muted-foreground">{new Date(key.created_at).toLocaleDateString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="p-6 text-sm text-muted-foreground italic">{t('user.no_keys')}</div>
              )}
            </div>
          </div>

          {/* Recent Activity */}
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
                        <p className="text-xs text-muted-foreground">{new Date(log.created_at).toLocaleString()}</p>
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
                <div className="p-6 text-sm text-muted-foreground italic">No recent activity.</div>
              )}
            </div>
          </div>

        </div>
      </div>
    </div>
  );
};

export default UserDetails;