/**
 * Common types used across the SDK
 */

/** Standard message response */
export interface MessageResponse {
  message: string;
}

/** Email-based message response */
export interface EmailMessageResponse extends MessageResponse {
  email: string;
}

/** Phone-based message response */
export interface PhoneMessageResponse extends MessageResponse {
  phone: string;
}

/** Validation response */
export interface ValidationResponse {
  valid: boolean;
  message?: string;
}

/** Timestamp fields */
export interface Timestamps {
  created_at: string;
  updated_at: string;
}

/** Entity with ID */
export interface Entity {
  id: string;
}

/** Entity with timestamps */
export interface TimestampedEntity extends Entity, Timestamps {}

/** Generic list response */
export interface ListResponse<T> {
  total: number;
  page?: number;
  page_size?: number;
  total_pages?: number;
}

/** Error response from API */
export interface ErrorResponse {
  error: string;
  message: string;
  details?: string;
}

/** Account types */
export type AccountType = 'human' | 'service';

/** @deprecated Use Role model from rbac.ts instead */
export type UserRole = 'user' | 'moderator' | 'admin';

/** OTP types */
export type OTPType = 'verification' | 'password_reset' | '2fa' | 'login';

/** OAuth providers */
export type OAuthProvider = 'google' | 'yandex' | 'github' | 'instagram' | 'telegram';

/** API key scopes */
export type APIKeyScope =
  | 'users:read'
  | 'users:write'
  | 'profile:read'
  | 'profile:write'
  | 'admin:all'
  | 'token:validate'
  | 'token:introspect'
  | 'all';

/** IP filter types */
export type IPFilterType = 'whitelist' | 'blacklist';
