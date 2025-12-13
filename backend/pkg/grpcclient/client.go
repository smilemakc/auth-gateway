package grpcclient

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client provides a convenient interface to the Auth Gateway gRPC service
type Client struct {
	conn    *grpc.ClientConn
	address string
	timeout time.Duration
}

// NewClient creates a new gRPC client for the Auth Gateway
func NewClient(address string, opts ...Option) (*Client, error) {
	c := &Client{
		address: address,
		timeout: 10 * time.Second,
	}

	for _, opt := range opts {
		opt(c)
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	c.conn = conn
	return c, nil
}

// Option is a functional option for configuring the client
type Option func(*Client)

// WithTimeout sets the default timeout for requests
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ValidateToken validates a JWT access token or API key
func (c *Client) ValidateToken(ctx context.Context, accessToken string) (*ValidateTokenResponse, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
	}

	req := &ValidateTokenRequest{AccessToken: accessToken}
	resp := &ValidateTokenResponse{}

	err := c.conn.Invoke(ctx, "/auth.AuthService/ValidateToken", req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetUser retrieves user information by ID
func (c *Client) GetUser(ctx context.Context, userID string) (*GetUserResponse, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
	}

	req := &GetUserRequest{UserId: userID}
	resp := &GetUserResponse{}

	err := c.conn.Invoke(ctx, "/auth.AuthService/GetUser", req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CheckPermission checks if a user has a specific permission
func (c *Client) CheckPermission(ctx context.Context, userID, resource, action string) (*CheckPermissionResponse, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
	}

	req := &CheckPermissionRequest{
		UserId:   userID,
		Resource: resource,
		Action:   action,
	}
	resp := &CheckPermissionResponse{}

	err := c.conn.Invoke(ctx, "/auth.AuthService/CheckPermission", req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// IntrospectToken provides detailed information about a token
func (c *Client) IntrospectToken(ctx context.Context, accessToken string) (*IntrospectTokenResponse, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
	}

	req := &IntrospectTokenRequest{AccessToken: accessToken}
	resp := &IntrospectTokenResponse{}

	err := c.conn.Invoke(ctx, "/auth.AuthService/IntrospectToken", req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
