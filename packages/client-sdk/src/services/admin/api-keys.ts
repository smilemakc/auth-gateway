/**
 * Admin API Keys service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type { AdminAPIKeyResponse } from '../../types/api-key';
import { BaseService } from '../base';

/** Admin API key list response */
interface AdminAPIKeyListResponse {
  apiKeys: AdminAPIKeyResponse[];
  total: number;
  page: number;
  pageSize: number;
}

/** Admin API Keys service for system-wide API key management */
export class AdminAPIKeysService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all API keys in the system
   * @param page Page number
   * @param pageSize Items per page
   * @returns Paginated list of all API keys
   */
  async list(page = 1, pageSize = 50): Promise<AdminAPIKeyListResponse> {
    const response = await this.http.get<AdminAPIKeyListResponse>(
      '/admin/api-keys',
      { query: { page, page_size: pageSize } }
    );
    return response.data;
  }

  /**
   * Revoke an API key
   * @param id API key ID
   * @returns Success message
   */
  async revoke(id: string): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/admin/api-keys/${id}/revoke`
    );
    return response.data;
  }

  /**
   * Get API keys for a specific user
   * @param userId User ID
   * @returns User's API keys
   */
  async getByUser(userId: string): Promise<AdminAPIKeyResponse[]> {
    const allKeys: AdminAPIKeyResponse[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const response = await this.list(page, 100);
      const userKeys = response.apiKeys.filter((k) => k.userId === userId);
      allKeys.push(...userKeys);
      hasMore = page * 100 < response.total;
      page++;
    }

    return allKeys;
  }

  /**
   * Get all active API keys
   * @returns List of active API keys
   */
  async getActive(): Promise<AdminAPIKeyResponse[]> {
    const allKeys: AdminAPIKeyResponse[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const response = await this.list(page, 100);
      const activeKeys = response.apiKeys.filter((k) => k.isActive);
      allKeys.push(...activeKeys);
      hasMore = page * 100 < response.total;
      page++;
    }

    return allKeys;
  }

  /**
   * Revoke all API keys for a user
   * @param userId User ID
   * @returns Number of keys revoked
   */
  async revokeAllForUser(userId: string): Promise<number> {
    const userKeys = await this.getByUser(userId);
    const activeKeys = userKeys.filter((k) => k.isActive);

    for (const key of activeKeys) {
      await this.revoke(key.id);
    }

    return activeKeys.length;
  }
}
