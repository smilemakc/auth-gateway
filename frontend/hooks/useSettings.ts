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
    queryFn: () => apiClient.admin.system.getPasswordPolicy(),
  });
}

export function useUpdatePasswordPolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (policy: any) => apiClient.admin.system.updatePasswordPolicy(policy),
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
    queryFn: async () => {
      try {
        return await apiClient.admin.smsSettings.getActive();
      } catch (err: any) {
        if (err?.status === 404 || err?.response?.status === 404) {
          return null;
        }
        throw err;
      }
    },
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
