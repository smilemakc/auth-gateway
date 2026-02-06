import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type { CreateAPIKeyRequest, UpdateAPIKeyRequest } from '@auth-gateway/client-sdk';
import { useCurrentAppId } from './useAppAwareQuery';

export function useApiKeys(page: number = 1, pageSize: number = 50) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.apiKeys.list(page, pageSize), appId],
    queryFn: async () => {
      const response = await apiClient.admin.apiKeys.list(page, pageSize);
      return response;
    },
  });
}

export function useApiKeyDetail(id: string) {
  return useQuery({
    queryKey: queryKeys.apiKeys.detail(id),
    queryFn: () => apiClient.apiKeys.get(id),
    enabled: !!id,
  });
}

export function useCreateApiKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateAPIKeyRequest) =>
      apiClient.apiKeys.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.all });
    },
  });
}

export function useUpdateApiKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateAPIKeyRequest }) =>
      apiClient.apiKeys.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.detail(variables.id) });
    },
  });
}

export function useDeleteApiKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.apiKeys.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.all });
    },
  });
}

export function useRevokeApiKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.admin.apiKeys.revoke(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.detail(id) });
    },
  });
}
