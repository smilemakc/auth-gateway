/**
 * @auth-gateway/client-sdk
 *
 * Full-featured TypeScript client SDK for Auth Gateway API
 *
 * @example
 * ```typescript
 * import { createClient, createLocalStorageTokenStorage } from '@auth-gateway/client-sdk';
 *
 * // Create client with localStorage persistence
 * const client = createClient({
 *   baseUrl: 'https://api.example.com',
 *   tokenStorage: createLocalStorageTokenStorage(),
 *   callbacks: {
 *     onAuthFailure: () => window.location.href = '/login',
 *   },
 * });
 *
 * // Sign in
 * try {
 *   const { user } = await client.auth.signIn({
 *     email: 'user@example.com',
 *     password: 'password123',
 *   });
 *   console.log('Logged in as:', user.email);
 * } catch (error) {
 *   if (error instanceof TwoFactorRequiredError) {
 *     // Handle 2FA
 *     const result = await client.twoFactor.verifyLogin({
 *       twoFactorToken: error.twoFactorToken,
 *       code: await getUserInput('Enter 2FA code'),
 *     });
 *   }
 * }
 *
 * // Use OAuth
 * const providers = await client.oauth.getEnabledProviders();
 * const googleUrl = client.oauth.getAuthorizationUrl('google');
 *
 * // Admin operations
 * const stats = await client.admin.users.getStats();
 * ```
 */

// Main client
export {
  AuthGatewayClient,
  createClient,
  createApiKeyClient,
  createLocalStorageTokenStorage,
  createSessionStorageTokenStorage,
  MemoryTokenStorage,
  type AdminServices,
} from './client';

// Configuration
export * from './config';

// Core utilities
export {
  // Errors
  AuthGatewayError,
  NetworkError,
  TimeoutError,
  AuthenticationError,
  AuthorizationError,
  NotFoundError,
  ValidationError,
  ConflictError,
  RateLimitError,
  ServerError,
  TwoFactorRequiredError,
  createErrorFromResponse,
  isRetryableError,
  // Retry
  withRetry,
  createRetryWrapper,
  calculateRetryDelay,
  sleep,
  type RetryContext,
  // HTTP
  HttpClient,
  // WebSocket
  WebSocketClient,
  createWebSocketClient,
  type WebSocketEventType,
  type WebSocketMessage,
  type WebSocketState,
  type WebSocketConfig,
} from './core';

// Services
export {
  BaseService,
  AuthService,
  OAuthService,
  TwoFactorService,
  OTPService,
  SMSService,
  PasswordlessService,
  APIKeysService,
  SessionsService,
  HealthService,
  // Admin services
  AdminUsersService,
  AdminRBACService,
  AdminSessionsService,
  AdminIPFiltersService,
  AdminAuditService,
  AdminBrandingService,
  AdminSystemService,
  AdminAPIKeysService,
  AdminSMSSettingsService,
  AdminOAuthProvidersService,
  AdminOAuthClientsService,
  AdminTemplatesService,
  AdminWebhooksService,
  AdminGroupsService,
  AdminLDAPService,
  AdminSAMLService,
  AdminBulkService,
  AdminSCIMService,
  AdminAppOAuthProvidersService,
  AdminTelegramBotsService,
  AdminUserTelegramService,
  type AuditLogQueryOptions,
  type ListClientsParams,
} from './services';

// OAuth Provider Client (for apps using Auth Gateway as OAuth provider)
export {
  OAuthProviderClient,
  DeviceFlowPendingError,
  type OAuthProviderClientConfig,
} from './oauth-provider-client';

// Types
export * from './types';
