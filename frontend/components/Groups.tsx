import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Search, Plus, Edit, Trash2, Users, Eye } from 'lucide-react';
import type { Group } from '@auth-gateway/client-sdk';
import { useLanguage } from '../services/i18n';
import { useGroups, useDeleteGroup } from '../hooks/useGroups';

const Groups: React.FC = () => {
  const [page, setPage] = useState(1);
  const [searchTerm, setSearchTerm] = useState('');
  const pageSize = 20;
  const navigate = useNavigate();
  const { t } = useLanguage();

  const { data, isLoading, error } = useGroups(page, pageSize);
  const deleteGroup = useDeleteGroup();

  const handleDelete = async (id: string, name: string) => {
    if (window.confirm(`${t('groups.delete_confirm')} "${name}"?`)) {
      try {
        await deleteGroup.mutateAsync(id);
      } catch (error) {
        console.error('Failed to delete group:', error);
        alert(t('common.failed_to_load'));
      }
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
        <p className="text-destructive">{t('groups.error_loading')}: {(error as Error).message}</p>
      </div>
    );
  }

  const groups = data?.groups || [];
  const filteredGroups = searchTerm
    ? groups.filter(
        (g) =>
          g.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          g.display_name.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : groups;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h1 className="text-2xl font-bold text-foreground">{t('groups.title')}</h1>
        <button
          type="button"
          onClick={(e) => {
            e.preventDefault();
            navigate('/groups/new');
          }}
          className="bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
        >
          <Plus size={18} />
          {t('groups.create')}
        </button>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        {/* Search */}
        <div className="p-4 border-b border-border">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" size={20} />
            <input
              type="text"
              placeholder={t('groups.search')}
              className="w-full pl-10 pr-4 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('groups.col_name')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('groups.col_display_name')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('groups.col_description')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('groups.col_members')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('groups.col_created')}
                </th>
                <th scope="col" className="relative px-6 py-3">
                  <span className="sr-only">{t('groups.col_actions')}</span>
                </th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {filteredGroups.map((group) => (
                <tr key={group.id} className="hover:bg-accent transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-foreground">{group.name}</div>
                    {group.is_system_group && (
                      <span className="text-xs text-muted-foreground">{t('groups.system_group')}</span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-foreground">{group.display_name}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-muted-foreground max-w-xs truncate">
                      {group.description || '-'}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Users size={16} />
                      {group.member_count || 0}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {new Date(group.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <Link
                        to={`/groups/${group.id}`}
                        className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                        title={t('common.view_details')}
                      >
                        <Eye size={18} />
                      </Link>
                      <Link
                        to={`/groups/${group.id}/edit`}
                        className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                        title={t('common.edit')}
                      >
                        <Edit size={18} />
                      </Link>
                      {!group.is_system_group && (
                        <button
                          onClick={() => handleDelete(group.id, group.display_name)}
                          className="p-1 text-muted-foreground hover:text-destructive rounded-md hover:bg-accent"
                          title={t('common.delete')}
                          disabled={deleteGroup.isPending}
                        >
                          <Trash2 size={18} />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {filteredGroups.length === 0 && (
            <div className="p-12 text-center text-muted-foreground">{t('groups.no_groups')}</div>
          )}
        </div>

        {/* Pagination */}
        {data && data.total > pageSize && (
          <div className="px-6 py-4 border-t border-border flex items-center justify-between">
            <div className="text-sm text-muted-foreground">
              {t('common.showing')} {(page - 1) * pageSize + 1} {t('common.to')} {Math.min(page * pageSize, data.total)} {t('common.of')} {data.total} {t('groups.showing')}
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 border border-input rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
              >
                {t('common.previous')}
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * pageSize >= data.total}
                className="px-3 py-1 border border-input rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
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

export default Groups;

