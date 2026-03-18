import React from 'react';
import { Plus, Shield } from 'lucide-react';
import { SearchInput, EmptyState, ActionButton } from '../ui';
import { useLanguage } from '../../services/i18n';
import AccessControlRoleCard from './AccessControlRoleCard';
import type { Permission, RoleDefinition } from '../../types';

interface AccessControlRoleListProps {
  roles: RoleDefinition[];
  groupedPermissions: Record<string, Permission[]>;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  expandedRoles: Set<string>;
  onToggleExpand: (roleId: string) => void;
  onCreateRole: () => void;
}

const AccessControlRoleList: React.FC<AccessControlRoleListProps> = ({
  roles,
  groupedPermissions,
  searchTerm,
  onSearchChange,
  expandedRoles,
  onToggleExpand,
  onCreateRole,
}) => {
  const { t } = useLanguage();

  return (
    <>
      <div className="flex flex-col sm:flex-row gap-4">
        <SearchInput
          value={searchTerm}
          onChange={onSearchChange}
          placeholder={t('access_control.search_roles')}
        />
        <ActionButton icon={<Plus size={18} />} onClick={onCreateRole}>
          {t('access_control.create_role')}
        </ActionButton>
      </div>

      <div className="space-y-4">
        {roles.map((role) => (
          <AccessControlRoleCard
            key={role.id}
            role={role}
            groupedPermissions={groupedPermissions}
            isExpanded={expandedRoles.has(role.id)}
            onToggleExpand={() => onToggleExpand(role.id)}
          />
        ))}

        {roles.length === 0 && (
          <EmptyState
            icon={<Shield size={48} />}
            message={t('access_control.no_roles')}
            action={{ label: t('access_control.create_first_role'), onClick: onCreateRole }}
          />
        )}
      </div>
    </>
  );
};

export default AccessControlRoleList;
