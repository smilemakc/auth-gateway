/**
 * API Key types
 */

import type { APIKeyScope, TimestampedEntity } from './common';

/** API Key entity */
export interface APIKey extends TimestampedEntity {
  userId: string;
  name: string;
  description?: string;
  keyPrefix: string;
  scopes: APIKeyScope[];
  isActive: boolean;
  lastUsedAt?: string;
  expiresAt?: string;
}

/** Create API key request */
export interface CreateAPIKeyRequest {
  name: string;
  description?: string;
  scopes: APIKeyScope[];
  expiresAt?: string;
}

/** Create API key response */
export interface CreateAPIKeyResponse {
  apiKey: APIKey;
  /** The plain API key - only returned once at creation */
  plainKey: string;
}

/** Update API key request */
export interface UpdateAPIKeyRequest {
  name?: string;
  description?: string;
  scopes?: APIKeyScope[];
  isActive?: boolean;
}

/** List API keys response */
export interface ListAPIKeysResponse {
  apiKeys: APIKey[];
  total: number;
}

/** Admin API key response with user info */
export interface AdminAPIKeyResponse extends APIKey {
  userEmail?: string;
  userName?: string;
}
