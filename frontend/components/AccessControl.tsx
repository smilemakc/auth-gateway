import React, { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { ArrowLeft, Plus, Edit2, Trash2, Shield, Lock, Search } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useRoles, useDeleteRole, usePermissions, useDeletePermission } from '../hooks/useRBAC';

type TabType = 'roles' | 'permissions';

const AccessControl: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [searchParams, setSearchParams] = useSearchParams();

  const initialTab = (searchParams.get('tab') as TabType) || 'roles';
  const [activeTab, setActiveTab] = useState<TabType>(initialTab);
  const [searchTerm, setSearchTerm] = useState('');

  // Roles data
  const { data: roles = [], isLoading: rolesLoading, error: rolesError } = useRoles();
  const deleteRoleMutation = useDeleteRole();

  // Permissions data
  const { data: permissions = [], isLoading: permissionsLoading, error: permissionsError } = usePermissions();
  const deletePermissionMutation = useDeletePermission();

  const handleTabChange = (tab: TabType) => {
    setActiveTab(tab);
    setSearchParams({ tab });
    setSearchTerm('');
  };

  const handleDeleteRole = async (id: string) => {
    if (window.confirm(t('common.confirm_delete'))) {
      try {
        await deleteRoleMutation.mutateAsync(id);
      } catch (err) {
        console.error('Failed to delete role:', err);
      }
    }
  };

  const handleDeletePermission = async (id: string) => {
    if (window.confirm(t('common.confirm_delete'))) {
      try {
        await deletePermissionMutation.mutateAsync(id);
      } catch (err) {
        console.error('Failed to delete permission:', err);
      }
    }
  };

  const filteredRoles = roles.filter(r =>
    (r.display_name || r.name).toLowerCase().includes(searchTerm.toLowerCase()) ||
    r.description?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const filteredPermissions = permissions.filter(p =>
    p.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    p.resource.toLowerCase().includes(searchTerm.toLowerCase()) ||
    p.action.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const isLoading = activeTab === 'roles' ? rolesLoading : permissionsLoading;
  const error = activeTab === 'roles' ? rolesError : permissionsError;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 text-destructive">
        Failed to load data. Please try again.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/settings')}
            className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h1 className="text-2xl font-bold text-foreground">{t('settings.roles_desc')}</h1>
            <p className="text-sm text-muted-foreground mt-1">
              {activeTab === 'roles'
                ? 'Manage user roles and their permission sets'
                : 'Create and manage granular permissions'}
            </p>
          </div>
        </div>
        <Link
          to={activeTab === 'roles' ? '/settings/access-control/roles/new' : '/settings/access-control/permissions/new'}
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('common.create')}
        </Link>
      </div>

      {/* Tabs and Search */}
      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="border-b border-border">
          <div className="flex items-center justify-between px-4">
            <div className="flex">
              <button
                onClick={() => handleTabChange('roles')}
                className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'roles'
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                }`}
              >
                <Shield size={18} />
                {t('roles.title')}
                <span className={`ml-1 px-2 py-0.5 rounded-full text-xs ${
                  activeTab === 'roles' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'
                }`}>
                  {roles.length}
                </span>
              </button>
              <button
                onClick={() => handleTabChange('permissions')}
                className={`flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'permissions'
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                }`}
              >
                <Lock size={18} />
                {t('perms.title')}
                <span className={`ml-1 px-2 py-0.5 rounded-full text-xs ${
                  activeTab === 'permissions' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'
                }`}>
                  {permissions.length}
                </span>
              </button>
            </div>
            <div className="relative">
              <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder={t('common.search')}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 pr-4 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring w-64"
              />
            </div>
          </div>
        </div>

        {/* Roles Tab Content */}
        {activeTab === 'roles' && (
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
                {filteredRoles.map((role) => (
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
                      {role.created_at ? new Date(role.created_at).toLocaleDateString() : '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex justify-end gap-2">
                        <Link
                          to={`/settings/access-control/roles/${role.id}`}
                          className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                        >
                          <Edit2 size={18} />
                        </Link>
                        {!role.is_system_role && (
                          <button
                            onClick={() => handleDeleteRole(role.id)}
                            className="p-1 text-muted-foreground hover:text-destructive rounded-md hover:bg-accent"
                          >
                            <Trash2 size={18} />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
                {filteredRoles.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-6 py-12 text-center text-muted-foreground">
                      No roles found.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}

        {/* Permissions Tab Content */}
        {activeTab === 'permissions' && (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border">
              <thead className="bg-muted">
                <tr>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('perms.name')}</th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('perms.resource')}</th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('perms.action')}</th>
                  <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Description</th>
                  <th scope="col" className="relative px-6 py-3"><span className="sr-only">Actions</span></th>
                </tr>
              </thead>
              <tbody className="bg-card divide-y divide-border">
                {filteredPermissions.map((perm) => (
                  <tr key={perm.id} className="hover:bg-accent transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-3">
                        <div className="p-1.5 rounded-lg bg-muted text-muted-foreground">
                          <Lock size={16} />
                        </div>
                        <span className="font-medium text-foreground">{perm.name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary font-mono">
                        {perm.resource}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-accent text-accent-foreground font-mono">
                        {perm.action}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-sm text-muted-foreground">{perm.description}</span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex justify-end gap-2">
                        <Link
                          to={`/settings/access-control/permissions/${perm.id}`}
                          className="p-1 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                        >
                          <Edit2 size={18} />
                        </Link>
                        <button
                          onClick={() => handleDeletePermission(perm.id)}
                          className="p-1 text-muted-foreground hover:text-destructive rounded-md hover:bg-accent"
                        >
                          <Trash2 size={18} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
                {filteredPermissions.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-6 py-12 text-center text-muted-foreground">
                      No permissions found.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default AccessControl;
