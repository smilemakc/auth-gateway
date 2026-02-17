import React, { useState } from 'react';
import { Plus, X, Loader } from 'lucide-react';
import type { AdminUserResponse } from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';
import { confirm } from '../../services/confirm';
import { toast } from '../../services/toast';
import { logger } from '@/lib/logger';

interface GroupMembersSectionProps {
  groupId: string;
  members: AdminUserResponse[];
  availableUsers: AdminUserResponse[];
  isLoadingMembers: boolean;
  membersTotal: number;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onAddMembers: (userIds: string[]) => Promise<void>;
  isAddingMembers: boolean;
  onRemoveMember: (userId: string) => Promise<void>;
  isRemovingMember: boolean;
}

export const GroupMembersSection: React.FC<GroupMembersSectionProps> = ({
  groupId,
  members,
  availableUsers,
  isLoadingMembers,
  membersTotal,
  page,
  pageSize,
  onPageChange,
  onAddMembers,
  isAddingMembers,
  onRemoveMember,
  isRemovingMember,
}) => {
  const { t } = useLanguage();
  const [showAddMembers, setShowAddMembers] = useState(false);
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);

  const handleAddMembers = async () => {
    if (selectedUserIds.length === 0) return;
    try {
      await onAddMembers(selectedUserIds);
      setShowAddMembers(false);
      setSelectedUserIds([]);
    } catch (error) {
      logger.error('Failed to add members:', error);
      toast.error(t('group_details.add_error'));
    }
  };

  const handleRemoveMember = async (userId: string) => {
    const ok = await confirm({
      description: t('group_details.remove_member_confirm'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await onRemoveMember(userId);
      } catch (error) {
        logger.error('Failed to remove member:', error);
        toast.error(t('group_details.remove_error'));
      }
    }
  };

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="p-4 border-b border-border flex items-center justify-between">
        <h2 className="text-lg font-semibold text-foreground">{t('groups.col_members')}</h2>
        <button
          onClick={() => setShowAddMembers(true)}
          className="px-3 py-1.5 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm transition-colors flex items-center gap-2"
        >
          <Plus size={16} />
          {t('group_details.add_members')}
        </button>
      </div>

      {showAddMembers && (
        <div className="p-4 border-b border-border bg-muted">
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-medium text-foreground">{t('group_details.select_users')}</h3>
            <button
              onClick={() => {
                setShowAddMembers(false);
                setSelectedUserIds([]);
              }}
              className="text-muted-foreground hover:text-foreground"
            >
              <X size={20} />
            </button>
          </div>
          <div className="max-h-48 overflow-y-auto space-y-2 mb-3">
            {availableUsers.map((user) => (
              <label key={user.id} className="flex items-center gap-2 p-2 hover:bg-card rounded cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedUserIds.includes(user.id)}
                  onChange={(e) => {
                    if (e.target.checked) {
                      setSelectedUserIds([...selectedUserIds, user.id]);
                    } else {
                      setSelectedUserIds(selectedUserIds.filter((id) => id !== user.id));
                    }
                  }}
                  className="rounded border-input text-primary focus:ring-ring"
                />
                <span className="text-sm text-foreground">{user.email}</span>
              </label>
            ))}
            {availableUsers.length === 0 && (
              <p className="text-sm text-muted-foreground text-center py-4">{t('group_details.all_members')}</p>
            )}
          </div>
          <button
            onClick={handleAddMembers}
            disabled={selectedUserIds.length === 0 || isAddingMembers}
            className="w-full px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            {isAddingMembers && <Loader size={16} className="animate-spin" />}
            {t('group_details.add')} {selectedUserIds.length > 0 && `(${selectedUserIds.length})`}
          </button>
        </div>
      )}

      {isLoadingMembers ? (
        <div className="p-8 text-center">
          <Loader size={24} className="animate-spin text-primary mx-auto" />
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('users.col_user')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('auth.email')}
                </th>
                <th scope="col" className="relative px-6 py-3">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {members.map((member) => (
                <tr key={member.id} className="hover:bg-accent transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-foreground">{member.username}</div>
                    {member.full_name && <div className="text-sm text-muted-foreground">{member.full_name}</div>}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-muted-foreground">{member.email}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button
                      onClick={() => handleRemoveMember(member.id)}
                      className="text-destructive hover:text-destructive p-1 rounded-md hover:bg-destructive/10"
                      disabled={isRemovingMember}
                    >
                      <X size={18} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {members.length === 0 && (
            <div className="p-12 text-center text-muted-foreground">{t('group_details.no_members')}</div>
          )}
        </div>
      )}

      {membersTotal > pageSize && (
        <div className="px-6 py-4 border-t border-border flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            {t('common.showing')} {(page - 1) * pageSize + 1} {t('common.to')} {Math.min(page * pageSize, membersTotal)} {t('common.of')}{' '}
            {membersTotal} {t('groups.col_members')}
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => onPageChange(Math.max(1, page - 1))}
              disabled={page === 1}
              className="px-3 py-1 border border-input rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.previous')}
            </button>
            <button
              onClick={() => onPageChange(page + 1)}
              disabled={page * pageSize >= membersTotal}
              className="px-3 py-1 border border-input rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.next')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
