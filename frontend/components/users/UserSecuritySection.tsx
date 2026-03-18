import React from 'react';
import {
  Shield,
  AlertTriangle,
  RotateCcw,
  Send,
} from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface UserSecuritySectionProps {
  totpEnabled: boolean;
  emailVerified: boolean;
  onReset2FA: () => void;
  onSendPasswordReset: () => void;
  isResetting2FA: boolean;
  isSendingPasswordReset: boolean;
}

const UserSecuritySection: React.FC<UserSecuritySectionProps> = ({
  totpEnabled,
  emailVerified,
  onReset2FA,
  onSendPasswordReset,
  isResetting2FA,
  isSendingPasswordReset,
}) => {
  const { t } = useLanguage();

  return (
    <>
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
          <Shield size={18} className="text-primary" />
          {t('user.security')}
        </h3>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">{t('user.two_factor')}</span>
            {totpEnabled ? (
              <span className="text-xs font-medium text-success bg-success/10 px-2 py-1 rounded-full">{t('user.enabled')}</span>
            ) : (
              <span className="text-xs font-medium text-muted-foreground bg-muted px-2 py-1 rounded-full">{t('user.disabled')}</span>
            )}
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">{t('user.email_verified')}</span>
            {emailVerified ? (
              <span className="text-xs font-medium text-success bg-success/10 px-2 py-1 rounded-full">{t('common.yes')}</span>
            ) : (
              <span className="text-xs font-medium text-yellow-700 bg-yellow-50 px-2 py-1 rounded-full">{t('common.no')}</span>
            )}
          </div>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-destructive/20 p-6">
        <h3 className="font-semibold text-destructive mb-4 flex items-center gap-2">
          <AlertTriangle size={18} />
          {t('user.danger_zone')}
        </h3>
        <div className="space-y-3">
          {totpEnabled && (
            <button
              onClick={onReset2FA}
              disabled={isResetting2FA}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium text-destructive bg-destructive/10 hover:bg-destructive/20 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <RotateCcw size={16} />
              {isResetting2FA ? t('user.resetting') : t('user.reset_2fa')}
            </button>
          )}
          <button
            onClick={onSendPasswordReset}
            disabled={isSendingPasswordReset}
            className="w-full flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send size={16} />
            {isSendingPasswordReset ? t('user.sending') : t('user.reset_password_email')}
          </button>
        </div>
      </div>
    </>
  );
};

export default UserSecuritySection;
