import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useOAuthProviders(page: number = 1, pageSize: number = 50) {
  return useQuery({
    queryKey: queryKeys.oauth.list(page, pageSize),
    queryFn: () => apiClient.admin.oauth.providers.list(page, pageSize),
  });
}

export function useOAuthProviderDetail(providerId: string) {
  return useQuery({
    queryKey: queryKeys.oauth.detail(providerId),
    queryFn: () => apiClient.admin.oauth.providers.get(providerId),
    enabled: !!providerId,
  });
}

export function useCreateOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      name: string;
      provider: string;
      clientId: string;
      clientSecret: string;
      redirectUri?: string;
      scopes?: string[];
      enabled?: boolean;
    }) => apiClient.admin.oauth.providers.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.all });
    },
  });
}

export function useUpdateOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      apiClient.admin.oauth.providers.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.detail(variables.id) });
    },
  });
}

export function useDeleteOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (providerId: string) => apiClient.admin.oauth.providers.delete(providerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.all });
    },
  });
}

export function useToggleOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      apiClient.admin.oauth.providers.update(id, { enabled }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.detail(variables.id) });
    },
  });
}
