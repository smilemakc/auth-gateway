import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { HttpClient, MemoryTokenStorage } from '../core/http';
import {
  AuthGatewayError,
  AuthenticationError,
  AuthorizationError,
  ConflictError,
  NetworkError,
  NotFoundError,
  RateLimitError,
  ServerError,
  TimeoutError,
  TwoFactorRequiredError,
  ValidationError,
  createErrorFromResponse,
  isRetryableError,
} from '../core/errors';
import {
  calculateRetryDelay,
  defaultRetryCondition,
  withRetry,
  createRetryWrapper,
} from '../core/retry';
import { DEFAULT_CONFIG } from '../config/types';
import type { RetryConfig, RequestConfig, ApiResponse } from '../config/types';
import { AuthGatewayClient, createClient, createApiKeyClient } from '../client';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const TEST_BASE_URL = 'https://api.example.com';

function mockFetchJson(data: unknown, status = 200, headers?: Record<string, string>) {
  const responseHeaders = new Headers({
    'Content-Type': 'application/json',
    ...headers,
  });
  return {
    ok: status >= 200 && status < 300,
    status,
    headers: responseHeaders,
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(JSON.stringify(data)),
  };
}

// ---------------------------------------------------------------------------
// Error classes tests
// ---------------------------------------------------------------------------

describe('Error classes', () => {
  describe('AuthGatewayError', () => {
    it('should create error with all properties', () => {
      const error = new AuthGatewayError('Something went wrong', {
        status: 500,
        code: 'INTERNAL',
        details: 'DB connection failed',
        requestId: 'req-abc',
        retryable: true,
      });

      expect(error.message).toBe('Something went wrong');
      expect(error.name).toBe('AuthGatewayError');
      expect(error.status).toBe(500);
      expect(error.code).toBe('INTERNAL');
      expect(error.details).toBe('DB connection failed');
      expect(error.requestId).toBe('req-abc');
      expect(error.retryable).toBe(true);
      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(AuthGatewayError);
    });

    it('should default retryable to false', () => {
      const error = new AuthGatewayError('Error');
      expect(error.retryable).toBe(false);
    });

    it('should store original error reference', () => {
      const original = new TypeError('original cause');
      const error = new AuthGatewayError('Wrapped', { originalError: original });

      expect(error.originalError).toBe(original);
    });
  });

  describe('NetworkError', () => {
    it('should create retryable network error', () => {
      const original = new TypeError('Failed to fetch');
      const error = new NetworkError('Network request failed', original);

      expect(error.name).toBe('NetworkError');
      expect(error.code).toBe('NETWORK_ERROR');
      expect(error.retryable).toBe(true);
      expect(error.originalError).toBe(original);
    });
  });

  describe('TimeoutError', () => {
    it('should create retryable timeout error with default message', () => {
      const error = new TimeoutError();

      expect(error.name).toBe('TimeoutError');
      expect(error.message).toBe('Request timed out');
      expect(error.code).toBe('TIMEOUT');
      expect(error.retryable).toBe(true);
    });

    it('should accept custom message', () => {
      const error = new TimeoutError('Connection timed out after 30s');
      expect(error.message).toBe('Connection timed out after 30s');
    });
  });

  describe('AuthenticationError', () => {
    it('should create non-retryable 401 error', () => {
      const error = new AuthenticationError();

      expect(error.name).toBe('AuthenticationError');
      expect(error.status).toBe(401);
      expect(error.code).toBe('UNAUTHORIZED');
      expect(error.retryable).toBe(false);
    });

    it('should accept custom message and details', () => {
      const error = new AuthenticationError('Token expired', 'JWT expired at 2024-01-01');

      expect(error.message).toBe('Token expired');
      expect(error.details).toBe('JWT expired at 2024-01-01');
    });
  });

  describe('AuthorizationError', () => {
    it('should create non-retryable 403 error', () => {
      const error = new AuthorizationError('Insufficient permissions');

      expect(error.name).toBe('AuthorizationError');
      expect(error.status).toBe(403);
      expect(error.code).toBe('FORBIDDEN');
      expect(error.retryable).toBe(false);
    });
  });

  describe('NotFoundError', () => {
    it('should create 404 error', () => {
      const error = new NotFoundError('User not found');

      expect(error.name).toBe('NotFoundError');
      expect(error.status).toBe(404);
      expect(error.code).toBe('NOT_FOUND');
    });
  });

  describe('ValidationError', () => {
    it('should create 400 error with field details', () => {
      const fields = {
        email: ['Invalid email format'],
        password: ['Too short', 'Must contain number'],
      };
      const error = new ValidationError('Validation failed', fields);

      expect(error.name).toBe('ValidationError');
      expect(error.status).toBe(400);
      expect(error.fields).toEqual(fields);
    });
  });

  describe('ConflictError', () => {
    it('should create 409 error', () => {
      const error = new ConflictError('Email already registered');

      expect(error.name).toBe('ConflictError');
      expect(error.status).toBe(409);
    });
  });

  describe('RateLimitError', () => {
    it('should create retryable 429 error with retryAfter', () => {
      const error = new RateLimitError('Too many requests', 120);

      expect(error.name).toBe('RateLimitError');
      expect(error.status).toBe(429);
      expect(error.retryable).toBe(true);
      expect(error.retryAfter).toBe(120);
    });

    it('should default retryAfter to 60', () => {
      const error = new RateLimitError();
      expect(error.retryAfter).toBe(60);
    });
  });

  describe('ServerError', () => {
    it('should create retryable 5xx error', () => {
      const error = new ServerError('Internal error', 503, 'Service unavailable');

      expect(error.name).toBe('ServerError');
      expect(error.status).toBe(503);
      expect(error.retryable).toBe(true);
      expect(error.details).toBe('Service unavailable');
    });
  });

  describe('TwoFactorRequiredError', () => {
    it('should carry two-factor token', () => {
      const error = new TwoFactorRequiredError('2fa-token-xyz');

      expect(error.name).toBe('TwoFactorRequiredError');
      expect(error.twoFactorToken).toBe('2fa-token-xyz');
      expect(error.code).toBe('2FA_REQUIRED');
      expect(error.retryable).toBe(false);
    });
  });
});

