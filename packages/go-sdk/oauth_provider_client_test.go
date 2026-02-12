package authgateway

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// discoveryResponse returns a mock OIDC discovery document
func discoveryResponse(baseURL string) models.OIDCDiscovery {
	return models.OIDCDiscovery{
		Issuer:                           baseURL,
		AuthorizationEndpoint:            baseURL + "/oauth/authorize",
		TokenEndpoint:                    baseURL + "/oauth/token",
		UserInfoEndpoint:                 baseURL + "/oauth/userinfo",
		JwksURI:                          baseURL + "/oauth/jwks",
		RevocationEndpoint:               baseURL + "/oauth/revoke",
		IntrospectionEndpoint:            baseURL + "/oauth/introspect",
		DeviceAuthorizationEndpoint:      baseURL + "/oauth/device",
		ResponseTypesSupported:           []string{"code"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		ScopesSupported:                  []string{"openid", "profile", "email"},
		GrantTypesSupported:              []string{"authorization_code", "refresh_token", "client_credentials"},
		CodeChallengeMethodsSupported:    []string{"S256"},
	}
}

// testServer wraps httptest.Server with helper methods
type testServer struct {
	*httptest.Server
}

// newTestServer creates a test HTTP server
func newTestServer(t *testing.T, handler http.Handler) *testServer {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &testServer{server}
}

// discoveryHandler returns a handler that serves the discovery document
func discoveryHandler(serverURL *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		discovery := discoveryResponse(*serverURL)
		json.NewEncoder(w).Encode(discovery)
	}
}

// TestNewOAuthProviderClient_Configuration tests the NewOAuthProviderClient configuration
func TestNewOAuthProviderClient_Configuration(t *testing.T) {
	t.Run("ShouldUseDefaultHTTPClient_WhenNotProvided", func(t *testing.T) {
		// Arrange & Act
		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   "https://auth.example.com",
			ClientID: "test-client",
		})

		// Assert
		if client.httpClient == nil {
			t.Error("expected HTTP client to be initialized")
		}
	})

	t.Run("ShouldUseCustomHTTPClient_WhenProvided", func(t *testing.T) {
		// Arrange
		customClient := &http.Client{Timeout: 60 * time.Second}

		// Act
		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:     "https://auth.example.com",
			ClientID:   "test-client",
			HTTPClient: customClient,
		})

		// Assert
		if client.httpClient != customClient {
			t.Error("expected custom HTTP client to be used")
		}
	})

	t.Run("ShouldUseDefaultScopes_WhenNotProvided", func(t *testing.T) {
		// Arrange & Act
		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   "https://auth.example.com",
			ClientID: "test-client",
		})

		// Assert
		if len(client.config.Scopes) != 1 || client.config.Scopes[0] != "openid" {
			t.Errorf("expected default scopes ['openid'], got %v", client.config.Scopes)
		}
	})

	t.Run("ShouldUseCustomScopes_WhenProvided", func(t *testing.T) {
		// Arrange
		customScopes := []string{"openid", "profile", "email"}

		// Act
		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   "https://auth.example.com",
			ClientID: "test-client",
			Scopes:   customScopes,
		})

		// Assert
		if len(client.config.Scopes) != 3 {
			t.Errorf("expected 3 scopes, got %d", len(client.config.Scopes))
		}
	})

	t.Run("ShouldEnablePKCEByDefault", func(t *testing.T) {
		// Arrange & Act
		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   "https://auth.example.com",
			ClientID: "test-client",
		})

		// Assert
		if !client.config.UsePKCE {
			t.Error("expected PKCE to be enabled by default")
		}
	})

	t.Run("ShouldPreserveAllConfigFields", func(t *testing.T) {
		// Arrange
		config := OAuthProviderConfig{
			Issuer:       "https://auth.example.com",
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			RedirectURI:  "https://app.example.com/callback",
			Scopes:       []string{"openid", "profile"},
			UsePKCE:      true,
		}

		// Act
		client := NewOAuthProviderClient(config)

		// Assert
		if client.config.Issuer != config.Issuer {
			t.Errorf("expected issuer %s, got %s", config.Issuer, client.config.Issuer)
		}
		if client.config.ClientID != config.ClientID {
			t.Errorf("expected client ID %s, got %s", config.ClientID, client.config.ClientID)
		}
		if client.config.ClientSecret != config.ClientSecret {
			t.Errorf("expected client secret %s, got %s", config.ClientSecret, client.config.ClientSecret)
		}
		if client.config.RedirectURI != config.RedirectURI {
			t.Errorf("expected redirect URI %s, got %s", config.RedirectURI, client.config.RedirectURI)
		}
	})
}

