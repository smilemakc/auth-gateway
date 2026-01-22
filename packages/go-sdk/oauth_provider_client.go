package authgateway

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// OAuthProviderClient implements an OAuth 2.0 / OIDC client for applications
// that want to use Auth Gateway as their identity provider.
//
// Use Cases:
//   - Your web/mobile app wants to authenticate users via Auth Gateway
//   - Your service needs to integrate with Auth Gateway as an OAuth provider
//   - You're building a CLI tool or IoT device that needs Auth Gateway login
//
// Supported Flows:
//   - Authorization Code Flow with PKCE (web/mobile apps)
//   - Device Authorization Flow (TVs, CLI tools, IoT)
//   - Client Credentials Flow (service-to-service)
//   - Refresh Token Flow
//
// Security Features:
//   - PKCE (Proof Key for Code Exchange) enabled by default
//   - State parameter for CSRF protection
//   - Nonce for ID token validation
//   - Automatic OIDC discovery
//
// Example:
//
//	client := authgateway.NewOAuthProviderClient(authgateway.OAuthProviderConfig{
//	    Issuer:       "https://auth.example.com",
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    RedirectURI:  "https://yourapp.com/callback",
//	    Scopes:       []string{"openid", "profile", "email"},
//	})
//
//	authURL, _ := client.GetAuthorizationURL(ctx, nil)
//	// Redirect user to authURL.URL
//
//	tokens, _ := client.ExchangeCode(ctx, code, authURL.CodeVerifier)
//	userInfo, _ := client.GetUserInfo(ctx, tokens.AccessToken)
//
// For complete documentation, see OAUTH_PROVIDER.md

var (
	ErrAuthorizationPending = errors.New("authorization_pending")
	ErrSlowDown             = errors.New("slow_down")
	ErrAccessDenied         = errors.New("access_denied")
	ErrExpiredToken         = errors.New("expired_token")
)

type OAuthProviderConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	UsePKCE      bool
	HTTPClient   *http.Client

	// Headers contains custom headers to include in every request.
	// Common headers: X-Application-ID, X-Client-Name, etc.
	Headers map[string]string
}

type OAuthProviderClient struct {
	config     OAuthProviderConfig
	httpClient *http.Client

	discovery     *models.OIDCDiscovery
	discoveryOnce sync.Once

	jwks   *models.JWKS
	jwksMu sync.RWMutex

	// Custom headers to include in every request
	headers   map[string]string
	headersMu sync.RWMutex
}

func NewOAuthProviderClient(config OAuthProviderConfig) *OAuthProviderClient {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if config.Scopes == nil {
		config.Scopes = []string{"openid"}
	}
	if !config.UsePKCE {
		config.UsePKCE = true
	}

	return &OAuthProviderClient{
		config:     config,
		httpClient: config.HTTPClient,
		headers:    config.Headers,
	}
}

// SetHeader sets a custom header to be included in all requests.
func (c *OAuthProviderClient) SetHeader(key, value string) {
	c.headersMu.Lock()
	defer c.headersMu.Unlock()

	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value
}

// SetHeaders sets multiple custom headers to be included in all requests.
func (c *OAuthProviderClient) SetHeaders(headers map[string]string) {
	c.headersMu.Lock()
	defer c.headersMu.Unlock()

	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	for k, v := range headers {
		c.headers[k] = v
	}
}

// RemoveHeader removes a custom header.
func (c *OAuthProviderClient) RemoveHeader(key string) {
	c.headersMu.Lock()
	defer c.headersMu.Unlock()

	delete(c.headers, key)
}

// GetHeaders returns a copy of the current custom headers.
func (c *OAuthProviderClient) GetHeaders() map[string]string {
	c.headersMu.RLock()
	defer c.headersMu.RUnlock()

	result := make(map[string]string)
	for k, v := range c.headers {
		result[k] = v
	}
	return result
}

// SetApplicationID sets the X-Application-ID header for multi-tenant support.
func (c *OAuthProviderClient) SetApplicationID(appID string) {
	c.SetHeader("X-Application-ID", appID)
}

// SetClientName sets the X-Client-Name header for client identification.
func (c *OAuthProviderClient) SetClientName(name string) {
	c.SetHeader("X-Client-Name", name)
}

