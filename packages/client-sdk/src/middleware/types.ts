/**
 * Express middleware types
 */

/** Auth data attached to Express request */
export interface AuthData {
  userId: string;
  email: string;
  username: string;
  roles: string[];
  appRoles: string[];
  applicationId?: string;
}

/** Token extractor function */
export type TokenExtractor = (req: any) => string | null;

/** Middleware options */
export interface AuthMiddlewareOptions {
  tokenExtractors?: TokenExtractor[];
  skipPaths?: string[];
  onError?: (req: any, res: any, error: Error) => void;
  cache?: boolean;
  cacheTTL?: number;
}

/** Client interface for middleware (supports both REST and gRPC clients) */
export interface AuthValidator {
  validateToken(token: string): Promise<TokenValidationResult>;
}

export interface TokenValidationResult {
  valid: boolean;
  userId?: string;
  email?: string;
  username?: string;
  roles?: string[];
  appRoles?: string[];
  applicationId?: string;
}

declare global {
  namespace Express {
    interface Request {
      auth?: AuthData;
    }
  }
}
