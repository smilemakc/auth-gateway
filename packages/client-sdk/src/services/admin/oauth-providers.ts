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
      '/admin/oauth/providers'
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
      `/admin/oauth/providers/${id}`
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
      '/admin/oauth/providers',
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
      `/admin/oauth/providers/${id}`,
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
      `/admin/oauth/providers/${id}`
    );
    return response.data;
  }

  /**
   * Enable OAuth provider
   * @param id Provider ID
   * @returns Updated provider configuration
   */
  async enable(id: string): Promise<OAuthProviderConfig> {
    return this.update(id, { is_enabled: true });
  }

  /**
   * Disable OAuth provider
   * @param id Provider ID
   * @returns Updated provider configuration
   */
  async disable(id: string): Promise<OAuthProviderConfig> {
    return this.update(id, { is_enabled: false });
  }

  /**
   * Get enabled OAuth providers
   * @returns List of enabled providers
   */
  async getEnabled(): Promise<OAuthProviderConfig[]> {
    const providers = await this.list();
    return providers.filter((p) => p.is_enabled);
  }

  /**
   * Update redirect URIs for a provider
   * @param id Provider ID
   * @param redirectUris New redirect URIs
   * @returns Updated provider configuration
   */
  async updateRedirectUris(
    id: string,
    redirectUris: string[]
  ): Promise<OAuthProviderConfig> {
    return this.update(id, { redirect_uris: redirectUris });
  }
}
