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
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  const availableParentGroups = groupsData?.groups.filter((g) => g.id !== id) || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">{isNew ? 'Create Group' : 'Edit Group'}</h1>
        <button
          onClick={() => navigate('/groups')}
          className="text-muted-foreground hover:text-foreground flex items-center gap-2"
        >
          <X size={20} />
          Cancel
        </button>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border p-6 space-y-6">
        {!isNew && (
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">Name</label>
            <input
              type="text"
              value={formData.name}
              disabled
              className="w-full px-3 py-2 border border-input rounded-lg bg-muted text-muted-foreground"
            />
            <p className="mt-1 text-xs text-muted-foreground">Group name cannot be changed after creation</p>
          </div>
        )}

        {isNew && (
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">
              Name <span className="text-destructive">*</span>
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value.toLowerCase() })}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                errors.name ? 'border-destructive' : 'border-input'
              }`}
              placeholder="engineering"
            />
            {errors.name && <p className="mt-1 text-sm text-destructive">{errors.name}</p>}
            <p className="mt-1 text-xs text-muted-foreground">Lowercase letters, numbers, hyphens, and underscores only</p>
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-foreground mb-1">
            Display Name <span className="text-destructive">*</span>
          </label>
          <input
            type="text"
            value={formData.display_name}
            onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
            className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
              errors.display_name ? 'border-destructive' : 'border-input'
            }`}
            placeholder="Engineering Department"
          />
          {errors.display_name && <p className="mt-1 text-sm text-destructive">{errors.display_name}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-1">Description</label>
          <textarea
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            rows={3}
            className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            placeholder="Engineering team responsible for product development"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-1">Parent Group</label>
          <select
            value={formData.parent_group_id || ''}
            onChange={(e) => {
              const value = e.target.value;
              setFormData({ ...formData, parent_group_id: value ? value : undefined });
            }}
            className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="">None (Top-level group)</option>
            {availableParentGroups.map((g) => (
              <option key={g.id} value={g.id}>
                {g.display_name}
              </option>
            ))}
          </select>
          <p className="mt-1 text-xs text-muted-foreground">Optional: Select a parent group to create a hierarchy</p>
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-border">
          <button
            type="button"
            onClick={() => navigate('/groups')}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createGroup.isPending || updateGroup.isPending}
            className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
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

