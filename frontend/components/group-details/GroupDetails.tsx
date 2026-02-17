import React, { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { Edit, Trash2, Users } from 'lucide-react';
import { LoadingSpinner } from '../ui';
import { useGroup, useDeleteGroup, useGroupMembers, useAddGroupMembers, useRemoveGroupMember } from '../../hooks/useGroups';
import { useUsers } from '../../hooks/useUsers';
import { formatDate } from '../../lib/date';
import { toast } from '../../services/toast';
import { confirm } from '../../services/confirm';
import { useLanguage } from '../../services/i18n';
import { logger } from '@/lib/logger';
import { GroupMembersSection } from './GroupMembersSection';

const GroupDetails: React.FC = () => {
  const { t } = useLanguage();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const { data: group, isLoading: isLoadingGroup } = useGroup(id || '');
  const { data: membersData, isLoading: isLoadingMembers } = useGroupMembers(id || '', page, pageSize);
  const { data: usersData } = useUsers(1, 100);
  const deleteGroup = useDeleteGroup();
  const addMembers = useAddGroupMembers();
  const removeMember = useRemoveGroupMember();

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

  const handleAddMembers = async (userIds: string[]) => {
    if (!id) return;
    await addMembers.mutateAsync({
      id,
      data: { user_ids: userIds },
    });
  };

  const handleRemoveMember = async (userId: string) => {
    if (!id) return;
    await removeMember.mutateAsync({ groupId: id, userId });
  };

  if (isLoadingGroup) {
    return <LoadingSpinner className="min-h-screen" />;
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

      <GroupMembersSection
        groupId={id || ''}
        members={members}
        availableUsers={availableUsers}
        isLoadingMembers={isLoadingMembers}
        membersTotal={membersData?.total || 0}
        page={page}
        pageSize={pageSize}
        onPageChange={setPage}
        onAddMembers={handleAddMembers}
        isAddingMembers={addMembers.isPending}
        onRemoveMember={handleRemoveMember}
        isRemovingMember={removeMember.isPending}
      />
    </div>
  );
};

export default GroupDetails;