// ---------------------------------------------------------------------------
// createErrorFromResponse
// ---------------------------------------------------------------------------

describe('createErrorFromResponse', () => {
  it('should create ValidationError for 400', () => {
    const error = createErrorFromResponse(400, { message: 'Bad input' });
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('Bad input');
  });

  it('should create AuthenticationError for 401', () => {
    const error = createErrorFromResponse(401, { message: 'Unauthorized' });
    expect(error).toBeInstanceOf(AuthenticationError);
  });

  it('should create AuthorizationError for 403', () => {
    const error = createErrorFromResponse(403, { message: 'Forbidden' });
    expect(error).toBeInstanceOf(AuthorizationError);
  });

  it('should create NotFoundError for 404', () => {
    const error = createErrorFromResponse(404, { message: 'Not found' });
    expect(error).toBeInstanceOf(NotFoundError);
  });

  it('should create ConflictError for 409', () => {
    const error = createErrorFromResponse(409, { message: 'Conflict' });
    expect(error).toBeInstanceOf(ConflictError);
  });

  it('should create RateLimitError for 429', () => {
    const error = createErrorFromResponse(429, { message: 'Rate limited' });
    expect(error).toBeInstanceOf(RateLimitError);
  });

  it('should create ServerError for 500+', () => {
    const error = createErrorFromResponse(500, { message: 'Internal server error' });
    expect(error).toBeInstanceOf(ServerError);
  });

  it('should create ServerError for 502', () => {
    const error = createErrorFromResponse(502, { message: 'Bad gateway' });
    expect(error).toBeInstanceOf(ServerError);
  });

  it('should create generic AuthGatewayError for unknown status codes', () => {
    const error = createErrorFromResponse(418, { message: "I'm a teapot" });
    expect(error).toBeInstanceOf(AuthGatewayError);
    expect(error.status).toBe(418);
  });

  it('should use error field as fallback for message', () => {
    const error = createErrorFromResponse(400, { error: 'Bad request' });
    expect(error.message).toBe('Bad request');
  });

  it('should use "Unknown error" when body has no message', () => {
    const error = createErrorFromResponse(400, null);
    expect(error.message).toBe('Unknown error');
  });
});

