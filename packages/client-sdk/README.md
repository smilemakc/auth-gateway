# @auth-gateway/client-sdk

Full-featured TypeScript client SDK for Auth Gateway API.

## Features

- **Complete API Coverage**: All 70+ endpoints covered
- **Type Safety**: Full TypeScript support with comprehensive type definitions
- **Easy Configuration**: Runtime configuration updates, header management
- **Token Management**: Automatic token refresh, multiple storage backends
- **Retry Policy**: Exponential backoff with configurable retry strategies
- **Interceptors**: Request, response, and error interceptors
- **WebSocket Support**: Real-time events with auto-reconnect
- **gRPC Client**: Server-to-server communication

## Installation

```bash
npm install @auth-gateway/client-sdk
# or
yarn add @auth-gateway/client-sdk
# or
pnpm add @auth-gateway/client-sdk
```

## Quick Start

```typescript
import { createClient, createLocalStorageTokenStorage } from '@auth-gateway/client-sdk';

// Create client
const client = createClient({
  baseUrl: 'https://api.example.com',
  tokenStorage: createLocalStorageTokenStorage(),
  callbacks: {
    onAuthFailure: () => window.location.href = '/login',
    onTokenRefresh: (tokens) => console.log('Tokens refreshed'),
  },
});

// Sign up
const { user } = await client.auth.signUp({
  email: 'user@example.com',
  username: 'johndoe',
  password: 'SecurePass123!',
  fullName: 'John Doe',
});

// Sign in
const { user, accessToken } = await client.auth.signIn({
  email: 'user@example.com',
  password: 'SecurePass123!',
});

// Get profile
const profile = await client.auth.getProfile();
```

## Configuration

### Basic Configuration

```typescript
const client = createClient({
  // Required
  baseUrl: 'https://api.example.com',

  // Optional
  timeout: 30000,              // Request timeout in ms
  debug: false,                // Enable debug logging
  autoRefreshTokens: true,     // Auto-refresh expired tokens

  // Custom headers
  headers: {
    'X-App-Version': '1.0.0',
  },

  // Retry policy
  retry: {
    maxRetries: 3,
    initialDelayMs: 1000,
    maxDelayMs: 30000,
    backoffMultiplier: 2,
  },

  // Lifecycle callbacks
  callbacks: {
    onRequest: (config) => console.log('Request:', config.url),
    onResponse: (response) => console.log('Response:', response.status),
    onError: (error) => console.error('Error:', error.message),
    onTokenRefresh: (tokens) => console.log('Tokens refreshed'),
    onAuthFailure: () => redirectToLogin(),
    onRateLimited: (retryAfter) => console.log(`Rate limited, retry after ${retryAfter}s`),
  },
});
```

### Runtime Configuration

```typescript
// Update configuration on the fly
client.configure({
  baseUrl: 'https://new-api.example.com',
  timeout: 60000,
});

// Set custom headers
client.setHeader('X-Custom-Header', 'value');
client.removeHeader('X-Custom-Header');
```

### Token Storage

```typescript
import {
  createLocalStorageTokenStorage,
  createSessionStorageTokenStorage,
  MemoryTokenStorage,
} from '@auth-gateway/client-sdk';

// LocalStorage (persists across sessions)
const client = createClient({
  baseUrl: 'https://api.example.com',
  tokenStorage: createLocalStorageTokenStorage(),
});

// SessionStorage (cleared on tab close)
const client = createClient({
  baseUrl: 'https://api.example.com',
  tokenStorage: createSessionStorageTokenStorage(),
});

// Memory (default, cleared on page reload)
const client = createClient({
  baseUrl: 'https://api.example.com',
  tokenStorage: new MemoryTokenStorage(),
});

// Custom storage
const client = createClient({
  baseUrl: 'https://api.example.com',
  tokenStorage: {
    getAccessToken: () => secureStore.get('access_token'),
    setAccessToken: (token) => secureStore.set('access_token', token),
    getRefreshToken: () => secureStore.get('refresh_token'),
    setRefreshToken: (token) => secureStore.set('refresh_token', token),
    clear: () => secureStore.clear(),
  },
});
```

## Authentication

### Email/Password Authentication

```typescript
// Sign up
const { user, accessToken, refreshToken } = await client.auth.signUp({
  email: 'user@example.com',
  username: 'johndoe',
  password: 'SecurePass123!',
  fullName: 'John Doe',
});

// Sign in
try {
  const result = await client.auth.signIn({
    email: 'user@example.com',
    password: 'SecurePass123!',
  });
  console.log('Logged in:', result.user.email);
} catch (error) {
  if (error instanceof TwoFactorRequiredError) {
    // Handle 2FA (see below)
  }
}

// Sign out
await client.signOut();

// Refresh token manually
const newTokens = await client.auth.refreshToken();
```

