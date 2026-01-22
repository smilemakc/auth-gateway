/**
 * Admin SAML service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  SAMLServiceProvider,
  CreateSAMLSPRequest,
  UpdateSAMLSPRequest,
  SAMLSPListResponse,
  SAMLMetadataResponse,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin SAML service for SAML Service Provider management */
export class AdminSAMLService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all SAML Service Providers
   * @param page Page number
   * @param pageSize Items per page
   * @returns Paginated list of SAML SPs
   */
  async listSPs(page = 1, pageSize = 20): Promise<SAMLSPListResponse> {
    const response = await this.http.get<SAMLSPListResponse>('/admin/saml/sp', {
      headers: {},
      query: { page, page_size: pageSize },
    });
    return response.data;
  }

  /**
   * Get a specific SAML Service Provider by ID
   * @param id SP ID
   * @returns SAML SP details
   */
  async getSP(id: string): Promise<SAMLServiceProvider> {
    const response = await this.http.get<SAMLServiceProvider>(`/admin/saml/sp/${id}`);
    return response.data;
  }

  /**
   * Create a new SAML Service Provider
   * @param data SP data
   * @returns Created SP
   */
  async createSP(data: CreateSAMLSPRequest): Promise<SAMLServiceProvider> {
    const response = await this.http.post<SAMLServiceProvider>('/admin/saml/sp', data);
    return response.data;
  }

  /**
   * Update a SAML Service Provider
   * @param id SP ID
   * @param data Update data
   * @returns Updated SP
   */
  async updateSP(id: string, data: UpdateSAMLSPRequest): Promise<SAMLServiceProvider> {
    const response = await this.http.put<SAMLServiceProvider>(`/admin/saml/sp/${id}`, data);
    return response.data;
  }

  /**
   * Delete a SAML Service Provider
   * @param id SP ID
   * @returns Success message
   */
  async deleteSP(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(`/admin/saml/sp/${id}`);
    return response.data;
  }

  /**
   * Get SAML IdP metadata
   * @returns SAML metadata XML
   */
  async getMetadata(): Promise<SAMLMetadataResponse> {
    const response = await this.http.get<SAMLMetadataResponse>('/saml/metadata', {
      headers: {
        Accept: 'application/xml',
      },
    });
    return response.data;
  }
}

