import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Loader, Search, Shield } from 'lucide-react';
import type { BulkOperationResult, AdminUserResponse, Role } from '@auth-gateway/client-sdk';
import { useBulkAssignRoles } from '../hooks/useBulkOperations';
import { useUsers } from '../hooks/useUsers';
import { useRoles } from '../hooks/useRBAC';

const BulkAssignRoles: React.FC = () => {
  const navigate = useNavigate();
  const bulkAssignRoles = useBulkAssignRoles();
  const { data: usersData } = useUsers(1, 1000);
  const { data: rolesData } = useRoles();

  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([]);
  const [result, setResult] = useState<BulkOperationResult | null>(null);
  const [searchTerm, setSearchTerm] = useState('');

  const users = usersData?.users || [];
  const roles = rolesData || [];
  const filteredUsers = searchTerm
    ? users.filter(
        (u) =>
          u.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
          u.username.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : users;

  const toggleUserSelection = (userId: string) => {
    if (selectedUserIds.includes(userId)) {
      setSelectedUserIds(selectedUserIds.filter((id) => id !== userId));
    } else {
      setSelectedUserIds([...selectedUserIds, userId]);
    }
  };

  const toggleRoleSelection = (roleId: string) => {
    if (selectedRoleIds.includes(roleId)) {
      setSelectedRoleIds(selectedRoleIds.filter((id) => id !== roleId));
    } else {
      setSelectedRoleIds([...selectedRoleIds, roleId]);
    }
  };

  const handleSelectAllUsers = () => {
    if (selectedUserIds.length === filteredUsers.length) {
      setSelectedUserIds([]);
    } else {
      setSelectedUserIds(filteredUsers.map((u) => u.id));
    }
  };

  const handleSubmit = async () => {
    if (selectedUserIds.length === 0) {
      alert('Please select at least one user');
      return;
    }
    if (selectedRoleIds.length === 0) {
      alert('Please select at least one role');
      return;
    }

    try {
      const result = await bulkAssignRoles.mutateAsync({
        user_ids: selectedUserIds,
        role_ids: selectedRoleIds,
      });
      setResult(result);
    } catch (error) {
      console.error('Bulk assign roles failed:', error);
      alert('Failed to assign roles');
    }
  };

  if (result) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/bulk')} className="text-gray-500 hover:text-gray-700 flex items-center gap-2">
            <ArrowLeft size={20} />
            Back
          </button>
          <h1 className="text-2xl font-bold text-gray-900">Bulk Assign Roles Results</h1>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <div className="grid grid-cols-3 gap-4 mb-6">
            <div className="bg-gray-50 rounded-lg p-4">
              <div className="text-sm text-gray-500">Total</div>
              <div className="text-2xl font-bold text-gray-900">{result.total}</div>
            </div>
            <div className="bg-green-50 rounded-lg p-4">
              <div className="text-sm text-green-600">Success</div>
              <div className="text-2xl font-bold text-green-700">{result.success}</div>
            </div>
            <div className="bg-red-50 rounded-lg p-4">
              <div className="text-sm text-red-600">Failed</div>
              <div className="text-2xl font-bold text-red-700">{result.failed}</div>
            </div>
          </div>

          {result.errors && result.errors.length > 0 && (
            <div className="mb-4">
              <h3 className="font-semibold text-gray-900 mb-2">Errors</h3>
              <div className="space-y-2">
                {result.errors.map((error, index) => (
                  <div key={index} className="bg-red-50 border border-red-200 rounded-lg p-3">
                    <div className="text-sm text-red-800">{error.message}</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
            <button
              onClick={() => {
                setResult(null);
                setSelectedUserIds([]);
                setSelectedRoleIds([]);
              }}
              className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
            >
              Assign More
            </button>
            <button
              onClick={() => navigate('/bulk')}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
            >
              Done
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <button onClick={() => navigate('/bulk')} className="text-gray-500 hover:text-gray-700 flex items-center gap-2">
          <ArrowLeft size={20} />
          Back
        </button>
        <h1 className="text-2xl font-bold text-gray-900">Bulk Assign Roles</h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Role Selection */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <Shield size={20} />
            Select Roles ({selectedRoleIds.length} selected)
          </h2>
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {roles.map((role) => (
              <label
                key={role.id}
                className="flex items-center gap-3 p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer"
              >
                <input
                  type="checkbox"
                  checked={selectedRoleIds.includes(role.id)}
                  onChange={() => toggleRoleSelection(role.id)}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <div className="flex-1">
                  <div className="text-sm font-medium text-gray-900">{role.display_name || role.name}</div>
                  {role.description && <div className="text-xs text-gray-500">{role.description}</div>}
                </div>
              </label>
            ))}
            {roles.length === 0 && <p className="text-sm text-gray-500 text-center py-4">No roles available</p>}
          </div>
        </div>

        {/* User Selection */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">
              Select Users ({selectedUserIds.length} selected)
            </h2>
            <div className="flex items-center gap-2">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={18} />
                <input
                  type="text"
                  placeholder="Search users..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                />
              </div>
              <button
                onClick={handleSelectAllUsers}
                className="px-3 py-2 border border-gray-300 rounded-lg text-sm text-gray-700 hover:bg-gray-50"
              >
                {selectedUserIds.length === filteredUsers.length ? 'Deselect All' : 'Select All'}
              </button>
            </div>
          </div>

          <div className="border border-gray-200 rounded-lg max-h-96 overflow-y-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50 sticky top-0">
                <tr>
                  <th scope="col" className="px-4 py-3 text-left">
                    <input
                      type="checkbox"
                      checked={selectedUserIds.length === filteredUsers.length && filteredUsers.length > 0}
                      onChange={handleSelectAllUsers}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                  </th>
                  <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    User
                  </th>
                  <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Current Roles
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {filteredUsers.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedUserIds.includes(user.id)}
                        onChange={() => toggleUserSelection(user.id)}
                        className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      />
                    </td>
                    <td className="px-4 py-3">
                      <div className="text-sm font-medium text-gray-900">{user.username}</div>
                      <div className="text-xs text-gray-500">{user.email}</div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {user.roles?.map((role) => (
                          <span
                            key={role.id}
                            className="inline-flex px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800"
                          >
                            {role.display_name || role.name}
                          </span>
                        ))}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div className="flex justify-end gap-3">
        <button
          onClick={() => navigate('/bulk')}
          className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
        >
          Cancel
        </button>
        <button
          onClick={handleSubmit}
          disabled={bulkAssignRoles.isPending || selectedUserIds.length === 0 || selectedRoleIds.length === 0}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
        >
          {bulkAssignRoles.isPending && <Loader size={16} className="animate-spin" />}
          Assign Roles to {selectedUserIds.length > 0 && `${selectedUserIds.length} User(s)`}
        </button>
      </div>
    </div>
  );
};

export default BulkAssignRoles;

