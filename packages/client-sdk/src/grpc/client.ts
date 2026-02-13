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
  GrpcCallOptions,
  GrpcClientConfig,
  IntrospectTokenRequest,
  IntrospectTokenResponse,
  SyncUsersOptions,
  TokenExchangeResult,
  TokenExchangeRedeemResult,
  ValidateTokenRequest,
  ValidateTokenResponse,
  GetUserRequest,
  GetUserResponse,
  CreateUserRequest,
  CreateUserResponse,
  LoginRequest,
  LoginResponse,
  InitPasswordlessRegistrationRequest,
  InitPasswordlessRegistrationResponse,
  CompletePasswordlessRegistrationRequest,
  CompletePasswordlessRegistrationResponse,
  SendOTPRequest,
  SendOTPResponse,
  VerifyOTPRequest,
  VerifyOTPResponse,
  LoginWithOTPRequest,
  LoginWithOTPResponse,
  VerifyLoginOTPRequest,
  VerifyLoginOTPResponse,
  RegisterWithOTPRequest,
  RegisterWithOTPResponse,
  VerifyRegistrationOTPRequest,
  VerifyRegistrationOTPResponse,
  IntrospectOAuthTokenRequest,
  IntrospectOAuthTokenResponse,
  ValidateOAuthClientRequest,
  ValidateOAuthClientResponse,
  GetOAuthClientRequest,
  GetOAuthClientResponse,
  GetUserAppProfileRequest,
  UserAppProfileResponse,
  GetUserTelegramBotsRequest,
  UserTelegramBotsResponse,
  SendEmailRequest,
  SendEmailResponse,
  SyncUsersResponse,
  GetApplicationAuthConfigRequest,
  CreateTokenExchangeGrpcRequest,
  RedeemTokenExchangeGrpcRequest,
} from './types';

