package authgateway

import (
	"context"
	"fmt"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// OAuthService handles OAuth/social login operations.
type OAuthService struct {
	client *Client
}

// OAuth provider names
const (
	OAuthProviderGoogle    = "google"
	OAuthProviderGitHub    = "github"
	OAuthProviderYandex    = "yandex"
	OAuthProviderInstagram = "instagram"
	OAuthProviderTelegram  = "telegram"
)

// GetProviders retrieves the list of enabled OAuth providers.
func (s *OAuthService) GetProviders(ctx context.Context) ([]models.OAuthProvider, error) {
	var resp []models.OAuthProvider
	if err := s.client.get(ctx, "/api/auth/providers", &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetAuthURL returns the OAuth authorization URL for a provider.
// Note: This URL should be opened in a browser for the user to authenticate.
func (s *OAuthService) GetAuthURL(provider string) string {
	return fmt.Sprintf("%s/api/auth/%s", s.client.baseURL, provider)
}

// HandleCallback processes the OAuth callback (typically used in server-side scenarios).
// For client-side applications, the callback is handled automatically by the server.
func (s *OAuthService) HandleCallback(ctx context.Context, provider, code, state string) (*models.AuthResponse, error) {
	// This is typically handled server-side via redirect
	// This method is provided for custom callback handling scenarios
	path := fmt.Sprintf("/api/auth/%s/callback?code=%s&state=%s", provider, code, state)

	var resp models.AuthResponse
	if err := s.client.get(ctx, path, &resp); err != nil {
		return nil, err
	}

	// Store tokens if received
	if resp.AccessToken != "" {
		s.client.SetTokens(resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)
	}

	return &resp, nil
}

// PasswordlessService handles passwordless login operations.
type PasswordlessService struct {
	client *Client
}

// Request initiates passwordless login by sending an OTP.
func (s *PasswordlessService) Request(ctx context.Context, req *models.PasswordlessRequest) (*models.OTPResponse, error) {
	var resp models.OTPResponse
	if err := s.client.post(ctx, "/api/auth/passwordless/request", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RequestWithEmail is a convenience method for email-based passwordless login.
func (s *PasswordlessService) RequestWithEmail(ctx context.Context, email string) (*models.OTPResponse, error) {
	return s.Request(ctx, &models.PasswordlessRequest{Email: &email})
}

// RequestWithPhone is a convenience method for phone-based passwordless login.
func (s *PasswordlessService) RequestWithPhone(ctx context.Context, phone string) (*models.OTPResponse, error) {
	return s.Request(ctx, &models.PasswordlessRequest{Phone: &phone})
}

// Verify completes passwordless login with OTP verification.
func (s *PasswordlessService) Verify(ctx context.Context, req *models.PasswordlessVerifyRequest) (*models.AuthResponse, error) {
	var resp models.AuthResponse
	if err := s.client.post(ctx, "/api/auth/passwordless/verify", req, &resp); err != nil {
		return nil, err
	}

	// Store tokens
	if resp.AccessToken != "" {
		s.client.SetTokens(resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)
	}

	return &resp, nil
}

// VerifyWithEmail is a convenience method to verify email-based passwordless login.
func (s *PasswordlessService) VerifyWithEmail(ctx context.Context, email, code string) (*models.AuthResponse, error) {
	return s.Verify(ctx, &models.PasswordlessVerifyRequest{
		Email: &email,
		Code:  code,
	})
}

// VerifyWithPhone is a convenience method to verify phone-based passwordless login.
func (s *PasswordlessService) VerifyWithPhone(ctx context.Context, phone, code string) (*models.AuthResponse, error) {
	return s.Verify(ctx, &models.PasswordlessVerifyRequest{
		Phone: &phone,
		Code:  code,
	})
}
