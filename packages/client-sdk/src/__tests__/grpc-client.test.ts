import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { GrpcClientConfig } from '../grpc/types';

/**
 * These tests verify the gRPC client's public API and behavior
 * without establishing a real gRPC connection. We mock the underlying
 * gRPC dependencies so the tests can run in any environment.
 */

// ---------------------------------------------------------------------------
// Use vi.hoisted() to define mock variables BEFORE vi.mock hoisting.
// This is required because vi.mock factories execute at module scope.
// ---------------------------------------------------------------------------

const {
  mockWaitForReady,
  mockClose,
  mockGrpcMethods,
  MockServiceClient,
  methodResponses,
} = vi.hoisted(() => {
  const mockWaitForReady = vi.fn();
  const mockClose = vi.fn();
  const mockGrpcMethods: Record<string, ReturnType<typeof vi.fn>> = {};

  // Shared response registry: setupGrpcMethodResponse writes here,
  // and gRPC method mocks read from it. This survives client recreation.
  const methodResponses: Record<string, { response: unknown; error: { message: string; code: number } | null }> = {};

  function createMockGrpcMethod(name: string) {
    const fn = vi.fn().mockImplementation(
      (
        _request: unknown,
        _metadata: unknown,
        _options: unknown,
        callback: (error: unknown, response: unknown) => void
      ) => {
        const cfg = methodResponses[name];
        if (cfg?.error) {
          callback({ message: cfg.error.message, code: cfg.error.code }, null);
        } else {
          callback(null, cfg?.response ?? null);
        }
      }
    );
    mockGrpcMethods[name] = fn;
    return fn;
  }

  const MockServiceClient = vi.fn().mockImplementation(() => {
    const instance: Record<string, unknown> = {
      waitForReady: mockWaitForReady,
      close: mockClose,
    };

    const methods = [
      'validateToken', 'getUser', 'checkPermission', 'introspectToken',
      'createUser', 'login', 'initPasswordlessRegistration', 'completePasswordlessRegistration',
      'sendOTP', 'verifyOTP', 'loginWithOTP', 'verifyLoginOTP',
      'registerWithOTP', 'verifyRegistrationOTP', 'introspectOAuthToken',
      'validateOAuthClient', 'getOAuthClient', 'sendEmail',
      'getUserApplicationProfile', 'getUserTelegramBots', 'syncUsers',
      'getApplicationAuthConfig', 'createTokenExchange', 'redeemTokenExchange',
    ];

    for (const method of methods) {
      instance[method] = createMockGrpcMethod(method);
    }

    return instance;
  });

  return { mockWaitForReady, mockClose, mockGrpcMethods, MockServiceClient, methodResponses };
});

// ---------------------------------------------------------------------------
// Mock @grpc/grpc-js and @grpc/proto-loader
// ---------------------------------------------------------------------------

vi.mock('@grpc/grpc-js', () => {
  return {
    credentials: {
      createInsecure: vi.fn().mockReturnValue('insecure-creds'),
      createSsl: vi.fn().mockReturnValue('ssl-creds'),
    },
    Metadata: vi.fn().mockImplementation(() => {
      const store: Record<string, string> = {};
      return {
        add: vi.fn((key: string, value: string) => {
          store[key] = value;
        }),
        get: vi.fn((key: string) => store[key] ? [store[key]] : []),
        _store: store,
      };
    }),
    loadPackageDefinition: vi.fn().mockReturnValue({
      auth: {
        AuthService: MockServiceClient,
      },
    }),
  };
});

vi.mock('@grpc/proto-loader', () => ({
  loadSync: vi.fn().mockReturnValue({}),
}));

// Import AFTER mocks are set up
import { AuthGrpcClient, createGrpcClient } from '../grpc/client';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function createClient(overrides: Partial<GrpcClientConfig> = {}): AuthGrpcClient {
  return new AuthGrpcClient({
    address: 'localhost:50051',
    ...overrides,
  });
}

