/**
 * Admin Email sending service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import { BaseService } from '../base';

/** Request to send an email via template */
export interface SendEmailRequest {
  template_type: string;
  to_email: string;
  variables?: Record<string, string>;
  profile_id?: string;
  application_id?: string;
}

/** Admin Email service for sending emails via templates */
export class AdminEmailService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Send an email using a specified template
   * @param data Email sending request
   * @returns Success message
   */
  async send(data: SendEmailRequest): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      '/api/admin/email/send',
      data
    );
    return response.data;
  }
}
