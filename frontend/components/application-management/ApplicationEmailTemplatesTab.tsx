import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Loader2, RefreshCw } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import {
  useApplicationTemplates,
  useInitializeApplicationTemplates
} from '../../hooks/useApplicationTemplates';
import type { EmailTemplate } from '@auth-gateway/client-sdk';
import { logger } from '@/lib/logger';
import EmailTemplatesEmptyState from './EmailTemplatesEmptyState';
import EmailTemplateCard from './EmailTemplateCard';

interface ApplicationEmailTemplatesTabProps {
  applicationId: string;
}

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
      logger.error('Failed to initialize templates:', err);
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
        {t('email_tpl.load_error')}
      </div>
    );
  }

  if (templates.length === 0) {
    return (
      <EmailTemplatesEmptyState
        isPending={initializeTemplates.isPending}
        onInitialize={handleInitializeTemplates}
      />
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold text-foreground">{t('email_tpl.title')}</h2>
          <p className="text-sm text-muted-foreground mt-1">
            {t('email_tpl.desc')}
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
            {t('email_tpl.reset_defaults')}
          </button>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {templates.map((template: EmailTemplate) => (
          <EmailTemplateCard
            key={template.id}
            template={template}
            onEdit={handleEditTemplate}
          />
        ))}
      </div>
    </div>
  );
};

export default ApplicationEmailTemplatesTab;
