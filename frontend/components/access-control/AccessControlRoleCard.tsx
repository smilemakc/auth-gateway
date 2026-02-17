import React, { useCallback, useState } from 'react';
import { AlertCircle, Check, ChevronDown, ChevronRight, Lock, Shield, Trash2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useDeleteRole, useUpdateRole } from '../../hooks/rbac';
import type { Permission, RoleDefinition } from '../../types';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';

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

    let newPermissions: string[];
    if (select) {
      newPermissions = [...new Set([...currentPermIds, ...resourcePermIds])];
    } else {
      newPermissions = currentPermIds.filter(id => !resourcePermIds.includes(id));
    }

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
      {/* Role Header */}
      <div
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-accent/50 transition-colors"
        onClick={onToggleExpand}
      >
        <div className="flex items-center gap-4">
          <div className={`p-2.5 rounded-xl ${role.is_system_role ? 'bg-primary/10' : 'bg-muted'}`}>
            <Shield size={20} className={role.is_system_role ? 'text-primary' : 'text-muted-foreground'}/>
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="font-semibold text-foreground">
                {getRoleDisplayName(role)}
              </h3>
              {role.is_system_role && (
                <span
                  className="px-2 py-0.5 rounded text-[10px] font-bold bg-primary/10 text-primary uppercase">
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
                <Trash2 size={18}/>
              </button>
            )}
            {isExpanded ? (
              <ChevronDown size={20} className="text-muted-foreground"/>
            ) : (
              <ChevronRight size={20} className="text-muted-foreground"/>
            )}
          </div>
        </div>
      </div>

      {/* Expanded Permissions */}
      {isExpanded && (
        <div className="border-t border-border bg-muted/30">
          <div className="p-4">
            <div className="flex items-center justify-between mb-4">
              <h4 className="font-medium text-foreground flex items-center gap-2">
                <Lock size={16} className="text-primary"/>
                {t('roles.permissions')}
              </h4>
              <p className="text-xs text-muted-foreground">
                {t('access_control.toggle_hint')}
              </p>
            </div>

            {Object.keys(groupedPermissions).length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <AlertCircle size={32} className="mx-auto mb-2 opacity-50"/>
                <p>{t('access_control.no_permissions_hint')}</p>
              </div>
            ) : (
              <div className="space-y-6">
                {(Object.entries(groupedPermissions) as [string, Permission[]][]).map(([resource, perms]) => {
                  const resourcePermIds = perms.map(p => p.id);
                  const selectedCount = resourcePermIds.filter(id => rolePermIds.includes(id)).length;
                  const allSelected = selectedCount === perms.length;

                  return (
                    <div key={resource} className="bg-card rounded-lg border border-border p-4">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2">
                          <h5 className="font-medium text-foreground capitalize">
                            {resource.replace(/_/g, ' ')}
                          </h5>
                          <span className="text-xs text-muted-foreground">
                            ({selectedCount}/{perms.length})
                          </span>
                        </div>
                        <button
                          onClick={() => handleSelectAllResource(resource, !allSelected)}
                          disabled={pendingUpdate}
                          className={`text-xs font-medium px-2 py-1 rounded transition-colors ${
                            allSelected
                              ? 'bg-primary/10 text-primary hover:bg-primary/20'
                              : 'bg-muted text-muted-foreground hover:bg-accent'
                          }`}
                        >
                          {allSelected ? t('access_control.deselect_all') : t('access_control.select_all')}
                        </button>
                      </div>

                      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2">
                        {perms.map(perm => {
                          const isChecked = rolePermIds.includes(perm.id);

                          return (
                            <button
                              key={perm.id}
                              onClick={() => handlePermissionToggle(perm.id)}
                              disabled={pendingUpdate}
                              className={`
                                flex items-center gap-2 px-3 py-2 rounded-lg text-left transition-all text-sm
                                ${isChecked
                                ? 'bg-primary text-primary-foreground shadow-sm'
                                : 'bg-muted/50 text-foreground hover:bg-muted border border-transparent hover:border-border'
                              }
                                ${pendingUpdate ? 'opacity-50 cursor-wait' : 'cursor-pointer'}
                              `}
                              title={perm.description || `${perm.resource}:${perm.action}`}
                            >
                              <div className={`
                                w-4 h-4 rounded flex items-center justify-center shrink-0
                                ${isChecked ? 'bg-primary-foreground/20' : 'bg-card border border-input'}
                              `}>
                                {isChecked && <Check size={12}/>}
                              </div>
                              <span className="truncate font-medium capitalize">
                                {perm.action}
                              </span>
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default AccessControlRoleCard;
