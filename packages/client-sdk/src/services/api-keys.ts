/**
 * API Keys service
 */

import type { HttpClient } from '../core/http';
import type { MessageResponse } from '../types/common';
import type {
  APIKey,
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
  ListAPIKeysResponse,
  UpdateAPIKeyRequest,
} from '../types/api-key';
import { BaseService } from './base';

/** API Keys service for managing API keys */
export class APIKeysService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Create a new API key
   * @param data API key creation data
   * @returns Created API key with plain key (only shown once!)
   */
  async create(data: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> {
    const response = await this.http.post<CreateAPIKeyResponse>(
      '/api/api-keys',
      data
    );
    return response.data;
  }

  /**
   * List all API keys for the current user
   * @returns List of API keys (without plain key values)
   */
  async list(): Promise<ListAPIKeysResponse> {
    const response = await this.http.get<ListAPIKeysResponse>('/api/api-keys');
    return response.data;
  }

  /**
   * Get a specific API key by ID
   * @param id API key ID
   * @returns API key details
   */
  async get(id: string): Promise<APIKey> {
    const response = await this.http.get<APIKey>(`/api-keys/${id}`);
    return response.data;
  }

  /**
   * Update an API key
   * @param id API key ID
   * @param data Update data
   * @returns Updated API key
   */
  async update(id: string, data: UpdateAPIKeyRequest): Promise<APIKey> {
    const response = await this.http.put<APIKey>(`/api-keys/${id}`, data);
    return response.data;
  }

  /**
   * Revoke an API key (deactivate but keep record)
   * @param id API key ID
   * @returns Success message
   */
  async revoke(id: string): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/api-keys/${id}/revoke`
    );
    return response.data;
  }

  /**
   * Delete an API key permanently
   * @param id API key ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(`/api-keys/${id}`);
    return response.data;
  }

  /**
   * Get all active API keys
   * @returns List of active API keys
   */
  async getActive(): Promise<APIKey[]> {
    const { api_keys } = await this.list();
    return api_keys.filter((key: APIKey) => key.is_active);
  }
}
