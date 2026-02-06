import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Plus, Edit, Trash2, FileText, CheckCircle, XCircle } from 'lucide-react';
import type { SAMLServiceProvider } from '@auth-gateway/client-sdk';
import { useSAMLSPs, useDeleteSAMLSP } from '../hooks/useSAML';
import { useLanguage } from '../services/i18n';

const SAMLSPs: React.FC = () => {
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const navigate = useNavigate();
  const { t } = useLanguage();

  const { data, isLoading, error } = useSAMLSPs(page, pageSize);
  const deleteSP = useDeleteSAMLSP();

  const handleDelete = async (id: string, name: string) => {
    if (window.confirm(`${t('saml.delete_confirm')} "${name}"?`)) {
      try {
        await deleteSP.mutateAsync(id);
      } catch (error) {
        console.error('Failed to delete SAML SP:', error);
        alert(t('common.failed_to_load'));
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive">{t('saml.error_loading')}: {(error as Error).message}</p>
      </div>
    );
  }

  const sps = data?.sps || [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t('saml.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('saml.desc')}</p>
        </div>
        <div className="flex gap-2">
          <Link
            to="/saml/metadata"
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors flex items-center gap-2"
          >
            <FileText size={18} />
            {t('saml.view_metadata')}
          </Link>
          <button
            onClick={() => navigate('/saml/new')}
            className="bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
          >
            <Plus size={18} />
            {t('saml.create')}
          </button>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('saml.col_name')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('saml.col_entity_id')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('saml.col_acs_url')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('saml.col_status')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  {t('saml.col_created')}
                </th>
                <th scope="col" className="relative px-6 py-3">
                  <span className="sr-only">{t('saml.col_actions')}</span>
                </th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {sps.map((sp) => (
                <tr key={sp.id} className="hover:bg-accent transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-foreground">{sp.name}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-muted-foreground max-w-xs truncate">{sp.entity_id}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="text-sm text-muted-foreground max-w-xs truncate">{sp.acs_url}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {sp.is_active ? (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-success/10 text-success">
                        <CheckCircle size={12} />
                        {t('saml.active')}
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-muted text-foreground">
                        <XCircle size={12} />
                        {t('saml.inactive')}
                      </span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                    {new Date(sp.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <Link
                        to={`/saml/${sp.id}`}
                        className="p-1.5 text-muted-foreground hover:text-primary rounded-md hover:bg-accent"
                        title={t('common.edit')}
                      >
                        <Edit size={16} />
                      </Link>
                      <button
                        onClick={() => handleDelete(sp.id, sp.name)}
                        className="p-1.5 text-muted-foreground hover:text-destructive rounded-md hover:bg-accent"
                        title={t('common.delete')}
                        disabled={deleteSP.isPending}
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {sps.length === 0 && (
            <div className="p-12 text-center text-muted-foreground">{t('saml.no_sps')}</div>
          )}
        </div>

        {data && data.total > pageSize && (
          <div className="px-6 py-4 border-t border-border flex items-center justify-between">
            <div className="text-sm text-muted-foreground">
              {t('common.showing')} {(page - 1) * pageSize + 1} {t('common.to')} {Math.min(page * pageSize, data.total)} {t('common.of')} {data.total} {t('saml.showing_sps')}
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 border border-input rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
              >
                {t('common.previous')}
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * pageSize >= data.total}
                className="px-3 py-1 border border-input rounded-md text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
              >
                {t('common.next')}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default SAMLSPs;

