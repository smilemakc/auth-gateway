import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type { AdminCreateUserRequest, AdminUpdateUserRequest } from '@auth-gateway/client-sdk';
import { useCurrentAppId } from './useAppAwareQuery';

export function useUsers(page: number = 1, pageSize: number = 50) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.users.list(page, pageSize), appId],
    queryFn: async () => {
      const response = await apiClient.admin.users.list(page, pageSize);
      return { users: response.users, total: response.total };
    },
  });
}

export function useUserDetail(userId: string) {
  return useQuery({
    queryKey: queryKeys.users.detail(userId),
    queryFn: () => apiClient.admin.users.get(userId),
    enabled: !!userId,
  });
}

export function useUserStats() {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.users.stats, appId],
    queryFn: () => apiClient.admin.users.getStats(),
  });
}

export function useUpdateUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: AdminUpdateUserRequest }) =>
      apiClient.admin.users.update(id, data),
    onSuccess: (_, variables) => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.detail(variables.id) });
    },
  });
}

export function useDeleteUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: string) => apiClient.admin.users.delete(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}

export function useCreateUser() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: AdminCreateUserRequest) => apiClient.admin.users.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}

/**
 * Hook for resetting user 2FA (admin only)
 */
export function useReset2FA() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (userId: string) => {
      return apiClient.admin.users.reset2FA(userId);
    },
    onSuccess: (_, userId) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.detail(userId) });
    },
  });
}

/**
 * Hook for sending password reset email (admin only)
 */
export function useSendPasswordReset() {
  return useMutation({
    mutationFn: async (userId: string) => {
      return apiClient.admin.users.sendPasswordReset(userId);
    },
  });
}

/**
 * Hook for fetching user OAuth accounts (admin only)
 */
export function useUserOAuthAccounts(userId: string) {
  return useQuery({
    queryKey: ['users', 'oauth-accounts', userId],
    queryFn: () => apiClient.admin.users.getOAuthAccounts(userId),
    enabled: !!userId,
  });
}
