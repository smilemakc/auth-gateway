import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type {
  CreateLDAPConfigRequest,
  UpdateLDAPConfigRequest,
  LDAPTestConnectionRequest,
  LDAPSyncRequest,
} from '@auth-gateway/client-sdk';

export function useLDAPConfigs() {
  return useQuery({
    queryKey: queryKeys.ldap.configs,
    queryFn: () => apiClient.admin.ldap.listConfigs(),
  });
}

export function useLDAPConfig(configId: string) {
  return useQuery({
    queryKey: queryKeys.ldap.detail(configId),
    queryFn: () => apiClient.admin.ldap.getConfig(configId),
    enabled: !!configId,
  });
}

export function useActiveLDAPConfig() {
  return useQuery({
    queryKey: queryKeys.ldap.active,
    queryFn: () => apiClient.admin.ldap.getActiveConfig(),
  });
}

export function useCreateLDAPConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateLDAPConfigRequest) => apiClient.admin.ldap.createConfig(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ldap.all });
    },
  });
}

export function useUpdateLDAPConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateLDAPConfigRequest }) =>
      apiClient.admin.ldap.updateConfig(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ldap.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.ldap.detail(variables.id) });
    },
  });
}

export function useDeleteLDAPConfig() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (configId: string) => apiClient.admin.ldap.deleteConfig(configId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ldap.all });
    },
  });
}

export function useTestLDAPConnection() {
  return useMutation({
    mutationFn: (data: LDAPTestConnectionRequest) => apiClient.admin.ldap.testConnection(data),
  });
}

export function useSyncLDAP() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data?: LDAPSyncRequest }) =>
      apiClient.admin.ldap.sync(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.ldap.detail(variables.id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.ldap.syncLogs(variables.id) });
    },
  });
}

export function useLDAPSyncLogs(configId: string) {
  return useQuery({
    queryKey: queryKeys.ldap.syncLogs(configId),
    queryFn: () => apiClient.admin.ldap.getSyncLogs(configId),
    enabled: !!configId,
  });
}

