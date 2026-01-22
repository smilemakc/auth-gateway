import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus, Edit2, Trash2, Eye, ToggleLeft, ToggleRight, Boxes, Users, Palette } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useApplications, useDeleteApplication, useUpdateApplication } from '../hooks/useApplications';
import type { Application } from '../types';

const Applications: React.FC = () => {
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const { t } = useLanguage();

  const { data, isLoading, error } = useApplications(page, pageSize);
  const deleteApplication = useDeleteApplication();
  const updateApplication = useUpdateApplication();

  const handleToggle = async (app: Application) => {
    await updateApplication.mutateAsync({
      id: app.id,
      data: { is_active: !app.is_active },
    });
  };

  const handleDelete = async (id: string) => {
    if (window.confirm(t('common.confirm_delete'))) {
      await deleteApplication.mutateAsync(id);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 border border-destructive rounded-lg p-4 text-destructive">
        {t('apps.load_error') || 'Failed to load applications. Please try again.'}
      </div>
    );
  }

  const applications = data?.applications || [];
  const total = data?.total || 0;
  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('apps.title') || 'Applications'}</h1>
          <p className="text-muted-foreground mt-1">{t('apps.desc') || 'Manage multi-tenant applications'}</p>
        </div>
        <Link
          to="/applications/new"
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('apps.add') || 'Add Application'}
        </Link>
      </div>

      {/* Applications Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {applications.map((app) => (
          <div key={app.id} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col">
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
                        {app.is_active ? (t('common.active') || 'Active') : (t('common.inactive') || 'Inactive')}
                      </span>
                      {app.is_system && (
                        <>
                          <span className="text-xs text-muted-foreground">|</span>
                          <span className="text-xs text-warning font-medium uppercase tracking-wide">
                            {t('apps.system') || 'System'}
                          </span>
                        </>
                      )}
                    </div>
                  </div>
                </div>
                {!app.is_system && (
                  <button
                    onClick={() => handleToggle(app)}
                    className={`transition-colors ${app.is_active ? 'text-success hover:text-success' : 'text-muted-foreground hover:text-muted-foreground'}`}
                    disabled={updateApplication.isPending}
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
                    {t('apps.name_slug') || 'Name (Slug)'}
                  </label>
                  <code className="bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono block truncate border border-border">
                    {app.name}
                  </code>
                </div>
                {app.homepage_url && (
                  <div>
                    <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">
                      {t('apps.homepage') || 'Homepage'}
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
                    {t('apps.callbacks') || 'Callback URLs'}
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
                      <span className="italic">{t('apps.none_configured') || 'None configured'}</span>
                    )}
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-muted px-6 py-4 border-t border-border flex items-center justify-between">
              <span className="text-xs text-muted-foreground">
                {new Date(app.created_at).toLocaleDateString()}
              </span>
              <div className="flex items-center gap-1">
                <Link
                  to={`/applications/${app.id}`}
                  className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                  title={t('common.view') || 'View'}
                >
                  <Eye size={18} />
                </Link>
                <Link
                  to={`/applications/${app.id}/edit`}
                  className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                  title={t('common.edit') || 'Edit'}
                >
                  <Edit2 size={18} />
                </Link>
                <Link
                  to={`/applications/${app.id}/branding`}
                  className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                  title={t('apps.branding') || 'Branding'}
                >
                  <Palette size={18} />
                </Link>
                <Link
                  to={`/applications/${app.id}/users`}
                  className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                  title={t('apps.users') || 'Users'}
                >
                  <Users size={18} />
                </Link>
                {!app.is_system && (
                  <button
                    onClick={() => handleDelete(app.id)}
                    className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                    disabled={deleteApplication.isPending}
                    title={t('common.delete') || 'Delete'}
                  >
                    <Trash2 size={18} />
                  </button>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {applications.length === 0 && (
        <div className="text-center py-12">
          <Boxes className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-medium text-foreground">{t('apps.no_apps') || 'No applications'}</h3>
          <p className="mt-1 text-sm text-muted-foreground">{t('apps.get_started') || 'Get started by creating a new application.'}</p>
          <div className="mt-6">
            <Link
              to="/applications/new"
              className="inline-flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium"
            >
              <Plus size={18} />
              {t('apps.add') || 'Add Application'}
            </Link>
          </div>
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between bg-card px-4 py-3 rounded-lg border border-border">
          <div className="text-sm text-foreground">
            {t('common.showing') || 'Showing'} <span className="font-medium">{(page - 1) * pageSize + 1}</span> {t('common.to') || 'to'}{' '}
            <span className="font-medium">{Math.min(page * pageSize, total)}</span> {t('common.of') || 'of'}{' '}
            <span className="font-medium">{total}</span> {t('common.results') || 'results'}
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.previous') || 'Previous'}
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.next') || 'Next'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Applications;
