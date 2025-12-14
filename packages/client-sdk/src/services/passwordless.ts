/**
 * Passwordless authentication service
 */

import type { HttpClient } from '../core/http';
import type { AuthResponse, PasswordlessRequestRequest, PasswordlessVerifyRequest } from '../types/auth';
import type { EmailMessageResponse } from '../types/common';
import { BaseService } from './base';

/** Passwordless authentication service (magic link / email code) */
export class PasswordlessService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
  }

  /**
   * Request passwordless login code
   * @param email Email address
   * @returns Success message
   */
  async request(email: string): Promise<EmailMessageResponse> {
    const response = await this.http.post<EmailMessageResponse>(
      '/auth/passwordless/request',
      { email } satisfies PasswordlessRequestRequest,
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Verify passwordless login code
   * @param email Email address
   * @param code Login code from email
   * @returns Authentication response with tokens
   */
  async verify(email: string, code: string): Promise<AuthResponse> {
    const response = await this.http.post<AuthResponse>(
      '/auth/passwordless/verify',
      { email, code } satisfies PasswordlessVerifyRequest,
      { skipAuth: true }
    );

    // Store tokens
    const tokenStorage = this.http.getTokenStorage();
    await tokenStorage.setAccessToken(response.data.access_token);
    await tokenStorage.setRefreshToken(response.data.refresh_token);

    return response.data;
  }

  /**
   * Complete passwordless login flow
   * Combines request and verify steps
   * @param email Email address
   * @param getCode Function that returns the code (e.g., from user input)
   * @returns Authentication response
   */
  async login(
    email: string,
    getCode: () => Promise<string>
  ): Promise<AuthResponse> {
    // Request the code
    await this.request(email);

    // Get code from user
    const code = await getCode();

    // Verify and complete login
    return this.verify(email, code);
  }
}
