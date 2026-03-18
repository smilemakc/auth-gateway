import React, { useCallback, useState } from 'react';
import { useLanguage } from '../../services/i18n';
import { useCreatePermission } from '../../hooks/rbac';
import { FormField, TextInput, ActionButton } from '../ui';
import { logger } from '@/lib/logger';

const COMMON_ACTIONS = ['create', 'read', 'update', 'delete', 'list', 'manage', 'export', 'import'];

interface AccessControlPermissionFormProps {
  existingResources: string[];
  initialResource?: string;
  onClose: () => void;
}

const AccessControlPermissionForm: React.FC<AccessControlPermissionFormProps> = ({
  existingResources,
  initialResource = '',
  onClose,
}) => {
  const { t } = useLanguage();
  const createPermissionMutation = useCreatePermission();

  const [resource, setResource] = useState(initialResource);
  const [action, setAction] = useState('');
  const [description, setDescription] = useState('');

  const handleCreate = useCallback(async () => {
    if (!resource.trim() || !action.trim()) return;
    try {
      await createPermissionMutation.mutateAsync({
        name: `${resource}:${action}`,
        resource: resource.toLowerCase().replace(/\s+/g, '_'),
        action: action.toLowerCase().replace(/\s+/g, '_'),
        description,
      });
      setResource('');
      setAction('');
      setDescription('');
      onClose();
    } catch (err) {
      logger.error('Failed to create permission:', err);
    }
  }, [resource, action, description, createPermissionMutation, onClose]);

  return (
    <div className="p-4 bg-muted/30 border-b border-border">
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
        <FormField label={t('access_control.resource_label')} required>
          <TextInput
            value={resource}
            onChange={(e) => setResource(e.target.value)}
            placeholder={t('perm_edit.resource_placeholder')}
            variant="compact"
            datalistId="resources-list"
            datalistOptions={existingResources}
          />
          {existingResources.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-2">
              {existingResources.slice(0, 5).map((r) => (
                <button
                  key={r}
                  type="button"
                  onClick={() => setResource(r)}
                  className="px-2 py-0.5 text-xs bg-muted hover:bg-accent rounded transition-colors"
                >
                  {r}
                </button>
              ))}
            </div>
          )}
        </FormField>

        <FormField label={t('access_control.action_label')} required>
          <TextInput
            value={action}
            onChange={(e) => setAction(e.target.value)}
            placeholder={t('perm_edit.action_placeholder')}
            variant="compact"
            datalistId="actions-list"
            datalistOptions={COMMON_ACTIONS}
          />
          <div className="flex flex-wrap gap-1 mt-2">
            {COMMON_ACTIONS.slice(0, 5).map((a) => (
              <button
                key={a}
                type="button"
                onClick={() => setAction(a)}
                className="px-2 py-0.5 text-xs bg-muted hover:bg-accent rounded transition-colors"
              >
                {a}
              </button>
            ))}
          </div>
        </FormField>

        <FormField label={t('common.description')}>
          <TextInput
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t('common.description')}
            variant="compact"
          />
        </FormField>
      </div>

      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {t('access_control.perm_name_will_be')}:{' '}
          <code className="bg-muted px-1.5 py-0.5 rounded">
            {resource || 'resource'}:{action || 'action'}
          </code>
        </p>
        <ActionButton
          onClick={handleCreate}
          disabled={!resource.trim() || !action.trim()}
          isLoading={createPermissionMutation.isPending}
        >
          {t('access_control.create_permission')}
        </ActionButton>
      </div>
    </div>
  );
};

export default AccessControlPermissionForm;
