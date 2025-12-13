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
    detail: (id: string) => ['webhooks', 'detail', id] as const,
  },

  // Service Accounts
  serviceAccounts: {
    all: ['serviceAccounts'] as const,
    detail: (id: string) => ['serviceAccounts', 'detail', id] as const,
  },
};
