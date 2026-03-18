import React from 'react';
import { Lock, Shield, Users } from 'lucide-react';
import { StatCard } from '../ui';
import { useLanguage } from '../../services/i18n';

interface AccessControlStatsProps {
  rolesCount: number;
  permissionsCount: number;
  resourcesCount: number;
}

const AccessControlStats: React.FC<AccessControlStatsProps> = (
  {
    rolesCount,
    permissionsCount,
    resourcesCount,
  }) => {
  const { t } = useLanguage();

  return (
    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <StatCard
        icon={<Shield className="h-5 w-5 text-primary"/>}
        iconBgClass="bg-primary/10"
        value={rolesCount}
        label={t('roles.title')}
      />
      <StatCard
        icon={<Lock className="h-5 w-5 text-accent-foreground"/>}
        iconBgClass="bg-accent"
        value={permissionsCount}
        label={t('perms.title')}
      />
      <StatCard
        icon={<Users className="h-5 w-5 text-muted-foreground"/>}
        iconBgClass="bg-muted"
        value={resourcesCount}
        label={t('access_control.resources')}
      />
    </div>
  );
};

export default AccessControlStats;
