import React, { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { Edit, Trash2, Users, Plus, X, Loader } from 'lucide-react';
import type { AdminUserResponse } from '@auth-gateway/client-sdk';
import { useGroup, useDeleteGroup, useGroupMembers, useAddGroupMembers, useRemoveGroupMember } from '../hooks/useGroups';
import { useUsers } from '../hooks/useUsers';

const GroupDetails: React.FC = () => {
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
    if (window.confirm(`Are you sure you want to delete group "${group.display_name}"?`)) {
      try {
        await deleteGroup.mutateAsync(group.id);
        navigate('/groups');
      } catch (error) {
        console.error('Failed to delete group:', error);
        alert('Failed to delete group');
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
      console.error('Failed to add members:', error);
      alert('Failed to add members');
    }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!id) return;
    if (window.confirm('Are you sure you want to remove this member from the group?')) {
      try {
        await removeMember.mutateAsync({ groupId: id, userId });
      } catch (error) {
        console.error('Failed to remove member:', error);
        alert('Failed to remove member');
      }
    }
  };

  if (isLoadingGroup) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (!group) {
    return (
      <div className="p-8 text-center">
        <p className="text-red-600">Group not found</p>
        <Link to="/groups" className="text-blue-600 hover:underline mt-4 inline-block">
          Back to Groups
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
          <h1 className="text-2xl font-bold text-gray-900">{group.display_name}</h1>
          {group.description && <p className="text-gray-500 mt-1">{group.description}</p>}
        </div>
        <div className="flex gap-2">
          <Link
            to={`/groups/${group.id}/edit`}
            className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors flex items-center gap-2"
          >
            <Edit size={16} />
            Edit
          </Link>
          {!group.is_system_group && (
            <button
              onClick={handleDelete}
              disabled={deleteGroup.isPending}
              className="px-4 py-2 border border-red-300 rounded-lg text-red-700 hover:bg-red-50 transition-colors flex items-center gap-2 disabled:opacity-50"
            >
              <Trash2 size={16} />
              Delete
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <div className="text-sm text-gray-500">Group Name</div>
          <div className="text-lg font-semibold text-gray-900 mt-1">{group.name}</div>
        </div>
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <div className="text-sm text-gray-500">Members</div>
          <div className="text-lg font-semibold text-gray-900 mt-1 flex items-center gap-2">
            <Users size={20} />
            {group.member_count || 0}
          </div>
        </div>
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <div className="text-sm text-gray-500">Created</div>
          <div className="text-lg font-semibold text-gray-900 mt-1">
            {new Date(group.created_at).toLocaleDateString()}
          </div>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-4 border-b border-gray-100 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Members</h2>
          <button
            onClick={() => setShowAddMembers(true)}
            className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm transition-colors flex items-center gap-2"
          >
            <Plus size={16} />
            Add Members
          </button>
        </div>

        {showAddMembers && (
          <div className="p-4 border-b border-gray-100 bg-gray-50">
            <div className="flex items-center justify-between mb-3">
              <h3 className="font-medium text-gray-900">Select users to add</h3>
              <button
                onClick={() => {
                  setShowAddMembers(false);
                  setSelectedUserIds([]);
                }}
                className="text-gray-500 hover:text-gray-700"
              >
                <X size={20} />
              </button>
            </div>
            <div className="max-h-48 overflow-y-auto space-y-2 mb-3">
              {availableUsers.map((user) => (
                <label key={user.id} className="flex items-center gap-2 p-2 hover:bg-white rounded cursor-pointer">
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
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-900">{user.email}</span>
                </label>
              ))}
              {availableUsers.length === 0 && (
                <p className="text-sm text-gray-500 text-center py-4">All users are already members</p>
              )}
            </div>
            <button
              onClick={handleAddMembers}
              disabled={selectedUserIds.length === 0 || addMembers.isPending}
              className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {addMembers.isPending && <Loader size={16} className="animate-spin" />}
              Add {selectedUserIds.length > 0 && `(${selectedUserIds.length})`}
            </button>
          </div>
        )}

        {isLoadingMembers ? (
          <div className="p-8 text-center">
            <Loader size={24} className="animate-spin text-blue-600 mx-auto" />
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    User
                  </th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Email
                  </th>
                  <th scope="col" className="relative px-6 py-3">
                    <span className="sr-only">Actions</span>
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {members.map((member) => (
                  <tr key={member.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">{member.username}</div>
                      {member.full_name && <div className="text-sm text-gray-500">{member.full_name}</div>}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-500">{member.email}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => handleRemoveMember(member.id)}
                        className="text-red-600 hover:text-red-800 p-1 rounded-md hover:bg-red-50"
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
              <div className="p-12 text-center text-gray-500">No members in this group.</div>
            )}
          </div>
        )}

        {membersData && membersData.total > pageSize && (
          <div className="px-6 py-4 border-t border-gray-100 flex items-center justify-between">
            <div className="text-sm text-gray-500">
              Showing {(page - 1) * pageSize + 1} to {Math.min(page * pageSize, membersData.total)} of{' '}
              {membersData.total} members
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                Previous
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * pageSize >= membersData.total}
                className="px-3 py-1 border border-gray-300 rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default GroupDetails;

