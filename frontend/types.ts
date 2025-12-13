/**
 * @deprecated Use RoleDefinition and check by ID instead
 */
export enum UserRole {
  ADMIN = 'admin',
  MODERATOR = 'moderator',
  USER = 'user'
}

export interface RoleInfo {
  id: string;
  name: string;
  displayName: string;
}

export interface User {
  id: string;
  email: string;
  username: string;
  phone?: string;
  fullName: string;
  roles: RoleInfo[];
  isActive: boolean;
  isEmailVerified: boolean;
  is2FAEnabled: boolean;
  avatarUrl?: string;
  createdAt: string;
  lastLogin?: string;
}

export interface UserSession {
  id: string;
  userId: string;
  deviceType: 'mobile' | 'desktop' | 'tablet';
  os: string;
  browser: string;
  ipAddress: string;
  location: string;
  lastActive: string;
  isCurrent: boolean;
}

export interface ApiKey {
  id: string;
  name: string;
  prefix: string;
  ownerId: string;
  ownerName: string;
  scopes: string[];
  status: 'active' | 'revoked';
  lastUsed?: string;
  createdAt: string;
  expiresAt?: string;
}

export interface AuditLog {
  id: string;
  action: string;
  userId?: string;
  userEmail?: string; // For display convenience
  resource: string;
  ip: string;
  status: 'success' | 'failed' | 'blocked';
  timestamp: string;
  details?: string;
}

export interface DashboardStats {
  totalUsers: number;
  activeUsers: number;
  usersWith2FA: number;
  totalApiKeys: number;
  recentRegistrations: { date: string; count: number }[];
  loginActivity: { date: string; success: number; failed: number }[];
}

export interface OAuthAccount {
  id: string;
  provider: 'google' | 'github' | 'yandex' | 'telegram';
  userId: string;
  userName: string;
  connectedAt: string;
}

export interface OAuthProviderConfig {
  id: string;
  provider: 'google' | 'github' | 'yandex' | 'telegram' | 'instagram';
  clientId: string;
  clientSecret: string;
  redirectUris: string[];
  isEnabled: boolean;
  createdAt: string;
}

export interface EmailTemplate {
  id: string;
  type: 'verification' | 'reset_password' | 'welcome' | 'magic_link';
  name: string;
  subject: string;
  bodyHtml: string;
  variables: string[];
  lastUpdated: string;
}

export interface Permission {
  id: string;
  name: string;
  resource: string; // e.g., 'users', 'api_keys'
  action: string;   // e.g., 'read', 'write', 'delete'
  description: string;
}

export interface RoleDefinition {
  id: string;
  name: string;
  description: string;
  isSystem: boolean; // Cannot be deleted if true
  permissions: string[]; // List of Permission IDs
  userCount: number;
  createdAt: string;
}

export interface IpRule {
  id: string;
  type: 'whitelist' | 'blacklist';
  ipAddress: string; // CIDR or single IP
  description?: string;
  createdAt: string;
  createdBy: string;
}

export interface WebhookEndpoint {
  id: string;
  url: string;
  description?: string;
  events: string[];
  secret: string;
  isActive: boolean;
  failureCount: number;
  lastTriggeredAt?: string;
  createdAt: string;
}

export interface ServiceAccount {
  id: string;
  name: string;
  description: string;
  clientId: string;
  // clientSecret is not stored here for security, but returned once on creation
  isActive: boolean;
  createdAt: string;
  lastUsedAt?: string;
}

export interface BrandingConfig {
  companyName: string;
  logoUrl: string;
  faviconUrl: string;
  primaryColor: string;
  accentColor: string;
  backgroundColor: string;
  backgroundImageUrl?: string;
  loginPageTitle: string;
  loginPageSubtitle: string;
  showSocialLogins: boolean;
  customCss?: string;
}

export type SmsProviderType = 'aws' | 'twilio' | 'mock';

export interface SmsConfig {
  provider: SmsProviderType;
  awsRegion?: string;
  awsAccessKeyId?: string;
  awsSecretAccessKey?: string;
  twilioAccountSid?: string;
  twilioAuthToken?: string;
  twilioPhoneNumber?: string;
}

export interface SystemStatus {
  status: 'healthy' | 'degraded' | 'down';
  database: 'connected' | 'disconnected';
  redis: 'connected' | 'disconnected';
  uptime: number; // seconds
  version: string;
  maintenanceMode: boolean;
  maintenanceMessage: string;
}

export interface PasswordPolicy {
  minLength: number;
  requireUppercase: boolean;
  requireLowercase: boolean;
  requireNumbers: boolean;
  requireSpecial: boolean;
  historyCount: number; // Number of previous passwords to remember
  expiryDays: number; // 0 for no expiry
  jwtTtlMinutes: number;
  refreshTtlDays: number;
}