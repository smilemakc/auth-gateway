
import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Save,
  AlertCircle,
  Shield,
  Check,
  ToggleLeft,
  ToggleRight
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
        // For update, send all editable fields in snake_case
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
          <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
            
            <div className="sm:col-span-3">
              <label htmlFor="username" className="block text-sm font-medium text-foreground">{t('user.form.username')}</label>
              <div className="mt-1">
                <input
                  type="text"
                  name="username"
                  id="username"
                  value={formData.username}
                  onChange={handleChange}
                  className="shadow-sm focus:ring-ring focus:border-ring block w-full sm:text-sm border-input rounded-md p-2.5 border"
                  required
                />
              </div>
            </div>

            <div className="sm:col-span-6">
              <label className="block text-sm font-medium text-foreground mb-3">
                {t('user.form.role')}
              </label>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {availableRoles.map(role => {
                  const isSelected = formData.role_ids.includes(role.id);
                  return (
                    <button
                      key={role.id}
                      type="button"
                      onClick={() => {
                        if (isSelected) {
                          setFormData({
                            ...formData,
                            role_ids: formData.role_ids.filter((id: string) => id !== role.id)
                          });
                        } else {
                          setFormData({
                            ...formData,
                            role_ids: [...formData.role_ids, role.id]
                          });
                        }
                      }}
                      className={`
                        flex items-start gap-3 p-4 rounded-xl border-2 text-left transition-all
                        ${isSelected
                          ? 'border-primary bg-primary/5 ring-1 ring-primary/20'
                          : 'border-border bg-card hover:border-input hover:bg-accent/50'
                        }
                      `}
                    >
                      <div className={`
                        p-2 rounded-lg flex-shrink-0
                        ${isSelected ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'}
                      `}>
                        <Shield size={18} />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className={`font-medium ${isSelected ? 'text-primary' : 'text-foreground'}`}>
                            {role.display_name || role.name}
                          </span>
                          {isSelected && (
                            <div className="w-5 h-5 rounded-full bg-primary flex items-center justify-center">
                              <Check size={12} className="text-primary-foreground" />
                            </div>
                          )}
                        </div>
                        {role.description && (
                          <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                            {role.description}
                          </p>
                        )}
                        {role.permissions && role.permissions.length > 0 && (
                          <p className="text-xs text-muted-foreground mt-1">
                            {role.permissions.length} permissions
                          </p>
                        )}
                      </div>
                    </button>
                  );
                })}
              </div>
              {availableRoles.length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-4 bg-muted/50 rounded-lg">
                  No roles available. Create roles in Access Settings first.
                </p>
              )}
            </div>

            <div className="sm:col-span-6">
              <label htmlFor="full_name" className="block text-sm font-medium text-foreground">{t('user.form.fullname')}</label>
              <div className="mt-1">
                <input
                  type="text"
                  name="full_name"
                  id="full_name"
                  value={formData.full_name}
                  onChange={handleChange}
                  className="shadow-sm focus:ring-ring focus:border-ring block w-full sm:text-sm border-input rounded-md p-2.5 border"
                />
              </div>
            </div>

            <div className="sm:col-span-3">
              <label htmlFor="email" className="block text-sm font-medium text-foreground">{t('auth.email')}</label>
              <div className="mt-1">
                <input
                  id="email"
                  name="email"
                  type="email"
                  value={formData.email}
                  onChange={handleChange}
                  className="shadow-sm focus:ring-ring focus:border-ring block w-full sm:text-sm border-input rounded-md p-2.5 border"
                  required
                />
              </div>
            </div>

            {!isEditMode && (
              <div className="sm:col-span-3">
                <label htmlFor="password" className="block text-sm font-medium text-foreground">{t('auth.password')}</label>
                <div className="mt-1">
                  <input
                    id="password"
                    name="password"
                    type="password"
                    value={formData.password}
                    onChange={handleChange}
                    className="shadow-sm focus:ring-ring focus:border-ring block w-full sm:text-sm border-input rounded-md p-2.5 border"
                    required
                    minLength={8}
                    placeholder={t('user.form.password_placeholder')}
                  />
                </div>
              </div>
            )}

            <div className={isEditMode ? "sm:col-span-3" : "sm:col-span-6"}>
              <label htmlFor="phone" className="block text-sm font-medium text-foreground">{t('user.form.phone')}</label>
              <div className="mt-1">
                <input
                  type="text"
                  name="phone"
                  id="phone"
                  value={formData.phone || ''}
                  onChange={handleChange}
                  className="shadow-sm focus:ring-ring focus:border-ring block w-full sm:text-sm border-input rounded-md p-2.5 border"
                />
              </div>
            </div>
          </div>

          {!isEditMode && (
            <div className="pt-6 border-t border-border">
              <h3 className="text-sm font-medium text-foreground mb-4">{t('user.form.account_type')}</h3>
              <div className="flex gap-4">
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="account_type"
                    value="human"
                    checked={formData.account_type === 'human'}
                    onChange={handleChange}
                    className="focus:ring-ring h-4 w-4 text-primary border-input"
                  />
                  <span className="ml-2 text-sm text-foreground">{t('user.form.account_human')}</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="account_type"
                    value="service"
                    checked={formData.account_type === 'service'}
                    onChange={handleChange}
                    className="focus:ring-ring h-4 w-4 text-primary border-input"
                  />
                  <span className="ml-2 text-sm text-foreground">{t('user.form.account_service')}</span>
                </label>
              </div>
            </div>
          )}

          <div className="pt-6 border-t border-border">
            <h3 className="text-sm font-medium text-foreground mb-4">{t('settings.roles')} & {t('common.status')}</h3>
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <button
                  type="button"
                  onClick={() => setFormData(prev => ({ ...prev, is_active: !prev.is_active }))}
                  className={`transition-colors mt-0.5 ${formData.is_active ? 'text-success' : 'text-muted-foreground'}`}
                >
                  {formData.is_active ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                </button>
                <div className="text-sm">
                  <span className="font-medium text-foreground">{t('user.form.active')}</span>
                  <p className="text-muted-foreground">{t('user.form.active_desc')}</p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <button
                  type="button"
                  onClick={() => setFormData(prev => ({ ...prev, email_verified: !prev.email_verified }))}
                  className={`transition-colors mt-0.5 ${formData.email_verified ? 'text-success' : 'text-muted-foreground'}`}
                >
                  {formData.email_verified ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                </button>
                <div className="text-sm">
                  <span className="font-medium text-foreground">{t('user.email_verified')}</span>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <button
                  type="button"
                  onClick={() => setFormData(prev => ({ ...prev, totp_enabled: !prev.totp_enabled }))}
                  className={`transition-colors mt-0.5 ${formData.totp_enabled ? 'text-success' : 'text-muted-foreground'}`}
                >
                  {formData.totp_enabled ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                </button>
                <div className="text-sm">
                  <span className="font-medium text-foreground">{t('user.form.2fa_force')}</span>
                </div>
              </div>
            </div>
          </div>
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