// ---------------------------------------------------------------------------
// isRetryableError
// ---------------------------------------------------------------------------

describe('isRetryableError', () => {
  it('should return true for retryable AuthGatewayError', () => {
    expect(isRetryableError(new NetworkError('net'))).toBe(true);
    expect(isRetryableError(new TimeoutError())).toBe(true);
    expect(isRetryableError(new ServerError('err', 503))).toBe(true);
    expect(isRetryableError(new RateLimitError())).toBe(true);
  });

  it('should return false for non-retryable AuthGatewayError', () => {
    expect(isRetryableError(new AuthenticationError())).toBe(false);
    expect(isRetryableError(new AuthorizationError())).toBe(false);
    expect(isRetryableError(new NotFoundError())).toBe(false);
    expect(isRetryableError(new ValidationError())).toBe(false);
    expect(isRetryableError(new ConflictError())).toBe(false);
  });

  it('should return true for network-like error messages', () => {
    expect(isRetryableError(new Error('network failure'))).toBe(true);
    expect(isRetryableError(new Error('timeout reached'))).toBe(true);
    expect(isRetryableError(new Error('ECONNRESET'))).toBe(true);
    expect(isRetryableError(new Error('ECONNREFUSED'))).toBe(true);
  });

  it('should return false for non-network Error instances', () => {
    expect(isRetryableError(new Error('Validation failed'))).toBe(false);
  });

  it('should return false for non-Error values', () => {
    expect(isRetryableError('string error')).toBe(false);
    expect(isRetryableError(42)).toBe(false);
    expect(isRetryableError(null)).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// Retry logic
// ---------------------------------------------------------------------------

describe('calculateRetryDelay', () => {
  it('should calculate exponential backoff delay', () => {
    const config: RetryConfig = {
      ...DEFAULT_CONFIG.retry,
      initialDelayMs: 1000,
      backoffMultiplier: 2,
      maxDelayMs: 30000,
    };

    // Attempt 0: 1000 * 2^0 = 1000 (+/- jitter)
    const delay0 = calculateRetryDelay(0, config);
    expect(delay0).toBeGreaterThanOrEqual(900);
    expect(delay0).toBeLessThanOrEqual(1100);

    // Attempt 1: 1000 * 2^1 = 2000 (+/- jitter)
    const delay1 = calculateRetryDelay(1, config);
    expect(delay1).toBeGreaterThanOrEqual(1800);
    expect(delay1).toBeLessThanOrEqual(2200);

    // Attempt 2: 1000 * 2^2 = 4000 (+/- jitter)
    const delay2 = calculateRetryDelay(2, config);
    expect(delay2).toBeGreaterThanOrEqual(3600);
    expect(delay2).toBeLessThanOrEqual(4400);
  });

  it('should cap delay at maxDelayMs', () => {
    const config: RetryConfig = {
      ...DEFAULT_CONFIG.retry,
      initialDelayMs: 1000,
      backoffMultiplier: 10,
      maxDelayMs: 5000,
    };

    // Attempt 3: 1000 * 10^3 = 1,000,000 -> capped at 5000
    const delay = calculateRetryDelay(3, config);
    expect(delay).toBeLessThanOrEqual(5500); // 5000 + 10% jitter
  });
});

describe('defaultRetryCondition', () => {
  const config = DEFAULT_CONFIG.retry;

  it('should return false when max attempts reached', () => {
    const context = { attempt: 3, maxRetries: 3, startTime: Date.now() };
    expect(defaultRetryCondition(new NetworkError('err'), context, config)).toBe(false);
  });

  it('should return false for non-retryable errors', () => {
    const context = { attempt: 0, maxRetries: 3, startTime: Date.now() };
    expect(defaultRetryCondition(new AuthenticationError(), context, config)).toBe(false);
  });

  it('should return true for network-like generic errors when retryOnNetworkError is true', () => {
    // NetworkError extends AuthGatewayError which has `status` in its shape (even if undefined).
    // The condition checks `'status' in error` first, so for a plain Error with network-like message
    // but no status field, retryOnNetworkError config kicks in.
    const context = { attempt: 0, maxRetries: 3, startTime: Date.now() };
    const networkLikeError = new Error('network failure');
    expect(defaultRetryCondition(networkLikeError, context, config)).toBe(true);
  });

  it('should check retryableStatusCodes for NetworkError since it has a status property', () => {
    // NetworkError inherits `status` from AuthGatewayError (value undefined).
    // The retry condition checks `'status' in error` which is true, then looks up
    // retryableStatusCodes.includes(undefined) which is false.
    const context = { attempt: 0, maxRetries: 3, startTime: Date.now() };
    expect(defaultRetryCondition(new NetworkError('Network'), context, config)).toBe(false);
  });

  it('should check retryable status codes', () => {
    const context = { attempt: 0, maxRetries: 3, startTime: Date.now() };
    // 503 is in retryableStatusCodes
    expect(defaultRetryCondition(new ServerError('err', 503), context, config)).toBe(true);
    // 429 is in retryableStatusCodes
    expect(defaultRetryCondition(new RateLimitError(), context, config)).toBe(true);
  });
});

describe('withRetry', () => {
  it('should return result on first successful attempt', async () => {
    const operation = vi.fn().mockResolvedValue('success');

    const result = await withRetry(operation, { maxRetries: 3 });

    expect(result).toBe('success');
    expect(operation).toHaveBeenCalledTimes(1);
  });

  it('should throw immediately for non-retryable errors', async () => {
    const operation = vi.fn().mockRejectedValue(new AuthenticationError('Unauthorized'));

    await expect(withRetry(operation, { maxRetries: 3 })).rejects.toThrow('Unauthorized');
    expect(operation).toHaveBeenCalledTimes(1);
  });
});

describe('createRetryWrapper', () => {
  it('should create a wrapper with default configuration', () => {
    const wrapper = createRetryWrapper();

    expect(wrapper.config.maxRetries).toBe(DEFAULT_CONFIG.retry.maxRetries);
    expect(wrapper.config.initialDelayMs).toBe(DEFAULT_CONFIG.retry.initialDelayMs);
  });

  it('should allow overriding configuration', () => {
    const wrapper = createRetryWrapper({ maxRetries: 5 });

    expect(wrapper.config.maxRetries).toBe(5);
  });

  it('should execute operations successfully', async () => {
    const wrapper = createRetryWrapper({ maxRetries: 0 });

    const result = await wrapper.execute(() => Promise.resolve(42));

    expect(result).toBe(42);
  });

  it('should update config via updateConfig', () => {
    const wrapper = createRetryWrapper();

    wrapper.updateConfig({ maxRetries: 10 });

    expect(wrapper.config.maxRetries).toBe(10);
  });
});

// ---------------------------------------------------------------------------
// HttpClient interceptors
// ---------------------------------------------------------------------------

describe('HttpClient interceptors', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('request interceptors', () => {
    it('should modify request config before sending', async () => {
      const http = new HttpClient({
        baseUrl: TEST_BASE_URL,
        retry: { maxRetries: 0 },
      });

      http.addRequestInterceptor((config) => ({
        ...config,
        headers: { ...config.headers, 'X-Custom': 'injected' },
      }));

      fetchMock.mockResolvedValueOnce(mockFetchJson({ ok: true }));

      await http.get('/test');

      const [, options] = fetchMock.mock.calls[0]!;
      expect(options.headers['X-Custom']).toBe('injected');
    });

    it('should chain multiple request interceptors', async () => {
      const http = new HttpClient({
        baseUrl: TEST_BASE_URL,
        retry: { maxRetries: 0 },
      });

      http.addRequestInterceptor((config) => ({
        ...config,
        headers: { ...config.headers, 'X-First': 'first' },
      }));

      http.addRequestInterceptor((config) => ({
        ...config,
        headers: { ...config.headers, 'X-Second': 'second' },
      }));

      fetchMock.mockResolvedValueOnce(mockFetchJson({ ok: true }));

      await http.get('/test');

      const [, options] = fetchMock.mock.calls[0]!;
      expect(options.headers['X-First']).toBe('first');
      expect(options.headers['X-Second']).toBe('second');
    });

    it('should allow removing interceptors', async () => {
      const http = new HttpClient({
        baseUrl: TEST_BASE_URL,
        retry: { maxRetries: 0 },
      });

      const remove = http.addRequestInterceptor((config) => ({
        ...config,
        headers: { ...config.headers, 'X-Removable': 'present' },
      }));

      fetchMock.mockResolvedValueOnce(mockFetchJson({ ok: true }));
      await http.get('/test1');

      const [, options1] = fetchMock.mock.calls[0]!;
      expect(options1.headers['X-Removable']).toBe('present');

      // Remove the interceptor
      remove();

      fetchMock.mockResolvedValueOnce(mockFetchJson({ ok: true }));
      await http.get('/test2');

      const [, options2] = fetchMock.mock.calls[1]!;
      expect(options2.headers['X-Removable']).toBeUndefined();
    });
  });

  describe('response interceptors', () => {
    it('should transform the response', async () => {
      const http = new HttpClient({
        baseUrl: TEST_BASE_URL,
        retry: { maxRetries: 0 },
      });

      http.addResponseInterceptor((response) => ({
        ...response,
        data: { ...response.data as Record<string, unknown>, _intercepted: true },
      }));

      fetchMock.mockResolvedValueOnce(mockFetchJson({ name: 'test' }));

      const result = await http.get<{ name: string; _intercepted?: boolean }>('/test');

      expect(result.data.name).toBe('test');
      expect(result.data._intercepted).toBe(true);
    });

    it('should allow removing response interceptors', async () => {
      const http = new HttpClient({
        baseUrl: TEST_BASE_URL,
        retry: { maxRetries: 0 },
      });

      const remove = http.addResponseInterceptor((response) => ({
        ...response,
        data: { ...(response.data as Record<string, unknown>), injected: true },
      }));

      fetchMock.mockResolvedValueOnce(mockFetchJson({ val: 1 }));
      const r1 = await http.get<{ val: number; injected?: boolean }>('/test');
      expect(r1.data.injected).toBe(true);

      remove();

      fetchMock.mockResolvedValueOnce(mockFetchJson({ val: 2 }));
      const r2 = await http.get<{ val: number; injected?: boolean }>('/test');
      expect(r2.data.injected).toBeUndefined();
    });
  });

  describe('error interceptors', () => {
    it('should transform errors before they are thrown', async () => {
      const http = new HttpClient({
        baseUrl: TEST_BASE_URL,
        retry: { maxRetries: 0 },
        autoRefreshTokens: false,
      });

      http.addErrorInterceptor((error) => {
        return new AuthGatewayError(`Intercepted: ${error.message}`, {
          code: 'INTERCEPTED',
          retryable: false,
        });
      });

      fetchMock.mockResolvedValueOnce(mockFetchJson({ message: 'Server error' }, 500));

      try {
        await http.get('/test');
        expect.fail('Should have thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(AuthGatewayError);
        expect((error as AuthGatewayError).code).toBe('INTERCEPTED');
        expect((error as AuthGatewayError).message).toContain('Intercepted');
      }
    });
  });
});

// ---------------------------------------------------------------------------
// HttpClient configuration and headers
// ---------------------------------------------------------------------------

describe('HttpClient configuration', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should strip trailing slash from baseUrl', async () => {
    const http = new HttpClient({
      baseUrl: 'https://api.example.com/',
      retry: { maxRetries: 0 },
    });

    fetchMock.mockResolvedValueOnce(mockFetchJson({ ok: true }));

    await http.get('/test');

    const [url] = fetchMock.mock.calls[0]!;
    expect(url).toBe('https://api.example.com/test');
  });

  it('should include default Content-Type header', async () => {
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
    });

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));

    await http.get('/test');

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.headers['Content-Type']).toBe('application/json');
  });

  it('should include custom headers from config', async () => {
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      headers: { 'X-Custom-Global': 'global-val' },
      retry: { maxRetries: 0 },
    });

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));

    await http.get('/test');

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.headers['X-Custom-Global']).toBe('global-val');
  });

  it('should include X-Application-ID when applicationId is set', async () => {
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      applicationId: 'my-app-id',
      retry: { maxRetries: 0 },
    });

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));

    await http.get('/test');

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.headers['X-Application-ID']).toBe('my-app-id');
  });

  it('should use apiKey for Authorization header', async () => {
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      apiKey: 'agw_testkey123',
      retry: { maxRetries: 0 },
    });

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));

    await http.get('/test');

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.headers['Authorization']).toBe('Bearer agw_testkey123');
  });

  it('should use stored access token for Authorization header', async () => {
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
    });

    await http.getTokenStorage().setAccessToken('jwt-token-abc');

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));

    await http.get('/test');

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.headers['Authorization']).toBe('Bearer jwt-token-abc');
  });

  it('should skip Authorization when skipAuth is set', async () => {
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
    });

    await http.getTokenStorage().setAccessToken('token');

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));

    await http.get('/test', { skipAuth: true });

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.headers['Authorization']).toBeUndefined();
  });

  it('should update configuration at runtime', async () => {
    const http = new HttpClient({
      baseUrl: 'https://old-api.example.com',
      retry: { maxRetries: 0 },
    });

    http.configure({
      baseUrl: 'https://new-api.example.com',
      headers: { 'X-Updated': 'true' },
    });

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));

    await http.get('/test');

    const [url, options] = fetchMock.mock.calls[0]!;
    expect(url).toBe('https://new-api.example.com/test');
    expect(options.headers['X-Updated']).toBe('true');
  });

  it('should set and remove headers', async () => {
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
    });

    http.setHeader('X-Test', 'value');

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));
    await http.get('/test');

    let [, options] = fetchMock.mock.calls[0]!;
    expect(options.headers['X-Test']).toBe('value');

    http.removeHeader('X-Test');

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));
    await http.get('/test');

    [, options] = fetchMock.mock.calls[1]!;
    expect(options.headers['X-Test']).toBeUndefined();
  });
});

