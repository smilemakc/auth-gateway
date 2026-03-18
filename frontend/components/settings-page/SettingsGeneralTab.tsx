import React from 'react';
import { Link } from 'react-router-dom';
import { Server, Palette, Shield, ShieldAlert, ChevronRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface SystemStatus {
  status: string;
  services?: {
    database?: string;
    redis?: string;
  };
}

interface MaintenanceStatus {
  enabled: boolean;
}

interface SettingsGeneralTabProps {
  systemStatus: SystemStatus | undefined;
  maintenanceStatus: MaintenanceStatus | undefined;
  onToggleMaintenance: () => void;
}

export const SettingsGeneralTab: React.FC<SettingsGeneralTabProps> = ({
  systemStatus,
  maintenanceStatus,
  onToggleMaintenance,
}) => {
  const { t } = useLanguage();

  return (
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
                  onClick={onToggleMaintenance}
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
  );
};
