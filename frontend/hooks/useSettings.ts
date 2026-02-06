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
      // Password policy is currently config-based (environment variables)
      // A backend endpoint to update these dynamically needs to be implemented
      // For now, attempt to call the system settings endpoint
      const response = await fetch('/api/admin/system/password-policy', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
        body: JSON.stringify(policy),
      });
      if (!response.ok) {
        // If the endpoint doesn't exist (404), provide a clear message
        if (response.status === 404) {
          throw new Error('Password policy update endpoint not implemented. Password policy is currently configured via environment variables.');
        }
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `Failed to update password policy: ${response.status}`);
      }
      return response.json();
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

export function useMaintenanceModeStatus() {
  return useQuery({
    queryKey: queryKeys.settings.maintenance,
    queryFn: () => apiClient.health.isMaintenanceMode(),
    refetchInterval: 10000, // Refresh every 10 seconds
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
      queryClient.invalidateQueries({ queryKey: queryKeys.settings.maintenance });
    },
  });
}

// SMS Settings
export function useSmsSettings() {
  return useQuery({
    queryKey: ['smsSettings'],
    queryFn: () => apiClient.admin.smsSettings.list(),
  });
}

export function useSmsSettingsActive() {
  return useQuery({
    queryKey: ['smsSettings', 'active'],
    queryFn: () => apiClient.admin.smsSettings.getActive(),
  });
}

export function useCreateSmsSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: any) => apiClient.admin.smsSettings.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['smsSettings'] });
    },
  });
}

export function useUpdateSmsSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      apiClient.admin.smsSettings.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['smsSettings'] });
    },
  });
}

export function useActivateSmsSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.admin.smsSettings.activate(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['smsSettings'] });
    },
  });
}
