/**
 * OAuth-related types
 */

import type { OAuthProvider } from './common';
import type { User } from './user';

/** OAuth provider info */
export interface OAuthProviderInfo {
  name: OAuthProvider;
  displayName: string;
  iconUrl?: string;
  enabled: boolean;
}

/** OAuth login response */
export interface OAuthLoginResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
  isNewUser: boolean;
}

/** OAuth callback query parameters */
export interface OAuthCallbackParams {
  code: string;
  state: string;
  responseType?: 'json';
}

/** Telegram OAuth callback data */
export interface TelegramCallbackData {
  id: number;
  firstName: string;
  lastName?: string;
  username?: string;
  photoUrl?: string;
  authDate: number;
  hash: string;
}

/** OAuth account info (linked to user) */
export interface OAuthAccount {
  id: string;
  userId: string;
  provider: OAuthProvider;
  providerUserId: string;
  tokenExpiresAt?: string;
  createdAt: string;
  updatedAt: string;
}