function setupWaitForReady(succeed = true, error?: Error) {
  mockWaitForReady.mockImplementation((_deadline: number, callback: (error?: Error) => void) => {
    if (succeed) {
      callback();
    } else {
      callback(error ?? new Error('Connection refused'));
    }
  });
}

function setupGrpcMethodResponse(
  method: string,
  response: unknown,
  error: { message: string; code: number } | null = null
) {
  // Write to the shared registry; the mock reads from it at call time
  methodResponses[method] = { response, error };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('AuthGrpcClient', () => {
  beforeEach(async () => {
    // Reset call history without clearing mock implementations
    MockServiceClient.mockClear();
    mockWaitForReady.mockClear();
    mockClose.mockClear();
    for (const fn of Object.values(mockGrpcMethods)) {
      fn.mockClear();
    }
    // Clear the shared response registry
    for (const key of Object.keys(methodResponses)) {
      delete methodResponses[key];
    }

    // Re-setup loadPackageDefinition return value
    const grpc = await import('@grpc/grpc-js');
    (grpc.loadPackageDefinition as ReturnType<typeof vi.fn>).mockReturnValue({
      auth: {
        AuthService: MockServiceClient,
      },
    });

    setupWaitForReady(true);
  });

  // -----------------------------------------------------------------------
  // Connection lifecycle
  // -----------------------------------------------------------------------
  describe('connection lifecycle', () => {
    it('should connect successfully', async () => {
      const client = createClient();

      await client.connect();

      expect(client.isConnected()).toBe(true);
    });

    it('should not reconnect if already connected', async () => {
      const client = createClient();

      await client.connect();
      await client.connect(); // Second call should be a no-op

      expect(MockServiceClient).toHaveBeenCalledTimes(1);
    });

    it('should reject with error when connection fails', async () => {
      setupWaitForReady(false, new Error('ECONNREFUSED'));

      const client = createClient();

      await expect(client.connect()).rejects.toThrow(
        'Failed to connect to gRPC server: ECONNREFUSED'
      );
    });

    it('should disconnect and clear methods', async () => {
      const client = createClient();

      await client.connect();
      expect(client.isConnected()).toBe(true);

      client.disconnect();
      expect(client.isConnected()).toBe(false);
      expect(mockClose).toHaveBeenCalledTimes(1);
    });

    it('should handle disconnect when not connected', () => {
      const client = createClient();

      client.disconnect();
      expect(client.isConnected()).toBe(false);
      expect(mockClose).not.toHaveBeenCalled();
    });
  });

  // -----------------------------------------------------------------------
  // setAPIKey
  // -----------------------------------------------------------------------
  describe('setAPIKey', () => {
    it('should update the API key used in subsequent calls', async () => {
      const client = createClient({ apiKey: 'initial-key' });

      client.setAPIKey('new-key');

      await client.connect();
      setupGrpcMethodResponse('validateToken', { valid: true, userId: 'u1' });

      await client.validateToken('token-abc');

      // Metadata constructor should have been called to build call metadata
      const { Metadata } = await import('@grpc/grpc-js');
      expect(Metadata).toHaveBeenCalled();
    });
  });

  // -----------------------------------------------------------------------
  // validateToken
  // -----------------------------------------------------------------------
  describe('validateToken', () => {
    it('should validate token and return response', async () => {
      const client = createClient({ apiKey: 'agw_testkey' });
      await client.connect();

      const mockResponse = {
        valid: true,
        userId: 'user-123',
        email: 'test@example.com',
        roles: ['user'],
      };
      setupGrpcMethodResponse('validateToken', mockResponse);

      const result = await client.validateToken('jwt-token-abc');

      expect(result.valid).toBe(true);
      expect(result.userId).toBe('user-123');

      // Verify request structure
      const callArgs = mockGrpcMethods['validateToken']!.mock.calls[0]!;
      expect(callArgs[0]).toEqual({ accessToken: 'jwt-token-abc', applicationId: '' });
    });

    it('should auto-connect if not connected', async () => {
      const client = createClient();
      setupGrpcMethodResponse('validateToken', { valid: false });

      const result = await client.validateToken('some-token-12345678');

      expect(result.valid).toBe(false);
      expect(client.isConnected()).toBe(true);
    });

    it('should propagate gRPC error', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('validateToken', null, {
        message: 'UNAUTHENTICATED: Invalid API key',
        code: 16,
      });

      await expect(client.validateToken('bad-token-1234567890')).rejects.toThrow(
        'gRPC error: UNAUTHENTICATED: Invalid API key (code: 16)'
      );
    });
  });

  // -----------------------------------------------------------------------
  // getUser
  // -----------------------------------------------------------------------
  describe('getUser', () => {
    it('should get user by ID', async () => {
      const client = createClient();
      await client.connect();

      const mockResponse = {
        user: {
          id: 'user-42',
          email: 'user@example.com',
          username: 'testuser',
          fullName: 'Test User',
        },
      };
      setupGrpcMethodResponse('getUser', mockResponse);

      const result = await client.getUser('user-42');

      expect(result.user.email).toBe('user@example.com');

      const callArgs = mockGrpcMethods['getUser']!.mock.calls[0]!;
      expect(callArgs[0]).toEqual({ userId: 'user-42', applicationId: '' });
    });

    it('should throw on user not found', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('getUser', null, {
        message: 'NOT_FOUND: User not found',
        code: 5,
      });

      await expect(client.getUser('nonexistent')).rejects.toThrow(
        'gRPC error: NOT_FOUND: User not found (code: 5)'
      );
    });
  });

  // -----------------------------------------------------------------------
  // checkPermission
  // -----------------------------------------------------------------------
  describe('checkPermission', () => {
    it('should check permission and return allowed=true', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('checkPermission', { allowed: true });

      const result = await client.checkPermission('user-1', 'articles', 'write');

      expect(result.allowed).toBe(true);

      const callArgs = mockGrpcMethods['checkPermission']!.mock.calls[0]!;
      expect(callArgs[0]).toEqual({
        userId: 'user-1',
        resource: 'articles',
        action: 'write',
        applicationId: '',
      });
    });

    it('should return allowed=false when user lacks permission', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('checkPermission', { allowed: false });

      const result = await client.checkPermission('user-1', 'admin', 'delete');

      expect(result.allowed).toBe(false);
    });
  });

  // -----------------------------------------------------------------------
  // introspectToken
  // -----------------------------------------------------------------------
  describe('introspectToken', () => {
    it('should introspect token and return detailed information', async () => {
      const client = createClient();
      await client.connect();

      const mockResponse = {
        active: true,
        userId: 'user-123',
        email: 'test@example.com',
        roles: ['admin'],
        expiresAt: '2025-01-01T00:00:00Z',
        issuedAt: '2024-12-01T00:00:00Z',
      };
      setupGrpcMethodResponse('introspectToken', mockResponse);

      const result = await client.introspectToken('jwt-token-for-introspection');

      expect(result.active).toBe(true);
      expect(result.userId).toBe('user-123');
    });
  });

  // -----------------------------------------------------------------------
  // createUser
  // -----------------------------------------------------------------------
  describe('createUser', () => {
    it('should create a user via gRPC', async () => {
      const client = createClient();
      await client.connect();

      const mockResponse = {
        userId: 'user-new',
        accessToken: 'new-access-token',
        refreshToken: 'new-refresh-token',
      };
      setupGrpcMethodResponse('createUser', mockResponse);

      const result = await client.createUser({
        email: 'new@example.com',
        password: 'Secure123!',
        username: 'newuser',
        fullName: 'New User',
        applicationId: '',
      });

      expect(result.userId).toBe('user-new');
      expect(result.accessToken).toBe('new-access-token');

      const callArgs = mockGrpcMethods['createUser']!.mock.calls[0]!;
      expect(callArgs[0].email).toBe('new@example.com');
    });

    it('should throw on duplicate email', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('createUser', null, {
        message: 'ALREADY_EXISTS: Email already registered',
        code: 6,
      });

      await expect(
        client.createUser({
          email: 'existing@example.com',
          password: 'Pass123!',
          username: 'existing',
          fullName: 'Existing',
          applicationId: '',
        })
      ).rejects.toThrow('ALREADY_EXISTS');
    });
  });

  // -----------------------------------------------------------------------
  // login
  // -----------------------------------------------------------------------
  describe('login', () => {
    it('should login and return tokens', async () => {
      const client = createClient();
      await client.connect();

      const mockResponse = {
        accessToken: 'access-123',
        refreshToken: 'refresh-456',
        userId: 'user-42',
        requires_2fa: false,
      };
      setupGrpcMethodResponse('login', mockResponse);

      const result = await client.login({
        email: 'user@example.com',
        password: 'Password123!',
        applicationId: '',
      });

      expect(result.accessToken).toBe('access-123');
      expect(result.userId).toBe('user-42');
    });

    it('should propagate invalid credentials error', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('login', null, {
        message: 'UNAUTHENTICATED: Invalid credentials',
        code: 16,
      });

      await expect(
        client.login({
          email: 'user@example.com',
          password: 'wrong',
          applicationId: '',
        })
      ).rejects.toThrow('UNAUTHENTICATED: Invalid credentials');
    });
  });

  // -----------------------------------------------------------------------
  // OTP methods
  // -----------------------------------------------------------------------
  describe('sendOTP', () => {
    it('should send OTP and return success', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('sendOTP', { success: true, message: 'OTP sent' });

      const result = await client.sendOTP({
        email: 'user@example.com',
        otpType: 'login',
        applicationId: '',
      });

      expect(result.success).toBe(true);
      expect(result.message).toBe('OTP sent');
    });
  });

  describe('verifyOTP', () => {
    it('should verify OTP and return success', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('verifyOTP', { valid: true });

      const result = await client.verifyOTP({
        email: 'user@example.com',
        code: '123456',
        applicationId: '',
      });

      expect(result.valid).toBe(true);
    });
  });

  // -----------------------------------------------------------------------
  // sendEmail
  // -----------------------------------------------------------------------
  describe('sendEmail', () => {
    it('should send email via template', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('sendEmail', { success: true, messageId: 'msg-123' });

      const result = await client.sendEmail({
        templateType: 'welcome',
        toEmail: 'user@example.com',
        variables: { name: 'John' },
        applicationId: '',
      });

      expect(result.success).toBe(true);
      expect(result.messageId).toBe('msg-123');
    });
  });

  // -----------------------------------------------------------------------
  // syncUsers
  // -----------------------------------------------------------------------
  describe('syncUsers', () => {
    it('should sync users with Date object', async () => {
      const client = createClient();
      await client.connect();

      const mockResponse = {
        users: [{ id: 'u1', email: 'a@b.com' }],
        total: 1,
      };
      setupGrpcMethodResponse('syncUsers', mockResponse);

      const date = new Date('2024-06-01T00:00:00Z');
      const result = await client.syncUsers({ updatedAfter: date, limit: 50 });

      expect(result.users).toHaveLength(1);

      const callArgs = mockGrpcMethods['syncUsers']!.mock.calls[0]!;
      expect(callArgs[0].updatedAfter).toBe('2024-06-01T00:00:00.000Z');
      expect(callArgs[0].limit).toBe(50);
    });

    it('should sync users with string timestamp', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('syncUsers', { users: [], total: 0 });

      await client.syncUsers({ updatedAfter: '2024-01-01T00:00:00Z' });

      const callArgs = mockGrpcMethods['syncUsers']!.mock.calls[0]!;
      expect(callArgs[0].updatedAfter).toBe('2024-01-01T00:00:00Z');
    });

    it('should use default limit and offset', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('syncUsers', { users: [], total: 0 });

      await client.syncUsers({ updatedAfter: '2024-01-01T00:00:00Z' });

      const callArgs = mockGrpcMethods['syncUsers']!.mock.calls[0]!;
      expect(callArgs[0].limit).toBe(100);
      expect(callArgs[0].offset).toBe(0);
    });
  });

  // -----------------------------------------------------------------------
  // getApplicationAuthConfig
  // -----------------------------------------------------------------------
  describe('getApplicationAuthConfig', () => {
    it('should get auth config for application', async () => {
      const client = createClient();
      await client.connect();

      const mockResponse = {
        allowedAuthMethods: ['password', 'otp_email'],
        oauthProviders: ['google'],
      };
      setupGrpcMethodResponse('getApplicationAuthConfig', mockResponse);

      const result = await client.getApplicationAuthConfig('app-123');

      expect(result.allowedAuthMethods).toContain('password');

      const callArgs = mockGrpcMethods['getApplicationAuthConfig']!.mock.calls[0]!;
      expect(callArgs[0]).toEqual({ applicationId: 'app-123' });
    });
  });

  // -----------------------------------------------------------------------
  // Token exchange
  // -----------------------------------------------------------------------
  describe('createTokenExchange', () => {
    it('should create a token exchange code', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('createTokenExchange', {
        exchangeCode: 'exch-abc-123',
        expiresAt: '2024-12-01T01:00:00Z',
      });

      const result = await client.createTokenExchange('my-access-token', 'target-app-id');

      expect(result.exchangeCode).toBe('exch-abc-123');

      const callArgs = mockGrpcMethods['createTokenExchange']!.mock.calls[0]!;
      expect(callArgs[0]).toEqual({
        accessToken: 'my-access-token',
        targetApplicationId: 'target-app-id',
      });
    });
  });

  describe('redeemTokenExchange', () => {
    it('should redeem a token exchange code', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('redeemTokenExchange', {
        accessToken: 'target-access-token',
        refreshToken: 'target-refresh-token',
        userId: 'user-123',
      });

      const result = await client.redeemTokenExchange('exch-abc-123');

      expect(result.accessToken).toBe('target-access-token');

      const callArgs = mockGrpcMethods['redeemTokenExchange']!.mock.calls[0]!;
      expect(callArgs[0]).toEqual({ exchangeCode: 'exch-abc-123' });
    });
  });

  // -----------------------------------------------------------------------
  // Call options (timeout, metadata)
  // -----------------------------------------------------------------------
  describe('call options', () => {
    it('should pass custom metadata to gRPC calls', async () => {
      const client = createClient({ apiKey: 'agw_key' });
      await client.connect();

      setupGrpcMethodResponse('validateToken', { valid: true });

      await client.validateToken('token-xyz-12345678901', {
        metadata: { 'x-custom-header': 'custom-value' },
      });

      // Metadata constructor should have been invoked during the call
      const { Metadata } = await import('@grpc/grpc-js');
      expect(Metadata).toHaveBeenCalled();
    });

    it('should pass timeout as deadline in call options', async () => {
      const client = createClient();
      await client.connect();

      setupGrpcMethodResponse('getUser', { user: { id: 'u1' } });

      const beforeCall = Date.now();
      await client.getUser('u1', { timeout: 3000 });

      const callArgs = mockGrpcMethods['getUser']!.mock.calls[0]!;
      const callOptions = callArgs[2]; // Third arg is callOptions
      expect(callOptions.deadline).toBeGreaterThanOrEqual(beforeCall + 3000);
    });
  });

  // -----------------------------------------------------------------------
  // createGrpcClient factory
  // -----------------------------------------------------------------------
  describe('createGrpcClient', () => {
    it('should create an AuthGrpcClient instance', () => {
      const client = createGrpcClient({ address: 'localhost:50051' });

      expect(client).toBeInstanceOf(AuthGrpcClient);
      expect(client.isConnected()).toBe(false);
    });
  });
});
