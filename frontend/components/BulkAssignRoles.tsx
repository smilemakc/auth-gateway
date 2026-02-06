import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Loader, Search, Shield } from 'lucide-react';
import type { BulkOperationResult, AdminUserResponse, Role } from '@auth-gateway/client-sdk';
import { useBulkAssignRoles } from '../hooks/useBulkOperations';
import { useUsers } from '../hooks/useUsers';
import { useRoles } from '../hooks/useRBAC';
import { toast } from '../services/toast';
import { useLanguage } from '../services/i18n';

const BulkAssignRoles: React.FC = () => {
  const { t } = useLanguage();
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
      toast.warning(t('bulk_update.select_user_warning'));
      return;
    }
    if (selectedRoleIds.length === 0) {
      toast.warning(t('bulk_assign.select_role_warning'));
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
      toast.error(t('bulk_assign.assign_error'));
    }
  };

  if (result) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/bulk')} className="text-muted-foreground hover:text-foreground flex items-center gap-2">
            <ArrowLeft size={20} />
            {t('common.back')}
          </button>
          <h1 className="text-2xl font-bold text-foreground">{t('bulk_assign.results_title')}</h1>
        </div>

        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="grid grid-cols-3 gap-4 mb-6">
            <div className="bg-muted rounded-lg p-4">
              <div className="text-sm text-muted-foreground">{t('bulk_update.total')}</div>
              <div className="text-2xl font-bold text-foreground">{result.total}</div>
            </div>
            <div className="bg-success/10 rounded-lg p-4">
              <div className="text-sm text-success">{t('bulk_update.success')}</div>
              <div className="text-2xl font-bold text-success">{result.success}</div>
            </div>
            <div className="bg-destructive/10 rounded-lg p-4">
              <div className="text-sm text-destructive">{t('bulk_update.failed')}</div>
              <div className="text-2xl font-bold text-destructive">{result.failed}</div>
            </div>
          </div>

          {result.errors && result.errors.length > 0 && (
            <div className="mb-4">
              <h3 className="font-semibold text-foreground mb-2">{t('bulk_update.errors')}</h3>
              <div className="space-y-2">
                {result.errors.map((error, index) => (
                  <div key={index} className="bg-destructive/10 border border-destructive rounded-lg p-3">
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
                setSelectedRoleIds([]);
              }}
              className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
            >
              {t('bulk_assign.assign_more')}
            </button>
            <button
              onClick={() => navigate('/bulk')}
              className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors"
            >
              {t('common.done')}
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
          {t('common.back')}
        </button>
        <h1 className="text-2xl font-bold text-foreground">{t('bulk.assign_roles')}</h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Role Selection */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4 flex items-center gap-2">
            <Shield size={20} />
            {t('bulk_assign.select_roles')} ({selectedRoleIds.length} {t('bulk_update.selected')})
          </h2>
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {roles.map((role) => (
              <label
                key={role.id}
                className="flex items-center gap-3 p-3 border border-border rounded-lg hover:bg-accent cursor-pointer"
              >
                <input
                  type="checkbox"
                  checked={selectedRoleIds.includes(role.id)}
                  onChange={() => toggleRoleSelection(role.id)}
                  className="rounded border-input text-primary focus:ring-ring"
                />
                <div className="flex-1">
                  <div className="text-sm font-medium text-foreground">{role.display_name || role.name}</div>
                  {role.description && <div className="text-xs text-muted-foreground">{role.description}</div>}
                </div>
              </label>
            ))}
            {roles.length === 0 && <p className="text-sm text-muted-foreground text-center py-4">{t('bulk_assign.no_roles')}</p>}
          </div>
        </div>

        {/* User Selection */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-foreground">
              {t('bulk_update.select_users')} ({selectedUserIds.length} {t('bulk_update.selected')})
            </h2>
            <div className="flex items-center gap-2">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" size={18} />
                <input
                  type="text"
                  placeholder={t('common.search')}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 pr-4 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring text-sm"
                />
              </div>
              <button
                onClick={handleSelectAllUsers}
                className="px-3 py-2 border border-input rounded-lg text-sm text-foreground hover:bg-accent"
              >
                {selectedUserIds.length === filteredUsers.length ? t('bulk_update.deselect_all') : t('bulk_update.select_all')}
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
                      onChange={handleSelectAllUsers}
                      className="rounded border-input text-primary focus:ring-ring"
                    />
                  </th>
                  <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase">
                    {t('users.col_user')}
                  </th>
                  <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase">
                    {t('bulk_assign.current_roles')}
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
                    <td className="px-4 py-3">
                      <div className="text-sm font-medium text-foreground">{user.username}</div>
                      <div className="text-xs text-muted-foreground">{user.email}</div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {user.roles?.map((role) => (
                          <span
                            key={role.id}
                            className="inline-flex px-2 py-0.5 rounded-full text-xs font-medium bg-muted text-foreground"
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
          className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
        >
          {t('common.cancel')}
        </button>
        <button
          onClick={handleSubmit}
          disabled={bulkAssignRoles.isPending || selectedUserIds.length === 0 || selectedRoleIds.length === 0}
          className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
        >
          {bulkAssignRoles.isPending && <Loader size={16} className="animate-spin" />}
          {t('bulk_assign.assign_btn')} {selectedUserIds.length > 0 && `(${selectedUserIds.length})`}
        </button>
      </div>
    </div>
  );
};

export default BulkAssignRoles;

