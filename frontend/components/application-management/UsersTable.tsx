import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Ban, UserCheck, Eye, Users, Loader2, AlertCircle } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { UserApplicationProfile } from '../../types';
import { formatRelative } from '../../lib/date';

interface UsersTableProps {
  profiles: UserApplicationProfile[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  isBanPending: boolean;
  isUnbanPending: boolean;
  onPageChange: (page: number) => void;
  onBan: (profile: UserApplicationProfile, reason: string) => void;
  onUnban: (profile: UserApplicationProfile) => void;
}

const UsersTable: React.FC<UsersTableProps> = ({
  profiles,
  total,
  page,
  pageSize,
  totalPages,
  isBanPending,
  isUnbanPending,
  onPageChange,
  onBan,
  onUnban,
}) => {
  const { t } = useLanguage();
  const [banReason, setBanReason] = useState('');
  const [banningUserId, setBanningUserId] = useState<string | null>(null);

  const handleBanClick = (profile: UserApplicationProfile) => {
    onBan(profile, banReason);
    if (banReason.trim()) {
      setBanningUserId(null);
      setBanReason('');
    }
  };

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-muted">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {t('apps.user')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {t('apps.profile')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {t('common.status')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {t('apps.last_access')}
              </th>
              <th className="px-6 py-3 text-right text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {t('common.actions')}
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {profiles.map((profile) => (
              <React.Fragment key={profile.id}>
                <tr className="hover:bg-accent transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-3">
                      {profile.avatar_url ? (
                        <img
                          src={profile.avatar_url}
                          alt=""
                          className="w-10 h-10 rounded-full object-cover"
                        />
                      ) : (
                        <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                          <span className="text-primary font-medium">
                            {profile.user?.email?.charAt(0).toUpperCase() || '?'}
                          </span>
                        </div>
                      )}
                      <div>
                        <p className="font-medium text-foreground">
                          {profile.user?.full_name || profile.user?.username || profile.user?.email}
                        </p>
                        <p className="text-sm text-muted-foreground">{profile.user?.email}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div>
                      {profile.display_name && (
                        <p className="text-sm text-foreground">{profile.display_name}</p>
                      )}
                      {profile.nickname && (
                        <p className="text-sm text-muted-foreground">@{profile.nickname}</p>
                      )}
                      {profile.app_roles && profile.app_roles.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {profile.app_roles.map((role) => (
                            <span
                              key={role}
                              className="px-2 py-0.5 bg-primary/10 text-primary text-xs rounded"
                            >
                              {role}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {profile.is_banned ? (
                      <span className="inline-flex items-center gap-1 px-2 py-1 bg-destructive/10 text-destructive text-xs font-medium rounded">
                        <Ban size={12} />
                        {t('apps.banned')}
                      </span>
                    ) : profile.is_active ? (
                      <span className="inline-flex items-center gap-1 px-2 py-1 bg-success/10 text-success text-xs font-medium rounded">
                        <UserCheck size={12} />
                        {t('common.active')}
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 px-2 py-1 bg-muted text-muted-foreground text-xs font-medium rounded">
                        {t('common.inactive')}
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {profile.last_access_at
                      ? formatRelative(profile.last_access_at)
                      : t('apps.never')}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Link
                        to={`/users/${profile.user_id}`}
                        className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                        title={t('common.view')}
                      >
                        <Eye size={18} />
                      </Link>
                      {profile.is_banned ? (
                        <button
                          onClick={() => onUnban(profile)}
                          disabled={isUnbanPending}
                          className="p-2 text-muted-foreground hover:text-success hover:bg-success/10 rounded-lg transition-colors"
                          title={t('apps.unban')}
                        >
                          <UserCheck size={18} />
                        </button>
                      ) : (
                        <button
                          onClick={() => setBanningUserId(profile.user_id)}
                          className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                          title={t('apps.ban')}
                        >
                          <Ban size={18} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>

                {banningUserId === profile.user_id && (
                  <tr className="bg-destructive/5">
                    <td colSpan={5} className="px-6 py-4">
                      <div className="flex items-center gap-4">
                        <AlertCircle className="text-destructive" size={20} />
                        <div className="flex-1">
                          <p className="text-sm font-medium text-destructive mb-2">
                            {t('apps.ban_user')}
                          </p>
                          <input
                            type="text"
                            value={banReason}
                            onChange={e => setBanReason(e.target.value)}
                            placeholder={t('apps.ban_reason_placeholder')}
                            className="w-full px-3 py-2 bg-background border border-destructive/30 rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-destructive"
                          />
                        </div>
                        <div className="flex gap-2">
                          <button
                            onClick={() => {
                              setBanningUserId(null);
                              setBanReason('');
                            }}
                            className="px-3 py-2 text-sm text-foreground hover:bg-accent rounded-lg transition-colors"
                          >
                            {t('common.cancel')}
                          </button>
                          <button
                            onClick={() => handleBanClick(profile)}
                            disabled={isBanPending || !banReason.trim()}
                            className="flex items-center gap-2 px-3 py-2 bg-destructive hover:bg-destructive/90 text-destructive-foreground rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
                          >
                            {isBanPending && <Loader2 className="w-4 h-4 animate-spin" />}
                            {t('apps.confirm_ban')}
                          </button>
                        </div>
                      </div>
                      {profile.is_banned && profile.ban_reason && (
                        <p className="mt-2 text-sm text-muted-foreground">
                          <span className="font-medium">{t('apps.current_reason')}:</span> {profile.ban_reason}
                        </p>
                      )}
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      </div>

      {profiles.length === 0 && (
        <div className="text-center py-12">
          <Users className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-medium text-foreground">{t('apps.no_users')}</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {t('apps.no_users_desc')}
          </p>
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-between px-4 py-3 border-t border-border">
          <div className="text-sm text-foreground">
            {t('common.showing')} <span className="font-medium">{(page - 1) * pageSize + 1}</span>{' '}
            {t('common.to')} <span className="font-medium">{Math.min(page * pageSize, total)}</span>{' '}
            {t('common.of')} <span className="font-medium">{total}</span>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => onPageChange(Math.max(1, page - 1))}
              disabled={page === 1}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.previous')}
            </button>
            <button
              onClick={() => onPageChange(Math.min(totalPages, page + 1))}
              disabled={page === totalPages}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.next')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default UsersTable;
