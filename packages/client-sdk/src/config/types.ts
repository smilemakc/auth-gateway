/**
 * Configuration types for Auth Gateway Client SDK
 */

/** Retry policy configuration */
export interface RetryConfig {
  /** Maximum number of retry attempts (default: 3) */
  maxRetries: number;
  /** Initial delay in ms before first retry (default: 1000) */
  initialDelayMs: number;
  /** Maximum delay in ms between retries (default: 30000) */
  maxDelayMs: number;
  /** Multiplier for exponential backoff (default: 2) */
  backoffMultiplier: number;
  /** HTTP status codes that should trigger a retry */
  retryableStatusCodes: number[];
  /** Whether to retry on network errors (default: true) */
  retryOnNetworkError: boolean;
}

/** Token storage interface for custom implementations */
export interface TokenStorage {
  getAccessToken(): string | null | Promise<string | null>;
  setAccessToken(token: string): void | Promise<void>;
  getRefreshToken(): string | null | Promise<string | null>;
  setRefreshToken(token: string): void | Promise<void>;
  clear(): void | Promise<void>;
}

/** Request interceptor */
export type RequestInterceptor = (
  config: RequestConfig
) => RequestConfig | Promise<RequestConfig>;

/** Response interceptor */
export type ResponseInterceptor<T = unknown> = (
  response: ApiResponse<T>
) => ApiResponse<T> | Promise<ApiResponse<T>>;

/** Error interceptor */
export type ErrorInterceptor = (
  error: ApiError
) => ApiError | Promise<ApiError> | never;

/** Lifecycle callbacks */
export interface ClientCallbacks {
  /** Called before each request */
  onRequest?: (config: RequestConfig) => void;
  /** Called after each successful response */
  onResponse?: <T>(response: ApiResponse<T>) => void;
  /** Called on any error */
  onError?: (error: ApiError) => void;
  /** Called when tokens are refreshed */
  onTokenRefresh?: (tokens: { accessToken: string; refreshToken: string }) => void;
  /** Called when authentication fails (401) after token refresh attempt */
  onAuthFailure?: () => void;
  /** Called when rate limited (429) */
  onRateLimited?: (retryAfter: number) => void;
}

/** Main client configuration */
export interface ClientConfig {
  /** Base URL for REST API (e.g., 'https://api.example.com') */
  baseUrl: string;
  /** gRPC server address (e.g., 'localhost:50051') */
  grpcAddress?: string;
  /** WebSocket URL (e.g., 'wss://api.example.com/ws') */
  wsUrl?: string;
  /** Default headers to include in all requests */
  headers?: Record<string, string>;
  /** Request timeout in milliseconds (default: 30000) */
  timeout?: number;
  /** Retry configuration */
  retry?: Partial<RetryConfig>;
  /** Token storage implementation */
  tokenStorage?: TokenStorage;
  /** Lifecycle callbacks */
  callbacks?: ClientCallbacks;
  /** Whether to automatically refresh tokens (default: true) */
  autoRefreshTokens?: boolean;
  /** API key for server-to-server communication */
  apiKey?: string;
  /** Application ID for multi-tenant environments */
  applicationId?: string;
  /** Enable debug logging */
  debug?: boolean;
}

/** Request configuration */
export interface RequestConfig {
  method: HttpMethod;
  url: string;
  headers?: Record<string, string>;
  body?: unknown;
  query?: Record<string, string | number | boolean | undefined>;
  timeout?: number;
  skipAuth?: boolean;
  retryConfig?: Partial<RetryConfig>;
}

/** API response wrapper */
export interface ApiResponse<T> {
  data: T;
  status: number;
  headers: Record<string, string>;
  requestId?: string;
}

/** API error */
export interface ApiError extends Error {
  status?: number;
  code?: string;
  details?: string;
  requestId?: string;
  retryable: boolean;
  originalError?: Error;
}

/** HTTP methods */
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

/** Pagination parameters */
export interface PaginationParams {
  page?: number;
  pageSize?: number;
  perPage?: number;
}

/** Paginated response */
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

/** Default configuration values */
export const DEFAULT_CONFIG: Required<
  Pick<ClientConfig, 'timeout' | 'autoRefreshTokens' | 'debug'>
> & { retry: RetryConfig } = {
  timeout: 30000,
  autoRefreshTokens: true,
  debug: false,
  retry: {
    maxRetries: 3,
    initialDelayMs: 1000,
    maxDelayMs: 30000,
    backoffMultiplier: 2,
    retryableStatusCodes: [408, 429, 500, 502, 503, 504],
    retryOnNetworkError: true,
  },
};
