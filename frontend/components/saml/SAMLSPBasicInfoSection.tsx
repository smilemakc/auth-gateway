import React from 'react';
import { useLanguage } from '../../services/i18n';
import type { CreateSAMLSPRequest } from '@auth-gateway/client-sdk';

interface SAMLSPBasicInfoSectionProps {
  formData: CreateSAMLSPRequest;
  errors: Record<string, string>;
  onChange: (field: keyof CreateSAMLSPRequest, value: string) => void;
}

export const SAMLSPBasicInfoSection: React.FC<SAMLSPBasicInfoSectionProps> = ({
  formData,
  errors,
  onChange,
}) => {
  const { t } = useLanguage();

  return (
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
            onChange={(e) => onChange('name', e.target.value)}
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
            onChange={(e) => onChange('entity_id', e.target.value)}
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
            onChange={(e) => onChange('acs_url', e.target.value)}
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
            onChange={(e) => onChange('slo_url', e.target.value)}
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
  );
};
