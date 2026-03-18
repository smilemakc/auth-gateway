import React from 'react';
import {
  Mail, Edit2, CheckCircle, XCircle, ShieldCheck,
  KeyRound, UserPlus, LogIn, UserCheck,
} from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { EmailTemplate, EmailTemplateType } from '@auth-gateway/client-sdk';

interface EmailTemplateCardProps {
  template: EmailTemplate;
  onEdit: (templateId: string) => void;
}

const getTemplateTypeInfo = (type: EmailTemplateType): { icon: React.ReactNode; color: string; label: string } => {
  switch (type) {
    case 'verification':
      return {
        icon: <ShieldCheck size={18} />,
        color: 'text-blue-600 bg-blue-100 dark:text-blue-400 dark:bg-blue-950',
        label: 'email_tpl.verification'
      };
    case 'password_reset':
      return {
        icon: <KeyRound size={18} />,
        color: 'text-orange-600 bg-orange-100 dark:text-orange-400 dark:bg-orange-950',
        label: 'email_tpl.password_reset'
      };
    case 'welcome':
      return {
        icon: <UserPlus size={18} />,
        color: 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-950',
        label: 'email_tpl.welcome'
      };
    case '2fa':
      return {
        icon: <ShieldCheck size={18} />,
        color: 'text-purple-600 bg-purple-100 dark:text-purple-400 dark:bg-purple-950',
        label: 'email_tpl.2fa_code'
      };
    case 'otp_login':
      return {
        icon: <LogIn size={18} />,
        color: 'text-indigo-600 bg-indigo-100 dark:text-indigo-400 dark:bg-indigo-950',
        label: 'email_tpl.otp_login'
      };
    case 'otp_registration':
      return {
        icon: <UserCheck size={18} />,
        color: 'text-teal-600 bg-teal-100 dark:text-teal-400 dark:bg-teal-950',
        label: 'email_tpl.otp_registration'
      };
    case 'custom':
      return {
        icon: <Mail size={18} />,
        color: 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-800',
        label: 'email_tpl.custom'
      };
    default:
      return {
        icon: <Mail size={18} />,
        color: 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-800',
        label: type
      };
  }
};

const EmailTemplateCard: React.FC<EmailTemplateCardProps> = ({ template, onEdit }) => {
  const { t } = useLanguage();
  const typeInfo = getTemplateTypeInfo(template.type);

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col hover:border-primary transition-colors group">
      <div className="p-6 flex-1">
        <div className="flex items-start justify-between mb-4">
          <div className={`flex items-center gap-2 px-3 py-1.5 rounded-lg ${typeInfo.color}`}>
            {typeInfo.icon}
            <span className="text-xs font-medium">{t(typeInfo.label)}</span>
          </div>
          <button
            onClick={() => onEdit(template.id)}
            className="text-muted-foreground hover:text-primary transition-colors"
            title={t('email_tpl.edit')}
          >
            <Edit2 size={18} />
          </button>
        </div>

        <h3 className="text-lg font-semibold text-foreground mb-2">{template.name}</h3>
        <p className="text-sm text-muted-foreground mb-4 line-clamp-1">
          <span className="font-medium">{t('email.subject')}:</span> {template.subject}
        </p>

        <div className="space-y-3">
          <div className="flex items-center gap-2">
            {template.is_active ? (
              <div className="flex items-center gap-1.5 text-xs text-green-600 dark:text-green-400">
                <CheckCircle size={14} />
                <span className="font-medium">{t('common.active')}</span>
              </div>
            ) : (
              <div className="flex items-center gap-1.5 text-xs text-gray-500">
                <XCircle size={14} />
                <span className="font-medium">{t('common.inactive')}</span>
              </div>
            )}
          </div>

          {template.variables && template.variables.length > 0 && (
            <div>
              <p className="text-xs text-muted-foreground mb-2 font-medium">
                {t('email.vars')}:
              </p>
              <div className="flex flex-wrap gap-1.5">
                {template.variables.slice(0, 4).map((v: string) => (
                  <span
                    key={v}
                    className="bg-muted text-muted-foreground px-2 py-0.5 rounded text-xs font-mono border border-border"
                  >
                    {v}
                  </span>
                ))}
                {template.variables.length > 4 && (
                  <span className="text-xs text-muted-foreground self-center">
                    +{template.variables.length - 4}
                  </span>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="bg-muted px-6 py-3 border-t border-border">
        <button
          onClick={() => onEdit(template.id)}
          className="w-full flex items-center justify-center gap-2 text-sm font-medium text-foreground hover:text-primary transition-colors"
        >
          <Edit2 size={14} />
          {t('email_tpl.edit')}
        </button>
      </div>
    </div>
  );
};

export default EmailTemplateCard;
