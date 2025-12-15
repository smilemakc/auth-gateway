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
  per_page: number;
}

/** Audit log entry */
export interface AuditLogEntry {
  id: string;
  user_id: string;
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
  per_page: number;
}

/** Webhook delivery list response */
export interface WebhookDeliveryListResponse {
  deliveries: WebhookDelivery[];
  total: number;
  page: number;
  per_page: number;
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
export type EmailTemplateType = 'verification' | 'password_reset' | 'welcome' | '2fa' | 'custom';

/** Email template entity */
export interface EmailTemplate extends TimestampedEntity {
  type: EmailTemplateType;
  name: string;
  subject: string;
  html_body: string;
  text_body?: string;
  variables: string[];
  is_active: boolean;
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
export type OAuthProviderType = 'google' | 'github' | 'yandex' | 'telegram' | 'instagram';

/** OAuth provider configuration entity */
export interface OAuthProviderConfig extends TimestampedEntity {
  provider: OAuthProviderType;
  client_id: string;
  client_secret?: string;
  redirect_uris: string[];
  is_enabled: boolean;
}

/** Create OAuth provider request */
export interface CreateOAuthProviderRequest {
  provider: OAuthProviderType;
  client_id: string;
  client_secret: string;
  redirect_uris: string[];
}

/** Update OAuth provider request */
export interface UpdateOAuthProviderRequest {
  client_id?: string;
  client_secret?: string;
  redirect_uris?: string[];
  is_enabled?: boolean;
}

/** OAuth provider list response */
export interface OAuthProviderListResponse {
  providers: OAuthProviderConfig[];
}

// Note: OAuth Client, Scope, and Consent types have been moved to oauth-provider.ts
// for the full OAuth 2.0/OIDC provider implementation.
