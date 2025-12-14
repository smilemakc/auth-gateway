/**
 * Two-Factor Authentication service
 */

import type { HttpClient } from '../core/http';
import type { AuthResponse } from '../types/auth';
import type { MessageResponse } from '../types/common';
import type {
  RegenerateBackupCodesRequest,
  RegenerateBackupCodesResponse,
  TwoFactorDisableRequest,
  TwoFactorLoginVerifyRequest,
  TwoFactorLoginVerifyResponse,
  TwoFactorSetupRequest,
  TwoFactorSetupResponse,
  TwoFactorStatusResponse,
  TwoFactorVerifyRequest,
} from '../types/two-factor';
import { BaseService } from './base';

/** Two-Factor Authentication service */
export class TwoFactorService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Initialize 2FA setup
   * Returns TOTP secret, QR code URL, and backup codes
   * @param data Password for verification
   * @returns Setup response with secret and QR code
   */
  async setup(data: TwoFactorSetupRequest): Promise<TwoFactorSetupResponse> {
    const response = await this.http.post<TwoFactorSetupResponse>(
      '/auth/2fa/setup',
      data
    );
    return response.data;
  }

  /**
   * Verify and enable 2FA after setup
   * @param data TOTP code from authenticator app
   * @returns Success message
   */
  async verify(data: TwoFactorVerifyRequest): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      '/auth/2fa/verify',
      data
    );
    return response.data;
  }

  /**
   * Complete 2FA verification during login
   * @param data Two-factor token and TOTP code
   * @returns Full authentication response with tokens
   */
  async verifyLogin(
    data: TwoFactorLoginVerifyRequest
  ): Promise<TwoFactorLoginVerifyResponse> {
    const response = await this.http.post<AuthResponse>(
      '/auth/2fa/login/verify',
      data,
      { skipAuth: true }
    );

    // Store tokens
    const tokenStorage = this.http.getTokenStorage();
    await tokenStorage.setAccessToken(response.data.access_token);
    await tokenStorage.setRefreshToken(response.data.refresh_token);

    return response.data;
  }

  /**
   * Disable 2FA
   * @param data Password and current TOTP code
   * @returns Success message
   */
  async disable(data: TwoFactorDisableRequest): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      '/auth/2fa/disable',
      data
    );
    return response.data;
  }

  /**
   * Get 2FA status
   * @returns 2FA status including whether enabled and backup codes remaining
   */
  async getStatus(): Promise<TwoFactorStatusResponse> {
    const response = await this.http.get<TwoFactorStatusResponse>(
      '/auth/2fa/status'
    );
    return response.data;
  }

  /**
   * Regenerate backup codes
   * @param data Password for verification
   * @returns New backup codes
   */
  async regenerateBackupCodes(
    data: RegenerateBackupCodesRequest
  ): Promise<RegenerateBackupCodesResponse> {
    const response = await this.http.post<RegenerateBackupCodesResponse>(
      '/auth/2fa/backup-codes/regenerate',
      data
    );
    return response.data;
  }

  /**
   * Check if 2FA is enabled for the current user
   * @returns True if 2FA is enabled
   */
  async isEnabled(): Promise<boolean> {
    const status = await this.getStatus();
    return status.enabled;
  }
}
