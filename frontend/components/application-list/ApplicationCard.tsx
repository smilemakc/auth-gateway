import React from 'react';
import { Link } from 'react-router-dom';
import { Edit2, Trash2, Eye, ToggleLeft, ToggleRight, Boxes, Users, Palette } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { Application } from '../../types';
import { formatDate } from '../../lib/date';

interface ApplicationCardProps {
  app: Application;
  isToggling: boolean;
  isDeleting: boolean;
  onToggle: (app: Application) => void;
  onDelete: (id: string) => void;
}

export const ApplicationCard: React.FC<ApplicationCardProps> = ({
  app,
  isToggling,
  isDeleting,
  onToggle,
  onDelete,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col">
      <div className="p-6 flex-1">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center shadow-sm">
              {app.branding?.logo_url ? (
                <img src={app.branding.logo_url} alt={app.display_name} className="w-8 h-8 object-contain" />
              ) : (
                <Boxes className="text-primary" size={24} />
              )}
            </div>
            <div>
              <h3 className="font-semibold text-foreground text-lg">{app.display_name}</h3>
              <div className="flex items-center gap-2 mt-1">
                <span className={`w-2 h-2 rounded-full ${app.is_active ? 'bg-success' : 'bg-muted-foreground'}`}></span>
                <span className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                  {app.is_active ? t('common.active') : t('common.inactive')}
                </span>
                {app.is_system && (
                  <>
                    <span className="text-xs text-muted-foreground">|</span>
                    <span className="text-xs text-warning font-medium uppercase tracking-wide">
                      {t('apps.system')}
                    </span>
                  </>
                )}
              </div>
            </div>
          </div>
          {!app.is_system && (
            <button
              onClick={() => onToggle(app)}
              className={`transition-colors ${app.is_active ? 'text-success hover:text-success' : 'text-muted-foreground hover:text-muted-foreground'}`}
              disabled={isToggling}
            >
              {app.is_active ? <ToggleRight size={36} /> : <ToggleLeft size={36} />}
            </button>
          )}
        </div>

        {app.description && (
          <p className="text-sm text-muted-foreground mb-4 line-clamp-2">{app.description}</p>
        )}

        <div className="space-y-3">
          <div>
            <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
              {t('apps.name_slug')}
            </label>
            <code className="bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono block truncate border border-border">
              {app.name}
            </code>
          </div>
          {app.homepage_url && (
            <div>
              <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
                {t('apps.homepage')}
              </label>
              <a
                href={app.homepage_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-primary hover:underline truncate block"
              >
                {app.homepage_url}
              </a>
            </div>
          )}
          <div>
            <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
              {t('apps.callbacks')}
            </label>
            <div className="text-xs text-muted-foreground truncate">
              {app.callback_urls && app.callback_urls.length > 0 ? (
                <>
                  {app.callback_urls[0]}
                  {app.callback_urls.length > 1 && (
                    <span className="text-muted-foreground"> (+{app.callback_urls.length - 1} more)</span>
                  )}
                </>
              ) : (
                <span className="italic">{t('apps.none_configured')}</span>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="bg-muted px-6 py-4 border-t border-border flex items-center justify-between">
        <span className="text-xs text-muted-foreground">
          {formatDate(app.created_at)}
        </span>
        <div className="flex items-center gap-1">
          <Link
            to={`/applications/${app.id}`}
            className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
            title={t('common.view')}
          >
            <Eye size={18} />
          </Link>
          <Link
            to={`/applications/${app.id}/edit`}
            className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
            title={t('common.edit')}
          >
            <Edit2 size={18} />
          </Link>
          <Link
            to={`/applications/${app.id}?tab=branding`}
            className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
            title={t('apps.branding')}
          >
            <Palette size={18} />
          </Link>
          <Link
            to={`/applications/${app.id}?tab=users`}
            className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
            title={t('apps.users')}
          >
            <Users size={18} />
          </Link>
          {!app.is_system && (
            <button
              onClick={() => onDelete(app.id)}
              className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
              disabled={isDeleting}
              title={t('common.delete')}
            >
              <Trash2 size={18} />
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
