/**
 * Admin Audit Logs service
 */

import type { HttpClient } from '../../core/http';
import type { AuditLogEntry, AuditLogListResponse } from '../../types/admin';
import { BaseService } from '../base';

/** Query options for audit logs */
export interface AuditLogQueryOptions {
  userId?: string;
  action?: string;
  resource?: string;
  status?: 'success' | 'failure';
  startDate?: Date;
  endDate?: Date;
  page?: number;
  pageSize?: number;
}

/** Admin Audit service for audit log management */
export class AdminAuditService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List audit logs with filters
   * @param options Query options
   * @returns Paginated audit logs
   */
  async list(options: AuditLogQueryOptions = {}): Promise<AuditLogListResponse> {
    const response = await this.http.get<AuditLogEntry[] | AuditLogListResponse>(
      '/admin/audit-logs',
      {
        headers: {},
        query: {
          user_id: options.userId,
          action: options.action,
          resource: options.resource,
          status: options.status,
          start_date: options.startDate?.toISOString(),
          end_date: options.endDate?.toISOString(),
          page: options.page ?? 1,
          page_size: options.pageSize ?? 50,
        },
      }
    );

    // Backend returns an array directly, wrap it for consistency
    if (Array.isArray(response.data)) {
      return {
        logs: response.data,
        total: response.data.length,
        page: options.page ?? 1,
        pageSize: options.pageSize ?? 50,
      };
    }

    return response.data;
  }

  /**
   * Get audit logs for a specific user
   * @param userId User ID
   * @param page Page number
   * @param pageSize Items per page
   * @returns User's audit logs
   */
  async getByUser(
    userId: string,
    page = 1,
    pageSize = 50
  ): Promise<AuditLogListResponse> {
    return this.list({ userId, page, pageSize });
  }

  /**
   * Get audit logs for a specific action
   * @param action Action name (e.g., 'login', 'logout', 'password_change')
   * @param page Page number
   * @param pageSize Items per page
   * @returns Audit logs for the action
   */
  async getByAction(
    action: string,
    page = 1,
    pageSize = 50
  ): Promise<AuditLogListResponse> {
    return this.list({ action, page, pageSize });
  }

  /**
   * Get failed audit logs
   * @param page Page number
   * @param pageSize Items per page
   * @returns Failed audit logs
   */
  async getFailures(page = 1, pageSize = 50): Promise<AuditLogListResponse> {
    return this.list({ status: 'failure', page, pageSize });
  }

  /**
   * Get recent audit logs
   * @param hours Number of hours to look back
   * @param pageSize Maximum results
   * @returns Recent audit logs
   */
  async getRecent(hours = 24, pageSize = 100): Promise<AuditLogEntry[]> {
    const startDate = new Date();
    startDate.setHours(startDate.getHours() - hours);

    const response = await this.list({
      startDate,
      page: 1,
      pageSize,
    });

    return response.logs;
  }

  /**
   * Get login attempts for a specific IP
   * @param ipAddress IP address
   * @param page Page number
   * @param pageSize Items per page
   * @returns Login attempts from the IP
   */
  async getLoginAttemptsByIP(
    ipAddress: string,
    page = 1,
    pageSize = 50
  ): Promise<AuditLogEntry[]> {
    // Fetch login logs and filter by IP
    const response = await this.list({
      action: 'login',
      page,
      pageSize: pageSize * 5, // Fetch more to filter
    });

    return response.logs
      .filter((log) => log.ipAddress === ipAddress)
      .slice(0, pageSize);
  }
}
