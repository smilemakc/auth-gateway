/**
 * Admin Users service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  AdminStatsResponse,
  AdminUpdateUserRequest,
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
   * Update a user
   * @param id User ID
   * @param data Update data (role, isActive)
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
    return this.update(id, { isActive: true });
  }

  /**
   * Deactivate a user
   * @param id User ID
   * @returns Updated user
   */
  async deactivate(id: string): Promise<AdminUserResponse> {
    return this.update(id, { isActive: false });
  }

  /**
   * Change user role
   * @param id User ID
   * @param role New role
   * @returns Updated user
   */
  async setRole(
    id: string,
    role: 'user' | 'moderator' | 'admin'
  ): Promise<AdminUserResponse> {
    return this.update(id, { role });
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
          user.fullName.toLowerCase().includes(lowerQuery)
      );

      results.push(...matches);

      if (response.users.length < 100 || page * 100 >= response.total) {
        break;
      }
      page++;
    }

    return results.slice(0, maxResults);
  }
}
