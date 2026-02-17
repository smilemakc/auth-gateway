
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Permission } from '../../types';
import { ArrowLeft, Save, AlertCircle } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useRoleDetail, useCreateRole, useUpdateRole, usePermissions } from '../../hooks/rbac';
import { useApplication } from '../../services/appContext';
import { RolePermissionsMatrix } from './RolePermissionsMatrix';

interface RoleFormState {
  name: string;
  display_name: string;
  description: string;
  permissions: string[];
  is_system_role: boolean;
}

const RoleEditor: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { currentApplication } = useApplication();
  const isEditMode = id && id !== 'new';
  const isNewMode = id === 'new';

  const [role, setRole] = useState<RoleFormState>({
    name: '',
    display_name: '',
    description: '',
    permissions: [],
    is_system_role: false
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const { data: existingRole, isLoading: roleLoading } = useRoleDetail(isEditMode ? id! : '');
  const { data: availablePermissions = [], isLoading: permissionsLoading } = usePermissions();
  const createMutation = useCreateRole();
  const updateMutation = useUpdateRole();

  useEffect(() => {
    if (isEditMode && existingRole) {
      setRole({
        name: existingRole.name,
        display_name: existingRole.display_name || existingRole.name,
        description: existingRole.description || '',
        permissions: existingRole.permissions?.map(p => p.id) || [],
        is_system_role: existingRole.is_system_role
      });
    }
  }, [existingRole, isEditMode]);

  const handlePermissionToggle = (permId: string) => {
    setRole(prev => {
      const currentPerms = prev.permissions || [];
      if (currentPerms.includes(permId)) {
        return { ...prev, permissions: currentPerms.filter(p => p !== permId) };
      } else {
        return { ...prev, permissions: [...currentPerms, permId] };
      }
    });
  };

  const handleSelectAllResource = (resource: string, select: boolean) => {
    const resourcePerms = availablePermissions.filter(p => p.resource === resource).map(p => p.id);
    setRole(prev => {
      let newPerms = prev.permissions || [];
      if (select) {
        const toAdd = resourcePerms.filter(p => !newPerms.includes(p));
        newPerms = [...newPerms, ...toAdd];
      } else {
        newPerms = newPerms.filter(p => !resourcePerms.includes(p));
      }
      return { ...prev, permissions: newPerms };
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    const displayName = role.display_name || role.name;
    if (!displayName) {
      setError(t('role_edit.err_name_required'));
      setLoading(false);
      return;
    }

    try {
      if (isNewMode) {
        await createMutation.mutateAsync({
          name: displayName.toLowerCase().replace(/\s+/g, '_'),
          display_name: displayName,
          description: role.description,
          permissions: role.permissions || []
        });
      } else if (id) {
        await updateMutation.mutateAsync({
          id,
          data: {
            display_name: displayName,
            description: role.description,
            permissions: role.permissions
          }
        });
      }
      navigate('/settings/access-control?tab=roles');
    } catch (err: any) {
      setError(err.message || t('role_edit.save_error'));
    } finally {
      setLoading(false);
    }
  };

  const isPageLoading = (isEditMode && roleLoading) || permissionsLoading;

  if (isPageLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  const groupedPermissions = availablePermissions.reduce((acc, perm) => {
    if (!acc[perm.resource]) acc[perm.resource] = [];
    acc[perm.resource].push(perm);
    return acc;
  }, {} as Record<string, Permission[]>);

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/settings/access-control?tab=roles')}
          className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-foreground">{isNewMode ? t('common.create') : t('common.edit')}</h1>
          {currentApplication && isNewMode && (
            <p className="text-sm text-muted-foreground mt-1">
              {t('role_edit.created_for')}: <span className="font-medium text-foreground">{currentApplication.name}</span>
            </p>
          )}
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Role Details */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6 space-y-6">
           {error && (
            <div className="bg-destructive/10 border-l-4 border-destructive p-4 flex items-center">
              <AlertCircle className="h-5 w-5 text-destructive/60 mr-2" />
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}

          <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">{t('users.col_role')}</label>
              <input
                type="text"
                value={role.display_name || role.name}
                onChange={(e) => setRole(prev => ({ ...prev, display_name: e.target.value, name: e.target.value }))}
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
                placeholder="e.g. Content Manager"
                disabled={role.is_system_role}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">{t('common.description')}</label>
              <input
                type="text"
                value={role.description}
                onChange={(e) => setRole(prev => ({ ...prev, description: e.target.value }))}
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
              />
            </div>
          </div>
        </div>

        {/* Permissions Matrix */}
        <RolePermissionsMatrix
          groupedPermissions={groupedPermissions}
          selectedPermissions={role.permissions}
          onTogglePermission={handlePermissionToggle}
          onSelectAllResource={handleSelectAllResource}
        />

        {/* Footer Actions */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={() => navigate('/settings/access-control?tab=roles')}
            className="px-4 py-2 text-sm font-medium text-foreground bg-card border border-input rounded-lg hover:bg-accent focus:outline-none"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={loading}
            className={`flex items-center px-6 py-2 text-sm font-medium text-primary-foreground bg-primary border border-transparent rounded-lg hover:bg-primary-600 focus:outline-none
              ${loading ? 'opacity-70 cursor-not-allowed' : ''}`}
          >
            {loading ? (
               <span className="w-5 h-5 border-2 border-primary-foreground border-t-transparent rounded-full animate-spin mr-2"></span>
            ) : (
              <Save size={16} className="mr-2" />
            )}
            {t('common.save')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default RoleEditor;
