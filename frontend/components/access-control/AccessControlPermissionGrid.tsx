import React from 'react';
import { AlertCircle, Lock } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { Permission } from '../../types';
import AccessControlResourceGroup from './AccessControlResourceGroup';

interface AccessControlPermissionGridProps {
  groupedPermissions: Record<string, Permission[]>;
  selectedPermIds: string[];
  pendingUpdate: boolean;
  onTogglePermission: (permissionId: string) => void;
  onSelectAllResource: (resource: string, select: boolean) => void;
}

const AccessControlPermissionGrid: React.FC<AccessControlPermissionGridProps> = ({
  groupedPermissions,
  selectedPermIds,
  pendingUpdate,
  onTogglePermission,
  onSelectAllResource,
}) => {
  const { t } = useLanguage();

  return (
    <div className="border-t border-border bg-muted/30">
      <div className="p-4">
        <div className="flex items-center justify-between mb-4">
          <h4 className="font-medium text-foreground flex items-center gap-2">
            <Lock size={16} className="text-primary" />
            {t('roles.permissions')}
          </h4>
          <p className="text-xs text-muted-foreground">
            {t('access_control.toggle_hint')}
          </p>
        </div>

        {Object.keys(groupedPermissions).length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <AlertCircle size={32} className="mx-auto mb-2 opacity-50" />
            <p>{t('access_control.no_permissions_hint')}</p>
          </div>
        ) : (
          <div className="space-y-6">
            {(Object.entries(groupedPermissions) as [string, Permission[]][]).map(([resource, perms]) => (
              <AccessControlResourceGroup
                key={resource}
                resource={resource}
                permissions={perms}
                selectedPermIds={selectedPermIds}
                pendingUpdate={pendingUpdate}
                onTogglePermission={onTogglePermission}
                onSelectAll={onSelectAllResource}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default AccessControlPermissionGrid;
