package authgateway

import (
	"context"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// AuthService handles authentication operations.
type AuthService struct {
	client *Client
}

// SignUp registers a new user account.
func (s *AuthService) SignUp(ctx context.Context, req *models.SignUpRequest) (*models.AuthResponse, error) {
	var resp models.AuthResponse
	if err := s.client.post(ctx, "/api/auth/signup", req, &resp); err != nil {
		return nil, err
	}

	// Store tokens if received
	if resp.AccessToken != "" {
		s.client.SetTokens(resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)
	}

	return &resp, nil
}

// SignIn authenticates a user with email/phone and password.
// Returns AuthResponse with tokens or TwoFactorRequiredError if 2FA is required.
func (s *AuthService) SignIn(ctx context.Context, req *models.SignInRequest) (*models.AuthResponse, error) {
	var resp models.AuthResponse
	if err := s.client.post(ctx, "/api/auth/signin", req, &resp); err != nil {
		return nil, err
	}

	// If 2FA is required, don't store tokens yet
	if resp.Requires2FA {
		return &resp, &TwoFactorRequiredError{TwoFactorToken: resp.TwoFactorToken}
	}

	// Store tokens
	if resp.AccessToken != "" {
		s.client.SetTokens(resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)
	}

	return &resp, nil
}

// SignInWithEmail is a convenience method for email/password login.
func (s *AuthService) SignInWithEmail(ctx context.Context, email, password string) (*models.AuthResponse, error) {
	return s.SignIn(ctx, &models.SignInRequest{
		Email:    email,
		Password: password,
	})
}

// SignInWithPhone is a convenience method for phone/password login.
func (s *AuthService) SignInWithPhone(ctx context.Context, phone, password string) (*models.AuthResponse, error) {
	return s.SignIn(ctx, &models.SignInRequest{
		Phone:    &phone,
		Password: password,
	})
}

// Verify2FA completes login with 2FA verification.
func (s *AuthService) Verify2FA(ctx context.Context, twoFactorToken, code string) (*models.AuthResponse, error) {
	req := &models.TwoFactorLoginVerifyRequest{
		TwoFactorToken: twoFactorToken,
		Code:           code,
	}

	var resp models.AuthResponse
	if err := s.client.post(ctx, "/api/auth/2fa/login/verify", req, &resp); err != nil {
		return nil, err
	}

	// Store tokens
	if resp.AccessToken != "" {
		s.client.SetTokens(resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)
	}

	return &resp, nil
}

// RefreshTokens refreshes the access token using the refresh token.
func (s *AuthService) RefreshTokens(ctx context.Context) error {
	refreshToken := s.client.GetRefreshToken()
	if refreshToken == "" {
		return &AuthenticationError{Message: "no refresh token available"}
	}

	req := &models.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	var resp models.TokenResponse
	if err := s.client.post(ctx, "/api/auth/refresh", req, &resp); err != nil {
		return err
	}

	// Update tokens
	s.client.SetTokens(resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	return nil
}

// Logout logs out the current user and invalidates tokens.
func (s *AuthService) Logout(ctx context.Context) error {
	err := s.client.post(ctx, "/api/auth/logout", nil, nil)
	// Clear tokens regardless of server response
	s.client.ClearTokens()
	return err
}

// ResendVerificationEmail resends the email verification code.
func (s *AuthService) ResendVerificationEmail(ctx context.Context) (*models.MessageResponse, error) {
	var resp models.MessageResponse
	if err := s.client.post(ctx, "/api/auth/verify/resend", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// VerifyEmail verifies the user's email with an OTP code.
func (s *AuthService) VerifyEmail(ctx context.Context, code string) (*models.MessageResponse, error) {
	req := &models.VerifyEmailRequest{Code: code}

	var resp models.MessageResponse
	if err := s.client.post(ctx, "/api/auth/verify/email", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RequestPasswordReset initiates the password reset process.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (*models.MessageResponse, error) {
	req := &models.ForgotPasswordRequest{Email: email}

	var resp models.MessageResponse
	if err := s.client.post(ctx, "/api/auth/password/reset/request", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ResetPassword completes the password reset with OTP verification.
func (s *AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) (*models.MessageResponse, error) {
	req := &models.ResetPasswordRequest{
		Email:       email,
		Code:        code,
		NewPassword: newPassword,
	}

	var resp models.MessageResponse
	if err := s.client.post(ctx, "/api/auth/password/reset/complete", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
