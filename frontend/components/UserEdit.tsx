
import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Save,
  AlertCircle
} from 'lucide-react';
import type {
  AdminUserResponse,
  AdminCreateUserRequest,
  AdminUpdateUserRequest,
  AccountType
} from '@auth-gateway/client-sdk';
import type { Role } from '@auth-gateway/client-sdk';
import { useLanguage } from '../services/i18n';
import { useUserDetail, useUpdateUser, useCreateUser } from '../hooks/useUsers';
import { useRoles } from '../hooks/useRBAC';

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

  // Fetch user data if in edit mode
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

  // Sync user data to form when loaded
  useEffect(() => {
    if (isEditMode && user) {
      setFormData({
        full_name: user.full_name || '',
        username: user.username || '',
        email: user.email || '',
        password: '', // Don't load password
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      if (isEditMode) {
        // For update, send only updatable fields in snake_case
        const updateData: AdminUpdateUserRequest = {
          role_ids: formData.role_ids,
          is_active: formData.is_active
        };
        await updateUserMutation.mutateAsync({ id: id!, data: updateData });
        navigate(`/users/${id}`);
      } else {
        // For create, already in snake_case
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
      console.error('Failed to save user:', err);
      const message = err instanceof Error ? err.message : 'Failed to save user';
      setError(message);
    }
  };

  if (userLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-4 mb-6">
        <button 
          onClick={() => navigate(isEditMode ? `/users/${id}` : '/users')}
          className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
        >
          <ArrowLeft size={24} />
        </button>
        <h1 className="text-2xl font-bold text-gray-900">{isEditMode ? t('user.edit.title') : t('user.create.title')}</h1>
      </div>

      <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        
        {error && (
          <div className="bg-red-50 border-l-4 border-red-500 p-4 mb-4 mx-6 mt-6">
            <div className="flex">
              <div className="flex-shrink-0">
                <AlertCircle className="h-5 w-5 text-red-400" aria-hidden="true" />
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-700">{error}</p>
              </div>
            </div>
          </div>
        )}

        <div className="p-6 space-y-6">
          <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
            
            <div className="sm:col-span-3">
              <label htmlFor="username" className="block text-sm font-medium text-gray-700">{t('user.form.username')}</label>
              <div className="mt-1">
                <input
                  type="text"
                  name="username"
                  id="username"
                  value={formData.username}
                  onChange={handleChange}
                  className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md p-2.5 border"
                  required
                />
              </div>
            </div>

            <div className="sm:col-span-3">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                {t('user.form.role')}
              </label>
              <div className="space-y-2 mt-1">
                {availableRoles.map(role => (
                  <label key={role.id} className="flex items-center">
                    <input
                      type="checkbox"
                      checked={formData.role_ids.includes(role.id)}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setFormData({
                            ...formData,
                            role_ids: [...formData.role_ids, role.id]
                          });
                        } else {
                          setFormData({
                            ...formData,
                            role_ids: formData.role_ids.filter((id: string) => id !== role.id)
                          });
                        }
                      }}
                      className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                    />
                    <span className="ml-2 text-sm text-gray-700">{role.display_name || role.name}</span>
                  </label>
                ))}
              </div>
            </div>

            <div className="sm:col-span-6">
              <label htmlFor="full_name" className="block text-sm font-medium text-gray-700">{t('user.form.fullname')}</label>
              <div className="mt-1">
                <input
                  type="text"
                  name="full_name"
                  id="full_name"
                  value={formData.full_name}
                  onChange={handleChange}
                  className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md p-2.5 border"
                />
              </div>
            </div>

            <div className="sm:col-span-3">
              <label htmlFor="email" className="block text-sm font-medium text-gray-700">{t('auth.email')}</label>
              <div className="mt-1">
                <input
                  id="email"
                  name="email"
                  type="email"
                  value={formData.email}
                  onChange={handleChange}
                  className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md p-2.5 border"
                  required
                />
              </div>
            </div>

            {!isEditMode && (
              <div className="sm:col-span-3">
                <label htmlFor="password" className="block text-sm font-medium text-gray-700">{t('auth.password')}</label>
                <div className="mt-1">
                  <input
                    id="password"
                    name="password"
                    type="password"
                    value={formData.password}
                    onChange={handleChange}
                    className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md p-2.5 border"
                    required
                    minLength={8}
                    placeholder={t('user.form.password_placeholder') || 'Minimum 8 characters'}
                  />
                </div>
              </div>
            )}

            <div className={isEditMode ? "sm:col-span-3" : "sm:col-span-6"}>
              <label htmlFor="phone" className="block text-sm font-medium text-gray-700">{t('user.form.phone')}</label>
              <div className="mt-1">
                <input
                  type="text"
                  name="phone"
                  id="phone"
                  value={formData.phone || ''}
                  onChange={handleChange}
                  className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md p-2.5 border"
                />
              </div>
            </div>
          </div>

          {!isEditMode && (
            <div className="pt-6 border-t border-gray-200">
              <h3 className="text-sm font-medium text-gray-900 mb-4">{t('user.form.account_type') || 'Account Type'}</h3>
              <div className="flex gap-4">
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="account_type"
                    value="human"
                    checked={formData.account_type === 'human'}
                    onChange={handleChange}
                    className="focus:ring-blue-500 h-4 w-4 text-blue-600 border-gray-300"
                  />
                  <span className="ml-2 text-sm text-gray-700">{t('user.form.account_human') || 'Human'}</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="account_type"
                    value="service"
                    checked={formData.account_type === 'service'}
                    onChange={handleChange}
                    className="focus:ring-blue-500 h-4 w-4 text-blue-600 border-gray-300"
                  />
                  <span className="ml-2 text-sm text-gray-700">{t('user.form.account_service') || 'Service'}</span>
                </label>
              </div>
            </div>
          )}

          <div className="pt-6 border-t border-gray-200">
            <h3 className="text-sm font-medium text-gray-900 mb-4">{t('settings.roles')} & {t('common.status')}</h3>
            <div className="space-y-4">
              <div className="flex items-start">
                <div className="flex items-center h-5">
                  <input
                    id="is_active"
                    name="is_active"
                    type="checkbox"
                    checked={formData.is_active}
                    onChange={handleChange}
                    className="focus:ring-blue-500 h-4 w-4 text-blue-600 border-gray-300 rounded"
                  />
                </div>
                <div className="ml-3 text-sm">
                  <label htmlFor="is_active" className="font-medium text-gray-700">{t('user.form.active')}</label>
                  <p className="text-gray-500">{t('user.form.active_desc')}</p>
                </div>
              </div>

              <div className="flex items-start">
                <div className="flex items-center h-5">
                  <input
                    id="email_verified"
                    name="email_verified"
                    type="checkbox"
                    checked={formData.email_verified}
                    onChange={handleChange}
                    className="focus:ring-blue-500 h-4 w-4 text-blue-600 border-gray-300 rounded"
                  />
                </div>
                <div className="ml-3 text-sm">
                  <label htmlFor="email_verified" className="font-medium text-gray-700">{t('user.email_verified')}</label>
                </div>
              </div>

              <div className="flex items-start">
                <div className="flex items-center h-5">
                  <input
                    id="totp_enabled"
                    name="totp_enabled"
                    type="checkbox"
                    checked={formData.totp_enabled}
                    onChange={handleChange}
                    className="focus:ring-blue-500 h-4 w-4 text-blue-600 border-gray-300 rounded"
                  />
                </div>
                <div className="ml-3 text-sm">
                  <label htmlFor="totp_enabled" className="font-medium text-gray-700">{t('user.form.2fa_force')}</label>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate(isEditMode ? `/users/${id}` : '/users')}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={updateUserMutation.isPending || createUserMutation.isPending}
            className={`flex items-center px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500
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
