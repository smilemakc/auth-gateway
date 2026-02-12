/**
 * Admin OAuth Clients service
 */

import type { HttpClient } from '../../core/http';
import type {
  OAuthClient,
  CreateOAuthClientRequest,
  CreateOAuthClientResponse,
  UpdateOAuthClientRequest,
  RotateSecretResponse,
  OAuthClientListResponse,
  OAuthScope,
  CreateScopeRequest,
  OAuthScopeListResponse,
  UserConsentListResponse,
} from '../../types/oauth-provider';
import { BaseService } from '../base';

/** List clients query parameters */
export interface ListClientsParams extends Record<string, string | number | boolean | undefined> {
  page?: number;
  per_page?: number;
  owner_id?: string;
  is_active?: boolean;
}

/** Admin OAuth Clients service for OAuth client management */
export class AdminOAuthClientsService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  // ============= Client Management =============

  /**
   * Create a new OAuth client
   * @param data Client creation data
   * @returns Client with plain client_secret (shown only once)
   */
  async create(data: CreateOAuthClientRequest): Promise<CreateOAuthClientResponse> {
    const response = await this.http.post<CreateOAuthClientResponse>(
      '/api/admin/oauth/clients',
      data
    );
    return response.data;
  }

  /**
   * List OAuth clients with pagination
   * @param params Query parameters
   * @returns Paginated list of clients
   */
  async list(params?: ListClientsParams): Promise<OAuthClientListResponse> {
    const response = await this.http.get<OAuthClientListResponse>(
      '/api/admin/oauth/clients',
      { query: params }
    );
    return response.data;
  }

  /**
   * Get a single OAuth client by ID
   * @param id Client ID
   * @returns Client details
   */
  async get(id: string): Promise<OAuthClient> {
    const response = await this.http.get<OAuthClient>(`/api/admin/oauth/clients/${id}`);
    return response.data;
  }

  /**
   * Update an OAuth client
   * @param id Client ID
   * @param data Update data
   * @returns Updated client
   */
  async update(id: string, data: UpdateOAuthClientRequest): Promise<OAuthClient> {
    const response = await this.http.put<OAuthClient>(
      `/api/admin/oauth/clients/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete an OAuth client (soft delete)
   * @param id Client ID
   */
  async delete(id: string): Promise<void> {
    await this.http.delete(`/api/admin/oauth/clients/${id}`);
  }

  /**
   * Rotate client secret - generates a new secret
   * @param id Client ID
   * @returns New client_secret (shown only once)
   */
  async rotateSecret(id: string): Promise<RotateSecretResponse> {
    const response = await this.http.post<RotateSecretResponse>(
      `/api/admin/oauth/clients/${id}/rotate-secret`
    );
    return response.data;
  }

  /**
   * Activate an OAuth client
   * @param id Client ID
   * @returns Updated client
   */
  async activate(id: string): Promise<OAuthClient> {
    return this.update(id, { is_active: true });
  }

  /**
   * Deactivate an OAuth client
   * @param id Client ID
   * @returns Updated client
   */
  async deactivate(id: string): Promise<OAuthClient> {
    return this.update(id, { is_active: false });
  }

  // ============= Scope Management =============

  /**
   * List all OAuth scopes
   * @returns List of available scopes
   */
  async listScopes(): Promise<OAuthScopeListResponse> {
    const response = await this.http.get<OAuthScopeListResponse>('/api/admin/oauth/scopes');
    return response.data;
  }

  /**
   * Create a custom OAuth scope
   * @param data Scope creation data
   * @returns Created scope
   */
  async createScope(data: CreateScopeRequest): Promise<OAuthScope> {
    const response = await this.http.post<OAuthScope>('/api/admin/oauth/scopes', data);
    return response.data;
  }

  /**
   * Delete a non-system OAuth scope
   * @param id Scope ID
   */
  async deleteScope(id: string): Promise<void> {
    await this.http.delete(`/api/admin/oauth/scopes/${id}`);
  }

  // ============= Consent Management =============

  /**
   * List all user consents for a client
   * @param clientId Client ID
   * @returns List of user consents
   */
  async listClientConsents(clientId: string): Promise<UserConsentListResponse> {
    const response = await this.http.get<UserConsentListResponse>(
      `/api/admin/oauth/clients/${clientId}/consents`
    );
    return response.data;
  }

  /**
   * Revoke a user's consent for a client
   * @param clientId Client ID
   * @param userId User ID
   */
  async revokeUserConsent(clientId: string, userId: string): Promise<void> {
    await this.http.delete(`/api/admin/oauth/clients/${clientId}/consents/${userId}`);
  }
}
