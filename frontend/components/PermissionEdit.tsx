
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Permission } from '../types';
import { ArrowLeft, Save, Lock, AlertCircle } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { usePermissionDetail, useCreatePermission } from '../hooks/useRBAC';

const PermissionEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isEditMode = id && id !== 'new';
  const isNewMode = id === 'new';

  const [formData, setFormData] = useState<Partial<Permission>>({
    name: '',
    resource: '',
    action: '',
    description: ''
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Fetch data from API
  const { data: existingPermission, isLoading: permissionLoading } = usePermissionDetail(isEditMode ? id! : '');
  const createMutation = useCreatePermission();

  // Populate form when existing permission is loaded
  useEffect(() => {
    if (isEditMode && existingPermission) {
      setFormData({
        name: existingPermission.name,
        resource: existingPermission.resource,
        action: existingPermission.action,
        description: existingPermission.description || ''
      });
    }
  }, [existingPermission, isEditMode]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    if (!formData.resource || !formData.action) {
      setError(t('perm_edit.err_resource_action'));
      setLoading(false);
      return;
    }

    try {
      if (isNewMode) {
        await createMutation.mutateAsync({
          name: formData.name || `${formData.resource}.${formData.action}`,
          resource: formData.resource,
          action: formData.action,
          description: formData.description
        });
        navigate('/settings/access-control?tab=permissions');
      } else {
        // Note: Update is not fully implemented in SDK
        setError(t('perm_edit.edit_not_supported'));
        setLoading(false);
        return;
      }
    } catch (err: any) {
      setError(err.message || t('perm_edit.save_error'));
    } finally {
      setLoading(false);
    }
  };

  // Loading state
  if (isEditMode && permissionLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  const generatedId = formData.resource && formData.action
    ? `${formData.resource.toLowerCase()}:${formData.action.toLowerCase()}`
    : '...';

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/settings/access-control?tab=permissions')}
          className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-foreground">{isNewMode ? t('common.create') : t('common.edit')}</h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="p-6 space-y-6">
          {error && (
            <div className="bg-destructive/10 border-l-4 border-destructive p-4 flex items-center">
              <AlertCircle className="h-5 w-5 text-destructive/60 mr-2" />
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}

          <div className="flex items-center gap-4 bg-muted p-4 rounded-lg border border-border mb-6">
            <Lock size={24} className="text-muted-foreground flex-shrink-0" />
            <p className="text-sm text-muted-foreground">
              {t('perm_edit.info_text')} <code>resource:action</code>.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('perms.name')}</label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              placeholder={t('perm_edit.name_placeholder')}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">{t('perms.resource')}</label>
              <input
                type="text"
                required
                value={formData.resource}
                onChange={(e) => setFormData(prev => ({ ...prev, resource: e.target.value.toLowerCase().replace(/\s+/g, '_') }))}
                placeholder={t('perm_edit.resource_placeholder')}
                disabled={isEditMode}
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none font-mono text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">{t('perms.action')}</label>
              <input
                type="text"
                required
                value={formData.action}
                onChange={(e) => setFormData(prev => ({ ...prev, action: e.target.value.toLowerCase().replace(/\s+/g, '_') }))}
                placeholder={t('perm_edit.action_placeholder')}
                disabled={isEditMode}
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none font-mono text-sm"
              />
            </div>
          </div>

          <div className="text-xs text-muted-foreground">
            {t('perm_edit.resulting_id')}: <code className="bg-muted px-1 py-0.5 rounded border border-border">{isEditMode ? id : generatedId}</code>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('common.description')}</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              rows={3}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none resize-none"
            />
          </div>
        </div>

        <div className="px-6 py-4 bg-muted border-t border-border flex justify-end gap-3">
           <button
            type="button"
            onClick={() => navigate('/settings/access-control?tab=permissions')}
            className="px-4 py-2 text-sm font-medium text-foreground bg-card border border-input rounded-lg hover:bg-accent focus:outline-none"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={loading}
            className={`flex items-center px-6 py-2 text-sm font-medium text-primary-foreground bg-primary border border-transparent rounded-lg hover:bg-primary-600 focus:outline-none
              ${loading ? 'opacity-70 cursor-not-allowed' : ''}`}
          >
            {loading ? (
              <span className="w-5 h-5 border-2 border-primary-foreground border-t-transparent rounded-full animate-spin mr-2"></span>
            ) : (
              <Save size={16} className="mr-2" />
            )}
            {isNewMode ? t('common.create') : t('common.save')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default PermissionEdit;
