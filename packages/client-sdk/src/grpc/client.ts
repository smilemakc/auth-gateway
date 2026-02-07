/**
 * gRPC client for Auth Gateway
 */

import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import * as path from 'path';
import type {
  AuthConfigResult,
  CheckPermissionRequest,
  CheckPermissionResponse,
  GetUserRequest,
  GetUserResponse,
  GrpcCallOptions,
  GrpcClientConfig,
  IntrospectTokenRequest,
  IntrospectTokenResponse,
  SyncUsersOptions,
  SyncUsersResult,
  ValidateTokenRequest,
  ValidateTokenResponse,
} from './types';

/** Default gRPC configuration */
const DEFAULT_CONFIG: Required<Omit<GrpcClientConfig, 'address' | 'caCertPath'>> = {
  useTls: false,
  timeout: 5000,
  debug: false,
};

/** Proto definition for dynamic loading */
const PROTO_DEFINITION = `
syntax = "proto3";

package auth;

service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionResponse);
  rpc IntrospectToken(IntrospectTokenRequest) returns (IntrospectTokenResponse);
}

message ValidateTokenRequest {
  string access_token = 1;
}

message ValidateTokenResponse {
  bool valid = 1;
  string user_id = 2;
  string email = 3;
  string username = 4;
  string role = 5;
  string error_message = 6;
  int64 expires_at = 7;
}

message GetUserRequest {
  string user_id = 1;
}

message User {
  string id = 1;
  string email = 2;
  string username = 3;
  string full_name = 4;
  string profile_picture_url = 5;
  string role = 6;
  bool email_verified = 7;
  bool is_active = 8;
  int64 created_at = 9;
  int64 updated_at = 10;
}

message GetUserResponse {
  User user = 1;
  string error_message = 2;
}

message CheckPermissionRequest {
  string user_id = 1;
  string resource = 2;
  string action = 3;
}

message CheckPermissionResponse {
  bool allowed = 1;
  string role = 2;
  string error_message = 3;
}

message IntrospectTokenRequest {
  string access_token = 1;
}

message IntrospectTokenResponse {
  bool active = 1;
  string user_id = 2;
  string email = 3;
  string username = 4;
  string role = 5;
  int64 issued_at = 6;
  int64 expires_at = 7;
  int64 not_before = 8;
  string subject = 9;
  bool blacklisted = 10;
  string error_message = 11;
}
`;

/** gRPC Auth Service client */
export class AuthGrpcClient {
  private client: grpc.Client | null = null;
  private config: GrpcClientConfig & typeof DEFAULT_CONFIG;
  private serviceMethods: Record<string, Function> = {};

  constructor(config: GrpcClientConfig) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  /** Log debug message */
  private log(...args: unknown[]): void {
    if (this.config.debug) {
      console.log('[AuthGrpcClient]', ...args);
    }
  }

  /** Connect to gRPC server */
  async connect(): Promise<void> {
    if (this.client) {
      this.log('Already connected');
      return;
    }

    this.log(`Connecting to ${this.config.address}`);

    // Load proto definition from string
    const packageDefinition = protoLoader.loadSync(
      path.join(__dirname, 'auth.proto'),
      {
        keepCase: false,
        longs: Number,
        enums: String,
        defaults: true,
        oneofs: true,
      }
    );

    // If proto file doesn't exist, use inline definition
    let protoDescriptor: grpc.GrpcObject;
    try {
      protoDescriptor = grpc.loadPackageDefinition(packageDefinition);
    } catch {
      // Fallback: create client directly with proto definition string
      this.log('Using inline proto definition');
      const tempProtoPath = '/tmp/auth.proto';
      const fs = await import('fs');
      fs.writeFileSync(tempProtoPath, PROTO_DEFINITION);

      const tempPackageDef = protoLoader.loadSync(tempProtoPath, {
        keepCase: false,
        longs: Number,
        enums: String,
        defaults: true,
        oneofs: true,
      });
      protoDescriptor = grpc.loadPackageDefinition(tempPackageDef);
    }

    // Get credentials
    const credentials = this.config.useTls
      ? this.config.caCertPath
        ? grpc.credentials.createSsl(
            (await import('fs')).readFileSync(this.config.caCertPath)
          )
        : grpc.credentials.createSsl()
      : grpc.credentials.createInsecure();

    // Create client
    const AuthService = (protoDescriptor.auth as grpc.GrpcObject)
      .AuthService as grpc.ServiceClientConstructor;
    this.client = new AuthService(this.config.address, credentials);

    // Store methods
    this.serviceMethods = {
      validateToken: this.promisify('validateToken'),
      getUser: this.promisify('getUser'),
      checkPermission: this.promisify('checkPermission'),
      introspectToken: this.promisify('introspectToken'),
      syncUsers: this.promisify('syncUsers'),
      getApplicationAuthConfig: this.promisify('getApplicationAuthConfig'),
    };

    // Wait for connection
    await this.waitForReady();
  }

