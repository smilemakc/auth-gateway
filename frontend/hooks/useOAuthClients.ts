import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type {
  CreateOAuthClientRequest,
  UpdateOAuthClientRequest,
} from '@auth-gateway/client-sdk';
import { useCurrentAppId } from './useAppAwareQuery';

// OAuth Clients
export function useOAuthClients(page: number = 1, pageSize: number = 20, isActive?: boolean) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.oauthClients.list(page, pageSize), appId, isActive],
    queryFn: () => apiClient.admin.oauthClients.list({ page, per_page: pageSize, is_active: isActive }),
  });
}

export function useOAuthClientDetail(clientId: string) {
  return useQuery({
    queryKey: queryKeys.oauthClients.detail(clientId),
    queryFn: () => apiClient.admin.oauthClients.get(clientId),
    enabled: !!clientId,
  });
}

export function useCreateOAuthClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateOAuthClientRequest) =>
      apiClient.admin.oauthClients.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.all });
    },
  });
}

export function useUpdateOAuthClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateOAuthClientRequest }) =>
      apiClient.admin.oauthClients.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.detail(variables.id) });
    },
  });
}

export function useDeleteOAuthClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (clientId: string) => apiClient.admin.oauthClients.delete(clientId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.all });
    },
  });
}

export function useRotateOAuthClientSecret() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (clientId: string) => apiClient.admin.oauthClients.rotateSecret(clientId),
    onSuccess: (_, clientId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.detail(clientId) });
    },
  });
}

export function useActivateOAuthClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (clientId: string) => apiClient.admin.oauthClients.activate(clientId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.all });
    },
  });
}

export function useDeactivateOAuthClient() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (clientId: string) => apiClient.admin.oauthClients.deactivate(clientId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.all });
    },
  });
}

// OAuth Scopes
export function useOAuthScopes() {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.oauthClients.scopes, appId],
    queryFn: () => apiClient.admin.oauthClients.listScopes(),
  });
}

export function useCreateOAuthScope() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; display_name: string; description?: string }) =>
      apiClient.admin.oauthClients.createScope(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.scopes });
    },
  });
}

export function useDeleteOAuthScope() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (scopeId: string) => apiClient.admin.oauthClients.deleteScope(scopeId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.scopes });
    },
  });
}

// User Consents
export function useOAuthClientConsents(clientId: string) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.oauthClients.consents(clientId), appId],
    queryFn: () => apiClient.admin.oauthClients.listClientConsents(clientId),
    enabled: !!clientId,
  });
}

export function useRevokeOAuthConsent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ clientId, userId }: { clientId: string; userId: string }) =>
      apiClient.admin.oauthClients.revokeUserConsent(clientId, userId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.oauthClients.consents(variables.clientId) });
    },
  });
}
