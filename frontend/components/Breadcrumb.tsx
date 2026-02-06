import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { ChevronRight, Home } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const Breadcrumb: React.FC = () => {
  const { t } = useLanguage();
  const location = useLocation();
  const pathSegments = location.pathname.split('/').filter(Boolean);

  if (pathSegments.length === 0) return null;

  const ROUTE_LABELS: Record<string, string> = {
    'users': t('nav.users'),
    'groups': t('nav.groups'),
    'applications': t('nav.applications') || 'Applications',
    'sessions': t('nav.sessions'),
    'email': t('nav.email'),
    'templates': t('nav.email_templates') || 'Templates',
    'providers': t('nav.email_providers') || 'Providers',
    'sms': 'SMS',
    'settings': t('nav.settings'),
    'branding': t('settings.branding') || 'Branding',
    'access-control': t('nav.access_settings') || 'Access Control',
    'oauth': t('nav.oauth'),
    'oauth-clients': t('nav.oauth_clients'),
    'ip-security': t('nav.ip_security'),
    'audit-logs': t('nav.audit_logs'),
    'api-keys': t('nav.api_keys'),
    'roles': t('nav.roles'),
    'permissions': t('nav.permissions'),
    'developers': t('nav.developers'),
    'webhooks': t('nav.webhooks'),
    'service-accounts': t('nav.service_accounts'),
    'token-inspector': t('nav.token_inspector'),
    'ldap': 'LDAP',
    'saml': 'SAML',
    'bulk': t('nav.bulk_operations'),
    'new': t('common.create'),
    'edit': t('common.edit'),
    'create': t('common.create'),
    'update': t('common.edit'),
    'delete': t('common.delete'),
    'assign-roles': t('nav.roles'),
    'logs': t('nav.audit_logs'),
    'metadata': 'Metadata',
    'scim': 'SCIM',
  };

  const isUUID = (segment: string) => /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(segment);

  const breadcrumbs: { label: string; path: string }[] = [];
  let currentPath = '';

  for (const segment of pathSegments) {
    currentPath += `/${segment}`;
    if (isUUID(segment)) {
      breadcrumbs.push({ label: t('breadcrumb.details'), path: currentPath });
    } else {
      const label = ROUTE_LABELS[segment] || segment;
      breadcrumbs.push({ label, path: currentPath });
    }
  }

  return (
    <nav className="flex items-center gap-1 text-sm text-muted-foreground">
      <Link to="/" className="hover:text-foreground transition-colors flex items-center">
        <Home size={14} />
      </Link>
      {breadcrumbs.map((crumb, index) => (
        <React.Fragment key={crumb.path}>
          <ChevronRight size={14} className="text-muted-foreground/50" />
          {index === breadcrumbs.length - 1 ? (
            <span className="text-foreground font-medium truncate max-w-[200px]">{crumb.label}</span>
          ) : (
            <Link to={crumb.path} className="hover:text-foreground transition-colors truncate max-w-[150px]">
              {crumb.label}
            </Link>
          )}
        </React.Fragment>
      ))}
    </nav>
  );
};

export default Breadcrumb;
