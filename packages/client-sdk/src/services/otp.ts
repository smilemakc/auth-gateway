/**
 * OTP (One-Time Password) service
 */

import type { HttpClient } from '../core/http';
import type { SendOTPRequest, VerifyOTPRequest, VerifyOTPResponse } from '../types/auth';
import type { EmailMessageResponse, PhoneMessageResponse } from '../types/common';
import { BaseService } from './base';

/** OTP service for email and phone verification */
export class OTPService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Send OTP code to email
   * @param email Email address
   * @param type OTP type (verification, password_reset, 2fa, login)
   * @returns Success message
   */
  async sendToEmail(
    email: string,
    type: SendOTPRequest['type']
  ): Promise<EmailMessageResponse> {
    const response = await this.http.post<EmailMessageResponse>(
      '/otp/send',
      { email, type } satisfies SendOTPRequest,
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Send OTP code to phone
   * @param phone Phone number
   * @param type OTP type (verification, password_reset, 2fa, login)
   * @returns Success message
   */
  async sendToPhone(
    phone: string,
    type: SendOTPRequest['type']
  ): Promise<PhoneMessageResponse> {
    const response = await this.http.post<PhoneMessageResponse>(
      '/otp/send',
      { phone, type } satisfies SendOTPRequest,
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Verify OTP code for email
   * @param email Email address
   * @param code OTP code
   * @param type OTP type
   * @returns Verification result (may include tokens for login type)
   */
  async verifyEmail(
    email: string,
    code: string,
    type: VerifyOTPRequest['type']
  ): Promise<VerifyOTPResponse> {
    const response = await this.http.post<VerifyOTPResponse>(
      '/otp/verify',
      { email, code, type } satisfies VerifyOTPRequest,
      { skipAuth: true }
    );

    // Store tokens if returned (login OTP)
    if (response.data.access_token && response.data.refresh_token) {
      const tokenStorage = this.http.getTokenStorage();
      await tokenStorage.setAccessToken(response.data.access_token);
      await tokenStorage.setRefreshToken(response.data.refresh_token);
    }

    return response.data;
  }

  /**
   * Verify OTP code for phone
   * @param phone Phone number
   * @param code OTP code
   * @param type OTP type
   * @returns Verification result (may include tokens for login type)
   */
  async verifyPhone(
    phone: string,
    code: string,
    type: VerifyOTPRequest['type']
  ): Promise<VerifyOTPResponse> {
    const response = await this.http.post<VerifyOTPResponse>(
      '/otp/verify',
      { phone, code, type } satisfies VerifyOTPRequest,
      { skipAuth: true }
    );

    // Store tokens if returned (login OTP)
    if (response.data.access_token && response.data.refresh_token) {
      const tokenStorage = this.http.getTokenStorage();
      await tokenStorage.setAccessToken(response.data.access_token);
      await tokenStorage.setRefreshToken(response.data.refresh_token);
    }

    return response.data;
  }
}
