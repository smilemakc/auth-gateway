/**
 * OAuth 2.0 / OIDC provider types
 */

import type { TimestampedEntity, ListResponse } from './common';

// OAuth Client Types

export type ClientType = 'confidential' | 'public';

export type GrantType =
  | 'authorization_code'
  | 'client_credentials'
  | 'refresh_token'
  | 'urn:ietf:params:oauth:grant-type:device_code';

export interface OAuthClient extends TimestampedEntity {
  client_id: string;
  name: string;
  description?: string;
  logo_url?: string;
  client_type: ClientType;
  redirect_uris: string[];
  allowed_grant_types: GrantType[];
  allowed_scopes: string[];
  default_scopes: string[];
  access_token_ttl: number;
  refresh_token_ttl: number;
  id_token_ttl: number;
  require_pkce: boolean;
  require_consent: boolean;
  first_party: boolean;
  is_active: boolean;
  owner_id?: string;
}

export interface CreateOAuthClientRequest {
  name: string;
  description?: string;
  logo_url?: string;
  client_type?: ClientType;
  redirect_uris?: string[];
  allowed_grant_types?: GrantType[];
  allowed_scopes?: string[];
  default_scopes?: string[];
  access_token_ttl?: number;
  refresh_token_ttl?: number;
  id_token_ttl?: number;
  require_pkce?: boolean;
  require_consent?: boolean;
  first_party?: boolean;
}

export interface CreateOAuthClientResponse {
  client: OAuthClient;
  client_secret: string;
}

export interface UpdateOAuthClientRequest {
  name?: string;
  description?: string;
  logo_url?: string;
  redirect_uris?: string[];
  allowed_grant_types?: GrantType[];
  allowed_scopes?: string[];
  default_scopes?: string[];
  access_token_ttl?: number;
  refresh_token_ttl?: number;
  id_token_ttl?: number;
  require_pkce?: boolean;
  require_consent?: boolean;
  first_party?: boolean;
  is_active?: boolean;
}

export interface RotateSecretResponse {
  client_secret: string;
}

// OAuth Scope Types

export interface OAuthScope extends TimestampedEntity {
  name: string;
  display_name: string;
  description?: string;
  is_default: boolean;
  is_system: boolean;
}

export interface CreateScopeRequest {
  name: string;
  display_name: string;
  description?: string;
}

// Token Types

export interface TokenResponse {
  access_token: string;
  token_type: 'Bearer';
  expires_in: number;
  refresh_token?: string;
  id_token?: string;
  scope?: string;
}

export interface TokenIntrospectionResponse {
  active: boolean;
  scope?: string;
  client_id?: string;
  username?: string;
  token_type?: string;
  exp?: number;
  iat?: number;
  nbf?: number;
  sub?: string;
  aud?: string;
  iss?: string;
  jti?: string;
}

// Device Flow Types

export interface DeviceAuthRequest {
  client_id: string;
  scope?: string;
}

export interface DeviceAuthResponse {
  device_code: string;
  user_code: string;
  verification_uri: string;
  verification_uri_complete?: string;
  expires_in: number;
  interval: number;
}

// User Consent Types

export interface UserConsent extends TimestampedEntity {
  user_id: string;
  client_id: string;
  scopes: string[];
  revoked_at?: string;
}

// OIDC Discovery Types

export interface OIDCDiscoveryDocument {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint: string;
  jwks_uri: string;
  revocation_endpoint: string;
  introspection_endpoint: string;
  device_authorization_endpoint: string;
  scopes_supported: string[];
  response_types_supported: string[];
  grant_types_supported: string[];
  subject_types_supported: string[];
  id_token_signing_alg_values_supported: string[];
  token_endpoint_auth_methods_supported: string[];
  code_challenge_methods_supported: string[];
  claims_supported: string[];
}

// JWKS Types

export interface JWK {
  kty: string;
  use: string;
  alg: string;
  kid: string;
  n?: string;
  e?: string;
  crv?: string;
  x?: string;
  y?: string;
}

export interface JWKS {
  keys: JWK[];
}

// UserInfo Response (OIDC)

export interface UserInfoResponse {
  sub: string;
  name?: string;
  given_name?: string;
  family_name?: string;
  preferred_username?: string;
  picture?: string;
  email?: string;
  email_verified?: boolean;
  phone_number?: string;
  phone_number_verified?: boolean;
  updated_at?: number;
}

// ID Token Claims

export interface IDTokenClaims {
  iss: string;
  sub: string;
  aud: string;
  exp: number;
  iat: number;
  auth_time?: number;
  nonce?: string;
  acr?: string;
  amr?: string[];
  azp?: string;
  name?: string;
  given_name?: string;
  family_name?: string;
  preferred_username?: string;
  picture?: string;
  email?: string;
  email_verified?: boolean;
}

// PKCE Types

export interface PKCEParams {
  code_verifier: string;
  code_challenge: string;
  code_challenge_method: 'S256' | 'plain';
}

// Authorization URL Options

export interface AuthorizationUrlOptions {
  scope?: string;
  state?: string;
  nonce?: string;
  prompt?: 'none' | 'login' | 'consent' | 'select_account';
  login_hint?: string;
  acr_values?: string;
}

export interface AuthorizationUrlResult {
  url: string;
  state: string;
  nonce?: string;
  codeVerifier?: string;
}

// OAuth Error

export interface OAuthError {
  error: string;
  error_description?: string;
  error_uri?: string;
}

// List Response wrappers

export interface OAuthClientListResponse extends ListResponse<OAuthClient> {
  clients: OAuthClient[];
}

export interface OAuthScopeListResponse {
  scopes: OAuthScope[];
}

export interface UserConsentListResponse {
  consents: UserConsent[];
}
