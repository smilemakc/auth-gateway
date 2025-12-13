/**
 * Main AuthGatewayClient class
 * Unified client that provides access to all Auth Gateway services
 */

import type { ClientConfig, RequestInterceptor, ResponseInterceptor, ErrorInterceptor, TokenStorage, ClientCallbacks, RetryConfig } from './config/types';
import { HttpClient, MemoryTokenStorage } from './core/http';
import { WebSocketClient, type WebSocketConfig } from './core/websocket';
import {
  AuthService,
  OAuthService,
  TwoFactorService,
  OTPService,
  SMSService,
  PasswordlessService,
  APIKeysService,
  SessionsService,
  HealthService,
  AdminUsersService,
  AdminRBACService,
  AdminSessionsService,
  AdminIPFiltersService,
  AdminAuditService,
  AdminBrandingService,
  AdminSystemService,
  AdminAPIKeysService,
  AdminSMSSettingsService,
} from './services';

/** Admin services container */
export interface AdminServices {
  /** User management */
  users: AdminUsersService;
  /** Roles and permissions */
  rbac: AdminRBACService;
  /** Session management */
  sessions: AdminSessionsService;
  /** IP whitelist/blacklist */
  ipFilters: AdminIPFiltersService;
  /** Audit logs */
  audit: AdminAuditService;
  /** Branding and customization */
  branding: AdminBrandingService;
  /** System management */
  system: AdminSystemService;
  /** API keys management */
  apiKeys: AdminAPIKeysService;
  /** SMS settings */
  smsSettings: AdminSMSSettingsService;
}

/**
 * Main Auth Gateway Client
 *
 * Provides unified access to all Auth Gateway services with:
 * - Easy configuration and runtime updates
 * - Automatic token management
 * - Retry policy with exponential backoff
 * - Request/response interceptors
 * - WebSocket support for real-time events
 *
 * @example
 * ```typescript
 * // Create client
 * const client = new AuthGatewayClient({
 *   baseUrl: 'https://api.example.com',
 *   callbacks: {
 *     onTokenRefresh: (tokens) => console.log('Tokens refreshed'),
 *     onAuthFailure: () => router.push('/login'),
 *   },
 * });
 *
 * // Sign in
 * const { user, accessToken } = await client.auth.signIn({
 *   email: 'user@example.com',
 *   password: 'password123',
 * });
 *
 * // Get profile
 * const profile = await client.auth.getProfile();
 *
 * // Enable 2FA
 * const setup = await client.twoFactor.setup({ password: 'password123' });
 * ```
 */
export class AuthGatewayClient {
  private readonly http: HttpClient;
  private wsClient: WebSocketClient | null = null;
  private readonly config: ClientConfig;

  // ==================== SERVICES ====================

  /** Authentication service (sign up, sign in, tokens, profile) */
  public readonly auth: AuthService;

  /** OAuth service (social login) */
  public readonly oauth: OAuthService;

  /** Two-factor authentication service */
  public readonly twoFactor: TwoFactorService;

  /** OTP service (email/phone verification) */
  public readonly otp: OTPService;

  /** SMS service */
  public readonly sms: SMSService;

  /** Passwordless authentication service */
  public readonly passwordless: PasswordlessService;

  /** API keys service */
  public readonly apiKeys: APIKeysService;

  /** Sessions service */
  public readonly sessions: SessionsService;

  /** Health check service */
  public readonly health: HealthService;

  /** Admin services (requires admin role) */
  public readonly admin: AdminServices;

  /**
   * Create a new Auth Gateway Client
   * @param config Client configuration
   */
  constructor(config: ClientConfig) {
    this.config = config;
    this.http = new HttpClient(config);

    // Initialize services
    this.auth = new AuthService(this.http);
    this.oauth = new OAuthService(this.http, config.baseUrl);
    this.twoFactor = new TwoFactorService(this.http);
    this.otp = new OTPService(this.http);
    this.sms = new SMSService(this.http);
    this.passwordless = new PasswordlessService(this.http);
    this.apiKeys = new APIKeysService(this.http);
    this.sessions = new SessionsService(this.http);
    this.health = new HealthService(this.http);

    // Initialize admin services
    this.admin = {
      users: new AdminUsersService(this.http),
      rbac: new AdminRBACService(this.http),
      sessions: new AdminSessionsService(this.http),
      ipFilters: new AdminIPFiltersService(this.http),
      audit: new AdminAuditService(this.http),
      branding: new AdminBrandingService(this.http),
      system: new AdminSystemService(this.http),
      apiKeys: new AdminAPIKeysService(this.http),
      smsSettings: new AdminSMSSettingsService(this.http),
    };
  }

