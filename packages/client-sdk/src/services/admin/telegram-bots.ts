/**
 * Admin Telegram Bots service
 */

import { BaseService } from '../base';
import type { HttpClient } from '../../core/http';
import type {
  TelegramBot,
  CreateTelegramBotRequest,
  UpdateTelegramBotRequest,
} from '../../types';

/** Admin Telegram Bots service for managing Telegram bots per application */
export class AdminTelegramBotsService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * List all Telegram bots for an application
   * @param appId Application ID
   * @returns Array of Telegram bots
   */
  async list(appId: string): Promise<TelegramBot[]> {
    const response = await this.http.get<{ bots: TelegramBot[]; total: number }>(
      `/api/admin/applications/${appId}/telegram-bots`
    );
    return response.data.bots;
  }

  /**
   * Create a new Telegram bot for an application
   * @param appId Application ID
   * @param data Telegram bot creation data
   * @returns Created Telegram bot
   */
  async create(appId: string, data: CreateTelegramBotRequest): Promise<TelegramBot> {
    const response = await this.http.post<TelegramBot>(
      `/api/admin/applications/${appId}/telegram-bots`,
      data
    );
    return response.data;
  }

  /**
   * Get a specific Telegram bot by ID
   * @param appId Application ID
   * @param id Telegram bot ID
   * @returns Telegram bot details
   */
  async getById(appId: string, id: string): Promise<TelegramBot> {
    const response = await this.http.get<TelegramBot>(
      `/api/admin/applications/${appId}/telegram-bots/${id}`
    );
    return response.data;
  }

  /**
   * Update a Telegram bot
   * @param appId Application ID
   * @param id Telegram bot ID
   * @param data Update data
   * @returns Updated Telegram bot
   */
  async update(appId: string, id: string, data: UpdateTelegramBotRequest): Promise<TelegramBot> {
    const response = await this.http.put<TelegramBot>(
      `/api/admin/applications/${appId}/telegram-bots/${id}`,
      data
    );
    return response.data;
  }

  /**
   * Delete a Telegram bot
   * @param appId Application ID
   * @param id Telegram bot ID
   * @returns void
   */
  async delete(appId: string, id: string): Promise<void> {
    await this.http.delete(`/api/admin/applications/${appId}/telegram-bots/${id}`);
  }
}
