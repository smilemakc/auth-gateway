
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getPermission, createPermission, updatePermission } from '../services/mockData';
import { Permission } from '../types';
import { ArrowLeft, Save, Lock } from 'lucide-react';
import { useLanguage } from '../services/i18n';

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

  useEffect(() => {
    if (isEditMode) {
      const existing = getPermission(id);
      if (existing) {
        setFormData(existing);
      } else {
        navigate('/settings/permissions');
      }
    }
  }, [id, isEditMode, navigate]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    
    setTimeout(() => {
      if (isNewMode) {
        createPermission(formData);
      } else if (id) {
        updatePermission(id, formData);
      }
      setLoading(false);
      navigate('/settings/permissions');
    }, 800);
  };

  const generatedId = formData.resource && formData.action 
    ? `${formData.resource.toLowerCase()}:${formData.action.toLowerCase()}` 
    : '...';

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button 
          onClick={() => navigate('/settings/permissions')}
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
            onClick={() => navigate('/settings/permissions')}
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
            {loading ? t('common.saving') : (
              <>
                <Save size={16} className="mr-2" />
                {isNewMode ? t('common.create') : t('common.save')}
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
};

export default PermissionEdit;
