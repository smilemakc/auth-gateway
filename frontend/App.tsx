import React from 'react';
import { HashRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { queryClient } from './services/queryClient';
import { AuthProvider, useAuth } from './services/authContext';
import { LanguageProvider } from './services/i18n';
import Layout from './components/Layout';
import Dashboard from './components/Dashboard';
import Users from './components/Users';
import UserDetails from './components/UserDetails';
import UserEdit from './components/UserEdit';
import Sessions from './components/Sessions';
import ApiKeys from './components/ApiKeys';
import OAuthProviders from './components/OAuthProviders';
import OAuthProviderEdit from './components/OAuthProviderEdit';
import AuditLogs from './components/AuditLogs';
import Settings from './components/Settings';
import EmailTemplates from './components/EmailTemplates';
import EmailTemplateEditor from './components/EmailTemplateEditor';
import Roles from './components/Roles';
import RoleEditor from './components/RoleEditor';
import Permissions from './components/Permissions';
import PermissionEdit from './components/PermissionEdit';
import IpSecurity from './components/IpSecurity';
import Branding from './components/Branding';
import Webhooks from './components/Webhooks';
import WebhookEdit from './components/WebhookEdit';
import ServiceAccounts from './components/ServiceAccounts';
import ServiceAccountEdit from './components/ServiceAccountEdit';
import SmsSettings from './components/SmsSettings';
import TokenInspector from './components/TokenInspector';
import Login from './components/Login';

const AppRoutes: React.FC = () => {
  const { isAuthenticated, isLoading, logout } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
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
          <Route path="/settings/email-templates" element={<EmailTemplates />} />
          <Route path="/settings/email-templates/:id" element={<EmailTemplateEditor />} />
          <Route path="/settings/roles" element={<Roles />} />
          <Route path="/settings/roles/:id" element={<RoleEditor />} />
          <Route path="/settings/permissions" element={<Permissions />} />
          <Route path="/settings/permissions/:id" element={<PermissionEdit />} />
          <Route path="/settings/security/ip-rules" element={<IpSecurity />} />
          <Route path="/settings/sms" element={<SmsSettings />} />
          <Route path="/developers/webhooks" element={<Webhooks />} />
          <Route path="/developers/webhooks/:id" element={<WebhookEdit />} />
          <Route path="/developers/service-accounts" element={<ServiceAccounts />} />
          <Route path="/developers/service-accounts/:id" element={<ServiceAccountEdit />} />
          <Route path="/developers/token-inspector" element={<TokenInspector />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout>
    </Router>
  );
};

const App: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <LanguageProvider>
        <AuthProvider>
          <AppRoutes />
        </AuthProvider>
      </LanguageProvider>
      {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
    </QueryClientProvider>
  );
};

export default App;