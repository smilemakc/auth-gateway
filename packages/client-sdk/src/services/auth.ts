/**
 * Authentication service
 */

import type { HttpClient } from '../core/http';
import { TwoFactorRequiredError } from '../core/errors';
import type {
  AuthResponse,
  PasswordResetCompleteRequest,
  PasswordResetRequestRequest,
  RefreshTokenRequest,
  ResendVerificationRequest,
  SignInRequest,
  SignUpRequest,
  TokenValidationResponse,
  VerifyEmailRequest,
} from '../types/auth';
import type {
  ChangePasswordRequest,
  UpdateProfileRequest,
  User,
} from '../types/user';
import type { MessageResponse, EmailMessageResponse, ValidationResponse } from '../types/common';
import { BaseService } from './base';

/** Authentication service for sign up, sign in, token management, and profile */
export class AuthService extends BaseService {
  constructor(http: HttpClient) {
    super(http);
    // Set up token refresh function
    http.setRefreshTokenFn(this.refreshTokenInternal.bind(this));
  }

  /**
   * Register a new user
   * @param data Sign up data
   * @returns Authentication response with tokens and user
   */
  async signUp(data: SignUpRequest): Promise<AuthResponse> {
    const response = await this.http.post<AuthResponse>('/auth/signup', data, {
      skipAuth: true,
    });

    // Store tokens
    await this.storeTokens(response.data);

    return response.data;
  }

  /**
   * Sign in with email/phone and password
   * @param data Sign in credentials
   * @returns Authentication response (may require 2FA)
   * @throws TwoFactorRequiredError if 2FA is required
   */
  async signIn(data: SignInRequest): Promise<AuthResponse> {
    const response = await this.http.post<AuthResponse>('/auth/signin', data, {
      skipAuth: true,
    });

    // Check if 2FA is required
    if (response.data.requires_2fa && response.data.two_factor_token) {
      throw new TwoFactorRequiredError(response.data.two_factor_token);
    }

    // Store tokens
    await this.storeTokens(response.data);

    return response.data;
  }

  /**
   * Refresh access token
   * @param refreshToken Optional refresh token (uses stored token if not provided)
   * @returns New authentication response with tokens
   */
  async refreshToken(refreshToken?: string): Promise<AuthResponse> {
    const tokenStorage = this.http.getTokenStorage();
    const token = refreshToken ?? (await tokenStorage.getRefreshToken());

    if (!token) {
      throw new Error('No refresh token available');
    }

    const response = await this.http.post<AuthResponse>(
      '/auth/refresh',
      { refresh_token: token } satisfies RefreshTokenRequest,
      { skipAuth: true }
    );

    // Store new tokens
    await this.storeTokens(response.data);

    return response.data;
  }

  /** Internal refresh token function for auto-refresh */
  private async refreshTokenInternal(): Promise<{
    accessToken: string;
    refreshToken: string;
  }> {
    const result = await this.refreshToken();
    return {
      accessToken: result.access_token,
      refreshToken: result.refresh_token,
    };
  }

  /**
   * Log out the current user
   * @returns Success message
   */
  async logout(): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>('/auth/logout');

    // Clear stored tokens
    const tokenStorage = this.http.getTokenStorage();
    await tokenStorage.clear();

    return response.data;
  }

  /**
   * Get current user profile
   * @returns User profile
   */
  async getProfile(): Promise<User> {
    const response = await this.http.get<User>('/auth/profile');
    return response.data;
  }

  /**
   * Update current user profile
   * @param data Profile update data
   * @returns Updated user profile
   */
  async updateProfile(data: UpdateProfileRequest): Promise<User> {
    const response = await this.http.put<User>('/auth/profile', data);
    return response.data;
  }

  /**
   * Change password
   * @param data Old and new password
   * @returns Success message
   */
  async changePassword(data: ChangePasswordRequest): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      '/auth/change-password',
      data
    );
    return response.data;
  }

  /**
   * Verify email address with code
   * @param data Email and verification code
   * @returns Validation result
   */
  async verifyEmail(data: VerifyEmailRequest): Promise<ValidationResponse> {
    const response = await this.http.post<ValidationResponse>(
      '/auth/verify/email',
      data,
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Resend email verification code
   * @param data Email address
   * @returns Success message
   */
  async resendVerification(
    data: ResendVerificationRequest
  ): Promise<EmailMessageResponse> {
    const response = await this.http.post<EmailMessageResponse>(
      '/auth/verify/resend',
      data,
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Request password reset
   * @param data Email address
   * @returns Success message
   */
  async requestPasswordReset(
    data: PasswordResetRequestRequest
  ): Promise<EmailMessageResponse> {
    const response = await this.http.post<EmailMessageResponse>(
      '/auth/password/reset/request',
      data,
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Complete password reset
   * @param data Email, code, and new password
   * @returns Success message
   */
  async completePasswordReset(
    data: PasswordResetCompleteRequest
  ): Promise<MessageResponse> {
    const response = await this.http.post<MessageResponse>(
      '/auth/password/reset/complete',
      data,
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Set tokens manually (useful for SSR or external auth)
   * @param accessToken Access token
   * @param refreshToken Refresh token
   */
  async setTokens(accessToken: string, refreshToken: string): Promise<void> {
    const tokenStorage = this.http.getTokenStorage();
    await tokenStorage.setAccessToken(accessToken);
    await tokenStorage.setRefreshToken(refreshToken);
  }

  /**
   * Get current access token
   * @returns Access token or null
   */
  async getAccessToken(): Promise<string | null> {
    return this.http.getTokenStorage().getAccessToken();
  }

  /**
   * Check if user is authenticated (has valid tokens)
   * @returns True if tokens exist
   */
  async isAuthenticated(): Promise<boolean> {
    const token = await this.http.getTokenStorage().getAccessToken();
    return token !== null;
  }

  /**
   * Clear all stored tokens
   */
  async clearTokens(): Promise<void> {
    await this.http.getTokenStorage().clear();
  }

  /**
   * Validate a token (JWT or API key)
   * This is useful for external services to validate tokens
   * @param accessToken The token to validate
   * @returns Token validation result
   */
  async validateToken(accessToken: string): Promise<TokenValidationResponse> {
    const response = await this.http.post<TokenValidationResponse>(
      '/v1/token/validate',
      { access_token: accessToken },
      { skipAuth: true }
    );
    return response.data;
  }

  /** Store tokens from auth response */
  private async storeTokens(response: AuthResponse): Promise<void> {
    const tokenStorage = this.http.getTokenStorage();
    await tokenStorage.setAccessToken(response.access_token);
    await tokenStorage.setRefreshToken(response.refresh_token);
  }
}
