/**
 * Health check service
 */

import type { HttpClient } from '../core/http';
import type { HealthResponse, MaintenanceModeResponse } from '../types/admin';
import { BaseService } from './base';

/** Status response for readiness and liveness */
interface StatusResponse {
  status: string;
}

/** Health check service */
export class HealthService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Check overall health status
   * @returns Health status of all services
   */
  async check(): Promise<HealthResponse> {
    const response = await this.http.get<HealthResponse>('/health', {
      skipAuth: true,
    });
    return response.data;
  }

  /**
   * Check readiness status
   * @returns Readiness status
   */
  async ready(): Promise<boolean> {
    try {
      const response = await this.http.get<StatusResponse>('/ready', {
        skipAuth: true,
      });
      return response.data.status === 'ready';
    } catch {
      return false;
    }
  }

  /**
   * Check liveness status
   * @returns Liveness status
   */
  async live(): Promise<boolean> {
    try {
      const response = await this.http.get<StatusResponse>('/live', {
        skipAuth: true,
      });
      return response.data.status === 'alive';
    } catch {
      return false;
    }
  }

  /**
   * Check if server is in maintenance mode
   * @returns Maintenance mode status
   */
  async isMaintenanceMode(): Promise<MaintenanceModeResponse> {
    const response = await this.http.get<MaintenanceModeResponse>(
      '/system/maintenance',
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Wait for server to become ready
   * @param maxWaitMs Maximum wait time in milliseconds
   * @param intervalMs Check interval in milliseconds
   * @returns True if server became ready, false if timeout
   */
  async waitForReady(maxWaitMs = 30000, intervalMs = 1000): Promise<boolean> {
    const startTime = Date.now();

    while (Date.now() - startTime < maxWaitMs) {
      const isReady = await this.ready();
      if (isReady) {
        return true;
      }
      await new Promise((resolve) => setTimeout(resolve, intervalMs));
    }

    return false;
  }

  /**
   * Check if specific service is healthy
   * @param serviceName Service name (database, redis)
   * @returns True if service is healthy
   */
  async isServiceHealthy(
    serviceName: 'database' | 'redis'
  ): Promise<boolean> {
    try {
      const health = await this.check();
      return health.services[serviceName] === 'healthy';
    } catch {
      return false;
    }
  }
}
