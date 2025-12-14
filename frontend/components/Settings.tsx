
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Save, Lock, Mail, Globe, Shield, ExternalLink, ChevronRight, ShieldAlert, Palette, Key, MessageSquare, Activity, Server, Sliders } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useSystemStatus, usePasswordPolicy, useUpdatePasswordPolicy, useMaintenanceMode, useMaintenanceModeStatus } from '../hooks/useSettings';
import { PasswordPolicy } from '../types';

const Settings: React.FC = () => {
  const { t } = useLanguage();

  // Fetch data with React Query
  const { data: systemStatus, isLoading: statusLoading, error: statusError } = useSystemStatus();
  const { data: maintenanceStatus, isLoading: maintenanceLoading } = useMaintenanceModeStatus();
  const { data: apiPasswordPolicy, isLoading: policyLoading, error: policyError } = usePasswordPolicy();
  const updatePolicyMutation = useUpdatePasswordPolicy();
  const maintenanceMutation = useMaintenanceMode();

  // Local state for form editing
  const [localPasswordPolicy, setLocalPasswordPolicy] = useState<PasswordPolicy | null>(null);

  // Sync API data to local state when it loads
  useEffect(() => {
    if (apiPasswordPolicy) {
      setLocalPasswordPolicy(apiPasswordPolicy);
    }
  }, [apiPasswordPolicy]);

  const toggleMaintenance = async () => {
    if (maintenanceStatus) {
      const newStatus = !maintenanceStatus.enabled;
      if (window.confirm(newStatus ? t('sys.confirm_enable') : t('sys.confirm_disable'))) {
        try {
          await maintenanceMutation.mutateAsync({
            enabled: newStatus,
            message: 'System maintenance in progress',
          });
        } catch (error) {
          console.error('Failed to toggle maintenance mode:', error);
          alert('Failed to update maintenance mode');
        }
      }
    }
  };

  const handlePolicyChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!localPasswordPolicy) return;
    const { name, value, type, checked } = e.target;
    setLocalPasswordPolicy({
      ...localPasswordPolicy,
      [name]: type === 'checkbox' ? checked : parseInt(value)
    });
  };

  const handleSave = async () => {
    if (localPasswordPolicy) {
      try {
        await updatePolicyMutation.mutateAsync(localPasswordPolicy);
        alert(t('common.saved'));
      } catch (error) {
        console.error('Failed to save password policy:', error);
        alert('Failed to save settings');
      }
    }
  };

  if (statusLoading || policyLoading || maintenanceLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (statusError || policyError) {
    return (
      <div className="p-8 text-center">
        <p className="text-red-600">
          Error loading settings: {((statusError || policyError) as Error).message}
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{t('settings.title')}</h1>
        <button
          onClick={handleSave}
          disabled={updatePolicyMutation.isPending}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Save size={18} />
          {updatePolicyMutation.isPending ? 'Saving...' : t('common.save')}
        </button>
      </div>

      {/* System Status & Maintenance */}
      <section className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-100 flex items-center justify-between">
          <div className="flex items-center gap-3">
             <div className="p-2 bg-gray-100 text-gray-600 rounded-lg">
                <Server size={20} />
             </div>
             <div>
                <h2 className="text-lg font-semibold text-gray-900">{t('sys.health')}</h2>
                <div className="flex items-center gap-2 text-sm text-gray-500 mt-1">
                   <span className={`inline-block w-2 h-2 rounded-full ${systemStatus?.status === 'healthy' ? 'bg-green-500' : 'bg-red-500'}`}></span>
                   Status: <span className="uppercase font-medium">{systemStatus?.status}</span>
                   <span className="mx-1">•</span>
                   <span>DB: {systemStatus?.services?.database || 'unknown'}</span>
                   <span className="mx-1">•</span>
                   <span>Redis: {systemStatus?.services?.redis || 'unknown'}</span>
                </div>
             </div>
          </div>
          <div className="flex items-center gap-3">
             <div className="flex items-center">
                <span className={`mr-2 text-sm font-medium ${maintenanceStatus?.enabled ? 'text-orange-600' : 'text-gray-500'}`}>
                  {maintenanceStatus?.enabled ? t('sys.maintenance_on') : t('sys.maintenance_off')}
                </span>
                <button
                  onClick={toggleMaintenance}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${
                    maintenanceStatus?.enabled ? 'bg-orange-600' : 'bg-gray-200'
                  }`}
                >
                  <span
                    className={`${
                      maintenanceStatus?.enabled ? 'translate-x-6' : 'translate-x-1'
                    } inline-block h-4 w-4 transform rounded-full bg-white transition-transform`}
                  />
                </button>
             </div>
          </div>
        </div>
      </section>

      {/* Branding Section */}
      <section className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-100 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-pink-50 text-pink-600 rounded-lg">
               <Palette size={20} />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900">{t('settings.branding')}</h2>
              <p className="text-sm text-gray-500">{t('settings.branding_desc')}</p>
            </div>
          </div>
          <Link 
            to="/settings/branding" 
            className="text-sm bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 px-4 py-2 rounded-md font-medium transition-colors flex items-center gap-2"
          >
            {t('oauth.configure')} <ChevronRight size={16} />
          </Link>
        </div>
      </section>

      {/* Access Control Section */}
      <section className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-100 flex items-center gap-3">
          <div className="p-2 bg-indigo-50 text-indigo-600 rounded-lg">
             <Shield size={20} />
          </div>
          <h2 className="text-lg font-semibold text-gray-900">{t('settings.roles_desc')}</h2>
        </div>
        <div className="divide-y divide-gray-100">
           <div className="p-6 flex items-center justify-between hover:bg-gray-50 transition-colors">
              <div>
                <h3 className="text-md font-medium text-gray-900">{t('roles.title')}</h3>
                <p className="text-sm text-gray-500">Define user roles and assign permission sets.</p>
              </div>
              <Link 
                to="/settings/roles" 
                className="text-sm bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 px-4 py-2 rounded-md font-medium transition-colors flex items-center gap-2"
              >
                {t('common.edit')} <ChevronRight size={16} />
              </Link>
           </div>
           <div className="p-6 flex items-center justify-between hover:bg-gray-50 transition-colors">
              <div>
                <h3 className="text-md font-medium text-gray-900">{t('perms.title')}</h3>
                <p className="text-sm text-gray-500">Create granular permissions for system resources.</p>
              </div>
              <Link 
                to="/settings/permissions" 
                className="text-sm bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 px-4 py-2 rounded-md font-medium transition-colors flex items-center gap-2"
              >
                {t('common.edit')} <ChevronRight size={16} />
              </Link>
           </div>
        </div>
      </section>

      {/* IP Security Section */}
      <section className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-100 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-red-50 text-red-600 rounded-lg">
               <ShieldAlert size={20} />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900">{t('settings.ip_security')}</h2>
              <p className="text-sm text-gray-500">{t('settings.ip_desc')}</p>
            </div>
          </div>
          <Link 
            to="/settings/security/ip-rules" 
            className="text-sm bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 px-4 py-2 rounded-md font-medium transition-colors flex items-center gap-2"
          >
            {t('oauth.configure')} <ChevronRight size={16} />
          </Link>
        </div>
      </section>

      {/* Security Policies */}
      <section className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-100 flex items-center gap-3">
          <div className="p-2 bg-blue-50 text-blue-600 rounded-lg">
            <Lock size={20} />
          </div>
          <h2 className="text-lg font-semibold text-gray-900">{t('settings.security_policies')}</h2>
        </div>
        {localPasswordPolicy && (
          <div className="p-6 space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">{t('settings.jwt_ttl')}</label>
                <input
                  type="number"
                  name="jwtTtlMinutes"
                  value={localPasswordPolicy.jwtTtlMinutes}
                  onChange={handlePolicyChange}
                  className="w-full border-gray-300 border rounded-lg p-2.5 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">{t('settings.refresh_ttl')}</label>
                <input
                  type="number"
                  name="refreshTtlDays"
                  value={localPasswordPolicy.refreshTtlDays}
                  onChange={handlePolicyChange}
                  className="w-full border-gray-300 border rounded-lg p-2.5 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>

            <div className="border-t border-gray-100 pt-6">
              <h3 className="text-md font-medium text-gray-900 mb-4">{t('settings.password_policy')}</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">{t('settings.min_pass')}</label>
                  <input
                    type="number"
                    name="minLength"
                    value={localPasswordPolicy.minLength}
                    onChange={handlePolicyChange}
                    className="w-full border-gray-300 border rounded-lg p-2.5 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">{t('settings.pass_history')}</label>
                  <input
                    type="number"
                    name="historyCount"
                    value={localPasswordPolicy.historyCount}
                    onChange={handlePolicyChange}
                    className="w-full border-gray-300 border rounded-lg p-2.5 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                 <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">{t('settings.pass_expiry')}</label>
                  <input
                    type="number"
                    name="expiryDays"
                    value={localPasswordPolicy.expiryDays}
                    onChange={handlePolicyChange}
                    className="w-full border-gray-300 border rounded-lg p-2.5 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <label className="flex items-center cursor-pointer">
                   <input type="checkbox" name="requireUppercase" checked={localPasswordPolicy.requireUppercase} onChange={handlePolicyChange} className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500" />
                   <span className="ms-2 text-sm text-gray-700">{t('settings.req_uppercase')}</span>
                </label>
                <label className="flex items-center cursor-pointer">
                   <input type="checkbox" name="requireLowercase" checked={localPasswordPolicy.requireLowercase} onChange={handlePolicyChange} className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500" />
                   <span className="ms-2 text-sm text-gray-700">{t('settings.req_lowercase')}</span>
                </label>
                <label className="flex items-center cursor-pointer">
                   <input type="checkbox" name="requireNumbers" checked={localPasswordPolicy.requireNumbers} onChange={handlePolicyChange} className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500" />
                   <span className="ms-2 text-sm text-gray-700">{t('settings.req_numbers')}</span>
                </label>
                <label className="flex items-center cursor-pointer">
                   <input type="checkbox" name="requireSpecial" checked={localPasswordPolicy.requireSpecial} onChange={handlePolicyChange} className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500" />
                   <span className="ms-2 text-sm text-gray-700">{t('settings.req_special')}</span>
                </label>
              </div>
            </div>
          </div>
        )}
      </section>

      {/* OAuth Providers */}
      <section className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-100 flex items-center gap-3">
          <div className="p-2 bg-purple-50 text-purple-600 rounded-lg">
             <Globe size={20} />
          </div>
          <h2 className="text-lg font-semibold text-gray-900">{t('oauth.title')}</h2>
        </div>
        <div className="p-6">
           <div className="space-y-4">
             {['Google', 'GitHub', 'Yandex', 'Telegram'].map((provider) => (
               <div key={provider} className="flex items-center justify-between p-4 border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors">
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-full bg-white border border-gray-200 flex items-center justify-center font-bold text-gray-600 text-lg">
                      {provider[0]}
                    </div>
                    <div>
                      <h3 className="font-medium text-gray-900">{provider}</h3>
                      <p className="text-xs text-gray-500">{t('oauth.client_id')}: ************392a</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs font-medium text-green-600 bg-green-50 px-2 py-1 rounded">Enabled</span>
                    <button className="text-sm text-blue-600 hover:text-blue-800 font-medium">{t('oauth.configure')}</button>
                  </div>
               </div>
             ))}
           </div>
           <div className="mt-4 text-center">
             <Link to="/oauth" className="text-sm text-blue-600 font-medium hover:underline inline-flex items-center gap-1">
               {t('oauth.manage_desc')} <ExternalLink size={14} />
             </Link>
           </div>
        </div>
      </section>

       {/* Email & SMS */}
       <section className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-100 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-50 text-green-600 rounded-lg">
               <Mail size={20} />
            </div>
            <h2 className="text-lg font-semibold text-gray-900">{t('settings.email_smtp')}</h2>
          </div>
          <Link 
            to="/settings/email-templates" 
            className="text-sm bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 px-3 py-1.5 rounded-md font-medium transition-colors"
          >
            {t('settings.manage_templates')}
          </Link>
        </div>
        <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">{t('settings.smtp_host')}</label>
              <input type="text" defaultValue="smtp.example.com" className="w-full border-gray-300 border rounded-lg p-2.5 focus:ring-blue-500 focus:border-blue-500" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">{t('settings.smtp_port')}</label>
              <input type="number" defaultValue={587} className="w-full border-gray-300 border rounded-lg p-2.5 focus:ring-blue-500 focus:border-blue-500" />
            </div>
             <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">{t('settings.from_addr')}</label>
              <input type="email" defaultValue="noreply@authgateway.com" className="w-full border-gray-300 border rounded-lg p-2.5 focus:ring-blue-500 focus:border-blue-500" />
            </div>
        </div>
        
        {/* SMS Settings Link */}
        <div className="bg-gray-50 p-6 border-t border-gray-100 flex items-center justify-between">
            <div className="flex items-center gap-3">
               <MessageSquare size={18} className="text-gray-500" />
               <span className="text-sm font-medium text-gray-700">{t('sms.title')}</span>
            </div>
            <Link 
              to="/settings/sms" 
              className="text-sm bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 px-4 py-2 rounded-md font-medium transition-colors flex items-center gap-2"
            >
              {t('oauth.configure')} <ChevronRight size={16} />
            </Link>
        </div>
      </section>
    </div>
  );
};

export default Settings;