import React from 'react';
import { Check } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { Permission } from '../../types';

interface AccessControlResourceGroupProps {
  resource: string;
  permissions: Permission[];
  selectedPermIds: string[];
  pendingUpdate: boolean;
  onTogglePermission: (permissionId: string) => void;
  onSelectAll: (resource: string, select: boolean) => void;
}

const AccessControlResourceGroup: React.FC<AccessControlResourceGroupProps> = ({
  resource,
  permissions,
  selectedPermIds,
  pendingUpdate,
  onTogglePermission,
  onSelectAll,
}) => {
  const { t } = useLanguage();
  const selectedCount = permissions.filter(p => selectedPermIds.includes(p.id)).length;
  const allSelected = selectedCount === permissions.length;

  return (
    <div className="bg-card rounded-lg border border-border p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <h5 className="font-medium text-foreground capitalize">
            {resource.replace(/_/g, ' ')}
          </h5>
          <span className="text-xs text-muted-foreground">
            ({selectedCount}/{permissions.length})
          </span>
        </div>
        <button
          onClick={() => onSelectAll(resource, !allSelected)}
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
        {permissions.map(perm => {
          const isChecked = selectedPermIds.includes(perm.id);
          return (
            <button
              key={perm.id}
              onClick={() => onTogglePermission(perm.id)}
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
                {isChecked && <Check size={12} />}
              </div>
              <span className="truncate font-medium capitalize">{perm.action}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
};

export default AccessControlResourceGroup;
