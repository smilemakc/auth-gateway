
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
      setError('Resource and action are required');
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
        setError('Permission editing is not yet supported. Please delete and recreate.');
        setLoading(false);
        return;
      }
    } catch (err: any) {
      setError(err.message || 'Failed to save permission');
    } finally {
      setLoading(false);
    }
  };

  // Loading state
  if (isEditMode && permissionLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
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
          className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{isNewMode ? t('common.create') : t('common.edit')}</h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 space-y-6">
          {error && (
            <div className="bg-red-50 border-l-4 border-red-500 p-4 flex items-center">
              <AlertCircle className="h-5 w-5 text-red-400 mr-2" />
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}

          <div className="flex items-center gap-4 bg-gray-50 p-4 rounded-lg border border-gray-100 mb-6">
            <Lock size={24} className="text-gray-500 flex-shrink-0" />
            <p className="text-sm text-gray-600">
              Permissions define granular access controls. They follow the format <code>resource:action</code>.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t('perms.name')}</label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              placeholder="e.g. Read Users"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('perms.resource')}</label>
              <input
                type="text"
                required
                value={formData.resource}
                onChange={(e) => setFormData(prev => ({ ...prev, resource: e.target.value.toLowerCase().replace(/\s+/g, '_') }))}
                placeholder="e.g. users"
                disabled={isEditMode}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none font-mono text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('perms.action')}</label>
              <input
                type="text"
                required
                value={formData.action}
                onChange={(e) => setFormData(prev => ({ ...prev, action: e.target.value.toLowerCase().replace(/\s+/g, '_') }))}
                placeholder="e.g. read"
                disabled={isEditMode}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none font-mono text-sm"
              />
            </div>
          </div>

          <div className="text-xs text-gray-500">
            Resulting ID: <code className="bg-gray-100 px-1 py-0.5 rounded border border-gray-200">{isEditMode ? id : generatedId}</code>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              rows={3}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none resize-none"
            />
          </div>
        </div>

        <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
           <button
            type="button"
            onClick={() => navigate('/settings/access-control?tab=permissions')}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 focus:outline-none"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={loading}
            className={`flex items-center px-6 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-lg hover:bg-blue-700 focus:outline-none
              ${loading ? 'opacity-70 cursor-not-allowed' : ''}`}
          >
            {loading ? (
              <span className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></span>
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
