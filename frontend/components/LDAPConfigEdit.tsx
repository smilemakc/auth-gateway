import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Save, X, TestTube, Loader, AlertCircle, CheckCircle } from 'lucide-react';
import type { CreateLDAPConfigRequest, UpdateLDAPConfigRequest, LDAPConfig } from '@auth-gateway/client-sdk';
import { useLDAPConfig, useCreateLDAPConfig, useUpdateLDAPConfig, useTestLDAPConnection } from '../hooks/useLDAP';

const LDAPConfigEdit: React.FC = () => {
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
        bind_password: '', // Don't load password
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

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.server.trim()) {
      newErrors.server = 'Server is required';
    }
    if (!formData.port || formData.port < 1 || formData.port > 65535) {
      newErrors.port = 'Port must be between 1 and 65535';
    }
    if (!formData.bind_dn.trim()) {
      newErrors.bind_dn = 'Bind DN is required';
    }
    if (isNew && !formData.bind_password.trim()) {
      newErrors.bind_password = 'Bind password is required';
    }
    if (!formData.base_dn.trim()) {
      newErrors.base_dn = 'Base DN is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleTest = async () => {
    if (!formData.server || !formData.bind_dn || !formData.base_dn) {
      setTestResult({ success: false, message: 'Please fill in server, bind DN, and base DN' });
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
          ? `Connection successful! Users: ${result.user_count || 0}, Groups: ${result.group_count || 0}`
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
      console.error('Failed to save LDAP config:', error);
      alert('Failed to save LDAP configuration');
    }
  };

  if (isLoadingConfig && !isNew) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">{isNew ? 'Create LDAP Configuration' : 'Edit LDAP Configuration'}</h1>
        <button onClick={() => navigate('/ldap')} className="text-muted-foreground hover:text-foreground flex items-center gap-2">
          <X size={20} />
          Cancel
        </button>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border p-6 space-y-6">
        {/* Test Result */}
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
                {testResult.success ? 'Connection Successful' : 'Connection Failed'}
              </p>
              <p className={`text-sm mt-1 ${testResult.success ? 'text-success' : 'text-destructive'}`}>
                {testResult.message}
              </p>
            </div>
          </div>
        )}

        {/* Basic Settings */}
        <div className="border-b border-border pb-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">Connection Settings</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                Server <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                value={formData.server}
                onChange={(e) => setFormData({ ...formData, server: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  errors.server ? 'border-destructive' : 'border-input'
                }`}
                placeholder="ldap.example.com"
              />
              {errors.server && <p className="mt-1 text-sm text-destructive">{errors.server}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                Port <span className="text-destructive">*</span>
              </label>
              <input
                type="number"
                value={formData.port}
                onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) || 389 })}
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
                Bind DN <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                value={formData.bind_dn}
                onChange={(e) => setFormData({ ...formData, bind_dn: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  errors.bind_dn ? 'border-destructive' : 'border-input'
                }`}
                placeholder="cn=admin,dc=example,dc=com"
              />
              {errors.bind_dn && <p className="mt-1 text-sm text-destructive">{errors.bind_dn}</p>}
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                Bind Password {isNew && <span className="text-destructive">*</span>}
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={formData.bind_password}
                  onChange={(e) => setFormData({ ...formData, bind_password: e.target.value })}
                  className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                    errors.bind_password ? 'border-destructive' : 'border-input'
                  }`}
                  placeholder={isNew ? 'Enter password' : 'Leave empty to keep current'}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? 'Hide' : 'Show'}
                </button>
              </div>
              {errors.bind_password && <p className="mt-1 text-sm text-destructive">{errors.bind_password}</p>}
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-foreground mb-1">
                Base DN <span className="text-destructive">*</span>
              </label>
              <input
                type="text"
                value={formData.base_dn}
                onChange={(e) => setFormData({ ...formData, base_dn: e.target.value })}
                className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  errors.base_dn ? 'border-destructive' : 'border-input'
                }`}
                placeholder="dc=example,dc=com"
              />
              {errors.base_dn && <p className="mt-1 text-sm text-destructive">{errors.base_dn}</p>}
            </div>

            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={formData.use_tls}
                  onChange={(e) => setFormData({ ...formData, use_tls: e.target.checked })}
                  className="rounded border-input text-primary focus:ring-ring"
                />
                <span className="text-sm text-foreground">Use TLS</span>
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={formData.use_ssl}
                  onChange={(e) => setFormData({ ...formData, use_ssl: e.target.checked })}
                  className="rounded border-input text-primary focus:ring-ring"
                />
                <span className="text-sm text-foreground">Use SSL</span>
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={formData.insecure}
                  onChange={(e) => setFormData({ ...formData, insecure: e.target.checked })}
                  className="rounded border-input text-primary focus:ring-ring"
                />
                <span className="text-sm text-foreground">Skip certificate verification</span>
              </label>
            </div>
          </div>

          <div className="mt-4">
            <button
              type="button"
              onClick={handleTest}
              disabled={testConnection.isPending}
              className="px-4 py-2 bg-muted hover:bg-accent text-foreground rounded-lg text-sm transition-colors flex items-center gap-2 disabled:opacity-50"
            >
              {testConnection.isPending ? <Loader size={16} className="animate-spin" /> : <TestTube size={16} />}
              Test Connection
            </button>
          </div>
        </div>

        {/* User Search Settings */}
        <div className="border-b border-border pb-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">User Search Settings</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">User Search Base</label>
              <input
                type="text"
                value={formData.user_search_base}
                onChange={(e) => setFormData({ ...formData, user_search_base: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="ou=users,dc=example,dc=com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">User Search Filter</label>
              <input
                type="text"
                value={formData.user_search_filter}
                onChange={(e) => setFormData({ ...formData, user_search_filter: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="(objectClass=person)"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">User ID Attribute</label>
              <input
                type="text"
                value={formData.user_id_attribute}
                onChange={(e) => setFormData({ ...formData, user_id_attribute: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="uid"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">User Email Attribute</label>
              <input
                type="text"
                value={formData.user_email_attribute}
                onChange={(e) => setFormData({ ...formData, user_email_attribute: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="mail"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">User Name Attribute</label>
              <input
                type="text"
                value={formData.user_name_attribute}
                onChange={(e) => setFormData({ ...formData, user_name_attribute: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="cn"
              />
            </div>
          </div>
        </div>

        {/* Group Search Settings */}
        <div className="border-b border-border pb-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">Group Search Settings</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Group Search Base</label>
              <input
                type="text"
                value={formData.group_search_base}
                onChange={(e) => setFormData({ ...formData, group_search_base: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="ou=groups,dc=example,dc=com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Group Search Filter</label>
              <input
                type="text"
                value={formData.group_search_filter}
                onChange={(e) => setFormData({ ...formData, group_search_filter: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="(objectClass=group)"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Group ID Attribute</label>
              <input
                type="text"
                value={formData.group_id_attribute}
                onChange={(e) => setFormData({ ...formData, group_id_attribute: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="cn"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Group Name Attribute</label>
              <input
                type="text"
                value={formData.group_name_attribute}
                onChange={(e) => setFormData({ ...formData, group_name_attribute: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="cn"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Group Member Attribute</label>
              <input
                type="text"
                value={formData.group_member_attribute}
                onChange={(e) => setFormData({ ...formData, group_member_attribute: e.target.value })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="member"
              />
            </div>
          </div>
        </div>

        {/* Sync Settings */}
        <div>
          <h2 className="text-lg font-semibold text-foreground mb-4">Synchronization Settings</h2>
          <div className="space-y-4">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={formData.sync_enabled}
                onChange={(e) => setFormData({ ...formData, sync_enabled: e.target.checked })}
                className="rounded border-input text-primary focus:ring-ring"
              />
              <span className="text-sm text-foreground">Enable automatic synchronization</span>
            </label>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Sync Interval (seconds)</label>
              <input
                type="number"
                value={formData.sync_interval}
                onChange={(e) => setFormData({ ...formData, sync_interval: parseInt(e.target.value) || 3600 })}
                className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                min="60"
              />
              <p className="mt-1 text-xs text-muted-foreground">Minimum: 60 seconds (1 minute)</p>
            </div>
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-border">
          <button
            type="button"
            onClick={() => navigate('/ldap')}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createConfig.isPending || updateConfig.isPending}
            className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {(createConfig.isPending || updateConfig.isPending) && <Loader size={16} className="animate-spin" />}
            <Save size={16} />
            {isNew ? 'Create Configuration' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default LDAPConfigEdit;

