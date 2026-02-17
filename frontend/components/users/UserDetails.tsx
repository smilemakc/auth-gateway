import React from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { Edit } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useUserDetail, useReset2FA, useSendPasswordReset, useUserOAuthAccounts } from '../../hooks/useUsers';
import { useUserSessions, useRevokeSession } from '../../hooks/useSessions';
import { useApiKeys } from '../../hooks/useApiKeys';
import { useUserAuditLogs } from '../../hooks/useAuditLogs';
import { useApplication } from '../../services/appContext';
import { useUserApplicationProfile } from '../../hooks/useApplications';
import { toast } from '../../services/toast';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';
import { PageHeader } from '../ui';
import UserProfileCard from './UserProfileCard';
import UserSecuritySection from './UserSecuritySection';
import UserSessionsSection from './UserSessionsSection';
import UserAppProfileSection from './UserAppProfileSection';
import UserOAuthSection from './UserOAuthSection';
import UserApiKeysSection from './UserApiKeysSection';
import UserAuditSection from './UserAuditSection';

const UserDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { currentApplicationId } = useApplication();

  const { data: user, isLoading: userLoading, error: userError } = useUserDetail(id!);
  const { data: sessionsData } = useUserSessions(id!, 1, 50);
  const { data: apiKeysData } = useApiKeys(1, 50);
  const { data: logsData } = useUserAuditLogs(id!, 1, 5);
  const { data: oauthAccountsData } = useUserOAuthAccounts(id!);
  const { data: appProfile, isLoading: appProfileLoading } = useUserApplicationProfile(id!, currentApplicationId || '');

  const revokeSessionMutation = useRevokeSession();
  const reset2FAMutation = useReset2FA();
  const sendPasswordResetMutation = useSendPasswordReset();

  const sessions = sessionsData?.sessions || [];
  const apiKeys = (apiKeysData?.api_keys || []).filter((key: any) => key.user_id === id);
  const logs = (logsData?.logs || logsData?.items || []).slice(0, 5);
  const oauthAccounts = oauthAccountsData?.accounts || [];

  const handleRevokeSession = async (sessionId: string) => {
    const ok = await confirm({ description: t('user.revoke_confirm'), variant: 'danger' });
    if (ok) {
      try {
        await revokeSessionMutation.mutateAsync(sessionId);
        toast.success(t('user.revoked_success'));
      } catch (error) {
        logger.error('Failed to revoke session:', error);
        toast.error('Failed to revoke session');
      }
    }
  };

  const handleReset2FA = async () => {
    const ok = await confirm({ description: t('user.reset_2fa_confirm'), variant: 'danger' });
    if (ok) {
      try {
        await reset2FAMutation.mutateAsync(id!);
        toast.success(t('user.reset_2fa_success'));
      } catch (error) {
        toast.error('Failed to reset 2FA: ' + (error as Error).message);
      }
    }
  };

  const handleSendPasswordReset = async () => {
    const ok = await confirm({ description: t('user.reset_password_confirm') });
    if (ok) {
      try {
        const result = await sendPasswordResetMutation.mutateAsync(id!);
        toast.success(`Password reset email sent to ${result.email}`);
      } catch (error) {
        toast.error('Failed to send password reset: ' + (error as Error).message);
      }
    }
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
        <button onClick={() => navigate('/users')} className="mt-4 text-primary hover:underline">
          {t('user.back_to_users')}
        </button>
      </div>
    );
  }

  const headerSubtitle = <>{t('user.id')}: <span className="font-mono">{user.id}</span></>;
  const headerAction = (
    <Link
      to={`/users/${user.id}/edit`}
      className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg font-medium transition-colors"
    >
      <Edit size={18} />
      {t('user.edit_profile')}
    </Link>
  );

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <PageHeader
        title={user.full_name}
        subtitle={headerSubtitle}
        onBack={() => navigate('/users')}
        action={headerAction}
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="space-y-6">
          <UserProfileCard user={user} />
          <UserSecuritySection
            totpEnabled={user.totp_enabled}
            emailVerified={user.email_verified}
            onReset2FA={handleReset2FA}
            onSendPasswordReset={handleSendPasswordReset}
            isResetting2FA={reset2FAMutation.isPending}
            isSendingPasswordReset={sendPasswordResetMutation.isPending}
          />
        </div>

        <div className="lg:col-span-2 space-y-6">
          <UserSessionsSection sessions={sessions} onRevokeSession={handleRevokeSession} />
          {currentApplicationId && (
            <UserAppProfileSection appProfile={appProfile} isLoading={appProfileLoading} />
          )}
          <UserOAuthSection oauthAccounts={oauthAccounts} />
          <UserApiKeysSection apiKeys={apiKeys} />
          <UserAuditSection logs={logs} />
        </div>
      </div>
    </div>
  );
};

export default UserDetails;
