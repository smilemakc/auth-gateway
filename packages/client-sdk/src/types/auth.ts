/**
 * Authentication-related types
 */

import type { AccountType, OTPType, ValidationResponse } from './common';
import type { User } from './user';

/** Sign up request */
export interface SignUpRequest {
  email: string;
  username: string;
  password: string;
  fullName?: string;
  phone?: string;
  accountType?: AccountType;
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
  accessToken: string;
  refreshToken: string;
  user: User;
  expiresIn: number;
  requires2FA?: boolean;
  twoFactorToken?: string;
}

/** Refresh token request */
export interface RefreshTokenRequest {
  refreshToken: string;
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
  newPassword: string;
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
  accessToken?: string;
  refreshToken?: string;
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
