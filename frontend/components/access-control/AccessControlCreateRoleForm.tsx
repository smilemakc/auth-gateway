import React, { useCallback, useState } from 'react';
import { Loader2, Plus } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useCreateRole } from '../../hooks/rbac';
import { logger } from '@/lib/logger';

interface AccessControlCreateRoleFormProps {
  onClose: () => void;
}

export default function AccessControlCreateRoleForm({ onClose }: AccessControlCreateRoleFormProps) {
  const { t } = useLanguage();
  const [newRoleName, setNewRoleName] = useState('');
  const [newRoleDescription, setNewRoleDescription] = useState('');
  const createRoleMutation = useCreateRole();

  const handleCreateRole = useCallback(async () => {
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
      onClose();
    } catch (err) {
      logger.error('Failed to create role:', err);
    }
  }, [newRoleName, newRoleDescription, createRoleMutation, onClose]);

  return (
    <div className="bg-card border border-primary/20 rounded-xl p-6 shadow-lg">
      <h3 className="text-lg font-semibold text-foreground mb-4 flex items-center gap-2">
        <Plus size={20} className="text-primary"/>
        {t('access_control.create_new_role')}
      </h3>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
        <div>
          <label className="block text-sm font-medium text-foreground mb-1">
            {t('access_control.role_name')} *
          </label>
          <input
            type="text"
            value={newRoleName}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewRoleName(e.target.value)}
            placeholder="e.g. Content Manager"
            className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-foreground mb-1">
            {t('common.description')}
          </label>
          <input
            type="text"
            value={newRoleDescription}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewRoleDescription(e.target.value)}
            placeholder={t('common.description')}
            className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
          />
        </div>
      </div>
      <div className="flex justify-end gap-2">
        <button
          onClick={onClose}
          className="px-4 py-2 text-sm font-medium text-foreground bg-card border border-input rounded-lg hover:bg-accent"
        >
          {t('common.cancel')}
        </button>
        <button
          onClick={handleCreateRole}
          disabled={!newRoleName.trim() || createRoleMutation.isPending}
          className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50"
        >
          {createRoleMutation.isPending && <Loader2 size={16} className="animate-spin"/>}
          {t('access_control.create_role')}
        </button>
      </div>
    </div>
  );
}
