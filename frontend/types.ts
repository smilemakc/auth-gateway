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
  // OAuth Client (OIDC Provider) types
  OAuthClient,
  CreateOAuthClientRequest,
  CreateOAuthClientResponse,
  UpdateOAuthClientRequest,
  OAuthScope,
  UserConsent,
  GrantType,
  ClientType,
  // Application OAuth Provider types
  ApplicationOAuthProvider,
  CreateAppOAuthProviderRequest,
  UpdateAppOAuthProviderRequest,
  // Telegram types
  TelegramBot,
  CreateTelegramBotRequest,
  UpdateTelegramBotRequest,
  UserTelegramAccount,
  UserTelegramBotAccess,
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
  OAuthClient,
  CreateOAuthClientRequest,
  CreateOAuthClientResponse,
  UpdateOAuthClientRequest,
  OAuthScope,
  UserConsent,
  GrantType,
  ClientType,
  ApplicationOAuthProvider,
  CreateAppOAuthProviderRequest,
  UpdateAppOAuthProviderRequest,
  TelegramBot,
  CreateTelegramBotRequest,
  UpdateTelegramBotRequest,
  UserTelegramAccount,
  UserTelegramBotAccess,
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

// ============================================
// Application types (Multi-tenant)
// ============================================

/** Application entity for multi-tenant support */
export interface Application {
  id: string;
  name: string;
  display_name: string;
  description?: string;
  homepage_url?: string;
  callback_urls: string[];
  allowed_auth_methods?: string[];
  secret_prefix?: string;
  secret_last_rotated_at?: string;
  is_active: boolean;
  is_system: boolean;
  owner_id?: string;
  owner?: User;
  branding?: ApplicationBranding;
  created_at: string;
  updated_at: string;
}

/** Application branding settings */
export interface ApplicationBranding {
  id: string;
  application_id: string;
  logo_url?: string;
  favicon_url?: string;
  primary_color?: string;
  secondary_color?: string;
  background_color?: string;
  custom_css?: string;
  company_name?: string;
  support_email?: string;
  terms_url?: string;
  privacy_url?: string;
  updated_at: string;
}

/** User profile within an application */
export interface UserApplicationProfile {
  id: string;
  user_id: string;
  application_id: string;
  display_name?: string;
  avatar_url?: string;
  nickname?: string;
  metadata?: Record<string, unknown>;
  app_roles: string[];
  is_active: boolean;
  is_banned: boolean;
  ban_reason?: string;
  banned_at?: string;
  banned_by?: string;
  last_access_at?: string;
  created_at: string;
  updated_at: string;
  user?: User;
  application?: Application;
}

/** Request to create an application */
export interface CreateApplicationRequest {
  name: string;
  display_name: string;
  description?: string;
  homepage_url?: string;
  callback_urls?: string[];
  allowed_auth_methods?: string[];
  owner_id?: string;
}

/** Request to update an application */
export interface UpdateApplicationRequest {
  display_name?: string;
  description?: string;
  homepage_url?: string;
  callback_urls?: string[];
  allowed_auth_methods?: string[];
  is_active?: boolean;
  owner_id?: string;
}

/** Request to update application branding */
export interface UpdateApplicationBrandingRequest {
  logo_url?: string;
  favicon_url?: string;
  primary_color?: string;
  secondary_color?: string;
  background_color?: string;
  custom_css?: string;
  company_name?: string;
  support_email?: string;
  terms_url?: string;
  privacy_url?: string;
}

/** Request to ban a user from application */
export interface BanUserFromApplicationRequest {
  reason: string;
}

/** Response for listing applications */
export interface ListApplicationsResponse {
  applications: Application[];
  total: number;
  page: number;
  page_size: number;
}

/** Response for listing application users */
export interface ListApplicationUsersResponse {
  profiles: UserApplicationProfile[];
  total: number;
  page: number;
  page_size: number;
}

/** User entry for import */
export interface ImportUserEntry {
  id?: string;
  email: string;
  username?: string;
  password_hash_import?: string;
  full_name?: string;
  is_active?: boolean;
  skip_email_verification?: boolean;
  app_roles?: string[];
}

/** Request to import users */
export interface ImportUsersRequest {
  users: ImportUserEntry[];
  on_conflict: 'skip' | 'update' | 'error';
}

/** Response from import users */
export interface ImportUsersResponse {
  imported: number;
  skipped: number;
  updated: number;
  errors: number;
  details: ImportDetail[];
}

/** Detail of a single import operation */
export interface ImportDetail {
  email: string;
  status: string;
  reason?: string;
  user_id?: string;
}
