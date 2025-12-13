import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useApiKeys(page: number = 1, pageSize: number = 50) {
  return useQuery({
    queryKey: queryKeys.apiKeys.list(page, pageSize),
    queryFn: async () => {
      const response = await apiClient.admin.apiKeys.list(page, pageSize);
      return response;
    },
  });
}

export function useApiKeyDetail(id: string) {
  return useQuery({
    queryKey: queryKeys.apiKeys.detail(id),
    queryFn: () => apiClient.admin.apiKeys.get(id),
    enabled: !!id,
  });
}

export function useCreateApiKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; scopes: string[]; expiresAt?: string }) =>
      apiClient.admin.apiKeys.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.all });
    },
  });
}

export function useUpdateApiKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      apiClient.admin.apiKeys.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.detail(variables.id) });
    },
  });
}

export function useDeleteApiKey() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.admin.apiKeys.delete(id),
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