// TestGetAuthorizationURL tests authorization URL generation
func TestGetAuthorizationURL(t *testing.T) {
	t.Run("ShouldGenerateValidURL_WithPKCE", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:      serverURL,
			ClientID:    "test-client",
			RedirectURI: "https://app.example.com/callback",
			Scopes:      []string{"openid", "profile"},
			UsePKCE:     true,
		})

		// Act
		result, err := client.GetAuthorizationURL(context.Background(), nil)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.URL == "" {
			t.Error("expected non-empty URL")
		}
		if result.State == "" {
			t.Error("expected non-empty state")
		}
		if result.Nonce == "" {
			t.Error("expected non-empty nonce")
		}
		if result.CodeVerifier == "" {
			t.Error("expected non-empty code verifier with PKCE enabled")
		}

		// Parse and validate URL
		parsedURL, err := url.Parse(result.URL)
		if err != nil {
			t.Fatalf("failed to parse URL: %v", err)
		}

		queryParams := parsedURL.Query()
		if queryParams.Get("response_type") != "code" {
			t.Error("expected response_type=code")
		}
		if queryParams.Get("client_id") != "test-client" {
			t.Errorf("expected client_id=test-client, got %s", queryParams.Get("client_id"))
		}
		if queryParams.Get("redirect_uri") != "https://app.example.com/callback" {
			t.Errorf("expected correct redirect_uri, got %s", queryParams.Get("redirect_uri"))
		}
		if queryParams.Get("state") != result.State {
			t.Error("state parameter does not match result")
		}
		if queryParams.Get("nonce") != result.Nonce {
			t.Error("nonce parameter does not match result")
		}
		if queryParams.Get("code_challenge") == "" {
			t.Error("expected code_challenge parameter with PKCE")
		}
		if queryParams.Get("code_challenge_method") != "S256" {
			t.Error("expected code_challenge_method=S256")
		}
	})

	t.Run("ShouldUseProvidedStateAndNonce", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:      serverURL,
			ClientID:    "test-client",
			RedirectURI: "https://app.example.com/callback",
		})

		// Act
		result, err := client.GetAuthorizationURL(context.Background(), &AuthorizationURLOptions{
			State: "custom-state",
			Nonce: "custom-nonce",
		})

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.State != "custom-state" {
			t.Errorf("expected state=custom-state, got %s", result.State)
		}
		if result.Nonce != "custom-nonce" {
			t.Errorf("expected nonce=custom-nonce, got %s", result.Nonce)
		}
	})

	t.Run("ShouldIncludePromptAndLoginHint_WhenProvided", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:      serverURL,
			ClientID:    "test-client",
			RedirectURI: "https://app.example.com/callback",
		})

		// Act
		result, err := client.GetAuthorizationURL(context.Background(), &AuthorizationURLOptions{
			Prompt:    "consent",
			LoginHint: "user@example.com",
		})

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsedURL, _ := url.Parse(result.URL)
		queryParams := parsedURL.Query()
		if queryParams.Get("prompt") != "consent" {
			t.Errorf("expected prompt=consent, got %s", queryParams.Get("prompt"))
		}
		if queryParams.Get("login_hint") != "user@example.com" {
			t.Errorf("expected login_hint=user@example.com, got %s", queryParams.Get("login_hint"))
		}
	})

	t.Run("ShouldUseCustomScope_WhenProvided", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:      serverURL,
			ClientID:    "test-client",
			RedirectURI: "https://app.example.com/callback",
			Scopes:      []string{"openid", "profile"},
		})

		// Act
		result, err := client.GetAuthorizationURL(context.Background(), &AuthorizationURLOptions{
			Scope: "openid email offline_access",
		})

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsedURL, _ := url.Parse(result.URL)
		queryParams := parsedURL.Query()
		if queryParams.Get("scope") != "openid email offline_access" {
			t.Errorf("expected custom scope, got %s", queryParams.Get("scope"))
		}
	})

	t.Run("ShouldReturnError_WhenDiscoveryFails", func(t *testing.T) {
		// Arrange
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		server := newTestServer(t, mux)

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   server.URL,
			ClientID: "test-client",
		})

		// Act
		result, err := client.GetAuthorizationURL(context.Background(), nil)

		// Assert
		if err == nil {
			t.Error("expected error when discovery fails")
		}
		if result != nil {
			t.Error("expected nil result when discovery fails")
		}
	})
}

