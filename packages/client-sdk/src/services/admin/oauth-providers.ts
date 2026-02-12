/**
 * Admin OAuth Providers service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  OAuthProviderConfig,
  OAuthProviderType,
  OAuthProviderListResponse,
  CreateOAuthProviderRequest,
  UpdateOAuthProviderRequest,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin OAuth Providers service for OAuth configuration management */
export class AdminOAuthProvidersService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all OAuth provider configurations
   * @returns List of OAuth providers
   */
  async list(): Promise<OAuthProviderConfig[]> {
    const response = await this.http.get<OAuthProviderListResponse>(
      '/api/admin/oauth/providers'
    );
    return response.data.providers;
  }

  /**
   * Get OAuth provider by ID
   * @param id Provider ID
   * @returns Provider configuration
   */
  async get(id: string): Promise<OAuthProviderConfig> {
    const response = await this.http.get<OAuthProviderConfig>(
      `/api/admin/oauth/providers/${id}`
    );
    return response.data;
  }

  /**
   * Get OAuth provider by type
   * @param provider Provider type (google, github, etc.)
   * @returns Provider configuration or undefined
   */
  async getByType(
    provider: OAuthProviderType
  ): Promise<OAuthProviderConfig | undefined> {
    const providers = await this.list();
    return providers.find((p) => p.provider === provider);
  }

  /**
   * Create a new OAuth provider configuration
   * @param data Provider creation data
   * @returns Created provider configuration
   */
  async create(data: CreateOAuthProviderRequest): Promise<OAuthProviderConfig> {
    const response = await this.http.post<OAuthProviderConfig>(
      '/api/admin/oauth/providers',
      data
    );
    return response.data;
  }

  /**
   * Update OAuth provider configuration
   * @param id Provider ID
   * @param data Provider update data
   * @returns Updated provider configuration
   */
  async update(
    id: string,
    data: UpdateOAuthProviderRequest
  ): Promise<OAuthProviderConfig> {
    const response = await this.http.put<OAuthProviderConfig>(
      `/api/admin/oauth/providers/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete OAuth provider configuration
   * @param id Provider ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/api/admin/oauth/providers/${id}`
    );
    return response.data;
  }

  /**
   * Enable OAuth provider
   * @param id Provider ID
   * @returns Updated provider configuration
   */
  async enable(id: string): Promise<OAuthProviderConfig> {
    return this.update(id, { is_active: true });
  }

  /**
   * Disable OAuth provider
   * @param id Provider ID
   * @returns Updated provider configuration
   */
  async disable(id: string): Promise<OAuthProviderConfig> {
    return this.update(id, { is_active: false });
  }

  /**
   * Get enabled OAuth providers
   * @returns List of enabled providers
   */
  async getEnabled(): Promise<OAuthProviderConfig[]> {
    const providers = await this.list();
    return providers.filter((p) => p.is_active);
  }

  /**
   * Update callback URL for a provider
   * @param id Provider ID
   * @param callbackUrl New callback URL
   * @returns Updated provider configuration
   */
  async updateCallbackUrl(
    id: string,
    callbackUrl: string
  ): Promise<OAuthProviderConfig> {
    return this.update(id, { callback_url: callbackUrl });
  }
}
