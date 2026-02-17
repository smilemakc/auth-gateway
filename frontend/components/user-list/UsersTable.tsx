import React from 'react';
import { Link } from 'react-router-dom';
import { Shield, ShieldOff, Check, X, Eye, Edit } from 'lucide-react';
import type { AdminUserResponse } from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';
import { formatDate } from '../../lib/date';
import type { SortState } from '../../hooks/useSort';
import SortableHeader from '../SortableHeader';

interface UsersTableProps {
  users: AdminUserResponse[];
  sortState: SortState;
  onSort: (key: string) => void;
  onToggleStatus: (id: string, currentStatus: boolean) => void;
  isToggling: boolean;
}

export const UsersTable: React.FC<UsersTableProps> = ({
  users,
  sortState,
  onSort,
  onToggleStatus,
  isToggling,
}) => {
  const { t } = useLanguage();

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-border">
        <thead className="bg-muted">
          <tr>
            <SortableHeader label={t('users.col_user')} sortKey="email" currentSortKey={sortState.key} currentDirection={sortState.direction} onSort={onSort} />
            <SortableHeader label={t('users.col_role')} sortKey="roles" currentSortKey={sortState.key} currentDirection={sortState.direction} onSort={onSort} />
            <SortableHeader label={t('users.col_status')} sortKey="is_active" currentSortKey={sortState.key} currentDirection={sortState.direction} onSort={onSort} />
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('users.col_2fa')}</th>
            <SortableHeader label={t('users.col_created')} sortKey="created_at" currentSortKey={sortState.key} currentDirection={sortState.direction} onSort={onSort} />
            <th scope="col" className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody className="bg-card divide-y divide-border">
          {users.map((user) => (
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
                    onClick={() => onToggleStatus(user.id, user.is_active)}
                    className={`p-1 rounded-md hover:bg-accent ${user.is_active ? 'text-muted-foreground hover:text-destructive' : 'text-muted-foreground hover:text-success'}`}
                    disabled={isToggling}
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
    </div>
  );
};
