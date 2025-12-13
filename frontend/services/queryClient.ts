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
      console.error(`[Mutation Error] ${friendlyError.title}: ${friendlyError.message}`);
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
    list: (page: number, pageSize: number, search?: string, role?: string) =>
      ['users', 'list', { page, pageSize, search, role }] as const,
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
  },

  // RBAC
  rbac: {
    roles: ['rbac', 'roles'] as const,
    role: (id: string) => ['rbac', 'role', id] as const,
    permissions: ['rbac', 'permissions'] as const,
    permission: (id: string) => ['rbac', 'permission', id] as const,
  },

  // Sessions
  sessions: {
    all: ['sessions'] as const,
    user: (userId: string) => ['sessions', 'user', userId] as const,
    current: ['sessions', 'current'] as const,
  },

  // Settings
  settings: {
    system: ['settings', 'system'] as const,
    branding: ['settings', 'branding'] as const,
    sms: ['settings', 'sms'] as const,
    passwordPolicy: ['settings', 'passwordPolicy'] as const,
  },

  // IP Filters
  ipFilters: {
    all: ['ipFilters'] as const,
    byType: (type: 'whitelist' | 'blacklist') => ['ipFilters', type] as const,
  },

  // Webhooks
  webhooks: {
    all: ['webhooks'] as const,
    detail: (id: string) => ['webhooks', 'detail', id] as const,
  },

  // Service Accounts
  serviceAccounts: {
    all: ['serviceAccounts'] as const,
    detail: (id: string) => ['serviceAccounts', 'detail', id] as const,
  },
};
