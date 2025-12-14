import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useDashboardStats() {
  return useQuery({
    queryKey: queryKeys.dashboard.stats,
    queryFn: async () => {
      const stats = await apiClient.admin.users.getStats();
      return stats;
    },
    refetchInterval: 30000,
  });
}
