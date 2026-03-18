import React, { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../../services/i18n';
import { usePermissions, useRoles } from '../../hooks/rbac';
import { LoadingSpinner, PageHeader } from '../ui';
import type { Permission } from '../../types';
import AccessControlStats from './AccessControlStats';
import AccessControlRoleList from './AccessControlRoleList';
import AccessControlCreateRoleForm from './AccessControlCreateRoleForm';
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

  if (rolesLoading || permissionsLoading) {
    return <LoadingSpinner />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('nav.access_settings')}
        subtitle={t('access_control.subtitle')}
        onBack={() => navigate('/settings')}
      />

      <AccessControlStats
        rolesCount={roles.length}
        permissionsCount={permissions.length}
        resourcesCount={Object.keys(groupedPermissions).length}
      />

      <AccessControlRoleList
        roles={filteredRoles}
        groupedPermissions={groupedPermissions}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        expandedRoles={expandedRoles}
        onToggleExpand={toggleRoleExpand}
        onCreateRole={() => setShowCreateRole(true)}
      />

      {showCreateRole && (
        <AccessControlCreateRoleForm onClose={() => setShowCreateRole(false)} />
      )}

      <AccessControlPermissionsSection
        permissions={permissions}
        groupedPermissions={groupedPermissions}
      />
    </div>
  );
};

export default AccessControl;
