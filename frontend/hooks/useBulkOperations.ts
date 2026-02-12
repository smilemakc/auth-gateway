import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import type {
  BulkCreateUsersRequest,
  BulkUpdateUsersRequest,
  BulkDeleteUsersRequest,
  BulkAssignRolesRequest,
} from '@auth-gateway/client-sdk';

export function useBulkCreateUsers() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: BulkCreateUsersRequest) => apiClient.admin.bulk.bulkCreateUsers(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}

export function useBulkUpdateUsers() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: BulkUpdateUsersRequest) => apiClient.admin.bulk.bulkUpdateUsers(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}

export function useBulkDeleteUsers() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: BulkDeleteUsersRequest) => apiClient.admin.bulk.bulkDeleteUsers(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}

export function useBulkAssignRoles() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: BulkAssignRolesRequest) => apiClient.admin.bulk.bulkAssignRoles(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.all });
    },
  });
}

