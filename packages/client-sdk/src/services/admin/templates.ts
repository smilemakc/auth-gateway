/**
 * Admin Email Templates service
 */

import type { HttpClient } from '../../core/http';
import type { MessageResponse } from '../../types/common';
import type {
  EmailTemplate,
  EmailTemplateType,
  EmailTemplateListResponse,
  CreateEmailTemplateRequest,
  UpdateEmailTemplateRequest,
  PreviewEmailTemplateRequest,
  PreviewEmailTemplateResponse,
  EmailTemplateTypesResponse,
  EmailTemplateVariablesResponse,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin Email Templates service for template management */
export class AdminTemplatesService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all email templates
   * @returns List of email templates
   */
  async list(): Promise<EmailTemplateListResponse> {
    const response = await this.http.get<EmailTemplateListResponse>(
      '/admin/templates'
    );
    return response.data;
  }

  /**
   * Get template by ID
   * @param id Template ID
   * @returns Template details
   */
  async get(id: string): Promise<EmailTemplate> {
    const response = await this.http.get<EmailTemplate>(
      `/admin/templates/${id}`
    );
    return response.data;
  }

  /**
   * Create a new email template
   * @param data Template creation data
   * @returns Created template
   */
  async create(data: CreateEmailTemplateRequest): Promise<EmailTemplate> {
    const response = await this.http.post<EmailTemplate>(
      '/admin/templates',
      data
    );
    return response.data;
  }

  /**
   * Update email template
   * @param id Template ID
   * @param data Template update data
   * @returns Updated template
   */
  async update(
    id: string,
    data: UpdateEmailTemplateRequest
  ): Promise<EmailTemplate> {
    const response = await this.http.put<EmailTemplate>(
      `/admin/templates/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete email template
   * @param id Template ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/admin/templates/${id}`
    );
    return response.data;
  }

  /**
   * Preview email template with variables
   * @param data Preview request with HTML body and variables
   * @returns Rendered HTML and text
   */
  async preview(
    data: PreviewEmailTemplateRequest
  ): Promise<PreviewEmailTemplateResponse> {
    const response = await this.http.post<PreviewEmailTemplateResponse>(
      '/admin/templates/preview',
      data
    );
    return response.data;
  }

  /**
   * Get available template types
   * @returns List of available template types
   */
  async getTypes(): Promise<EmailTemplateType[]> {
    const response = await this.http.get<EmailTemplateTypesResponse>(
      '/admin/templates/types'
    );
    return response.data.types;
  }

  /**
   * Get default variables for a template type
   * @param type Template type
   * @returns List of available variables
   */
  async getVariables(type: EmailTemplateType): Promise<string[]> {
    const response = await this.http.get<EmailTemplateVariablesResponse>(
      `/admin/templates/variables/${type}`
    );
    return response.data.variables;
  }

  /**
   * Get templates by type
   * @param type Template type
   * @returns List of templates of the specified type
   */
  async getByType(type: EmailTemplateType): Promise<EmailTemplate[]> {
    const templates = await this.list();
    return templates.templates.filter((t) => t.type === type);
  }

  /**
   * Enable template
   * @param id Template ID
   * @returns Updated template
   */
  async enable(id: string): Promise<EmailTemplate> {
    return this.update(id, { is_active: true });
  }

  /**
   * Disable template
   * @param id Template ID
   * @returns Updated template
   */
  async disable(id: string): Promise<EmailTemplate> {
    return this.update(id, { is_active: false });
  }
}
