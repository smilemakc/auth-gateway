
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Permission } from '../types';
import { ArrowLeft, Save, Shield, Check, AlertCircle } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useRoleDetail, useCreateRole, useUpdateRole, usePermissions } from '../hooks/useRBAC';

interface RoleFormState {
  name: string;
  display_name: string;
  description: string;
  permissions: string[];
  is_system_role: boolean;
}

const RoleEditor: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isEditMode = id && id !== 'new';
  const isNewMode = id === 'new';

  const [role, setRole] = useState<RoleFormState>({
    name: '',
    display_name: '',
    description: '',
    permissions: [],
    is_system_role: false
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Fetch data from API
  const { data: existingRole, isLoading: roleLoading } = useRoleDetail(isEditMode ? id! : '');
  const { data: availablePermissions = [], isLoading: permissionsLoading } = usePermissions();
  const createMutation = useCreateRole();
  const updateMutation = useUpdateRole();

  // Populate form when existing role is loaded
  useEffect(() => {
    if (isEditMode && existingRole) {
      setRole({
        name: existingRole.name,
        display_name: existingRole.display_name || existingRole.name,
        description: existingRole.description || '',
        permissions: existingRole.permissions?.map(p => p.id) || [],
        is_system_role: existingRole.is_system_role
      });
    }
  }, [existingRole, isEditMode]);

  const handlePermissionToggle = (permId: string) => {
    setRole(prev => {
      const currentPerms = prev.permissions || [];
      if (currentPerms.includes(permId)) {
        return { ...prev, permissions: currentPerms.filter(p => p !== permId) };
      } else {
        return { ...prev, permissions: [...currentPerms, permId] };
      }
    });
  };

  const handleSelectAllResource = (resource: string, select: boolean) => {
    const resourcePerms = availablePermissions.filter(p => p.resource === resource).map(p => p.id);
    setRole(prev => {
      let newPerms = prev.permissions || [];
      if (select) {
        const toAdd = resourcePerms.filter(p => !newPerms.includes(p));
        newPerms = [...newPerms, ...toAdd];
      } else {
        newPerms = newPerms.filter(p => !resourcePerms.includes(p));
      }
      return { ...prev, permissions: newPerms };
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    const displayName = role.display_name || role.name;
    if (!displayName) {
      setError('Role name is required');
      setLoading(false);
      return;
    }

    try {
      if (isNewMode) {
        await createMutation.mutateAsync({
          name: displayName.toLowerCase().replace(/\s+/g, '_'),
          display_name: displayName,
          description: role.description,
          permissions: role.permissions || []
        });
      } else if (id) {
        await updateMutation.mutateAsync({
          id,
          data: {
            display_name: displayName,
            description: role.description,
            permissions: role.permissions
          }
        });
      }
      navigate('/settings/access-control?tab=roles');
    } catch (err: any) {
      setError(err.message || 'Failed to save role');
    } finally {
      setLoading(false);
    }
  };

  // Loading state for page
  const isPageLoading = (isEditMode && roleLoading) || permissionsLoading;

  if (isPageLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  // Group permissions for display by RESOURCE
  const groupedPermissions = availablePermissions.reduce((acc, perm) => {
    if (!acc[perm.resource]) acc[perm.resource] = [];
    acc[perm.resource].push(perm);
    return acc;
  }, {} as Record<string, Permission[]>);

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/settings/access-control?tab=roles')}
          className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{isNewMode ? t('common.create') : t('common.edit')}</h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Role Details */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 space-y-6">
           {error && (
            <div className="bg-red-50 border-l-4 border-red-500 p-4 flex items-center">
              <AlertCircle className="h-5 w-5 text-red-400 mr-2" />
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}

          <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('users.col_role')}</label>
              <input
                type="text"
                value={role.display_name || role.name}
                onChange={(e) => setRole(prev => ({ ...prev, display_name: e.target.value, name: e.target.value }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
                placeholder="e.g. Content Manager"
                disabled={role.is_system_role}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <input
                type="text"
                value={role.description}
                onChange={(e) => setRole(prev => ({ ...prev, description: e.target.value }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
              />
            </div>
          </div>
        </div>

        {/* Permissions Matrix */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
          <div className="p-6 border-b border-gray-100">
            <h2 className="text-lg font-bold text-gray-900 flex items-center gap-2">
              <Shield size={20} className="text-blue-500" />
              {t('roles.permissions')}
            </h2>
          </div>

          <div className="divide-y divide-gray-100">
            {Object.entries(groupedPermissions).map(([resource, permissions]: [string, Permission[]]) => {
              const allSelected = permissions.every(p => role.permissions?.includes(p.id));

              return (
                <div key={resource} className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-md font-bold text-gray-900 capitalize flex items-center gap-2">
                        {resource.replace(/_/g, ' ')}
                    </h3>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => handleSelectAllResource(resource, !allSelected)}
                        className="text-xs text-blue-600 hover:text-blue-800 font-medium"
                      >
                        {allSelected ? 'Deselect Resource' : 'Select Resource'}
                      </button>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {permissions.map(perm => {
                      const isChecked = role.permissions?.includes(perm.id);
                      return (
                        <div
                          key={perm.id}
                          onClick={() => handlePermissionToggle(perm.id)}
                          className={`
                            cursor-pointer border rounded-lg p-3 flex items-start gap-3 transition-all relative overflow-hidden
                            ${isChecked
                              ? 'bg-blue-50 border-blue-200 ring-1 ring-blue-200'
                              : 'bg-white border-gray-200 hover:border-gray-300'}
                          `}
                        >
                          <div className={`mt-0.5 w-5 h-5 rounded border flex items-center justify-center flex-shrink-0 transition-colors ${isChecked ? 'bg-blue-600 border-blue-600' : 'border-gray-300 bg-white'}`}>
                            {isChecked && <Check size={12} className="text-white" />}
                          </div>
                          <div className="flex-1">
                            <p className={`text-sm font-semibold capitalize ${isChecked ? 'text-blue-900' : 'text-gray-900'}`}>{perm.action}</p>
                            <p className="text-xs text-gray-500 mt-0.5">{perm.description}</p>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Footer Actions */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={() => navigate('/settings/access-control?tab=roles')}
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
            {t('common.save')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default RoleEditor;
