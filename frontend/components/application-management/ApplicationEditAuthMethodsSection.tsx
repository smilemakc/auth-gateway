import React from 'react';
import {
  ToggleLeft, ToggleRight, ShieldCheck, KeyRound, Mail, Smartphone,
  Chrome, Github, Globe, Send, Key,
} from 'lucide-react';
import { useLanguage } from '../../services/i18n';

const AUTH_METHODS = [
  { value: 'password', label: 'apps.auth_methods.password', icon: KeyRound },
  { value: 'otp_email', label: 'apps.auth_methods.otp_email', icon: Mail },
  { value: 'otp_sms', label: 'apps.auth_methods.otp_sms', icon: Smartphone },
  { value: 'oauth_google', label: 'apps.auth_methods.oauth_google', icon: Chrome },
  { value: 'oauth_github', label: 'apps.auth_methods.oauth_github', icon: Github },
  { value: 'oauth_yandex', label: 'apps.auth_methods.oauth_yandex', icon: Globe },
  { value: 'oauth_telegram', label: 'apps.auth_methods.oauth_telegram', icon: Send },
  { value: 'totp', label: 'apps.auth_methods.totp', icon: ShieldCheck },
  { value: 'api_key', label: 'apps.auth_methods.api_key', icon: Key },
] as const;

interface ApplicationEditAuthMethodsSectionProps {
  selectedMethods: string[];
  isActive: boolean;
  isEditMode: boolean;
  error?: string;
  onToggleMethod: (method: string) => void;
  onToggleActive: () => void;
}

const ApplicationEditAuthMethodsSection: React.FC<ApplicationEditAuthMethodsSectionProps> = ({
  selectedMethods,
  isActive,
  isEditMode,
  error,
  onToggleMethod,
  onToggleActive,
}) => {
  const { t } = useLanguage();

  return (
    <>
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <div className="flex items-center gap-3 mb-2">
          <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
            <ShieldCheck className="text-primary" size={20} />
          </div>
          <div>
            <h2 className="text-lg font-semibold text-foreground">{t('apps.auth_methods.title')}</h2>
            <p className="text-sm text-muted-foreground">{t('apps.auth_methods.description')}</p>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 mt-4">
          {AUTH_METHODS.map(method => {
            const Icon = method.icon;
            const isSelected = selectedMethods.includes(method.value);
            return (
              <button
                type="button"
                key={method.value}
                onClick={() => onToggleMethod(method.value)}
                className={`flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors text-left ${
                  isSelected
                    ? 'border-primary bg-primary/5'
                    : 'border-border hover:border-muted-foreground/30'
                }`}
              >
                <Icon size={18} className={isSelected ? 'text-primary' : 'text-muted-foreground'} />
                <span className={`text-sm font-medium ${isSelected ? 'text-foreground' : 'text-muted-foreground'}`}>
                  {t(method.label)}
                </span>
              </button>
            );
          })}
        </div>

        {error && (
          <p className="text-destructive text-sm mt-2">{error}</p>
        )}
      </div>

      {isEditMode && (
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('apps.status')}</h2>
          <div className="flex items-center gap-3">
            <button type="button" onClick={onToggleActive}
              className={`transition-colors ${isActive ? 'text-success' : 'text-muted-foreground'}`}>
              {isActive ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
            </button>
            <span className="text-sm text-foreground">{t('apps.active')}</span>
          </div>
          <p className="text-xs text-muted-foreground mt-2">
            {t('apps.active_hint')}
          </p>
        </div>
      )}
    </>
  );
};

export default ApplicationEditAuthMethodsSection;
