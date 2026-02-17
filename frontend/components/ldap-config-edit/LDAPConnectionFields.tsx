import React from 'react';
import { ToggleLeft, ToggleRight, TestTube, Loader, CheckCircle, AlertCircle } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import type { CreateLDAPConfigRequest } from '@auth-gateway/client-sdk';

interface LDAPConnectionFieldsProps {
  formData: CreateLDAPConfigRequest;
  errors: Record<string, string>;
  isNew: boolean;
  showPassword: boolean;
  isTestPending: boolean;
  testResult: { success: boolean; message: string } | null;
  onFormChange: (data: Partial<CreateLDAPConfigRequest>) => void;
  onTogglePassword: () => void;
  onTest: () => void;
}

const LDAPConnectionFields: React.FC<LDAPConnectionFieldsProps> = ({
  formData,
  errors,
  isNew,
  showPassword,
  isTestPending,
  testResult,
  onFormChange,
  onTogglePassword,
  onTest,
}) => {
  const { t } = useLanguage();

  return (
    <>
      {testResult && (
        <div
          className={`p-4 rounded-lg flex items-start gap-3 ${
            testResult.success ? 'bg-success/10 border border-success/20' : 'bg-destructive/10 border border-destructive/20'
          }`}
        >
          {testResult.success ? (
            <CheckCircle className="text-success mt-0.5" size={20} />
          ) : (
            <AlertCircle className="text-destructive mt-0.5" size={20} />
          )}
          <div className="flex-1">
            <p className={`font-medium ${testResult.success ? 'text-success' : 'text-destructive'}`}>
              {testResult.success ? t('ldap.connection_success') : t('ldap.connection_failed')}
            </p>
            <p className={`text-sm mt-1 ${testResult.success ? 'text-success' : 'text-destructive'}`}>
              {testResult.message}
            </p>
          </div>
        </div>
      )}

      <div className="border-b border-border pb-6">
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('ldap_edit.connection_settings')}</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">
              {t('ldap_edit.server')} <span className="text-destructive">*</span>
            </label>
            <input
              type="text"
              value={formData.server}
              onChange={(e) => onFormChange({ server: e.target.value })}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                errors.server ? 'border-destructive' : 'border-input'
              }`}
              placeholder="ldap.example.com"
            />
            {errors.server && <p className="mt-1 text-sm text-destructive">{errors.server}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">
              {t('common.port')} <span className="text-destructive">*</span>
            </label>
            <input
              type="number"
              value={formData.port}
              onChange={(e) => onFormChange({ port: parseInt(e.target.value) || 389 })}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                errors.port ? 'border-destructive' : 'border-input'
              }`}
              min="1"
              max="65535"
            />
            {errors.port && <p className="mt-1 text-sm text-destructive">{errors.port}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">
              {t('ldap_edit.bind_dn')} <span className="text-destructive">*</span>
            </label>
            <input
              type="text"
              value={formData.bind_dn}
              onChange={(e) => onFormChange({ bind_dn: e.target.value })}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                errors.bind_dn ? 'border-destructive' : 'border-input'
              }`}
              placeholder="cn=admin,dc=example,dc=com"
            />
            {errors.bind_dn && <p className="mt-1 text-sm text-destructive">{errors.bind_dn}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">
              {t('ldap_edit.bind_password')} {isNew && <span className="text-destructive">*</span>}
            </label>
            <div className="relative">
              <input
                type={showPassword ? 'text' : 'password'}
                value={formData.bind_password}
                onChange={(e) => onFormChange({ bind_password: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  errors.bind_password ? 'border-destructive' : 'border-input'
                }`}
                placeholder={isNew ? t('ldap_edit.enter_password') : t('ldap_edit.leave_empty')}
              />
              <button
                type="button"
                onClick={onTogglePassword}
                className="absolute right-2 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                {showPassword ? t('ldap_edit.hide') : t('ldap_edit.show')}
              </button>
            </div>
            {errors.bind_password && <p className="mt-1 text-sm text-destructive">{errors.bind_password}</p>}
          </div>

          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-foreground mb-1">
              {t('ldap_edit.base_dn')} <span className="text-destructive">*</span>
            </label>
            <input
              type="text"
              value={formData.base_dn}
              onChange={(e) => onFormChange({ base_dn: e.target.value })}
              className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                errors.base_dn ? 'border-destructive' : 'border-input'
              }`}
              placeholder="dc=example,dc=com"
            />
            {errors.base_dn && <p className="mt-1 text-sm text-destructive">{errors.base_dn}</p>}
          </div>

          <div className="flex items-center gap-4">
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={() => onFormChange({ use_tls: !formData.use_tls })}
                className={`transition-colors ${formData.use_tls ? 'text-success' : 'text-muted-foreground'}`}
              >
                {formData.use_tls ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
              </button>
              <span className="text-sm text-foreground">{t('ldap_edit.use_tls')}</span>
            </div>
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={() => onFormChange({ use_ssl: !formData.use_ssl })}
                className={`transition-colors ${formData.use_ssl ? 'text-success' : 'text-muted-foreground'}`}
              >
                {formData.use_ssl ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
              </button>
              <span className="text-sm text-foreground">{t('ldap_edit.use_ssl')}</span>
            </div>
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={() => onFormChange({ insecure: !formData.insecure })}
                className={`transition-colors ${formData.insecure ? 'text-success' : 'text-muted-foreground'}`}
              >
                {formData.insecure ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
              </button>
              <span className="text-sm text-foreground">{t('ldap_edit.skip_cert')}</span>
            </div>
          </div>
        </div>

        <div className="mt-4">
          <button
            type="button"
            onClick={onTest}
            disabled={isTestPending}
            className="px-4 py-2 bg-muted hover:bg-accent text-foreground rounded-lg text-sm transition-colors flex items-center gap-2 disabled:opacity-50"
          >
            {isTestPending ? <Loader size={16} className="animate-spin" /> : <TestTube size={16} />}
            {t('ldap.test_connection')}
          </button>
        </div>
      </div>
    </>
  );
};

export default LDAPConnectionFields;