  /** Wait for gRPC channel to be ready */
  private waitForReady(): Promise<void> {
    return new Promise((resolve, reject) => {
      const deadline = Date.now() + this.config.timeout;

      (this.client as grpc.Client).waitForReady(deadline, (error?: Error) => {
        if (error) {
          reject(new Error(`Failed to connect to gRPC server: ${error.message}`));
        } else {
          this.log('Connected successfully');
          resolve();
        }
      });
    });
  }

  /** Promisify a gRPC method */
  private promisify(methodName: string): Function {
    return (request: unknown, options?: GrpcCallOptions): Promise<unknown> => {
      return new Promise((resolve, reject) => {
        if (!this.client) {
          reject(new Error('Client not connected'));
          return;
        }

        const metadata = new grpc.Metadata();
        if (options?.metadata) {
          for (const [key, value] of Object.entries(options.metadata)) {
            metadata.add(key, value);
          }
        }

        const callOptions: grpc.CallOptions = {};
        if (options?.timeout) {
          callOptions.deadline = Date.now() + options.timeout;
        }

        (this.client as any)[methodName](
          request,
          metadata,
          callOptions,
          (error: grpc.ServiceError | null, response: unknown) => {
            if (error) {
              reject(new Error(`gRPC error: ${error.message} (code: ${error.code})`));
            } else {
              resolve(response);
            }
          }
        );
      });
    };
  }

  /** Disconnect from gRPC server */
  disconnect(): void {
    if (this.client) {
      this.log('Disconnecting');
      (this.client as grpc.Client).close();
      this.client = null;
      this.serviceMethods = {};
    }
  }

  /** Check if connected */
  isConnected(): boolean {
    return this.client !== null;
  }

  /**
   * Validate an access token or API key
   * @param accessToken JWT token or API key (starting with 'agw_')
   * @param options Call options
   * @returns Token validation result
   */
  async validateToken(
    accessToken: string,
    options?: GrpcCallOptions
  ): Promise<ValidateTokenResponse> {
    if (!this.isConnected()) {
      await this.connect();
    }

    if (!this.serviceMethods.validateToken) {
      throw new Error('ValidateToken method not available');
    }

    const request: ValidateTokenRequest = { accessToken };
    this.log('ValidateToken:', { accessToken: accessToken.substring(0, 20) + '...' });

    const response = await this.serviceMethods.validateToken(request, options);
    return this.transformResponse(response as Record<string, unknown>) as unknown as ValidateTokenResponse;
  }

  /**
   * Get user information by ID
   * @param userId User UUID
   * @param options Call options
   * @returns User information
   */
  async getUser(
    userId: string,
    options?: GrpcCallOptions
  ): Promise<GetUserResponse> {
    if (!this.isConnected()) {
      await this.connect();
    }

    if (!this.serviceMethods.getUser) {
      throw new Error('GetUser method not available');
    }

    const request: GetUserRequest = { userId };
    this.log('GetUser:', request);

    const response = await this.serviceMethods.getUser(request, options);
    return this.transformResponse(response as Record<string, unknown>) as unknown as GetUserResponse;
  }

