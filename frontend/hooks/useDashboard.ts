import { useQuery } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useDashboardStats() {
  return useQuery({
    queryKey: queryKeys.dashboard.stats,
    queryFn: async () => {
      // Fetch stats from the admin users service
      const stats = await apiClient.admin.users.getStats();

      // Transform backend response to match frontend DashboardStats type
      // Note: The actual API response structure may differ, adjust as needed
      return {
        totalUsers: stats.totalUsers || stats.total_users || 0,
        activeUsers: stats.activeUsers || stats.active_users || 0,
        usersWith2FA: stats.usersWith2FA || stats.users_2fa_enabled || stats.twoFactorEnabled || 0,
        totalApiKeys: stats.totalApiKeys || stats.total_api_keys || stats.activeApiKeys || 0,
        // These might need to be fetched from different endpoints
        recentRegistrations: stats.recentRegistrations || stats.registrationsByDay || [],
        loginActivity: stats.loginActivity || [],
      };
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  });
}
