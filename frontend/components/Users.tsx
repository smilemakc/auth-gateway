
import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Search, Shield, ShieldOff, Check, X, Filter, Eye, Edit, Users as UsersIcon } from 'lucide-react';
import type { AdminUserResponse } from '@auth-gateway/client-sdk';
import { useLanguage } from '../services/i18n';
import { useUsers, useUpdateUser } from '../hooks/useUsers';
import { useApplication } from '../services/appContext';
import { formatDate } from '../lib/date';
import { useSort } from '../hooks/useSort';
import SortableHeader from './SortableHeader';
import { toast } from '../services/toast';

const Users: React.FC = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('all');
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { currentApplication } = useApplication();

  // Fetch users with React Query
  const { data, isLoading, error } = useUsers(page, pageSize);
  const updateUserMutation = useUpdateUser();
  const { sortState, requestSort, sortData } = useSort<AdminUserResponse>();

  useEffect(() => {
    setPage(1);
  }, [roleFilter]);

  const toggleStatus = async (id: string, currentStatus: boolean) => {
    try {
      await updateUserMutation.mutateAsync({
        id,
        data: { is_active: !currentStatus },
      });
    } catch (error) {
      console.error('Failed to toggle user status:', error);
      toast.error('Failed to update user status');
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
  const total = data?.total || 0;
  const totalPages = Math.ceil(total / pageSize);

  const filterUsers = (userList: AdminUserResponse[]) => {
    let filtered = userList;
    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      filtered = filtered.filter((u) =>
        u.username?.toLowerCase().includes(term) ||
        u.email?.toLowerCase().includes(term) ||
        u.full_name?.toLowerCase().includes(term)
      );
    }
    if (roleFilter !== 'all') {
      filtered = filtered.filter((u) =>
        u.roles?.some((r) => r.name === roleFilter)
      );
    }
    return filtered;
  };

  const filteredUsers = filterUsers(users);
  const sortedUsers = sortData(filteredUsers);

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

      {currentApplication && (
        <div className="px-4 py-2 bg-muted/50 rounded-lg border border-border">
          <span className="text-sm text-muted-foreground">
            Showing users for: <span className="font-medium text-foreground">{currentApplication.name}</span>
          </span>
        </div>
      )}

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
                <SortableHeader label={t('users.col_user')} sortKey="email" currentSortKey={sortState.key} currentDirection={sortState.direction} onSort={requestSort} />
                <SortableHeader label={t('users.col_role')} sortKey="roles" currentSortKey={sortState.key} currentDirection={sortState.direction} onSort={requestSort} />
                <SortableHeader label={t('users.col_status')} sortKey="is_active" currentSortKey={sortState.key} currentDirection={sortState.direction} onSort={requestSort} />
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('users.col_2fa')}</th>
                <SortableHeader label={t('users.col_created')} sortKey="created_at" currentSortKey={sortState.key} currentDirection={sortState.direction} onSort={requestSort} />
                <th scope="col" className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {sortedUsers.map((user) => (
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
                    {formatDate(user.created_at)}
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
                        title={user.is_active ? t('users.block_user') : t('users.unblock_user')}
                      >
                        {user.is_active ? <ShieldOff size={18} /> : <Shield size={18} />}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {filteredUsers.length === 0 && !isLoading && (
            <div className="text-center py-12 bg-card rounded-xl border border-border">
              <UsersIcon className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-2 text-sm font-semibold text-foreground">{t('users.no_users')}</h3>
              <p className="mt-1 text-sm text-muted-foreground">{t('users.no_users_desc')}</p>
              <div className="mt-6">
                <Link to="/users/new" className="inline-flex items-center gap-2 bg-primary text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary/90">
                  {t('users.create_user')}
                </Link>
              </div>
            </div>
          )}
        </div>

        {totalPages > 1 && (
          <div className="flex items-center justify-between bg-card px-4 py-3 rounded-lg border border-border mt-4">
            <div className="text-sm text-muted-foreground">
              {t('common.showing')} <span className="font-medium">{(page - 1) * pageSize + 1}</span> {t('common.to')}{' '}
              <span className="font-medium">{Math.min(page * pageSize, total)}</span> {t('common.of')}{' '}
              <span className="font-medium">{total}</span>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
              >
                {t('common.previous')}
              </button>
              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
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

export default Users;
