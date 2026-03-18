/**
 * Admin System service
 */

import type { HttpClient } from '../../core/http';
import type {
  GeoDistributionResponse,
  MaintenanceModeResponse,
  SystemHealthResponse,
  UpdateMaintenanceModeRequest,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin System service for system management */
export class AdminSystemService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Get system health status
   * @returns Detailed system health information
   */
  async getHealth(): Promise<SystemHealthResponse> {
    const response = await this.http.get<SystemHealthResponse>(
      '/api/admin/system/health'
    );
    return response.data;
  }

  /**
   * Get maintenance mode status
   * @returns Maintenance mode status
   */
  async getMaintenanceMode(): Promise<MaintenanceModeResponse> {
    const response = await this.http.get<MaintenanceModeResponse>(
      '/system/maintenance',
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Update maintenance mode
   * @param data Maintenance mode configuration
   * @returns Updated maintenance mode status
   */
  async setMaintenanceMode(
    data: UpdateMaintenanceModeRequest
  ): Promise<MaintenanceModeResponse> {
    const response = await this.http.put<MaintenanceModeResponse>(
      '/api/admin/system/maintenance',
      data
    );
    return response.data;
  }

  /**
   * Enable maintenance mode
   * @param message Optional maintenance message
   * @returns Updated status
   */
  async enableMaintenanceMode(
    message?: string
  ): Promise<MaintenanceModeResponse> {
    return this.setMaintenanceMode({ enabled: true, message });
  }

  /**
   * Disable maintenance mode
   * @returns Updated status
   */
  async disableMaintenanceMode(): Promise<MaintenanceModeResponse> {
    return this.setMaintenanceMode({ enabled: false });
  }

  /**
   * Get geo distribution analytics
   * @param days Number of days to analyze
   * @returns Geo distribution of logins
   */
  async getGeoDistribution(days = 30): Promise<GeoDistributionResponse> {
    const response = await this.http.get<GeoDistributionResponse>(
      '/api/admin/analytics/geo-distribution',
      { query: { days } }
    );
    return response.data;
  }

  /**
   * Check if system is healthy
   * @returns True if system is healthy
   */
  async isHealthy(): Promise<boolean> {
    try {
      const health = await this.getHealth();
      return health.status === 'healthy';
    } catch {
      return false;
    }
  }

  /**
   * Get top countries by login count
   * @param limit Number of countries to return
   * @param days Number of days to analyze
   * @returns Top countries
   */
  async getTopCountries(
    limit = 10,
    days = 30
  ): Promise<
    Array<{ country_code: string; country_name: string; login_count: number }>
  > {
    const geo = await this.getGeoDistribution(days);

    // Aggregate by country
    const countryMap = new Map<
      string,
      { country_code: string; country_name: string; login_count: number }
    >();

    for (const loc of geo.locations) {
      const existing = countryMap.get(loc.country_code);
      if (existing) {
        existing.login_count += loc.login_count;
      } else {
        countryMap.set(loc.country_code, {
          country_code: loc.country_code,
          country_name: loc.country_name,
          login_count: loc.login_count,
        });
      }
    }

    return Array.from(countryMap.values())
      .sort((a, b) => b.login_count - a.login_count)
      .slice(0, limit);
  }

  /**
   * Get password policy configuration
   * @returns Password policy settings
   */
  async getPasswordPolicy(): Promise<any> {
    const response = await this.http.get<any>('/api/admin/system/password-policy');
    return response.data;
  }

  /**
   * Update password policy configuration
   * @param policy Password policy settings
   * @returns Updated password policy
   */
  async updatePasswordPolicy(policy: any): Promise<any> {
    const response = await this.http.put<any>('/api/admin/system/password-policy', policy);
    return response.data;
  }
}
