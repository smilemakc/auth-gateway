/**
 * SMS service for phone-based authentication
 */

import type { HttpClient } from '../core/http';
import type { OTPType } from '../types/common';
import type {
  SendSMSRequest,
  SendSMSResponse,
  SMSStatsResponse,
  VerifySMSRequest,
  VerifySMSResponse,
} from '../types/sms';
import { BaseService } from './base';

/** SMS service for phone verification and authentication */
export class SMSService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Send SMS OTP code
   * @param phone Phone number in E.164 format (e.g., +1234567890)
   * @param type OTP type (verification, password_reset, 2fa, login)
   * @returns SMS send response with message ID and expiration
   */
  async send(phone: string, type: OTPType): Promise<SendSMSResponse> {
    const response = await this.http.post<SendSMSResponse>(
      '/sms/send',
      { phone, type } satisfies SendSMSRequest,
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Verify SMS OTP code
   * @param phone Phone number
   * @param code OTP code from SMS
   * @returns Verification result (may include tokens for login type)
   */
  async verify(phone: string, code: string): Promise<VerifySMSResponse> {
    const response = await this.http.post<VerifySMSResponse>(
      '/sms/verify',
      { phone, code } satisfies VerifySMSRequest,
      { skipAuth: true }
    );

    // Store tokens if returned (login SMS)
    if (response.data.accessToken && response.data.refreshToken) {
      const tokenStorage = this.http.getTokenStorage();
      await tokenStorage.setAccessToken(response.data.accessToken);
      await tokenStorage.setRefreshToken(response.data.refreshToken);
    }

    return response.data;
  }

  /**
   * Get SMS statistics (admin only)
   * @returns SMS statistics including delivery rates
   */
  async getStats(): Promise<SMSStatsResponse> {
    const response = await this.http.get<SMSStatsResponse>('/sms/stats');
    return response.data;
  }
}
