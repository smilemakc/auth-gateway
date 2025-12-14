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

	ctx, cancel := context.WithTimeout(context.Background(), config.DialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.Address, opts...)
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

// Raw returns the underlying gRPC client for advanced usage.
func (c *GRPCClient) Raw() proto.AuthServiceClient {
	return c.client
}
