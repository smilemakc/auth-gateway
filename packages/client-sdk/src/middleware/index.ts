/**
 * Express middleware for Auth Gateway
 */
export {
  createAuthMiddleware,
  requireRole,
  requireAppRole,
  bearerTokenExtractor,
  cookieTokenExtractor,
  queryTokenExtractor,
  headerTokenExtractor,
} from './express';

export {
  RestClientAdapter,
  GrpcClientAdapter,
  createRestValidator,
  createGrpcValidator,
} from './adapters';

export { createSSOCallbackHandler } from './sso-callback';

export type {
  AuthData,
  TokenExtractor,
  AuthMiddlewareOptions,
  AuthValidator,
  TokenValidationResult,
} from './types';
