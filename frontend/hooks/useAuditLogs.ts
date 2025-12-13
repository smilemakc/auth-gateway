import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

interface AuditLogFilters {
  userId?: string;
  action?: string;
  resource?: string;
  startDate?: string;
  endDate?: string;
}

export function useAuditLogs(
  page: number = 1,
  pageSize: number = 50,
  filters?: AuditLogFilters
) {
  return useQuery({
    queryKey: queryKeys.auditLogs.list(page, pageSize, filters),
    queryFn: async () => {
      // Try different possible SDK paths for audit logs
      try {
        // Try: admin.auditLogs.list()
        if (apiClient.admin.auditLogs?.list) {
          const response = await apiClient.admin.auditLogs.list({ page, pageSize });
          return processAuditLogsResponse(response, filters);
        }

        // Try: admin.audit.list()
        if ((apiClient.admin as any).audit?.list) {
          const response = await (apiClient.admin as any).audit.list({ page, pageSize });
          return processAuditLogsResponse(response, filters);
        }

        // Try: auditLogs.list()
        if ((apiClient as any).auditLogs?.list) {
          const response = await (apiClient as any).auditLogs.list(page, pageSize);
          return processAuditLogsResponse(response, filters);
        }

        // If none of the above work, return empty data
        console.warn('[Audit Logs] SDK method not found, returning empty data');
        return { logs: [], items: [], total: 0 };
      } catch (error) {
        console.error('[Audit Logs] Error fetching:', error);
        throw error;
      }
    },
  });
}

// Helper function to process response and apply filters
function processAuditLogsResponse(response: any, filters?: AuditLogFilters) {
  // Client-side filtering if needed (ideally backend should handle this)
  let logs = response.logs || response.items || [];

  if (filters) {
    if (filters.userId) {
      logs = logs.filter((log: any) => log.userId === filters.userId);
    }
    if (filters.action) {
      logs = logs.filter((log: any) =>
        log.action?.toLowerCase().includes(filters.action!.toLowerCase())
      );
    }
    if (filters.resource) {
      logs = logs.filter((log: any) =>
        log.resource?.toLowerCase().includes(filters.resource!.toLowerCase())
      );
    }
  }

  return {
    ...response,
    logs,
    items: logs,
  };
}

export function useAuditLogDetail(id: string) {
  return useQuery({
    queryKey: queryKeys.auditLogs.detail(id),
    queryFn: async () => {
      // Try different SDK paths
      if (apiClient.admin.auditLogs?.get) {
        return await apiClient.admin.auditLogs.get(id);
      }
      if ((apiClient.admin as any).audit?.get) {
        return await (apiClient.admin as any).audit.get(id);
      }
      console.warn('[Audit Logs] SDK get method not found');
      return null;
    },
    enabled: !!id,
  });
}

export function useUserAuditLogs(userId: string, page: number = 1, pageSize: number = 50) {
  return useQuery({
    queryKey: queryKeys.auditLogs.byUser(userId, page, pageSize),
    queryFn: async () => {
      // Try to use a specific user audit logs endpoint if available
      try {
        if (apiClient.admin.auditLogs?.list) {
          const response = await apiClient.admin.auditLogs.list(page, pageSize);
          return processAuditLogsResponse(response, { userId });
        }
        if ((apiClient.admin as any).audit?.list) {
          const response = await (apiClient.admin as any).audit.list(page, pageSize);
          return processAuditLogsResponse(response, { userId });
        }
        return { logs: [], items: [], total: 0 };
      } catch (error) {
        console.error('Failed to fetch user audit logs:', error);
        throw error;
      }
    },
    enabled: !!userId,
  });
}
