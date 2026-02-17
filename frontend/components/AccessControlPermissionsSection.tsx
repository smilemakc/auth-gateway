import React, { useCallback, useMemo, useState } from 'react';
import { Loader2, Lock, Plus, X } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useCreatePermission, useDeletePermission } from '../hooks/rbac';
import type { Permission } from '../types';
import { confirm } from '../services/confirm';
import { logger } from '@/lib/logger';

const COMMON_ACTIONS = ['create', 'read', 'update', 'delete', 'list', 'manage', 'export', 'import'];

interface AccessControlPermissionsSectionProps {
  permissions: Permission[];
  groupedPermissions: Record<string, Permission[]>;
}

const AccessControlPermissionsSection: React.FC<AccessControlPermissionsSectionProps> = ({
  permissions,
  groupedPermissions,
}) => {
  const { t } = useLanguage();
  const createPermissionMutation = useCreatePermission();
  const deletePermissionMutation = useDeletePermission();

  const [showCreatePermission, setShowCreatePermission] = useState(false);
  const [newPermResource, setNewPermResource] = useState('');
  const [newPermAction, setNewPermAction] = useState('');
  const [newPermDescription, setNewPermDescription] = useState('');

  const existingResources = useMemo(() => {
    return [...new Set(permissions.map(p => p.resource))].sort();
  }, [permissions]);

  const handleCreatePermission = useCallback(async () => {
    if (!newPermResource.trim() || !newPermAction.trim()) return;

    try {
      await createPermissionMutation.mutateAsync({
        name: `${newPermResource}:${newPermAction}`,
        resource: newPermResource.toLowerCase().replace(/\s+/g, '_'),
        action: newPermAction.toLowerCase().replace(/\s+/g, '_'),
        description: newPermDescription,
      });
      setNewPermResource('');
      setNewPermAction('');
      setNewPermDescription('');
      setShowCreatePermission(false);
    } catch (err) {
      logger.error('Failed to create permission:', err);
    }
  }, [newPermResource, newPermAction, newPermDescription, createPermissionMutation]);

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
            <Lock size={20} className="text-accent-foreground"/>
          </div>
          <div>
            <h3 className="font-semibold text-foreground">{t('perms.title')}</h3>
            <p className="text-sm text-muted-foreground">
              {permissions.length} {t('access_control.across_resources').replace('{count}', String(Object.keys(groupedPermissions).length))}
            </p>
          </div>
        </div>
        <button
          onClick={() => setShowCreatePermission(!showCreatePermission)}
          className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
            showCreatePermission
              ? 'bg-muted text-foreground'
              : 'bg-primary text-primary-foreground hover:bg-primary/90'
          }`}
        >
          {showCreatePermission ? (
            <>
              <X size={16}/>
              {t('common.cancel')}
            </>
          ) : (
            <>
              <Plus size={16}/>
              {t('access_control.add_permission')}
            </>
          )}
        </button>
      </div>

      {showCreatePermission && (
        <div className="p-4 bg-muted/30 border-b border-border">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                {t('access_control.resource_label')} *
              </label>
              <div className="relative">
                <input
                  type="text"
                  value={newPermResource}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewPermResource(e.target.value)}
                  placeholder={t('perm_edit.resource_placeholder')}
                  list="resources-list"
                  className="w-full px-3 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none text-sm"
                />
                <datalist id="resources-list">
                  {existingResources.map((r: string) => (
                    <option key={r} value={r}/>
                  ))}
                </datalist>
              </div>
              {existingResources.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {existingResources.slice(0, 5).map((r: string) => (
                    <button
                      key={r}
                      type="button"
                      onClick={() => setNewPermResource(r)}
                      className="px-2 py-0.5 text-xs bg-muted hover:bg-accent rounded transition-colors"
                    >
                      {r}
                    </button>
                  ))}
                </div>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                {t('access_control.action_label')} *
              </label>
              <input
                type="text"
                value={newPermAction}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewPermAction(e.target.value)}
                placeholder={t('perm_edit.action_placeholder')}
                list="actions-list"
                className="w-full px-3 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none text-sm"
              />
              <datalist id="actions-list">
                {COMMON_ACTIONS.map(a => (
                  <option key={a} value={a}/>
                ))}
              </datalist>
              <div className="flex flex-wrap gap-1 mt-2">
                {COMMON_ACTIONS.slice(0, 5).map(a => (
                  <button
                    key={a}
                    type="button"
                    onClick={() => setNewPermAction(a)}
                    className="px-2 py-0.5 text-xs bg-muted hover:bg-accent rounded transition-colors"
                  >
                    {a}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                {t('common.description')}
              </label>
              <input
                type="text"
                value={newPermDescription}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewPermDescription(e.target.value)}
                placeholder={t('common.description')}
                className="w-full px-3 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none text-sm"
              />
            </div>
          </div>
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">
              {t('access_control.perm_name_will_be')}: <code className="bg-muted px-1.5 py-0.5 rounded">
              {newPermResource || 'resource'}:{newPermAction || 'action'}
            </code>
            </p>
            <button
              onClick={handleCreatePermission}
              disabled={!newPermResource.trim() || !newPermAction.trim() || createPermissionMutation.isPending}
              className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50"
            >
              {createPermissionMutation.isPending && <Loader2 size={16} className="animate-spin"/>}
              {t('access_control.create_permission')}
            </button>
          </div>
        </div>
      )}

      <div className="p-4">
        {Object.keys(groupedPermissions).length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <Lock size={32} className="mx-auto mb-2 opacity-50"/>
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
                        <X size={14}/>
                      </button>
                    </div>
                  ))}
                  <button
                    onClick={() => {
                      setNewPermResource(resource);
                      setShowCreatePermission(true);
                    }}
                    className="flex items-center gap-1 px-3 py-1.5 border border-dashed border-border hover:border-primary hover:text-primary rounded-lg text-sm text-muted-foreground transition-colors"
                  >
                    <Plus size={14}/>
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
