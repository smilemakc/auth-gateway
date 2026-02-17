import React from 'react';
import { ToggleLeft, ToggleRight } from 'lucide-react';
import type { AccountType } from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';

interface AuthFieldsData {
  account_type: AccountType;
  is_active: boolean;
  email_verified: boolean;
  totp_enabled: boolean;
}

interface UserEditAuthFieldsProps {
  formData: AuthFieldsData;
  isEditMode: boolean;
  onToggle: (field: keyof AuthFieldsData) => void;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const UserEditAuthFields: React.FC<UserEditAuthFieldsProps> = ({
  formData,
  isEditMode,
  onToggle,
  onChange,
}) => {
  const { t } = useLanguage();

  return (
    <>
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
                onChange={onChange}
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
                onChange={onChange}
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
              onClick={() => onToggle('is_active')}
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
              onClick={() => onToggle('email_verified')}
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
              onClick={() => onToggle('totp_enabled')}
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
    </>
  );
};

export default UserEditAuthFields;
