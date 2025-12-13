/**
 * Retry policy implementation with exponential backoff
 */

import type { RetryConfig } from '../config/types';
import { DEFAULT_CONFIG } from '../config/types';
import { isRetryableError, RateLimitError } from './errors';

export interface RetryContext {
  attempt: number;
  maxRetries: number;
  lastError?: Error;
  startTime: number;
}

export type RetryCondition = (error: unknown, context: RetryContext) => boolean;

/** Calculate delay for next retry attempt with exponential backoff and jitter */
export function calculateRetryDelay(
  attempt: number,
  config: RetryConfig
): number {
  // Exponential backoff: initialDelay * (multiplier ^ attempt)
  const exponentialDelay =
    config.initialDelayMs * Math.pow(config.backoffMultiplier, attempt);

  // Cap at max delay
  const cappedDelay = Math.min(exponentialDelay, config.maxDelayMs);

  // Add jitter (±10%) to prevent thundering herd
  const jitter = cappedDelay * 0.1 * (Math.random() * 2 - 1);

  return Math.floor(cappedDelay + jitter);
}

/** Sleep for specified milliseconds */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/** Default retry condition */
export function defaultRetryCondition(
  error: unknown,
  context: RetryContext,
  config: RetryConfig
): boolean {
  // Don't retry if max attempts reached
  if (context.attempt >= config.maxRetries) {
    return false;
  }

  // Check if error is retryable
  if (!isRetryableError(error)) {
    return false;
  }

  // Check status code if available
  if (error instanceof Error && 'status' in error) {
    const status = (error as { status: number }).status;
    return config.retryableStatusCodes.includes(status);
  }

  // For network errors, check config
  return config.retryOnNetworkError;
}

/** Retry operation with exponential backoff */
export async function withRetry<T>(
  operation: () => Promise<T>,
  config: Partial<RetryConfig> = {},
  onRetry?: (context: RetryContext, delay: number) => void
): Promise<T> {
  const fullConfig: RetryConfig = {
    ...DEFAULT_CONFIG.retry,
    ...config,
  };

  const context: RetryContext = {
    attempt: 0,
    maxRetries: fullConfig.maxRetries,
    startTime: Date.now(),
  };

  while (true) {
    try {
      return await operation();
    } catch (error) {
      context.lastError = error instanceof Error ? error : new Error(String(error));

      // Handle rate limit with Retry-After header
      if (error instanceof RateLimitError) {
        const retryAfter = error.retryAfter * 1000; // Convert to ms
        if (context.attempt < fullConfig.maxRetries) {
          context.attempt++;
          onRetry?.(context, retryAfter);
          await sleep(retryAfter);
          continue;
        }
      }

      // Check if we should retry
      if (!defaultRetryCondition(error, context, fullConfig)) {
        throw error;
      }

      // Calculate delay and wait
      const delay = calculateRetryDelay(context.attempt, fullConfig);
      context.attempt++;

      onRetry?.(context, delay);
      await sleep(delay);
    }
  }
}

/** Create a retry wrapper with pre-configured options */
export function createRetryWrapper(config: Partial<RetryConfig> = {}) {
  const fullConfig: RetryConfig = {
    ...DEFAULT_CONFIG.retry,
    ...config,
  };

  return {
    config: fullConfig,

    /** Execute operation with retry */
    async execute<T>(
      operation: () => Promise<T>,
      overrideConfig?: Partial<RetryConfig>,
      onRetry?: (context: RetryContext, delay: number) => void
    ): Promise<T> {
      return withRetry(
        operation,
        { ...fullConfig, ...overrideConfig },
        onRetry
      );
    },

    /** Update retry configuration */
    updateConfig(newConfig: Partial<RetryConfig>): void {
      Object.assign(fullConfig, newConfig);
    },
  };
}

/** Retry decorator for class methods */
export function Retryable(config?: Partial<RetryConfig>) {
  return function <T>(
    _target: object,
    _propertyKey: string,
    descriptor: TypedPropertyDescriptor<(...args: unknown[]) => Promise<T>>
  ): TypedPropertyDescriptor<(...args: unknown[]) => Promise<T>> {
    const originalMethod = descriptor.value;

    if (!originalMethod) {
      return descriptor;
    }

    descriptor.value = async function (
      this: unknown,
      ...args: unknown[]
    ): Promise<T> {
      return withRetry(() => originalMethod.apply(this, args), config);
    };

    return descriptor;
  };
}