/** Default gRPC configuration */
const DEFAULT_CONFIG: Required<Omit<GrpcClientConfig, 'address' | 'caCertPath' | 'apiKey'>> = {
  useTls: false,
  timeout: 5000,
  debug: false,
};

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

    // Load proto definition from file (copied by generate-proto.sh)
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

    const protoDescriptor = grpc.loadPackageDefinition(packageDefinition);

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

    // Store all 24 RPC methods
    this.serviceMethods = {
      validateToken: this.promisify('validateToken'),
      getUser: this.promisify('getUser'),
      checkPermission: this.promisify('checkPermission'),
      introspectToken: this.promisify('introspectToken'),
      createUser: this.promisify('createUser'),
      login: this.promisify('login'),
      initPasswordlessRegistration: this.promisify('initPasswordlessRegistration'),
      completePasswordlessRegistration: this.promisify('completePasswordlessRegistration'),
      sendOTP: this.promisify('sendOTP'),
      verifyOTP: this.promisify('verifyOTP'),
      loginWithOTP: this.promisify('loginWithOTP'),
      verifyLoginOTP: this.promisify('verifyLoginOTP'),
      registerWithOTP: this.promisify('registerWithOTP'),
      verifyRegistrationOTP: this.promisify('verifyRegistrationOTP'),
      introspectOAuthToken: this.promisify('introspectOAuthToken'),
      validateOAuthClient: this.promisify('validateOAuthClient'),
      getOAuthClient: this.promisify('getOAuthClient'),
      sendEmail: this.promisify('sendEmail'),
      getUserApplicationProfile: this.promisify('getUserApplicationProfile'),
      getUserTelegramBots: this.promisify('getUserTelegramBots'),
      syncUsers: this.promisify('syncUsers'),
      getApplicationAuthConfig: this.promisify('getApplicationAuthConfig'),
      createTokenExchange: this.promisify('createTokenExchange'),
      redeemTokenExchange: this.promisify('redeemTokenExchange'),
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
        // Inject API key or application secret if configured
        if (this.config.apiKey) {
          metadata.add('x-api-key', this.config.apiKey);
        }
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

  /** Set or update the API key or application secret for authentication */
  setAPIKey(apiKey: string): void {
    this.config.apiKey = apiKey;
  }

  // ========== Ensure connection helper ==========

  private async ensureConnected(): Promise<void> {
    if (!this.isConnected()) {
      await this.connect();
    }
  }

  private ensureMethod(name: string): Function {
    const method = this.serviceMethods[name];
    if (!method) {
      throw new Error(`${name} method not available`);
    }
    return method;
  }

  // ========== Token & Auth Methods ==========

  /**
   * Validate an access token, API key, or application secret
   * @param accessToken JWT token, API key (starting with 'agw_'), or application secret (starting with 'app_')
   * @param options Call options
   * @returns Token validation result
   */
  async validateToken(
    accessToken: string,
    options?: GrpcCallOptions
  ): Promise<ValidateTokenResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('validateToken');

    const request: ValidateTokenRequest = { accessToken, applicationId: '' };
    this.log('ValidateToken:', { accessToken: accessToken.substring(0, 20) + '...' });

    return await method(request, options) as ValidateTokenResponse;
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
    await this.ensureConnected();
    const method = this.ensureMethod('getUser');

    const request: GetUserRequest = { userId, applicationId: '' };
    this.log('GetUser:', { userId });

    return await method(request, options) as GetUserResponse;
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
    await this.ensureConnected();
    const method = this.ensureMethod('checkPermission');

    const request: CheckPermissionRequest = { userId, resource, action, applicationId: '' };
    this.log('CheckPermission:', { userId, resource, action });

    return await method(request, options) as CheckPermissionResponse;
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
    await this.ensureConnected();
    const method = this.ensureMethod('introspectToken');

    const request: IntrospectTokenRequest = { accessToken };
    this.log('IntrospectToken:', { accessToken: accessToken.substring(0, 20) + '...' });

    return await method(request, options) as IntrospectTokenResponse;
  }

  // ========== User Management Methods ==========

  /**
   * Create a new user account
   * @param request User creation data
   * @param options Call options
   * @returns Created user with tokens
   */
  async createUser(
    request: CreateUserRequest,
    options?: GrpcCallOptions
  ): Promise<CreateUserResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('createUser');

    this.log('CreateUser:', { email: request.email });

    return await method(request, options) as CreateUserResponse;
  }

  /**
   * Authenticate a user with email/phone and password
   * @param request Login credentials
   * @param options Call options
   * @returns Login result with tokens
   */
  async login(
    request: LoginRequest,
    options?: GrpcCallOptions
  ): Promise<LoginResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('login');

    this.log('Login:', { email: request.email });

    return await method(request, options) as LoginResponse;
  }

  // ========== Passwordless Registration Methods ==========

  /**
   * Initiate passwordless two-step registration
   * @param request Registration initiation data
   * @param options Call options
   * @returns Initiation result
   */
  async initPasswordlessRegistration(
    request: InitPasswordlessRegistrationRequest,
    options?: GrpcCallOptions
  ): Promise<InitPasswordlessRegistrationResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('initPasswordlessRegistration');

    this.log('InitPasswordlessRegistration:', { email: request.email });

    return await method(request, options) as InitPasswordlessRegistrationResponse;
  }

  /**
   * Complete registration after OTP verification
   * @param request Registration completion data
   * @param options Call options
   * @returns Completed registration with tokens
   */
  async completePasswordlessRegistration(
    request: CompletePasswordlessRegistrationRequest,
    options?: GrpcCallOptions
  ): Promise<CompletePasswordlessRegistrationResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('completePasswordlessRegistration');

    this.log('CompletePasswordlessRegistration:', { email: request.email });

    return await method(request, options) as CompletePasswordlessRegistrationResponse;
  }

  // ========== OTP Methods ==========

  /**
   * Send a one-time password to email
   * @param request OTP send data
   * @param options Call options
   * @returns Send result
   */
  async sendOTP(
    request: SendOTPRequest,
    options?: GrpcCallOptions
  ): Promise<SendOTPResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('sendOTP');

    this.log('SendOTP:', { email: request.email, otpType: request.otpType });

    return await method(request, options) as SendOTPResponse;
  }

  /**
   * Verify a one-time password
   * @param request OTP verification data
   * @param options Call options
   * @returns Verification result
   */
  async verifyOTP(
    request: VerifyOTPRequest,
    options?: GrpcCallOptions
  ): Promise<VerifyOTPResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('verifyOTP');

    this.log('VerifyOTP:', { email: request.email });

    return await method(request, options) as VerifyOTPResponse;
  }

  /**
   * Initiate passwordless login by sending OTP to email
   * @param request Login OTP request data
   * @param options Call options
   * @returns Send result
   */
  async loginWithOTP(
    request: LoginWithOTPRequest,
    options?: GrpcCallOptions
  ): Promise<LoginWithOTPResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('loginWithOTP');

    this.log('LoginWithOTP:', { email: request.email });

    return await method(request, options) as LoginWithOTPResponse;
  }

  /**
   * Complete passwordless login by verifying OTP
   * @param request Login OTP verification data
   * @param options Call options
   * @returns Login result with tokens
   */
  async verifyLoginOTP(
    request: VerifyLoginOTPRequest,
    options?: GrpcCallOptions
  ): Promise<VerifyLoginOTPResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('verifyLoginOTP');

    this.log('VerifyLoginOTP:', { email: request.email });

    return await method(request, options) as VerifyLoginOTPResponse;
  }

  /**
   * Initiate OTP-based registration by sending verification code
   * @param request Registration OTP request data
   * @param options Call options
   * @returns Send result
   */
  async registerWithOTP(
    request: RegisterWithOTPRequest,
    options?: GrpcCallOptions
  ): Promise<RegisterWithOTPResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('registerWithOTP');

    this.log('RegisterWithOTP:', { email: request.email });

    return await method(request, options) as RegisterWithOTPResponse;
  }

  /**
   * Complete OTP-based registration
   * @param request Registration OTP verification data
   * @param options Call options
   * @returns Registration result with tokens
   */
  async verifyRegistrationOTP(
    request: VerifyRegistrationOTPRequest,
    options?: GrpcCallOptions
  ): Promise<VerifyRegistrationOTPResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('verifyRegistrationOTP');

    this.log('VerifyRegistrationOTP:', { email: request.email });

    return await method(request, options) as VerifyRegistrationOTPResponse;
  }

  // ========== OAuth Provider Methods ==========

  /**
   * Validate OAuth access token (RFC 7662)
   * @param request OAuth token introspection data
   * @param options Call options
   * @returns Introspection result
   */
  async introspectOAuthToken(
    request: IntrospectOAuthTokenRequest,
    options?: GrpcCallOptions
  ): Promise<IntrospectOAuthTokenResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('introspectOAuthToken');

    this.log('IntrospectOAuthToken');

    return await method(request, options) as IntrospectOAuthTokenResponse;
  }

  /**
   * Validate OAuth client credentials
   * @param request Client credentials
   * @param options Call options
   * @returns Validation result
   */
  async validateOAuthClient(
    request: ValidateOAuthClientRequest,
    options?: GrpcCallOptions
  ): Promise<ValidateOAuthClientResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('validateOAuthClient');

    this.log('ValidateOAuthClient:', { clientId: request.clientId });

    return await method(request, options) as ValidateOAuthClientResponse;
  }

  /**
   * Get OAuth client information by client_id
   * @param request Client ID data
   * @param options Call options
   * @returns OAuth client information
   */
  async getOAuthClient(
    request: GetOAuthClientRequest,
    options?: GrpcCallOptions
  ): Promise<GetOAuthClientResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('getOAuthClient');

    this.log('GetOAuthClient:', { clientId: request.clientId });

    return await method(request, options) as GetOAuthClientResponse;
  }

  // ========== Email Methods ==========

  /**
   * Send an email using a specified template
   * @param request Email send data
   * @param options Call options
   * @returns Send result
   */
  async sendEmail(
    request: SendEmailRequest,
    options?: GrpcCallOptions
  ): Promise<SendEmailResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('sendEmail');

    this.log('SendEmail:', { templateType: request.templateType, toEmail: request.toEmail });

    return await method(request, options) as SendEmailResponse;
  }

  // ========== Multi-Application Methods ==========

  /**
   * Get user's profile for a specific application
   * @param request User and application IDs
   * @param options Call options
   * @returns User application profile
   */
  async getUserApplicationProfile(
    request: GetUserAppProfileRequest,
    options?: GrpcCallOptions
  ): Promise<UserAppProfileResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('getUserApplicationProfile');

    this.log('GetUserApplicationProfile:', { userId: request.userId, applicationId: request.applicationId });

    return await method(request, options) as UserAppProfileResponse;
  }

  /**
   * Get user's Telegram bot access for an application
   * @param request User and application IDs
   * @param options Call options
   * @returns Telegram bot access records
   */
  async getUserTelegramBots(
    request: GetUserTelegramBotsRequest,
    options?: GrpcCallOptions
  ): Promise<UserTelegramBotsResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('getUserTelegramBots');

    this.log('GetUserTelegramBots:', { userId: request.userId });

    return await method(request, options) as UserTelegramBotsResponse;
  }

  // ========== Sync & Config Methods ==========

  /**
   * Sync users updated after a timestamp
   * @param opts Sync options including updatedAfter timestamp
   * @param callOptions gRPC call options
   * @returns List of users updated after the timestamp
   */
  async syncUsers(
    opts: SyncUsersOptions,
    callOptions?: GrpcCallOptions
  ): Promise<SyncUsersResponse> {
    await this.ensureConnected();
    const method = this.ensureMethod('syncUsers');

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

    return await method(request, callOptions) as SyncUsersResponse;
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
    await this.ensureConnected();
    const method = this.ensureMethod('getApplicationAuthConfig');

    const request: GetApplicationAuthConfigRequest = { applicationId };
    this.log('GetApplicationAuthConfig:', request);

    return await method(request, options) as AuthConfigResult;
  }

  // ========== SSO Token Exchange Methods ==========

  /**
   * Create a token exchange code for cross-application SSO
   * @param accessToken Current access token
   * @param targetAppId Target application UUID
   * @param options Call options
   * @returns Exchange code and expiration
   */
  async createTokenExchange(
    accessToken: string,
    targetAppId: string,
    options?: GrpcCallOptions
  ): Promise<TokenExchangeResult> {
    await this.ensureConnected();
    const method = this.ensureMethod('createTokenExchange');

    const request: CreateTokenExchangeGrpcRequest = { accessToken, targetApplicationId: targetAppId };
    this.log('CreateTokenExchange:', { targetApplicationId: targetAppId });

    return await method(request, options) as TokenExchangeResult;
  }

  /**
   * Redeem a token exchange code to get tokens for the target application
   * @param code Exchange code
   * @param options Call options
   * @returns Access token, refresh token, and user information
   */
  async redeemTokenExchange(
    code: string,
    options?: GrpcCallOptions
  ): Promise<TokenExchangeRedeemResult> {
    await this.ensureConnected();
    const method = this.ensureMethod('redeemTokenExchange');

    const request: RedeemTokenExchangeGrpcRequest = { exchangeCode: code };
    this.log('RedeemTokenExchange:', { exchangeCode: code.substring(0, 10) + '...' });

    return await method(request, options) as TokenExchangeRedeemResult;
  }
}

/** Create a gRPC client */
export function createGrpcClient(config: GrpcClientConfig): AuthGrpcClient {
  return new AuthGrpcClient(config);
}
