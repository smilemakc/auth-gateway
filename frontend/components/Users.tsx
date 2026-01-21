
import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Search, Shield, ShieldOff, Check, X, Filter, Eye, Edit } from 'lucide-react';
import type { AdminUserResponse } from '@auth-gateway/client-sdk';
import { useLanguage } from '../services/i18n';
import { useUsers, useUpdateUser } from '../hooks/useUsers';

const Users: React.FC = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('all');
  const navigate = useNavigate();
  const { t } = useLanguage();

  // Fetch users with React Query
  const { data, isLoading, error } = useUsers(1, 50, searchTerm, roleFilter);
  const updateUserMutation = useUpdateUser();

  const toggleStatus = async (id: string, currentStatus: boolean) => {
    try {
      await updateUserMutation.mutateAsync({
        id,
        data: { is_active: !currentStatus },
      });
    } catch (error) {
      console.error('Failed to toggle user status:', error);
      alert('Failed to update user status');
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive">Error loading users: {(error as Error).message}</p>
      </div>
    );
  }

  const users = data?.users || [];

  const filterUsersByRole = (userList: AdminUserResponse[]) => {
    if (roleFilter === 'all') return userList;
    return userList.filter((u) =>
      u.roles?.some((r) => r.name === roleFilter)
    );
  };

  const filteredUsers = filterUsersByRole(users);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h1 className="text-2xl font-bold text-foreground">{t('users.title')}</h1>
        <button
          onClick={() => navigate('/users/new')}
          className="bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          + {t('users.create_new')}
        </button>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        {/* Filters */}
        <div className="p-4 border-b border-border flex flex-col sm:flex-row gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" size={20} />
            <input
              type="text"
              placeholder={t('common.search')}
              className="w-full pl-10 pr-4 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          <div className="flex items-center gap-2">
            <Filter size={20} className="text-muted-foreground" />
            <select
              className="border border-input rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              value={roleFilter}
              onChange={(e) => setRoleFilter(e.target.value)}
            >
              <option value="all">{t('users.filter_role')}</option>
              <option value="admin">Admin</option>
              <option value="moderator">Moderator</option>
              <option value="user">User</option>
            </select>
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('users.col_user')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('users.col_role')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('users.col_status')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('users.col_2fa')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('users.col_created')}</th>
                <th scope="col" className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {filteredUsers.map((user) => (
                <tr key={user.id} className="hover:bg-accent transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div className="flex-shrink-0 h-10 w-10">
                        <img className="h-10 w-10 rounded-full" src={user.profile_picture_url || `https://ui-avatars.com/api/?name=${user.username}`} alt="" />
                      </div>
                      <div className="ml-4">
                        <div className="text-sm font-medium text-foreground">{user.username}</div>
                        <div className="text-sm text-muted-foreground">{user.email}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex gap-1 flex-wrap">
                      {user.roles?.map(role => (
                        <span
                          key={role.id}
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize
                            ${role.name === 'admin' ? 'bg-purple-100 text-purple-800' :
                              role.name === 'moderator' ? 'bg-indigo-100 text-indigo-800' : 'bg-muted text-foreground'}`}>
                          {role.display_name || role.name}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                     <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium
                      ${user.is_active ? 'bg-success/20 text-success' : 'bg-destructive/20 text-destructive'}`}>
                      <span className={`h-1.5 w-1.5 rounded-full ${user.is_active ? 'bg-success' : 'bg-destructive'}`}></span>
                      {user.is_active ? t('users.active') : t('users.blocked')}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {user.totp_enabled ? (
                      <Check size={16} className="text-success" />
                    ) : (
                      <X size={16} className="text-muted-foreground" />
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {new Date(user.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <Link
                        to={`/users/${user.id}`}
                        className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                        title={t('users.view_details')}
                      >
                        <Eye size={18} />
                      </Link>
                      <Link
                        to={`/users/${user.id}/edit`}
                        className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                        title={t('common.edit')}
                      >
                        <Edit size={18} />
                      </Link>
                      <button
                        onClick={() => toggleStatus(user.id, user.is_active)}
                        className={`p-1 rounded-md hover:bg-accent ${user.is_active ? 'text-muted-foreground hover:text-destructive' : 'text-muted-foreground hover:text-success'}`}
                        disabled={updateUserMutation.isPending}
                      >
                        {user.is_active ? <ShieldOff size={18} /> : <Shield size={18} />}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {filteredUsers.length === 0 && (
             <div className="p-12 text-center text-muted-foreground">
                No users found.
             </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Users;
