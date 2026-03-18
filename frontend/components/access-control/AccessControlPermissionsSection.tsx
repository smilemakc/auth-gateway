import React, { useCallback, useMemo, useState } from 'react';
import { Lock, Plus, X } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useDeletePermission } from '../../hooks/rbac';
import type { Permission } from '../../types';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';
import AccessControlPermissionForm from './AccessControlPermissionForm';

interface AccessControlPermissionsSectionProps {
  permissions: Permission[];
  groupedPermissions: Record<string, Permission[]>;
}

const AccessControlPermissionsSection: React.FC<AccessControlPermissionsSectionProps> = ({
  permissions,
  groupedPermissions,
}) => {
  const { t } = useLanguage();
  const deletePermissionMutation = useDeletePermission();

  const [showCreatePermission, setShowCreatePermission] = useState(false);
  const [initialResource, setInitialResource] = useState('');

  const existingResources = useMemo(() => {
    return [...new Set(permissions.map(p => p.resource))].sort();
  }, [permissions]);

  const handleDeletePermission = useCallback(async (perm: Permission) => {
    const ok = await confirm({
      description: `${t('access_control.confirm_delete_perm')} "${perm.name}"?`,
      title: t('confirm.delete_title'),
      variant: 'danger',
    });
    if (ok) {
      try {
        await deletePermissionMutation.mutateAsync(perm.id);
      } catch (err) {
        logger.error('Failed to delete permission:', err);
      }
    }
  }, [deletePermissionMutation, t]);

  return (
    <div className="bg-card border border-border rounded-xl overflow-hidden">
      <div className="p-4 border-b border-border flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-accent rounded-lg">
            <Lock size={20} className="text-accent-foreground" />
          </div>
          <div>
            <h3 className="font-semibold text-foreground">{t('perms.title')}</h3>
            <p className="text-sm text-muted-foreground">
              {permissions.length} {t('access_control.across_resources').replace('{count}', String(Object.keys(groupedPermissions).length))}
            </p>
          </div>
        </div>
        <button
          onClick={() => {
            if (showCreatePermission) setInitialResource('');
            setShowCreatePermission(!showCreatePermission);
          }}
          className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
            showCreatePermission
              ? 'bg-muted text-foreground'
              : 'bg-primary text-primary-foreground hover:bg-primary/90'
          }`}
        >
          {showCreatePermission ? (
            <>
              <X size={16} />
              {t('common.cancel')}
            </>
          ) : (
            <>
              <Plus size={16} />
              {t('access_control.add_permission')}
            </>
          )}
        </button>
      </div>

      {showCreatePermission && (
        <AccessControlPermissionForm
          existingResources={existingResources}
          initialResource={initialResource}
          onClose={() => {
            setShowCreatePermission(false);
            setInitialResource('');
          }}
        />
      )}

      <div className="p-4">
        {Object.keys(groupedPermissions).length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <Lock size={32} className="mx-auto mb-2 opacity-50" />
            <p>{t('access_control.no_permissions_yet')}</p>
            <button
              onClick={() => setShowCreatePermission(true)}
              className="mt-2 text-primary hover:underline text-sm font-medium"
            >
              {t('access_control.create_first_permission')}
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            {(Object.entries(groupedPermissions) as [string, Permission[]][]).map(([resource, perms]) => (
              <div key={resource} className="border border-border rounded-lg p-3">
                <div className="flex items-center justify-between mb-2">
                  <h4 className="font-medium text-foreground capitalize flex items-center gap-2">
                    <span className="px-2 py-0.5 bg-primary/10 text-primary rounded text-xs font-mono">
                      {resource}
                    </span>
                    <span className="text-xs text-muted-foreground font-normal">
                      {perms.length} {perms.length === 1 ? t('access_control.permission_singular') : t('access_control.permissions_plural')}
                    </span>
                  </h4>
                </div>
                <div className="flex flex-wrap gap-2">
                  {perms.map(perm => (
                    <div
                      key={perm.id}
                      className="group flex items-center gap-2 px-3 py-1.5 bg-muted/50 hover:bg-muted rounded-lg text-sm transition-colors"
                      title={perm.description || `${perm.resource}:${perm.action}`}
                    >
                      <span className="font-medium capitalize">{perm.action}</span>
                      <button
                        onClick={() => void handleDeletePermission(perm)}
                        className="opacity-0 group-hover:opacity-100 p-0.5 text-muted-foreground hover:text-destructive transition-all"
                      >
                        <X size={14} />
                      </button>
                    </div>
                  ))}
                  <button
                    onClick={() => {
                      setInitialResource(resource);
                      setShowCreatePermission(true);
                    }}
                    className="flex items-center gap-1 px-3 py-1.5 border border-dashed border-border hover:border-primary hover:text-primary rounded-lg text-sm text-muted-foreground transition-colors"
                  >
                    <Plus size={14} />
                    {t('access_control.add')}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default AccessControlPermissionsSection;
