/**
 * Admin Webhooks service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  Webhook,
  WebhookListResponse,
  CreateWebhookRequest,
  CreateWebhookResponse,
  UpdateWebhookRequest,
  TestWebhookRequest,
  WebhookDeliveryListResponse,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin Webhooks service for webhook management */
export class AdminWebhooksService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all webhooks
   * @param page Page number (default 1)
   * @param perPage Items per page (default 20)
   * @returns Paginated list of webhooks
   */
  async list(page = 1, perPage = 20): Promise<WebhookListResponse> {
    const response = await this.http.get<WebhookListResponse>(
      '/admin/webhooks',
      { query: { page, per_page: perPage } }
    );
    return response.data;
  }

  /**
   * Get webhook by ID
   * @param id Webhook ID
   * @returns Webhook details
   */
  async get(id: string): Promise<Webhook> {
    const response = await this.http.get<Webhook>(`/admin/webhooks/${id}`);
    return response.data;
  }

  /**
   * Create a new webhook
   * @param data Webhook creation data
   * @returns Created webhook with secret key
   */
  async create(data: CreateWebhookRequest): Promise<CreateWebhookResponse> {
    const response = await this.http.post<CreateWebhookResponse>(
      '/admin/webhooks',
      data
    );
    return response.data;
  }

  /**
   * Update webhook
   * @param id Webhook ID
   * @param data Webhook update data
   * @returns Updated webhook
   */
  async update(id: string, data: UpdateWebhookRequest): Promise<Webhook> {
    const response = await this.http.put<Webhook>(
      `/admin/webhooks/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete webhook
   * @param id Webhook ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/admin/webhooks/${id}`
    );
    return response.data;
  }

  /**
   * Test webhook with a specific event
   * @param id Webhook ID
   * @param data Test data with event type
   * @returns Success message
   */
  async test(id: string, data: TestWebhookRequest): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/admin/webhooks/${id}/test`,
      data
    );
    return response.data;
  }

  /**
   * Get webhook deliveries
   * @param id Webhook ID
   * @param page Page number (default 1)
   * @param perPage Items per page (default 20)
   * @returns Paginated list of deliveries
   */
  async getDeliveries(
    id: string,
    page = 1,
    perPage = 20
  ): Promise<WebhookDeliveryListResponse> {
    const response = await this.http.get<WebhookDeliveryListResponse>(
      `/admin/webhooks/${id}/deliveries`,
      { query: { page, per_page: perPage } }
    );
    return response.data;
  }

  /**
   * Get available webhook events
   * @returns List of available event types
   */
  async getAvailableEvents(): Promise<string[]> {
    const response = await this.http.get<{ events: string[] }>(
      '/admin/webhooks/events'
    );
    return response.data.events;
  }

  /**
   * Enable webhook
   * @param id Webhook ID
   * @returns Updated webhook
   */
  async enable(id: string): Promise<Webhook> {
    return this.update(id, { is_active: true });
  }

  /**
   * Disable webhook
   * @param id Webhook ID
   * @returns Updated webhook
   */
  async disable(id: string): Promise<Webhook> {
    return this.update(id, { is_active: false });
  }
}
