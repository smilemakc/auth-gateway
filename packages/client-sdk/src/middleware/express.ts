/**
 * Express middleware for Auth Gateway token validation
 */
import type { AuthMiddlewareOptions, TokenExtractor, AuthValidator, AuthData } from './types';

/** Extract token from Authorization: Bearer header */
export function bearerTokenExtractor(): TokenExtractor {
  return (req: any) => {
    const auth = req.headers?.authorization;
    if (auth && auth.startsWith('Bearer ')) {
      return auth.slice(7);
    }
    return null;
  };
}

/** Extract token from cookie */
export function cookieTokenExtractor(name: string): TokenExtractor {
  return (req: any) => req.cookies?.[name] || null;
}

/** Extract token from query parameter */
export function queryTokenExtractor(param: string): TokenExtractor {
  return (req: any) => req.query?.[param] || null;
}

/** Extract token from custom header */
export function headerTokenExtractor(header: string): TokenExtractor {
  return (req: any) => req.headers?.[header.toLowerCase()] || null;
}

/** Create Express authentication middleware */
export function createAuthMiddleware(
  validator: AuthValidator,
  options: AuthMiddlewareOptions = {}
) {
  const extractors = options.tokenExtractors || [bearerTokenExtractor()];
  const skipPaths = new Set(options.skipPaths || []);
  const cacheTTL = options.cacheTTL || 30000;
  const cache = options.cache ? new Map<string, { data: AuthData; expires: number }>() : null;

  return async (req: any, res: any, next: any) => {
    if (skipPaths.has(req.path)) {
      return next();
    }

    let token: string | null = null;
    for (const extractor of extractors) {
      token = extractor(req);
      if (token) break;
    }

    if (!token) {
      if (options.onError) {
        return options.onError(req, res, new Error('Missing authorization token'));
      }
      return res.status(401).json({ error: 'Missing authorization token' });
    }

    if (cache) {
      const cached = cache.get(token);
      if (cached && cached.expires > Date.now()) {
        req.auth = cached.data;
        return next();
      }
    }

    try {
      const validation = await validator.validateToken(token);

      if (!validation.valid) {
        if (options.onError) {
          return options.onError(req, res, new Error('Invalid token'));
        }
        return res.status(401).json({ error: 'Invalid or expired token' });
      }

      const authData: AuthData = {
        userId: validation.userId || '',
        email: validation.email || '',
        username: validation.username || '',
        roles: validation.roles || [],
        appRoles: validation.appRoles || [],
        applicationId: validation.applicationId,
      };

      if (cache) {
        cache.set(token, { data: authData, expires: Date.now() + cacheTTL });
      }

      req.auth = authData;
      next();
    } catch (err) {
      if (options.onError) {
        return options.onError(req, res, err instanceof Error ? err : new Error(String(err)));
      }
      return res.status(502).json({ error: 'Auth service unavailable' });
    }
  };
}

/** Require user to have at least one of the specified roles */
export function requireRole(...roles: string[]) {
  return (req: any, res: any, next: any) => {
    if (!req.auth) {
      return res.status(401).json({ error: 'Unauthorized' });
    }
    const hasRole = roles.some((r: string) => req.auth.roles.includes(r));
    if (!hasRole) {
      return res.status(403).json({ error: 'Insufficient permissions' });
    }
    next();
  };
}

/** Require user to have at least one of the specified app roles */
export function requireAppRole(...roles: string[]) {
  return (req: any, res: any, next: any) => {
    if (!req.auth) {
      return res.status(401).json({ error: 'Unauthorized' });
    }
    const hasRole = roles.some((r: string) => req.auth.appRoles.includes(r));
    if (!hasRole) {
      return res.status(403).json({ error: 'Insufficient permissions' });
    }
    next();
  };
}
