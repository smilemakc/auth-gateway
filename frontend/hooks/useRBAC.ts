import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/apiClient';
import { queryKeys } from '../services/queryClient';

// Roles
export function useRoles(page: number = 1, pageSize: number = 50) {
  return useQuery({
    queryKey: queryKeys.rbac.roles.list(page, pageSize),
    queryFn: () => apiClient.admin.rbac.roles.list(page, pageSize),
  });
}

export function useRoleDetail(roleId: string) {
  return useQuery({
    queryKey: queryKeys.rbac.roles.detail(roleId),
    queryFn: () => apiClient.admin.rbac.roles.get(roleId),
    enabled: !!roleId,
  });
}

export function useCreateRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; description?: string; permissions: string[] }) =>
      apiClient.admin.rbac.roles.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
    },
  });
}

export function useUpdateRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      apiClient.admin.rbac.roles.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.detail(variables.id) });
    },
  });
}

export function useDeleteRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (roleId: string) => apiClient.admin.rbac.roles.delete(roleId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.all });
    },
  });
}

// Permissions
export function usePermissions(page: number = 1, pageSize: number = 100) {
  return useQuery({
    queryKey: queryKeys.rbac.permissions.list(page, pageSize),
    queryFn: () => apiClient.admin.rbac.permissions.list(page, pageSize),
  });
}

export function usePermissionDetail(permissionId: string) {
  return useQuery({
    queryKey: queryKeys.rbac.permissions.detail(permissionId),
    queryFn: () => apiClient.admin.rbac.permissions.get(permissionId),
    enabled: !!permissionId,
  });
}

export function useCreatePermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; resource: string; action: string; description?: string }) =>
      apiClient.admin.rbac.permissions.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.all });
    },
  });
}

export function useUpdatePermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      apiClient.admin.rbac.permissions.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.permissions.detail(variables.id) });
    },
  });
}

export function useDeletePermission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (permissionId: string) => apiClient.admin.rbac.permissions.delete(permissionId),
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
      apiClient.admin.rbac.roles.assignPermission(roleId, permissionId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.rbac.roles.detail(variables.roleId) });
    },
  });
}

export function useRevokePermissionFromRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ roleId, permissionId }: { roleId: string; permissionId: string }) =>
      apiClient.admin.rbac.roles.revokePermission(roleId, permissionId),
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
      queryClient.invalidateQueries({ queryKey: queryKeys.users.list() });
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
      queryClient.invalidateQueries({ queryKey: queryKeys.users.list() });
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
      queryClient.invalidateQueries({ queryKey: queryKeys.users.list() });
    },
  });
}
