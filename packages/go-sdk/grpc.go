package authgateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCClient provides a high-level interface to the Auth Gateway gRPC API.
// This is typically used for server-to-server communication.
type GRPCClient struct {
	conn   *grpc.ClientConn
	client proto.AuthServiceClient

	// Custom metadata to include in every request
	metadata   map[string]string
	metadataMu sync.RWMutex

	apiKey string
}

// GRPCConfig contains configuration for the gRPC client.
type GRPCConfig struct {
	// Address is the gRPC server address (e.g., "localhost:50051")
	Address string

	// Insecure disables TLS (for development)
	Insecure bool

	// DialTimeout is the timeout for establishing the connection
	DialTimeout time.Duration

	// DialOptions allows passing custom gRPC dial options
	DialOptions []grpc.DialOption

	// Metadata contains custom metadata to include in every gRPC call.
	// Common metadata: x-application-id, x-client-name, x-request-id, etc.
	Metadata map[string]string

	// APIKey for gRPC authentication. Sent as x-api-key metadata on every call.
	APIKey string
}

// NewGRPCClient creates a new gRPC client for the Auth Gateway.
func NewGRPCClient(config GRPCConfig) (*GRPCClient, error) {
	if config.Address == "" {
		config.Address = "localhost:50051"
	}

	if config.DialTimeout == 0 {
		config.DialTimeout = 10 * time.Second
	}

	opts := config.DialOptions
	if config.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(config.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	c := &GRPCClient{
		conn:     conn,
		client:   proto.NewAuthServiceClient(conn),
		metadata: config.Metadata,
		apiKey:   config.APIKey,
	}

	if config.APIKey != "" {
		c.SetMetadata("x-api-key", config.APIKey)
	}

	return c, nil
}

// SetMetadata sets a custom metadata key-value pair to be included in all requests.
func (c *GRPCClient) SetMetadata(key, value string) {
	c.metadataMu.Lock()
	defer c.metadataMu.Unlock()

	if c.metadata == nil {
		c.metadata = make(map[string]string)
	}
	c.metadata[key] = value
}

// SetMetadataMap sets multiple custom metadata key-value pairs.
func (c *GRPCClient) SetMetadataMap(md map[string]string) {
	c.metadataMu.Lock()
	defer c.metadataMu.Unlock()

	if c.metadata == nil {
		c.metadata = make(map[string]string)
	}
	for k, v := range md {
		c.metadata[k] = v
	}
}

// RemoveMetadata removes a custom metadata key.
func (c *GRPCClient) RemoveMetadata(key string) {
	c.metadataMu.Lock()
	defer c.metadataMu.Unlock()

	delete(c.metadata, key)
}

// GetMetadata returns a copy of the current custom metadata.
func (c *GRPCClient) GetMetadata() map[string]string {
	c.metadataMu.RLock()
	defer c.metadataMu.RUnlock()

	result := make(map[string]string)
	for k, v := range c.metadata {
		result[k] = v
	}
	return result
}

// SetApplicationID sets the x-application-id metadata for multi-tenant support.
func (c *GRPCClient) SetApplicationID(appID string) {
	c.SetMetadata("x-application-id", appID)
}

// SetClientName sets the x-client-name metadata for client identification.
func (c *GRPCClient) SetClientName(name string) {
	c.SetMetadata("x-client-name", name)
}

// SetAPIKey sets the API key for gRPC authentication.
// The key is sent as x-api-key metadata on every call.
func (c *GRPCClient) SetAPIKey(apiKey string) {
	c.SetMetadata("x-api-key", apiKey)
}

// withMetadata returns a context with the client's metadata attached.
// If the context already has metadata, the client metadata is appended.
func (c *GRPCClient) withMetadata(ctx context.Context) context.Context {
	c.metadataMu.RLock()
	defer c.metadataMu.RUnlock()

	if len(c.metadata) == 0 {
		return ctx
	}

	pairs := make([]string, 0, len(c.metadata)*2)
	for k, v := range c.metadata {
		pairs = append(pairs, k, v)
	}

	md := metadata.Pairs(pairs...)

	// Merge with existing metadata if present
	if existingMD, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existingMD, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

// grpcMetadataContextKey is used for per-request metadata via context.
type grpcMetadataContextKey struct{}

// WithGRPCMetadata returns a context with custom metadata for a single gRPC call.
// These metadata will be merged with the client's default metadata.
func WithGRPCMetadata(ctx context.Context, md map[string]string) context.Context {
	pairs := make([]string, 0, len(md)*2)
	for k, v := range md {
		pairs = append(pairs, k, v)
	}
	grpcMD := metadata.Pairs(pairs...)

	// Merge with existing metadata if present
	if existingMD, ok := metadata.FromOutgoingContext(ctx); ok {
		grpcMD = metadata.Join(existingMD, grpcMD)
	}

	return metadata.NewOutgoingContext(ctx, grpcMD)
}

// WithGRPCRequestID returns a context with a request ID for tracing.
func WithGRPCRequestID(ctx context.Context, requestID string) context.Context {
	return WithGRPCMetadata(ctx, map[string]string{"x-request-id": requestID})
}

// Close closes the gRPC connection.
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ValidateToken validates a JWT access token and returns user information.
func (c *GRPCClient) ValidateToken(ctx context.Context, accessToken string) (*proto.ValidateTokenResponse, error) {
	resp, err := c.client.ValidateToken(c.withMetadata(ctx), &proto.ValidateTokenRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeInvalidToken,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// GetUser retrieves user information by ID.
func (c *GRPCClient) GetUser(ctx context.Context, userID string) (*proto.User, error) {
	resp, err := c.client.GetUser(c.withMetadata(ctx), &proto.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeNotFound,
			Message: resp.ErrorMessage,
		}
	}

	return resp.User, nil
}

// CheckPermission checks if a user has a specific permission.
func (c *GRPCClient) CheckPermission(ctx context.Context, userID, resource, action string) (*proto.CheckPermissionResponse, error) {
	resp, err := c.client.CheckPermission(c.withMetadata(ctx), &proto.CheckPermissionRequest{
		UserId:   userID,
		Resource: resource,
		Action:   action,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeForbidden,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// HasPermission is a convenience method that returns true if the user has the permission.
func (c *GRPCClient) HasPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	resp, err := c.CheckPermission(ctx, userID, resource, action)
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

// IntrospectToken provides detailed information about a token.
func (c *GRPCClient) IntrospectToken(ctx context.Context, accessToken string) (*proto.IntrospectTokenResponse, error) {
	resp, err := c.client.IntrospectToken(c.withMetadata(ctx), &proto.IntrospectTokenRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to introspect token: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeInvalidToken,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// CreateUser creates a new user account via gRPC.
func (c *GRPCClient) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	resp, err := c.client.CreateUser(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// InitPasswordlessRegistration initiates a two-step passwordless registration via gRPC.
// Step 1: Provide email or phone and optional name/username.
// An OTP is sent to the provided email or phone.
func (c *GRPCClient) InitPasswordlessRegistration(ctx context.Context, req *proto.InitPasswordlessRegistrationRequest) (*proto.InitPasswordlessRegistrationResponse, error) {
	resp, err := c.client.InitPasswordlessRegistration(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to init passwordless registration: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// CompletePasswordlessRegistration completes the two-step passwordless registration via gRPC.
// Step 2: Provide the OTP code received via email or SMS.
// Returns tokens and user info on success.
func (c *GRPCClient) CompletePasswordlessRegistration(ctx context.Context, req *proto.CompletePasswordlessRegistrationRequest) (*proto.CompletePasswordlessRegistrationResponse, error) {
	resp, err := c.client.CompletePasswordlessRegistration(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to complete passwordless registration: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// Raw returns the underlying gRPC client for advanced usage.
func (c *GRPCClient) Raw() proto.AuthServiceClient {
	return c.client
}

// ========== Authentication Methods ==========

// Login authenticates a user with email/phone and password.
func (c *GRPCClient) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	resp, err := c.client.Login(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeUnauthorized,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// LoginWithEmail is a convenience method for login with email and password.
func (c *GRPCClient) LoginWithEmail(ctx context.Context, email, password string) (*proto.LoginResponse, error) {
	return c.Login(ctx, &proto.LoginRequest{
		Email:    email,
		Password: password,
	})
}

// LoginWithPhone is a convenience method for login with phone and password.
func (c *GRPCClient) LoginWithPhone(ctx context.Context, phone, password string) (*proto.LoginResponse, error) {
	return c.Login(ctx, &proto.LoginRequest{
		Phone:    phone,
		Password: password,
	})
}

// ========== OTP Methods ==========

// SendOTP sends a one-time password to email.
func (c *GRPCClient) SendOTP(ctx context.Context, req *proto.SendOTPRequest) (*proto.SendOTPResponse, error) {
	resp, err := c.client.SendOTP(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to send OTP: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// VerifyOTP verifies a one-time password.
func (c *GRPCClient) VerifyOTP(ctx context.Context, req *proto.VerifyOTPRequest) (*proto.VerifyOTPResponse, error) {
	resp, err := c.client.VerifyOTP(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// LoginWithOTP initiates passwordless login by sending OTP to email.
func (c *GRPCClient) LoginWithOTP(ctx context.Context, req *proto.LoginWithOTPRequest) (*proto.LoginWithOTPResponse, error) {
	resp, err := c.client.LoginWithOTP(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate OTP login: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// VerifyLoginOTP completes passwordless login by verifying OTP.
func (c *GRPCClient) VerifyLoginOTP(ctx context.Context, req *proto.VerifyLoginOTPRequest) (*proto.VerifyLoginOTPResponse, error) {
	resp, err := c.client.VerifyLoginOTP(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify login OTP: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeUnauthorized,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// RegisterWithOTP initiates OTP-based registration by sending verification code.
func (c *GRPCClient) RegisterWithOTP(ctx context.Context, req *proto.RegisterWithOTPRequest) (*proto.RegisterWithOTPResponse, error) {
	resp, err := c.client.RegisterWithOTP(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate OTP registration: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// VerifyRegistrationOTP completes OTP-based registration.
func (c *GRPCClient) VerifyRegistrationOTP(ctx context.Context, req *proto.VerifyRegistrationOTPRequest) (*proto.VerifyRegistrationOTPResponse, error) {
	resp, err := c.client.VerifyRegistrationOTP(c.withMetadata(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to complete OTP registration: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// ========== OAuth Provider Methods ==========

// IntrospectOAuthToken validates an OAuth access token (RFC 7662).
func (c *GRPCClient) IntrospectOAuthToken(ctx context.Context, token string, tokenTypeHint string) (*proto.IntrospectOAuthTokenResponse, error) {
	resp, err := c.client.IntrospectOAuthToken(c.withMetadata(ctx), &proto.IntrospectOAuthTokenRequest{
		Token:         token,
		TokenTypeHint: tokenTypeHint,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to introspect OAuth token: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeInvalidToken,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// ValidateOAuthClient validates OAuth client credentials.
func (c *GRPCClient) ValidateOAuthClient(ctx context.Context, clientID, clientSecret string) (*proto.ValidateOAuthClientResponse, error) {
	resp, err := c.client.ValidateOAuthClient(c.withMetadata(ctx), &proto.ValidateOAuthClientRequest{
		ClientId:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to validate OAuth client: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeUnauthorized,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// GetOAuthClient retrieves OAuth client information by client_id.
func (c *GRPCClient) GetOAuthClient(ctx context.Context, clientID string) (*proto.OAuthClient, error) {
	resp, err := c.client.GetOAuthClient(c.withMetadata(ctx), &proto.GetOAuthClientRequest{
		ClientId: clientID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth client: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeNotFound,
			Message: resp.ErrorMessage,
		}
	}

	return resp.Client, nil
}

// GetUserApplicationProfile returns user's profile for a specific application
func (c *GRPCClient) GetUserApplicationProfile(ctx context.Context, userID, applicationID string) (*proto.UserAppProfileResponse, error) {
	resp, err := c.client.GetUserApplicationProfile(c.withMetadata(ctx), &proto.GetUserAppProfileRequest{
		UserId:        userID,
		ApplicationId: applicationID,
	})
	if err != nil {
		return nil, fmt.Errorf("get user application profile: %w", err)
	}
	if resp.ErrorMessage != "" {
		return nil, &APIError{Code: "GRPC_ERROR", Message: resp.ErrorMessage}
	}
	return resp, nil
}

// GetUserTelegramBots returns user's Telegram bot access for an application
func (c *GRPCClient) GetUserTelegramBots(ctx context.Context, userID, applicationID string) (*proto.UserTelegramBotsResponse, error) {
	resp, err := c.client.GetUserTelegramBots(c.withMetadata(ctx), &proto.GetUserTelegramBotsRequest{
		UserId:        userID,
		ApplicationId: applicationID,
	})
	if err != nil {
		return nil, fmt.Errorf("get user telegram bots: %w", err)
	}
	if resp.ErrorMessage != "" {
		return nil, &APIError{Code: "GRPC_ERROR", Message: resp.ErrorMessage}
	}
	return resp, nil
}

// ========== Sync & Config Methods ==========

// SyncUsers returns users updated after a given timestamp for shadow table sync.
func (c *GRPCClient) SyncUsers(ctx context.Context, updatedAfter string, applicationID string, limit, offset int32) (*proto.SyncUsersResponse, error) {
	resp, err := c.client.SyncUsers(c.withMetadata(ctx), &proto.SyncUsersRequest{
		UpdatedAfter:  updatedAfter,
		ApplicationId: applicationID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sync users: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}

// GetApplicationAuthConfig returns auth configuration for a specific application.
func (c *GRPCClient) GetApplicationAuthConfig(ctx context.Context, applicationID string) (*proto.GetApplicationAuthConfigResponse, error) {
	resp, err := c.client.GetApplicationAuthConfig(c.withMetadata(ctx), &proto.GetApplicationAuthConfigRequest{
		ApplicationId: applicationID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get application auth config: %w", err)
	}

	if resp.ErrorMessage != "" {
		return nil, &APIError{
			Code:    ErrCodeBadRequest,
			Message: resp.ErrorMessage,
		}
	}

	return resp, nil
}
