/**
 * Admin User Telegram service
 */

import { BaseService } from '../base';
import type { HttpClient } from '../../core/http';
import type { UserTelegramAccount, UserTelegramBotAccess } from '../../types';

/** Admin User Telegram service for managing user Telegram accounts and bot access */
export class AdminUserTelegramService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Get all Telegram accounts for a user
   * @param userId User ID
   * @returns Array of Telegram accounts
   */
  async getAccounts(userId: string): Promise<UserTelegramAccount[]> {
    const response = await this.http.get<UserTelegramAccount[]>(
      `/api/admin/users/${userId}/telegram-accounts`
    );
    return response.data;
  }

  /**
   * Get Telegram bot access records for a user
   * @param userId User ID
   * @param appId Optional application ID filter
   * @returns Array of bot access records
   */
  async getBotAccess(userId: string, appId?: string): Promise<UserTelegramBotAccess[]> {
    const query = appId ? { app_id: appId } : undefined;
    const response = await this.http.get<UserTelegramBotAccess[]>(
      `/api/admin/users/${userId}/telegram-bot-access`,
      { query }
    );
    return response.data;
  }
}
