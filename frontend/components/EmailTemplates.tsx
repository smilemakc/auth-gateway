import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Mail, Edit2, Calendar, FileText, Loader2, ChevronDown, ChevronRight, KeyRound, ShieldCheck } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useEmailTemplates } from '../hooks/useEmailTemplates';
import { formatDate } from '../lib/date';

interface EmailTemplatesProps {
  embedded?: boolean;
}

type TemplateType = 'verification' | 'welcome' | 'otp_login' | 'otp_registration' | 'password_reset' | '2fa' | 'custom';

interface GroupConfig {
  key: string;
  icon: React.ElementType;
  types: TemplateType[];
}

const GROUPS: GroupConfig[] = [
  {
    key: 'auth',
    icon: KeyRound,
    types: ['verification', 'welcome', 'otp_login', 'otp_registration'],
  },
  {
    key: 'security',
    icon: ShieldCheck,
    types: ['password_reset', '2fa'],
  },
  {
    key: 'other',
    icon: Mail,
    types: ['custom'],
  },
];

const TYPE_COLORS: Record<TemplateType, string> = {
  verification: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
  welcome: 'bg-green-500/10 text-green-500 border-green-500/20',
  otp_login: 'bg-indigo-500/10 text-indigo-500 border-indigo-500/20',
  otp_registration: 'bg-teal-500/10 text-teal-500 border-teal-500/20',
  password_reset: 'bg-orange-500/10 text-orange-500 border-orange-500/20',
  '2fa': 'bg-purple-500/10 text-purple-500 border-purple-500/20',
  custom: 'bg-gray-500/10 text-gray-500 border-gray-500/20',
};

const EmailTemplates: React.FC<EmailTemplatesProps> = ({ embedded = false }) => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { data: templatesResponse, isLoading, error } = useEmailTemplates();
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set(['auth', 'security', 'other']));

  const templates = templatesResponse?.templates || [];

  const toggleGroup = (groupKey: string) => {
    setExpandedGroups(prev => {
      const newSet = new Set(prev);
      if (newSet.has(groupKey)) {
        newSet.delete(groupKey);
      } else {
        newSet.add(groupKey);
      }
      return newSet;
    });
  };

  const getTemplatesByGroup = (groupTypes: TemplateType[]) => {
    return templates.filter(template => {
      const templateType = template.type as TemplateType;
      return groupTypes.includes(templateType);
    });
  };

  const getTemplateTypeLabel = (type: string): string => {
    const typeKey = `email.type_${type}`;
    return t(typeKey);
  };

  const getTemplateTypeColor = (type: string): string => {
    return TYPE_COLORS[type as TemplateType] || TYPE_COLORS.custom;
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
        {t('email.load_error')}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {!embedded && (
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('email.templates')}</h1>
        </div>
      )}

      <div className="space-y-4">
        {GROUPS.map(group => {
          const groupTemplates = getTemplatesByGroup(group.types);
          const isExpanded = expandedGroups.has(group.key);
          const GroupIcon = group.icon;

          return (
            <div key={group.key} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
              <button
                onClick={() => toggleGroup(group.key)}
                className="w-full px-6 py-4 flex items-center justify-between hover:bg-muted/50 transition-colors"
              >
                <div className="flex items-center gap-4">
                  <div className="p-2 bg-primary/10 text-primary rounded-lg">
                    <GroupIcon size={20} />
                  </div>
                  <div className="text-left">
                    <h2 className="text-lg font-bold text-foreground">
                      {t(`email.group_${group.key}`)}
                    </h2>
                    <p className="text-sm text-muted-foreground">
                      {t(`email.group_${group.key}_desc`)}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <span className="px-3 py-1 bg-primary/10 text-primary rounded-full text-sm font-semibold">
                    {groupTemplates.length}
                  </span>
                  {isExpanded ? (
                    <ChevronDown size={20} className="text-muted-foreground" />
                  ) : (
                    <ChevronRight size={20} className="text-muted-foreground" />
                  )}
                </div>
              </button>

              {isExpanded && (
                <div className="border-t border-border p-6">
                  {groupTemplates.length === 0 ? (
                    <div className="text-center py-8">
                      <Mail size={40} className="mx-auto mb-3 text-muted-foreground opacity-50" />
                      <p className="text-muted-foreground text-sm">{t('email.no_templates_in_group')}</p>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                      {groupTemplates.map((template) => (
                        <div key={template.id} className="bg-background rounded-lg border border-border overflow-hidden flex flex-col hover:border-primary transition-colors group">
                          <div className="p-4 flex-1">
                            <div className="flex items-start justify-between mb-3">
                              <div className="p-2 bg-primary/10 text-primary rounded-lg group-hover:bg-primary group-hover:text-primary-foreground transition-colors">
                                <Mail size={20} />
                              </div>
                              <Link
                                to={`/email/templates/${template.id}`}
                                className="text-muted-foreground hover:text-primary transition-colors"
                              >
                                <Edit2 size={18} />
                              </Link>
                            </div>

                            <div className="mb-2 flex flex-wrap gap-2">
                              <span className={`inline-block px-2 py-1 rounded text-xs font-semibold border ${getTemplateTypeColor(template.type)}`}>
                                {getTemplateTypeLabel(template.type)}
                              </span>
                              {template.application && (
                                <span className="inline-block px-2 py-1 rounded text-xs font-medium bg-muted text-muted-foreground border border-border">
                                  {template.application.display_name || template.application.name}
                                </span>
                              )}
                              {!template.application && (
                                <span className="inline-block px-2 py-1 rounded text-xs font-medium bg-primary/10 text-primary border border-primary/20">
                                  Global
                                </span>
                              )}
                            </div>

                            <h3 className="text-base font-bold text-foreground mb-1">{template.name}</h3>
                            <p className="text-xs text-muted-foreground mb-3 line-clamp-1">{t('email.subject')}: {template.subject}</p>

                            <div className="flex flex-wrap gap-1.5">
                              {(template.variables || []).slice(0, 3).map(v => (
                                <span key={v} className="bg-muted text-muted-foreground px-2 py-0.5 rounded text-xs font-mono border border-border">
                                  {v}
                                </span>
                              ))}
                              {(template.variables || []).length > 3 && (
                                <span className="text-xs text-muted-foreground self-center">+{(template.variables || []).length - 3}</span>
                              )}
                            </div>
                          </div>

                          <div className="bg-muted px-4 py-2.5 border-t border-border flex items-center justify-between text-xs text-muted-foreground">
                            <div className="flex items-center gap-1.5">
                              <FileText size={12} />
                              <span className="font-mono">HTML</span>
                            </div>
                            <div className="flex items-center gap-1.5">
                              <Calendar size={12} />
                              <span>{template.updated_at ? formatDate(template.updated_at) : '-'}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}

        {templates.length === 0 && (
          <div className="text-center py-12 bg-card rounded-xl border border-border">
            <Mail size={48} className="mx-auto mb-4 text-muted-foreground opacity-50" />
            <p className="text-muted-foreground">{t('email.no_templates_found')}</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default EmailTemplates;
