/**
 * Admin-related types
 */

import type { IPFilterType, TimestampedEntity } from './common';

/** IP Filter entity */
export interface IPFilter extends TimestampedEntity {
  ipAddress: string;
  type: IPFilterType;
  description?: string;
}

/** Create IP filter request */
export interface CreateIPFilterRequest {
  ipAddress: string;
  type: IPFilterType;
  description?: string;
}

/** IP filter list response */
export interface IPFilterListResponse {
  filters: IPFilter[];
  total: number;
  page: number;
  perPage: number;
}

/** Audit log entry */
export interface AuditLogEntry {
  id: string;
  userId: string;
  action: string;
  resource: string;
  status: 'success' | 'failure';
  ipAddress: string;
  userAgent: string;
  createdAt: string;
  details?: Record<string, unknown>;
}

/** Audit log list response */
export interface AuditLogListResponse {
  logs: AuditLogEntry[];
  total: number;
  page: number;
  pageSize: number;
}

/** Theme settings */
export interface ThemeSettings {
  primaryColor: string;
  secondaryColor: string;
  backgroundColor: string;
}

/** Branding settings */
export interface BrandingSettings extends TimestampedEntity {
  logoUrl?: string;
  faviconUrl?: string;
  theme: ThemeSettings;
  companyName: string;
  supportEmail?: string;
  termsUrl?: string;
  privacyUrl?: string;
}

/** Public branding response (no sensitive data) */
export interface PublicBrandingResponse {
  logoUrl?: string;
  faviconUrl?: string;
  theme: ThemeSettings;
  companyName: string;
  supportEmail?: string;
  termsUrl?: string;
  privacyUrl?: string;
}

/** Update branding request */
export interface UpdateBrandingRequest {
  logoUrl?: string;
  faviconUrl?: string;
  primaryColor?: string;
  secondaryColor?: string;
  backgroundColor?: string;
  companyName?: string;
  supportEmail?: string;
  termsUrl?: string;
  privacyUrl?: string;
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
  countryCode: string;
  countryName: string;
  city: string;
  loginCount: number;
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
