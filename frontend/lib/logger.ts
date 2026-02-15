/**
 * Development-only logger. All methods are no-ops in production builds.
 * Vite's tree-shaking removes the console calls entirely from the bundle
 * when import.meta.env.DEV is false.
 *
 * Usage: import { logger } from '@/lib/logger';
 *        logger.error('Failed to load:', err);
 */
export const logger = {
  error: (...args: unknown[]): void => {
    if (import.meta.env.DEV) console.error(...args);
  },
  warn: (...args: unknown[]): void => {
    if (import.meta.env.DEV) console.warn(...args);
  },
  log: (...args: unknown[]): void => {
    if (import.meta.env.DEV) console.log(...args);
  },
};