### Two-Factor Authentication

```typescript
import { TwoFactorRequiredError } from '@auth-gateway/client-sdk';

// Sign in with 2FA
try {
  await client.auth.signIn({ email, password });
} catch (error) {
  if (error instanceof TwoFactorRequiredError) {
    // Prompt user for 2FA code
    const code = await promptFor2FACode();

    const result = await client.twoFactor.verifyLogin({
      twoFactorToken: error.twoFactorToken,
      code,
    });
  }
}

// Enable 2FA
const setup = await client.twoFactor.setup({ password: 'current_password' });
console.log('Scan QR code:', setup.qrCodeUrl);
console.log('Backup codes:', setup.backupCodes);

// Verify and activate
await client.twoFactor.verify({ code: '123456' });

// Check status
const status = await client.twoFactor.getStatus();
console.log('2FA enabled:', status.enabled);

// Disable 2FA
await client.twoFactor.disable({
  password: 'current_password',
  code: '123456',
});
```

### OAuth / Social Login

```typescript
// Get available providers
const providers = await client.oauth.getEnabledProviders();
// [{ name: 'google', displayName: 'Google', enabled: true }, ...]

// Redirect to OAuth provider
window.location.href = client.oauth.getAuthorizationUrl('google');

// Or open in popup
const result = await client.oauth.openAuthPopup('google');
console.log('Logged in:', result.user.email, 'New user:', result.isNewUser);

// Handle callback (SPA routing)
const params = client.oauth.parseCallbackUrl(window.location.href);
if (params) {
  const result = await client.oauth.handleCallback(
    params.provider,
    params.code,
    params.state
  );
}
```

### Using Auth Gateway as OAuth Provider

If your application wants to use Auth Gateway as an OAuth/OIDC provider (instead of using it as a client), use the `OAuthProviderClient`:

```typescript
import { OAuthProviderClient } from '@auth-gateway/client-sdk';

// Create OAuth provider client
const oauth = new OAuthProviderClient({
  issuer: 'https://auth.example.com',
  clientId: 'your-client-id',
  clientSecret: 'your-client-secret', // Optional for public clients
  redirectUri: 'https://yourapp.com/callback',
  scopes: ['openid', 'profile', 'email'],
  usePKCE: true, // Recommended, required for public clients
});

// Authorization Code Flow with PKCE
// Step 1: Get authorization URL
const { url, state, codeVerifier } = await oauth.getAuthorizationUrl({
  prompt: 'consent',
  login_hint: 'user@example.com',
});

// Redirect user to authorization URL
window.location.href = url;

// Step 2: Handle callback (after user authorizes)
const params = new URLSearchParams(window.location.search);
const code = params.get('code');
const returnedState = params.get('state');

// Verify state matches
if (returnedState !== state) {
  throw new Error('Invalid state parameter');
}

// Exchange code for tokens
const tokens = await oauth.exchangeCode(code, codeVerifier);
console.log('Access token:', tokens.access_token);
console.log('ID token:', tokens.id_token);

// Get user info
const userInfo = await oauth.getUserInfo(tokens.access_token);
console.log('User:', userInfo);

// Decode ID token (basic validation)
const claims = oauth.decodeIdToken(tokens.id_token!);
console.log('ID token claims:', claims);

// Refresh tokens when access token expires
const newTokens = await oauth.refreshTokens(tokens.refresh_token!);

// Introspect token
const introspection = await oauth.introspectToken(tokens.access_token);
console.log('Token active:', introspection.active);

// Revoke token when done
await oauth.revokeToken(tokens.access_token);
```

#### Client Credentials Flow (Machine-to-Machine)

```typescript
// For server-to-server communication
const oauth = new OAuthProviderClient({
  issuer: 'https://auth.example.com',
  clientId: 'service-account',
  clientSecret: 'secret',
  redirectUri: '', // Not needed for client credentials
});

const tokens = await oauth.clientCredentialsGrant(['api:read', 'api:write']);
console.log('Service token:', tokens.access_token);
```

#### Device Authorization Flow