  // ==================== CONFIGURATION ====================

  /**
   * Update client configuration at runtime
   * @param config Partial configuration to update
   *
   * @example
   * ```typescript
   * // Change base URL
   * client.configure({ baseUrl: 'https://new-api.example.com' });
   *
   * // Update headers
   * client.configure({ headers: { 'X-Custom-Header': 'value' } });
   *
   * // Change retry policy
   * client.configure({ retry: { maxRetries: 5 } });
   * ```
   */
  configure(config: Partial<ClientConfig>): void {
    this.http.configure(config);
  }

  /**
   * Set a custom header for all requests
   * @param key Header name
   * @param value Header value
   */
  setHeader(key: string, value: string): void {
    this.http.setHeader(key, value);
  }

  /**
   * Remove a header
   * @param key Header name
   */
  removeHeader(key: string): void {
    this.http.removeHeader(key);
  }

  /**
   * Get current token storage
   * @returns Token storage instance
   */
  getTokenStorage(): TokenStorage {
    return this.http.getTokenStorage();
  }

  // ==================== INTERCEPTORS ====================

  /**
   * Add a request interceptor
   * @param interceptor Request interceptor function
   * @returns Function to remove the interceptor
   *
   * @example
   * ```typescript
   * const removeInterceptor = client.addRequestInterceptor((config) => {
   *   console.log(`Making request to ${config.url}`);
   *   return config;
   * });
   *
   * // Later: remove the interceptor
   * removeInterceptor();
   * ```
   */
  addRequestInterceptor(interceptor: RequestInterceptor): () => void {
    return this.http.addRequestInterceptor(interceptor);
  }

  /**
   * Add a response interceptor
   * @param interceptor Response interceptor function
   * @returns Function to remove the interceptor
   */
  addResponseInterceptor(interceptor: ResponseInterceptor): () => void {
    return this.http.addResponseInterceptor(interceptor);
  }

  /**
   * Add an error interceptor
   * @param interceptor Error interceptor function
   * @returns Function to remove the interceptor
   */
  addErrorInterceptor(interceptor: ErrorInterceptor): () => void {
    return this.http.addErrorInterceptor(interceptor);
  }

  // ==================== WEBSOCKET ====================

  /**
   * Connect to WebSocket for real-time events
   * @param config Optional WebSocket configuration override
   * @returns WebSocket client instance
   *
   * @example
   * ```typescript
   * const ws = await client.connectWebSocket();
   *
   * // Listen to events
   * ws.on('session_revoked', (msg) => {
   *   console.log('Session was revoked!');
   *   router.push('/login');
   * });
   *
   * ws.on('notification', (msg) => {
   *   showNotification(msg.payload);
   * });
   * ```
   */
  async connectWebSocket(
    config?: Partial<Omit<WebSocketConfig, 'tokenStorage'>>
  ): Promise<WebSocketClient> {
    if (this.wsClient?.isConnected()) {
      return this.wsClient;
    }

    const wsUrl = config?.url ?? this.config.wsUrl;
    if (!wsUrl) {
      throw new Error('WebSocket URL not configured. Provide wsUrl in client config or connectWebSocket options.');
    }

    this.wsClient = new WebSocketClient({
      url: wsUrl,
      tokenStorage: this.http.getTokenStorage(),
      debug: this.config.debug,
      ...config,
    });

    await this.wsClient.connect();
    return this.wsClient;
  }

  /**
   * Disconnect WebSocket
   */
  disconnectWebSocket(): void {
    this.wsClient?.disconnect();
    this.wsClient = null;
  }

  /**
   * Get current WebSocket client (if connected)
   * @returns WebSocket client or null
   */
  getWebSocket(): WebSocketClient | null {
    return this.wsClient;
  }

  // ==================== UTILITIES ====================

