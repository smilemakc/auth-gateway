import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Loader, CheckCircle, XCircle, Search } from 'lucide-react';
import type { BulkUserUpdate, BulkOperationResult, AdminUserResponse } from '@auth-gateway/client-sdk';
import { useBulkUpdateUsers } from '../hooks/useBulkOperations';
import { useUsers } from '../hooks/useUsers';
import { toast } from '../services/toast';

const BulkUpdateUsers: React.FC = () => {
  const navigate = useNavigate();
  const bulkUpdate = useBulkUpdateUsers();
  const { data: usersData } = useUsers(1, 1000);

  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [updateFields, setUpdateFields] = useState<{
    email?: string;
    username?: string;
    full_name?: string;
    is_active?: boolean;
  }>({});
  const [result, setResult] = useState<BulkOperationResult | null>(null);
  const [searchTerm, setSearchTerm] = useState('');

  const users = usersData?.users || [];
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

  const handleSelectAll = () => {
    if (selectedUserIds.length === filteredUsers.length) {
      setSelectedUserIds([]);
    } else {
      setSelectedUserIds(filteredUsers.map((u) => u.id));
    }
  };

  const handleSubmit = async () => {
    if (selectedUserIds.length === 0) {
      toast.warning('Please select at least one user');
      return;
    }

    const updates: BulkUserUpdate[] = selectedUserIds.map((id) => ({
      id,
      email: updateFields.email || undefined,
      username: updateFields.username || undefined,
      full_name: updateFields.full_name || undefined,
      is_active: updateFields.is_active !== undefined ? updateFields.is_active : undefined,
    }));

    try {
      const result = await bulkUpdate.mutateAsync({ users: updates });
      setResult(result);
    } catch (error) {
      console.error('Bulk update failed:', error);
      toast.error('Failed to update users');
    }
  };

  if (result) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/bulk')} className="text-muted-foreground hover:text-foreground flex items-center gap-2">
            <ArrowLeft size={20} />
            Back
          </button>
          <h1 className="text-2xl font-bold text-foreground">Bulk Update Results</h1>
        </div>

        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="grid grid-cols-3 gap-4 mb-6">
            <div className="bg-muted rounded-lg p-4">
              <div className="text-sm text-muted-foreground">Total</div>
              <div className="text-2xl font-bold text-foreground">{result.total}</div>
            </div>
            <div className="bg-success/10 rounded-lg p-4">
              <div className="text-sm text-success">Success</div>
              <div className="text-2xl font-bold text-success">{result.success}</div>
            </div>
            <div className="bg-destructive/10 rounded-lg p-4">
              <div className="text-sm text-destructive">Failed</div>
              <div className="text-2xl font-bold text-destructive">{result.failed}</div>
            </div>
          </div>

          {result.errors && result.errors.length > 0 && (
            <div className="mb-4">
              <h3 className="font-semibold text-foreground mb-2">Errors</h3>
              <div className="space-y-2">
                {result.errors.map((error, index) => (
                  <div key={index} className="bg-destructive/10 border border-border rounded-lg p-3">
                    <div className="text-sm text-destructive">{error.message}</div>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="flex justify-end gap-3 pt-4 border-t border-border">
            <button
              onClick={() => {
                setResult(null);
                setSelectedUserIds([]);
                setUpdateFields({});
              }}
              className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
            >
              Update More
            </button>
            <button
              onClick={() => navigate('/bulk')}
              className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors"
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
        <button onClick={() => navigate('/bulk')} className="text-muted-foreground hover:text-foreground flex items-center gap-2">
          <ArrowLeft size={20} />
          Back
        </button>
        <h1 className="text-2xl font-bold text-foreground">Bulk Update Users</h1>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6 space-y-6">
        {/* Update Fields */}
        <div>
          <h2 className="text-lg font-semibold text-foreground mb-4">Update Fields</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Email</label>
              <input
                type="email"
                value={updateFields.email || ''}
                onChange={(e) => setUpdateFields({ ...updateFields, email: e.target.value || undefined })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Leave empty to keep current"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Username</label>
              <input
                type="text"
                value={updateFields.username || ''}
                onChange={(e) => setUpdateFields({ ...updateFields, username: e.target.value || undefined })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Leave empty to keep current"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Full Name</label>
              <input
                type="text"
                value={updateFields.full_name || ''}
                onChange={(e) => setUpdateFields({ ...updateFields, full_name: e.target.value || undefined })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Leave empty to keep current"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Status</label>
              <select
                value={updateFields.is_active === undefined ? '' : updateFields.is_active ? 'true' : 'false'}
                onChange={(e) =>
                  setUpdateFields({
                    ...updateFields,
                    is_active: e.target.value === '' ? undefined : e.target.value === 'true',
                  })
                }
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="">Keep current</option>
                <option value="true">Active</option>
                <option value="false">Inactive</option>
              </select>
            </div>
          </div>
        </div>

        {/* User Selection */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-foreground">
              Select Users ({selectedUserIds.length} selected)
            </h2>
            <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" size={18} />
                <input
                  type="text"
                  placeholder="Search users..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 pr-4 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                />
              </div>
              <button
                onClick={handleSelectAll}
                className="px-3 py-2 border border-input rounded-lg text-sm text-foreground hover:bg-accent"
              >
                {selectedUserIds.length === filteredUsers.length ? 'Deselect All' : 'Select All'}
              </button>
            </div>
          </div>

          <div className="border border-border rounded-lg max-h-96 overflow-y-auto">
            <table className="min-w-full divide-y divide-border">
              <thead className="bg-muted sticky top-0">
                <tr>
                  <th scope="col" className="px-4 py-3 text-left">
                    <input
                      type="checkbox"
                      checked={selectedUserIds.length === filteredUsers.length && filteredUsers.length > 0}
                      onChange={handleSelectAll}
                      className="rounded border-input text-primary focus:ring-ring"
                    />
                  </th>
                  <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase">
                    User
                  </th>
                  <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase">
                    Email
                  </th>
                  <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase">
                    Status
                  </th>
                </tr>
              </thead>
              <tbody className="bg-card divide-y divide-border">
                {filteredUsers.map((user) => (
                  <tr key={user.id} className="hover:bg-accent">
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedUserIds.includes(user.id)}
                        onChange={() => toggleUserSelection(user.id)}
                        className="rounded border-input text-primary focus:ring-ring"
                      />
                    </td>
                    <td className="px-4 py-3 text-sm font-medium text-foreground">{user.username}</td>
                    <td className="px-4 py-3 text-sm text-muted-foreground">{user.email}</td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${
                          user.is_active ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'
                        }`}
                      >
                        {user.is_active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-border">
          <button
            onClick={() => navigate('/bulk')}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={bulkUpdate.isPending || selectedUserIds.length === 0}
            className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {bulkUpdate.isPending && <Loader size={16} className="animate-spin" />}
            Update {selectedUserIds.length > 0 && `(${selectedUserIds.length})`}
          </button>
        </div>
      </div>
    </div>
  );
};

export default BulkUpdateUsers;

