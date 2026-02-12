/**
 * OAuth Provider Client
 *
 * Utility for third-party applications integrating with Auth Gateway as an OAuth/OIDC provider.
 * Handles authorization code flow, PKCE, token exchange, and OIDC discovery.
 *
 * @example
 * ```typescript
 * import { OAuthProviderClient } from '@auth-gateway/client-sdk/oauth-provider';
 *
 * const client = new OAuthProviderClient({
 *   issuer: 'https://auth.example.com',
 *   clientId: 'your-client-id',
 *   clientSecret: 'your-secret', // Optional for public clients
 *   redirectUri: 'https://yourapp.com/callback',
 *   scopes: ['openid', 'profile', 'email'],
 *   usePKCE: true, // Recommended
 * });
 *
 * // Get authorization URL
 * const { url, state, codeVerifier } = await client.getAuthorizationUrl();
 * window.location.href = url;
 *
 * // Handle callback (after redirect)
 * const tokens = await client.exchangeCode(code, codeVerifier);
 *
 * // Get user info
 * const userInfo = await client.getUserInfo(tokens.access_token);
 * ```
 */

import type {
  OIDCDiscoveryDocument,
  JWKS,
  TokenResponse,
  TokenIntrospectionResponse,
  UserInfoResponse,
  DeviceAuthResponse,
  IDTokenClaims,
  PKCEParams,
  AuthorizationUrlOptions,
  AuthorizationUrlResult,
} from './types/oauth-provider';

export interface OAuthProviderClientConfig {
  issuer: string;
  clientId: string;
  clientSecret?: string;
  redirectUri: string;
  scopes?: string[];
  usePKCE?: boolean;
}

export class OAuthProviderClient {
  private config: OAuthProviderClientConfig;
  private discovery?: OIDCDiscoveryDocument;
  private jwks?: JWKS;

  constructor(config: OAuthProviderClientConfig) {
    this.config = {
      usePKCE: true,
      scopes: ['openid'],
      ...config,
    };
  }

  async getDiscovery(): Promise<OIDCDiscoveryDocument> {
    if (this.discovery) return this.discovery;

    const response = await fetch(
      `${this.config.issuer}/.well-known/openid-configuration`
    );

    if (!response.ok) {
      throw new Error(`Failed to fetch discovery document: ${response.status}`);
    }

    this.discovery = await response.json();
    return this.discovery!;
  }

  async getJWKS(): Promise<JWKS> {
    if (this.jwks) return this.jwks;

    const discovery = await this.getDiscovery();
    const response = await fetch(discovery.jwks_uri);

    if (!response.ok) {
      throw new Error(`Failed to fetch JWKS: ${response.status}`);
    }

    this.jwks = await response.json();
    return this.jwks!;
  }

  async getAuthorizationUrl(options?: AuthorizationUrlOptions): Promise<AuthorizationUrlResult> {
    const discovery = await this.getDiscovery();

    const state = options?.state || this.generateRandomString(32);
    const nonce = options?.nonce || this.generateRandomString(32);

    const params = new URLSearchParams({
      response_type: 'code',
      client_id: this.config.clientId,
      redirect_uri: this.config.redirectUri,
      scope: options?.scope || this.config.scopes!.join(' '),
      state,
      nonce,
    });

    if (options?.prompt) {
      params.set('prompt', options.prompt);
    }
    if (options?.login_hint) {
      params.set('login_hint', options.login_hint);
    }

    let codeVerifier: string | undefined;

    if (this.config.usePKCE) {
      const pkce = await this.generatePKCE();
      codeVerifier = pkce.code_verifier;
      params.set('code_challenge', pkce.code_challenge);
      params.set('code_challenge_method', pkce.code_challenge_method);
    }

    return {
      url: `${discovery.authorization_endpoint}?${params.toString()}`,
      state,
      nonce,
      codeVerifier,
    };
  }

