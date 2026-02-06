import React from 'react';
import { HashRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { queryClient } from './services/queryClient';
import { AuthProvider, useAuth } from './services/authContext';
import { ApplicationProvider } from './services/appContext';
import { LanguageProvider } from './services/i18n';
import { ThemeProvider } from './lib/theme';
import Layout from './components/Layout';
import Dashboard from './components/Dashboard';
import Users from './components/Users';
import UserDetails from './components/UserDetails';
import UserEdit from './components/UserEdit';
import Sessions from './components/Sessions';
import ApiKeys from './components/ApiKeys';
import OAuthProviders from './components/OAuthProviders';
import OAuthProviderEdit from './components/OAuthProviderEdit';
import OAuthClients from './components/OAuthClients';
import OAuthClientEdit from './components/OAuthClientEdit';
import AuditLogs from './components/AuditLogs';
import Settings from './components/Settings';
import EmailTemplates from './components/EmailTemplates';
import EmailTemplateEditor from './components/EmailTemplateEditor';
import Roles from './components/Roles';
import RoleEditor from './components/RoleEditor';
import Permissions from './components/Permissions';
import PermissionEdit from './components/PermissionEdit';
import AccessControl from './components/AccessControl';
import IpSecurity from './components/IpSecurity';
import Branding from './components/Branding';
import Webhooks from './components/Webhooks';
import WebhookEdit from './components/WebhookEdit';
import ServiceAccounts from './components/ServiceAccounts';
import ServiceAccountEdit from './components/ServiceAccountEdit';
import SmsSettings from './components/SmsSettings';
import TokenInspector from './components/TokenInspector';
import Login from './components/Login';
import Groups from './components/Groups';
import GroupEdit from './components/GroupEdit';
import GroupDetails from './components/GroupDetails';
import LDAPConfigs from './components/LDAPConfigs';
import LDAPConfigEdit from './components/LDAPConfigEdit';
import LDAPSyncLogs from './components/LDAPSyncLogs';
import SAMLSPs from './components/SAMLSPs';
import SAMLSPEdit from './components/SAMLSPEdit';
import SAMLMetadata from './components/SAMLMetadata';
import BulkOperations from './components/BulkOperations';
import BulkCreateUsers from './components/BulkCreateUsers';
import BulkUpdateUsers from './components/BulkUpdateUsers';
import BulkDeleteUsers from './components/BulkDeleteUsers';
import BulkAssignRoles from './components/BulkAssignRoles';
import SCIMSettings from './components/SCIMSettings';
import Applications from './components/Applications';
import ApplicationEdit from './components/ApplicationEdit';
import ApplicationDetails from './components/ApplicationDetails';
import ApplicationBrandingTab from './components/ApplicationBrandingTab';
import ApplicationUsersTab from './components/ApplicationUsersTab';
import ApplicationTemplateEditor from './components/ApplicationTemplateEditor';
import EmailProviders from './components/EmailProviders';
import EmailProviderEdit from './components/EmailProviderEdit';
import EmailSettings from './components/EmailSettings';
import ApplicationOAuthProviderEdit from './components/ApplicationOAuthProviderEdit';
import TelegramBotEdit from './components/TelegramBotEdit';

const AppRoutes: React.FC = () => {
  const { isAuthenticated, isLoading, logout } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Login />;
  }

  return (
    <Router>
      <Layout onLogout={logout}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/users" element={<Users />} />
          <Route path="/users/new" element={<UserEdit />} />
          <Route path="/users/:id" element={<UserDetails />} />
          <Route path="/users/:id/edit" element={<UserEdit />} />
          <Route path="/sessions" element={<Sessions />} />
          <Route path="/api-keys" element={<ApiKeys />} />
          <Route path="/oauth" element={<OAuthProviders />} />
          <Route path="/oauth/new" element={<OAuthProviderEdit />} />
          <Route path="/oauth/:id" element={<OAuthProviderEdit />} />
          <Route path="/oauth-clients" element={<OAuthClients />} />
          <Route path="/oauth-clients/new" element={<OAuthClientEdit />} />
          <Route path="/oauth-clients/:id" element={<OAuthClientEdit />} />
          <Route path="/roles" element={<Roles />} />
          <Route path="/roles/new" element={<RoleEditor />} />
          <Route path="/roles/:id" element={<RoleEditor />} />
          <Route path="/permissions" element={<Permissions />} />
          <Route path="/permissions/new" element={<PermissionEdit />} />
          <Route path="/permissions/:id" element={<PermissionEdit />} />
          <Route path="/ip-security" element={<IpSecurity />} />
          <Route path="/audit-logs" element={<AuditLogs />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/settings/branding" element={<Branding />} />
          <Route path="/settings/email" element={<EmailSettings />} />
          <Route path="/settings/email-templates" element={<EmailTemplates />} />
          <Route path="/settings/email-templates/:id" element={<EmailTemplateEditor />} />
          <Route path="/settings/roles" element={<Roles />} />
          <Route path="/settings/roles/:id" element={<RoleEditor />} />
          <Route path="/settings/permissions" element={<Permissions />} />
          <Route path="/settings/permissions/:id" element={<PermissionEdit />} />
          <Route path="/settings/access-control" element={<AccessControl />} />
          <Route path="/settings/access-control/roles/new" element={<RoleEditor />} />
          <Route path="/settings/access-control/roles/:id" element={<RoleEditor />} />
          <Route path="/settings/access-control/permissions/new" element={<PermissionEdit />} />
          <Route path="/settings/access-control/permissions/:id" element={<PermissionEdit />} />
          <Route path="/settings/security/ip-rules" element={<IpSecurity />} />
          <Route path="/settings/sms" element={<SmsSettings />} />
          <Route path="/settings/email-providers" element={<EmailProviders />} />
          <Route path="/settings/email-providers/:id" element={<EmailProviderEdit />} />
          <Route path="/developers/webhooks" element={<Webhooks />} />
          <Route path="/developers/webhooks/:id" element={<WebhookEdit />} />
          <Route path="/developers/service-accounts" element={<ServiceAccounts />} />
          <Route path="/developers/service-accounts/:id" element={<ServiceAccountEdit />} />
          <Route path="/developers/token-inspector" element={<TokenInspector />} />
          <Route path="/groups" element={<Groups />} />
          <Route path="/groups/new" element={<GroupEdit />} />
          <Route path="/groups/:id" element={<GroupDetails />} />
          <Route path="/groups/:id/edit" element={<GroupEdit />} />
          <Route path="/ldap" element={<LDAPConfigs />} />
          <Route path="/ldap/new" element={<LDAPConfigEdit />} />
          <Route path="/ldap/:id" element={<LDAPConfigEdit />} />
          <Route path="/ldap/:id/logs" element={<LDAPSyncLogs />} />
          <Route path="/saml" element={<SAMLSPs />} />
          <Route path="/saml/new" element={<SAMLSPEdit />} />
          <Route path="/saml/:id" element={<SAMLSPEdit />} />
          <Route path="/saml/metadata" element={<SAMLMetadata />} />
          <Route path="/bulk" element={<BulkOperations />} />
          <Route path="/bulk/create" element={<BulkCreateUsers />} />
          <Route path="/bulk/update" element={<BulkUpdateUsers />} />
          <Route path="/bulk/delete" element={<BulkDeleteUsers />} />
          <Route path="/bulk/assign-roles" element={<BulkAssignRoles />} />
          <Route path="/settings/scim" element={<SCIMSettings />} />
          <Route path="/applications" element={<Applications />} />
          <Route path="/applications/new" element={<ApplicationEdit />} />
          <Route path="/applications/:id" element={<ApplicationDetails />} />
          <Route path="/applications/:id/edit" element={<ApplicationEdit />} />
          <Route path="/applications/:appId/email-templates/:templateId" element={<ApplicationTemplateEditor />} />
          <Route path="/applications/:applicationId/oauth/new" element={<ApplicationOAuthProviderEdit />} />
          <Route path="/applications/:applicationId/oauth/:providerId" element={<ApplicationOAuthProviderEdit />} />
          <Route path="/applications/:applicationId/telegram-bots/new" element={<TelegramBotEdit />} />
          <Route path="/applications/:applicationId/telegram-bots/:botId" element={<TelegramBotEdit />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout>
    </Router>
  );
};

const App: React.FC = () => {
  return (
    <ThemeProvider defaultMode="system">
      <QueryClientProvider client={queryClient}>
        <LanguageProvider>
          <AuthProvider>
            <ApplicationProvider>
              <AppRoutes />
            </ApplicationProvider>
          </AuthProvider>
        </LanguageProvider>
        {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
      </QueryClientProvider>
    </ThemeProvider>
  );
};

export default App;