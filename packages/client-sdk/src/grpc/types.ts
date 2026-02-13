/**
 * gRPC type definitions
 *
 * Proto message types are auto-generated from auth.proto by ts-proto.
 * This file re-exports them and provides SDK-specific types.
 */

// Re-export all generated proto types
export type {
  ValidateTokenRequest,
  ValidateTokenResponse,
  GetUserRequest,
  GetUserResponse,
  User,
  CheckPermissionRequest,
  CheckPermissionResponse,
  IntrospectTokenRequest,
  IntrospectTokenResponse,
  CreateUserRequest,
  CreateUserResponse,
  LoginRequest,
  LoginResponse,
  InitPasswordlessRegistrationRequest,
  InitPasswordlessRegistrationResponse,
  CompletePasswordlessRegistrationRequest,
  CompletePasswordlessRegistrationResponse,
  SendOTPRequest,
  SendOTPResponse,
  VerifyOTPRequest,
  VerifyOTPResponse,
  LoginWithOTPRequest,
  LoginWithOTPResponse,
  VerifyLoginOTPRequest,
  VerifyLoginOTPResponse,
  RegisterWithOTPRequest,
  RegisterWithOTPResponse,
  VerifyRegistrationOTPRequest,
  VerifyRegistrationOTPResponse,
  IntrospectOAuthTokenRequest,
  IntrospectOAuthTokenResponse,
  ValidateOAuthClientRequest,
  ValidateOAuthClientResponse,
  GetOAuthClientRequest,
  GetOAuthClientResponse,
  OAuthClient,
  GetUserAppProfileRequest,
  UserAppProfileResponse,
  GetUserTelegramBotsRequest,
  TelegramBotAccess,
  UserTelegramBotsResponse,
  SendEmailRequest,
  SendEmailResponse,
  SyncUsersRequest,
  SyncUsersResponse,
  SyncUser,
  SyncUserAppProfile,
  GetApplicationAuthConfigRequest,
  GetApplicationAuthConfigResponse,
  CreateTokenExchangeGrpcRequest,
  CreateTokenExchangeGrpcResponse,
  RedeemTokenExchangeGrpcRequest,
  RedeemTokenExchangeGrpcResponse,
} from './generated';

export { OTPType } from './generated';

// Import types used in local aliases
import type {
  User,
  GetApplicationAuthConfigResponse,
  CreateTokenExchangeGrpcResponse,
  RedeemTokenExchangeGrpcResponse,
} from './generated';

// Backward-compatible alias
export type GrpcUser = User;

// ========== SDK-specific types ==========

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
  /** API key (prefix 'agw_') or application secret (prefix 'app_') for gRPC authentication. Sent as x-api-key metadata on every call. */
  apiKey?: string;
}

/** gRPC call options */
export interface GrpcCallOptions {
  /** Call timeout in ms */
  timeout?: number;
  /** Metadata/headers to include */
  metadata?: Record<string, string>;
}

/** Sync users options (SDK convenience wrapper) */
export interface SyncUsersOptions {
  updatedAfter: Date | string;
  applicationId?: string;
  limit?: number;
  offset?: number;
}

/** Auth config result (SDK convenience alias) */
export type AuthConfigResult = GetApplicationAuthConfigResponse;

/** Token exchange result (SDK convenience alias) */
export type TokenExchangeResult = CreateTokenExchangeGrpcResponse;

/** Token exchange redeem result (SDK convenience alias) */
export type TokenExchangeRedeemResult = RedeemTokenExchangeGrpcResponse;