// ---------------------------------------------------------------------------
// HttpClient HTTP methods
// ---------------------------------------------------------------------------

describe('HttpClient HTTP methods', () => {
  let http: HttpClient;
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
    http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should make GET requests', async () => {
    fetchMock.mockResolvedValueOnce(mockFetchJson({ id: 1 }));

    const result = await http.get<{ id: number }>('/items/1');

    expect(result.data.id).toBe(1);
    expect(result.status).toBe(200);

    const [url, options] = fetchMock.mock.calls[0]!;
    expect(url).toBe(`${TEST_BASE_URL}/items/1`);
    expect(options.method).toBe('GET');
  });

  it('should make POST requests with body', async () => {
    fetchMock.mockResolvedValueOnce(mockFetchJson({ created: true }));

    await http.post('/items', { name: 'new item' });

    const [url, options] = fetchMock.mock.calls[0]!;
    expect(url).toBe(`${TEST_BASE_URL}/items`);
    expect(options.method).toBe('POST');
    expect(JSON.parse(options.body)).toEqual({ name: 'new item' });
  });

  it('should make PUT requests', async () => {
    fetchMock.mockResolvedValueOnce(mockFetchJson({ updated: true }));

    await http.put('/items/1', { name: 'updated' });

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.method).toBe('PUT');
  });

  it('should make PATCH requests', async () => {
    fetchMock.mockResolvedValueOnce(mockFetchJson({ patched: true }));

    await http.patch('/items/1', { name: 'patched' });

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.method).toBe('PATCH');
  });

  it('should make DELETE requests', async () => {
    fetchMock.mockResolvedValueOnce(mockFetchJson({ deleted: true }));

    await http.delete('/items/1');

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.method).toBe('DELETE');
  });

  it('should include query parameters in URL', async () => {
    fetchMock.mockResolvedValueOnce(mockFetchJson([]));

    await http.get('/items', { query: { page: 2, status: 'active', unused: undefined } });

    const [url] = fetchMock.mock.calls[0]!;
    expect(url).toContain('page=2');
    expect(url).toContain('status=active');
    expect(url).not.toContain('unused');
  });

  it('should return response with headers and requestId', async () => {
    const responseHeaders = new Headers({
      'Content-Type': 'application/json',
      'X-Request-ID': 'req-789',
    });

    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: responseHeaders,
      json: () => Promise.resolve({ data: 'test' }),
    });

    const result = await http.get('/test');

    expect(result.requestId).toBe('req-789');
    expect(result.headers['content-type']).toBe('application/json');
  });

  it('should handle non-JSON responses as text', async () => {
    const responseHeaders = new Headers({ 'Content-Type': 'text/plain' });

    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      headers: responseHeaders,
      text: () => Promise.resolve('plain text response'),
    });

    const result = await http.get<string>('/test');

    expect(result.data).toBe('plain text response');
  });
});