// TestGeneratePKCE tests PKCE generation
func TestGeneratePKCE(t *testing.T) {
	t.Run("ShouldGenerateValidPKCEParams", func(t *testing.T) {
		// Act
		pkce, err := GeneratePKCE()

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pkce.CodeVerifier == "" {
			t.Error("expected non-empty code verifier")
		}
		if pkce.CodeChallenge == "" {
			t.Error("expected non-empty code challenge")
		}
		if pkce.CodeChallengeMethod != "S256" {
			t.Errorf("expected S256 method, got %s", pkce.CodeChallengeMethod)
		}
	})

	t.Run("ShouldGenerateCodeVerifierWithCorrectLength", func(t *testing.T) {
		// Act
		pkce, _ := GeneratePKCE()

		// Assert - verifier should be 64 characters
		if len(pkce.CodeVerifier) != 64 {
			t.Errorf("expected 64 character verifier, got %d", len(pkce.CodeVerifier))
		}
	})

	t.Run("ShouldGenerateValidS256Challenge", func(t *testing.T) {
		// Act
		pkce, _ := GeneratePKCE()

		// Verify - compute the expected challenge
		hash := sha256.Sum256([]byte(pkce.CodeVerifier))
		expectedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

		// Assert
		if pkce.CodeChallenge != expectedChallenge {
			t.Errorf("code challenge mismatch: expected %s, got %s", expectedChallenge, pkce.CodeChallenge)
		}
	})

	t.Run("ShouldGenerateUniqueVerifiers", func(t *testing.T) {
		// Arrange
		seen := make(map[string]bool)

		// Act & Assert
		for i := 0; i < 100; i++ {
			pkce, _ := GeneratePKCE()
			if seen[pkce.CodeVerifier] {
				t.Error("generated duplicate code verifier")
			}
			seen[pkce.CodeVerifier] = true
		}
	})

	t.Run("ShouldGenerateURLSafeChallenge", func(t *testing.T) {
		// Act
		pkce, _ := GeneratePKCE()

		// Assert - challenge should not contain URL-unsafe characters
		if strings.ContainsAny(pkce.CodeChallenge, "+/=") {
			t.Error("code challenge contains URL-unsafe characters")
		}
	})
}

// TestGenerateRandomString tests random string generation
func TestGenerateRandomString(t *testing.T) {
	t.Run("ShouldGenerateStringOfCorrectLength", func(t *testing.T) {
		testCases := []int{16, 32, 64, 128}
		for _, length := range testCases {
			// Act
			result := generateRandomString(length)

			// Assert
			if len(result) != length {
				t.Errorf("expected length %d, got %d", length, len(result))
			}
		}
	})

	t.Run("ShouldGenerateUniqueStrings", func(t *testing.T) {
		// Arrange
		seen := make(map[string]bool)

		// Act & Assert
		for i := 0; i < 100; i++ {
			result := generateRandomString(32)
			if seen[result] {
				t.Error("generated duplicate string")
			}
			seen[result] = true
		}
	})

	t.Run("ShouldOnlyContainValidCharacters", func(t *testing.T) {
		// Arrange
		validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"

		// Act
		result := generateRandomString(1000)

		// Assert
		for _, c := range result {
			if !strings.ContainsRune(validChars, c) {
				t.Errorf("invalid character found: %c", c)
			}
		}
	})
}

