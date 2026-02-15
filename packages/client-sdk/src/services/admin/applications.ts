/**
 * Admin Applications service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  Application,
  ApplicationListResponse,
  CreateApplicationRequest,
  CreateApplicationResponse,
  UpdateApplicationRequest,
  AppRotateSecretResponse,
  ApplicationBranding,
  UpdateBrandingRequest,
  UserAppProfileListResponse,
  BanUserRequest,
  AuthConfigResponse,
  EmailTemplateListResponse,
  EmailTemplate,
  CreateEmailTemplateRequest,
  UpdateEmailTemplateRequest,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin Applications service for application management */
export class AdminApplicationsService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Create a new application
   * @param data Application creation data
   * @returns Created application with secret
   */
  async create(data: CreateApplicationRequest): Promise<CreateApplicationResponse> {
    const response = await this.http.post<CreateApplicationResponse>(
      '/api/admin/applications',
      data
    );
    return response.data;
  }

  /**
   * List all applications
   * @param page Page number (default 1)
   * @param perPage Items per page (default 20)
   * @param isActive Filter by active status (optional)
   * @returns Paginated list of applications
   */
  async list(page = 1, perPage = 20, isActive?: boolean): Promise<ApplicationListResponse> {
    const query: Record<string, number | boolean> = { page, page_size: perPage };
    if (isActive !== undefined) {
      query.is_active = isActive;
    }
    const response = await this.http.get<ApplicationListResponse>(
      '/api/admin/applications',
      { query }
    );
    return response.data;
  }

  /**
   * Get application by ID
   * @param id Application ID
   * @returns Application details
   */
  async getById(id: string): Promise<Application> {
    const response = await this.http.get<Application>(`/api/admin/applications/${id}`);
    return response.data;
  }

  /**
   * Update application
   * @param id Application ID
   * @param data Application update data
   * @returns Updated application
   */
  async update(id: string, data: UpdateApplicationRequest): Promise<Application> {
    const response = await this.http.put<Application>(
      `/api/admin/applications/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete application
   * @param id Application ID
   * @returns Success message
   */
  async remove(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/api/admin/applications/${id}`
    );
    return response.data;
  }

  /**
   * Rotate application secret
   * @param id Application ID
   * @returns New secret with warning
   */
  async rotateSecret(id: string): Promise<AppRotateSecretResponse> {
    const response = await this.http.post<AppRotateSecretResponse>(
      `/api/admin/applications/${id}/rotate-secret`,
      {}
    );
    return response.data;
  }

  /**
   * Get application branding
   * @param appId Application ID
   * @returns Application branding settings
   */
  async getBranding(appId: string): Promise<ApplicationBranding> {
    const response = await this.http.get<ApplicationBranding>(
      `/api/admin/applications/${appId}/branding`
    );
    return response.data;
  }

  /**
   * Update application branding
   * @param appId Application ID
   * @param data Branding update data
   * @returns Updated branding settings
   */
  async updateBranding(appId: string, data: UpdateBrandingRequest): Promise<ApplicationBranding> {
    const response = await this.http.put<ApplicationBranding>(
      `/api/admin/applications/${appId}/branding`,
      data
    );
    return response.data;
  }

  /**
   * List users of an application
   * @param appId Application ID
   * @param page Page number (default 1)
   * @param perPage Items per page (default 20)
   * @returns Paginated list of user app profiles
   */
  async listUsers(appId: string, page = 1, perPage = 20): Promise<UserAppProfileListResponse> {
    const response = await this.http.get<UserAppProfileListResponse>(
      `/api/admin/applications/${appId}/users`,
      { query: { page, page_size: perPage } }
    );
    return response.data;
  }

  /**
   * Ban a user from an application
   * @param appId Application ID
   * @param userId User ID
   * @param data Ban request with reason
   * @returns Success message
   */
  async banUser(appId: string, userId: string, data: BanUserRequest): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/api/admin/applications/${appId}/users/${userId}/ban`,
      data
    );
    return response.data;
  }

  /**
   * Unban a user from an application
   * @param appId Application ID
   * @param userId User ID
   * @returns Success message
   */
  async unbanUser(appId: string, userId: string): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/api/admin/applications/${appId}/users/${userId}/unban`,
      {}
    );
    return response.data;
  }

  /**
   * Get public auth configuration for an application
   * @param appId Application ID
   * @returns Auth configuration with allowed methods and branding
   */
  async getAuthConfig(appId: string): Promise<AuthConfigResponse> {
    const response = await this.http.get<AuthConfigResponse>(
      `/api/applications/${appId}/auth-config`
    );
    return response.data;
  }

  /**
   * List email templates for an application
   * @param appId Application ID
   * @returns Email templates list
   */
  async listTemplates(appId: string): Promise<EmailTemplateListResponse> {
    const response = await this.http.get<EmailTemplateListResponse>(
      `/api/admin/applications/${appId}/email-templates`
    );
    return response.data;
  }

  /**
   * Get email template by ID
   * @param appId Application ID
   * @param templateId Template ID
   * @returns Email template details
   */
  async getTemplate(appId: string, templateId: string): Promise<EmailTemplate> {
    const response = await this.http.get<EmailTemplate>(
      `/api/admin/applications/${appId}/email-templates/${templateId}`
    );
    return response.data;
  }

  /**
   * Create email template for an application
   * @param appId Application ID
   * @param data Template creation data
   * @returns Created email template
   */
  async createTemplate(appId: string, data: CreateEmailTemplateRequest): Promise<EmailTemplate> {
    const response = await this.http.post<EmailTemplate>(
      `/api/admin/applications/${appId}/email-templates`,
      data
    );
    return response.data;
  }

  /**
   * Update email template
   * @param appId Application ID
   * @param templateId Template ID
   * @param data Template update data
   * @returns Updated email template
   */
  async updateTemplate(appId: string, templateId: string, data: UpdateEmailTemplateRequest): Promise<EmailTemplate> {
    const response = await this.http.put<EmailTemplate>(
      `/api/admin/applications/${appId}/email-templates/${templateId}`,
      data
    );
    return response.data;
  }

  /**
   * Delete email template
   * @param appId Application ID
   * @param templateId Template ID
   * @returns Success message
   */
  async deleteTemplate(appId: string, templateId: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/api/admin/applications/${appId}/email-templates/${templateId}`
    );
    return response.data;
  }

  /**
   * Initialize default email templates for an application
   * @param appId Application ID
   * @returns Success message
   */
  async initializeTemplates(appId: string): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      `/api/admin/applications/${appId}/email-templates/initialize`,
      {}
    );
    return response.data;
  }

  /**
   * Enable email template
   * @param appId Application ID
   * @param templateId Template ID
   * @returns Updated email template
   */
  async enableTemplate(appId: string, templateId: string): Promise<EmailTemplate> {
    const response = await this.http.post<EmailTemplate>(
      `/api/admin/applications/${appId}/email-templates/${templateId}/enable`,
      {}
    );
    return response.data;
  }

  /**
   * Disable email template
   * @param appId Application ID
   * @param templateId Template ID
   * @returns Updated email template
   */
  async disableTemplate(appId: string, templateId: string): Promise<EmailTemplate> {
    const response = await this.http.post<EmailTemplate>(
      `/api/admin/applications/${appId}/email-templates/${templateId}/disable`,
      {}
    );
    return response.data;
  }

  /**
   * Import users from file or data
   * @param data Import data
   * @returns Import result
   */
  async importUsers(data: any): Promise<any> {
    const response = await this.http.post<any>(
      '/api/admin/users/import',
      data
    );
    return response.data;
  }

  /**
   * Get public branding for an application (no auth required)
   * @param appId Application ID
   * @returns Application branding settings
   */
  async getPublicBranding(appId: string): Promise<ApplicationBranding> {
    const response = await this.http.get<ApplicationBranding>(
      `/api/applications/${appId}/branding`,
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Get current user's application profile
   * @param appId Optional application ID (uses X-Application-ID header if not provided)
   * @returns User application profile
   */
  async getMyApplicationProfile(appId?: string): Promise<any> {
    const query = appId ? { application_id: appId } : undefined;
    const response = await this.http.get<any>(
      '/api/user/application-profile',
      { query }
    );
    return response.data;
  }

  /**
   * Update current user's application profile
   * @param data Profile update data
   * @param appId Optional application ID (uses X-Application-ID header if not provided)
   * @returns Updated user application profile
   */
  async updateMyApplicationProfile(data: any, appId?: string): Promise<any> {
    const query = appId ? { application_id: appId } : undefined;
    const response = await this.http.put<any>(
      '/api/user/application-profile',
      data,
      { query }
    );
    return response.data;
  }
}
