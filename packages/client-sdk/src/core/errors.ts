/**
 * Custom error classes for Auth Gateway Client SDK
 */

import type { ApiError } from '../config/types';

/** Base API error class */
export class AuthGatewayError extends Error implements ApiError {
  public readonly status?: number;
  public readonly code?: string;
  public readonly details?: string;
  public readonly requestId?: string;
  public readonly retryable: boolean;
  public readonly originalError?: Error;

  constructor(
    message: string,
    options: {
      status?: number;
      code?: string;
      details?: string;
      requestId?: string;
      retryable?: boolean;
      originalError?: Error;
    } = {}
  ) {
    super(message);
    this.name = 'AuthGatewayError';
    this.status = options.status;
    this.code = options.code;
    this.details = options.details;
    this.requestId = options.requestId;
    this.retryable = options.retryable ?? false;
    this.originalError = options.originalError;

    // Maintains proper stack trace for where error was thrown
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, AuthGatewayError);
    }
  }

  toJSON(): Record<string, unknown> {
    return {
      name: this.name,
      message: this.message,
      status: this.status,
      code: this.code,
      details: this.details,
      requestId: this.requestId,
      retryable: this.retryable,
    };
  }
}

/** Network error (connection issues, timeouts) */
export class NetworkError extends AuthGatewayError {
  constructor(message: string, originalError?: Error) {
    super(message, {
      code: 'NETWORK_ERROR',
      retryable: true,
      originalError,
    });
    this.name = 'NetworkError';
  }
}

/** Timeout error */
export class TimeoutError extends AuthGatewayError {
  constructor(message: string = 'Request timed out') {
    super(message, {
      code: 'TIMEOUT',
      retryable: true,
    });
    this.name = 'TimeoutError';
  }
}

/** Authentication error (401) */
export class AuthenticationError extends AuthGatewayError {
  constructor(message: string = 'Authentication failed', details?: string) {
    super(message, {
      status: 401,
      code: 'UNAUTHORIZED',
      details,
      retryable: false,
    });
    this.name = 'AuthenticationError';
  }
}

/** Authorization error (403) */
export class AuthorizationError extends AuthGatewayError {
  constructor(message: string = 'Access denied', details?: string) {
    super(message, {
      status: 403,
      code: 'FORBIDDEN',
      details,
      retryable: false,
    });
    this.name = 'AuthorizationError';
  }
}

/** Not found error (404) */
export class NotFoundError extends AuthGatewayError {
  constructor(message: string = 'Resource not found', details?: string) {
    super(message, {
      status: 404,
      code: 'NOT_FOUND',
      details,
      retryable: false,
    });
    this.name = 'NotFoundError';
  }
}

/** Validation error (400) */
export class ValidationError extends AuthGatewayError {
  public readonly fields?: Record<string, string[]>;

  constructor(
    message: string = 'Validation failed',
    fields?: Record<string, string[]>
  ) {
    super(message, {
      status: 400,
      code: 'VALIDATION_ERROR',
      retryable: false,
    });
    this.name = 'ValidationError';
    this.fields = fields;
  }
}

/** Conflict error (409) */
export class ConflictError extends AuthGatewayError {
  constructor(message: string = 'Resource already exists', details?: string) {
    super(message, {
      status: 409,
      code: 'CONFLICT',
      details,
      retryable: false,
    });
    this.name = 'ConflictError';
  }
}

/** Rate limit error (429) */
export class RateLimitError extends AuthGatewayError {
  public readonly retryAfter: number;

  constructor(message: string = 'Rate limit exceeded', retryAfter: number = 60) {
    super(message, {
      status: 429,
      code: 'RATE_LIMITED',
      retryable: true,
    });
    this.name = 'RateLimitError';
    this.retryAfter = retryAfter;
  }
}

/** Server error (5xx) */
export class ServerError extends AuthGatewayError {
  constructor(
    message: string = 'Internal server error',
    status: number = 500,
    details?: string
  ) {
    super(message, {
      status,
      code: 'SERVER_ERROR',
      details,
      retryable: status >= 500 && status < 600,
    });
    this.name = 'ServerError';
  }
}

/** Two-factor authentication required error */
export class TwoFactorRequiredError extends AuthGatewayError {
  public readonly twoFactorToken: string;

  constructor(twoFactorToken: string) {
    super('Two-factor authentication required', {
      code: '2FA_REQUIRED',
      retryable: false,
    });
    this.name = 'TwoFactorRequiredError';
    this.twoFactorToken = twoFactorToken;
  }
}

/** Create appropriate error from HTTP response */
export function createErrorFromResponse(
  status: number,
  body: { error?: string; message?: string; details?: string } | null,
  requestId?: string
): AuthGatewayError {
  const message = body?.message || body?.error || 'Unknown error';
  const details = body?.details;

  switch (status) {
    case 400:
      return new ValidationError(message);
    case 401:
      return new AuthenticationError(message, details);
    case 403:
      return new AuthorizationError(message, details);
    case 404:
      return new NotFoundError(message, details);
    case 409:
      return new ConflictError(message, details);
    case 429:
      return new RateLimitError(message);
    default:
      if (status >= 500) {
        return new ServerError(message, status, details);
      }
      return new AuthGatewayError(message, {
        status,
        details,
        requestId,
        retryable: false,
      });
  }
}

/** Check if error is retryable */
export function isRetryableError(error: unknown): boolean {
  if (error instanceof AuthGatewayError) {
    return error.retryable;
  }
  // Network errors are typically retryable
  if (error instanceof Error) {
    const message = error.message.toLowerCase();
    return (
      message.includes('network') ||
      message.includes('timeout') ||
      message.includes('econnreset') ||
      message.includes('econnrefused')
    );
  }
  return false;
}
