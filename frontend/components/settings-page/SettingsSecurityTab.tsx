import React from 'react';
import { Lock, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { PasswordPolicy } from '../../types';

interface SettingsSecurityTabProps {
  localPasswordPolicy: PasswordPolicy | null;
  onPolicyChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onTogglePolicy: (field: keyof PasswordPolicy) => void;
}

export const SettingsSecurityTab: React.FC<SettingsSecurityTabProps> = ({
  localPasswordPolicy,
  onPolicyChange,
  onTogglePolicy,
}) => {
  const { t } = useLanguage();

  return (
    <section className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="p-6 border-b border-border flex items-center gap-3">
        <div className="p-2 bg-primary/10 text-primary rounded-lg">
          <Lock size={20} />
        </div>
        <h2 className="text-lg font-semibold text-foreground">{t('settings.security_policies')}</h2>
      </div>
      {localPasswordPolicy && (
        <div className="p-6 space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">{t('settings.jwt_ttl')}</label>
              <input
                type="number"
                name="jwtTtlMinutes"
                value={localPasswordPolicy.jwtTtlMinutes}
                onChange={onPolicyChange}
                className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">{t('settings.refresh_ttl')}</label>
              <input
                type="number"
                name="refreshTtlDays"
                value={localPasswordPolicy.refreshTtlDays}
                onChange={onPolicyChange}
                className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
              />
            </div>
          </div>

          <div className="border-t border-border pt-6">
            <h3 className="text-md font-medium text-foreground mb-4">{t('settings.password_policy')}</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">{t('settings.min_pass')}</label>
                <input
                  type="number"
                  name="minLength"
                  value={localPasswordPolicy.minLength}
                  onChange={onPolicyChange}
                  className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">{t('settings.pass_history')}</label>
                <input
                  type="number"
                  name="historyCount"
                  value={localPasswordPolicy.historyCount}
                  onChange={onPolicyChange}
                  className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
                />
              </div>
               <div>
                <label className="block text-sm font-medium text-foreground mb-2">{t('settings.pass_expiry')}</label>
                <input
                  type="number"
                  name="expiryDays"
                  value={localPasswordPolicy.expiryDays}
                  onChange={onPolicyChange}
                  className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <label className="flex items-center gap-3 cursor-pointer"
                onClick={() => onTogglePolicy('requireUppercase')}>
                <span className={`transition-colors ${localPasswordPolicy.requireUppercase ? 'text-success' : 'text-muted-foreground'}`}>
                  {localPasswordPolicy.requireUppercase ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                </span>
                <span className="text-sm text-foreground">{t('settings.req_uppercase')}</span>
              </label>
              <label className="flex items-center gap-3 cursor-pointer"
                onClick={() => onTogglePolicy('requireLowercase')}>
                <span className={`transition-colors ${localPasswordPolicy.requireLowercase ? 'text-success' : 'text-muted-foreground'}`}>
                  {localPasswordPolicy.requireLowercase ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                </span>
                <span className="text-sm text-foreground">{t('settings.req_lowercase')}</span>
              </label>
              <label className="flex items-center gap-3 cursor-pointer"
                onClick={() => onTogglePolicy('requireNumbers')}>
                <span className={`transition-colors ${localPasswordPolicy.requireNumbers ? 'text-success' : 'text-muted-foreground'}`}>
                  {localPasswordPolicy.requireNumbers ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                </span>
                <span className="text-sm text-foreground">{t('settings.req_numbers')}</span>
              </label>
              <label className="flex items-center gap-3 cursor-pointer"
                onClick={() => onTogglePolicy('requireSpecial')}>
                <span className={`transition-colors ${localPasswordPolicy.requireSpecial ? 'text-success' : 'text-muted-foreground'}`}>
                  {localPasswordPolicy.requireSpecial ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                </span>
                <span className="text-sm text-foreground">{t('settings.req_special')}</span>
              </label>
            </div>
          </div>
        </div>
      )}
    </section>
  );
};
