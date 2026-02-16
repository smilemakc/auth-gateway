import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { HttpClient, MemoryTokenStorage } from '../core/http';
import { AuthService } from '../services/auth';
import { TwoFactorRequiredError, AuthenticationError } from '../core/errors';
import type { AuthResponse } from '../types/auth';
import type { User } from '../types/user';
import type { ApiResponse, TokenStorage } from '../config/types';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const TEST_BASE_URL = 'https://api.example.com';

function createMockUser(overrides: Partial<User> = {}): User {
  return {
    id: 'user-123',
    email: 'test@example.com',
    username: 'testuser',
    full_name: 'Test User',
    roles: [{ id: 'role-1', name: 'user', display_name: 'User' }],
    account_type: 'human',
    email_verified: true,
    phone_verified: false,
    is_active: true,
    totp_enabled: false,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

function createMockAuthResponse(overrides: Partial<AuthResponse> = {}): AuthResponse {
  return {
    access_token: 'mock-access-token',
    refresh_token: 'mock-refresh-token',
    user: createMockUser(),
    expires_in: 900,
    ...overrides,
  };
}

function wrapApiResponse<T>(data: T, status = 200): ApiResponse<T> {
  return {
    data,
    status,
    headers: { 'content-type': 'application/json' },
    requestId: 'req-123',
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('AuthService', () => {
  let http: HttpClient;
  let authService: AuthService;
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore – override global fetch for tests
    global.fetch = fetchMock;

    http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      autoRefreshTokens: false,
      retry: { maxRetries: 0 },
    });
    authService = new AuthService(http);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  // -----------------------------------------------------------------------
  // signUp
  // -----------------------------------------------------------------------
  describe('signUp', () => {
    it('should register a new user and store tokens', async () => {
      const mockResponse = createMockAuthResponse();

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(mockResponse),
      });

      const result = await authService.signUp({
        email: 'new@example.com',
        username: 'newuser',
        password: 'Password123!',
        full_name: 'New User',
      });

      expect(result.access_token).toBe('mock-access-token');
      expect(result.refresh_token).toBe('mock-refresh-token');
      expect(result.user.email).toBe('test@example.com');

      // Verify tokens are stored
      const tokenStorage = http.getTokenStorage();
      expect(await tokenStorage.getAccessToken()).toBe('mock-access-token');
      expect(await tokenStorage.getRefreshToken()).toBe('mock-refresh-token');

      // Verify request
      expect(fetchMock).toHaveBeenCalledTimes(1);
      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/signup`);
      expect(options.method).toBe('POST');
      expect(JSON.parse(options.body)).toEqual({
        email: 'new@example.com',
        username: 'newuser',
        password: 'Password123!',
        full_name: 'New User',
      });
    });

    it('should send request without Authorization header (skipAuth)', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(createMockAuthResponse()),
      });

      // Pre-set a token to verify it is NOT sent
      await http.getTokenStorage().setAccessToken('existing-token');

      await authService.signUp({
        email: 'new@example.com',
        username: 'newuser',
        password: 'Password123!',
      });

      const [, options] = fetchMock.mock.calls[0]!;
      expect(options.headers['Authorization']).toBeUndefined();
    });

    it('should throw on server error', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 400,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ error: 'Validation failed', message: 'Email already exists' }),
      });

      await expect(
        authService.signUp({
          email: 'existing@example.com',
          username: 'existuser',
          password: 'Password123!',
        })
      ).rejects.toThrow('Email already exists');
    });
  });

  // -----------------------------------------------------------------------
  // signIn
  // -----------------------------------------------------------------------
  describe('signIn', () => {
    it('should authenticate user and store tokens', async () => {
      const mockResponse = createMockAuthResponse();

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(mockResponse),
      });

      const result = await authService.signIn({
        email: 'test@example.com',
        password: 'Password123!',
      });

      expect(result.access_token).toBe('mock-access-token');
      expect(result.user.email).toBe('test@example.com');

      // Tokens should be stored
      const tokenStorage = http.getTokenStorage();
      expect(await tokenStorage.getAccessToken()).toBe('mock-access-token');
      expect(await tokenStorage.getRefreshToken()).toBe('mock-refresh-token');

      // Verify correct endpoint and body
      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/signin`);
      expect(JSON.parse(options.body)).toEqual({
        email: 'test@example.com',
        password: 'Password123!',
      });
    });

    it('should throw TwoFactorRequiredError when 2FA is required', async () => {
      const mockResponse = createMockAuthResponse({
        requires_2fa: true,
        two_factor_token: '2fa-token-abc',
      });

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(mockResponse),
      });

      try {
        await authService.signIn({
          email: 'test@example.com',
          password: 'Password123!',
        });
        expect.fail('Should have thrown TwoFactorRequiredError');
      } catch (error) {
        expect(error).toBeInstanceOf(TwoFactorRequiredError);
        expect((error as TwoFactorRequiredError).twoFactorToken).toBe('2fa-token-abc');
      }
    });

    it('should throw on invalid credentials (401)', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 401,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ message: 'Invalid credentials' }),
      });

      await expect(
        authService.signIn({
          email: 'test@example.com',
          password: 'wrong',
        })
      ).rejects.toThrow('Invalid credentials');
    });

    it('should support phone-based sign in', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(createMockAuthResponse()),
      });

      await authService.signIn({
        phone: '+1234567890',
        password: 'Password123!',
      });

      const [, options] = fetchMock.mock.calls[0]!;
      expect(JSON.parse(options.body)).toEqual({
        phone: '+1234567890',
        password: 'Password123!',
      });
    });
  });

  // -----------------------------------------------------------------------
  // refreshToken
  // -----------------------------------------------------------------------
  describe('refreshToken', () => {
    it('should refresh tokens using stored refresh token', async () => {
      // Pre-store a refresh token
      await http.getTokenStorage().setRefreshToken('old-refresh-token');

      const mockResponse = createMockAuthResponse({
        access_token: 'new-access-token',
        refresh_token: 'new-refresh-token',
      });

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(mockResponse),
      });

      const result = await authService.refreshToken();

      expect(result.access_token).toBe('new-access-token');
      expect(result.refresh_token).toBe('new-refresh-token');

      // Verify stored tokens are updated
      const tokenStorage = http.getTokenStorage();
      expect(await tokenStorage.getAccessToken()).toBe('new-access-token');
      expect(await tokenStorage.getRefreshToken()).toBe('new-refresh-token');

      // Verify request body
      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/refresh`);
      expect(JSON.parse(options.body)).toEqual({
        refresh_token: 'old-refresh-token',
      });
    });

    it('should use explicitly provided refresh token', async () => {
      const mockResponse = createMockAuthResponse();

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(mockResponse),
      });

      await authService.refreshToken('explicit-refresh-token');

      const [, options] = fetchMock.mock.calls[0]!;
      expect(JSON.parse(options.body)).toEqual({
        refresh_token: 'explicit-refresh-token',
      });
    });

    it('should throw when no refresh token is available', async () => {
      await expect(authService.refreshToken()).rejects.toThrow(
        'No refresh token available'
      );
    });

    it('should throw on server error during refresh', async () => {
      await http.getTokenStorage().setRefreshToken('expired-token');

      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 401,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ message: 'Token expired' }),
      });

      await expect(authService.refreshToken()).rejects.toThrow('Token expired');
    });
  });

  // -----------------------------------------------------------------------
  // logout
  // -----------------------------------------------------------------------
  describe('logout', () => {
    it('should call logout endpoint and clear tokens', async () => {
      // Pre-store tokens
      const tokenStorage = http.getTokenStorage();
      await tokenStorage.setAccessToken('access-token');
      await tokenStorage.setRefreshToken('refresh-token');

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ message: 'Logged out successfully' }),
      });

      const result = await authService.logout();

      expect(result.message).toBe('Logged out successfully');
      expect(await tokenStorage.getAccessToken()).toBeNull();
      expect(await tokenStorage.getRefreshToken()).toBeNull();

      // Verify endpoint
      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/logout`);
      expect(options.method).toBe('POST');
    });
  });

  // -----------------------------------------------------------------------
  // getProfile / updateProfile
  // -----------------------------------------------------------------------
  describe('getProfile', () => {
    it('should return the current user profile', async () => {
      const mockUser = createMockUser();

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(mockUser),
      });

      const result = await authService.getProfile();

      expect(result.id).toBe('user-123');
      expect(result.email).toBe('test@example.com');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/profile`);
      expect(options.method).toBe('GET');
    });
  });

  describe('updateProfile', () => {
    it('should update the user profile', async () => {
      const updatedUser = createMockUser({ full_name: 'Updated Name' });

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(updatedUser),
      });

      const result = await authService.updateProfile({ full_name: 'Updated Name' });

      expect(result.full_name).toBe('Updated Name');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/profile`);
      expect(options.method).toBe('PUT');
      expect(JSON.parse(options.body)).toEqual({ full_name: 'Updated Name' });
    });
  });

  // -----------------------------------------------------------------------
  // changePassword
  // -----------------------------------------------------------------------
  describe('changePassword', () => {
    it('should change the user password', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ message: 'Password changed successfully' }),
      });

      const result = await authService.changePassword({
        old_password: 'OldPass123!',
        new_password: 'NewPass456!',
      });

      expect(result.message).toBe('Password changed successfully');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/change-password`);
      expect(JSON.parse(options.body)).toEqual({
        old_password: 'OldPass123!',
        new_password: 'NewPass456!',
      });
    });
  });

  // -----------------------------------------------------------------------
  // verifyEmail / resendVerification
  // -----------------------------------------------------------------------
  describe('verifyEmail', () => {
    it('should verify email with code', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ valid: true, message: 'Email verified' }),
      });

      const result = await authService.verifyEmail({
        email: 'test@example.com',
        code: '123456',
      });

      expect(result.valid).toBe(true);

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/verify/email`);
    });
  });

  describe('resendVerification', () => {
    it('should resend verification email', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ message: 'Verification email sent', email: 'test@example.com' }),
      });

      const result = await authService.resendVerification({ email: 'test@example.com' });

      expect(result.message).toBe('Verification email sent');
      expect(result.email).toBe('test@example.com');
    });
  });

  // -----------------------------------------------------------------------
  // Password reset
  // -----------------------------------------------------------------------
  describe('requestPasswordReset', () => {
    it('should request a password reset', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ message: 'Reset email sent', email: 'test@example.com' }),
      });

      const result = await authService.requestPasswordReset({ email: 'test@example.com' });

      expect(result.message).toBe('Reset email sent');

      const [url] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/password/reset/request`);
    });
  });

  describe('completePasswordReset', () => {
    it('should complete the password reset', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ message: 'Password reset complete' }),
      });

      const result = await authService.completePasswordReset({
        email: 'test@example.com',
        code: '123456',
        new_password: 'NewSecure123!',
      });

      expect(result.message).toBe('Password reset complete');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/auth/password/reset/complete`);
      expect(JSON.parse(options.body)).toEqual({
        email: 'test@example.com',
        code: '123456',
        new_password: 'NewSecure123!',
      });
    });
  });

  // -----------------------------------------------------------------------
  // Token management helpers
  // -----------------------------------------------------------------------
  describe('setTokens', () => {
    it('should store access and refresh tokens', async () => {
      await authService.setTokens('new-access', 'new-refresh');

      const tokenStorage = http.getTokenStorage();
      expect(await tokenStorage.getAccessToken()).toBe('new-access');
      expect(await tokenStorage.getRefreshToken()).toBe('new-refresh');
    });
  });

  describe('getAccessToken', () => {
    it('should return stored access token', async () => {
      await http.getTokenStorage().setAccessToken('my-token');

      const token = await authService.getAccessToken();
      expect(token).toBe('my-token');
    });

    it('should return null when no token is stored', async () => {
      const token = await authService.getAccessToken();
      expect(token).toBeNull();
    });
  });

  describe('isAuthenticated', () => {
    it('should return true when access token exists', async () => {
      await http.getTokenStorage().setAccessToken('valid-token');

      const authenticated = await authService.isAuthenticated();
      expect(authenticated).toBe(true);
    });

    it('should return false when no access token exists', async () => {
      const authenticated = await authService.isAuthenticated();
      expect(authenticated).toBe(false);
    });
  });

  describe('clearTokens', () => {
    it('should clear all stored tokens', async () => {
      const tokenStorage = http.getTokenStorage();
      await tokenStorage.setAccessToken('token-a');
      await tokenStorage.setRefreshToken('token-r');

      await authService.clearTokens();

      expect(await tokenStorage.getAccessToken()).toBeNull();
      expect(await tokenStorage.getRefreshToken()).toBeNull();
    });
  });

  // -----------------------------------------------------------------------
  // validateToken
  // -----------------------------------------------------------------------
  describe('validateToken', () => {
    it('should validate a token and return result', async () => {
      const mockValidation = {
        valid: true,
        user_id: 'user-123',
        email: 'test@example.com',
        username: 'testuser',
        roles: ['user'],
        expires_at: Math.floor(Date.now() / 1000) + 900,
        is_active: true,
      };

      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve(mockValidation),
      });

      const result = await authService.validateToken('some-jwt-token');

      expect(result.valid).toBe(true);
      expect(result.user_id).toBe('user-123');

      const [url, options] = fetchMock.mock.calls[0]!;
      expect(url).toBe(`${TEST_BASE_URL}/api/v1/token/validate`);
      expect(JSON.parse(options.body)).toEqual({ access_token: 'some-jwt-token' });
    });

    it('should return invalid for expired tokens', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'Content-Type': 'application/json' }),
        json: () => Promise.resolve({ valid: false, error_message: 'Token expired' }),
      });

      const result = await authService.validateToken('expired-token');

      expect(result.valid).toBe(false);
      expect(result.error_message).toBe('Token expired');
    });
  });

  // -----------------------------------------------------------------------
  // Auto-refresh on 401 (integration with HttpClient)
  // -----------------------------------------------------------------------
  describe('auto-refresh on 401', () => {
    it('should automatically refresh token and retry on 401', async () => {
      // Enable auto-refresh
      const autoRefreshHttp = new HttpClient({
        baseUrl: TEST_BASE_URL,
        autoRefreshTokens: true,
        retry: { maxRetries: 0 },
      });
      const autoRefreshAuth = new AuthService(autoRefreshHttp);

      // Store initial tokens
      await autoRefreshHttp.getTokenStorage().setAccessToken('expired-access');
      await autoRefreshHttp.getTokenStorage().setRefreshToken('valid-refresh');

      // First call returns 401
      // Then refresh endpoint returns new tokens
      // Then retry of original call succeeds
      fetchMock
        // 1st attempt: getProfile returns 401
        .mockResolvedValueOnce({
          ok: false,
          status: 401,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: () => Promise.resolve({ message: 'Token expired' }),
        })
        // 2nd: refresh token call
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: () => Promise.resolve(
            createMockAuthResponse({
              access_token: 'fresh-access',
              refresh_token: 'fresh-refresh',
            })
          ),
        })
        // 3rd: retried getProfile succeeds
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: () => Promise.resolve(createMockUser()),
        });

      const profile = await autoRefreshAuth.getProfile();

      expect(profile.email).toBe('test@example.com');
      expect(fetchMock).toHaveBeenCalledTimes(3);

      // Verify new tokens are stored
      expect(await autoRefreshHttp.getTokenStorage().getAccessToken()).toBe('fresh-access');
      expect(await autoRefreshHttp.getTokenStorage().getRefreshToken()).toBe('fresh-refresh');
    });

    it('should call onAuthFailure callback when refresh also fails', async () => {
      const onAuthFailure = vi.fn();

      const autoRefreshHttp = new HttpClient({
        baseUrl: TEST_BASE_URL,
        autoRefreshTokens: true,
        retry: { maxRetries: 0 },
        callbacks: { onAuthFailure },
      });
      const autoRefreshAuth = new AuthService(autoRefreshHttp);

      await autoRefreshHttp.getTokenStorage().setAccessToken('expired-access');
      await autoRefreshHttp.getTokenStorage().setRefreshToken('also-expired');

      fetchMock
        // 1st: getProfile returns 401
        .mockResolvedValueOnce({
          ok: false,
          status: 401,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: () => Promise.resolve({ message: 'Token expired' }),
        })
        // 2nd: refresh also fails with 401
        .mockResolvedValueOnce({
          ok: false,
          status: 401,
          headers: new Headers({ 'Content-Type': 'application/json' }),
          json: () => Promise.resolve({ message: 'Refresh token expired' }),
        });

      await expect(autoRefreshAuth.getProfile()).rejects.toThrow('Token expired');
      expect(onAuthFailure).toHaveBeenCalledTimes(1);
    });
  });
});

// ---------------------------------------------------------------------------
// MemoryTokenStorage
// ---------------------------------------------------------------------------
describe('MemoryTokenStorage', () => {
  it('should store and retrieve access token', () => {
    const storage = new MemoryTokenStorage();

    expect(storage.getAccessToken()).toBeNull();
    storage.setAccessToken('access-123');
    expect(storage.getAccessToken()).toBe('access-123');
  });

  it('should store and retrieve refresh token', () => {
    const storage = new MemoryTokenStorage();

    expect(storage.getRefreshToken()).toBeNull();
    storage.setRefreshToken('refresh-456');
    expect(storage.getRefreshToken()).toBe('refresh-456');
  });

  it('should clear all tokens', () => {
    const storage = new MemoryTokenStorage();

    storage.setAccessToken('a');
    storage.setRefreshToken('r');
    storage.clear();

    expect(storage.getAccessToken()).toBeNull();
    expect(storage.getRefreshToken()).toBeNull();
  });
});
