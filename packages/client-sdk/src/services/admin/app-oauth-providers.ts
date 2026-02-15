/**
 * Admin Application OAuth Providers service
 */

import { BaseService } from '../base';
import type { HttpClient } from '../../core/http';
import type {
  ApplicationOAuthProvider,
  CreateAppOAuthProviderRequest,
  UpdateAppOAuthProviderRequest,
} from '../../types';

/** Admin Application OAuth Providers service for managing OAuth providers per application */
export class AdminAppOAuthProvidersService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all OAuth providers for an application
   * @param appId Application ID
   * @returns Array of OAuth providers
   */
  async list(appId: string): Promise<ApplicationOAuthProvider[]> {
    const response = await this.http.get<{ providers: ApplicationOAuthProvider[]; total: number }>(
      `/api/admin/applications/${appId}/oauth-providers`
    );
    return response.data.providers;
  }

  /**
   * Create a new OAuth provider for an application
   * @param appId Application ID
   * @param data OAuth provider creation data
   * @returns Created OAuth provider
   */
  async create(appId: string, data: CreateAppOAuthProviderRequest): Promise<ApplicationOAuthProvider> {
    const response = await this.http.post<ApplicationOAuthProvider>(
      `/api/admin/applications/${appId}/oauth-providers`,
      data
    );
    return response.data;
  }

  /**
   * Get a specific OAuth provider by ID
   * @param appId Application ID
   * @param id OAuth provider ID
   * @returns OAuth provider details
   */
  async getById(appId: string, id: string): Promise<ApplicationOAuthProvider> {
    const response = await this.http.get<ApplicationOAuthProvider>(
      `/api/admin/applications/${appId}/oauth-providers/${id}`
    );
    return response.data;
  }

  /**
   * Update an OAuth provider
   * @param appId Application ID
   * @param id OAuth provider ID
   * @param data Update data
   * @returns Updated OAuth provider
   */
  async update(appId: string, id: string, data: UpdateAppOAuthProviderRequest): Promise<ApplicationOAuthProvider> {
    const response = await this.http.put<ApplicationOAuthProvider>(
      `/api/admin/applications/${appId}/oauth-providers/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete an OAuth provider
   * @param appId Application ID
   * @param id OAuth provider ID
   * @returns void
   */
  async delete(appId: string, id: string): Promise<void> {
    await this.http.delete(`/api/admin/applications/${appId}/oauth-providers/${id}`);
  }
}