// TestExchangeCode tests token exchange
func TestExchangeCode(t *testing.T) {
	t.Run("ShouldExchangeCodeSuccessfully", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedBody string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			capturedBody = r.Form.Encode()

			response := models.OAuthTokenResponse{
				AccessToken:  "test-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: "test-refresh-token",
				IDToken:      "test-id-token",
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       serverURL,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			RedirectURI:  "https://app.example.com/callback",
		})

		// Act
		result, err := client.ExchangeCode(context.Background(), "auth-code", "code-verifier")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != "test-access-token" {
			t.Errorf("expected access token test-access-token, got %s", result.AccessToken)
		}
		if result.RefreshToken != "test-refresh-token" {
			t.Errorf("expected refresh token test-refresh-token, got %s", result.RefreshToken)
		}
		if result.ExpiresIn != 3600 {
			t.Errorf("expected expires_in 3600, got %d", result.ExpiresIn)
		}

		// Verify request body
		if !strings.Contains(capturedBody, "grant_type=authorization_code") {
			t.Error("expected grant_type=authorization_code in request")
		}
		if !strings.Contains(capturedBody, "code=auth-code") {
			t.Error("expected code parameter in request")
		}
		if !strings.Contains(capturedBody, "code_verifier=code-verifier") {
			t.Error("expected code_verifier parameter in request")
		}
	})

	t.Run("ShouldExchangeCodeWithoutCodeVerifier", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedBody string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			capturedBody = r.Form.Encode()

			response := models.OAuthTokenResponse{
				AccessToken: "test-access-token",
				TokenType:   "Bearer",
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:      serverURL,
			ClientID:    "test-client",
			RedirectURI: "https://app.example.com/callback",
			UsePKCE:     false,
		})

		// Act
		result, err := client.ExchangeCode(context.Background(), "auth-code", "")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != "test-access-token" {
			t.Errorf("expected access token, got %s", result.AccessToken)
		}
		if strings.Contains(capturedBody, "code_verifier") {
			t.Error("should not include code_verifier when empty")
		}
	})

	t.Run("ShouldReturnError_WhenTokenEndpointFails", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.OAuthError{
				Error:            "invalid_grant",
				ErrorDescription: "authorization code expired",
			})
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:      serverURL,
			ClientID:    "test-client",
			RedirectURI: "https://app.example.com/callback",
		})

		// Act
		result, err := client.ExchangeCode(context.Background(), "invalid-code", "")

		// Assert
		if err == nil {
			t.Error("expected error when token endpoint fails")
		}
		if result != nil {
			t.Error("expected nil result when token endpoint fails")
		}
		if !strings.Contains(err.Error(), "invalid_grant") {
			t.Errorf("error should contain OAuth error code, got: %v", err)
		}
	})

	t.Run("ShouldIncludeClientSecret_WhenProvided", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedBody string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			capturedBody = r.Form.Encode()

			response := models.OAuthTokenResponse{
				AccessToken: "test-access-token",
				TokenType:   "Bearer",
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       serverURL,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			RedirectURI:  "https://app.example.com/callback",
		})

		// Act
		_, err := client.ExchangeCode(context.Background(), "auth-code", "")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(capturedBody, "client_secret=test-secret") {
			t.Error("expected client_secret in request body")
		}
	})
}

// TestRefreshTokens tests token refresh
func TestRefreshTokens(t *testing.T) {
	t.Run("ShouldRefreshTokensSuccessfully", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedBody string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			capturedBody = r.Form.Encode()

			response := models.OAuthTokenResponse{
				AccessToken:  "new-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: "new-refresh-token",
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       serverURL,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		})

		// Act
		result, err := client.RefreshTokens(context.Background(), "old-refresh-token")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != "new-access-token" {
			t.Errorf("expected new-access-token, got %s", result.AccessToken)
		}
		if result.RefreshToken != "new-refresh-token" {
			t.Errorf("expected new-refresh-token, got %s", result.RefreshToken)
		}

		// Verify request
		if !strings.Contains(capturedBody, "grant_type=refresh_token") {
			t.Error("expected grant_type=refresh_token")
		}
		if !strings.Contains(capturedBody, "refresh_token=old-refresh-token") {
			t.Error("expected refresh_token parameter")
		}
	})

	t.Run("ShouldReturnError_WhenRefreshFails", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.OAuthError{
				Error:            "invalid_grant",
				ErrorDescription: "refresh token expired",
			})
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		result, err := client.RefreshTokens(context.Background(), "expired-refresh-token")

		// Assert
		if err == nil {
			t.Error("expected error when refresh fails")
		}
		if result != nil {
			t.Error("expected nil result when refresh fails")
		}
	})
}

