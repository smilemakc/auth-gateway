import React from 'react';
import { Briefcase } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { formatDate, formatRelative } from '../../lib/date';
import { LoadingSpinner } from '../ui';

interface AppProfile {
  display_name?: string;
  nickname?: string;
  app_roles?: string[];
  is_active: boolean;
  is_banned: boolean;
  ban_reason?: string;
  banned_at?: string;
  last_access_at?: string;
}

interface UserAppProfileSectionProps {
  appProfile: AppProfile | null | undefined;
  isLoading: boolean;
}

const UserAppProfileSection: React.FC<UserAppProfileSectionProps> = ({ appProfile, isLoading }) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="p-6 border-b border-border">
        <h3 className="font-semibold text-foreground flex items-center gap-2">
          <Briefcase size={18} className="text-primary" />
          {t('user.app_profile')}
        </h3>
      </div>
      <div className="p-6">
        {isLoading ? (
          <LoadingSpinner className="py-8" />
        ) : appProfile ? (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{t('user.display_name')}</span>
              <span className="text-sm font-medium text-foreground">{appProfile.display_name || '-'}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{t('user.nickname')}</span>
              <span className="text-sm font-medium text-foreground">{appProfile.nickname || '-'}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{t('user.app_roles')}</span>
              <div className="flex gap-2 flex-wrap justify-end">
                {appProfile.app_roles && appProfile.app_roles.length > 0 ? (
                  appProfile.app_roles.map((role, idx) => (
                    <span
                      key={idx}
                      className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800"
                    >
                      {role}
                    </span>
                  ))
                ) : (
                  <span className="text-sm text-muted-foreground">{t('user.no_roles')}</span>
                )}
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{t('common.status')}</span>
              <span className={`text-xs font-medium px-2 py-1 rounded-full ${
                appProfile.is_banned
                  ? 'bg-destructive/20 text-destructive'
                  : appProfile.is_active
                  ? 'bg-success/20 text-success'
                  : 'bg-yellow-100 text-yellow-800'
              }`}>
                {appProfile.is_banned ? t('apps.banned') : appProfile.is_active ? t('common.active') : t('common.inactive')}
              </span>
            </div>
            {appProfile.is_banned && appProfile.ban_reason && (
              <div className="pt-3 border-t border-border">
                <p className="text-xs text-muted-foreground mb-1">{t('user.ban_reason')}</p>
                <p className="text-sm text-destructive">{appProfile.ban_reason}</p>
                {appProfile.banned_at && (
                  <p className="text-xs text-muted-foreground mt-1">
                    {t('user.banned_on')} {formatDate(appProfile.banned_at)}
                  </p>
                )}
              </div>
            )}
            {appProfile.last_access_at && (
              <div className="flex items-center justify-between pt-3 border-t border-border">
                <span className="text-sm text-muted-foreground">{t('apps.last_access')}</span>
                <span className="text-sm text-foreground">
                  {formatRelative(appProfile.last_access_at)}
                </span>
              </div>
            )}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground italic">
            {t('user.not_associated')}
          </p>
        )}
      </div>
    </div>
  );
};

export default UserAppProfileSection;
