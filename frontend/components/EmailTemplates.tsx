
import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { getEmailTemplates } from '../services/mockData';
import { EmailTemplate } from '../types';
import { Mail, Edit2, Calendar, FileText, ArrowLeft } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const EmailTemplates: React.FC = () => {
  const [templates, setTemplates] = useState<EmailTemplate[]>([]);
  const navigate = useNavigate();
  const { t } = useLanguage();

  useEffect(() => {
    setTemplates(getEmailTemplates());
  }, []);

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
                {template.variables.slice(0, 3).map(v => (
                  <span key={v} className="bg-muted text-muted-foreground px-2 py-1 rounded text-xs font-mono border border-border">
                    {v}
                  </span>
                ))}
                {template.variables.length > 3 && (
                  <span className="text-xs text-muted-foreground self-center">+{template.variables.length - 3} more</span>
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
                <span>Updated {new Date(template.updated_at).toLocaleDateString()}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default EmailTemplates;
