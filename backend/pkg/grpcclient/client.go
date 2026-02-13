package grpcclient

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client provides a convenient interface to the Auth Gateway gRPC service
type Client struct {
	conn        *grpc.ClientConn
	address     string
	timeout     time.Duration
	apiKey      string
	insecure    bool
	tlsCertFile string
}

// NewClient creates a new gRPC client for the Auth Gateway
func NewClient(address string, opts ...Option) (*Client, error) {
	c := &Client{
		address:  address,
		timeout:  10 * time.Second,
		insecure: true,
	}

	for _, opt := range opts {
		opt(c)
	}

	var creds credentials.TransportCredentials
	if c.insecure {
		creds = insecure.NewCredentials()
	} else {
		if c.tlsCertFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(c.tlsCertFile, "")
			if err != nil {
				return nil, err
			}
		} else {
			creds = credentials.NewClientTLSFromCert(nil, "")
		}
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	if c.apiKey != "" {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(apiKeyUnaryInterceptor(c.apiKey)))
	}

	conn, err := grpc.NewClient(address, dialOpts...)
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

// WithAPIKey sets the API key or application secret for authentication.
// Accepts both API keys (agw_*) and application secrets (app_*).
// The key is sent as x-api-key gRPC metadata on every call.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithTLS enables TLS with an optional CA certificate file.
// If certFile is empty, system CA certificates are used.
func WithTLS(certFile string) Option {
	return func(c *Client) {
		c.insecure = false
		c.tlsCertFile = certFile
	}
}

// WithInsecure explicitly enables insecure (non-TLS) connection.
func WithInsecure() Option {
	return func(c *Client) {
		c.insecure = true
	}
}

func apiKeyUnaryInterceptor(apiKey string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		md.Set("x-api-key", apiKey)
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
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
