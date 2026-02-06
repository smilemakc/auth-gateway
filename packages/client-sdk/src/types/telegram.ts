/**
 * Telegram integration types
 */

export interface TelegramBot {
  id: string;
  application_id: string;
  bot_token?: string;
  bot_username: string;
  display_name: string;
  is_auth_bot: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateTelegramBotRequest {
  bot_token: string;
  bot_username: string;
  display_name: string;
  is_auth_bot?: boolean;
  is_active?: boolean;
}

export interface UpdateTelegramBotRequest {
  bot_token?: string;
  bot_username?: string;
  display_name?: string;
  is_auth_bot?: boolean;
  is_active?: boolean;
}

export interface UserTelegramAccount {
  id: string;
  user_id: string;
  telegram_user_id: string;
  telegram_username: string;
  first_name: string;
  last_name: string;
  photo_url: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface UserTelegramBotAccess {
  id: string;
  user_id: string;
  telegram_bot_id: string;
  telegram_account_id: string;
  is_active: boolean;
  first_interaction_at: string;
  last_interaction_at: string;
  bot?: TelegramBot;
  account?: UserTelegramAccount;
}
