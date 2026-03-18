import React from 'react';
import { ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { CreateLDAPConfigRequest } from '@auth-gateway/client-sdk';

interface LDAPMappingFieldsProps {
  formData: CreateLDAPConfigRequest;
  onFormChange: (data: Partial<CreateLDAPConfigRequest>) => void;
}

const LDAPMappingFields: React.FC<LDAPMappingFieldsProps> = ({
  formData,
  onFormChange,
}) => {
  const { t } = useLanguage();

  return (
    <div>
      <h2 className="text-lg font-semibold text-foreground mb-4">{t('ldap_edit.sync_settings')}</h2>
      <div className="space-y-4">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => onFormChange({ sync_enabled: !formData.sync_enabled })}
            className={`transition-colors ${formData.sync_enabled ? 'text-success' : 'text-muted-foreground'}`}
          >
            {formData.sync_enabled ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
          </button>
          <span className="text-sm text-foreground">{t('ldap_edit.enable_auto_sync')}</span>
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-1">{t('ldap_edit.sync_interval')}</label>
          <input
            type="number"
            value={formData.sync_interval}
            onChange={(e) => onFormChange({ sync_interval: parseInt(e.target.value) || 3600 })}
            className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            min="60"
          />
          <p className="mt-1 text-xs text-muted-foreground">{t('ldap_edit.sync_min_hint')}</p>
        </div>
      </div>
    </div>
  );
};

export default LDAPMappingFields;
