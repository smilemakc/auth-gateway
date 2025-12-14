/**
 * User-related types
 */

import type { AccountType, TimestampedEntity, UserRole } from './common';

/** Role information */
export interface RoleInfo {
  id: string;
  name: string;
  display_name: string;
}

/** User entity */
export interface User extends TimestampedEntity {
  email: string;
  phone?: string;
  username: string;
  full_name: string;
  profile_picture_url?: string;
  roles: RoleInfo[];
  account_type: AccountType;
  email_verified: boolean;
  phone_verified: boolean;
  is_active: boolean;
  totp_enabled: boolean;
  totp_enabled_at?: string;
}

/** User for public display (minimal info) */
export interface PublicUser {
  id: string;
  username: string;
  full_name: string;
  profile_picture_url?: string;
}

/** User profile update request */
export interface UpdateProfileRequest {
  full_name?: string;
  profile_picture_url?: string;
}

/** Change password request */
export interface ChangePasswordRequest {
  old_password: string;
  new_password: string;
}

/** Admin user response with additional fields */
export interface AdminUserResponse extends User {
  last_login?: string;
}

/** Admin user update request (snake_case for backend API) */
export interface AdminUpdateUserRequest {
  role_ids?: string[];
  is_active?: boolean;
}

/** Admin user create request (snake_case for backend API) */
export interface AdminCreateUserRequest {
  email: string;
  username: string;
  password: string;
  full_name: string;
  role_ids?: string[];
  account_type?: 'human' | 'service';
}


/** Assign role request (snake_case for backend API) */
export interface AssignRoleRequest {
  role_id: string;
}

/** Admin user list response */
export interface AdminUserListResponse {
  users: AdminUserResponse[];
  total: number;
  page: number;
  page_size: number;
}

/** Admin statistics response */
export interface AdminStatsResponse {
  total_users: number;
  active_users: number;
  new_users_today: number;
  total_api_keys: number;
  active_api_keys: number;
  login_attempts_today: number;
  failed_login_attempts_today: number;
}
