import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

interface AuditLogFilters {
  userId?: string;
  action?: string;
  resource?: string;
  status?: 'success' | 'failure';
  startDate?: Date;
  endDate?: Date;
}

export function useAuditLogs(
  page: number = 1,
  pageSize: number = 50,
  filters?: AuditLogFilters
) {
  return useQuery({
    queryKey: queryKeys.auditLogs.list({ page, pageSize, ...filters }),
    queryFn: async () => {
      const response = await apiClient.admin.audit.list({
        page,
        pageSize,
        ...filters,
      });
      return {
        logs: response.logs,
        items: response.logs,
        total: response.total,
        page: response.page,
        pageSize: response.page_size,
      };
    },
  });
}

export function useAuditLogDetail(id: string) {
  return useQuery({
    queryKey: ['auditLogs', 'detail', id],
    queryFn: async () => {
      const response = await apiClient.admin.audit.list({
        page: 1,
        pageSize: 1000,
      });
      return response.logs.find(log => log.id === id) || null;
    },
    enabled: !!id,
  });
}

export function useUserAuditLogs(userId: string, page: number = 1, pageSize: number = 50) {
  return useQuery({
    queryKey: ['auditLogs', 'user', userId, { page, pageSize }],
    queryFn: async () => {
      const response = await apiClient.admin.audit.list({
        userId,
        page,
        pageSize,
      });
      return {
        logs: response.logs,
        items: response.logs,
        total: response.total,
        page: response.page,
        pageSize: response.page_size,
      };
    },
    enabled: !!userId,
  });
}
