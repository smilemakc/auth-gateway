import React from 'react';
import { Link } from 'react-router-dom';
import { ExternalLink, Info, CheckCircle } from 'lucide-react';
import { useSCIMConfig, useSCIMMetadata } from '../hooks/useSCIM';
import { useLanguage } from '../services/i18n';

const SCIMSettings: React.FC = () => {
  const { t } = useLanguage();
  const { data: config, isLoading: isLoadingConfig } = useSCIMConfig();
  const { data: metadata, isLoading: isLoadingMetadata } = useSCIMMetadata();

  if (isLoadingConfig || isLoadingMetadata) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">{t('scim.title')}</h1>
        <p className="text-muted-foreground mt-1">{t('scim.desc')}</p>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6 space-y-6">
        {/* Status */}
        <div className="flex items-center gap-3 p-4 bg-primary/10 border border-primary/20 rounded-lg">
          <CheckCircle className="text-primary" size={24} />
          <div>
            <h3 className="font-semibold text-primary">{t('scim.enabled')}</h3>
            <p className="text-sm text-primary">{t('scim.endpoints_available')}</p>
          </div>
        </div>

        {/* Endpoints */}
        <div>
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('scim.endpoints')}</h2>
          <div className="space-y-3">
            <div className="border border-border rounded-lg p-4">
              <div className="text-sm font-medium text-foreground mb-1">{t('scim.base_url')}</div>
              <code className="text-sm text-foreground bg-muted px-2 py-1 rounded">
                {config?.base_url || window.location.origin}/scim/v2
              </code>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="border border-border rounded-lg p-4">
                <div className="text-sm font-medium text-foreground mb-2">{t('scim.users_endpoint')}</div>
                <code className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded block">
                  GET/POST /scim/v2/Users
                </code>
                <code className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded block mt-1">
                  GET/PUT/PATCH/DELETE /scim/v2/Users/{'{id}'}
                </code>
              </div>

              <div className="border border-border rounded-lg p-4">
                <div className="text-sm font-medium text-foreground mb-2">{t('scim.groups_endpoint')}</div>
                <code className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded block">
                  GET/POST /scim/v2/Groups
                </code>
                <code className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded block mt-1">
                  GET/PUT/PATCH/DELETE /scim/v2/Groups/{'{id}'}
                </code>
              </div>
            </div>

            <div className="border border-border rounded-lg p-4">
              <div className="text-sm font-medium text-foreground mb-2">{t('scim.sp_config')}</div>
              <code className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded block">
                GET /scim/v2/ServiceProviderConfig
              </code>
            </div>
          </div>
        </div>

        {/* Supported Operations */}
        {metadata && (
          <div>
            <h2 className="text-lg font-semibold text-foreground mb-4">{t('scim.supported_ops')}</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="border border-border rounded-lg p-4">
                <div className="text-sm font-medium text-foreground mb-2">{t('scim.users')}</div>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>• {t('scim.op_create')}</li>
                  <li>• {t('scim.op_read')}</li>
                  <li>• {t('scim.op_update')}</li>
                  <li>• {t('scim.op_delete')}</li>
                  <li>• {t('scim.op_list')}</li>
                  <li>• {t('scim.op_filter')}</li>
                </ul>
              </div>

              <div className="border border-border rounded-lg p-4">
                <div className="text-sm font-medium text-foreground mb-2">{t('scim.groups')}</div>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>• {t('scim.op_create')}</li>
                  <li>• {t('scim.op_read')}</li>
                  <li>• {t('scim.op_update')}</li>
                  <li>• {t('scim.op_delete')}</li>
                  <li>• {t('scim.op_list')}</li>
                  <li>• {t('scim.op_members')}</li>
                </ul>
              </div>
            </div>
          </div>
        )}

        {/* Documentation */}
        <div className="bg-muted border border-border rounded-lg p-4">
          <div className="flex items-start gap-3">
            <Info className="text-primary mt-0.5" size={20} />
            <div className="flex-1">
              <h3 className="font-semibold text-foreground mb-1">{t('scim.integration_guide')}</h3>
              <p className="text-sm text-muted-foreground mb-3">
                {t('scim.integration_desc')}
              </p>
              <Link
                to="/docs/scim-integration"
                target="_blank"
                className="inline-flex items-center gap-2 text-sm text-primary hover:text-primary/80"
              >
                {t('scim.view_guide')}
                <ExternalLink size={16} />
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SCIMSettings;

