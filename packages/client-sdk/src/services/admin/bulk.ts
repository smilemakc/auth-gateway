/**
 * Admin Bulk Operations service
 */

import type { HttpClient } from '../../core/http';
import type {
  BulkCreateUsersRequest,
  BulkUpdateUsersRequest,
  BulkDeleteUsersRequest,
  BulkAssignRolesRequest,
  BulkOperationResult,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin Bulk Operations service for bulk user management */
export class AdminBulkService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Bulk create users
   * @param data Bulk create request
   * @returns Operation result
   */
  async bulkCreateUsers(data: BulkCreateUsersRequest): Promise<BulkOperationResult> {
    const response = await this.http.post<BulkOperationResult>('/api/admin/users/bulk-create', data);
    return response.data;
  }

  /**
   * Bulk update users
   * @param data Bulk update request
   * @returns Operation result
   */
  async bulkUpdateUsers(data: BulkUpdateUsersRequest): Promise<BulkOperationResult> {
    const response = await this.http.put<BulkOperationResult>('/api/admin/users/bulk-update', data);
    return response.data;
  }

  /**
   * Bulk delete users
   * @param data Bulk delete request
   * @returns Operation result
   */
  async bulkDeleteUsers(data: BulkDeleteUsersRequest): Promise<BulkOperationResult> {
    const response = await this.http.post<BulkOperationResult>('/api/admin/users/bulk-delete', data);
    return response.data;
  }

  /**
   * Bulk assign roles to users
   * @param data Bulk assign roles request
   * @returns Operation result
   */
  async bulkAssignRoles(data: BulkAssignRolesRequest): Promise<BulkOperationResult> {
    const response = await this.http.post<BulkOperationResult>('/api/admin/users/bulk-assign-roles', data);
    return response.data;
  }
}
