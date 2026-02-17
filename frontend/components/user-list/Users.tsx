
import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Search, Filter, Users as UsersIcon } from 'lucide-react';
import type { AdminUserResponse } from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';
import { useUsers, useUpdateUser } from '../../hooks/useUsers';
import { useApplication } from '../../services/appContext';
import { useSort } from '../../hooks/useSort';
import { toast } from '../../services/toast';
import { logger } from '@/lib/logger';
import { UsersTable } from './UsersTable';

const Users: React.FC = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('all');
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { currentApplication } = useApplication();

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
      logger.error('Failed to toggle user status:', error);
      toast.error(t('users.status_update_error'));
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
        <p className="text-destructive">{t('users.loading_error')} {(error as Error).message}</p>
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
            {t('users.showing_for')} <span className="font-medium text-foreground">{currentApplication.name}</span>
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
        <UsersTable
          users={sortedUsers}
          sortState={sortState}
          onSort={requestSort}
          onToggleStatus={toggleStatus}
          isToggling={updateUserMutation.isPending}
        />

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
