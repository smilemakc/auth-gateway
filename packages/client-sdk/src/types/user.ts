/**
 * User-related types
 */

import type { AccountType, TimestampedEntity, UserRole } from './common';

/** User entity */
export interface User extends TimestampedEntity {
  email: string;
  phone?: string;
  username: string;
  fullName: string;
  profilePictureUrl?: string;
  role: UserRole;
  roleId?: string;
  accountType: AccountType;
  emailVerified: boolean;
  phoneVerified: boolean;
  isActive: boolean;
  totpEnabled: boolean;
  totpEnabledAt?: string;
}

/** User for public display (minimal info) */
export interface PublicUser {
  id: string;
  username: string;
  fullName: string;
  profilePictureUrl?: string;
}

/** User profile update request */
export interface UpdateProfileRequest {
  fullName?: string;
  profilePictureUrl?: string;
}

/** Change password request */
export interface ChangePasswordRequest {
  oldPassword: string;
  newPassword: string;
}

/** Admin user response with additional fields */
export interface AdminUserResponse extends User {
  lastLogin?: string;
}

/** Admin user update request */
export interface AdminUpdateUserRequest {
  role?: UserRole;
  isActive?: boolean;
}

/** Admin user list response */
export interface AdminUserListResponse {
  users: AdminUserResponse[];
  total: number;
  page: number;
  pageSize: number;
}

/** Admin statistics response */
export interface AdminStatsResponse {
  totalUsers: number;
  activeUsers: number;
  newUsersToday: number;
  totalApiKeys: number;
  activeApiKeys: number;
  loginAttemptsToday: number;
  failedLoginAttemptsToday: number;
}
