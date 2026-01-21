import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Loader, AlertTriangle, Search } from 'lucide-react';
import type { BulkOperationResult, AdminUserResponse } from '@auth-gateway/client-sdk';
import { useBulkDeleteUsers } from '../hooks/useBulkOperations';
import { useUsers } from '../hooks/useUsers';

const BulkDeleteUsers: React.FC = () => {
  const navigate = useNavigate();
  const bulkDelete = useBulkDeleteUsers();
  const { data: usersData } = useUsers(1, 1000);

  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
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
      alert('Please select at least one user');
      return;
    }

    if (
      !window.confirm(
        `Are you sure you want to delete ${selectedUserIds.length} user(s)? This action cannot be undone.`
      )
    ) {
      return;
    }

    try {
      const result = await bulkDelete.mutateAsync({ user_ids: selectedUserIds });
      setResult(result);
    } catch (error) {
      console.error('Bulk delete failed:', error);
      alert('Failed to delete users');
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
          <h1 className="text-2xl font-bold text-foreground">Bulk Delete Results</h1>
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
        <h1 className="text-2xl font-bold text-foreground">Bulk Delete Users</h1>
      </div>

      <div className="bg-destructive/10 border border-border rounded-lg p-4 flex items-start gap-3">
        <AlertTriangle className="text-destructive mt-0.5" size={20} />
        <div>
          <h3 className="font-semibold text-destructive mb-1">Warning</h3>
          <p className="text-sm text-destructive">
            This will permanently delete the selected users. This action cannot be undone.
          </p>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-foreground">
            Select Users to Delete ({selectedUserIds.length} selected)
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

        <div className="flex justify-end gap-3 pt-4 border-t border-border">
          <button
            onClick={() => navigate('/bulk')}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={bulkDelete.isPending || selectedUserIds.length === 0}
            className="px-4 py-2 bg-destructive hover:bg-destructive/90 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {bulkDelete.isPending && <Loader size={16} className="animate-spin" />}
            Delete {selectedUserIds.length > 0 && `(${selectedUserIds.length})`}
          </button>
        </div>
      </div>
    </div>
  );
};

export default BulkDeleteUsers;

