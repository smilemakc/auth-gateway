import React, { useCallback, useState } from 'react';
import { ChevronDown, ChevronRight, Shield, Trash2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useDeleteRole, useUpdateRole } from '../../hooks/rbac';
import type { Permission, RoleDefinition } from '../../types';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';
import AccessControlPermissionGrid from './AccessControlPermissionGrid';

interface AccessControlRoleCardProps {
  role: RoleDefinition;
  groupedPermissions: Record<string, Permission[]>;
  isExpanded: boolean;
  onToggleExpand: () => void;
}

function getRoleDisplayName(role: RoleDefinition): string {
  return role.display_name || role.name;
}

const AccessControlRoleCard: React.FC<AccessControlRoleCardProps> = ({
  role,
  groupedPermissions,
  isExpanded,
  onToggleExpand,
}) => {
  const { t } = useLanguage();
  const updateRoleMutation = useUpdateRole();
  const deleteRoleMutation = useDeleteRole();
  const [pendingUpdate, setPendingUpdate] = useState(false);

  const rolePermIds = role.permissions?.map(p => p.id) || [];

  const handlePermissionToggle = useCallback(async (permissionId: string) => {
    const currentPermIds = role.permissions?.map(p => p.id) || [];
    const hasPermission = currentPermIds.includes(permissionId);
    const newPermissions = hasPermission
      ? currentPermIds.filter(id => id !== permissionId)
      : [...currentPermIds, permissionId];

    try {
      setPendingUpdate(true);
      await updateRoleMutation.mutateAsync({
        id: role.id,
        data: {
          display_name: getRoleDisplayName(role),
          description: role.description,
          permissions: newPermissions,
        },
      });
    } catch (err) {
      logger.error('Failed to update role permissions:', err);
    } finally {
      setPendingUpdate(false);
    }
  }, [role, updateRoleMutation]);

  const handleSelectAllResource = useCallback(async (resource: string, select: boolean) => {
    const currentPermIds = role.permissions?.map(p => p.id) || [];
    const resourcePermIds = groupedPermissions[resource]?.map(p => p.id) || [];
    const newPermissions = select
      ? [...new Set([...currentPermIds, ...resourcePermIds])]
      : currentPermIds.filter(id => !resourcePermIds.includes(id));

    try {
      setPendingUpdate(true);
      await updateRoleMutation.mutateAsync({
        id: role.id,
        data: {
          display_name: getRoleDisplayName(role),
          description: role.description,
          permissions: newPermissions,
        },
      });
    } catch (err) {
      logger.error('Failed to update role permissions:', err);
    } finally {
      setPendingUpdate(false);
    }
  }, [role, groupedPermissions, updateRoleMutation]);

  const handleDeleteRole = useCallback(async () => {
    const ok = await confirm({
      description: t('common.confirm_delete'),
      title: t('confirm.delete_title'),
      variant: 'danger',
    });
    if (ok) {
      try {
        await deleteRoleMutation.mutateAsync(role.id);
      } catch (err) {
        logger.error('Failed to delete role:', err);
      }
    }
  }, [role.id, deleteRoleMutation, t]);

  return (
    <div className="bg-card border border-border rounded-xl overflow-hidden transition-shadow hover:shadow-md">
      <div
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-accent/50 transition-colors"
        onClick={onToggleExpand}
      >
        <div className="flex items-center gap-4">
          <div className={`p-2.5 rounded-xl ${role.is_system_role ? 'bg-primary/10' : 'bg-muted'}`}>
            <Shield size={20} className={role.is_system_role ? 'text-primary' : 'text-muted-foreground'} />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-semibold text-foreground">
                {getRoleDisplayName(role)}
              </h3>
              {role.is_system_role && (
                <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-primary/10 text-primary uppercase">
                  {t('roles.system_role')}
                </span>
              )}
            </div>
            <p className="text-sm text-muted-foreground mt-0.5">
              {role.description || t('common.no_description')}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-4">
          <div className="text-right mr-4">
            <p className="text-lg font-bold text-foreground">{rolePermIds.length}</p>
            <p className="text-xs text-muted-foreground">{t('access_control.permissions_count')}</p>
          </div>
          <div className="flex items-center gap-2">
            {!role.is_system_role && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  void handleDeleteRole();
                }}
                className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
              >
                <Trash2 size={18} />
              </button>
            )}
            {isExpanded ? (
              <ChevronDown size={20} className="text-muted-foreground" />
            ) : (
              <ChevronRight size={20} className="text-muted-foreground" />
            )}
          </div>
        </div>
      </div>

      {isExpanded && (
        <AccessControlPermissionGrid
          groupedPermissions={groupedPermissions}
          selectedPermIds={rolePermIds}
          pendingUpdate={pendingUpdate}
          onTogglePermission={handlePermissionToggle}
          onSelectAllResource={handleSelectAllResource}
        />
      )}
    </div>
  );
};

export default AccessControlRoleCard;
