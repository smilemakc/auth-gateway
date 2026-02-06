import React, { useState, useEffect } from 'react';
import { useNavigate, useParams, Link } from 'react-router-dom';
import { Save, X, Loader, Download, FileText } from 'lucide-react';
import type { CreateSAMLSPRequest, UpdateSAMLSPRequest, SAMLServiceProvider } from '@auth-gateway/client-sdk';
import { useSAMLSP, useCreateSAMLSP, useUpdateSAMLSP } from '../hooks/useSAML';
import { toast } from '../services/toast';
import { useLanguage } from '../services/i18n';

const SAMLSPEdit: React.FC = () => {
  const { t } = useLanguage();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = !id;

  const { data: sp, isLoading: isLoadingSP } = useSAMLSP(id || '');
  const createSP = useCreateSAMLSP();
  const updateSP = useUpdateSAMLSP();

  const [formData, setFormData] = useState<CreateSAMLSPRequest>({
    name: '',
    entity_id: '',
    acs_url: '',
    slo_url: '',
    x509_cert: '',
    metadata_url: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (sp && !isNew) {
      setFormData({
        name: sp.name,
        entity_id: sp.entity_id,
        acs_url: sp.acs_url,
        slo_url: sp.slo_url || '',
        x509_cert: sp.x509_cert || '',
        metadata_url: sp.metadata_url || '',
      });
    }
  }, [sp, isNew]);

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = t('saml_edit.err_name');
    }
    if (!formData.entity_id.trim()) {
      newErrors.entity_id = t('saml_edit.err_entity_id');
    } else if (!formData.entity_id.startsWith('http://') && !formData.entity_id.startsWith('https://')) {
      newErrors.entity_id = t('saml_edit.err_entity_id_url');
    }
    if (!formData.acs_url.trim()) {
      newErrors.acs_url = t('saml_edit.err_acs_url');
    } else if (!formData.acs_url.startsWith('http://') && !formData.acs_url.startsWith('https://')) {
      newErrors.acs_url = t('saml_edit.err_acs_url_url');
    }
    if (formData.slo_url && !formData.slo_url.startsWith('http://') && !formData.slo_url.startsWith('https://')) {
      newErrors.slo_url = t('saml_edit.err_slo_url');
    }
    if (formData.metadata_url && !formData.metadata_url.startsWith('http://') && !formData.metadata_url.startsWith('https://')) {
      newErrors.metadata_url = t('saml_edit.err_metadata_url');
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    try {
      if (isNew) {
        await createSP.mutateAsync(formData);
      } else {
        const updateData: UpdateSAMLSPRequest = {
          name: formData.name,
          entity_id: formData.entity_id,
          acs_url: formData.acs_url,
          slo_url: formData.slo_url || undefined,
          x509_cert: formData.x509_cert || undefined,
          metadata_url: formData.metadata_url || undefined,
        };
        await updateSP.mutateAsync({ id: id!, data: updateData });
      }
      navigate('/saml');
    } catch (error) {
      console.error('Failed to save SAML SP:', error);
      toast.error(t('saml_edit.save_error'));
    }
  };

  if (isLoadingSP && !isNew) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">{isNew ? t('saml_edit.create_title') : t('saml_edit.edit_title')}</h1>
        <div className="flex gap-2">
          <Link
            to="/saml/metadata"
            className="px-3 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors flex items-center gap-2 text-sm"
          >
            <Download size={16} />
            {t('saml_edit.download_metadata')}
          </Link>
          <button onClick={() => navigate('/saml')} className="text-muted-foreground hover:text-foreground flex items-center gap-2">
            <X size={20} />
            {t('common.cancel')}
          </button>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border p-6 space-y-6">
        {/* Basic Information */}
        <div className="border-b border-border pb-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('saml_edit.basic_info')}</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                {t('common.name')} <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  errors.name ? 'border-destructive' : 'border-input'
                }`}
                placeholder="Salesforce"
              />
              {errors.name && <p className="mt-1 text-sm text-destructive">{errors.name}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                {t('saml_edit.entity_id')} <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                value={formData.entity_id}
                onChange={(e) => setFormData({ ...formData, entity_id: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  errors.entity_id ? 'border-destructive' : 'border-input'
                }`}
                placeholder="https://saml.salesforce.com"
              />
              {errors.entity_id && <p className="mt-1 text-sm text-destructive">{errors.entity_id}</p>}
              <p className="mt-1 text-xs text-muted-foreground">{t('saml_edit.entity_id_hint')}</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                {t('saml_edit.acs_url')} <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                value={formData.acs_url}
                onChange={(e) => setFormData({ ...formData, acs_url: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  errors.acs_url ? 'border-destructive' : 'border-input'
                }`}
                placeholder="https://saml.salesforce.com/sp/ACS"
              />
              {errors.acs_url && <p className="mt-1 text-sm text-destructive">{errors.acs_url}</p>}
              <p className="mt-1 text-xs text-muted-foreground">{t('saml_edit.acs_url_hint')}</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">{t('saml_edit.slo_url')}</label>
              <input
                type="text"
                value={formData.slo_url}
                onChange={(e) => setFormData({ ...formData, slo_url: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  errors.slo_url ? 'border-destructive' : 'border-input'
                }`}
                placeholder="https://saml.salesforce.com/sp/SLO"
              />
              {errors.slo_url && <p className="mt-1 text-sm text-destructive">{errors.slo_url}</p>}
              <p className="mt-1 text-xs text-muted-foreground">{t('saml_edit.slo_url_hint')}</p>
            </div>
          </div>
        </div>

        {/* Certificate */}
        <div className="border-b border-border pb-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('saml_edit.sp_cert')}</h2>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('saml_edit.x509_cert')}</label>
            <textarea
              value={formData.x509_cert}
              onChange={(e) => setFormData({ ...formData, x509_cert: e.target.value })}
              rows={8}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring font-mono text-xs"
              placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----"
            />
            <p className="mt-1 text-xs text-muted-foreground">{t('saml_edit.x509_hint')}</p>
          </div>
        </div>

        {/* Metadata */}
        <div>
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('saml_edit.metadata')}</h2>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('saml_edit.metadata_url')}</label>
            <input
              type="text"
              value={formData.metadata_url}
              onChange={(e) => setFormData({ ...formData, metadata_url: e.target.value })}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                errors.metadata_url ? 'border-destructive' : 'border-input'
              }`}
              placeholder="https://saml.salesforce.com/metadata"
            />
            {errors.metadata_url && <p className="mt-1 text-sm text-destructive">{errors.metadata_url}</p>}
            <p className="mt-1 text-xs text-muted-foreground">{t('saml_edit.metadata_url_hint')}</p>
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-border">
          <button
            type="button"
            onClick={() => navigate('/saml')}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={createSP.isPending || updateSP.isPending}
            className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {(createSP.isPending || updateSP.isPending) && <Loader size={16} className="animate-spin" />}
            <Save size={16} />
            {isNew ? t('saml.create') : t('common.save')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default SAMLSPEdit;

