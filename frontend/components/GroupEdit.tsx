import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Save, X, Loader } from 'lucide-react';
import type { CreateGroupRequest, UpdateGroupRequest, Group } from '@auth-gateway/client-sdk';
import { useGroup, useCreateGroup, useUpdateGroup, useGroups } from '../hooks/useGroups';

const GroupEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = !id;

  const { data: group, isLoading: isLoadingGroup } = useGroup(id || '');
  const { data: groupsData } = useGroups(1, 100); // For parent group selection
  const createGroup = useCreateGroup();
  const updateGroup = useUpdateGroup();

  const [formData, setFormData] = useState<CreateGroupRequest>({
    name: '',
    display_name: '',
    description: '',
    parent_group_id: undefined,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (group && !isNew) {
      setFormData({
        name: group.name,
        display_name: group.display_name,
        description: group.description || '',
        parent_group_id: group.parent_group_id || undefined,
      });
    }
  }, [group, isNew]);

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    } else if (!/^[a-z0-9_-]+$/.test(formData.name)) {
      newErrors.name = 'Name must contain only lowercase letters, numbers, hyphens, and underscores';
    }

    if (!formData.display_name.trim()) {
      newErrors.display_name = 'Display name is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    try {
      if (isNew) {
        // Prepare create data - remove undefined parent_group_id if empty
        const createData: CreateGroupRequest = {
          name: formData.name,
          display_name: formData.display_name,
          description: formData.description || undefined,
          parent_group_id: formData.parent_group_id || undefined,
        };
        await createGroup.mutateAsync(createData);
      } else {
        const updateData: UpdateGroupRequest = {
          display_name: formData.display_name,
          description: formData.description || undefined,
          parent_group_id: formData.parent_group_id || undefined,
        };
        await updateGroup.mutateAsync({ id: id!, data: updateData });
      }
      navigate('/groups');
    } catch (error: any) {
      console.error('Failed to save group:', error);
      const errorMessage = error?.response?.data?.message || error?.message || 'Failed to save group';
      alert(`Error: ${errorMessage}`);
    }
  };

  if (isLoadingGroup && !isNew) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  const availableParentGroups = groupsData?.groups.filter((g) => g.id !== id) || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{isNew ? 'Create Group' : 'Edit Group'}</h1>
        <button
          onClick={() => navigate('/groups')}
          className="text-gray-500 hover:text-gray-700 flex items-center gap-2"
        >
          <X size={20} />
          Cancel
        </button>
      </div>

      <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 space-y-6">
        {!isNew && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              type="text"
              value={formData.name}
              disabled
              className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500"
            />
            <p className="mt-1 text-xs text-gray-500">Group name cannot be changed after creation</p>
          </div>
        )}

        {isNew && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Name <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value.toLowerCase() })}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                errors.name ? 'border-red-300' : 'border-gray-300'
              }`}
              placeholder="engineering"
            />
            {errors.name && <p className="mt-1 text-sm text-red-600">{errors.name}</p>}
            <p className="mt-1 text-xs text-gray-500">Lowercase letters, numbers, hyphens, and underscores only</p>
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Display Name <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            value={formData.display_name}
            onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
            className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${
              errors.display_name ? 'border-red-300' : 'border-gray-300'
            }`}
            placeholder="Engineering Department"
          />
          {errors.display_name && <p className="mt-1 text-sm text-red-600">{errors.display_name}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Engineering team responsible for product development"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Parent Group</label>
          <select
            value={formData.parent_group_id || ''}
            onChange={(e) => {
              const value = e.target.value;
              setFormData({ ...formData, parent_group_id: value ? value : undefined });
            }}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">None (Top-level group)</option>
            {availableParentGroups.map((g) => (
              <option key={g.id} value={g.id}>
                {g.display_name}
              </option>
            ))}
          </select>
          <p className="mt-1 text-xs text-gray-500">Optional: Select a parent group to create a hierarchy</p>
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
          <button
            type="button"
            onClick={() => navigate('/groups')}
            className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createGroup.isPending || updateGroup.isPending}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {(createGroup.isPending || updateGroup.isPending) && <Loader size={16} className="animate-spin" />}
            <Save size={16} />
            {isNew ? 'Create Group' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default GroupEdit;

