import { QueryClient, DefaultOptions } from '@tanstack/react-query';
import { getUserFriendlyError } from './errorHandler';

const queryConfig: DefaultOptions = {
  queries: {
    // Global defaults for queries
    retry: 1,
    refetchOnWindowFocus: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
  },
  mutations: {
    // Global defaults for mutations
    retry: 0,
    onError: (error) => {
      const friendlyError = getUserFriendlyError(error);
      if (import.meta.env.DEV) console.error(`[Mutation Error] ${friendlyError.title}: ${friendlyError.message}`);
      // Could integrate with a toast notification system here
    },
  },
};

export const queryClient = new QueryClient({
  defaultOptions: queryConfig,
});

// Query keys factory for consistency
export const queryKeys = {
  // Auth
  profile: ['profile'] as const,

  // Users
  users: {
    all: ['users'] as const,
    list: (page: number, pageSize: number) =>
      ['users', 'list', { page, pageSize }] as const,
    detail: (id: string) => ['users', 'detail', id] as const,
    stats: ['users', 'stats'] as const,
  },

  // API Keys
  apiKeys: {
    all: ['apiKeys'] as const,
    list: (page: number, pageSize: number) => ['apiKeys', 'list', { page, pageSize }] as const,
    detail: (id: string) => ['apiKeys', 'detail', id] as const,
  },

  // Audit Logs
  auditLogs: {
    all: ['auditLogs'] as const,
    list: (filters?: Record<string, unknown>) => ['auditLogs', 'list', filters] as const,
  },

  // Dashboard
  dashboard: {
    stats: ['dashboard', 'stats'] as const,
  },

  // OAuth
  oauth: {
    providers: ['oauth', 'providers'] as const,
    provider: (id: string) => ['oauth', 'provider', id] as const,
  },

  // Email Templates
  emailTemplates: {
    all: ['emailTemplates'] as const,
    detail: (id: string) => ['emailTemplates', 'detail', id] as const,
    types: ['emailTemplates', 'types'] as const,
    variables: (type: string) => ['emailTemplates', 'variables', type] as const,
    byType: (type: string) => ['emailTemplates', 'byType', type] as const,
  },

  // RBAC
  rbac: {
    roles: {
      all: ['rbac', 'roles'] as const,
      list: (page?: number, pageSize?: number) => ['rbac', 'roles', 'list', { page, pageSize }] as const,
      detail: (id: string) => ['rbac', 'roles', 'detail', id] as const,
    },
    permissions: {
      all: ['rbac', 'permissions'] as const,
      list: (page?: number, pageSize?: number) => ['rbac', 'permissions', 'list', { page, pageSize }] as const,
      detail: (id: string) => ['rbac', 'permissions', 'detail', id] as const,
    },
  },

  // Sessions
  sessions: {
    all: ['sessions'] as const,
    list: (page?: number, pageSize?: number) => ['sessions', 'list', { page, pageSize }] as const,
    detail: (id: string) => ['sessions', 'detail', id] as const,
    byUser: (userId: string, page?: number, pageSize?: number) => ['sessions', 'user', userId, { page, pageSize }] as const,
    current: ['sessions', 'current'] as const,
  },

  // Settings
  settings: {
    system: ['settings', 'system'] as const,
    branding: ['settings', 'branding'] as const,
    sms: ['settings', 'sms'] as const,
    passwordPolicy: ['settings', 'passwordPolicy'] as const,
    maintenance: ['settings', 'maintenance'] as const,
  },

  // IP Filters
  ipFilters: {
    all: ['ipFilters'] as const,
    list: (page?: number, pageSize?: number) => ['ipFilters', 'list', { page, pageSize }] as const,
    detail: (id: string) => ['ipFilters', 'detail', id] as const,
    byType: (type: 'whitelist' | 'blacklist') => ['ipFilters', type] as const,
  },

  // Webhooks
  webhooks: {
    all: ['webhooks'] as const,
    list: (page: number, pageSize: number) => ['webhooks', 'list', { page, pageSize }] as const,
    detail: (id: string) => ['webhooks', 'detail', id] as const,
  },

  // Service Accounts
  serviceAccounts: {
    all: ['serviceAccounts'] as const,
    detail: (id: string) => ['serviceAccounts', 'detail', id] as const,
  },

  // OAuth Clients (OIDC Provider)
  oauthClients: {
    all: ['oauthClients'] as const,
    list: (page: number, pageSize: number) => ['oauthClients', 'list', { page, pageSize }] as const,
    detail: (id: string) => ['oauthClients', 'detail', id] as const,
    scopes: ['oauthClients', 'scopes'] as const,
    consents: (clientId: string) => ['oauthClients', 'consents', clientId] as const,
  },

  // Groups
  groups: {
    all: ['groups'] as const,
    list: (page: number, pageSize: number) => ['groups', 'list', { page, pageSize }] as const,
    detail: (id: string) => ['groups', 'detail', id] as const,
    members: (id: string, page: number, pageSize: number) =>
      ['groups', 'members', id, { page, pageSize }] as const,
  },

  // LDAP
  ldap: {
    all: ['ldap'] as const,
    configs: ['ldap', 'configs'] as const,
    active: ['ldap', 'active'] as const,
    detail: (id: string) => ['ldap', 'detail', id] as const,
    syncLogs: (id: string) => ['ldap', 'syncLogs', id] as const,
  },

  // SAML
  saml: {
    all: ['saml'] as const,
    list: (page: number, pageSize: number) => ['saml', 'list', { page, pageSize }] as const,
    detail: (id: string) => ['saml', 'detail', id] as const,
    metadata: ['saml', 'metadata'] as const,
  },

  // SCIM
  scim: {
    config: ['scim', 'config'] as const,
    metadata: ['scim', 'metadata'] as const,
  },

  // Applications (Multi-tenant)
  applications: {
    all: ['applications'] as const,
    list: (page: number, pageSize: number) => ['applications', 'list', { page, pageSize }] as const,
    detail: (id: string) => ['applications', 'detail', id] as const,
    branding: (id: string) => ['applications', 'branding', id] as const,
    users: (id: string, page: number, pageSize: number) => ['applications', 'users', id, { page, pageSize }] as const,
    templates: (id: string) => ['applications', 'templates', id] as const,
    templateDetail: (appId: string, templateId: string) => ['applications', 'templates', appId, templateId] as const,
    oauthProviders: (appId: string) => ['applications', 'oauthProviders', appId] as const,
    oauthProvider: (appId: string, providerId: string) => ['applications', 'oauthProviders', appId, providerId] as const,
    telegramBots: (appId: string) => ['applications', 'telegramBots', appId] as const,
    telegramBot: (appId: string, botId: string) => ['applications', 'telegramBots', appId, botId] as const,
  },

  // User Telegram
  userTelegram: {
    accounts: (userId: string) => ['userTelegram', 'accounts', userId] as const,
    botAccess: (userId: string, appId?: string) => ['userTelegram', 'botAccess', userId, appId] as const,
  },
};
