/**
 * RBAC (Role-Based Access Control) types
 */

import type { TimestampedEntity } from './common';

/** Permission entity */
export interface Permission extends TimestampedEntity {
  name: string;
  resource: string;
  action: string;
  description?: string;
}

/** Create permission request */
export interface CreatePermissionRequest {
  name: string;
  resource: string;
  action: string;
  description?: string;
}

/** Role entity */
export interface Role extends TimestampedEntity {
  name: string;
  display_name: string;
  description?: string;
  is_system_role: boolean;
  permissions: Permission[];
}

/** Create role request */
export interface CreateRoleRequest {
  name: string;
  display_name: string;
  description?: string;
  permissions: string[]; // Permission IDs
}

/** Update role request */
export interface UpdateRoleRequest {
  display_name?: string;
  description?: string;
  permissions?: string[]; // Permission IDs
}

/** Permission matrix entry */
export interface PermissionMatrixEntry {
  role: string;
  permissions: Record<string, boolean>;
}

/** Permission matrix response */
export interface PermissionMatrix {
  roles: string[];
  resources: string[];
  actions: string[];
  matrix: PermissionMatrixEntry[];
}
