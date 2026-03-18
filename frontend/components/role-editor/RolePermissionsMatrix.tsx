import React from 'react';
import { Shield, Check } from 'lucide-react';
import { Permission } from '../../types';
import { useLanguage } from '../../services/i18n';

interface RolePermissionsMatrixProps {
  groupedPermissions: Record<string, Permission[]>;
  selectedPermissions: string[];
  onTogglePermission: (permId: string) => void;
  onSelectAllResource: (resource: string, select: boolean) => void;
}

export const RolePermissionsMatrix: React.FC<RolePermissionsMatrixProps> = ({
  groupedPermissions,
  selectedPermissions,
  onTogglePermission,
  onSelectAllResource,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="p-6 border-b border-border">
        <h2 className="text-lg font-bold text-foreground flex items-center gap-2">
          <Shield size={20} className="text-primary" />
          {t('roles.permissions')}
        </h2>
      </div>

      <div className="divide-y divide-border">
        {Object.entries(groupedPermissions).map(([resource, permissions]: [string, Permission[]]) => {
          const allSelected = permissions.every(p => selectedPermissions?.includes(p.id));

          return (
            <div key={resource} className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-md font-bold text-foreground capitalize flex items-center gap-2">
                    {resource.replace(/_/g, ' ')}
                </h3>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => onSelectAllResource(resource, !allSelected)}
                    className="text-xs text-primary hover:text-primary/80 font-medium"
                  >
                    {allSelected ? t('role_edit.deselect_resource') : t('role_edit.select_resource')}
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {permissions.map(perm => {
                  const isChecked = selectedPermissions?.includes(perm.id);
                  return (
                    <div
                      key={perm.id}
                      onClick={() => onTogglePermission(perm.id)}
                      className={`
                        cursor-pointer border rounded-lg p-3 flex items-start gap-3 transition-all relative overflow-hidden
                        ${isChecked
                          ? 'bg-primary/10 border-primary/20 ring-1 ring-primary/20'
                          : 'bg-card border-border hover:border-input'}
                      `}
                    >
                      <div className={`mt-0.5 w-5 h-5 rounded border flex items-center justify-center flex-shrink-0 transition-colors ${isChecked ? 'bg-primary border-primary' : 'border-input bg-card'}`}>
                        {isChecked && <Check size={12} className="text-primary-foreground" />}
                      </div>
                      <div className="flex-1">
                        <p className={`text-sm font-semibold capitalize ${isChecked ? 'text-primary' : 'text-foreground'}`}>{perm.action}</p>
                        <p className="text-xs text-muted-foreground mt-0.5">{perm.description}</p>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