  /**
   * Check if a user has permission to perform an action on a resource
   * @param userId User UUID
   * @param resource Resource name (e.g., 'users', 'products')
   * @param action Action name (e.g., 'read', 'write', 'delete')
   * @param options Call options
   * @returns Permission check result
   */
  async checkPermission(
    userId: string,
    resource: string,
    action: string,
    options?: GrpcCallOptions
  ): Promise<CheckPermissionResponse> {
    if (!this.isConnected()) {
      await this.connect();
    }

    if (!this.serviceMethods.checkPermission) {
      throw new Error('CheckPermission method not available');
    }

    const request: CheckPermissionRequest = { userId, resource, action };
    this.log('CheckPermission:', request);

    const response = await this.serviceMethods.checkPermission(request, options);
    return this.transformResponse(response as Record<string, unknown>) as unknown as CheckPermissionResponse;
  }

  /**
   * Introspect a token for detailed information
   * @param accessToken JWT token
   * @param options Call options
   * @returns Detailed token information
   */
  async introspectToken(
    accessToken: string,
    options?: GrpcCallOptions
  ): Promise<IntrospectTokenResponse> {
    if (!this.isConnected()) {
      await this.connect();
    }

    if (!this.serviceMethods.introspectToken) {
      throw new Error('IntrospectToken method not available');
    }

    const request: IntrospectTokenRequest = { accessToken };
    this.log('IntrospectToken:', { accessToken: accessToken.substring(0, 20) + '...' });

    const response = await this.serviceMethods.introspectToken(request, options);
    return this.transformResponse(response as Record<string, unknown>) as unknown as IntrospectTokenResponse;
  }

  /**
   * Sync users updated after a timestamp
   * @param opts Sync options including updatedAfter timestamp
   * @param callOptions gRPC call options
   * @returns List of users updated after the timestamp
   */
  async syncUsers(
    opts: SyncUsersOptions,
    callOptions?: GrpcCallOptions
  ): Promise<SyncUsersResult> {
    if (!this.isConnected()) {
      await this.connect();
    }

    if (!this.serviceMethods.syncUsers) {
      throw new Error('SyncUsers method not available');
    }

    const updatedAfter = typeof opts.updatedAfter === 'string'
      ? opts.updatedAfter
      : opts.updatedAfter.toISOString();

    const request = {
      updatedAfter,
      applicationId: opts.applicationId || '',
      limit: opts.limit || 100,
      offset: opts.offset || 0,
    };

    this.log('SyncUsers:', { updatedAfter, applicationId: opts.applicationId, limit: opts.limit, offset: opts.offset });

    const response = await this.serviceMethods.syncUsers(request, callOptions);
    return this.transformResponse(response as Record<string, unknown>) as unknown as SyncUsersResult;
  }

  /**
   * Get auth config for an application
   * @param applicationId Application UUID
   * @param options Call options
   * @returns Application auth configuration
   */
  async getApplicationAuthConfig(
    applicationId: string,
    options?: GrpcCallOptions
  ): Promise<AuthConfigResult> {
    if (!this.isConnected()) {
      await this.connect();
    }

    if (!this.serviceMethods.getApplicationAuthConfig) {
      throw new Error('GetApplicationAuthConfig method not available');
    }

    const request = { applicationId };
    this.log('GetApplicationAuthConfig:', request);

    const response = await this.serviceMethods.getApplicationAuthConfig(request, options);
    return this.transformResponse(response as Record<string, unknown>) as unknown as AuthConfigResult;
  }

  /** Transform snake_case response to camelCase */
  private transformResponse(obj: Record<string, unknown>): Record<string, unknown> {
    const result: Record<string, unknown> = {};

    for (const [key, value] of Object.entries(obj)) {
      const camelKey = key.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());

      if (Array.isArray(value)) {
        result[camelKey] = value.map(item =>
          item && typeof item === 'object'
            ? this.transformResponse(item as Record<string, unknown>)
            : item
        );
      } else if (value && typeof value === 'object') {
        result[camelKey] = this.transformResponse(value as Record<string, unknown>);
      } else {
        result[camelKey] = value;
      }
    }

    return result;
  }
}

/** Create a gRPC client */
export function createGrpcClient(config: GrpcClientConfig): AuthGrpcClient {
  return new AuthGrpcClient(config);
}
