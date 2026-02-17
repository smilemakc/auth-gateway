import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../services/apiClient';
import { queryKeys } from '../../services/queryClient';

export function useAssignPermissionToRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) =>
      apiClient.admin.rbac.addPermissionsToRole(roleId, [permissionId]),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.detail(variables.roleId) });
    },
  });
}

export function useRevokePermissionFromRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) =>
      apiClient.admin.rbac.removePermissionsFromRole(roleId, [permissionId]),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.detail(variables.roleId) });
    },
  });
}