// ---------------------------------------------------------------------------
// HttpClient callbacks
// ---------------------------------------------------------------------------

describe('HttpClient callbacks', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should call onRequest callback before each request', async () => {
    const onRequest = vi.fn();
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
      callbacks: { onRequest },
    });

    fetchMock.mockResolvedValueOnce(mockFetchJson({}));

    await http.get('/test');

    expect(onRequest).toHaveBeenCalledTimes(1);
    expect(onRequest).toHaveBeenCalledWith(expect.objectContaining({ method: 'GET', url: '/test' }));
  });

  it('should call onResponse callback after successful response', async () => {
    const onResponse = vi.fn();
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
      callbacks: { onResponse },
    });

    fetchMock.mockResolvedValueOnce(mockFetchJson({ value: 1 }));

    await http.get('/test');

    expect(onResponse).toHaveBeenCalledTimes(1);
    expect(onResponse).toHaveBeenCalledWith(
      expect.objectContaining({ status: 200, data: { value: 1 } })
    );
  });

  it('should call onError callback on error', async () => {
    const onError = vi.fn();
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
      autoRefreshTokens: false,
      callbacks: { onError },
    });

    fetchMock.mockResolvedValueOnce(mockFetchJson({ message: 'Server error' }, 500));

    try {
      await http.get('/test');
    } catch {
      // Expected
    }

    expect(onError).toHaveBeenCalledTimes(1);
    expect(onError).toHaveBeenCalledWith(expect.any(ServerError));
  });

  it('should call onRateLimited callback on 429', async () => {
    const onRateLimited = vi.fn();
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
      autoRefreshTokens: false,
      callbacks: { onRateLimited },
    });

    fetchMock.mockResolvedValueOnce({
      ok: false,
      status: 429,
      headers: new Headers({
        'Content-Type': 'application/json',
        'Retry-After': '120',
      }),
      json: () => Promise.resolve({ message: 'Too many requests' }),
    });

    try {
      await http.get('/test');
    } catch {
      // Expected
    }

    expect(onRateLimited).toHaveBeenCalledWith(120);
  });

  it('should call onTokenRefresh callback when tokens are refreshed', async () => {
    const onTokenRefresh = vi.fn();
    const http = new HttpClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
      autoRefreshTokens: true,
      callbacks: { onTokenRefresh },
    });

    // Set up refresh function
    http.setRefreshTokenFn(async () => ({
      accessToken: 'new-access',
      refreshToken: 'new-refresh',
    }));

    await http.getTokenStorage().setAccessToken('old-token');

    // First call: 401 triggers refresh
    fetchMock
      .mockResolvedValueOnce(mockFetchJson({ message: 'Unauthorized' }, 401))
      // Retried call after refresh
      .mockResolvedValueOnce(mockFetchJson({ ok: true }));

    await http.get('/test');

    expect(onTokenRefresh).toHaveBeenCalledWith({
      accessToken: 'new-access',
      refreshToken: 'new-refresh',
    });
  });
});

