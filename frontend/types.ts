/**
 * Frontend-specific types
 * Common types are imported from @auth-gateway/client-sdk
 */

// Import types for use in aliases
import type {
  // User types
  User,
  AdminUserResponse,
  RoleInfo,
  AccountType,
  AdminCreateUserRequest,
  AdminUpdateUserRequest,
  OAuthAccount,
  // RBAC types
  Role,
  Permission,
  // Session types
  Session,
  // API Key types
  APIKey,
  AdminAPIKeyResponse,
  // Audit types
  AuditLogEntry,
  // Admin types
  IPFilter,
  BrandingSettings,
  ThemeSettings,
  SystemHealthResponse,
  MaintenanceModeResponse,
  // New types
  Webhook,
  EmailTemplate,
  EmailTemplateType,
  OAuthProviderConfig,
  OAuthProviderType,
} from '@auth-gateway/client-sdk';

// Re-export common SDK types for convenience
export type {
  User,
  AdminUserResponse,
  RoleInfo,
  AccountType,
  AdminCreateUserRequest,
  AdminUpdateUserRequest,
  OAuthAccount,
  Role,
  Permission,
  Session,
  APIKey,
  AdminAPIKeyResponse,
  AuditLogEntry,
  IPFilter,
  BrandingSettings,
  ThemeSettings,
  SystemHealthResponse,
  MaintenanceModeResponse,
  Webhook,
  EmailTemplate,
  EmailTemplateType,
  OAuthProviderConfig,
  OAuthProviderType,
} from '@auth-gateway/client-sdk';

// ============================================
// Backward compatibility type aliases
// ============================================

export type BrandingConfig = BrandingSettings;
export type IpRule = IPFilter;
export type RoleDefinition = Role;
export type WebhookEndpoint = Webhook;
export type ApiKey = APIKey;
export type AuditLog = AuditLogEntry;
export type UserSession = Session;

// Keep deprecated enum for backward compatibility
export enum UserRole {
  ADMIN = 'admin',
  MODERATOR = 'moderator',
  USER = 'user',
}

// ============================================
// Frontend-specific types (not in SDK)
// ============================================

/** Dashboard statistics - aggregated view for admin dashboard */
export interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  newUsersToday: number;
  totalApiKeys: number;
  activeApiKeys: number;
  loginAttemptsToday: number;
  failedLoginAttemptsToday: number;
}

/** Password policy settings for security configuration */
export interface PasswordPolicy {
  minLength: number;
  requireUppercase: boolean;
  requireLowercase: boolean;
  requireNumbers: boolean;
  requireSpecial: boolean;
  historyCount: number;
  expiryDays: number;
  jwtTtlMinutes: number;
  refreshTtlDays: number;
}

/** SMS provider type for SMS settings */
export type SmsProviderType = 'aws' | 'twilio' | 'mock';

/** SMS configuration */
export interface SmsConfig {
  provider: SmsProviderType;
  awsRegion?: string;
  awsAccessKeyId?: string;
  awsSecretAccessKey?: string;
  twilioAccountSid?: string;
  twilioAuthToken?: string;
  twilioPhoneNumber?: string;
}

/** Service account for service-to-service auth */
export interface ServiceAccount {
  id: string;
  name: string;
  description: string;
  client_id: string;
  is_active: boolean;
  created_at: string;
  last_used_at?: string;
}
