/**
 * Admin Groups service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  Group,
  CreateGroupRequest,
  UpdateGroupRequest,
  AddGroupMembersRequest,
  GroupListResponse,
  GroupMembersResponse,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin Groups service for group management */
export class AdminGroupsService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all groups with pagination
   * @param page Page number
   * @param pageSize Items per page
   * @returns Paginated list of groups
   */
  async list(page = 1, pageSize = 20): Promise<GroupListResponse> {
    const response = await this.http.get<GroupListResponse>('/api/admin/groups', {
      headers: {},
      query: { page, page_size: pageSize },
    });
    return response.data;
  }

  /**
   * Get a specific group by ID
   * @param id Group ID
   * @returns Group details
   */
  async get(id: string): Promise<Group> {
    const response = await this.http.get<Group>(`/api/admin/groups/${id}`);
    return response.data;
  }

  /**
   * Create a new group
   * @param data Group data
   * @returns Created group
   */
  async create(data: CreateGroupRequest): Promise<Group> {
    const response = await this.http.post<Group>('/api/admin/groups', data);
    return response.data;
  }

  /**
   * Update a group
   * @param id Group ID
   * @param data Update data
   * @returns Updated group
   */
  async update(id: string, data: UpdateGroupRequest): Promise<Group> {
    const response = await this.http.put<Group>(`/api/admin/groups/${id}`, data);
    return response.data;
  }

  /**
   * Delete a group
   * @param id Group ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(`/api/admin/groups/${id}`);
    return response.data;
  }

  /**
   * Get group members
   * @param id Group ID
   * @param page Page number
   * @param pageSize Items per page
   * @returns Paginated list of group members
   */
  async getMembers(id: string, page = 1, pageSize = 20): Promise<GroupMembersResponse> {
    const response = await this.http.get<GroupMembersResponse>(`/api/admin/groups/${id}/members`, {
      headers: {},
      query: { page, page_size: pageSize },
    });
    return response.data;
  }

  /**
   * Add members to a group
   * @param id Group ID
   * @param data Member data
   * @returns Success message
   */
  async addMembers(id: string, data: AddGroupMembersRequest): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(`/api/admin/groups/${id}/members`, data);
    return response.data;
  }

  /**
   * Remove a member from a group
   * @param id Group ID
   * @param userId User ID
   * @returns Success message
   */
  async removeMember(id: string, userId: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(`/api/admin/groups/${id}/members/${userId}`);
    return response.data;
  }
}
