
import React, { useState, useEffect } from 'react';
import { Save } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { LoadingSpinner } from '../ui';
import { useSystemStatus, usePasswordPolicy, useUpdatePasswordPolicy, useMaintenanceMode, useMaintenanceModeStatus } from '../../hooks/useSettings';
import { PasswordPolicy } from '../../types';
import { toast } from '../../services/toast';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';
import { SettingsGeneralTab } from './SettingsGeneralTab';
import { SettingsSecurityTab } from './SettingsSecurityTab';

const Settings: React.FC = () => {
  const { t } = useLanguage();
  const [activeTab, setActiveTab] = useState<'general' | 'security'>('general');

  const { data: systemStatus, isLoading: statusLoading, error: statusError } = useSystemStatus();
  const { data: maintenanceStatus, isLoading: maintenanceLoading } = useMaintenanceModeStatus();
  const { data: apiPasswordPolicy, isLoading: policyLoading, error: policyError } = usePasswordPolicy();
  const updatePolicyMutation = useUpdatePasswordPolicy();
  const maintenanceMutation = useMaintenanceMode();

  const [localPasswordPolicy, setLocalPasswordPolicy] = useState<PasswordPolicy | null>(null);

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

  const handleTogglePolicy = (field: keyof PasswordPolicy) => {
    setLocalPasswordPolicy(prev => {
      if (!prev) return prev;
      return { ...prev, [field]: !prev[field] };
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
    return <LoadingSpinner className="min-h-screen" />;
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
        <SettingsGeneralTab
          systemStatus={systemStatus}
          maintenanceStatus={maintenanceStatus}
          onToggleMaintenance={toggleMaintenance}
        />
      )}

      {activeTab === 'security' && (
        <SettingsSecurityTab
          localPasswordPolicy={localPasswordPolicy}
          onPolicyChange={handlePolicyChange}
          onTogglePolicy={handleTogglePolicy}
        />
      )}

    </div>
  );
};

export default Settings;
