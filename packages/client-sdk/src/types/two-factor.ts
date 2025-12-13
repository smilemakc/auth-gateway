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
  qrCodeUrl: string;
  backupCodes: string[];
}

/** 2FA verify request */
export interface TwoFactorVerifyRequest {
  code: string;
}

/** 2FA login verify request */
export interface TwoFactorLoginVerifyRequest {
  twoFactorToken: string;
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
  enabledAt?: string;
  backupCodesRemaining: number;
}

/** Regenerate backup codes request */
export interface RegenerateBackupCodesRequest {
  password: string;
}

/** Regenerate backup codes response */
export interface RegenerateBackupCodesResponse {
  backupCodes: string[];
  message: string;
}
