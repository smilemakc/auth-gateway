import React from 'react';
import { Mail, RefreshCw, Loader2 } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface EmailTemplatesEmptyStateProps {
  isPending: boolean;
  onInitialize: () => void;
}

const EmailTemplatesEmptyState: React.FC<EmailTemplatesEmptyStateProps> = ({
  isPending,
  onInitialize,
}) => {
  const { t } = useLanguage();

  return (
    <div className="flex flex-col items-center justify-center py-12 bg-card rounded-xl border border-border">
      <div className="p-4 bg-muted rounded-full mb-4">
        <Mail size={48} className="text-muted-foreground" />
      </div>
      <h3 className="text-lg font-semibold text-foreground mb-2">
        {t('email_tpl.no_templates')}
      </h3>
      <p className="text-sm text-muted-foreground mb-6 text-center max-w-md">
        {t('email_tpl.no_templates_desc')}
      </p>
      <button
        onClick={onInitialize}
        disabled={isPending}
        className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
      >
        {isPending ? (
          <>
            <Loader2 className="w-4 h-4 animate-spin" />
            {t('common.loading')}
          </>
        ) : (
          <>
            <RefreshCw size={18} />
            {t('email_tpl.initialize')}
          </>
        )}
      </button>
    </div>
  );
};

export default EmailTemplatesEmptyState;