```typescript
// For devices with limited input (TVs, IoT, etc.)
const deviceAuth = await oauth.requestDeviceCode(['openid', 'profile']);

console.log('User code:', deviceAuth.user_code);
console.log('Go to:', deviceAuth.verification_uri);
console.log('Or visit:', deviceAuth.verification_uri_complete);

// Poll for authorization
const pollInterval = deviceAuth.interval * 1000;
let tokens;

while (!tokens) {
  try {
    tokens = await oauth.pollDeviceToken(deviceAuth.device_code);
  } catch (error) {
    if (error instanceof DeviceFlowPendingError) {
      if (error.code === 'slow_down') {
        pollInterval *= 2; // Back off
      }
      await new Promise(resolve => setTimeout(resolve, pollInterval));
    } else {
      throw error;
    }
  }
}

console.log('Device authorized:', tokens);
```

#### OIDC Discovery

```typescript
// Get OpenID Connect discovery document
const discovery = await oauth.getDiscovery();
console.log('Issuer:', discovery.issuer);
console.log('Authorization endpoint:', discovery.authorization_endpoint);
console.log('Supported scopes:', discovery.scopes_supported);

// Get JSON Web Key Set (for token verification)
const jwks = await oauth.getJWKS();
console.log('Keys:', jwks.keys);
```

### Passwordless Login

```typescript
// Request login code
await client.passwordless.request('user@example.com');

// Verify code
const result = await client.passwordless.verify('user@example.com', '123456');

// Or use the combined flow
const result = await client.passwordless.login(
  'user@example.com',
  async () => prompt('Enter the code sent to your email:')
);
```

## Profile Management

```typescript
// Get profile
const profile = await client.auth.getProfile();

// Update profile
const updated = await client.auth.updateProfile({
  fullName: 'New Name',
  profilePictureUrl: 'https://example.com/avatar.jpg',
});

// Change password
await client.auth.changePassword({
  oldPassword: 'current_password',
  newPassword: 'new_secure_password',
});

// Password reset flow
await client.auth.requestPasswordReset({ email: 'user@example.com' });
await client.auth.completePasswordReset({
  email: 'user@example.com',
  code: '123456',
  newPassword: 'new_password',
});
```

## Sessions Management

```typescript
// List sessions
const { sessions, total } = await client.sessions.list();

// Get current session
const current = await client.sessions.getCurrent();

// Revoke a specific session
await client.sessions.revoke(sessionId);

// Revoke all other sessions
await client.sessions.revokeAll();
```

## API Keys

```typescript
// Create API key
const { apiKey, plainKey } = await client.apiKeys.create({
  name: 'My API Key',
  scopes: ['users:read', 'profile:read'],
  expiresAt: '2025-12-31T23:59:59Z',
});
console.log('Save this key:', plainKey); // Only shown once!

// List API keys
const { apiKeys } = await client.apiKeys.list();

// Revoke API key
await client.apiKeys.revoke(apiKeyId);

// Delete API key
await client.apiKeys.delete(apiKeyId);
```

## Admin Operations

```typescript
// User management
const stats = await client.admin.users.getStats();
const { users, total } = await client.admin.users.list(1, 20);
const user = await client.admin.users.get(userId);
await client.admin.users.update(userId, { role: 'moderator' });
await client.admin.users.delete(userId);

// RBAC
const permissions = await client.admin.rbac.listPermissions();
const roles = await client.admin.rbac.listRoles();
const matrix = await client.admin.rbac.getPermissionMatrix();

// Sessions
const sessionStats = await client.admin.sessions.getStats();
await client.admin.sessions.revokeUserSessions(userId);

// IP Filters
await client.admin.ipFilters.whitelist('192.168.1.1', 'Office IP');
await client.admin.ipFilters.blacklist('10.0.0.1', 'Suspicious IP');

// Audit logs
const { logs } = await client.admin.audit.list({
  userId: 'user-id',
  action: 'login',
  status: 'failure',
});

// System
const health = await client.admin.system.getHealth();
await client.admin.system.enableMaintenanceMode('Scheduled maintenance');
```

## Interceptors

```typescript
// Request interceptor
const removeRequestInterceptor = client.addRequestInterceptor((config) => {
  console.log(`Making ${config.method} request to ${config.url}`);
  config.headers['X-Request-Time'] = Date.now().toString();
  return config;
});

// Response interceptor
const removeResponseInterceptor = client.addResponseInterceptor((response) => {
  console.log(`Response ${response.status} from ${response.requestId}`);
  return response;
});

// Error interceptor
const removeErrorInterceptor = client.addErrorInterceptor((error) => {
  console.error('API Error:', error.message, error.status);
  // Can transform or rethrow error
  return error;
});

// Remove interceptors when done
removeRequestInterceptor();
removeResponseInterceptor();
removeErrorInterceptor();
```

