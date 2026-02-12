/**
 * OAuth service for social authentication
 */

import type { HttpClient } from '../core/http';
import type { OAuthProvider } from '../types/common';
import type {
  OAuthLoginResponse,
  OAuthProviderInfo,
  TelegramCallbackData,
} from '../types/oauth';
import { BaseService } from './base';

/** OAuth service for social login */
export class OAuthService extends BaseService {
  private baseUrl: string;

  constructor(http: HttpClient, baseUrl: string) {
    super(http);
    this.baseUrl = baseUrl.replace(/\/$/, '');
  }

  /**
   * Get list of available OAuth providers
   * @returns List of OAuth provider info
   */
  async getProviders(): Promise<OAuthProviderInfo[]> {
    const response = await this.http.get<OAuthProviderInfo[]>(
      '/api/auth/providers',
      { skipAuth: true }
    );
    return response.data;
  }

  /**
   * Get enabled OAuth providers only
   * @returns List of enabled OAuth providers
   */
  async getEnabledProviders(): Promise<OAuthProviderInfo[]> {
    const providers = await this.getProviders();
    return providers.filter((p) => p.enabled);
  }

  /**
   * Get OAuth authorization URL for a provider
   * This URL should be used to redirect the user to the OAuth provider
   * @param provider OAuth provider name
   * @returns Authorization URL
   */
  getAuthorizationUrl(provider: OAuthProvider): string {
    return `${this.baseUrl}/api/auth/${provider}`;
  }

  /**
   * Open OAuth authorization in a popup window
   * @param provider OAuth provider name
   * @param options Popup window options
   * @returns Promise that resolves with OAuth response when popup completes
   */
  openAuthPopup(
    provider: OAuthProvider,
    options?: {
      width?: number;
      height?: number;
    }
  ): Promise<OAuthLoginResponse> {
    const width = options?.width ?? 600;
    const height = options?.height ?? 700;
    const left = window.screenX + (window.outerWidth - width) / 2;
    const top = window.screenY + (window.outerHeight - height) / 2;

    const url = this.getAuthorizationUrl(provider);
    const popup = window.open(
      url,
      'oauth_popup',
      `width=${width},height=${height},left=${left},top=${top},popup=true`
    );

    if (!popup) {
      return Promise.reject(new Error('Failed to open OAuth popup'));
    }

    return new Promise((resolve, reject) => {
      const checkClosed = setInterval(() => {
        if (popup.closed) {
          clearInterval(checkClosed);
          reject(new Error('OAuth popup was closed'));
        }
      }, 500);

      // Listen for message from popup
      const handleMessage = (event: MessageEvent): void => {
        // Verify origin
        if (!event.origin.includes(new URL(this.baseUrl).hostname)) {
          return;
        }

        if (event.data?.type === 'oauth_callback') {
          clearInterval(checkClosed);
          window.removeEventListener('message', handleMessage);
          popup.close();

          if (event.data.error) {
            reject(new Error(event.data.error));
          } else {
            resolve(event.data.result as OAuthLoginResponse);
          }
        }
      };

      window.addEventListener('message', handleMessage);

      // Cleanup on timeout (5 minutes)
      setTimeout(() => {
        clearInterval(checkClosed);
        window.removeEventListener('message', handleMessage);
        if (!popup.closed) {
          popup.close();
        }
        reject(new Error('OAuth timeout'));
      }, 5 * 60 * 1000);
    });
  }

  /**
   * Handle OAuth callback (for SPA routing)
   * Call this when your app receives the callback redirect
   * @param provider OAuth provider name
   * @param code Authorization code from callback
   * @param state CSRF state from callback
   * @returns OAuth login response
   */
  async handleCallback(
    provider: OAuthProvider,
    code: string,
    state: string
  ): Promise<OAuthLoginResponse> {
    const response = await this.http.get<OAuthLoginResponse>(
      `/api/auth/${provider}/callback`,
      {
        query: {
          code,
          state,
          response_type: 'json',
        },
        skipAuth: true,
      }
    );

    // Store tokens
    if (response.data.access_token) {
      const tokenStorage = this.http.getTokenStorage();
      await tokenStorage.setAccessToken(response.data.access_token);
      await tokenStorage.setRefreshToken(response.data.refresh_token);
    }

    return response.data;
  }

  /**
   * Handle Telegram authentication callback
   * @param data Telegram callback data
   * @returns OAuth login response
   */
  async handleTelegramCallback(
    data: TelegramCallbackData
  ): Promise<OAuthLoginResponse> {
    const response = await this.http.post<OAuthLoginResponse>(
      '/api/auth/telegram/callback',
      data,
      { skipAuth: true }
    );

    // Store tokens
    if (response.data.access_token) {
      const tokenStorage = this.http.getTokenStorage();
      await tokenStorage.setAccessToken(response.data.access_token);
      await tokenStorage.setRefreshToken(response.data.refresh_token);
    }

    return response.data;
  }

  /**
   * Generate Telegram login widget script URL
   * @param botUsername Telegram bot username
   * @param callbackUrl Callback URL for Telegram login
   * @returns Script URL for Telegram widget
   */
  getTelegramWidgetUrl(botUsername: string): string {
    return `https://telegram.org/js/telegram-widget.js?22`;
  }

  /**
   * Parse OAuth callback from URL (for handling redirects)
   * @param url Current URL with callback parameters
   * @returns Parsed callback parameters or null if not a callback
   */
  parseCallbackUrl(url: string): {
    provider: OAuthProvider;
    code: string;
    state: string;
  } | null {
    const urlObj = new URL(url);
    const pathParts = urlObj.pathname.split('/');

    // Check if this is an OAuth callback URL
    // Expected format: /auth/{provider}/callback
    const authIndex = pathParts.indexOf('auth');
    if (
      authIndex === -1 ||
      pathParts[authIndex + 2] !== 'callback'
    ) {
      return null;
    }

    const provider = pathParts[authIndex + 1] as OAuthProvider;
    const code = urlObj.searchParams.get('code');
    const state = urlObj.searchParams.get('state');

    if (!code || !state) {
      return null;
    }

    return { provider, code, state };
  }
}
