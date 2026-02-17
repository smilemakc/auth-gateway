import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus, Boxes } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useApplications, useDeleteApplication, useUpdateApplication } from '../../hooks/useApplications';
import type { Application } from '../../types';
import { confirm } from '../../services/confirm';
import { ApplicationCard } from './ApplicationCard';

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
    const ok = await confirm({
      title: t('confirm.delete_title'),
      description: t('common.confirm_delete'),
      variant: 'danger'
    });
    if (ok) {
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
        {t('apps.load_error')}
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
          <h1 className="text-2xl font-bold text-foreground">{t('apps.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('apps.desc')}</p>
        </div>
        <Link
          to="/applications/new"
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('apps.add')}
        </Link>
      </div>

      {/* Applications Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {applications.map((app) => (
          <ApplicationCard
            key={app.id}
            app={app}
            isToggling={updateApplication.isPending}
            isDeleting={deleteApplication.isPending}
            onToggle={handleToggle}
            onDelete={handleDelete}
          />
        ))}
      </div>

      {applications.length === 0 && (
        <div className="text-center py-12">
          <Boxes className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-medium text-foreground">{t('apps.no_apps')}</h3>
          <p className="mt-1 text-sm text-muted-foreground">{t('apps.get_started')}</p>
          <div className="mt-6">
            <Link
              to="/applications/new"
              className="inline-flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium"
            >
              <Plus size={18} />
              {t('apps.add')}
            </Link>
          </div>
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between bg-card px-4 py-3 rounded-lg border border-border">
          <div className="text-sm text-foreground">
            {t('common.showing')} <span className="font-medium">{(page - 1) * pageSize + 1}</span> {t('common.to')}{' '}
            <span className="font-medium">{Math.min(page * pageSize, total)}</span> {t('common.of')}{' '}
            <span className="font-medium">{total}</span> {t('common.results')}
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.previous')}
            </button>
            <button
              onClick={() => setPage(p => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-3 py-1 border border-input rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
            >
              {t('common.next')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Applications;
