/**
 * API Key types
 */

import type { APIKeyScope, TimestampedEntity } from './common';

/** API Key entity */
export interface APIKey extends TimestampedEntity {
  user_id: string;
  name: string;
  description?: string;
  key_prefix: string;
  scopes: APIKeyScope[];
  is_active: boolean;
  last_used_at?: string;
  expires_at?: string;
}

/** Create API key request */
export interface CreateAPIKeyRequest {
  name: string;
  description?: string;
  scopes: APIKeyScope[];
  expires_at?: string;
}

/** Create API key response */
export interface CreateAPIKeyResponse {
  api_key: APIKey;
  /** The plain API key - only returned once at creation */
  plain_key: string;
}

/** Update API key request */
export interface UpdateAPIKeyRequest {
  name?: string;
  description?: string;
  scopes?: APIKeyScope[];
  is_active?: boolean;
}

/** List API keys response */
export interface ListAPIKeysResponse {
  api_keys: APIKey[];
  total: number;
}

/** Admin API key response with user info */
export interface AdminAPIKeyResponse extends APIKey {
  user_email?: string;
  owner_name?: string;
}
