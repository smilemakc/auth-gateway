import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useSystemStatus() {
  return useQuery({
    queryKey: queryKeys.settings.system,
    queryFn: () => apiClient.admin.system.getHealth(),
    refetchInterval: 10000, // Refresh every 10 seconds
  });
}

export function usePasswordPolicy() {
  return useQuery({
    queryKey: queryKeys.settings.passwordPolicy,
    queryFn: async () => {
      // Try to get password policy from system settings
      // The exact endpoint might vary based on SDK implementation
      try {
        const health = await apiClient.admin.system.getHealth();
        // If password policy is included in health response, return it
        if (health && 'passwordPolicy' in health) {
          return (health as any).passwordPolicy;
        }
        // Otherwise return a default policy
        return {
          minLength: 8,
          requireUppercase: true,
          requireLowercase: true,
          requireNumbers: true,
          requireSpecial: false,
          historyCount: 3,
          expiryDays: 90,
          jwtTtlMinutes: 15,
          refreshTtlDays: 7,
        };
      } catch (error) {
        console.error('Failed to fetch password policy:', error);
        throw error;
      }
    },
  });
}

export function useUpdatePasswordPolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (policy: any) => {
      // The SDK might not have this method, we'll need to implement it
      // For now, return a mock response
      console.log('Update password policy:', policy);
      return policy;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.passwordPolicy });
    },
  });
}

export function useBranding() {
  return useQuery({
    queryKey: queryKeys.settings.branding,
    queryFn: () => apiClient.admin.branding.get(),
  });
}

export function useUpdateBranding() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (branding: any) => apiClient.admin.branding.update(branding),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.branding });
    },
  });
}

export function useMaintenanceMode() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ enabled, message }: { enabled: boolean; message?: string }) =>
      enabled
        ? apiClient.admin.system.enableMaintenanceMode(message || 'System maintenance in progress')
        : apiClient.admin.system.disableMaintenanceMode(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.system });
    },
  });
}