  /**
   * Check if user is authenticated
   * @returns True if access token exists
   */
  async isAuthenticated(): Promise<boolean> {
    return this.auth.isAuthenticated();
  }

  /**
   * Sign out and cleanup
   * Logs out, clears tokens, and disconnects WebSocket
   */
  async signOut(): Promise<void> {
    try {
      await this.auth.logout();
    } catch {
      // Ignore logout errors
    }
    await this.auth.clearTokens();
    this.disconnectWebSocket();
  }

  /**
   * Get HTTP client for advanced usage
   * @returns HTTP client instance
   */
  getHttpClient(): HttpClient {
    return this.http;
  }
}

// ==================== FACTORY FUNCTIONS ====================

/**
 * Create a new Auth Gateway Client
 * @param config Client configuration
 * @returns AuthGatewayClient instance
 *
 * @example
 * ```typescript
 * const client = createClient({
 *   baseUrl: 'https://api.example.com',
 * });
 * ```
 */
export function createClient(config: ClientConfig): AuthGatewayClient {
  return new AuthGatewayClient(config);
}

/**
 * Create a client with API key authentication (server-to-server)
 * @param baseUrl API base URL
 * @param apiKey API key
 * @param options Additional options
 * @returns AuthGatewayClient instance
 *
 * @example
 * ```typescript
 * const client = createApiKeyClient(
 *   'https://api.example.com',
 *   'agw_xxxxxxxxxxxx'
 * );
 *
 * // Make authenticated requests
 * const user = await client.auth.getProfile();
 * ```
 */
export function createApiKeyClient(
  baseUrl: string,
  apiKey: string,
  options?: Partial<Omit<ClientConfig, 'baseUrl' | 'apiKey'>>
): AuthGatewayClient {
  return new AuthGatewayClient({
    baseUrl,
    apiKey,
    autoRefreshTokens: false, // No token refresh with API keys
    ...options,
  });
}

/**
 * Create token storage that persists to localStorage (browser only)
 * @param prefix Optional key prefix (default: 'auth_gateway_')
 * @returns TokenStorage implementation
 */
export function createLocalStorageTokenStorage(
  prefix = 'auth_gateway_'
): TokenStorage {
  return {
    getAccessToken(): string | null {
      if (typeof localStorage === 'undefined') return null;
      return localStorage.getItem(`${prefix}access_token`);
    },
    setAccessToken(token: string): void {
      if (typeof localStorage === 'undefined') return;
      localStorage.setItem(`${prefix}access_token`, token);
    },
    getRefreshToken(): string | null {
      if (typeof localStorage === 'undefined') return null;
      return localStorage.getItem(`${prefix}refresh_token`);
    },
    setRefreshToken(token: string): void {
      if (typeof localStorage === 'undefined') return;
      localStorage.setItem(`${prefix}refresh_token`, token);
    },
    clear(): void {
      if (typeof localStorage === 'undefined') return;
      localStorage.removeItem(`${prefix}access_token`);
      localStorage.removeItem(`${prefix}refresh_token`);
    },
  };
}

/**
 * Create token storage that persists to sessionStorage (browser only)
 * @param prefix Optional key prefix (default: 'auth_gateway_')
 * @returns TokenStorage implementation
 */
export function createSessionStorageTokenStorage(
  prefix = 'auth_gateway_'
): TokenStorage {
  return {
    getAccessToken(): string | null {
      if (typeof sessionStorage === 'undefined') return null;
      return sessionStorage.getItem(`${prefix}access_token`);
    },
    setAccessToken(token: string): void {
      if (typeof sessionStorage === 'undefined') return;
      sessionStorage.setItem(`${prefix}access_token`, token);
    },
    getRefreshToken(): string | null {
      if (typeof sessionStorage === 'undefined') return null;
      return sessionStorage.getItem(`${prefix}refresh_token`);
    },
    setRefreshToken(token: string): void {
      if (typeof sessionStorage === 'undefined') return;
      sessionStorage.setItem(`${prefix}refresh_token`, token);
    },
    clear(): void {
      if (typeof sessionStorage === 'undefined') return;
      sessionStorage.removeItem(`${prefix}access_token`);
      sessionStorage.removeItem(`${prefix}refresh_token`);
    },
  };
}

// Re-export token storage class
export { MemoryTokenStorage };
