import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../services/apiClient';
import { queryKeys } from '../../services/queryClient';
import { useCurrentAppId } from '../useAppAwareQuery';

export function usePermissions(page: number = 1, pageSize: number = 100) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.rbac.permissions.list(page, pageSize), appId],
    queryFn: () => apiClient.admin.rbac.listPermissions(),
  });
}

export function usePermissionDetail(permissionId: string) {
  return useQuery({
    queryKey: queryKeys.rbac.permissions.detail(permissionId),
    queryFn: () => apiClient.admin.rbac.getPermission(permissionId),
    enabled: !!permissionId,
  });
}

export function useCreatePermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; resource: string; action: string; description?: string }) =>
      apiClient.admin.rbac.createPermission(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.all });
    },
  });
}

export function useUpdatePermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { name?: string; description?: string } }) =>
      apiClient.admin.rbac.updatePermission(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.detail(variables.id) });
    },
  });
}

export function useDeletePermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (permissionId: string) => apiClient.admin.rbac.deletePermission(permissionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
    },
  });
}