// TestGetDiscovery tests OIDC discovery
func TestGetDiscovery(t *testing.T) {
	t.Run("ShouldFetchDiscoveryDocument", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		discovery, err := client.GetDiscovery(context.Background())

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if discovery.Issuer != serverURL {
			t.Errorf("expected issuer %s, got %s", serverURL, discovery.Issuer)
		}
		if discovery.AuthorizationEndpoint != serverURL+"/oauth/authorize" {
			t.Error("authorization endpoint mismatch")
		}
		if discovery.TokenEndpoint != serverURL+"/oauth/token" {
			t.Error("token endpoint mismatch")
		}
	})

	t.Run("ShouldCacheDiscoveryDocument", func(t *testing.T) {
		// Arrange
		var serverURL string
		var callCount int32
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act - call multiple times
		client.GetDiscovery(context.Background())
		client.GetDiscovery(context.Background())
		client.GetDiscovery(context.Background())

		// Assert
		if atomic.LoadInt32(&callCount) != 1 {
			t.Errorf("expected 1 call to discovery endpoint, got %d", callCount)
		}
	})

	t.Run("ShouldReturnError_WhenDiscoveryEndpointFails", func(t *testing.T) {
		// Arrange
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		server := newTestServer(t, mux)

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   server.URL,
			ClientID: "test-client",
		})

		// Act
		discovery, err := client.GetDiscovery(context.Background())

		// Assert
		if err == nil {
			t.Error("expected error when discovery endpoint fails")
		}
		if discovery != nil {
			t.Error("expected nil discovery when endpoint fails")
		}
	})
}

// TestGetJWKS tests JWKS fetching
func TestGetJWKS(t *testing.T) {
	t.Run("ShouldFetchJWKS", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/jwks", func(w http.ResponseWriter, r *http.Request) {
			jwks := models.JWKS{
				Keys: []models.JWK{
					{
						Kty: "RSA",
						Use: "sig",
						Kid: "key-1",
						Alg: "RS256",
						N:   "test-modulus",
						E:   "AQAB",
					},
				},
			}
			json.NewEncoder(w).Encode(jwks)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		jwks, err := client.GetJWKS(context.Background())

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(jwks.Keys) != 1 {
			t.Errorf("expected 1 key, got %d", len(jwks.Keys))
		}
		if jwks.Keys[0].Kid != "key-1" {
			t.Errorf("expected key ID key-1, got %s", jwks.Keys[0].Kid)
		}
	})

	t.Run("ShouldCacheJWKS", func(t *testing.T) {
		// Arrange
		var serverURL string
		var jwksCallCount int32
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/jwks", func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&jwksCallCount, 1)
			jwks := models.JWKS{Keys: []models.JWK{}}
			json.NewEncoder(w).Encode(jwks)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		client.GetJWKS(context.Background())
		client.GetJWKS(context.Background())

		// Assert
		if atomic.LoadInt32(&jwksCallCount) != 1 {
			t.Errorf("expected 1 call to JWKS endpoint, got %d", jwksCallCount)
		}
	})

	t.Run("ShouldReturnError_WhenJWKSEndpointFails", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/jwks", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		jwks, err := client.GetJWKS(context.Background())

		// Assert
		if err == nil {
			t.Error("expected error when JWKS endpoint fails")
		}
		if jwks != nil {
			t.Error("expected nil JWKS when endpoint fails")
		}
	})
}

// TestIntrospectToken tests token introspection
func TestIntrospectToken(t *testing.T) {
	t.Run("ShouldIntrospectActiveToken", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedAuth string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/introspect", func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			response := models.TokenIntrospectionResponse{
				Active:    true,
				Scope:     "openid profile",
				ClientID:  "test-client",
				Username:  "testuser",
				TokenType: "Bearer",
				Exp:       time.Now().Add(time.Hour).Unix(),
				Sub:       "user-123",
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       serverURL,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		})

		// Act
		result, err := client.IntrospectToken(context.Background(), "test-token")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Active {
			t.Error("expected token to be active")
		}
		if result.Username != "testuser" {
			t.Errorf("expected username testuser, got %s", result.Username)
		}
		if result.Sub != "user-123" {
			t.Errorf("expected sub user-123, got %s", result.Sub)
		}

		// Verify Basic auth header
		if !strings.HasPrefix(capturedAuth, "Basic ") {
			t.Error("expected Basic authentication header")
		}
	})

	t.Run("ShouldReturnInactiveForExpiredToken", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/introspect", func(w http.ResponseWriter, r *http.Request) {
			response := models.TokenIntrospectionResponse{
				Active: false,
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       serverURL,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		})

		// Act
		result, err := client.IntrospectToken(context.Background(), "expired-token")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Active {
			t.Error("expected token to be inactive")
		}
	})

	t.Run("ShouldReturnError_WhenIntrospectionFails", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/introspect", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		result, err := client.IntrospectToken(context.Background(), "test-token")

		// Assert
		if err == nil {
			t.Error("expected error when introspection fails")
		}
		if result != nil {
			t.Error("expected nil result when introspection fails")
		}
	})
}

