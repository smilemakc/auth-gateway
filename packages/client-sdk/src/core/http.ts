/**
 * HTTP client wrapper with interceptors, retry, and auth handling
 */

import type {
  ApiResponse,
  ClientCallbacks,
  ClientConfig,
  ErrorInterceptor,
  HttpMethod,
  RequestConfig,
  RequestInterceptor,
  ResponseInterceptor,
  RetryConfig,
  TokenStorage,
} from '../config/types';
import { DEFAULT_CONFIG } from '../config/types';
import {
  AuthenticationError,
  AuthGatewayError,
  createErrorFromResponse,
  NetworkError,
  RateLimitError,
  TimeoutError,
} from './errors';
import { withRetry } from './retry';

/** In-memory token storage (default) */
export class MemoryTokenStorage implements TokenStorage {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  getAccessToken(): string | null {
    return this.accessToken;
  }

  setAccessToken(token: string): void {
    this.accessToken = token;
  }

  getRefreshToken(): string | null {
    return this.refreshToken;
  }

  setRefreshToken(token: string): void {
    this.refreshToken = token;
  }

  clear(): void {
    this.accessToken = null;
    this.refreshToken = null;
  }
}

/** HTTP client with full features */
export class HttpClient {
  private baseUrl: string;
  private defaultHeaders: Record<string, string>;
  private timeout: number;
  private retryConfig: RetryConfig;
  private tokenStorage: TokenStorage;
  private callbacks: ClientCallbacks;
  private apiKey?: string;
  private autoRefreshTokens: boolean;
  private debug: boolean;
  private isRefreshing = false;
  private refreshPromise: Promise<void> | null = null;

  // Interceptors
  private requestInterceptors: RequestInterceptor[] = [];
  private responseInterceptors: ResponseInterceptor[] = [];
  private errorInterceptors: ErrorInterceptor[] = [];

  // Refresh token function (set by AuthService)
  private refreshTokenFn?: () => Promise<{
    accessToken: string;
    refreshToken: string;
  }>;

  constructor(config: ClientConfig) {
    this.baseUrl = config.baseUrl.replace(/\/$/, '');
    this.defaultHeaders = {
      'Content-Type': 'application/json',
      ...config.headers,
    };
    if (config.applicationId) {
      this.defaultHeaders['X-Application-ID'] = config.applicationId;
    }
    this.timeout = config.timeout ?? DEFAULT_CONFIG.timeout;
    this.retryConfig = { ...DEFAULT_CONFIG.retry, ...config.retry };
    this.tokenStorage = config.tokenStorage ?? new MemoryTokenStorage();
    this.callbacks = config.callbacks ?? {};
    this.apiKey = config.apiKey;
    this.autoRefreshTokens = config.autoRefreshTokens ?? DEFAULT_CONFIG.autoRefreshTokens;
    this.debug = config.debug ?? DEFAULT_CONFIG.debug;
  }

  /** Set refresh token function */
  setRefreshTokenFn(
    fn: () => Promise<{ accessToken: string; refreshToken: string }>
  ): void {
    this.refreshTokenFn = fn;
  }

  /** Get token storage */
  getTokenStorage(): TokenStorage {
    return this.tokenStorage;
  }

  /** Update configuration at runtime */
  configure(config: Partial<ClientConfig>): void {
    if (config.baseUrl) {
      this.baseUrl = config.baseUrl.replace(/\/$/, '');
    }
    if (config.headers) {
      this.defaultHeaders = { ...this.defaultHeaders, ...config.headers };
    }
    if (config.timeout !== undefined) {
      this.timeout = config.timeout;
    }
    if (config.retry) {
      this.retryConfig = { ...this.retryConfig, ...config.retry };
    }
    if (config.tokenStorage) {
      this.tokenStorage = config.tokenStorage;
    }
    if (config.callbacks) {
      this.callbacks = { ...this.callbacks, ...config.callbacks };
    }
    if (config.apiKey !== undefined) {
      this.apiKey = config.apiKey;
    }
    if (config.autoRefreshTokens !== undefined) {
      this.autoRefreshTokens = config.autoRefreshTokens;
    }
    if (config.debug !== undefined) {
      this.debug = config.debug;
    }
  }

  /** Set a specific header */
  setHeader(key: string, value: string): void {
    this.defaultHeaders[key] = value;
  }

