/**
 * Admin-related types
 */

import type { IPFilterType, TimestampedEntity } from './common';

/** IP Filter entity */
export interface IPFilter extends TimestampedEntity {
  ip_address: string;
  type: IPFilterType;
  description?: string;
}

/** Create IP filter request */
export interface CreateIPFilterRequest {
  ip_address: string;
  type: IPFilterType;
  description?: string;
}

/** IP filter list response */
export interface IPFilterListResponse {
  filters: IPFilter[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

/** Audit log entry */
export interface AuditLogEntry {
  id: string;
  user_id: string;
  user_email?: string;
  action: string;
  resource: string;
  status: 'success' | 'failure';
  ip_address: string;
  user_agent: string;
  created_at: string;
  details?: Record<string, unknown>;
}

/** Audit log list response */
export interface AuditLogListResponse {
  logs: AuditLogEntry[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

/** Theme settings */
export interface ThemeSettings {
  primary_color: string;
  secondary_color: string;
  background_color: string;
}

/** Branding settings */
export interface BrandingSettings extends TimestampedEntity {
  logo_url?: string;
  favicon_url?: string;
  theme: ThemeSettings;
  company_name: string;
  support_email?: string;
  terms_url?: string;
  privacy_url?: string;
}

/** Public branding response (no sensitive data) */
export interface PublicBrandingResponse {
  logo_url?: string;
  favicon_url?: string;
  theme: ThemeSettings;
  company_name: string;
  support_email?: string;
  terms_url?: string;
  privacy_url?: string;
}

/** Update branding request */
export interface UpdateBrandingRequest {
  logo_url?: string;
  favicon_url?: string;
  primary_color?: string;
  secondary_color?: string;
  background_color?: string;
  company_name?: string;
  support_email?: string;
  terms_url?: string;
  privacy_url?: string;
}

/** Maintenance mode response */
export interface MaintenanceModeResponse {
  enabled: boolean;
  message?: string;
}

/** Update maintenance mode request */
export interface UpdateMaintenanceModeRequest {
  enabled: boolean;
  message?: string;
}

/** System health response */
export interface SystemHealthResponse {
  status: 'healthy' | 'degraded' | 'unhealthy';
  services: Record<string, 'healthy' | 'unhealthy'>;
  uptime: number;
  version: string;
}

/** Health check response */
export interface HealthResponse {
  status: 'healthy' | 'unhealthy';
  services: {
    database: 'healthy' | 'unhealthy';
    redis: 'healthy' | 'unhealthy';
  };
}

/** Geo distribution location */
export interface GeoLocation {
  country_code: string;
  country_name: string;
  city: string;
  login_count: number;
  latitude: number;
  longitude: number;
}

/** Geo distribution response */
export interface GeoDistributionResponse {
  locations: GeoLocation[];
  total: number;
  countries: number;
  cities: number;
}

// ============================================
// Webhook Types
// ============================================

/** Webhook retry configuration */
export interface WebhookRetryConfig {
  max_retries: number;
  initial_delay_ms: number;
  max_delay_ms: number;
}

/** Webhook entity */
export interface Webhook extends TimestampedEntity {
  name: string;
  url: string;
  events: string[];
  headers?: Record<string, string>;
  is_active: boolean;
  secret_key?: string;
  retry_config?: WebhookRetryConfig;
  last_triggered_at?: string;
  failure_count: number;
}

/** Create webhook request */
export interface CreateWebhookRequest {
  name: string;
  url: string;
  events: string[];
  headers?: Record<string, string>;
  retry_config?: {
    max_retries?: number;
    initial_delay_ms?: number;
    max_delay_ms?: number;
  };
}

/** Update webhook request */
export interface UpdateWebhookRequest {
  name?: string;
  url?: string;
  events?: string[];
  headers?: Record<string, string>;
  is_active?: boolean;
  retry_config?: {
    max_retries?: number;
    initial_delay_ms?: number;
    max_delay_ms?: number;
  };
}

/** Test webhook request */
export interface TestWebhookRequest {
  event_type: string;
  payload?: Record<string, unknown>;
}

/** Webhook delivery entry */
export interface WebhookDelivery {
  id: string;
  webhook_id: string;
  event_type: string;
  payload: Record<string, unknown>;
  status_code: number;
  response_body?: string;
  error?: string;
  attempt_number: number;
  delivered_at: string;
}

/** Webhook list response */
export interface WebhookListResponse {
  webhooks: Webhook[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

/** Webhook delivery list response */
export interface WebhookDeliveryListResponse {
  deliveries: WebhookDelivery[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

/** Create webhook response (includes secret key) */
export interface CreateWebhookResponse {
  webhook: Webhook;
  secret_key: string;
}

// ============================================
// Email Template Types
// ============================================

/** Email template type */
export type EmailTemplateType = 'verification' | 'password_reset' | 'welcome' | '2fa' | 'otp_login' | 'otp_registration' | 'custom';

/** Email template entity */
export interface EmailTemplate extends TimestampedEntity {
  type: EmailTemplateType;
  name: string;
  subject: string;
  html_body: string;
  text_body?: string;
  variables: string[];
  is_active: boolean;
  application_id?: string;
  application?: { id: string; name: string; display_name?: string };
}

/** Create email template request */
export interface CreateEmailTemplateRequest {
  type: EmailTemplateType;
  name: string;
  subject: string;
  html_body: string;
  text_body?: string;
  variables?: string[];
}

/** Update email template request */
export interface UpdateEmailTemplateRequest {
  name?: string;
  subject?: string;
  html_body?: string;
  text_body?: string;
  variables?: string[];
  is_active?: boolean;
}

/** Preview email template request */
export interface PreviewEmailTemplateRequest {
  html_body: string;
  text_body?: string;
  variables?: Record<string, string>;
}

/** Preview email template response */
export interface PreviewEmailTemplateResponse {
  html: string;
  text?: string;
}

/** Email template list response */
export interface EmailTemplateListResponse {
  templates: EmailTemplate[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

/** Email template types response */
export interface EmailTemplateTypesResponse {
  types: EmailTemplateType[];
}

/** Email template variables response */
export interface EmailTemplateVariablesResponse {
  variables: string[];
}

// ============================================
// OAuth Provider Configuration Types
// ============================================

/** OAuth provider type */
export type OAuthProviderType = 'google' | 'github' | 'yandex' | 'telegram' | 'instagram' | 'onec' | string;

/** OAuth provider configuration entity */
export interface OAuthProviderConfig extends TimestampedEntity {
  provider: OAuthProviderType;
  application_id: string;
  client_id: string;
  callback_url: string;
  scopes?: string[];
  auth_url?: string;
  token_url?: string;
  user_info_url?: string;
  is_active: boolean;
}

/** Create OAuth provider request */
export interface CreateOAuthProviderRequest {
  provider: OAuthProviderType;
  client_id: string;
  client_secret: string;
  callback_url: string;
  scopes?: string[];
  auth_url?: string;
  token_url?: string;
  user_info_url?: string;
}

/** Update OAuth provider request */
export interface UpdateOAuthProviderRequest {
  client_id?: string;
  client_secret?: string;
  callback_url?: string;
  scopes?: string[];
  auth_url?: string;
  token_url?: string;
  user_info_url?: string;
  is_active?: boolean;
}

/** OAuth provider list response */
export interface OAuthProviderListResponse {
  providers: OAuthProviderConfig[];
}

// Note: OAuth Client, Scope, and Consent types have been moved to oauth-provider.ts
// for the full OAuth 2.0/OIDC provider implementation.

// ============================================
// Groups Types
// ============================================

/** Group entity */
export interface Group extends TimestampedEntity {
  id: string;
  name: string;
  display_name: string;
  description?: string;
  parent_group_id?: string;
  is_system_group: boolean;
  member_count: number;
}

/** Create group request */
export interface CreateGroupRequest {
  name: string;
  display_name: string;
  description?: string;
  parent_group_id?: string;
}

/** Update group request */
export interface UpdateGroupRequest {
  display_name?: string;
  description?: string;
  parent_group_id?: string;
}

/** Add group members request */
export interface AddGroupMembersRequest {
  user_ids: string[];
}

/** Group list response */
export interface GroupListResponse {
  groups: Group[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

/** Group members response */
export interface GroupMembersResponse {
  users: Array<{
    id: string;
    email: string;
    username: string;
    full_name?: string;
  }>;
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// ============================================
// LDAP Types
// ============================================

/** LDAP configuration entity */
export interface LDAPConfig extends TimestampedEntity {
  id: string;
  server: string;
  port: number;
  use_tls: boolean;
  use_ssl: boolean;
  insecure: boolean;
  bind_dn: string;
  base_dn: string;
  user_search_base?: string;
  group_search_base?: string;
  user_search_filter: string;
  group_search_filter: string;
  user_id_attribute: string;
  user_email_attribute: string;
  user_name_attribute: string;
  group_id_attribute: string;
  group_name_attribute: string;
  group_member_attribute: string;
  sync_enabled: boolean;
  sync_interval: number; // in seconds
  is_active: boolean;
  last_sync_at?: string;
  next_sync_at?: string;
  last_test_at?: string;
  last_test_result?: string;
}

/** Create LDAP config request */
export interface CreateLDAPConfigRequest {
  server: string;
  port: number;
  use_tls?: boolean;
  use_ssl?: boolean;
  insecure?: boolean;
  bind_dn: string;
  bind_password: string;
  base_dn: string;
  user_search_base?: string;
  group_search_base?: string;
  user_search_filter?: string;
  group_search_filter?: string;
  user_id_attribute?: string;
  user_email_attribute?: string;
  user_name_attribute?: string;
  group_id_attribute?: string;
  group_name_attribute?: string;
  group_member_attribute?: string;
  sync_enabled?: boolean;
  sync_interval?: number; // in seconds
}

/** Update LDAP config request */
export interface UpdateLDAPConfigRequest {
  server?: string;
  port?: number;
  use_tls?: boolean;
  use_ssl?: boolean;
  insecure?: boolean;
  bind_dn?: string;
  bind_password?: string;
  base_dn?: string;
  user_search_base?: string;
  group_search_base?: string;
  user_search_filter?: string;
  group_search_filter?: string;
  sync_enabled?: boolean;
  sync_interval?: number; // in seconds
  is_active?: boolean;
}

/** LDAP test connection request */
export interface LDAPTestConnectionRequest {
  server: string;
  port: number;
  use_tls?: boolean;
  use_ssl?: boolean;
  insecure?: boolean;
  bind_dn: string;
  bind_password: string;
  base_dn: string;
}

/** LDAP test connection response */
export interface LDAPTestConnectionResponse {
  success: boolean;
  message: string;
  error?: string;
  user_count?: number;
  group_count?: number;
}

/** LDAP sync log */
export interface LDAPSyncLog {
  id: string;
  ldap_config_id: string;
  status: 'success' | 'failed' | 'partial';
  users_synced: number;
  users_created: number;
  users_updated: number;
  users_deleted: number;
  groups_synced: number;
  groups_created: number;
  groups_updated: number;
  error_message?: string;
  started_at: string;
  completed_at?: string;
  duration_ms: number;
}

/** LDAP sync request */
export interface LDAPSyncRequest {
  sync_users?: boolean;
  sync_groups?: boolean;
  dry_run?: boolean;
}

/** LDAP sync response */
export interface LDAPSyncResponse {
  status: string;
  sync_log_id: string;
  users_synced: number;
  users_created: number;
  users_updated: number;
  users_deleted: number;
  groups_synced: number;
  groups_created: number;
  groups_updated: number;
  message: string;
  error?: string;
}

/** LDAP config list response */
export interface LDAPConfigListResponse {
  configs: LDAPConfig[];
  total: number;
}

/** LDAP sync logs response */
export interface LDAPSyncLogsResponse {
  logs: LDAPSyncLog[];
  total: number;
}

// ============================================
// SAML Types
// ============================================

/** SAML Service Provider entity */
export interface SAMLServiceProvider extends TimestampedEntity {
  id: string;
  name: string;
  entity_id: string;
  acs_url: string;
  slo_url?: string;
  x509_cert?: string;
  metadata_url?: string;
  is_active: boolean;
}

/** Create SAML SP request */
export interface CreateSAMLSPRequest {
  name: string;
  entity_id: string;
  acs_url: string;
  slo_url?: string;
  x509_cert?: string;
  metadata_url?: string;
}

/** Update SAML SP request */
export interface UpdateSAMLSPRequest {
  name?: string;
  entity_id?: string;
  acs_url?: string;
  slo_url?: string;
  x509_cert?: string;
  metadata_url?: string;
  is_active?: boolean;
}

/** SAML SP list response */
export interface SAMLSPListResponse {
  sps: SAMLServiceProvider[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

/** SAML metadata response */
export interface SAMLMetadataResponse {
  metadata: string; // XML string
}

// ============================================
// Bulk Operations Types
// ============================================

/** Bulk user create item */
export interface BulkUserCreate {
  email: string;
  username: string;
  full_name?: string;
  password: string;
  is_active?: boolean;
  email_verified?: boolean;
}

/** Bulk user update item */
export interface BulkUserUpdate {
  id: string;
  email?: string;
  username?: string;
  full_name?: string;
  is_active?: boolean;
}

/** Bulk create users request */
export interface BulkCreateUsersRequest {
  users: BulkUserCreate[];
}

/** Bulk update users request */
export interface BulkUpdateUsersRequest {
  users: BulkUserUpdate[];
}

/** Bulk delete users request */
export interface BulkDeleteUsersRequest {
  user_ids: string[];
}

/** Bulk assign roles request */
export interface BulkAssignRolesRequest {
  user_ids: string[];
  role_ids: string[];
}

/** Bulk operation error */
export interface BulkOperationError {
  index: number;
  id?: string;
  email?: string;
  message: string;
}

/** Bulk operation item result */
export interface BulkOperationItemResult {
  index: number;
  id: string;
  email: string;
  success: boolean;
  message?: string;
}

/** Bulk operation result */
export interface BulkOperationResult {
  total: number;
  success: number;
  failed: number;
  errors?: BulkOperationError[];
  results?: BulkOperationItemResult[];
}

// ============================================
// SCIM Types
// ============================================

/** SCIM configuration */
export interface SCIMConfig {
  base_url: string;
  enabled: boolean;
  supported_operations: string[];
}

/** SCIM metadata response */
export interface SCIMMetadataResponse {
  metadata: Record<string, unknown>;
}

// ============================================
// Application Types
// ============================================

/** Authentication method */
export type AuthMethod = 'password' | 'otp_email' | 'otp_sms' | 'oauth_google' | 'oauth_github' | 'oauth_yandex' | 'oauth_telegram' | 'totp' | 'api_key';

/** Application entity */
export interface Application {
  id: string;
  name: string;
  display_name: string;
  description?: string;
  homepage_url?: string;
  callback_urls?: string[];
  allowed_auth_methods: (AuthMethod | string)[];
  is_active: boolean;
  is_system: boolean;
  owner_id?: string;
  secret_prefix?: string;
  secret_last_rotated_at?: string;
  owner?: { id: string; email?: string; username?: string; full_name?: string };
  created_at: string;
  updated_at: string;
  branding?: ApplicationBranding;
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
}

/** Create application request */
export interface CreateApplicationRequest {
  name: string;
  display_name: string;
  description?: string;
  homepage_url?: string;
  callback_urls?: string[];
  allowed_auth_methods?: (AuthMethod | string)[];
  is_active?: boolean;
}

/** Create application response */
export interface CreateApplicationResponse {
  application: Application;
  secret?: string;
  warning?: string;
}

/** Update application request */
export interface UpdateApplicationRequest {
  display_name?: string;
  description?: string;
  homepage_url?: string;
  callback_urls?: string[];
  allowed_auth_methods?: (AuthMethod | string)[];
  is_active?: boolean;
}

/** Application list response */
export interface ApplicationListResponse {
  applications: Application[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

/** Rotate application secret response */
export interface AppRotateSecretResponse {
  secret: string;
  warning: string;
}

/** User app profile entity */
export interface UserAppProfile {
  id: string;
  user_id: string;
  application_id: string;
  display_name?: string;
  avatar_url?: string;
  nickname?: string;
  metadata?: Record<string, unknown>;
  app_roles?: string[];
  is_active: boolean;
  is_banned: boolean;
  ban_reason?: string;
  banned_at?: string;
  banned_by?: string;
  last_access_at?: string;
  created_at: string;
  updated_at: string;
  user?: { id: string; email?: string; username?: string; full_name?: string };
  application?: Application;
}

/** User app profile list response */
export interface UserAppProfileListResponse {
  profiles: UserAppProfile[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

/** Ban user request */
export interface BanUserRequest {
  reason: string;
}

/** Update user app profile request */
export interface UpdateUserAppProfileRequest {
  display_name?: string;
  avatar_url?: string;
  nickname?: string;
  metadata?: Record<string, unknown>;
  app_roles?: string[];
  is_active?: boolean;
  is_banned?: boolean;
  ban_reason?: string;
}

/** Auth configuration response (public) */
export interface AuthConfigResponse {
  application_id: string;
  name: string;
  display_name: string;
  allowed_auth_methods: (AuthMethod | string)[];
  oauth_providers?: string[];
  branding?: ApplicationBranding;
}

/** Request to ban a user from application */
export interface BanUserFromApplicationRequest {
  reason: string;
}

/** Response for listing application users */
export interface ListApplicationUsersResponse {
  profiles: UserAppProfile[];
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

/** Update application branding request */
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
