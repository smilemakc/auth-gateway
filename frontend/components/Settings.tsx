
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Save, Lock, Globe, Shield, ExternalLink, ChevronRight, ShieldAlert, Palette, Key, Activity, Server, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useSystemStatus, usePasswordPolicy, useUpdatePasswordPolicy, useMaintenanceMode, useMaintenanceModeStatus } from '../hooks/useSettings';
import { PasswordPolicy } from '../types';
import { toast } from '../services/toast';
import { confirm } from '../services/confirm';
import { logger } from '@/lib/logger';

const Settings: React.FC = () => {
  const { t } = useLanguage();
  const [activeTab, setActiveTab] = useState<'general' | 'security'>('general');

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
      const ok = await confirm({ description: newStatus ? t('sys.confirm_enable') : t('sys.confirm_disable') });
      if (ok) {
        try {
          await maintenanceMutation.mutateAsync({
            enabled: newStatus,
            message: 'System maintenance in progress',
          });
        } catch (error) {
          logger.error('Failed to toggle maintenance mode:', error);
          toast.error('Failed to update maintenance mode');
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
        toast.success(t('common.saved'));
      } catch (error) {
        logger.error('Failed to save password policy:', error);
        toast.error('Failed to save settings');
      }
    }
  };

  if (statusLoading || policyLoading || maintenanceLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (statusError || policyError) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive">
          Error loading settings: {((statusError || policyError) as Error).message}
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">{t('settings.title')}</h1>
        {activeTab === 'security' && (
          <button
            onClick={handleSave}
            disabled={updatePolicyMutation.isPending}
            className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-6 py-2 rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Save size={18} />
            {updatePolicyMutation.isPending ? t('settings.saving') : t('common.save')}
          </button>
        )}
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-border">
        <button
          onClick={() => setActiveTab('general')}
          className={`px-4 py-2.5 text-sm font-medium transition-colors relative ${
            activeTab === 'general'
              ? 'text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          {t('settings.tab_general')}
          {activeTab === 'general' && (
            <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t" />
          )}
        </button>
        <button
          onClick={() => setActiveTab('security')}
          className={`px-4 py-2.5 text-sm font-medium transition-colors relative ${
            activeTab === 'security'
              ? 'text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          {t('settings.tab_security')}
          {activeTab === 'security' && (
            <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-t" />
          )}
        </button>
      </div>

      {activeTab === 'general' && (
        <>
          {/* System Status & Maintenance */}
          <section className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="p-6 border-b border-border flex items-center justify-between">
              <div className="flex items-center gap-3">
                 <div className="p-2 bg-muted text-muted-foreground rounded-lg">
                    <Server size={20} />
                 </div>
                 <div>
                    <h2 className="text-lg font-semibold text-foreground">{t('sys.health')}</h2>
                    <div className="flex items-center gap-2 text-sm text-muted-foreground mt-1">
                       <span className={`inline-block w-2 h-2 rounded-full ${systemStatus?.status === 'healthy' ? 'bg-success' : 'bg-destructive'}`}></span>
                       {t('settings.status_label')}: <span className="uppercase font-medium">{systemStatus?.status}</span>
                       <span className="mx-1">•</span>
                       <span>{t('settings.db_label')}: {systemStatus?.services?.database || 'unknown'}</span>
                       <span className="mx-1">•</span>
                       <span>{t('settings.redis_label')}: {systemStatus?.services?.redis || 'unknown'}</span>
                    </div>
                 </div>
              </div>
              <div className="flex items-center gap-3">
                 <div className="flex items-center">
                    <span className={`mr-2 text-sm font-medium ${maintenanceStatus?.enabled ? 'text-warning' : 'text-muted-foreground'}`}>
                      {maintenanceStatus?.enabled ? t('sys.maintenance_on') : t('sys.maintenance_off')}
                    </span>
                    <button
                      onClick={toggleMaintenance}
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 ${
                        maintenanceStatus?.enabled ? 'bg-warning' : 'bg-muted'
                      }`}
                    >
                      <span
                        className={`${
                          maintenanceStatus?.enabled ? 'translate-x-6' : 'translate-x-1'
                        } inline-block h-4 w-4 transform rounded-full bg-card transition-transform`}
                      />
                    </button>
                 </div>
              </div>
            </div>
          </section>

          {/* Branding Section */}
          <section className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="p-6 border-b border-border flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-pink-50 text-pink-600 rounded-lg">
                   <Palette size={20} />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-foreground">{t('settings.branding')}</h2>
                  <p className="text-sm text-muted-foreground">{t('settings.branding_desc')}</p>
                </div>
              </div>
              <Link
                to="/settings/branding"
                className="text-sm bg-card border border-input text-foreground hover:bg-accent px-4 py-2 rounded-md font-medium transition-colors flex items-center gap-2"
              >
                {t('oauth.configure')} <ChevronRight size={16} />
              </Link>
            </div>
          </section>

          {/* Access Control Section */}
          <section className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="p-6 border-b border-border flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-indigo-50 text-indigo-600 rounded-lg">
                   <Shield size={20} />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-foreground">{t('settings.roles_desc')}</h2>
                  <p className="text-sm text-muted-foreground">{t('settings.roles_manage_desc')}</p>
                </div>
              </div>
              <Link
                to="/settings/access-control"
                className="text-sm bg-card border border-input text-foreground hover:bg-accent px-4 py-2 rounded-md font-medium transition-colors flex items-center gap-2"
              >
                {t('oauth.configure')} <ChevronRight size={16} />
              </Link>
            </div>
          </section>

          {/* IP Security Section */}
          <section className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="p-6 border-b border-border flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2 bg-destructive/10 text-destructive rounded-lg">
                   <ShieldAlert size={20} />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-foreground">{t('settings.ip_security')}</h2>
                  <p className="text-sm text-muted-foreground">{t('settings.ip_desc')}</p>
                </div>
              </div>
              <Link
                to="/ip-security"
                className="text-sm bg-card border border-input text-foreground hover:bg-accent px-4 py-2 rounded-md font-medium transition-colors flex items-center gap-2"
              >
                {t('oauth.configure')} <ChevronRight size={16} />
              </Link>
            </div>
          </section>
        </>
      )}

      {activeTab === 'security' && (
        <>
          {/* Security Policies */}
          <section className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="p-6 border-b border-border flex items-center gap-3">
              <div className="p-2 bg-primary/10 text-primary rounded-lg">
                <Lock size={20} />
              </div>
              <h2 className="text-lg font-semibold text-foreground">{t('settings.security_policies')}</h2>
            </div>
            {localPasswordPolicy && (
              <div className="p-6 space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">{t('settings.jwt_ttl')}</label>
                    <input
                      type="number"
                      name="jwtTtlMinutes"
                      value={localPasswordPolicy.jwtTtlMinutes}
                      onChange={handlePolicyChange}
                      className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-2">{t('settings.refresh_ttl')}</label>
                    <input
                      type="number"
                      name="refreshTtlDays"
                      value={localPasswordPolicy.refreshTtlDays}
                      onChange={handlePolicyChange}
                      className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
                    />
                  </div>
                </div>

                <div className="border-t border-border pt-6">
                  <h3 className="text-md font-medium text-foreground mb-4">{t('settings.password_policy')}</h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                    <div>
                      <label className="block text-sm font-medium text-foreground mb-2">{t('settings.min_pass')}</label>
                      <input
                        type="number"
                        name="minLength"
                        value={localPasswordPolicy.minLength}
                        onChange={handlePolicyChange}
                        className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-foreground mb-2">{t('settings.pass_history')}</label>
                      <input
                        type="number"
                        name="historyCount"
                        value={localPasswordPolicy.historyCount}
                        onChange={handlePolicyChange}
                        className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
                      />
                    </div>
                     <div>
                      <label className="block text-sm font-medium text-foreground mb-2">{t('settings.pass_expiry')}</label>
                      <input
                        type="number"
                        name="expiryDays"
                        value={localPasswordPolicy.expiryDays}
                        onChange={handlePolicyChange}
                        className="w-full border-input border rounded-lg p-2.5 focus:ring-ring focus:border-ring"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <label className="flex items-center gap-3 cursor-pointer"
                      onClick={() => setLocalPasswordPolicy(prev => ({ ...prev, requireUppercase: !prev.requireUppercase }))}>
                      <span className={`transition-colors ${localPasswordPolicy.requireUppercase ? 'text-success' : 'text-muted-foreground'}`}>
                        {localPasswordPolicy.requireUppercase ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                      </span>
                      <span className="text-sm text-foreground">{t('settings.req_uppercase')}</span>
                    </label>
                    <label className="flex items-center gap-3 cursor-pointer"
                      onClick={() => setLocalPasswordPolicy(prev => ({ ...prev, requireLowercase: !prev.requireLowercase }))}>
                      <span className={`transition-colors ${localPasswordPolicy.requireLowercase ? 'text-success' : 'text-muted-foreground'}`}>
                        {localPasswordPolicy.requireLowercase ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                      </span>
                      <span className="text-sm text-foreground">{t('settings.req_lowercase')}</span>
                    </label>
                    <label className="flex items-center gap-3 cursor-pointer"
                      onClick={() => setLocalPasswordPolicy(prev => ({ ...prev, requireNumbers: !prev.requireNumbers }))}>
                      <span className={`transition-colors ${localPasswordPolicy.requireNumbers ? 'text-success' : 'text-muted-foreground'}`}>
                        {localPasswordPolicy.requireNumbers ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                      </span>
                      <span className="text-sm text-foreground">{t('settings.req_numbers')}</span>
                    </label>
                    <label className="flex items-center gap-3 cursor-pointer"
                      onClick={() => setLocalPasswordPolicy(prev => ({ ...prev, requireSpecial: !prev.requireSpecial }))}>
                      <span className={`transition-colors ${localPasswordPolicy.requireSpecial ? 'text-success' : 'text-muted-foreground'}`}>
                        {localPasswordPolicy.requireSpecial ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
                      </span>
                      <span className="text-sm text-foreground">{t('settings.req_special')}</span>
                    </label>
                  </div>
                </div>
              </div>
            )}
          </section>
        </>
      )}

    </div>
  );
};

export default Settings;