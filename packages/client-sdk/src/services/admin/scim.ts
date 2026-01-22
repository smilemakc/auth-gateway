/**
 * Admin SCIM service
 */

import type { HttpClient } from '../../core/http';
import type { SCIMConfig, SCIMMetadataResponse } from '../../types/admin';
import { BaseService } from '../base';

/** Admin SCIM service for SCIM configuration */
export class AdminSCIMService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Get SCIM configuration
   * @returns SCIM configuration
   */
  async getConfig(): Promise<SCIMConfig> {
    const response = await this.http.get<SCIMConfig>('/admin/scim/config');
    return response.data;
  }

  /**
   * Get SCIM metadata
   * @returns SCIM metadata
   */
  async getMetadata(): Promise<SCIMMetadataResponse> {
    const response = await this.http.get<SCIMMetadataResponse>('/scim/v2/ServiceProviderConfig');
    return response.data;
  }
}

