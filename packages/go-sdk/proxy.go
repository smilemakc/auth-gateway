package authgateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// ProxyConfig contains configuration for the auth proxy.
type ProxyConfig struct {
	BaseURL       string
	APIKey        string
	ApplicationID string
	HTTPClient    *http.Client
	Timeout       time.Duration
}

// AuthProxy proxies authentication requests to the auth gateway.
// It is designed to be framework-agnostic and can be used by any Go product
// to forward authentication requests from their frontend to the auth gateway.
type AuthProxy struct {
	baseURL       string
	apiKey        string
	applicationID string
	httpClient    *http.Client
}

// NewAuthProxy creates a new auth proxy.
func NewAuthProxy(config ProxyConfig) *AuthProxy {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:3000"
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	return &AuthProxy{
		baseURL:       config.BaseURL,
		apiKey:        config.APIKey,
		applicationID: config.ApplicationID,
		httpClient:    httpClient,
	}
}

// AuthProxyResponse contains the auth response from the gateway.
type AuthProxyResponse struct {
	AccessToken       string                 `json:"access_token"`
	RefreshToken      string                 `json:"refresh_token"`
	ExpiresIn         int64                  `json:"expires_in"`
	User              map[string]interface{} `json:"user,omitempty"`
	RequiresTwoFactor bool                   `json:"requires_two_factor,omitempty"`
	TwoFactorToken    string                 `json:"two_factor_token,omitempty"`
}

// SignInProxyRequest contains sign-in request data.
type SignInProxyRequest struct {
	Email    string  `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Password string  `json:"password"`
}

// SignUpProxyRequest contains sign-up request data.
type SignUpProxyRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username,omitempty"`
	FullName string `json:"full_name,omitempty"`
}

// SignIn proxies a sign-in request to the auth gateway.
func (p *AuthProxy) SignIn(ctx context.Context, req SignInProxyRequest) (*AuthProxyResponse, error) {
	var resp AuthProxyResponse
	if err := p.proxyRequest(ctx, http.MethodPost, "/api/auth/signin", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SignUp proxies a sign-up request to the auth gateway.
func (p *AuthProxy) SignUp(ctx context.Context, req SignUpProxyRequest) (*AuthProxyResponse, error) {
	var resp AuthProxyResponse
	if err := p.proxyRequest(ctx, http.MethodPost, "/api/auth/signup", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendOTP proxies an OTP send request to the auth gateway.
func (p *AuthProxy) SendOTP(ctx context.Context, email string) error {
	reqBody := map[string]string{
		"email": email,
	}
	return p.proxyRequest(ctx, http.MethodPost, "/api/auth/otp/send", reqBody, nil)
}

// VerifyOTP proxies an OTP verification request to the auth gateway.
func (p *AuthProxy) VerifyOTP(ctx context.Context, email, code string) (*AuthProxyResponse, error) {
	reqBody := map[string]string{
		"email": email,
		"code":  code,
	}

	var resp AuthProxyResponse
	if err := p.proxyRequest(ctx, http.MethodPost, "/api/auth/otp/verify", reqBody, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RefreshToken proxies a token refresh request to the auth gateway.
func (p *AuthProxy) RefreshToken(ctx context.Context, refreshToken string) (*AuthProxyResponse, error) {
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	var resp AuthProxyResponse
	if err := p.proxyRequest(ctx, http.MethodPost, "/api/auth/refresh", reqBody, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Logout proxies a logout request to the auth gateway.
func (p *AuthProxy) Logout(ctx context.Context, accessToken string) error {
	return p.proxyRequestWithAuth(ctx, http.MethodPost, "/api/auth/logout", nil, nil, accessToken)
}

// proxyRequest performs an HTTP request to the auth gateway with API key authentication.
func (p *AuthProxy) proxyRequest(ctx context.Context, method, path string, body, result interface{}) error {
	return p.proxyRequestWithAuth(ctx, method, path, body, result, "")
}

// proxyRequestWithAuth performs an HTTP request to the auth gateway with optional Bearer token.
func (p *AuthProxy) proxyRequestWithAuth(ctx context.Context, method, path string, body, result interface{}, bearerToken string) error {
	reqURL := p.baseURL + path

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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if p.apiKey != "" {
		req.Header.Set("X-API-Key", p.apiKey)
	}

	if p.applicationID != "" {
		req.Header.Set("X-Application-ID", p.applicationID)
	}

	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &NetworkError{Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &NetworkError{Err: err}
	}

	if resp.StatusCode >= 400 {
		return p.parseProxyErrorResponse(resp.StatusCode, respBody)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// parseProxyErrorResponse parses an error response from the API.
func (p *AuthProxy) parseProxyErrorResponse(statusCode int, body []byte) error {
	var errResp models.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return &APIError{
			StatusCode: statusCode,
			Code:       ErrCodeInternalServer,
			Message:    string(body),
		}
	}

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
