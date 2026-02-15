
import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { RoleDefinition } from '../types';
import { ArrowLeft, Shield, Plus, Edit2, Trash2 } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useRoles, useDeleteRole } from '../hooks/useRBAC';
import { useApplication } from '../services/appContext';
import { formatDate } from '../lib/date';
import { confirm } from '../services/confirm';
import { logger } from '@/lib/logger';

const Roles: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { currentApplication } = useApplication();
  const { data: roles = [], isLoading, error } = useRoles();
  const deleteRoleMutation = useDeleteRole();

  const handleDelete = async (id: string) => {
    const ok = await confirm({
      title: t('confirm.delete_title'),
      description: t('common.confirm_delete'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await deleteRoleMutation.mutateAsync(id);
      } catch (err) {
        logger.error('Failed to delete role:', err);
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
        Failed to load roles. Please try again.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/settings')}
            className="p-2 hover:bg-card rounded-lg transition-colors text-muted-foreground"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-foreground">{t('roles.title')}</h1>
            {currentApplication && (
              <p className="text-sm text-muted-foreground mt-1">
                Roles for: <span className="font-medium text-foreground">{currentApplication.name}</span>
              </p>
            )}
          </div>
        </div>
        <Link
          to="/settings/roles/new"
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('common.create')}
        </Link>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('users.col_role')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Description</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('roles.permissions')}</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('common.created')}</th>
                <th scope="col" className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {roles.map((role) => (
                <tr key={role.id} className="hover:bg-accent transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-3">
                      <div className={`p-2 rounded-lg ${role.is_system_role ? 'bg-accent text-accent-foreground' : 'bg-muted text-muted-foreground'}`}>
                        <Shield size={18} />
                      </div>
                      <span className="font-medium text-foreground">{role.display_name || role.name}</span>
                      {role.is_system_role && (
                        <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-muted text-muted-foreground uppercase">{t('roles.system_role')}</span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="text-sm text-muted-foreground">{role.description}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary">
                      {role.permissions?.length || 0}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {role.created_at ? formatDate(role.created_at) : '-'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <Link
                        to={`/settings/roles/${role.id}`}
                        className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-muted"
                      >
                        <Edit2 size={18} />
                      </Link>
                      {!role.is_system_role && (
                        <button
                          onClick={() => handleDelete(role.id)}
                          className="p-1 text-muted-foreground hover:text-red-600 rounded-md hover:bg-muted"
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
        </div>
      </div>
    </div>
  );
};

export default Roles;
