import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type { AdminCreateUserRequest, AdminUpdateUserRequest } from '@auth-gateway/client-sdk';

export function useUsers(page: number = 1, pageSize: number = 50, search?: string, role?: string) {
  return useQuery({
    queryKey: queryKeys.users.list(page, pageSize, search, role),
    queryFn: async () => {
      const response = await apiClient.admin.users.list(page, pageSize);
      const { users, total } = response;

      // Client-side filtering for search and role (if SDK doesn't support it)
      let filtered = users;
      if (search) {
        filtered = filtered.filter(
          (u: any) =>
            u.username?.toLowerCase().includes(search.toLowerCase()) ||
            u.email?.toLowerCase().includes(search.toLowerCase()) ||
            u.full_name?.toLowerCase().includes(search.toLowerCase())
        );
      }
      if (role && role !== 'all') {
        filtered = filtered.filter((u: any) =>
          u.roles?.some((r: any) => r.name === role)
        );
      }

      return { users: filtered, total };
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
  return useQuery({
    queryKey: queryKeys.users.stats,
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
    queryFn: async () => {
      const response = await fetch(`/api/admin/users/${userId}/oauth-accounts`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      });
      if (!response.ok) {
        throw new Error(`Failed to fetch OAuth accounts: ${response.status}`);
      }
      return response.json();
    },
    enabled: !!userId,
  });
}
