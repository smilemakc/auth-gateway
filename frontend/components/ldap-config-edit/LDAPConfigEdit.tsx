import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Save, X, Loader } from 'lucide-react';
import { LoadingSpinner } from '../ui';
import type { CreateLDAPConfigRequest, UpdateLDAPConfigRequest } from '@auth-gateway/client-sdk';
import { useLDAPConfig, useCreateLDAPConfig, useUpdateLDAPConfig, useTestLDAPConnection } from '../../hooks/useLDAP';
import { toast } from '../../services/toast';
import { useLanguage } from '../../services/i18n';
import { logger } from '@/lib/logger';
import LDAPConnectionFields from './LDAPConnectionFields';
import LDAPSearchFields from './LDAPSearchFields';
import LDAPMappingFields from './LDAPMappingFields';

const LDAPConfigEdit: React.FC = () => {
  const { t } = useLanguage();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = !id;

  const { data: config, isLoading: isLoadingConfig } = useLDAPConfig(id || '');
  const createConfig = useCreateLDAPConfig();
  const updateConfig = useUpdateLDAPConfig();
  const testConnection = useTestLDAPConnection();

  const [formData, setFormData] = useState<CreateLDAPConfigRequest>({
    server: '',
    port: 389,
    use_tls: false,
    use_ssl: false,
    insecure: false,
    bind_dn: '',
    bind_password: '',
    base_dn: '',
    user_search_base: '',
    group_search_base: '',
    user_search_filter: '(objectClass=person)',
    group_search_filter: '(objectClass=group)',
    user_id_attribute: 'uid',
    user_email_attribute: 'mail',
    user_name_attribute: 'cn',
    group_id_attribute: 'cn',
    group_name_attribute: 'cn',
    group_member_attribute: 'member',
    sync_enabled: false,
    sync_interval: 3600,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    if (config && !isNew) {
      setFormData({
        server: config.server,
        port: config.port,
        use_tls: config.use_tls,
        use_ssl: config.use_ssl,
        insecure: config.insecure,
        bind_dn: config.bind_dn,
        bind_password: '',
        base_dn: config.base_dn,
        user_search_base: config.user_search_base || '',
        group_search_base: config.group_search_base || '',
        user_search_filter: config.user_search_filter,
        group_search_filter: config.group_search_filter,
        user_id_attribute: config.user_id_attribute,
        user_email_attribute: config.user_email_attribute,
        user_name_attribute: config.user_name_attribute,
        group_id_attribute: config.group_id_attribute,
        group_name_attribute: config.group_name_attribute,
        group_member_attribute: config.group_member_attribute,
        sync_enabled: config.sync_enabled,
        sync_interval: Math.floor(config.sync_interval),
      });
    }
  }, [config, isNew]);

  const handleFormChange = (partial: Partial<CreateLDAPConfigRequest>) => {
    setFormData(prev => ({ ...prev, ...partial }));
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.server.trim()) {
      newErrors.server = t('ldap_edit.err_server');
    }
    if (!formData.port || formData.port < 1 || formData.port > 65535) {
      newErrors.port = t('ldap_edit.err_port');
    }
    if (!formData.bind_dn.trim()) {
      newErrors.bind_dn = t('ldap_edit.err_bind_dn');
    }
    if (isNew && !formData.bind_password.trim()) {
      newErrors.bind_password = t('ldap_edit.err_bind_password');
    }
    if (!formData.base_dn.trim()) {
      newErrors.base_dn = t('ldap_edit.err_base_dn');
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleTest = async () => {
    if (!formData.server || !formData.bind_dn || !formData.base_dn) {
      setTestResult({ success: false, message: t('ldap_edit.test_fields_required') });
      return;
    }

    try {
      const result = await testConnection.mutateAsync({
        server: formData.server,
        port: formData.port,
        use_tls: formData.use_tls,
        use_ssl: formData.use_ssl,
        insecure: formData.insecure,
        bind_dn: formData.bind_dn,
        bind_password: formData.bind_password,
        base_dn: formData.base_dn,
      });
      setTestResult({
        success: result.success,
        message: result.success
          ? `${t('ldap.connection_success_msg')} ${t('nav.users')}: ${result.user_count || 0}, ${t('nav.groups')}: ${result.group_count || 0}`
          : result.error || result.message,
      });
    } catch (error) {
      setTestResult({ success: false, message: (error as Error).message });
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    try {
      if (isNew) {
        await createConfig.mutateAsync(formData);
      } else {
        const updateData: UpdateLDAPConfigRequest = {
          server: formData.server,
          port: formData.port,
          use_tls: formData.use_tls,
          use_ssl: formData.use_ssl,
          insecure: formData.insecure,
          bind_dn: formData.bind_dn,
          bind_password: formData.bind_password || undefined,
          base_dn: formData.base_dn,
          user_search_base: formData.user_search_base || undefined,
          group_search_base: formData.group_search_base || undefined,
          user_search_filter: formData.user_search_filter,
          group_search_filter: formData.group_search_filter,
          sync_enabled: formData.sync_enabled,
          sync_interval: formData.sync_interval,
        };
        await updateConfig.mutateAsync({ id: id!, data: updateData });
      }
      navigate('/ldap');
    } catch (error) {
      logger.error('Failed to save LDAP config:', error);
      toast.error(t('ldap_edit.save_error'));
    }
  };

  if (isLoadingConfig && !isNew) {
    return <LoadingSpinner className="min-h-screen" />;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">{isNew ? t('ldap_edit.create_title') : t('ldap_edit.edit_title')}</h1>
        <button onClick={() => navigate('/ldap')} className="text-muted-foreground hover:text-foreground flex items-center gap-2">
          <X size={20} />
          {t('common.cancel')}
        </button>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border p-6 space-y-6">
        <LDAPConnectionFields
          formData={formData}
          errors={errors}
          isNew={isNew}
          showPassword={showPassword}
          isTestPending={testConnection.isPending}
          testResult={testResult}
          onFormChange={handleFormChange}
          onTogglePassword={() => setShowPassword(prev => !prev)}
          onTest={handleTest}
        />

        <LDAPSearchFields
          formData={formData}
          onFormChange={handleFormChange}
        />

        <LDAPMappingFields
          formData={formData}
          onFormChange={handleFormChange}
        />

        <div className="flex justify-end gap-3 pt-4 border-t border-border">
          <button
            type="button"
            onClick={() => navigate('/ldap')}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={createConfig.isPending || updateConfig.isPending}
            className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {(createConfig.isPending || updateConfig.isPending) && <Loader size={16} className="animate-spin" />}
            <Save size={16} />
            {isNew ? t('ldap_edit.create_config') : t('common.save')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default LDAPConfigEdit;
