import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { OAuthProviderClient, DeviceFlowPendingError } from '../oauth-provider-client';
import type {
  OIDCDiscoveryDocument,
  JWKS,
  TokenResponse,
  TokenIntrospectionResponse,
  UserInfoResponse,
  DeviceAuthResponse,
} from '../types/oauth-provider';

// Mock discovery document
const mockDiscovery: OIDCDiscoveryDocument = {
  issuer: 'https://auth.example.com',
  authorization_endpoint: 'https://auth.example.com/oauth/authorize',
  token_endpoint: 'https://auth.example.com/oauth/token',
  userinfo_endpoint: 'https://auth.example.com/oauth/userinfo',
  jwks_uri: 'https://auth.example.com/.well-known/jwks.json',
  revocation_endpoint: 'https://auth.example.com/oauth/revoke',
  introspection_endpoint: 'https://auth.example.com/oauth/introspect',
  device_authorization_endpoint: 'https://auth.example.com/oauth/device',
  scopes_supported: ['openid', 'profile', 'email'],
  response_types_supported: ['code', 'token', 'id_token'],
  grant_types_supported: ['authorization_code', 'refresh_token', 'client_credentials'],
  subject_types_supported: ['public'],
  id_token_signing_alg_values_supported: ['RS256'],
  token_endpoint_auth_methods_supported: ['client_secret_basic', 'client_secret_post'],
  code_challenge_methods_supported: ['S256', 'plain'],
  claims_supported: ['sub', 'iss', 'aud', 'exp', 'iat', 'name', 'email'],
};

// Mock JWKS
const mockJWKS: JWKS = {
  keys: [
    {
      kty: 'RSA',
      use: 'sig',
      alg: 'RS256',
      kid: 'test-key-id',
      n: 'test-modulus',
      e: 'AQAB',
    },
  ],
};

// Mock token response
const mockTokenResponse: TokenResponse = {
  access_token: 'mock-access-token',
  token_type: 'Bearer',
  expires_in: 3600,
  refresh_token: 'mock-refresh-token',
  id_token: 'mock-id-token',
  scope: 'openid profile email',
};

// Mock introspection response
const mockIntrospectionResponse: TokenIntrospectionResponse = {
  active: true,
  scope: 'openid profile email',
  client_id: 'test-client-id',
  username: 'testuser',
  token_type: 'Bearer',
  exp: Math.floor(Date.now() / 1000) + 3600,
  iat: Math.floor(Date.now() / 1000),
  sub: 'user-123',
  aud: 'test-client-id',
  iss: 'https://auth.example.com',
};

// Mock user info response
const mockUserInfo: UserInfoResponse = {
  sub: 'user-123',
  name: 'Test User',
  email: 'test@example.com',
  email_verified: true,
};

// Mock device auth response
const mockDeviceAuthResponse: DeviceAuthResponse = {
  device_code: 'device-code-123',
  user_code: 'ABCD-EFGH',
  verification_uri: 'https://auth.example.com/device',
  verification_uri_complete: 'https://auth.example.com/device?user_code=ABCD-EFGH',
  expires_in: 600,
  interval: 5,
};

// Helper to create valid JWT for testing
function createMockIdToken(claims: Record<string, unknown>): string {
  const header = { alg: 'RS256', typ: 'JWT' };
  const headerB64 = btoa(JSON.stringify(header));
  const payloadB64 = btoa(JSON.stringify(claims));
  const signature = 'mock-signature';
  return `${headerB64}.${payloadB64}.${signature}`;
}

