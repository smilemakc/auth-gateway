import React, { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Plus,
  Edit2,
  Trash2,
  Shield,
  Lock,
  Search,
  ChevronDown,
  ChevronRight,
  Check,
  X,
  Users,
  Loader2,
  AlertCircle
} from 'lucide-react';
import { useLanguage } from '../services/i18n';
import {
  useRoles,
  useDeleteRole,
  usePermissions,
  useDeletePermission,
  useUpdateRole,
  useCreateRole,
  useCreatePermission
} from '../hooks/useRBAC';
import type { Permission, RoleDefinition } from '../types';

const AccessControl: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();

  const [searchTerm, setSearchTerm] = useState('');
  const [expandedRoles, setExpandedRoles] = useState<Set<string>>(new Set());
  const [editingRole, setEditingRole] = useState<string | null>(null);
  const [showCreateRole, setShowCreateRole] = useState(false);
  const [newRoleName, setNewRoleName] = useState('');
  const [newRoleDescription, setNewRoleDescription] = useState('');

  // Permission creation state
  const [showCreatePermission, setShowCreatePermission] = useState(false);
  const [newPermResource, setNewPermResource] = useState('');
  const [newPermAction, setNewPermAction] = useState('');
  const [newPermDescription, setNewPermDescription] = useState('');

  // Data fetching
  const { data: roles = [], isLoading: rolesLoading } = useRoles();
  const { data: permissions = [], isLoading: permissionsLoading } = usePermissions();
  const deleteRoleMutation = useDeleteRole();
  const deletePermissionMutation = useDeletePermission();
  const updateRoleMutation = useUpdateRole();
  const createRoleMutation = useCreateRole();
  const createPermissionMutation = useCreatePermission();

  // Group permissions by resource
  const groupedPermissions = useMemo(() => {
    return permissions.reduce((acc, perm) => {
      if (!acc[perm.resource]) acc[perm.resource] = [];
      acc[perm.resource].push(perm);
      return acc;
    }, {} as Record<string, Permission[]>);
  }, [permissions]);

  const filteredRoles = roles.filter(r =>
    (r.display_name || r.name).toLowerCase().includes(searchTerm.toLowerCase()) ||
    r.description?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const toggleRoleExpand = (roleId: string) => {
    setExpandedRoles(prev => {
      const newSet = new Set(prev);
      if (newSet.has(roleId)) {
        newSet.delete(roleId);
      } else {
        newSet.add(roleId);
      }
      return newSet;
    });
  };

  const handlePermissionToggle = async (role: RoleDefinition, permissionId: string) => {
    const currentPermIds = role.permissions?.map(p => p.id) || [];
    const hasPermission = currentPermIds.includes(permissionId);

    const newPermissions = hasPermission
      ? currentPermIds.filter(id => id !== permissionId)
      : [...currentPermIds, permissionId];

    try {
      await updateRoleMutation.mutateAsync({
        id: role.id,
        data: {
          display_name: role.display_name || role.name,
          description: role.description,
          permissions: newPermissions
        }
      });
    } catch (err) {
      console.error('Failed to update role permissions:', err);
    }
  };

  const handleSelectAllResource = async (role: RoleDefinition, resource: string, select: boolean) => {
    const currentPermIds = role.permissions?.map(p => p.id) || [];
    const resourcePermIds = groupedPermissions[resource]?.map(p => p.id) || [];

    let newPermissions: string[];
    if (select) {
      const toAdd = resourcePermIds.filter(id => !currentPermIds.includes(id));
      newPermissions = [...currentPermIds, ...toAdd];
    } else {
      newPermissions = currentPermIds.filter(id => !resourcePermIds.includes(id));
    }

    try {
      await updateRoleMutation.mutateAsync({
        id: role.id,
        data: {
          display_name: role.display_name || role.name,
          description: role.description,
          permissions: newPermissions
        }
      });
    } catch (err) {
      console.error('Failed to update role permissions:', err);
    }
  };

  const handleCreateRole = async () => {
    if (!newRoleName.trim()) return;

    try {
      await createRoleMutation.mutateAsync({
        name: newRoleName.toLowerCase().replace(/\s+/g, '_'),
        display_name: newRoleName,
        description: newRoleDescription,
        permissions: []
      });
      setNewRoleName('');
      setNewRoleDescription('');
      setShowCreateRole(false);
    } catch (err) {
      console.error('Failed to create role:', err);
    }
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

  const handleCreatePermission = async () => {
    if (!newPermResource.trim() || !newPermAction.trim()) return;

    try {
      await createPermissionMutation.mutateAsync({
        name: `${newPermResource}:${newPermAction}`,
        resource: newPermResource.toLowerCase().replace(/\s+/g, '_'),
        action: newPermAction.toLowerCase().replace(/\s+/g, '_'),
        description: newPermDescription
      });
      setNewPermResource('');
      setNewPermAction('');
      setNewPermDescription('');
      setShowCreatePermission(false);
    } catch (err) {
      console.error('Failed to create permission:', err);
    }
  };

  // Get unique resources for quick selection
  const existingResources = useMemo(() => {
    return [...new Set(permissions.map(p => p.resource))].sort();
  }, [permissions]);

  const commonActions = ['create', 'read', 'update', 'delete', 'list', 'manage', 'export', 'import'];

  const isLoading = rolesLoading || permissionsLoading;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
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
            <h1 className="text-2xl font-bold text-foreground">
              {t('nav.access_settings') || 'Access Settings'}
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              Manage roles and assign permissions to control access
            </p>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-card border border-border rounded-xl p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary/10 rounded-lg">
              <Shield className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="text-2xl font-bold text-foreground">{roles.length}</p>
              <p className="text-sm text-muted-foreground">{t('roles.title')}</p>
            </div>
          </div>
        </div>
        <div className="bg-card border border-border rounded-xl p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-accent rounded-lg">
              <Lock className="h-5 w-5 text-accent-foreground" />
            </div>
            <div>
              <p className="text-2xl font-bold text-foreground">{permissions.length}</p>
              <p className="text-sm text-muted-foreground">{t('perms.title')}</p>
            </div>
          </div>
        </div>
        <div className="bg-card border border-border rounded-xl p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-muted rounded-lg">
              <Users className="h-5 w-5 text-muted-foreground" />
            </div>
            <div>
              <p className="text-2xl font-bold text-foreground">{Object.keys(groupedPermissions).length}</p>
              <p className="text-sm text-muted-foreground">Resources</p>
            </div>
          </div>
        </div>
      </div>

      {/* Search and Create */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search roles..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2.5 border border-input rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-ring bg-card"
          />
        </div>
        <button
          onClick={() => setShowCreateRole(true)}
          className="flex items-center gap-2 bg-primary hover:bg-primary/90 text-primary-foreground px-4 py-2.5 rounded-xl text-sm font-medium transition-colors whitespace-nowrap"
        >
          <Plus size={18} />
          Create Role
        </button>
      </div>

      {/* Create Role Form */}
      {showCreateRole && (
        <div className="bg-card border border-primary/20 rounded-xl p-6 shadow-lg">
          <h3 className="text-lg font-semibold text-foreground mb-4 flex items-center gap-2">
            <Plus size={20} className="text-primary" />
            Create New Role
          </h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                Role Name *
              </label>
              <input
                type="text"
                value={newRoleName}
                onChange={(e) => setNewRoleName(e.target.value)}
                placeholder="e.g. Content Manager"
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                Description
              </label>
              <input
                type="text"
                value={newRoleDescription}
                onChange={(e) => setNewRoleDescription(e.target.value)}
                placeholder="Brief description of this role"
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
              />
            </div>
          </div>
          <div className="flex justify-end gap-2">
            <button
              onClick={() => {
                setShowCreateRole(false);
                setNewRoleName('');
                setNewRoleDescription('');
              }}
              className="px-4 py-2 text-sm font-medium text-foreground bg-card border border-input rounded-lg hover:bg-accent"
            >
              Cancel
            </button>
            <button
              onClick={handleCreateRole}
              disabled={!newRoleName.trim() || createRoleMutation.isPending}
              className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50"
            >
              {createRoleMutation.isPending && <Loader2 size={16} className="animate-spin" />}
              Create Role
            </button>
          </div>
        </div>
      )}

      {/* Roles List */}
      <div className="space-y-4">
        {filteredRoles.map((role) => {
          const isExpanded = expandedRoles.has(role.id);
          const rolePermIds = role.permissions?.map(p => p.id) || [];

          return (
            <div
              key={role.id}
              className="bg-card border border-border rounded-xl overflow-hidden transition-shadow hover:shadow-md"
            >
              {/* Role Header */}
              <div
                className="flex items-center justify-between p-4 cursor-pointer hover:bg-accent/50 transition-colors"
                onClick={() => toggleRoleExpand(role.id)}
              >
                <div className="flex items-center gap-4">
                  <div className={`p-2.5 rounded-xl ${role.is_system_role ? 'bg-primary/10' : 'bg-muted'}`}>
                    <Shield size={20} className={role.is_system_role ? 'text-primary' : 'text-muted-foreground'} />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold text-foreground">
                        {role.display_name || role.name}
                      </h3>
                      {role.is_system_role && (
                        <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-primary/10 text-primary uppercase">
                          System
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-muted-foreground mt-0.5">
                      {role.description || 'No description'}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  <div className="text-right mr-4">
                    <p className="text-lg font-bold text-foreground">{rolePermIds.length}</p>
                    <p className="text-xs text-muted-foreground">permissions</p>
                  </div>

                  <div className="flex items-center gap-2">
                    {!role.is_system_role && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteRole(role.id);
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

              {/* Expanded Permissions */}
              {isExpanded && (
                <div className="border-t border-border bg-muted/30">
                  <div className="p-4">
                    <div className="flex items-center justify-between mb-4">
                      <h4 className="font-medium text-foreground flex items-center gap-2">
                        <Lock size={16} className="text-primary" />
                        Permissions
                      </h4>
                      <p className="text-xs text-muted-foreground">
                        Click to toggle permissions
                      </p>
                    </div>

                    {Object.keys(groupedPermissions).length === 0 ? (
                      <div className="text-center py-8 text-muted-foreground">
                        <AlertCircle size={32} className="mx-auto mb-2 opacity-50" />
                        <p>No permissions available. Create permissions first.</p>
                      </div>
                    ) : (
                      <div className="space-y-6">
                        {Object.entries(groupedPermissions).map(([resource, perms]) => {
                          const resourcePermIds = perms.map(p => p.id);
                          const selectedCount = resourcePermIds.filter(id => rolePermIds.includes(id)).length;
                          const allSelected = selectedCount === perms.length;
                          const someSelected = selectedCount > 0 && selectedCount < perms.length;

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
                                  onClick={() => handleSelectAllResource(role, resource, !allSelected)}
                                  className={`text-xs font-medium px-2 py-1 rounded transition-colors ${
                                    allSelected
                                      ? 'bg-primary/10 text-primary hover:bg-primary/20'
                                      : 'bg-muted text-muted-foreground hover:bg-accent'
                                  }`}
                                >
                                  {allSelected ? 'Deselect All' : 'Select All'}
                                </button>
                              </div>

                              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2">
                                {perms.map(perm => {
                                  const isChecked = rolePermIds.includes(perm.id);
                                  const isUpdating = updateRoleMutation.isPending;

                                  return (
                                    <button
                                      key={perm.id}
                                      onClick={() => handlePermissionToggle(role, perm.id)}
                                      disabled={isUpdating}
                                      className={`
                                        flex items-center gap-2 px-3 py-2 rounded-lg text-left transition-all text-sm
                                        ${isChecked
                                          ? 'bg-primary text-primary-foreground shadow-sm'
                                          : 'bg-muted/50 text-foreground hover:bg-muted border border-transparent hover:border-border'
                                        }
                                        ${isUpdating ? 'opacity-50 cursor-wait' : 'cursor-pointer'}
                                      `}
                                      title={perm.description || `${perm.resource}:${perm.action}`}
                                    >
                                      <div className={`
                                        w-4 h-4 rounded flex items-center justify-center flex-shrink-0
                                        ${isChecked ? 'bg-primary-foreground/20' : 'bg-card border border-input'}
                                      `}>
                                        {isChecked && <Check size={12} />}
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
        })}

        {filteredRoles.length === 0 && (
          <div className="text-center py-12 bg-card rounded-xl border border-border">
            <Shield size={48} className="mx-auto mb-4 text-muted-foreground opacity-50" />
            <p className="text-muted-foreground">No roles found</p>
            <button
              onClick={() => setShowCreateRole(true)}
              className="mt-4 text-primary hover:underline text-sm font-medium"
            >
              Create your first role
            </button>
          </div>
        )}
      </div>

      {/* Permissions Section */}
      <div className="bg-card border border-border rounded-xl overflow-hidden">
        <div className="p-4 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-accent rounded-lg">
              <Lock size={20} className="text-accent-foreground" />
            </div>
            <div>
              <h3 className="font-semibold text-foreground">{t('perms.title')}</h3>
              <p className="text-sm text-muted-foreground">
                {permissions.length} permissions across {Object.keys(groupedPermissions).length} resources
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
                <X size={16} />
                Cancel
              </>
            ) : (
              <>
                <Plus size={16} />
                Add Permission
              </>
            )}
          </button>
        </div>

        {/* Create Permission Form */}
        {showCreatePermission && (
          <div className="p-4 bg-muted/30 border-b border-border">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  Resource *
                </label>
                <div className="relative">
                  <input
                    type="text"
                    value={newPermResource}
                    onChange={(e) => setNewPermResource(e.target.value)}
                    placeholder="e.g. users, orders, reports"
                    list="resources-list"
                    className="w-full px-3 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none text-sm"
                  />
                  <datalist id="resources-list">
                    {existingResources.map(r => (
                      <option key={r} value={r} />
                    ))}
                  </datalist>
                </div>
                {existingResources.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-2">
                    {existingResources.slice(0, 5).map(r => (
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
                  Action *
                </label>
                <input
                  type="text"
                  value={newPermAction}
                  onChange={(e) => setNewPermAction(e.target.value)}
                  placeholder="e.g. create, read, delete"
                  list="actions-list"
                  className="w-full px-3 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none text-sm"
                />
                <datalist id="actions-list">
                  {commonActions.map(a => (
                    <option key={a} value={a} />
                  ))}
                </datalist>
                <div className="flex flex-wrap gap-1 mt-2">
                  {commonActions.slice(0, 5).map(a => (
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
                  Description
                </label>
                <input
                  type="text"
                  value={newPermDescription}
                  onChange={(e) => setNewPermDescription(e.target.value)}
                  placeholder="What this permission allows"
                  className="w-full px-3 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none text-sm"
                />
              </div>
            </div>
            <div className="flex items-center justify-between">
              <p className="text-xs text-muted-foreground">
                Permission name will be: <code className="bg-muted px-1.5 py-0.5 rounded">
                  {newPermResource || 'resource'}:{newPermAction || 'action'}
                </code>
              </p>
              <button
                onClick={handleCreatePermission}
                disabled={!newPermResource.trim() || !newPermAction.trim() || createPermissionMutation.isPending}
                className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50"
              >
                {createPermissionMutation.isPending && <Loader2 size={16} className="animate-spin" />}
                Create Permission
              </button>
            </div>
          </div>
        )}

        {/* Permissions List by Resource */}
        <div className="p-4">
          {Object.keys(groupedPermissions).length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Lock size={32} className="mx-auto mb-2 opacity-50" />
              <p>No permissions yet.</p>
              <button
                onClick={() => setShowCreatePermission(true)}
                className="mt-2 text-primary hover:underline text-sm font-medium"
              >
                Create your first permission
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {Object.entries(groupedPermissions).map(([resource, perms]) => (
                <div key={resource} className="border border-border rounded-lg p-3">
                  <div className="flex items-center justify-between mb-2">
                    <h4 className="font-medium text-foreground capitalize flex items-center gap-2">
                      <span className="px-2 py-0.5 bg-primary/10 text-primary rounded text-xs font-mono">
                        {resource}
                      </span>
                      <span className="text-xs text-muted-foreground font-normal">
                        {perms.length} {perms.length === 1 ? 'permission' : 'permissions'}
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
                          onClick={() => {
                            if (window.confirm(`Delete permission "${perm.name}"?`)) {
                              deletePermissionMutation.mutate(perm.id);
                            }
                          }}
                          className="opacity-0 group-hover:opacity-100 p-0.5 text-muted-foreground hover:text-destructive transition-all"
                        >
                          <X size={14} />
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
                      <Plus size={14} />
                      Add
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default AccessControl;
