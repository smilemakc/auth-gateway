/**
 * Adapters to make REST and gRPC clients compatible with middleware
 */
import type { AuthValidator, TokenValidationResult } from './types';

/** REST client validator adapter */
export class RestClientAdapter implements AuthValidator {
  constructor(private client: { auth: { validateToken: (token: string) => Promise<any> } }) {}

  async validateToken(token: string): Promise<TokenValidationResult> {
    const result = await this.client.auth.validateToken(token);
    return {
      valid: result.valid,
      userId: result.user_id,
      email: result.email,
      username: result.username,
      roles: result.roles,
      appRoles: [],
      applicationId: undefined,
    };
  }
}

/** gRPC client validator adapter */
export class GrpcClientAdapter implements AuthValidator {
  constructor(private client: { validateToken: (token: string) => Promise<any> }) {}

  async validateToken(token: string): Promise<TokenValidationResult> {
    const result = await this.client.validateToken(token);
    return {
      valid: result.valid,
      userId: result.userId,
      email: result.email,
      username: result.username,
      roles: result.roles || [],
      appRoles: [],
      applicationId: undefined,
    };
  }
}

/** Create validator from REST client */
export function createRestValidator(client: { auth: { validateToken: (token: string) => Promise<any> } }): AuthValidator {
  return new RestClientAdapter(client);
}

/** Create validator from gRPC client */
export function createGrpcValidator(client: { validateToken: (token: string) => Promise<any> }): AuthValidator {
  return new GrpcClientAdapter(client);
}