// TestRevokeToken tests token revocation
func TestRevokeToken(t *testing.T) {
	t.Run("ShouldRevokeTokenSuccessfully", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedBody string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/revoke", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			capturedBody = r.Form.Encode()
			w.WriteHeader(http.StatusOK)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       serverURL,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		})

		// Act
		err := client.RevokeToken(context.Background(), "test-token", "access_token")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(capturedBody, "token=test-token") {
			t.Error("expected token parameter in request")
		}
		if !strings.Contains(capturedBody, "token_type_hint=access_token") {
			t.Error("expected token_type_hint parameter in request")
		}
	})

	t.Run("ShouldRevokeTokenWithoutTypeHint", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedBody string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/revoke", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			capturedBody = r.Form.Encode()
			w.WriteHeader(http.StatusOK)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       serverURL,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		})

		// Act
		err := client.RevokeToken(context.Background(), "test-token", "")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(capturedBody, "token_type_hint") {
			t.Error("should not include token_type_hint when empty")
		}
	})
}

// TestGetUserInfo tests userinfo endpoint
func TestGetUserInfo(t *testing.T) {
	t.Run("ShouldFetchUserInfoSuccessfully", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedAuth string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/userinfo", func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			response := models.UserInfo{
				Sub:           "user-123",
				Name:          "Test User",
				Email:         "test@example.com",
				EmailVerified: true,
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		userInfo, err := client.GetUserInfo(context.Background(), "access-token")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if userInfo.Sub != "user-123" {
			t.Errorf("expected sub user-123, got %s", userInfo.Sub)
		}
		if userInfo.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", userInfo.Email)
		}
		if !userInfo.EmailVerified {
			t.Error("expected email to be verified")
		}

		// Verify Bearer auth
		if capturedAuth != "Bearer access-token" {
			t.Errorf("expected Bearer access-token, got %s", capturedAuth)
		}
	})

	t.Run("ShouldReturnError_WhenUnauthorized", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/userinfo", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		userInfo, err := client.GetUserInfo(context.Background(), "invalid-token")

		// Assert
		if err == nil {
			t.Error("expected error when unauthorized")
		}
		if userInfo != nil {
			t.Error("expected nil userinfo when unauthorized")
		}
	})
}

// TestRequestDeviceCode tests device authorization flow
func TestRequestDeviceCode(t *testing.T) {
	t.Run("ShouldRequestDeviceCodeSuccessfully", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/device", func(w http.ResponseWriter, r *http.Request) {
			response := models.DeviceAuthResponse{
				DeviceCode:              "device-code-123",
				UserCode:                "ABCD-1234",
				VerificationURI:         "https://auth.example.com/device",
				VerificationURIComplete: "https://auth.example.com/device?user_code=ABCD-1234",
				ExpiresIn:               1800,
				Interval:                5,
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
			Scopes:   []string{"openid", "profile"},
		})

		// Act
		result, err := client.RequestDeviceCode(context.Background(), nil)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.DeviceCode != "device-code-123" {
			t.Errorf("expected device code device-code-123, got %s", result.DeviceCode)
		}
		if result.UserCode != "ABCD-1234" {
			t.Errorf("expected user code ABCD-1234, got %s", result.UserCode)
		}
		if result.ExpiresIn != 1800 {
			t.Errorf("expected expires_in 1800, got %d", result.ExpiresIn)
		}
		if result.Interval != 5 {
			t.Errorf("expected interval 5, got %d", result.Interval)
		}
	})

	t.Run("ShouldUseCustomScopes", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedBody string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/device", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			capturedBody = r.Form.Encode()
			response := models.DeviceAuthResponse{
				DeviceCode: "device-code",
				UserCode:   "USER-CODE",
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
			Scopes:   []string{"openid"},
		})

		// Act
		_, err := client.RequestDeviceCode(context.Background(), []string{"openid", "offline_access"})

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// URL encoding uses + for space
		if !strings.Contains(capturedBody, "scope=openid+offline_access") && !strings.Contains(capturedBody, "scope=openid%20offline_access") {
			t.Errorf("expected custom scopes in request, got: %s", capturedBody)
		}
	})
}

