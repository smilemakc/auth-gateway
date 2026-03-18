import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../services/apiClient';
import { queryKeys } from '../../services/queryClient';
import { useCurrentAppId } from '../useAppAwareQuery';

export function useRoles(page: number = 1, pageSize: number = 50) {
  const appId = useCurrentAppId();
  return useQuery({
    queryKey: [...queryKeys.rbac.roles.list(page, pageSize), appId],
    queryFn: () => apiClient.admin.rbac.listRoles(),
  });
}

export function useRoleDetail(roleId: string) {
  return useQuery({
    queryKey: queryKeys.rbac.roles.detail(roleId),
    queryFn: () => apiClient.admin.rbac.getRole(roleId),
    enabled: !!roleId,
  });
}

export function useCreateRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: any) =>
      apiClient.admin.rbac.createRole(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
    },
  });
}

export function useUpdateRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      apiClient.admin.rbac.updateRole(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.detail(variables.id) });
    },
  });
}

export function useDeleteRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (roleId: string) => apiClient.admin.rbac.deleteRole(roleId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
    },
  });
}
