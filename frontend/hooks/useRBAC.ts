import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';
import { useCurrentAppId } from './useAppAwareQuery';

// Roles
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

// Permissions
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
    queryFn: async () => {
      // SDK doesn't have getPermission, fetch all and filter
      const permissions = await apiClient.admin.rbac.listPermissions();
      return permissions.find((p: any) => p.id === permissionId);
    },
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
    mutationFn: async ({ id, data }: { id: string; data: { name?: string; description?: string } }) => {
      const response = await fetch(`/api/admin/rbac/permissions/${id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
        body: JSON.stringify(data),
      });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `Failed to update permission: ${response.status}`);
      }
      return response.json();
    },
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
    },
  });
}

// Assign/Revoke permissions to/from roles
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

// User Role Management
export function useAssignUserRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ userId, roleId }: { userId: string; roleId: string }) => {
      return await apiClient.admin.users.assignRole(userId, roleId);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.detail(variables.userId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}

export function useRemoveUserRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ userId, roleId }: { userId: string; roleId: string }) => {
      return await apiClient.admin.users.removeRole(userId, roleId);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.detail(variables.userId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}

export function useSetUserRoles() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ userId, roleIds }: { userId: string; roleIds: string[] }) => {
      return await apiClient.admin.users.setRoles(userId, roleIds);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.users.detail(variables.userId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.users.all });
    },
  });
}
