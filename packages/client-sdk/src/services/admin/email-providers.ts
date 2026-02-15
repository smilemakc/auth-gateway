/**
 * Admin Email Providers service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  EmailProvider,
  CreateEmailProviderRequest,
  UpdateEmailProviderRequest,
} from '../../types/email-provider';
import { BaseService } from '../base';

/** Admin Email Providers service for email provider management */
export class AdminEmailProvidersService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all email providers
   * @returns List of email providers
   */
  async list(): Promise<EmailProvider[]> {
    const response = await this.http.get<{ providers: EmailProvider[]; total: number }>(
      '/api/admin/email-providers'
    );
    return response.data.providers;
  }

  /**
   * Get email provider by ID
   * @param id Email provider ID
   * @returns Email provider details
   */
  async get(id: string): Promise<EmailProvider> {
    const response = await this.http.get<EmailProvider>(
      `/api/admin/email-providers/${id}`
    );
    return response.data;
  }

  /**
   * Create a new email provider
   * @param data Email provider creation data
   * @returns Created email provider
   */
  async create(data: CreateEmailProviderRequest): Promise<EmailProvider> {
    const response = await this.http.post<EmailProvider>(
      '/api/admin/email-providers',
      data
    );
    return response.data;
  }

  /**
   * Update email provider
   * @param id Email provider ID
   * @param data Email provider update data
   * @returns Updated email provider
   */
  async update(id: string, data: UpdateEmailProviderRequest): Promise<EmailProvider> {
    const response = await this.http.put<EmailProvider>(
      `/api/admin/email-providers/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete email provider
   * @param id Email provider ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/api/admin/email-providers/${id}`
    );
    return response.data;
  }

  /**
   * Test email provider
   * @param id Email provider ID
   * @param email Test email address
   * @returns Success message
   */
  async test(id: string, email: string): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/api/admin/email-providers/${id}/test`,
      { email }
    );
    return response.data;
  }
}