  /** Remove a header */
  removeHeader(key: string): void {
    delete this.defaultHeaders[key];
  }

  /** Add request interceptor */
  addRequestInterceptor(interceptor: RequestInterceptor): () => void {
    this.requestInterceptors.push(interceptor);
    return () => {
      const index = this.requestInterceptors.indexOf(interceptor);
      if (index >= 0) {
        this.requestInterceptors.splice(index, 1);
      }
    };
  }

  /** Add response interceptor */
  addResponseInterceptor(interceptor: ResponseInterceptor): () => void {
    this.responseInterceptors.push(interceptor);
    return () => {
      const index = this.responseInterceptors.indexOf(interceptor);
      if (index >= 0) {
        this.responseInterceptors.splice(index, 1);
      }
    };
  }

  /** Add error interceptor */
  addErrorInterceptor(interceptor: ErrorInterceptor): () => void {
    this.errorInterceptors.push(interceptor);
    return () => {
      const index = this.errorInterceptors.indexOf(interceptor);
      if (index >= 0) {
        this.errorInterceptors.splice(index, 1);
      }
    };
  }

  /** Log debug message */
  private log(...args: unknown[]): void {
    if (this.debug) {
      console.log('[AuthGatewayClient]', ...args);
    }
  }

  /** Build query string from params */
  private buildQueryString(
    params?: Record<string, string | number | boolean | undefined>
  ): string {
    if (!params) return '';

    const searchParams = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined) {
        searchParams.append(key, String(value));
      }
    }

    const query = searchParams.toString();
    return query ? `?${query}` : '';
  }

  /** Get authorization header value */
  private async getAuthHeader(): Promise<string | null> {
    if (this.apiKey) {
      return `Bearer ${this.apiKey}`;
    }

    const token = await this.tokenStorage.getAccessToken();
    return token ? `Bearer ${token}` : null;
  }

  /** Refresh tokens */
  private async refreshTokens(): Promise<void> {
    if (!this.refreshTokenFn) {
      throw new AuthenticationError('No refresh token function configured');
    }

    // Prevent multiple concurrent refresh attempts
    if (this.isRefreshing && this.refreshPromise) {
      return this.refreshPromise;
    }

    this.isRefreshing = true;
    this.refreshPromise = (async () => {
      try {
        const { accessToken, refreshToken } = await this.refreshTokenFn!();
        await this.tokenStorage.setAccessToken(accessToken);
        await this.tokenStorage.setRefreshToken(refreshToken);
        this.callbacks.onTokenRefresh?.({ accessToken, refreshToken });
        this.log('Tokens refreshed successfully');
      } finally {
        this.isRefreshing = false;
        this.refreshPromise = null;
      }
    })();

    return this.refreshPromise;
  }

  /** Execute HTTP request */
  async request<T>(requestConfig: RequestConfig): Promise<ApiResponse<T>> {
    // Apply request interceptors
    let config = { ...requestConfig };
    for (const interceptor of this.requestInterceptors) {
      config = await interceptor(config);
    }

    // Build URL
    const url = `${this.baseUrl}${config.url}${this.buildQueryString(config.query)}`;

    // Build headers
    const headers: Record<string, string> = {
      ...this.defaultHeaders,
      ...config.headers,
    };

    // Add auth header if not skipped
    if (!config.skipAuth) {
      const authHeader = await this.getAuthHeader();
      if (authHeader) {
        headers['Authorization'] = authHeader;
      }
    }

    // Callback
    this.callbacks.onRequest?.(config);
    this.log(`${config.method} ${url}`);

    // Create abort controller for timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(
      () => controller.abort(),
      config.timeout ?? this.timeout
    );

    const executeRequest = async (): Promise<ApiResponse<T>> => {
      try {
        const response = await fetch(url, {
          method: config.method,
          headers,
          body: config.body ? JSON.stringify(config.body) : undefined,
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        const requestId = response.headers.get('X-Request-ID') ?? undefined;

        // Parse response body
        let data: T;
        const contentType = response.headers.get('Content-Type') ?? '';
        if (contentType.includes('application/json')) {
          data = await response.json();
        } else {
          data = (await response.text()) as unknown as T;
        }

        // Handle error responses
        if (!response.ok) {
          const error = createErrorFromResponse(
            response.status,
            data as { error?: string; message?: string; details?: string },
            requestId
          );

          // Handle rate limiting
          if (error instanceof RateLimitError) {
            const retryAfter = parseInt(
              response.headers.get('Retry-After') ?? '60',
              10
            );
            this.callbacks.onRateLimited?.(retryAfter);
            throw new RateLimitError(error.message, retryAfter);
          }

          // Handle authentication error
          if (
            error instanceof AuthenticationError &&
            this.autoRefreshTokens &&
            !config.skipAuth
          ) {
            // Try to refresh tokens
            try {
              await this.refreshTokens();
              // Retry the request with new token
              const newAuthHeader = await this.getAuthHeader();
              if (newAuthHeader) {
                headers['Authorization'] = newAuthHeader;
              }
              return this.request({ ...config, skipAuth: true });
            } catch {
              this.callbacks.onAuthFailure?.();
              throw error;
            }
          }

          throw error;
        }

        // Build response
        const responseHeaders: Record<string, string> = {};
        response.headers.forEach((value, key) => {
          responseHeaders[key] = value;
        });

        let apiResponse: ApiResponse<T> = {
          data,
          status: response.status,
          headers: responseHeaders,
          requestId,
        };

        // Apply response interceptors
        for (const interceptor of this.responseInterceptors) {
          apiResponse = (await interceptor(apiResponse)) as ApiResponse<T>;
        }

        // Callback
        this.callbacks.onResponse?.(apiResponse);

        return apiResponse;
      } catch (error) {
        clearTimeout(timeoutId);

        // Handle abort (timeout)
        if (error instanceof Error && error.name === 'AbortError') {
          throw new TimeoutError();
        }

        // Handle network errors
        if (error instanceof TypeError && error.message.includes('fetch')) {
          throw new NetworkError('Network request failed', error);
        }

        // Apply error interceptors
        let finalError =
          error instanceof AuthGatewayError
            ? error
            : new AuthGatewayError(
                error instanceof Error ? error.message : 'Unknown error',
                {
                  originalError: error instanceof Error ? error : undefined,
                  retryable: false,
                }
              );

        for (const interceptor of this.errorInterceptors) {
          finalError = await interceptor(finalError);
        }

        // Callback
        this.callbacks.onError?.(finalError);

        throw finalError;
      }
    };

    // Execute with retry if configured
    const retryConfig = config.retryConfig
      ? { ...this.retryConfig, ...config.retryConfig }
      : this.retryConfig;

    if (retryConfig.maxRetries > 0) {
      return withRetry(executeRequest, retryConfig, (context, delay) => {
        this.log(
          `Retrying request (attempt ${context.attempt}/${context.maxRetries}) after ${delay}ms`
        );
      });
    }

    return executeRequest();
  }

  /** GET request */
  get<T>(
    url: string,
    options?: Omit<RequestConfig, 'method' | 'url' | 'body'>
  ): Promise<ApiResponse<T>> {
    return this.request<T>({
      method: 'GET',
      url,
      headers: {},
      ...options,
    });
  }

  /** POST request */
  post<T>(
    url: string,
    body?: unknown,
    options?: Omit<RequestConfig, 'method' | 'url' | 'body'>
  ): Promise<ApiResponse<T>> {
    return this.request<T>({
      method: 'POST',
      url,
      body,
      headers: {},
      ...options,
    });
  }

  /** PUT request */
  put<T>(
    url: string,
    body?: unknown,
    options?: Omit<RequestConfig, 'method' | 'url' | 'body'>
  ): Promise<ApiResponse<T>> {
    return this.request<T>({
      method: 'PUT',
      url,
      body,
      headers: {},
      ...options,
    });
  }

  /** PATCH request */
  patch<T>(
    url: string,
    body?: unknown,
    options?: Omit<RequestConfig, 'method' | 'url' | 'body'>
  ): Promise<ApiResponse<T>> {
    return this.request<T>({
      method: 'PATCH',
      url,
      body,
      headers: {},
      ...options,
    });
  }

  /** DELETE request */
  delete<T>(
    url: string,
    options?: Omit<RequestConfig, 'method' | 'url' | 'body'>
  ): Promise<ApiResponse<T>> {
    return this.request<T>({
      method: 'DELETE',
      url,
      headers: {},
      ...options,
    });
  }
}
