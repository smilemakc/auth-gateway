package authgateway

import (
	"context"
	"fmt"
	"time"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClient provides a high-level interface to the Auth Gateway gRPC API.
// This is typically used for server-to-server communication.
type GRPCClient struct {
	conn   *grpc.ClientConn
	client proto.AuthServiceClient
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

	return &GRPCClient{
		conn:   conn,
		client: proto.NewAuthServiceClient(conn),
	}, nil
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
	resp, err := c.client.ValidateToken(ctx, &proto.ValidateTokenRequest{
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
	resp, err := c.client.GetUser(ctx, &proto.GetUserRequest{
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
	resp, err := c.client.CheckPermission(ctx, &proto.CheckPermissionRequest{
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
	resp, err := c.client.IntrospectToken(ctx, &proto.IntrospectTokenRequest{
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
	resp, err := c.client.CreateUser(ctx, req)
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
	resp, err := c.client.InitPasswordlessRegistration(ctx, req)
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
	resp, err := c.client.CompletePasswordlessRegistration(ctx, req)
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
	resp, err := c.client.Login(ctx, req)
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
	resp, err := c.client.SendOTP(ctx, req)
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
	resp, err := c.client.VerifyOTP(ctx, req)
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
	resp, err := c.client.LoginWithOTP(ctx, req)
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
	resp, err := c.client.VerifyLoginOTP(ctx, req)
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
	resp, err := c.client.RegisterWithOTP(ctx, req)
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
	resp, err := c.client.VerifyRegistrationOTP(ctx, req)
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
	resp, err := c.client.IntrospectOAuthToken(ctx, &proto.IntrospectOAuthTokenRequest{
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
	resp, err := c.client.ValidateOAuthClient(ctx, &proto.ValidateOAuthClientRequest{
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
	resp, err := c.client.GetOAuthClient(ctx, &proto.GetOAuthClientRequest{
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