// applyHeaders applies custom headers to an HTTP request.
func (c *OAuthProviderClient) applyHeaders(req *http.Request) {
	c.headersMu.RLock()
	defer c.headersMu.RUnlock()

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
}

func (c *OAuthProviderClient) GetDiscovery(ctx context.Context) (*models.OIDCDiscovery, error) {
	var err error
	c.discoveryOnce.Do(func() {
		url := c.config.Issuer + "/.well-known/openid-configuration"
		req, reqErr := http.NewRequestWithContext(ctx, "GET", url, nil)
		if reqErr != nil {
			err = reqErr
			return
		}
		c.applyHeaders(req)

		resp, respErr := c.httpClient.Do(req)
		if respErr != nil {
			err = respErr
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("discovery endpoint returned status %d", resp.StatusCode)
			return
		}

		c.discovery = &models.OIDCDiscovery{}
		err = json.NewDecoder(resp.Body).Decode(c.discovery)
	})

	if err != nil {
		return nil, err
	}
	return c.discovery, nil
}

func (c *OAuthProviderClient) GetJWKS(ctx context.Context) (*models.JWKS, error) {
	c.jwksMu.RLock()
	if c.jwks != nil {
		c.jwksMu.RUnlock()
		return c.jwks, nil
	}
	c.jwksMu.RUnlock()

	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", discovery.JwksURI, nil)
	if err != nil {
		return nil, err
	}
	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	c.jwksMu.Lock()
	defer c.jwksMu.Unlock()

	c.jwks = &models.JWKS{}
	if err := json.NewDecoder(resp.Body).Decode(c.jwks); err != nil {
		return nil, err
	}

	return c.jwks, nil
}

type AuthorizationURLResult struct {
	URL          string
	State        string
	Nonce        string
	CodeVerifier string
}

type AuthorizationURLOptions struct {
	Scope     string
	State     string
	Nonce     string
	Prompt    string
	LoginHint string
}

func (c *OAuthProviderClient) GetAuthorizationURL(ctx context.Context, opts *AuthorizationURLOptions) (*AuthorizationURLResult, error) {
	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &AuthorizationURLOptions{}
	}

	state := opts.State
	if state == "" {
		state = generateRandomString(32)
	}

	nonce := opts.Nonce
	if nonce == "" {
		nonce = generateRandomString(32)
	}

	scope := opts.Scope
	if scope == "" {
		scope = strings.Join(c.config.Scopes, " ")
	}

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {c.config.ClientID},
		"redirect_uri":  {c.config.RedirectURI},
		"scope":         {scope},
		"state":         {state},
		"nonce":         {nonce},
	}

	if opts.Prompt != "" {
		params.Set("prompt", opts.Prompt)
	}
	if opts.LoginHint != "" {
		params.Set("login_hint", opts.LoginHint)
	}

	result := &AuthorizationURLResult{
		State: state,
		Nonce: nonce,
	}

	if c.config.UsePKCE {
		pkce, err := GeneratePKCE()
		if err != nil {
			return nil, err
		}
		result.CodeVerifier = pkce.CodeVerifier
		params.Set("code_challenge", pkce.CodeChallenge)
		params.Set("code_challenge_method", pkce.CodeChallengeMethod)
	}

	result.URL = discovery.AuthorizationEndpoint + "?" + params.Encode()
	return result, nil
}

func (c *OAuthProviderClient) ExchangeCode(ctx context.Context, code, codeVerifier string) (*models.OAuthTokenResponse, error) {
	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	params := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.config.RedirectURI},
		"client_id":    {c.config.ClientID},
	}

	if c.config.ClientSecret != "" {
		params.Set("client_secret", c.config.ClientSecret)
	}

	if codeVerifier != "" {
		params.Set("code_verifier", codeVerifier)
	}

	return c.tokenRequest(ctx, discovery.TokenEndpoint, params)
}

func (c *OAuthProviderClient) RefreshTokens(ctx context.Context, refreshToken string) (*models.OAuthTokenResponse, error) {
	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	params := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {c.config.ClientID},
	}

	if c.config.ClientSecret != "" {
		params.Set("client_secret", c.config.ClientSecret)
	}

	return c.tokenRequest(ctx, discovery.TokenEndpoint, params)
}

