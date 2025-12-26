/**
 * Admin LDAP service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  LDAPConfig,
  CreateLDAPConfigRequest,
  UpdateLDAPConfigRequest,
  LDAPTestConnectionRequest,
  LDAPTestConnectionResponse,
  LDAPSyncRequest,
  LDAPSyncResponse,
  LDAPConfigListResponse,
  LDAPSyncLogsResponse,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin LDAP service for LDAP configuration management */
export class AdminLDAPService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all LDAP configurations
   * @returns List of LDAP configurations
   */
  async listConfigs(): Promise<LDAPConfigListResponse> {
    const response = await this.http.get<LDAPConfigListResponse>('/admin/ldap/configs');
    return response.data;
  }

  /**
   * Get active LDAP configuration
   * @returns Active LDAP configuration
   */
  async getActiveConfig(): Promise<LDAPConfig> {
    const response = await this.http.get<LDAPConfig>('/admin/ldap/config');
    return response.data;
  }

  /**
   * Get a specific LDAP configuration by ID
   * @param id Configuration ID
   * @returns LDAP configuration details
   */
  async getConfig(id: string): Promise<LDAPConfig> {
    const response = await this.http.get<LDAPConfig>(`/admin/ldap/config/${id}`);
    return response.data;
  }

  /**
   * Create a new LDAP configuration
   * @param data Configuration data
   * @returns Created configuration
   */
  async createConfig(data: CreateLDAPConfigRequest): Promise<LDAPConfig> {
    const response = await this.http.post<LDAPConfig>('/admin/ldap/config', data);
    return response.data;
  }

  /**
   * Update an LDAP configuration
   * @param id Configuration ID
   * @param data Update data
   * @returns Updated configuration
   */
  async updateConfig(id: string, data: UpdateLDAPConfigRequest): Promise<LDAPConfig> {
    const response = await this.http.put<LDAPConfig>(`/admin/ldap/config/${id}`, data);
    return response.data;
  }

  /**
   * Delete an LDAP configuration
   * @param id Configuration ID
   * @returns Success message
   */
  async deleteConfig(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(`/admin/ldap/config/${id}`);
    return response.data;
  }

  /**
   * Test LDAP connection
   * @param data Connection test data
   * @returns Test result
   */
  async testConnection(data: LDAPTestConnectionRequest): Promise<LDAPTestConnectionResponse> {
    const response = await this.http.post<LDAPTestConnectionResponse>('/admin/ldap/test-connection', data);
    return response.data;
  }

  /**
   * Trigger LDAP synchronization
   * @param id Configuration ID
   * @param data Sync options
   * @returns Sync result
   */
  async sync(id: string, data?: LDAPSyncRequest): Promise<LDAPSyncResponse> {
    const response = await this.http.post<LDAPSyncResponse>(`/admin/ldap/config/${id}/sync`, data || {});
    return response.data;
  }

  /**
   * Get LDAP sync logs
   * @param id Configuration ID
   * @returns Sync logs
   */
  async getSyncLogs(id: string): Promise<LDAPSyncLogsResponse> {
    const response = await this.http.get<LDAPSyncLogsResponse>(`/admin/ldap/config/${id}/sync-logs`);
    return response.data;
  }
}

