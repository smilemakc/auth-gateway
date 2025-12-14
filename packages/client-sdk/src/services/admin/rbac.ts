/**
 * Admin RBAC service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  CreatePermissionRequest,
  CreateRoleRequest,
  Permission,
  PermissionMatrix,
  Role,
  UpdateRoleRequest,
} from '../../types/rbac';
import { BaseService } from '../base';

/** Admin RBAC service for roles and permissions management */
export class AdminRBACService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  // ==================== PERMISSIONS ====================

  /**
   * List all permissions
   * @returns List of permissions
   */
  async listPermissions(): Promise<Permission[]> {
    const response = await this.http.get<Permission[]>('/admin/rbac/permissions');
    return response.data;
  }

  /**
   * Create a new permission
   * @param data Permission data
   * @returns Created permission
   */
  async createPermission(data: CreatePermissionRequest): Promise<Permission> {
    const response = await this.http.post<Permission>(
      '/admin/rbac/permissions',
      data
    );
    return response.data;
  }

  /**
   * Delete a permission
   * @param id Permission ID
   * @returns Success message
   */
  async deletePermission(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/admin/rbac/permissions/${id}`
    );
    return response.data;
  }

  /**
   * Get permissions by resource
   * @param resource Resource name
   * @returns Permissions for the resource
   */
  async getPermissionsByResource(resource: string): Promise<Permission[]> {
    const permissions = await this.listPermissions();
    return permissions.filter((p) => p.resource === resource);
  }

  // ==================== ROLES ====================

  /**
   * List all roles
   * @returns List of roles with their permissions
   */
  async listRoles(): Promise<Role[]> {
    const response = await this.http.get<Role[]>('/admin/rbac/roles');
    return response.data;
  }

  /**
   * Get a specific role by ID
   * @param id Role ID
   * @returns Role details with permissions
   */
  async getRole(id: string): Promise<Role> {
    const response = await this.http.get<Role>(`/admin/rbac/roles/${id}`);
    return response.data;
  }

  /**
   * Create a new role
   * @param data Role data
   * @returns Created role
   */
  async createRole(data: CreateRoleRequest): Promise<Role> {
    const response = await this.http.post<Role>('/admin/rbac/roles', data);
    return response.data;
  }

  /**
   * Update a role
   * @param id Role ID
   * @param data Update data
   * @returns Updated role
   */
  async updateRole(id: string, data: UpdateRoleRequest): Promise<Role> {
    const response = await this.http.put<Role>(`/admin/rbac/roles/${id}`, data);
    return response.data;
  }

  /**
   * Delete a role
   * @param id Role ID
   * @returns Success message
   */
  async deleteRole(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/admin/rbac/roles/${id}`
    );
    return response.data;
  }

  /**
   * Add permissions to a role
   * @param roleId Role ID
   * @param permissionIds Permission IDs to add
   * @returns Updated role
   */
  async addPermissionsToRole(
    roleId: string,
    permissionIds: string[]
  ): Promise<Role> {
    const role = await this.getRole(roleId);
    const existingIds = role.permissions.map((p) => p.id);
    const newIds = [...new Set([...existingIds, ...permissionIds])];

    return this.updateRole(roleId, { permissions: newIds });
  }

  /**
   * Remove permissions from a role
   * @param roleId Role ID
   * @param permissionIds Permission IDs to remove
   * @returns Updated role
   */
  async removePermissionsFromRole(
    roleId: string,
    permissionIds: string[]
  ): Promise<Role> {
    const role = await this.getRole(roleId);
    const removeSet = new Set(permissionIds);
    const newIds = role.permissions
      .filter((p) => !removeSet.has(p.id))
      .map((p) => p.id);

    return this.updateRole(roleId, { permissions: newIds });
  }

  // ==================== PERMISSION MATRIX ====================

  /**
   * Get the permission matrix
   * Shows which roles have which permissions
   * @returns Permission matrix
   */
  async getPermissionMatrix(): Promise<PermissionMatrix> {
    const response = await this.http.get<PermissionMatrix>(
      '/admin/rbac/permission-matrix'
    );
    return response.data;
  }

  /**
   * Check if a role has a specific permission
   * @param roleName Role name
   * @param permissionName Permission name
   * @returns True if role has permission
   */
  async roleHasPermission(
    roleName: string,
    permissionName: string
  ): Promise<boolean> {
    const matrix = await this.getPermissionMatrix();
    const roleEntry = matrix.matrix.find((m) => m.role === roleName);

    if (!roleEntry) {
      return false;
    }

    return roleEntry.permissions[permissionName] === true;
  }

  /**
   * Get all non-system roles (custom roles)
   * @returns List of custom roles
   */
  async getCustomRoles(): Promise<Role[]> {
    const roles = await this.listRoles();
    return roles.filter((r) => !r.is_system_role);
  }

  /**
   * Get all system roles
   * @returns List of system roles
   */
  async getSystemRoles(): Promise<Role[]> {
    const roles = await this.listRoles();
    return roles.filter((r) => r.is_system_role);
  }
}
