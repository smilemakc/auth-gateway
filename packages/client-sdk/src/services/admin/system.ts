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
      '/admin/system/health'
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
      '/admin/system/maintenance',
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
      '/admin/analytics/geo-distribution',
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
    Array<{ countryCode: string; countryName: string; loginCount: number }>
  > {
    const geo = await this.getGeoDistribution(days);

    // Aggregate by country
    const countryMap = new Map<
      string,
      { countryCode: string; countryName: string; loginCount: number }
    >();

    for (const loc of geo.locations) {
      const existing = countryMap.get(loc.countryCode);
      if (existing) {
        existing.loginCount += loc.loginCount;
      } else {
        countryMap.set(loc.countryCode, {
          countryCode: loc.countryCode,
          countryName: loc.countryName,
          loginCount: loc.loginCount,
        });
      }
    }

    return Array.from(countryMap.values())
      .sort((a, b) => b.loginCount - a.loginCount)
      .slice(0, limit);
  }
}
