/**
 * Admin Email Profiles service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  EmailProfile,
  CreateEmailProfileRequest,
  UpdateEmailProfileRequest,
} from '../../types/email-provider';
import { BaseService } from '../base';

/** Admin Email Profiles service for email profile management */
export class AdminEmailProfilesService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all email profiles
   * @returns List of email profiles
   */
  async list(): Promise<EmailProfile[]> {
    const response = await this.http.get<{ profiles: EmailProfile[]; total: number }>(
      '/api/admin/email-profiles'
    );
    return response.data.profiles;
  }

  /**
   * Get email profile by ID
   * @param id Email profile ID
   * @returns Email profile details
   */
  async get(id: string): Promise<EmailProfile> {
    const response = await this.http.get<EmailProfile>(
      `/api/admin/email-profiles/${id}`
    );
    return response.data;
  }

  /**
   * Create a new email profile
   * @param data Email profile creation data
   * @returns Created email profile
   */
  async create(data: CreateEmailProfileRequest): Promise<EmailProfile> {
    const response = await this.http.post<EmailProfile>(
      '/api/admin/email-profiles',
      data
    );
    return response.data;
  }

  /**
   * Update email profile
   * @param id Email profile ID
   * @param data Email profile update data
   * @returns Updated email profile
   */
  async update(id: string, data: UpdateEmailProfileRequest): Promise<EmailProfile> {
    const response = await this.http.put<EmailProfile>(
      `/api/admin/email-profiles/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete email profile
   * @param id Email profile ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/api/admin/email-profiles/${id}`
    );
    return response.data;
  }

  /**
   * Set email profile as default
   * @param id Email profile ID
   * @returns Success message
   */
  async setDefault(id: string): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/api/admin/email-profiles/${id}/set-default`,
      {}
    );
    return response.data;
  }

  /**
   * Test email profile
   * @param id Email profile ID
   * @param email Test email address
   * @returns Success message
   */
  async test(id: string, email: string): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/api/admin/email-profiles/${id}/test`,
      { email }
    );
    return response.data;
  }
}
