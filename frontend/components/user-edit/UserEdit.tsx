import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Save,
  AlertCircle,
} from 'lucide-react';
import type {
  AdminCreateUserRequest,
  AdminUpdateUserRequest,
  AccountType
} from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';
import { useUserDetail, useUpdateUser, useCreateUser } from '../../hooks/useUsers';
import { useRoles } from '../../hooks/rbac';
import { logger } from '@/lib/logger';
import UserEditBasicFields from './UserEditBasicFields';
import UserEditAuthFields from './UserEditAuthFields';

interface UserFormData {
  full_name: string;
  username: string;
  email: string;
  password: string;
  phone: string;
  role_ids: string[];
  account_type: AccountType;
  is_active: boolean;
  email_verified: boolean;
  totp_enabled: boolean;
}

const UserEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [error, setError] = useState('');

  const isEditMode = !!id;

  const { data: user, isLoading: userLoading } = useUserDetail(id!);
  const { data: rolesData } = useRoles();
  const updateUserMutation = useUpdateUser();
  const createUserMutation = useCreateUser();

  const availableRoles = rolesData || [];

  const [formData, setFormData] = useState<UserFormData>({
    full_name: '',
    username: '',
    email: '',
    password: '',
    phone: '',
    role_ids: [],
    account_type: 'human',
    is_active: true,
    email_verified: false,
    totp_enabled: false
  });

  useEffect(() => {
    if (isEditMode && user) {
      setFormData({
        full_name: user.full_name || '',
        username: user.username || '',
        email: user.email || '',
        password: '',
        phone: user.phone || '',
        role_ids: user.roles?.map((role) => role.id) || [],
        account_type: user.account_type || 'human',
        is_active: user.is_active ?? true,
        email_verified: user.email_verified ?? false,
        totp_enabled: user.totp_enabled ?? false
      });
    }
  }, [isEditMode, user]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;

    if (type === 'checkbox') {
      const checked = (e.target as HTMLInputElement).checked;
      setFormData(prev => ({ ...prev, [name]: checked }));
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
  };

  const handleRoleToggle = (roleId: string) => {
    const isSelected = formData.role_ids.includes(roleId);
    if (isSelected) {
      setFormData({
        ...formData,
        role_ids: formData.role_ids.filter((id: string) => id !== roleId)
      });
    } else {
      setFormData({
        ...formData,
        role_ids: [...formData.role_ids, roleId]
      });
    }
  };

  const handleAuthToggle = (field: 'account_type' | 'is_active' | 'email_verified' | 'totp_enabled') => {
    setFormData(prev => ({ ...prev, [field]: !prev[field] }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      if (isEditMode) {
        const updateData: AdminUpdateUserRequest = {
          email: formData.email,
          username: formData.username,
          full_name: formData.full_name,
          phone: formData.phone || undefined,
          role_ids: formData.role_ids,
          is_active: formData.is_active,
          email_verified: formData.email_verified,
        };
        await updateUserMutation.mutateAsync({ id: id!, data: updateData });
        navigate(`/users/${id}`);
      } else {
        const createData: AdminCreateUserRequest = {
          email: formData.email,
          username: formData.username,
          password: formData.password,
          full_name: formData.full_name,
          role_ids: formData.role_ids,
          account_type: formData.account_type
        };
        const newUser = await createUserMutation.mutateAsync(createData);
        navigate(`/users/${newUser.id}`);
      }
    } catch (err: unknown) {
      logger.error('Failed to save user:', err);
      const message = err instanceof Error ? err.message : 'Failed to save user';
      setError(message);
    }
  };

  if (userLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => navigate(isEditMode ? `/users/${id}` : '/users')}
          className="p-2 hover:bg-card rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <h1 className="text-2xl font-bold text-foreground">{isEditMode ? t('user.edit.title') : t('user.create.title')}</h1>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">

        {error && (
          <div className="bg-destructive/10 border-l-4 border-destructive p-4 mb-4 mx-6 mt-6">
            <div className="flex">
              <div className="flex-shrink-0">
                <AlertCircle className="h-5 w-5 text-destructive" aria-hidden="true" />
              </div>
              <div className="ml-3">
                <p className="text-sm text-destructive">{error}</p>
              </div>
            </div>
          </div>
        )}

        <div className="p-6 space-y-6">
          <UserEditBasicFields
            formData={formData}
            isEditMode={isEditMode}
            availableRoles={availableRoles}
            onChange={handleChange}
            onRoleToggle={handleRoleToggle}
          />

          <UserEditAuthFields
            formData={formData}
            isEditMode={isEditMode}
            onToggle={handleAuthToggle}
            onChange={handleChange}
          />
        </div>

        <div className="px-6 py-4 bg-muted border-t border-border flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate(isEditMode ? `/users/${id}` : '/users')}
            className="px-4 py-2 text-sm font-medium text-foreground bg-card border border-input rounded-md hover:bg-accent focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={updateUserMutation.isPending || createUserMutation.isPending}
            className={`flex items-center px-4 py-2 text-sm font-medium text-primary-foreground bg-primary border border-transparent rounded-md hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring
              ${updateUserMutation.isPending || createUserMutation.isPending ? 'opacity-70 cursor-not-allowed' : ''}`}
          >
            {(updateUserMutation.isPending || createUserMutation.isPending) ? t('common.saving') : (
              <>
                <Save size={16} className="mr-2" />
                {isEditMode ? t('user.form.save') : t('users.create_new')}
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
};

export default UserEdit;