func (c *OAuthProviderClient) IntrospectToken(ctx context.Context, token string) (*models.TokenIntrospectionResponse, error) {
	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	params := url.Values{"token": {token}}

	req, err := http.NewRequestWithContext(ctx, "POST", discovery.IntrospectionEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.applyHeaders(req)
	c.setClientAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection failed with status %d", resp.StatusCode)
	}

	result := &models.TokenIntrospectionResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *OAuthProviderClient) RevokeToken(ctx context.Context, token, tokenTypeHint string) error {
	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return err
	}

	params := url.Values{"token": {token}}
	if tokenTypeHint != "" {
		params.Set("token_type_hint", tokenTypeHint)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", discovery.RevocationEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.applyHeaders(req)
	c.setClientAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *OAuthProviderClient) GetUserInfo(ctx context.Context, accessToken string) (*models.UserInfo, error) {
	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", discovery.UserInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}

	c.applyHeaders(req)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo failed with status %d", resp.StatusCode)
	}

	result := &models.UserInfo{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *OAuthProviderClient) RequestDeviceCode(ctx context.Context, scopes []string) (*models.DeviceAuthResponse, error) {
	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	scope := strings.Join(c.config.Scopes, " ")
	if len(scopes) > 0 {
		scope = strings.Join(scopes, " ")
	}

	params := url.Values{
		"client_id": {c.config.ClientID},
		"scope":     {scope},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", discovery.DeviceAuthorizationEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device authorization failed with status %d", resp.StatusCode)
	}

	result := &models.DeviceAuthResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *OAuthProviderClient) PollDeviceToken(ctx context.Context, deviceCode string) (*models.OAuthTokenResponse, error) {
	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	params := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": {deviceCode},
		"client_id":   {c.config.ClientID},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", discovery.TokenEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var oauthErr models.OAuthError
		if err := json.NewDecoder(resp.Body).Decode(&oauthErr); err != nil {
			return nil, fmt.Errorf("token request failed with status %d", resp.StatusCode)
		}

		switch oauthErr.Error {
		case "authorization_pending":
			return nil, ErrAuthorizationPending
		case "slow_down":
			return nil, ErrSlowDown
		case "access_denied":
			return nil, ErrAccessDenied
		case "expired_token":
			return nil, ErrExpiredToken
		default:
			return nil, fmt.Errorf("%s: %s", oauthErr.Error, oauthErr.ErrorDescription)
		}
	}

	result := &models.OAuthTokenResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *OAuthProviderClient) ClientCredentialsGrant(ctx context.Context, scopes []string) (*models.OAuthTokenResponse, error) {
	if c.config.ClientSecret == "" {
		return nil, errors.New("client_credentials grant requires client_secret")
	}

	discovery, err := c.GetDiscovery(ctx)
	if err != nil {
		return nil, err
	}

	scope := ""
	if len(scopes) > 0 {
		scope = strings.Join(scopes, " ")
	}

	params := url.Values{
		"grant_type": {"client_credentials"},
	}
	if scope != "" {
		params.Set("scope", scope)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", discovery.TokenEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.applyHeaders(req)
	c.setClientAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client credentials grant failed with status %d", resp.StatusCode)
	}

	result := &models.OAuthTokenResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *OAuthProviderClient) tokenRequest(ctx context.Context, endpoint string, params url.Values) (*models.OAuthTokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var oauthErr models.OAuthError
		if err := json.NewDecoder(resp.Body).Decode(&oauthErr); err != nil {
			return nil, fmt.Errorf("token request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s: %s", oauthErr.Error, oauthErr.ErrorDescription)
	}

	result := &models.OAuthTokenResponse{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *OAuthProviderClient) setClientAuth(req *http.Request) {
	if c.config.ClientSecret != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(c.config.ClientID + ":" + c.config.ClientSecret))
		req.Header.Set("Authorization", "Basic "+auth)
	}
}

func GeneratePKCE() (*models.PKCEParams, error) {
	verifier := generateRandomString(64)

	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &models.PKCEParams{
		CodeVerifier:        verifier,
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	}, nil
}

func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}
