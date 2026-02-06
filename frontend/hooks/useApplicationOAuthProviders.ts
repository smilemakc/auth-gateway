import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type {
  ApplicationOAuthProvider,
  CreateAppOAuthProviderRequest,
  UpdateAppOAuthProviderRequest,
} from '@auth-gateway/client-sdk';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

export function useApplicationOAuthProviders(appId: string) {
  return useQuery({
    queryKey: queryKeys.applications.oauthProviders(appId),
    queryFn: () => apiClient.admin.appOAuthProviders.list(appId),
    enabled: !!appId,
  });
}

export function useApplicationOAuthProviderDetail(appId: string, providerId: string) {
  return useQuery({
    queryKey: queryKeys.applications.oauthProvider(appId, providerId),
    queryFn: () => apiClient.admin.appOAuthProviders.getById(appId, providerId),
    enabled: !!appId && !!providerId,
  });
}

export function useCreateApplicationOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, data }: { appId: string; data: CreateAppOAuthProviderRequest }) =>
      apiClient.admin.appOAuthProviders.create(appId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.oauthProviders(variables.appId) });
    },
  });
}

export function useUpdateApplicationOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, id, data }: { appId: string; id: string; data: UpdateAppOAuthProviderRequest }) =>
      apiClient.admin.appOAuthProviders.update(appId, id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.oauthProviders(variables.appId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.oauthProvider(variables.appId, variables.id) });
    },
  });
}

export function useDeleteApplicationOAuthProvider() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ appId, id }: { appId: string; id: string }) =>
      apiClient.admin.appOAuthProviders.delete(appId, id),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.applications.oauthProviders(variables.appId) });
    },
  });
}