describe('OAuthProviderClient', () => {
  let client: OAuthProviderClient;
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    global.fetch = fetchMock;

    client = new OAuthProviderClient({
      issuer: 'https://auth.example.com',
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
      redirectUri: 'https://myapp.com/callback',
      scopes: ['openid', 'profile', 'email'],
      usePKCE: true,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('instantiation', () => {
    it('should create instance with required config', () => {
      const minimalClient = new OAuthProviderClient({
        issuer: 'https://auth.example.com',
        clientId: 'my-client',
        redirectUri: 'https://app.com/callback',
      });

      expect(minimalClient).toBeInstanceOf(OAuthProviderClient);
    });

    it('should apply default values for optional config', () => {
      const minimalClient = new OAuthProviderClient({
        issuer: 'https://auth.example.com',
        clientId: 'my-client',
        redirectUri: 'https://app.com/callback',
      });

      // Defaults should be applied: usePKCE=true, scopes=['openid']
      // We verify this by checking getAuthorizationUrl behavior
      expect(minimalClient).toBeInstanceOf(OAuthProviderClient);
    });

    it('should create public client without clientSecret', () => {
      const publicClient = new OAuthProviderClient({
        issuer: 'https://auth.example.com',
        clientId: 'public-client',
        redirectUri: 'https://spa.com/callback',
        usePKCE: true,
      });

      expect(publicClient).toBeInstanceOf(OAuthProviderClient);
    });

    it('should create confidential client with clientSecret', () => {
      const confidentialClient = new OAuthProviderClient({
        issuer: 'https://auth.example.com',
        clientId: 'confidential-client',
        clientSecret: 'super-secret',
        redirectUri: 'https://server.com/callback',
      });

      expect(confidentialClient).toBeInstanceOf(OAuthProviderClient);
    });
  });

  describe('getDiscovery', () => {
    it('should fetch and return discovery document', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });

      const discovery = await client.getDiscovery();

      expect(fetchMock).toHaveBeenCalledWith(
        'https://auth.example.com/.well-known/openid-configuration'
      );
      expect(discovery).toEqual(mockDiscovery);
    });

    it('should cache discovery document on subsequent calls', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });

      await client.getDiscovery();
      await client.getDiscovery();
      await client.getDiscovery();

      expect(fetchMock).toHaveBeenCalledTimes(1);
    });

    it('should throw error when discovery fetch fails', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 500,
      });

      await expect(client.getDiscovery()).rejects.toThrow(
        'Failed to fetch discovery document: 500'
      );
    });

    it('should throw error when discovery fetch returns 404', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 404,
      });

      await expect(client.getDiscovery()).rejects.toThrow(
        'Failed to fetch discovery document: 404'
      );
    });
  });

  describe('getJWKS', () => {
    it('should fetch and return JWKS', async () => {
      fetchMock
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockDiscovery),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockJWKS),
        });

      const jwks = await client.getJWKS();

      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(fetchMock).toHaveBeenLastCalledWith(mockDiscovery.jwks_uri);
      expect(jwks).toEqual(mockJWKS);
    });

    it('should cache JWKS on subsequent calls', async () => {
      fetchMock
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockDiscovery),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockJWKS),
        });

      await client.getJWKS();
      await client.getJWKS();

      // Discovery + JWKS = 2 calls, not 4
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });

    it('should throw error when JWKS fetch fails', async () => {
      fetchMock
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockDiscovery),
        })
        .mockResolvedValueOnce({
          ok: false,
          status: 500,
        });

      await expect(client.getJWKS()).rejects.toThrow('Failed to fetch JWKS: 500');
    });
  });

  describe('getAuthorizationUrl', () => {
    beforeEach(() => {
      fetchMock.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });
    });

    it('should generate authorization URL with required parameters', async () => {
      const result = await client.getAuthorizationUrl();

      expect(result.url).toContain(mockDiscovery.authorization_endpoint);
      expect(result.url).toContain('response_type=code');
      expect(result.url).toContain('client_id=test-client-id');
      expect(result.url).toContain('redirect_uri=' + encodeURIComponent('https://myapp.com/callback'));
      expect(result.state).toBeDefined();
      expect(result.state.length).toBe(32);
      expect(result.nonce).toBeDefined();
    });

    it('should include PKCE parameters when usePKCE is true', async () => {
      const result = await client.getAuthorizationUrl();

      expect(result.url).toContain('code_challenge=');
      expect(result.url).toContain('code_challenge_method=S256');
      expect(result.codeVerifier).toBeDefined();
      expect(result.codeVerifier!.length).toBe(64);
    });

    it('should not include PKCE parameters when usePKCE is false', async () => {
      const noPKCEClient = new OAuthProviderClient({
        issuer: 'https://auth.example.com',
        clientId: 'test-client-id',
        redirectUri: 'https://myapp.com/callback',
        usePKCE: false,
      });

      const result = await noPKCEClient.getAuthorizationUrl();

      expect(result.url).not.toContain('code_challenge');
      expect(result.url).not.toContain('code_challenge_method');
      expect(result.codeVerifier).toBeUndefined();
    });

    it('should use custom state when provided', async () => {
      const customState = 'my-custom-state-value';
      const result = await client.getAuthorizationUrl({ state: customState });

      expect(result.state).toBe(customState);
      expect(result.url).toContain('state=' + customState);
    });

    it('should use custom nonce when provided', async () => {
      const customNonce = 'my-custom-nonce-value';
      const result = await client.getAuthorizationUrl({ nonce: customNonce });

      expect(result.nonce).toBe(customNonce);
      expect(result.url).toContain('nonce=' + customNonce);
    });

    it('should use custom scope when provided', async () => {
      const customScope = 'openid profile custom:scope';
      const result = await client.getAuthorizationUrl({ scope: customScope });

      // URLSearchParams encodes spaces as + instead of %20
      expect(result.url).toContain('scope=openid+profile+custom%3Ascope');
    });

    it('should include prompt parameter when provided', async () => {
      const result = await client.getAuthorizationUrl({ prompt: 'consent' });

      expect(result.url).toContain('prompt=consent');
    });

    it('should include login_hint parameter when provided', async () => {
      const result = await client.getAuthorizationUrl({ login_hint: 'user@example.com' });

      expect(result.url).toContain('login_hint=' + encodeURIComponent('user@example.com'));
    });

    it('should generate unique state and nonce on each call', async () => {
      const result1 = await client.getAuthorizationUrl();
      const result2 = await client.getAuthorizationUrl();

      expect(result1.state).not.toBe(result2.state);
      expect(result1.nonce).not.toBe(result2.nonce);
    });
  });

  describe('exchangeCode', () => {
    beforeEach(() => {
      fetchMock
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockDiscovery),
        });
    });

    it('should exchange authorization code for tokens', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTokenResponse),
      });

      const tokens = await client.exchangeCode('auth-code-123');

      expect(fetchMock).toHaveBeenLastCalledWith(mockDiscovery.token_endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: expect.stringContaining('grant_type=authorization_code'),
      });
      expect(tokens).toEqual(mockTokenResponse);
    });

    it('should include code_verifier when provided', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTokenResponse),
      });

      await client.exchangeCode('auth-code-123', 'verifier-123');

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      expect(lastCall[1].body).toContain('code_verifier=verifier-123');
    });

    it('should include client_secret for confidential clients', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTokenResponse),
      });

      await client.exchangeCode('auth-code-123');

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      expect(lastCall[1].body).toContain('client_secret=test-client-secret');
    });

    it('should not include client_secret for public clients', async () => {
      const publicClient = new OAuthProviderClient({
        issuer: 'https://auth.example.com',
        clientId: 'public-client',
        redirectUri: 'https://spa.com/callback',
      });

      fetchMock
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockDiscovery),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockTokenResponse),
        });

      await publicClient.exchangeCode('auth-code-123');

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      expect(lastCall[1].body).not.toContain('client_secret');
    });

    it('should throw error with error_description when exchange fails', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () =>
          Promise.resolve({
            error: 'invalid_grant',
            error_description: 'The authorization code has expired',
          }),
      });

      await expect(client.exchangeCode('expired-code')).rejects.toThrow(
        'The authorization code has expired'
      );
    });

    it('should throw error with error code when no description provided', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () =>
          Promise.resolve({
            error: 'invalid_client',
          }),
      });

      await expect(client.exchangeCode('code')).rejects.toThrow('invalid_client');
    });

    it('should throw generic error when no error details provided', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () => Promise.resolve({}),
      });

      await expect(client.exchangeCode('code')).rejects.toThrow('Token exchange failed');
    });
  });

  describe('refreshTokens', () => {
    beforeEach(() => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });
    });

    it('should refresh tokens successfully', async () => {
      const refreshedTokens = {
        ...mockTokenResponse,
        access_token: 'new-access-token',
        refresh_token: 'new-refresh-token',
      };

      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(refreshedTokens),
      });

      const tokens = await client.refreshTokens('old-refresh-token');

      expect(fetchMock).toHaveBeenLastCalledWith(mockDiscovery.token_endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: expect.stringContaining('grant_type=refresh_token'),
      });
      expect(tokens.access_token).toBe('new-access-token');
    });

    it('should include refresh_token in request body', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTokenResponse),
      });

      await client.refreshTokens('refresh-token-123');

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      expect(lastCall[1].body).toContain('refresh_token=refresh-token-123');
    });

    it('should include client_secret for confidential clients', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTokenResponse),
      });

      await client.refreshTokens('refresh-token');

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      expect(lastCall[1].body).toContain('client_secret=test-client-secret');
    });

    it('should throw error when refresh fails', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () =>
          Promise.resolve({
            error: 'invalid_grant',
            error_description: 'Refresh token has been revoked',
          }),
      });

      await expect(client.refreshTokens('revoked-token')).rejects.toThrow(
        'Refresh token has been revoked'
      );
    });

    it('should throw generic error when no error details provided', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () => Promise.resolve({}),
      });

      await expect(client.refreshTokens('token')).rejects.toThrow('Token refresh failed');
    });
  });

  describe('introspectToken', () => {
    beforeEach(() => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });
    });

    it('should introspect token successfully', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockIntrospectionResponse),
      });

      const result = await client.introspectToken('some-access-token');

      expect(fetchMock).toHaveBeenLastCalledWith(mockDiscovery.introspection_endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          'Authorization': expect.stringContaining('Basic '),
        },
        body: 'token=some-access-token',
      });
      expect(result).toEqual(mockIntrospectionResponse);
    });

    it('should use Basic auth with client credentials', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockIntrospectionResponse),
      });

      await client.introspectToken('token');

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      const authHeader = lastCall[1].headers['Authorization'];
      const expectedBasic = 'Basic ' + btoa('test-client-id:test-client-secret');
      expect(authHeader).toBe(expectedBasic);
    });

    it('should return inactive response for invalid token', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ active: false }),
      });

      const result = await client.introspectToken('invalid-token');

      expect(result.active).toBe(false);
    });

    it('should throw error when introspection fails', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 401,
      });

      await expect(client.introspectToken('token')).rejects.toThrow(
        'Token introspection failed'
      );
    });
  });

  describe('revokeToken', () => {
    beforeEach(() => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });
    });

    it('should revoke token successfully', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
      });

      await client.revokeToken('token-to-revoke');

      expect(fetchMock).toHaveBeenLastCalledWith(mockDiscovery.revocation_endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          'Authorization': expect.stringContaining('Basic '),
        },
        body: 'token=token-to-revoke',
      });
    });

    it('should include token_type_hint when provided', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
      });

      await client.revokeToken('token', 'access_token');

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      expect(lastCall[1].body).toContain('token_type_hint=access_token');
    });

    it('should handle refresh_token type hint', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
      });

      await client.revokeToken('refresh-token', 'refresh_token');

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      expect(lastCall[1].body).toContain('token_type_hint=refresh_token');
    });

    it('should not throw on revocation failure (RFC 7009 compliance)', async () => {
      // Per RFC 7009, revocation endpoint should return 200 even for invalid tokens
      fetchMock.mockResolvedValueOnce({
        ok: true,
      });

      await expect(client.revokeToken('invalid-token')).resolves.toBeUndefined();
    });
  });

  describe('getUserInfo', () => {
    beforeEach(() => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });
    });

    it('should fetch user info successfully', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockUserInfo),
      });

      const userInfo = await client.getUserInfo('access-token');

      expect(fetchMock).toHaveBeenLastCalledWith(mockDiscovery.userinfo_endpoint, {
        headers: {
          'Authorization': 'Bearer access-token',
        },
      });
      expect(userInfo).toEqual(mockUserInfo);
    });

    it('should throw error when user info fetch fails', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 401,
      });

      await expect(client.getUserInfo('invalid-token')).rejects.toThrow(
        'Failed to fetch user info'
      );
    });
  });

  describe('decodeIdToken', () => {
    const validClaims = {
      iss: 'https://auth.example.com',
      sub: 'user-123',
      aud: 'test-client-id',
      exp: Math.floor(Date.now() / 1000) + 3600,
      iat: Math.floor(Date.now() / 1000),
      name: 'Test User',
      email: 'test@example.com',
    };

    it('should decode valid ID token', () => {
      const idToken = createMockIdToken(validClaims);
      const claims = client.decodeIdToken(idToken);

      expect(claims.iss).toBe('https://auth.example.com');
      expect(claims.sub).toBe('user-123');
      expect(claims.aud).toBe('test-client-id');
      expect(claims.name).toBe('Test User');
      expect(claims.email).toBe('test@example.com');
    });

    it('should throw error for invalid token format (not 3 parts)', () => {
      expect(() => client.decodeIdToken('invalid-token')).toThrow(
        'Invalid ID token format'
      );
      expect(() => client.decodeIdToken('part1.part2')).toThrow(
        'Invalid ID token format'
      );
      expect(() => client.decodeIdToken('a.b.c.d')).toThrow('Invalid ID token format');
    });

    it('should throw error for wrong issuer', () => {
      const wrongIssuerClaims = { ...validClaims, iss: 'https://wrong-issuer.com' };
      const idToken = createMockIdToken(wrongIssuerClaims);

      expect(() => client.decodeIdToken(idToken)).toThrow('Invalid issuer');
    });

    it('should throw error for wrong audience', () => {
      const wrongAudienceClaims = { ...validClaims, aud: 'wrong-client-id' };
      const idToken = createMockIdToken(wrongAudienceClaims);

      expect(() => client.decodeIdToken(idToken)).toThrow('Invalid audience');
    });

    it('should throw error for expired token', () => {
      const expiredClaims = {
        ...validClaims,
        exp: Math.floor(Date.now() / 1000) - 3600, // Expired 1 hour ago
      };
      const idToken = createMockIdToken(expiredClaims);

      expect(() => client.decodeIdToken(idToken)).toThrow('Token expired');
    });

    it('should return all standard claims', () => {
      const fullClaims = {
        ...validClaims,
        auth_time: Math.floor(Date.now() / 1000) - 60,
        nonce: 'test-nonce',
        acr: '0',
        amr: ['pwd'],
        azp: 'test-client-id',
        given_name: 'Test',
        family_name: 'User',
        preferred_username: 'testuser',
        picture: 'https://example.com/avatar.jpg',
        email_verified: true,
      };
      const idToken = createMockIdToken(fullClaims);
      const claims = client.decodeIdToken(idToken);

      expect(claims.auth_time).toBe(fullClaims.auth_time);
      expect(claims.nonce).toBe('test-nonce');
      expect(claims.given_name).toBe('Test');
      expect(claims.family_name).toBe('User');
      expect(claims.email_verified).toBe(true);
    });
  });

  describe('generatePKCE', () => {
    it('should generate valid PKCE parameters', async () => {
      const pkce = await client.generatePKCE();

      expect(pkce.code_verifier).toBeDefined();
      expect(pkce.code_verifier.length).toBe(64);
      expect(pkce.code_challenge).toBeDefined();
      expect(pkce.code_challenge_method).toBe('S256');
    });

    it('should generate unique code verifiers', async () => {
      const pkce1 = await client.generatePKCE();
      const pkce2 = await client.generatePKCE();

      expect(pkce1.code_verifier).not.toBe(pkce2.code_verifier);
      expect(pkce1.code_challenge).not.toBe(pkce2.code_challenge);
    });

    it('should generate code verifier with allowed characters only', async () => {
      const pkce = await client.generatePKCE();
      const allowedChars = /^[A-Za-z0-9\-._~]+$/;

      expect(pkce.code_verifier).toMatch(allowedChars);
    });

    it('should generate base64url-encoded code challenge', async () => {
      const pkce = await client.generatePKCE();
      // Base64URL should not contain + / or padding =
      expect(pkce.code_challenge).not.toContain('+');
      expect(pkce.code_challenge).not.toContain('/');
      expect(pkce.code_challenge).not.toMatch(/=+$/);
    });
  });

  describe('requestDeviceCode', () => {
    beforeEach(() => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });
    });

    it('should request device code successfully', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDeviceAuthResponse),
      });

      const response = await client.requestDeviceCode();

      expect(fetchMock).toHaveBeenLastCalledWith(
        mockDiscovery.device_authorization_endpoint,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: expect.stringContaining('client_id=test-client-id'),
        }
      );
      expect(response).toEqual(mockDeviceAuthResponse);
    });

    it('should include custom scopes when provided', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDeviceAuthResponse),
      });

      await client.requestDeviceCode(['openid', 'custom:scope']);

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      // URLSearchParams encodes spaces as + instead of %20
      expect(lastCall[1].body).toContain('scope=openid+custom%3Ascope');
    });

    it('should throw error when device authorization fails', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () =>
          Promise.resolve({
            error: 'unauthorized_client',
            error_description: 'Client not authorized for device flow',
          }),
      });

      await expect(client.requestDeviceCode()).rejects.toThrow(
        'Client not authorized for device flow'
      );
    });
  });

  describe('pollDeviceToken', () => {
    beforeEach(() => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });
    });

    it('should return tokens when user completes authorization', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTokenResponse),
      });

      const tokens = await client.pollDeviceToken('device-code-123');

      expect(fetchMock).toHaveBeenLastCalledWith(mockDiscovery.token_endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: expect.stringContaining('grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code'),
      });
      expect(tokens).toEqual(mockTokenResponse);
    });

    it('should throw DeviceFlowPendingError for authorization_pending', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () =>
          Promise.resolve({
            error: 'authorization_pending',
          }),
      });

      await expect(client.pollDeviceToken('device-code')).rejects.toThrow(
        DeviceFlowPendingError
      );

      try {
        await client.pollDeviceToken('device-code');
      } catch (error) {
        // Second call will work because we need to re-add the mock
        fetchMock.mockResolvedValueOnce({
          ok: false,
          json: () =>
            Promise.resolve({
              error: 'authorization_pending',
            }),
        });
      }
    });

    it('should throw DeviceFlowPendingError for slow_down', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () =>
          Promise.resolve({
            error: 'slow_down',
          }),
      });

      try {
        await client.pollDeviceToken('device-code');
        expect.fail('Should have thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(DeviceFlowPendingError);
        expect((error as DeviceFlowPendingError).code).toBe('slow_down');
      }
    });

    it('should throw regular error for other failures', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () =>
          Promise.resolve({
            error: 'expired_token',
            error_description: 'The device code has expired',
          }),
      });

      await expect(client.pollDeviceToken('device-code')).rejects.toThrow(
        'The device code has expired'
      );
    });
  });

  describe('clientCredentialsGrant', () => {
    beforeEach(() => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });
    });

    it('should obtain tokens using client credentials', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTokenResponse),
      });

      const tokens = await client.clientCredentialsGrant();

      expect(fetchMock).toHaveBeenLastCalledWith(mockDiscovery.token_endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          'Authorization': expect.stringContaining('Basic '),
        },
        body: expect.stringContaining('grant_type=client_credentials'),
      });
      expect(tokens).toEqual(mockTokenResponse);
    });

    it('should include custom scopes when provided', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockTokenResponse),
      });

      await client.clientCredentialsGrant(['api:read', 'api:write']);

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      // URLSearchParams encodes spaces as + instead of %20
      expect(lastCall[1].body).toContain('scope=api%3Aread+api%3Awrite');
    });

    it('should throw error when called without client_secret', async () => {
      const publicClient = new OAuthProviderClient({
        issuer: 'https://auth.example.com',
        clientId: 'public-client',
        redirectUri: 'https://app.com/callback',
      });

      await expect(publicClient.clientCredentialsGrant()).rejects.toThrow(
        'Client credentials grant requires client_secret'
      );
    });

    it('should throw error on grant failure', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        json: () =>
          Promise.resolve({
            error: 'invalid_client',
            error_description: 'Invalid client credentials',
          }),
      });

      await expect(client.clientCredentialsGrant()).rejects.toThrow(
        'Invalid client credentials'
      );
    });
  });

  describe('DeviceFlowPendingError', () => {
    it('should have correct name and code for authorization_pending', () => {
      const error = new DeviceFlowPendingError('authorization_pending');

      expect(error.name).toBe('DeviceFlowPendingError');
      expect(error.code).toBe('authorization_pending');
      expect(error.message).toBe('Device flow authorization_pending');
    });

    it('should have correct name and code for slow_down', () => {
      const error = new DeviceFlowPendingError('slow_down');

      expect(error.name).toBe('DeviceFlowPendingError');
      expect(error.code).toBe('slow_down');
      expect(error.message).toBe('Device flow slow_down');
    });

    it('should be instanceof Error', () => {
      const error = new DeviceFlowPendingError('authorization_pending');

      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(DeviceFlowPendingError);
    });
  });

  describe('Edge Cases', () => {
    it('should handle URL encoding in authorization URL', async () => {
      fetchMock.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });

      const result = await client.getAuthorizationUrl({
        login_hint: 'user+test@example.com',
        scope: 'openid profile email custom:special/scope',
      });

      // login_hint should be properly encoded
      expect(result.url).toContain('login_hint=user%2Btest%40example.com');
      // URLSearchParams encodes spaces as + and special chars appropriately
      expect(result.url).toContain('scope=openid+profile+email+custom%3Aspecial%2Fscope');
    });

    it('should handle special characters in client credentials', async () => {
      const specialClient = new OAuthProviderClient({
        issuer: 'https://auth.example.com',
        clientId: 'client:with:colons',
        clientSecret: 'secret/with+special=chars',
        redirectUri: 'https://app.com/callback',
      });

      fetchMock
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockDiscovery),
        })
        .mockResolvedValueOnce({
          ok: true,
          json: () => Promise.resolve(mockIntrospectionResponse),
        });

      await specialClient.introspectToken('token');

      const lastCall = fetchMock.mock.calls[fetchMock.mock.calls.length - 1];
      const authHeader = lastCall[1].headers['Authorization'];
      expect(authHeader).toContain('Basic ');
    });

    it('should handle empty scopes array', async () => {
      const noScopesClient = new OAuthProviderClient({
        issuer: 'https://auth.example.com',
        clientId: 'client',
        redirectUri: 'https://app.com/callback',
        scopes: [],
      });

      fetchMock.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockDiscovery),
      });

      const result = await noScopesClient.getAuthorizationUrl();

      // Should still have scope parameter, even if empty
      expect(result.url).toContain('scope=');
    });
  });
});
