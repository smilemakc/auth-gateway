/**
 * Admin Branding service
 */

import type { HttpClient } from '../../core/http';
import type {
  BrandingSettings,
  PublicBrandingResponse,
  UpdateBrandingRequest,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin Branding service for customization management */
export class AdminBrandingService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Get public branding settings (no auth required)
   * @returns Public branding configuration
   */
  async getPublic(): Promise<PublicBrandingResponse> {
    const response = await this.http.get<PublicBrandingResponse>('/branding', {
      skipAuth: true,
    });
    return response.data;
  }

  /**
   * Get full branding settings (admin only)
   * @returns Complete branding configuration
   */
  async get(): Promise<BrandingSettings> {
    const response = await this.http.get<BrandingSettings>('/admin/branding');
    return response.data;
  }

  /**
   * Update branding settings
   * @param data Branding update data
   * @returns Updated branding settings
   */
  async update(data: UpdateBrandingRequest): Promise<BrandingSettings> {
    const response = await this.http.put<BrandingSettings>(
      '/admin/branding',
      data
    );
    return response.data;
  }

  /**
   * Update logo URL
   * @param logoUrl New logo URL
   * @returns Updated branding settings
   */
  async updateLogo(logoUrl: string): Promise<BrandingSettings> {
    return this.update({ logoUrl });
  }

  /**
   * Update favicon URL
   * @param faviconUrl New favicon URL
   * @returns Updated branding settings
   */
  async updateFavicon(faviconUrl: string): Promise<BrandingSettings> {
    return this.update({ faviconUrl });
  }

  /**
   * Update theme colors
   * @param theme Theme color configuration
   * @returns Updated branding settings
   */
  async updateTheme(theme: {
    primaryColor?: string;
    secondaryColor?: string;
    backgroundColor?: string;
  }): Promise<BrandingSettings> {
    return this.update(theme);
  }

  /**
   * Update company information
   * @param info Company information
   * @returns Updated branding settings
   */
  async updateCompanyInfo(info: {
    companyName?: string;
    supportEmail?: string;
    termsUrl?: string;
    privacyUrl?: string;
  }): Promise<BrandingSettings> {
    return this.update(info);
  }
}
