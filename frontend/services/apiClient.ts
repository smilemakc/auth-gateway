import {
  createClient,
  type AuthGatewayClient,
  type TokenStorage,
} from '@auth-gateway/client-sdk';

// Get base URL from environment
const baseUrl = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api`;

// Custom event for auth failure - allows React components to react to auth failures
export const AUTH_FAILURE_EVENT = 'auth:failure';

export function dispatchAuthFailure() {
  window.dispatchEvent(new CustomEvent(AUTH_FAILURE_EVENT));
}

// Custom token storage with debugging
const customTokenStorage: TokenStorage = {
  async getAccessToken(): Promise<string | null> {
    const token = localStorage.getItem('auth_gateway_access_token');
    console.log('[TokenStorage] Getting access token:', token ? 'exists' : 'null');
    return token;
  },

  async getRefreshToken(): Promise<string | null> {
    const token = localStorage.getItem('auth_gateway_refresh_token');
    console.log('[TokenStorage] Getting refresh token:', token ? 'exists' : 'null');
    return token;
  },

  async setAccessToken(token: string): Promise<void> {
    console.log('[TokenStorage] Storing access token:', !!token);
    localStorage.setItem('auth_gateway_access_token', token);
  },

  async setRefreshToken(token: string): Promise<void> {
    console.log('[TokenStorage] Storing refresh token:', !!token);
    localStorage.setItem('auth_gateway_refresh_token', token);
  },

  async clear(): Promise<void> {
    console.log('[TokenStorage] Clearing tokens');
    localStorage.removeItem('auth_gateway_access_token');
    localStorage.removeItem('auth_gateway_refresh_token');
  },
};

// Create the API client
export const apiClient: AuthGatewayClient = createClient({
  baseUrl,
  tokenStorage: customTokenStorage,
  autoRefreshTokens: true,
  timeout: 30000,
  debug: import.meta.env.DEV,

  callbacks: {
    onAuthFailure: () => {
      console.error('[Auth] Authentication failed - clearing tokens and dispatching auth failure event');
      // Clear all auth-related items
      localStorage.removeItem('auth_gateway_access_token');
      localStorage.removeItem('auth_gateway_refresh_token');
      localStorage.removeItem('auth_token');
      // Dispatch custom event so AuthContext can update its state
      dispatchAuthFailure();
    },

    onTokenRefresh: (tokens) => {
      console.log('[Auth] Tokens refreshed successfully', {
        hasAccessToken: !!tokens.accessToken,
        hasRefreshToken: !!tokens.refreshToken,
      });
    },

    onRateLimited: (retryAfter) => {
      console.warn(`[API] Rate limited. Retry after ${retryAfter} seconds`);
    },

    onError: (error) => {
      console.error('[API Error]', error.message, error);
      // If it's a token refresh error, log more details
      if (error.message?.includes('RefreshToken')) {
        console.error('[Auth] Refresh token error detected. Tokens in storage:', {
          accessToken: localStorage.getItem('auth_gateway_access_token'),
          refreshToken: localStorage.getItem('auth_gateway_refresh_token'),
        });
      }
    },
  },

  // Retry configuration for resilience
  retry: {
    maxRetries: 3,
    initialDelayMs: 1000,
    maxDelayMs: 10000,
    backoffMultiplier: 2,
  },
});

// Export type for use in other files
export type ApiClient = typeof apiClient;
