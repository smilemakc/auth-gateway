/**
 * Admin Users service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  AdminStatsResponse,
  AdminUpdateUserRequest,
  AdminCreateUserRequest,
  AdminUserListResponse,
  AdminUserResponse,
} from '../../types/user';
import { BaseService } from '../base';

/** Admin Users service for user management */
export class AdminUsersService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Get admin dashboard statistics
   * @returns System statistics
   */
  async getStats(): Promise<AdminStatsResponse> {
    const response = await this.http.get<AdminStatsResponse>('/admin/stats');
    return response.data;
  }

  /**
   * List all users with pagination
   * @param page Page number
   * @param pageSize Items per page
   * @returns Paginated list of users
   */
  async list(page = 1, pageSize = 20): Promise<AdminUserListResponse> {
    const response = await this.http.get<AdminUserListResponse>('/admin/users', {
      headers: {},
      query: { page, page_size: pageSize },
    });
    return response.data;
  }

  /**
   * Get a specific user by ID
   * @param id User ID
   * @returns User details
   */
  async get(id: string): Promise<AdminUserResponse> {
    const response = await this.http.get<AdminUserResponse>(`/admin/users/${id}`);
    return response.data;
  }

  /**
   * Create a new user
   * @param data User creation data
   * @returns Created user
   */
  async create(data: AdminCreateUserRequest): Promise<AdminUserResponse> {
    const response = await this.http.post<AdminUserResponse>('/admin/users', data);
    return response.data;
  }

  /**
   * Update a user
   * @param id User ID
   * @param data Update data (roleIds, isActive)
   * @returns Updated user
   */
  async update(id: string, data: AdminUpdateUserRequest): Promise<AdminUserResponse> {
    const response = await this.http.put<AdminUserResponse>(
      `/admin/users/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete a user
   * @param id User ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(`/admin/users/${id}`);
    return response.data;
  }

  /**
   * Activate a user
   * @param id User ID
   * @returns Updated user
   */
  async activate(id: string): Promise<AdminUserResponse> {
    return this.update(id, { is_active: true });
  }

  /**
   * Deactivate a user
   * @param id User ID
   * @returns Updated user
   */
  async deactivate(id: string): Promise<AdminUserResponse> {
    return this.update(id, { is_active: false });
  }

  /**
   * Change user role
   * @deprecated Use setRoles() instead for multi-role support
   * @param id User ID
   * @param roleId Role ID
   * @returns Updated user
   */
  async setRole(id: string, roleId: string): Promise<AdminUserResponse> {
    return this.setRoles(id, [roleId]);
  }

  /**
   * Set user roles (replaces all existing roles)
   * @param id User ID
   * @param roleIds Array of role IDs
   * @returns Updated user
   */
  async setRoles(id: string, roleIds: string[]): Promise<AdminUserResponse> {
    return this.update(id, { role_ids: roleIds });
  }

  /**
   * Assign a role to a user
   * @param id User ID
   * @param roleId Role ID to assign
   * @returns Success message
   */
  async assignRole(id: string, roleId: string): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/admin/users/${id}/roles`,
      { role_id: roleId }
    );
    return response.data;
  }

  /**
   * Remove a role from a user
   * @param id User ID
   * @param roleId Role ID to remove
   * @returns Success message
   */
  async removeRole(id: string, roleId: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/admin/users/${id}/roles/${roleId}`
    );
    return response.data;
  }

  /**
   * Search users (basic implementation - fetch and filter)
   * @param query Search query (email, username, full name)
   * @param maxResults Maximum results to return
   * @returns Matching users
   */
  async search(query: string, maxResults = 50): Promise<AdminUserResponse[]> {
    const results: AdminUserResponse[] = [];
    const lowerQuery = query.toLowerCase();
    let page = 1;

    while (results.length < maxResults) {
      const response = await this.list(page, 100);
      const matches = response.users.filter(
        (user) =>
          user.email.toLowerCase().includes(lowerQuery) ||
          user.username.toLowerCase().includes(lowerQuery) ||
          user.full_name.toLowerCase().includes(lowerQuery)
      );

      results.push(...matches);

      if (response.users.length < 100 || page * 100 >= response.total) {
        break;
      }
      page++;
    }

    return results.slice(0, maxResults);
  }

  /**
   * Reset 2FA for a user (admin only)
   * Disables TOTP and deletes backup codes
   * @param userId User ID
   * @returns Success message
   */
  async reset2FA(userId: string): Promise<{ message: string; user_id: string }> {
    const response = await this.http.post<{ message: string; user_id: string }>(
      `/admin/users/${userId}/reset-2fa`
    );
    return response.data;
  }

  /**
   * Send password reset email for a user (admin only)
   * Initiates password reset flow for the user
   * @param userId User ID
   * @returns Success message with email
   */
  async sendPasswordReset(userId: string): Promise<{ message: string; email: string }> {
    const response = await this.http.post<{ message: string; email: string }>(
      `/admin/users/${userId}/send-password-reset`
    );
    return response.data;
  }
}
