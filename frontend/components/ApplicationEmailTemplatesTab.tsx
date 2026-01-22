import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Mail,
  Edit2,
  Plus,
  Loader2,
  RefreshCw,
  CheckCircle,
  XCircle,
  ShieldCheck,
  KeyRound,
  UserPlus,
  LogIn,
  UserCheck
} from 'lucide-react';
import { useLanguage } from '../services/i18n';
import {
  useApplicationTemplates,
  useInitializeApplicationTemplates
} from '../hooks/useApplicationTemplates';
import type { EmailTemplate, EmailTemplateType } from '@auth-gateway/client-sdk';

interface ApplicationEmailTemplatesTabProps {
  applicationId: string;
}

const getTemplateTypeInfo = (type: EmailTemplateType): { icon: React.ReactNode; color: string; label: string } => {
  switch (type) {
    case 'verification':
      return {
        icon: <ShieldCheck size={18} />,
        color: 'text-blue-600 bg-blue-100 dark:text-blue-400 dark:bg-blue-950',
        label: 'Email Verification'
      };
    case 'password_reset':
      return {
        icon: <KeyRound size={18} />,
        color: 'text-orange-600 bg-orange-100 dark:text-orange-400 dark:bg-orange-950',
        label: 'Password Reset'
      };
    case 'welcome':
      return {
        icon: <UserPlus size={18} />,
        color: 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-950',
        label: 'Welcome Email'
      };
    case '2fa':
      return {
        icon: <ShieldCheck size={18} />,
        color: 'text-purple-600 bg-purple-100 dark:text-purple-400 dark:bg-purple-950',
        label: '2FA Code'
      };
    case 'otp_login':
      return {
        icon: <LogIn size={18} />,
        color: 'text-indigo-600 bg-indigo-100 dark:text-indigo-400 dark:bg-indigo-950',
        label: 'OTP Login'
      };
    case 'otp_registration':
      return {
        icon: <UserCheck size={18} />,
        color: 'text-teal-600 bg-teal-100 dark:text-teal-400 dark:bg-teal-950',
        label: 'OTP Registration'
      };
    case 'custom':
      return {
        icon: <Mail size={18} />,
        color: 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-800',
        label: 'Custom'
      };
    default:
      return {
        icon: <Mail size={18} />,
        color: 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-800',
        label: type
      };
  }
};

const ApplicationEmailTemplatesTab: React.FC<ApplicationEmailTemplatesTabProps> = ({ applicationId }) => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { data: templatesResponse, isLoading, error } = useApplicationTemplates(applicationId);
  const initializeTemplates = useInitializeApplicationTemplates(applicationId);

  const templates = templatesResponse?.templates || [];

  const handleInitializeTemplates = async () => {
    try {
      await initializeTemplates.mutateAsync();
    } catch (err) {
      console.error('Failed to initialize templates:', err);
    }
  };

  const handleEditTemplate = (templateId: string) => {
    navigate(`/applications/${applicationId}/email-templates/${templateId}`);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 text-destructive">
        Failed to load email templates. Please try again.
      </div>
    );
  }

  if (templates.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 bg-card rounded-xl border border-border">
        <div className="p-4 bg-muted rounded-full mb-4">
          <Mail size={48} className="text-muted-foreground" />
        </div>
        <h3 className="text-lg font-semibold text-foreground mb-2">
          No Email Templates
        </h3>
        <p className="text-sm text-muted-foreground mb-6 text-center max-w-md">
          This application doesn't have any custom email templates yet. Initialize default templates to get started.
        </p>
        <button
          onClick={handleInitializeTemplates}
          disabled={initializeTemplates.isPending}
          className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
        >
          {initializeTemplates.isPending ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              {t('common.loading') || 'Initializing...'}
            </>
          ) : (
            <>
              <RefreshCw size={18} />
              Initialize Default Templates
            </>
          )}
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-foreground">Email Templates</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Customize email templates for this application
          </p>
        </div>
        {templates.length > 0 && (
          <button
            onClick={handleInitializeTemplates}
            disabled={initializeTemplates.isPending}
            className="flex items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors disabled:opacity-50"
          >
            {initializeTemplates.isPending ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <RefreshCw size={16} />
            )}
            Reset to Defaults
          </button>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {templates.map((template: EmailTemplate) => {
          const typeInfo = getTemplateTypeInfo(template.type);

          return (
            <div
              key={template.id}
              className="bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col hover:border-primary transition-colors group"
            >
              <div className="p-6 flex-1">
                <div className="flex items-start justify-between mb-4">
                  <div className={`flex items-center gap-2 px-3 py-1.5 rounded-lg ${typeInfo.color}`}>
                    {typeInfo.icon}
                    <span className="text-xs font-medium">{typeInfo.label}</span>
                  </div>
                  <button
                    onClick={() => handleEditTemplate(template.id)}
                    className="text-muted-foreground hover:text-primary transition-colors"
                    title="Edit template"
                  >
                    <Edit2 size={18} />
                  </button>
                </div>

                <h3 className="text-lg font-semibold text-foreground mb-2">{template.name}</h3>
                <p className="text-sm text-muted-foreground mb-4 line-clamp-1">
                  <span className="font-medium">{t('email.subject') || 'Subject'}:</span> {template.subject}
                </p>

                <div className="space-y-3">
                  {/* Status Badge */}
                  <div className="flex items-center gap-2">
                    {template.is_active ? (
                      <div className="flex items-center gap-1.5 text-xs text-green-600 dark:text-green-400">
                        <CheckCircle size={14} />
                        <span className="font-medium">Active</span>
                      </div>
                    ) : (
                      <div className="flex items-center gap-1.5 text-xs text-gray-500">
                        <XCircle size={14} />
                        <span className="font-medium">Inactive</span>
                      </div>
                    )}
                  </div>

                  {/* Variables */}
                  {template.variables && template.variables.length > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-2 font-medium">
                        {t('email.vars') || 'Variables'}:
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
                  onClick={() => handleEditTemplate(template.id)}
                  className="w-full flex items-center justify-center gap-2 text-sm font-medium text-foreground hover:text-primary transition-colors"
                >
                  <Edit2 size={14} />
                  Edit Template
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ApplicationEmailTemplatesTab;