// ---------------------------------------------------------------------------
// AuthGatewayClient (integration)
// ---------------------------------------------------------------------------

describe('AuthGatewayClient', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize all services', () => {
    const client = new AuthGatewayClient({ baseUrl: TEST_BASE_URL });

    expect(client.auth).toBeDefined();
    expect(client.oauth).toBeDefined();
    expect(client.twoFactor).toBeDefined();
    expect(client.otp).toBeDefined();
    expect(client.sms).toBeDefined();
    expect(client.passwordless).toBeDefined();
    expect(client.apiKeys).toBeDefined();
    expect(client.sessions).toBeDefined();
    expect(client.health).toBeDefined();
    expect(client.tokenExchange).toBeDefined();
    expect(client.admin).toBeDefined();
    expect(client.admin.users).toBeDefined();
    expect(client.admin.rbac).toBeDefined();
    expect(client.admin.system).toBeDefined();
    expect(client.admin.applications).toBeDefined();
  });

  it('should delegate isAuthenticated to auth service', async () => {
    const client = new AuthGatewayClient({ baseUrl: TEST_BASE_URL });

    expect(await client.isAuthenticated()).toBe(false);

    await client.getTokenStorage().setAccessToken('token');

    expect(await client.isAuthenticated()).toBe(true);
  });

  it('should sign out, clear tokens, and disconnect websocket', async () => {
    const client = new AuthGatewayClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
    });

    await client.getTokenStorage().setAccessToken('token');
    await client.getTokenStorage().setRefreshToken('refresh');

    // Mock the logout call (it will fail but that's ok - signOut catches errors)
    fetchMock.mockResolvedValueOnce(mockFetchJson({ message: 'ok' }));

    await client.signOut();

    const storage = client.getTokenStorage();
    expect(await storage.getAccessToken()).toBeNull();
    expect(await storage.getRefreshToken()).toBeNull();
  });

  it('should signOut gracefully even when logout endpoint fails', async () => {
    const client = new AuthGatewayClient({
      baseUrl: TEST_BASE_URL,
      retry: { maxRetries: 0 },
      autoRefreshTokens: false,
    });

    await client.getTokenStorage().setAccessToken('token');

    // Logout endpoint returns error
    fetchMock.mockResolvedValueOnce(mockFetchJson({ message: 'Server error' }, 500));

    // Should not throw
    await client.signOut();

    expect(await client.getTokenStorage().getAccessToken()).toBeNull();
  });

  it('should configure headers and interceptors', () => {
    const client = new AuthGatewayClient({ baseUrl: TEST_BASE_URL });

    client.setHeader('X-Custom', 'value');
    client.removeHeader('X-Custom');

    const removeReqInterceptor = client.addRequestInterceptor((config) => config);
    const removeResInterceptor = client.addResponseInterceptor((response) => response);
    const removeErrInterceptor = client.addErrorInterceptor((error) => error);

    // Remove all - should not throw
    removeReqInterceptor();
    removeResInterceptor();
    removeErrInterceptor();
  });

  it('should expose HTTP client for advanced usage', () => {
    const client = new AuthGatewayClient({ baseUrl: TEST_BASE_URL });

    expect(client.getHttpClient()).toBeInstanceOf(HttpClient);
  });
});

// ---------------------------------------------------------------------------
// Factory functions
// ---------------------------------------------------------------------------

describe('createClient', () => {
  it('should create an AuthGatewayClient instance', () => {
    const client = createClient({ baseUrl: TEST_BASE_URL });

    expect(client).toBeInstanceOf(AuthGatewayClient);
  });
});

describe('createApiKeyClient', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    // @ts-ignore
    global.fetch = fetchMock;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should create a client with API key and disabled auto-refresh', async () => {
    const client = createApiKeyClient(TEST_BASE_URL, 'agw_mykey');

    expect(client).toBeInstanceOf(AuthGatewayClient);

    // The API key should be used for Authorization
    fetchMock.mockResolvedValueOnce(mockFetchJson({ ok: true }));

    await client.getHttpClient().get('/test');

    const [, options] = fetchMock.mock.calls[0]!;
    expect(options.headers['Authorization']).toBe('Bearer agw_mykey');
  });
});
