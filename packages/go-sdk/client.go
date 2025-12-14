package authgateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// Client is the main Auth Gateway SDK client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string

	// Token management
	accessToken  string
	refreshToken string
	tokenMu      sync.RWMutex
	expiresAt    time.Time

	// Auto-refresh configuration
	autoRefresh bool

	// Services
	Auth        *AuthService
	Profile     *ProfileService
	TwoFactor   *TwoFactorService
	APIKeys     *APIKeysService
	Sessions    *SessionsService
	OTP         *OTPService
	OAuth       *OAuthService
	Passwordless *PasswordlessService
	Admin       *AdminService
}

// Config contains client configuration options.
type Config struct {
	// BaseURL is the Auth Gateway server URL (e.g., "http://localhost:3000")
	BaseURL string

	// HTTPClient is an optional custom HTTP client
	HTTPClient *http.Client

	// Timeout for HTTP requests (default: 30s)
	Timeout time.Duration

	// APIKey for API key authentication (alternative to JWT)
	APIKey string

	// AutoRefresh enables automatic token refresh (default: true)
	AutoRefresh bool
}

// NewClient creates a new Auth Gateway client.
func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:3000"
	}

	// Remove trailing slash
	config.BaseURL = strings.TrimSuffix(config.BaseURL, "/")

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	c := &Client{
		baseURL:     config.BaseURL,
		httpClient:  httpClient,
		apiKey:      config.APIKey,
		autoRefresh: config.AutoRefresh,
	}

	// Initialize services
	c.Auth = &AuthService{client: c}
	c.Profile = &ProfileService{client: c}
	c.TwoFactor = &TwoFactorService{client: c}
	c.APIKeys = &APIKeysService{client: c}
	c.Sessions = &SessionsService{client: c}
	c.OTP = &OTPService{client: c}
	c.OAuth = &OAuthService{client: c}
	c.Passwordless = &PasswordlessService{client: c}
	c.Admin = &AdminService{client: c}

	return c
}

// SetTokens sets the access and refresh tokens.
func (c *Client) SetTokens(accessToken, refreshToken string, expiresIn int64) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	c.accessToken = accessToken
	c.refreshToken = refreshToken
	if expiresIn > 0 {
		c.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
}

// GetAccessToken returns the current access token.
func (c *Client) GetAccessToken() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.accessToken
}

// GetRefreshToken returns the current refresh token.
func (c *Client) GetRefreshToken() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.refreshToken
}

// ClearTokens clears all stored tokens.
func (c *Client) ClearTokens() {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	c.accessToken = ""
	c.refreshToken = ""
	c.expiresAt = time.Time{}
}

// IsAuthenticated returns true if the client has valid tokens.
func (c *Client) IsAuthenticated() bool {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.accessToken != "" || c.apiKey != ""
}

// IsTokenExpired checks if the access token is expired.
func (c *Client) IsTokenExpired() bool {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()

	if c.expiresAt.IsZero() {
		return false
	}
	// Consider token expired 30 seconds before actual expiry
	return time.Now().Add(30 * time.Second).After(c.expiresAt)
}

// request performs an HTTP request with authentication and JSON handling.
func (c *Client) request(ctx context.Context, method, path string, body, result interface{}) error {
	// Check if we need to refresh the token
	if c.autoRefresh && c.IsTokenExpired() && c.refreshToken != "" {
		if err := c.Auth.RefreshTokens(ctx); err != nil {
			// If refresh fails, continue with expired token (server will reject if truly expired)
		}
	}

	reqURL := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return &NetworkError{Err: err}
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set authentication
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	} else {
		c.tokenMu.RLock()
		token := c.accessToken
		c.tokenMu.RUnlock()
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &NetworkError{Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &NetworkError{Err: err}
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		return c.parseErrorResponse(resp.StatusCode, respBody)
	}

	// Parse successful response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// parseErrorResponse parses an error response from the API.
func (c *Client) parseErrorResponse(statusCode int, body []byte) error {
	var errResp models.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return &APIError{
			StatusCode: statusCode,
			Code:       ErrCodeInternalServer,
			Message:    string(body),
		}
	}

	// Handle specific error types
	if errResp.Code == ErrCodeTwoFactorRequired {
		token := ""
		if errResp.Details != nil {
			token = errResp.Details["two_factor_token"]
		}
		return &TwoFactorRequiredError{TwoFactorToken: token}
	}

	return &APIError{
		StatusCode: statusCode,
		Code:       errResp.Code,
		Message:    errResp.Message,
		Details:    errResp.Details,
	}
}

// get performs a GET request.
func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	return c.request(ctx, http.MethodGet, path, nil, result)
}

// post performs a POST request.
func (c *Client) post(ctx context.Context, path string, body, result interface{}) error {
	return c.request(ctx, http.MethodPost, path, body, result)
}

// put performs a PUT request.
func (c *Client) put(ctx context.Context, path string, body, result interface{}) error {
	return c.request(ctx, http.MethodPut, path, body, result)
}

// delete performs a DELETE request.
func (c *Client) delete(ctx context.Context, path string, result interface{}) error {
	return c.request(ctx, http.MethodDelete, path, nil, result)
}

// buildQueryString builds a query string from a struct with url tags.
func buildQueryString(params interface{}) string {
	if params == nil {
		return ""
	}

	v := reflect.ValueOf(params)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}

	t := v.Type()
	values := url.Values{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		tag := fieldType.Tag.Get("url")
		if tag == "" || tag == "-" {
			continue
		}

		// Parse tag options
		parts := strings.Split(tag, ",")
		name := parts[0]
		omitEmpty := len(parts) > 1 && parts[1] == "omitempty"

		// Skip zero values if omitempty
		if omitEmpty && field.IsZero() {
			continue
		}

		// Handle pointer types
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				continue
			}
			field = field.Elem()
		}

		// Convert to string
		var strVal string
		switch field.Kind() {
		case reflect.String:
			strVal = field.String()
		case reflect.Int, reflect.Int64:
			strVal = fmt.Sprintf("%d", field.Int())
		case reflect.Bool:
			strVal = fmt.Sprintf("%t", field.Bool())
		default:
			continue
		}

		if strVal != "" {
			values.Add(name, strVal)
		}
	}

	if len(values) == 0 {
		return ""
	}

	return "?" + values.Encode()
}
