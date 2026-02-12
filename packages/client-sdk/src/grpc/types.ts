/**
 * gRPC type definitions
 */

/** Role information */
export interface RoleInfo {
  id: string;
  name: string;
  displayName: string;
}

/** Validate token request */
export interface ValidateTokenRequest {
  accessToken: string;
}

/** Validate token response */
export interface ValidateTokenResponse {
  valid: boolean;
  userId: string;
  email: string;
  username: string;
  roles: string[];
  errorMessage: string;
  expiresAt: number;
}

/** Get user request */
export interface GetUserRequest {
  userId: string;
}

/** gRPC User object */
export interface GrpcUser {
  id: string;
  email: string;
  username: string;
  fullName: string;
  profilePictureUrl: string;
  roles: RoleInfo[];
  emailVerified: boolean;
  isActive: boolean;
  createdAt: number;
  updatedAt: number;
}

/** Get user response */
export interface GetUserResponse {
  user: GrpcUser | null;
  errorMessage: string;
}

/** Check permission request */
export interface CheckPermissionRequest {
  userId: string;
  resource: string;
  action: string;
}

/** Check permission response */
export interface CheckPermissionResponse {
  allowed: boolean;
  roles: string[];
  errorMessage: string;
}

/** Introspect token request */
export interface IntrospectTokenRequest {
  accessToken: string;
}

/** Introspect token response */
export interface IntrospectTokenResponse {
  active: boolean;
  userId: string;
  email: string;
  username: string;
  roles: string[];
  issuedAt: number;
  expiresAt: number;
  notBefore: number;
  subject: string;
  blacklisted: boolean;
  errorMessage: string;
}

/** gRPC client configuration */
export interface GrpcClientConfig {
  /** gRPC server address (e.g., 'localhost:50051') */
  address: string;
  /** Enable TLS (default: false for local development) */
  useTls?: boolean;
  /** Path to CA certificate file (for TLS) */
  caCertPath?: string;
  /** Connection timeout in ms (default: 5000) */
  timeout?: number;
  /** Enable debug logging */
  debug?: boolean;
}

/** gRPC call options */
export interface GrpcCallOptions {
  /** Call timeout in ms */
  timeout?: number;
  /** Metadata/headers to include */
  metadata?: Record<string, string>;
}

/** Sync users options */
export interface SyncUsersOptions {
  updatedAfter: Date | string;
  applicationId?: string;
  limit?: number;
  offset?: number;
}

/** Sync user object */
export interface SyncUser {
  id: string;
  email: string;
  username: string;
  fullName: string;
  isActive: boolean;
  emailVerified: boolean;
  updatedAt: string;
  appProfile?: SyncUserAppProfile;
}

/** Sync user app profile */
export interface SyncUserAppProfile {
  displayName?: string;
  avatarUrl?: string;
  appRoles?: string[];
  isActive: boolean;
  isBanned: boolean;
}

/** Sync users result */
export interface SyncUsersResult {
  users: SyncUser[];
  total: number;
  hasMore: boolean;
  syncTimestamp: string;
}

/** Auth config result */
export interface AuthConfigResult {
  applicationId: string;
  name: string;
  displayName: string;
  allowedAuthMethods: string[];
  oauthProviders: string[];
}

/** Token exchange result */
export interface TokenExchangeResult {
  exchangeCode: string;
  expiresAt: string;
  redirectUrl?: string;
}

/** Token exchange redeem result */
export interface TokenExchangeRedeemResult {
  accessToken: string;
  refreshToken: string;
  userId: string;
  email: string;
  applicationId: string;
}