## WebSocket (Real-time Events)

```typescript
// Connect to WebSocket
const ws = await client.connectWebSocket({
  url: 'wss://api.example.com/ws', // Or set wsUrl in client config
});

// Listen to events
ws.on('session_revoked', (msg) => {
  console.log('Session revoked!');
  client.signOut();
});

ws.on('password_changed', (msg) => {
  console.log('Password was changed from another device');
});

ws.on('notification', (msg) => {
  showNotification(msg.payload);
});

// Listen to all events
ws.on('*', (msg) => {
  console.log('Event:', msg.type, msg.payload);
});

// Monitor connection state
ws.onStateChange((state) => {
  console.log('WebSocket state:', state); // 'connecting' | 'connected' | 'disconnected' | 'reconnecting'
});

// Disconnect
client.disconnectWebSocket();
```

## gRPC Client (Server-to-Server)

> **Important:** All gRPC methods require API key authentication. Pass `apiKey` in the config to authenticate.

```typescript
import { createGrpcClient } from '@auth-gateway/client-sdk/grpc';

const grpc = createGrpcClient({
  address: 'localhost:50051',
  useTls: false,
  apiKey: 'agw_YOUR_API_KEY', // Required: API key for authentication
  debug: true,
});

// Connect
await grpc.connect();

// Validate token (requires scope: token:validate)
const result = await grpc.validateToken('eyJhbGc...');
console.log('Valid:', result.valid, 'User:', result.userId);

// Check permission (requires scope: users:read)
const permission = await grpc.checkPermission(userId, 'products', 'write');
console.log('Allowed:', permission.allowed);

// Get user (requires scope: users:read)
const { user } = await grpc.getUser(userId);
console.log('User:', user?.email);

// Introspect token (requires scope: token:introspect)
const introspection = await grpc.introspectToken(token);
console.log('Active:', introspection.active, 'Blacklisted:', introspection.blacklisted);

// Update API key at runtime
grpc.setAPIKey('agw_NEW_API_KEY');

// Disconnect
grpc.disconnect();
```

## Error Handling

```typescript
import {
  AuthGatewayError,
  AuthenticationError,
  AuthorizationError,
  ValidationError,
  ConflictError,
  RateLimitError,
  NetworkError,
  TimeoutError,
  TwoFactorRequiredError,
} from '@auth-gateway/client-sdk';

try {
  await client.auth.signIn({ email, password });
} catch (error) {
  if (error instanceof TwoFactorRequiredError) {
    // 2FA required
    handleTwoFactor(error.twoFactorToken);
  } else if (error instanceof AuthenticationError) {
    // 401 - Invalid credentials
    showError('Invalid email or password');
  } else if (error instanceof AuthorizationError) {
    // 403 - Access denied
    showError('You do not have permission');
  } else if (error instanceof ValidationError) {
    // 400 - Validation failed
    showError(error.message);
  } else if (error instanceof ConflictError) {
    // 409 - Resource exists
    showError('Email already registered');
  } else if (error instanceof RateLimitError) {
    // 429 - Too many requests
    showError(`Rate limited. Retry after ${error.retryAfter} seconds`);
  } else if (error instanceof NetworkError) {
    // Network issues
    showError('Network error. Check your connection.');
  } else if (error instanceof TimeoutError) {
    // Request timeout
    showError('Request timed out. Please try again.');
  } else if (error instanceof AuthGatewayError) {
    // Other API errors
    showError(error.message);
  }
}
```

## Server-Side Usage (API Key)

```typescript
import { createApiKeyClient } from '@auth-gateway/client-sdk';

const client = createApiKeyClient(
  'https://api.example.com',
  'agw_xxxxxxxxxxxx'
);

// Validate user token
const profile = await client.auth.getProfile();
```

## TypeScript

All types are exported and can be imported:

```typescript
import type {
  User,
  AuthResponse,
  SignUpRequest,
  SignInRequest,
  APIKey,
  CreateAPIKeyRequest,
  Permission,
  Role,
  Session,
  ClientConfig,
  TokenStorage,
  // OAuth Provider types
  OAuthProviderClientConfig,
  OIDCDiscoveryDocument,
  TokenResponse,
  TokenIntrospectionResponse,
  UserInfoResponse,
  IDTokenClaims,
  DeviceAuthResponse,
  PKCEParams,
} from '@auth-gateway/client-sdk';
```

## License

MIT
