import React from 'react';
import { Shield, Check } from 'lucide-react';
import type { Role } from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';

interface UserFormData {
  full_name: string;
  username: string;
  email: string;
  password: string;
  phone: string;
  role_ids: string[];
}

interface UserEditBasicFieldsProps {
  formData: UserFormData;
  isEditMode: boolean;
  availableRoles: Role[];
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void;
  onRoleToggle: (roleId: string) => void;
}

const UserEditBasicFields: React.FC<UserEditBasicFieldsProps> = ({
  formData,
  isEditMode,
  availableRoles,
  onChange,
  onRoleToggle,
}) => {
  const { t } = useLanguage();

  return (
    <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">

      <div className="sm:col-span-3">
        <label htmlFor="username" className="block text-sm font-medium text-foreground">{t('user.form.username')}</label>
        <div className="mt-1">
          <input
            type="text"
            name="username"
            id="username"
            value={formData.username}
            onChange={onChange}
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
                onClick={() => onRoleToggle(role.id)}
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
            onChange={onChange}
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
            onChange={onChange}
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
              onChange={onChange}
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
            onChange={onChange}
            className="shadow-sm focus:ring-ring focus:border-ring block w-full sm:text-sm border-input rounded-md p-2.5 border"
          />
        </div>
      </div>
    </div>
  );
};

export default UserEditBasicFields;
