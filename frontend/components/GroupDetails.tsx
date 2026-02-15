import React, { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { Edit, Trash2, Users, Plus, X, Loader } from 'lucide-react';
import type { AdminUserResponse } from '@auth-gateway/client-sdk';
import { useGroup, useDeleteGroup, useGroupMembers, useAddGroupMembers, useRemoveGroupMember } from '../hooks/useGroups';
import { useUsers } from '../hooks/useUsers';
import { formatDate } from '../lib/date';
import { toast } from '../services/toast';
import { confirm } from '../services/confirm';
import { useLanguage } from '../services/i18n';
import { logger } from '@/lib/logger';

const GroupDetails: React.FC = () => {
  const { t } = useLanguage();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [showAddMembers, setShowAddMembers] = useState(false);
  const pageSize = 20;

  const { data: group, isLoading: isLoadingGroup } = useGroup(id || '');
  const { data: membersData, isLoading: isLoadingMembers } = useGroupMembers(id || '', page, pageSize);
  const { data: usersData } = useUsers(1, 100); // For member selection
  const deleteGroup = useDeleteGroup();
  const addMembers = useAddGroupMembers();
  const removeMember = useRemoveGroupMember();

  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);

  const handleDelete = async () => {
    if (!group) return;
    const ok = await confirm({
      description: `${t('group_details.delete_confirm')} "${group.display_name}"?`,
      variant: 'danger'
    });
    if (ok) {
      try {
        await deleteGroup.mutateAsync(group.id);
        navigate('/groups');
      } catch (error) {
        logger.error('Failed to delete group:', error);
        toast.error(t('group_details.delete_error'));
      }
    }
  };

  const handleAddMembers = async () => {
    if (!id || selectedUserIds.length === 0) return;
    try {
      await addMembers.mutateAsync({
        id,
        data: { user_ids: selectedUserIds },
      });
      setShowAddMembers(false);
      setSelectedUserIds([]);
    } catch (error) {
      logger.error('Failed to add members:', error);
      toast.error(t('group_details.add_error'));
    }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!id) return;
    const ok = await confirm({
      description: t('group_details.remove_member_confirm'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await removeMember.mutateAsync({ groupId: id, userId });
      } catch (error) {
        logger.error('Failed to remove member:', error);
        toast.error(t('group_details.remove_error'));
      }
    }
  };

  if (isLoadingGroup) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (!group) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive">{t('group_details.not_found')}</p>
        <Link to="/groups" className="text-primary hover:underline mt-4 inline-block">
          {t('group_details.back_to_groups')}
        </Link>
      </div>
    );
  }

  const members = membersData?.users || [];
  const availableUsers = usersData?.users.filter(
    (u) => !members.some((m) => m.id === u.id)
  ) || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{group.display_name}</h1>
          {group.description && <p className="text-muted-foreground mt-1">{group.description}</p>}
        </div>
        <div className="flex gap-2">
          <Link
            to={`/groups/${group.id}/edit`}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors flex items-center gap-2"
          >
            <Edit size={16} />
            {t('common.edit')}
          </Link>
          {!group.is_system_group && (
            <button
              onClick={handleDelete}
              disabled={deleteGroup.isPending}
              className="px-4 py-2 border border-destructive rounded-lg text-destructive hover:bg-destructive/10 transition-colors flex items-center gap-2 disabled:opacity-50"
            >
              <Trash2 size={16} />
              {t('common.delete')}
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="text-sm text-muted-foreground">{t('group_details.group_name')}</div>
          <div className="text-lg font-semibold text-foreground mt-1">{group.name}</div>
        </div>
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="text-sm text-muted-foreground">{t('groups.col_members')}</div>
          <div className="text-lg font-semibold text-foreground mt-1 flex items-center gap-2">
            <Users size={20} />
            {group.member_count || 0}
          </div>
        </div>
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="text-sm text-muted-foreground">{t('common.created')}</div>
          <div className="text-lg font-semibold text-foreground mt-1">
            {formatDate(group.created_at)}
          </div>
        </div>
      </div>

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
              disabled={selectedUserIds.length === 0 || addMembers.isPending}
              className="w-full px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {addMembers.isPending && <Loader size={16} className="animate-spin" />}
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
                        disabled={removeMember.isPending}
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

        {membersData && membersData.total > pageSize && (
          <div className="px-6 py-4 border-t border-border flex items-center justify-between">
            <div className="text-sm text-muted-foreground">
              {t('common.showing')} {(page - 1) * pageSize + 1} {t('common.to')} {Math.min(page * pageSize, membersData.total)} {t('common.of')}{' '}
              {membersData.total} {t('groups.col_members')}
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 border border-input rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
              >
                {t('common.previous')}
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * pageSize >= membersData.total}
                className="px-3 py-1 border border-input rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
              >
                {t('common.next')}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default GroupDetails;

