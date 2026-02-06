/**
 * Admin SMS Settings service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type { SMSSettings, SMSSettingsRequest } from '../../types/sms';
import { BaseService } from '../base';

/** Admin SMS Settings service for SMS provider configuration */
export class AdminSMSSettingsService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all SMS settings
   * @returns List of SMS provider configurations
   */
  async list(): Promise<SMSSettings[]> {
    const response = await this.http.get<SMSSettings[]>('/api/admin/sms/settings');
    return response.data;
  }

  /**
   * Get a specific SMS setting by ID
   * @param id SMS setting ID
   * @returns SMS setting details
   */
  async get(id: string): Promise<SMSSettings> {
    const response = await this.http.get<SMSSettings>(
      `/admin/sms/settings/${id}`
    );
    return response.data;
  }

  /**
   * Get the active SMS setting
   * @returns Active SMS provider configuration
   */
  async getActive(): Promise<SMSSettings> {
    const response = await this.http.get<SMSSettings>(
      '/api/admin/sms/settings/active'
    );
    return response.data;
  }

  /**
   * Create a new SMS setting
   * @param data SMS setting data
   * @returns Created SMS setting
   */
  async create(data: SMSSettingsRequest): Promise<SMSSettings> {
    const response = await this.http.post<SMSSettings>(
      '/api/admin/sms/settings',
      data
    );
    return response.data;
  }

  /**
   * Update an SMS setting
   * @param id SMS setting ID
   * @param data Update data
   * @returns Updated SMS setting
   */
  async update(id: string, data: Partial<SMSSettingsRequest>): Promise<SMSSettings> {
    const response = await this.http.put<SMSSettings>(
      `/admin/sms/settings/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete an SMS setting
   * @param id SMS setting ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/admin/sms/settings/${id}`
    );
    return response.data;
  }

  /**
   * Activate an SMS setting
   * @param id SMS setting ID
   * @returns Updated SMS setting
   */
  async activate(id: string): Promise<SMSSettings> {
    // First deactivate all others
    const allSettings = await this.list();
    for (const setting of allSettings) {
      if (setting.id !== id && setting.is_active) {
        await this.update(setting.id, { is_active: false });
      }
    }

    // Activate the target setting
    return this.update(id, { is_active: true });
  }

  /**
   * Deactivate an SMS setting
   * @param id SMS setting ID
   * @returns Updated SMS setting
   */
  async deactivate(id: string): Promise<SMSSettings> {
    return this.update(id, { is_active: false });
  }
}