// TestPollDeviceToken tests device token polling
func TestPollDeviceToken(t *testing.T) {
	t.Run("ShouldReturnTokenWhenAuthorized", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			response := models.OAuthTokenResponse{
				AccessToken:  "device-access-token",
				TokenType:    "Bearer",
				RefreshToken: "device-refresh-token",
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		result, err := client.PollDeviceToken(context.Background(), "device-code")

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != "device-access-token" {
			t.Errorf("expected device-access-token, got %s", result.AccessToken)
		}
	})

	t.Run("ShouldReturnErrAuthorizationPending", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.OAuthError{
				Error:            "authorization_pending",
				ErrorDescription: "user has not yet authorized the device",
			})
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		result, err := client.PollDeviceToken(context.Background(), "device-code")

		// Assert
		if err != ErrAuthorizationPending {
			t.Errorf("expected ErrAuthorizationPending, got %v", err)
		}
		if result != nil {
			t.Error("expected nil result when pending")
		}
	})

	t.Run("ShouldReturnErrSlowDown", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.OAuthError{
				Error: "slow_down",
			})
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		result, err := client.PollDeviceToken(context.Background(), "device-code")

		// Assert
		if err != ErrSlowDown {
			t.Errorf("expected ErrSlowDown, got %v", err)
		}
		if result != nil {
			t.Error("expected nil result when slow_down")
		}
	})

	t.Run("ShouldReturnErrAccessDenied", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.OAuthError{
				Error: "access_denied",
			})
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		result, err := client.PollDeviceToken(context.Background(), "device-code")

		// Assert
		if err != ErrAccessDenied {
			t.Errorf("expected ErrAccessDenied, got %v", err)
		}
		if result != nil {
			t.Error("expected nil result when access_denied")
		}
	})

	t.Run("ShouldReturnErrExpiredToken", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.OAuthError{
				Error: "expired_token",
			})
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		// Act
		result, err := client.PollDeviceToken(context.Background(), "device-code")

		// Assert
		if err != ErrExpiredToken {
			t.Errorf("expected ErrExpiredToken, got %v", err)
		}
		if result != nil {
			t.Error("expected nil result when expired_token")
		}
	})
}

// TestClientCredentialsGrant tests client credentials flow
func TestClientCredentialsGrant(t *testing.T) {
	t.Run("ShouldObtainTokenSuccessfully", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedAuth string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			response := models.OAuthTokenResponse{
				AccessToken: "client-credentials-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "api:read api:write",
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       serverURL,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		})

		// Act
		result, err := client.ClientCredentialsGrant(context.Background(), []string{"api:read", "api:write"})

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != "client-credentials-token" {
			t.Errorf("expected client-credentials-token, got %s", result.AccessToken)
		}
		if result.Scope != "api:read api:write" {
			t.Errorf("expected scopes, got %s", result.Scope)
		}

		// Verify Basic auth
		if !strings.HasPrefix(capturedAuth, "Basic ") {
			t.Error("expected Basic authentication")
		}
	})

	t.Run("ShouldReturnError_WhenClientSecretMissing", func(t *testing.T) {
		// Arrange
		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   "https://auth.example.com",
			ClientID: "test-client",
			// No client secret
		})

		// Act
		result, err := client.ClientCredentialsGrant(context.Background(), nil)

		// Assert
		if err == nil {
			t.Error("expected error when client secret is missing")
		}
		if result != nil {
			t.Error("expected nil result when error")
		}
		if !strings.Contains(err.Error(), "client_credentials grant requires client_secret") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("ShouldWorkWithoutScopes", func(t *testing.T) {
		// Arrange
		var serverURL string
		var capturedBody string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			capturedBody = r.Form.Encode()
			response := models.OAuthTokenResponse{
				AccessToken: "token",
				TokenType:   "Bearer",
			}
			json.NewEncoder(w).Encode(response)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       serverURL,
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		})

		// Act
		_, err := client.ClientCredentialsGrant(context.Background(), nil)

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(capturedBody, "scope=") {
			t.Error("should not include scope when nil/empty")
		}
	})
}

