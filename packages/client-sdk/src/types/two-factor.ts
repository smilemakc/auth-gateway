/**
 * Two-Factor Authentication types
 */

import type { AuthResponse } from './auth';

/** 2FA setup request */
export interface TwoFactorSetupRequest {
  password: string;
}

/** 2FA setup response */
export interface TwoFactorSetupResponse {
  secret: string;
  qr_code_url: string;
  backup_codes: string[];
}

/** 2FA verify request */
export interface TwoFactorVerifyRequest {
  code: string;
}

/** 2FA login verify request */
export interface TwoFactorLoginVerifyRequest {
  two_factor_token: string;
  code: string;
}

/** 2FA login verify response (same as AuthResponse) */
export type TwoFactorLoginVerifyResponse = AuthResponse;

/** 2FA disable request */
export interface TwoFactorDisableRequest {
  password: string;
  code: string;
}

/** 2FA status response */
export interface TwoFactorStatusResponse {
  enabled: boolean;
  enabled_at?: string;
  backup_codes_remaining: number;
}

/** Regenerate backup codes request */
export interface RegenerateBackupCodesRequest {
  password: string;
}

/** Regenerate backup codes response */
export interface RegenerateBackupCodesResponse {
  backup_codes: string[];
  message: string;
}
