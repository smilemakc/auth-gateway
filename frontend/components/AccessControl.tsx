import React, { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Loader2,
  Lock,
  Plus,
  Search,
  Shield,
  Users
} from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { usePermissions, useRoles } from '../hooks/rbac';
import type { Permission } from '../types';
import AccessControlCreateRoleForm from './AccessControlCreateRoleForm';
import AccessControlRoleCard from './AccessControlRoleCard';
import AccessControlPermissionsSection from './AccessControlPermissionsSection';

const AccessControl: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();

  const [searchTerm, setSearchTerm] = useState('');
  const [expandedRoles, setExpandedRoles] = useState<Set<string>>(new Set());
  const [showCreateRole, setShowCreateRole] = useState(false);

  const { data: roles = [], isLoading: rolesLoading } = useRoles();
  const { data: permissions = [], isLoading: permissionsLoading } = usePermissions();

  const groupedPermissions = useMemo(() => {
    return permissions.reduce((acc, perm) => {
      if (!acc[perm.resource]) acc[perm.resource] = [];
      acc[perm.resource].push(perm);
      return acc;
    }, {} as Record<string, Permission[]>);
  }, [permissions]);

  const filteredRoles = useMemo(() =>
    roles.filter(r =>
      (r.display_name || r.name).toLowerCase().includes(searchTerm.toLowerCase()) ||
      r.description?.toLowerCase().includes(searchTerm.toLowerCase())
    ), [roles, searchTerm]);

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

  const isLoading = rolesLoading || permissionsLoading;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary"/>
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
            <ArrowLeft size={24}/>
          </button>
          <div>
            <h1 className="text-2xl font-bold text-foreground">
              {t('nav.access_settings')}
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              {t('access_control.subtitle')}
            </p>
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-card border border-border rounded-xl p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary/10 rounded-lg">
              <Shield className="h-5 w-5 text-primary"/>
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
              <Lock className="h-5 w-5 text-accent-foreground"/>
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
              <Users className="h-5 w-5 text-muted-foreground"/>
            </div>
            <div>
              <p className="text-2xl font-bold text-foreground">{Object.keys(groupedPermissions).length}</p>
              <p className="text-sm text-muted-foreground">{t('access_control.resources')}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Search and Create */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"/>
          <input
            type="text"
            placeholder={t('access_control.search_roles')}
            value={searchTerm}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2.5 border border-input rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-ring bg-card"
          />
        </div>
        <button
          onClick={() => setShowCreateRole(true)}
          className="flex items-center gap-2 bg-primary hover:bg-primary/90 text-primary-foreground px-4 py-2.5 rounded-xl text-sm font-medium transition-colors whitespace-nowrap"
        >
          <Plus size={18}/>
          {t('access_control.create_role')}
        </button>
      </div>

      {showCreateRole && (
        <AccessControlCreateRoleForm onClose={() => setShowCreateRole(false)}/>
      )}

      {/* Roles List */}
      <div className="space-y-4">
        {filteredRoles.map((role) => (
          <AccessControlRoleCard
            key={role.id}
            role={role}
            groupedPermissions={groupedPermissions}
            isExpanded={expandedRoles.has(role.id)}
            onToggleExpand={() => toggleRoleExpand(role.id)}
          />
        ))}

        {filteredRoles.length === 0 && (
          <div className="text-center py-12 bg-card rounded-xl border border-border">
            <Shield size={48} className="mx-auto mb-4 text-muted-foreground opacity-50"/>
            <p className="text-muted-foreground">{t('access_control.no_roles')}</p>
            <button
              onClick={() => setShowCreateRole(true)}
              className="mt-4 text-primary hover:underline text-sm font-medium"
            >
              {t('access_control.create_first_role')}
            </button>
          </div>
        )}
      </div>

      <AccessControlPermissionsSection
        permissions={permissions}
        groupedPermissions={groupedPermissions}
      />
    </div>
  );
};

export default AccessControl;