// TestSetClientAuth tests the client authentication helper
func TestSetClientAuth(t *testing.T) {
	t.Run("ShouldSetBasicAuthHeader_WhenClientSecretProvided", func(t *testing.T) {
		// Arrange
		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:       "https://auth.example.com",
			ClientID:     "test-client",
			ClientSecret: "test-secret",
		})

		req, _ := http.NewRequest("POST", "https://example.com/token", nil)

		// Act
		client.setClientAuth(req)

		// Assert
		authHeader := req.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Basic ") {
			t.Error("expected Basic auth header")
		}

		// Decode and verify
		encoded := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Fatalf("failed to decode Basic auth: %v", err)
		}
		if string(decoded) != "test-client:test-secret" {
			t.Errorf("expected test-client:test-secret, got %s", string(decoded))
		}
	})

	t.Run("ShouldNotSetHeader_WhenNoClientSecret", func(t *testing.T) {
		// Arrange
		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   "https://auth.example.com",
			ClientID: "test-client",
			// No client secret
		})

		req, _ := http.NewRequest("POST", "https://example.com/token", nil)

		// Act
		client.setClientAuth(req)

		// Assert
		authHeader := req.Header.Get("Authorization")
		if authHeader != "" {
			t.Error("should not set auth header when no client secret")
		}
	})
}

// TestContextCancellation tests that context cancellation is respected
func TestContextCancellation(t *testing.T) {
	t.Run("ShouldRespectContextCancellation", func(t *testing.T) {
		// Arrange
		var serverURL string
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			discovery := discoveryResponse(serverURL)
			json.NewEncoder(w).Encode(discovery)
		})
		server := newTestServer(t, mux)
		serverURL = server.URL

		client := NewOAuthProviderClient(OAuthProviderConfig{
			Issuer:   serverURL,
			ClientID: "test-client",
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Act
		_, err := client.GetDiscovery(ctx)

		// Assert
		if err == nil {
			t.Error("expected error when context is cancelled")
		}
	})
}

// TestBase64URLEncoding tests Base64URL encoding used in PKCE
func TestBase64URLEncoding(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "Simple input",
			input:    []byte("test"),
			expected: base64.RawURLEncoding.EncodeToString([]byte("test")),
		},
		{
			name:     "Input with special chars",
			input:    []byte{0xFF, 0xFE, 0xFD},
			expected: base64.RawURLEncoding.EncodeToString([]byte{0xFF, 0xFE, 0xFD}),
		},
		{
			name:     "SHA256 hash simulation",
			input:    sha256.New().Sum([]byte("verifier")),
			expected: base64.RawURLEncoding.EncodeToString(sha256.New().Sum([]byte("verifier"))),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := base64.RawURLEncoding.EncodeToString(tc.input)

			// Assert
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}

			// Verify no padding
			if strings.Contains(result, "=") {
				t.Error("Base64URL should not contain padding")
			}

			// Verify URL-safe
			if strings.ContainsAny(result, "+/") {
				t.Error("Base64URL should not contain + or /")
			}
		})
	}
}

// TestPKCECodeChallengeVerification tests that code challenge can be verified
func TestPKCECodeChallengeVerification(t *testing.T) {
	t.Run("ShouldVerifyCodeChallengeMatchesVerifier", func(t *testing.T) {
		// Arrange
		pkce, _ := GeneratePKCE()

		// Act - simulate server-side verification
		hash := sha256.Sum256([]byte(pkce.CodeVerifier))
		computedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

		// Assert
		if computedChallenge != pkce.CodeChallenge {
			t.Error("code challenge verification failed")
		}
	})

	t.Run("ShouldFailVerificationWithWrongVerifier", func(t *testing.T) {
		// Arrange
		pkce, _ := GeneratePKCE()

		// Act - simulate server-side verification with wrong verifier
		hash := sha256.Sum256([]byte("wrong-verifier"))
		computedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

		// Assert
		if computedChallenge == pkce.CodeChallenge {
			t.Error("verification should fail with wrong verifier")
		}
	})
}
