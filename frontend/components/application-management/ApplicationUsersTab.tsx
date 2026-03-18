import React, { useState } from 'react';
import { Upload } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { LoadingSpinner } from '../ui';
import {
  useApplicationUsers,
  useBanUserFromApplication,
  useUnbanUserFromApplication,
} from '../../hooks/useApplications';
import type { UserApplicationProfile } from '../../types';
import { toast } from '../../services/toast';
import { confirm } from '../../services/confirm';
import { UsersImportModal } from '../bulk-operations';
import { logger } from '@/lib/logger';
import UsersStatsCards from './UsersStatsCards';
import UsersTable from './UsersTable';

interface ApplicationUsersTabProps {
  applicationId: string;
}

const PAGE_SIZE = 20;

const ApplicationUsersTab: React.FC<ApplicationUsersTabProps> = ({ applicationId }) => {
  const { t } = useLanguage();
  const [page, setPage] = useState(1);
  const [showImportModal, setShowImportModal] = useState(false);

  const { data, isLoading, error, refetch } = useApplicationUsers(applicationId, page, PAGE_SIZE);
  const banUser = useBanUserFromApplication();
  const unbanUser = useUnbanUserFromApplication();

  const handleBan = async (profile: UserApplicationProfile, reason: string) => {
    if (!reason.trim()) {
      toast.warning(t('apps.ban_reason_required'));
      return;
    }
    try {
      await banUser.mutateAsync({
        applicationId,
        userId: profile.user_id,
        reason,
      });
    } catch (error) {
      logger.error('Failed to ban user:', error);
    }
  };

  const handleUnban = async (profile: UserApplicationProfile) => {
    const ok = await confirm({
      description: t('apps.confirm_unban'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await unbanUser.mutateAsync({
          applicationId,
          userId: profile.user_id,
        });
      } catch (error) {
        logger.error('Failed to unban user:', error);
      }
    }
  };

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (error) {
    return (
      <div className="bg-destructive/10 border border-destructive rounded-lg p-4 text-destructive">
        {t('apps.users_load_error')}
      </div>
    );
  }

  const profiles = data?.profiles || [];
  const total = data?.total || 0;
  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-end">
        <button
          onClick={() => setShowImportModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg font-medium transition-colors"
        >
          <Upload size={18} />
          {t('apps.users.import')}
        </button>
      </div>

      <UsersStatsCards
        total={total}
        activeCount={profiles.filter(p => p.is_active && !p.is_banned).length}
        bannedCount={profiles.filter(p => p.is_banned).length}
      />

      <UsersTable
        profiles={profiles}
        total={total}
        page={page}
        pageSize={PAGE_SIZE}
        totalPages={totalPages}
        isBanPending={banUser.isPending}
        isUnbanPending={unbanUser.isPending}
        onPageChange={setPage}
        onBan={handleBan}
        onUnban={handleUnban}
      />

      {showImportModal && (
        <UsersImportModal
          applicationId={applicationId}
          onClose={() => setShowImportModal(false)}
          onSuccess={() => {
            refetch();
            toast.success(t('common.saved'));
          }}
        />
      )}
    </div>
  );
};

export default ApplicationUsersTab;
