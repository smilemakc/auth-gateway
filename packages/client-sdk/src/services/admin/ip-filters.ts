/**
 * Admin IP Filters service
 */

import type { HttpClient } from '../../core/http';
import type { IPFilterType, MessageResponse } from '../../types/common';
import type {
  CreateIPFilterRequest,
  IPFilter,
  IPFilterListResponse,
} from '../../types/admin';
import { BaseService } from '../base';

/** Admin IP Filters service for IP whitelist/blacklist management */
export class AdminIPFiltersService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List IP filters
   * @param type Filter type (whitelist or blacklist)
   * @param page Page number
   * @param perPage Items per page
   * @returns List of IP filters
   */
  async list(
    type?: IPFilterType,
    page = 1,
    perPage = 20
  ): Promise<IPFilterListResponse> {
    const response = await this.http.get<IPFilterListResponse>(
      '/api/admin/ip-filters',
      { query: { type, page, per_page: perPage } }
    );
    return response.data;
  }

  /**
   * Create a new IP filter
   * @param data IP filter data
   * @returns Created IP filter
   */
  async create(data: CreateIPFilterRequest): Promise<IPFilter> {
    const response = await this.http.post<IPFilter>('/api/admin/ip-filters', data);
    return response.data;
  }

  /**
   * Delete an IP filter
   * @param id IP filter ID
   * @returns Success message
   */
  async delete(id: string): Promise<MessageResponse> {
    const response = await this.http.delete<MessageResponse>(
      `/admin/ip-filters/${id}`
    );
    return response.data;
  }

  /**
   * Add IP to whitelist
   * @param ipAddress IP address or CIDR
   * @param description Optional description
   * @returns Created IP filter
   */
  async whitelist(ipAddress: string, description?: string): Promise<IPFilter> {
    return this.create({
      ip_address: ipAddress,
      type: 'whitelist',
      description,
    });
  }

  /**
   * Add IP to blacklist
   * @param ipAddress IP address or CIDR
   * @param description Optional description
   * @returns Created IP filter
   */
  async blacklist(ipAddress: string, description?: string): Promise<IPFilter> {
    return this.create({
      ip_address: ipAddress,
      type: 'blacklist',
      description,
    });
  }

  /**
   * Get all whitelisted IPs
   * @returns List of whitelisted IPs
   */
  async getWhitelist(): Promise<IPFilter[]> {
    const filters: IPFilter[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const response = await this.list('whitelist', page, 100);
      filters.push(...response.filters);
      hasMore = filters.length < response.total;
      page++;
    }

    return filters;
  }

  /**
   * Get all blacklisted IPs
   * @returns List of blacklisted IPs
   */
  async getBlacklist(): Promise<IPFilter[]> {
    const filters: IPFilter[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const response = await this.list('blacklist', page, 100);
      filters.push(...response.filters);
      hasMore = filters.length < response.total;
      page++;
    }

    return filters;
  }

  /**
   * Check if an IP is in the whitelist
   * @param ipAddress IP address
   * @returns True if IP is whitelisted
   */
  async isWhitelisted(ipAddress: string): Promise<boolean> {
    const whitelist = await this.getWhitelist();
    return whitelist.some((f) => f.ip_address === ipAddress);
  }

  /**
   * Check if an IP is in the blacklist
   * @param ipAddress IP address
   * @returns True if IP is blacklisted
   */
  async isBlacklisted(ipAddress: string): Promise<boolean> {
    const blacklist = await this.getBlacklist();
    return blacklist.some((f) => f.ip_address === ipAddress);
  }
}
