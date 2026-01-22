/**
 * Authentication-related types
 */

import type { AccountType, OTPType, ValidationResponse } from './common';
import type { User } from './user';

/** JWT token claims */
export interface JWTClaims {
  user_id: string;
  email: string;
  username: string;
  roles: string[];
  exp: number;
  iat: number;
}

/** Sign up request */
export interface SignUpRequest {
  email: string;
  username: string;
  password: string;
  full_name?: string;
  phone?: string;
  account_type?: AccountType;
}

/** Sign in request (email-based) */
export interface SignInEmailRequest {
  email: string;
  password: string;
}

/** Sign in request (phone-based) */
export interface SignInPhoneRequest {
  phone: string;
  password: string;
}

/** Sign in request (unified) */
export type SignInRequest = SignInEmailRequest | SignInPhoneRequest;

/** Authentication response */
export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
  expires_in: number;
  requires_2fa?: boolean;
  two_factor_token?: string;
}

/** Refresh token request */
export interface RefreshTokenRequest {
  refresh_token: string;
}

/** Email verification request */
export interface VerifyEmailRequest {
  email: string;
  code: string;
}

/** Resend verification request */
export interface ResendVerificationRequest {
  email: string;
}

/** Password reset request */
export interface PasswordResetRequestRequest {
  email: string;
}

/** Password reset complete request */
export interface PasswordResetCompleteRequest {
  email: string;
  code: string;
  new_password: string;
}

/** Send OTP request */
export interface SendOTPRequest {
  email?: string;
  phone?: string;
  type: OTPType;
}

/** Verify OTP request */
export interface VerifyOTPRequest {
  email?: string;
  phone?: string;
  code: string;
  type: OTPType;
}

/** Verify OTP response */
export interface VerifyOTPResponse extends ValidationResponse {
  access_token?: string;
  refresh_token?: string;
  user?: User;
}

/** Passwordless login request */
export interface PasswordlessRequestRequest {
  email: string;
}

/** Passwordless verify request */
export interface PasswordlessVerifyRequest {
  email: string;
  code: string;
}

/** Token validation request */
export interface TokenValidationRequest {
  access_token: string;
}

/** Token validation response */
export interface TokenValidationResponse {
  valid: boolean;
  user_id?: string;
  email?: string;
  username?: string;
  roles?: string[];
  expires_at?: number;
  is_active?: boolean;
  error_message?: string;
}
