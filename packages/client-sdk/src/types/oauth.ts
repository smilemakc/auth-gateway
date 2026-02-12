/**
 * OAuth-related types
 */

import type { OAuthProvider } from './common';
import type { User } from './user';

/** OAuth provider info */
export interface OAuthProviderInfo {
  name: OAuthProvider;
  display_name: string;
  icon_url?: string;
  enabled: boolean;
}

/** OAuth login response */
export interface OAuthLoginResponse {
  access_token: string;
  refresh_token: string;
  user: User;
  is_new_user: boolean;
}

/** OAuth callback query parameters */
export interface OAuthCallbackParams {
  code: string;
  state: string;
  response_type?: 'json';
}

/** Telegram OAuth callback data */
export interface TelegramCallbackData {
  id: number;
  first_name: string;
  last_name?: string;
  username?: string;
  photo_url?: string;
  auth_date: number;
  hash: string;
}

/** OAuth account info (linked to user) */
export interface OAuthAccount {
  id: string;
  user_id: string;
  provider: OAuthProvider;
  provider_user_id: string;
  token_expires_at?: string;
  created_at: string;
  updated_at: string;
}
