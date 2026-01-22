
import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Mail, Edit2, Calendar, FileText, ArrowLeft, Loader2 } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useEmailTemplates } from '../hooks/useEmailTemplates';

const EmailTemplates: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { data: templatesResponse, isLoading, error } = useEmailTemplates();

  const templates = templatesResponse?.templates || [];

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

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/settings')}
          className="p-2 hover:bg-card rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
           <h1 className="text-2xl font-bold text-foreground">{t('email.templates')}</h1>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {templates.map((template) => (
          <div key={template.id} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col hover:border-primary transition-colors group">
            <div className="p-6 flex-1">
              <div className="flex items-start justify-between mb-4">
                <div className="p-3 bg-primary/10 text-primary rounded-lg group-hover:bg-primary group-hover:text-primary-foreground transition-colors">
                   <Mail size={24} />
                </div>
                <Link
                   to={`/settings/email-templates/${template.id}`}
                   className="text-muted-foreground hover:text-primary transition-colors"
                >
                  <Edit2 size={20} />
                </Link>
              </div>

              <h3 className="text-lg font-bold text-foreground mb-1">{template.name}</h3>
              <p className="text-sm text-muted-foreground mb-4 line-clamp-1">{t('email.subject')}: {template.subject}</p>

              <div className="flex flex-wrap gap-2 mt-4">
                {(template.variables || []).slice(0, 3).map(v => (
                  <span key={v} className="bg-muted text-muted-foreground px-2 py-1 rounded text-xs font-mono border border-border">
                    {v}
                  </span>
                ))}
                {(template.variables || []).length > 3 && (
                  <span className="text-xs text-muted-foreground self-center">+{(template.variables || []).length - 3} more</span>
                )}
              </div>
            </div>

            <div className="bg-muted px-6 py-4 border-t border-border flex items-center justify-between text-xs text-muted-foreground">
              <div className="flex items-center gap-2">
                <FileText size={14} />
                <span className="font-mono">HTML</span>
              </div>
              <div className="flex items-center gap-2">
                <Calendar size={14} />
                <span>Updated {template.updated_at ? new Date(template.updated_at).toLocaleDateString() : '-'}</span>
              </div>
            </div>
          </div>
        ))}

        {templates.length === 0 && (
          <div className="col-span-full text-center py-12 bg-card rounded-xl border border-border">
            <Mail size={48} className="mx-auto mb-4 text-muted-foreground opacity-50" />
            <p className="text-muted-foreground">No email templates found.</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default EmailTemplates;
