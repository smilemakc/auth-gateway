import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type {
  OAuthProviderConfig,
  CreateOAuthProviderRequest,
  UpdateOAuthProviderRequest,
} from '@auth-gateway/client-sdk';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useOAuthProviders() {
  return useQuery({
    queryKey: queryKeys.oauth.providers,
    queryFn: () => apiClient.admin.oauthProviders.list(),
  });
}

export function useOAuthProviderDetail(providerId: string) {
  return useQuery({
    queryKey: queryKeys.oauth.provider(providerId),
    queryFn: () => apiClient.admin.oauthProviders.get(providerId),
    enabled: !!providerId,
  });
}

export function useCreateOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateOAuthProviderRequest) =>
      apiClient.admin.oauthProviders.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.providers });
    },
  });
}

export function useUpdateOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateOAuthProviderRequest }) =>
      apiClient.admin.oauthProviders.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.providers });
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.provider(variables.id) });
    },
  });
}

export function useDeleteOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (providerId: string) => apiClient.admin.oauthProviders.delete(providerId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.providers });
    },
  });
}

export function useToggleOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      enabled ? apiClient.admin.oauthProviders.enable(id) : apiClient.admin.oauthProviders.disable(id),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.providers });
      queryClient.invalidateQueries({ queryKey: queryKeys.oauth.provider(variables.id) });
    },
  });
}