  async exchangeCode(code: string, codeVerifier?: string): Promise<TokenResponse> {
    const discovery = await this.getDiscovery();

    const params = new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      redirect_uri: this.config.redirectUri,
      client_id: this.config.clientId,
    });

    if (this.config.clientSecret) {
      params.set('client_secret', this.config.clientSecret);
    }

    if (codeVerifier) {
      params.set('code_verifier', codeVerifier);
    }

    const response = await fetch(discovery.token_endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params.toString(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error_description || error.error || 'Token exchange failed');
    }

    return response.json();
  }

  async refreshTokens(refreshToken: string): Promise<TokenResponse> {
    const discovery = await this.getDiscovery();

    const params = new URLSearchParams({
      grant_type: 'refresh_token',
      refresh_token: refreshToken,
      client_id: this.config.clientId,
    });

    if (this.config.clientSecret) {
      params.set('client_secret', this.config.clientSecret);
    }

    const response = await fetch(discovery.token_endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params.toString(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error_description || error.error || 'Token refresh failed');
    }

    return response.json();
  }

  async introspectToken(token: string): Promise<TokenIntrospectionResponse> {
    const discovery = await this.getDiscovery();

    const params = new URLSearchParams({ token });

    const headers: Record<string, string> = {
      'Content-Type': 'application/x-www-form-urlencoded',
    };

    if (this.config.clientSecret) {
      headers['Authorization'] = `Basic ${this.base64Encode(`${this.config.clientId}:${this.config.clientSecret}`)}`;
    }

    const response = await fetch(discovery.introspection_endpoint, {
      method: 'POST',
      headers,
      body: params.toString(),
    });

    if (!response.ok) {
      throw new Error('Token introspection failed');
    }

    return response.json();
  }

  async revokeToken(token: string, tokenTypeHint?: 'access_token' | 'refresh_token'): Promise<void> {
    const discovery = await this.getDiscovery();

    const params = new URLSearchParams({ token });
    if (tokenTypeHint) {
      params.set('token_type_hint', tokenTypeHint);
    }

    const headers: Record<string, string> = {
      'Content-Type': 'application/x-www-form-urlencoded',
    };

    if (this.config.clientSecret) {
      headers['Authorization'] = `Basic ${this.base64Encode(`${this.config.clientId}:${this.config.clientSecret}`)}`;
    }

    await fetch(discovery.revocation_endpoint, {
      method: 'POST',
      headers,
      body: params.toString(),
    });
  }

  async getUserInfo(accessToken: string): Promise<UserInfoResponse> {
    const discovery = await this.getDiscovery();

    const response = await fetch(discovery.userinfo_endpoint, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to fetch user info');
    }

    return response.json();
  }

  decodeIdToken(idToken: string): IDTokenClaims {
    const parts = idToken.split('.');
    if (parts.length !== 3) {
      throw new Error('Invalid ID token format');
    }

    const payload = JSON.parse(this.base64Decode(parts[1] as string));

    if (payload.iss !== this.config.issuer) {
      throw new Error('Invalid issuer');
    }
    if (payload.aud !== this.config.clientId) {
      throw new Error('Invalid audience');
    }
    if (payload.exp < Date.now() / 1000) {
      throw new Error('Token expired');
    }

    return payload as IDTokenClaims;
  }

  async requestDeviceCode(scopes?: string[]): Promise<DeviceAuthResponse> {
    const discovery = await this.getDiscovery();

    const params = new URLSearchParams({
      client_id: this.config.clientId,
      scope: scopes?.join(' ') || this.config.scopes!.join(' '),
    });

    const response = await fetch(discovery.device_authorization_endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params.toString(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error_description || error.error || 'Device authorization failed');
    }

    return response.json();
  }

  async pollDeviceToken(deviceCode: string): Promise<TokenResponse> {
    const discovery = await this.getDiscovery();

    const params = new URLSearchParams({
      grant_type: 'urn:ietf:params:oauth:grant-type:device_code',
      device_code: deviceCode,
      client_id: this.config.clientId,
    });

    const response = await fetch(discovery.token_endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params.toString(),
    });

    if (!response.ok) {
      const error = await response.json();
      if (error.error === 'authorization_pending' || error.error === 'slow_down') {
        throw new DeviceFlowPendingError(error.error);
      }
      throw new Error(error.error_description || error.error || 'Device token request failed');
    }

    return response.json();
  }

  async clientCredentialsGrant(scopes?: string[]): Promise<TokenResponse> {
    if (!this.config.clientSecret) {
      throw new Error('Client credentials grant requires client_secret');
    }

    const discovery = await this.getDiscovery();

    const params = new URLSearchParams({
      grant_type: 'client_credentials',
      scope: scopes?.join(' ') || '',
    });

    const response = await fetch(discovery.token_endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Authorization': `Basic ${this.base64Encode(`${this.config.clientId}:${this.config.clientSecret}`)}`,
      },
      body: params.toString(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error_description || error.error || 'Client credentials grant failed');
    }

    return response.json();
  }

  async generatePKCE(): Promise<PKCEParams> {
    const codeVerifier = this.generateRandomString(64);
    const encoder = new TextEncoder();
    const data = encoder.encode(codeVerifier);
    const hash = await crypto.subtle.digest('SHA-256', data);
    const codeChallenge = this.base64UrlEncode(new Uint8Array(hash));

    return {
      code_verifier: codeVerifier,
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
    };
  }

  private generateRandomString(length: number): string {
    const charset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';
    const randomValues = crypto.getRandomValues(new Uint8Array(length));
    return Array.from(randomValues)
      .map(x => charset[x % charset.length])
      .join('');
  }

  private base64UrlEncode(buffer: Uint8Array): string {
    return this.base64Encode(String.fromCharCode(...buffer))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '');
  }

  private base64Encode(str: string): string {
    if (typeof btoa !== 'undefined') {
      return btoa(str);
    }
    return Buffer.from(str).toString('base64');
  }

  private base64Decode(str: string): string {
    if (typeof atob !== 'undefined') {
      return atob(str);
    }
    return Buffer.from(str, 'base64').toString('utf-8');
  }
}

export class DeviceFlowPendingError extends Error {
  constructor(public code: 'authorization_pending' | 'slow_down') {
    super(`Device flow ${code}`);
    this.name = 'DeviceFlowPendingError';
  }
}
