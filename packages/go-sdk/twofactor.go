package authgateway

import (
	"context"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// TwoFactorService handles two-factor authentication operations.
type TwoFactorService struct {
	client *Client
}

// Setup initiates 2FA setup and returns the secret and QR code.
func (s *TwoFactorService) Setup(ctx context.Context) (*models.TwoFASetupResponse, error) {
	var resp models.TwoFASetupResponse
	if err := s.client.post(ctx, "/api/auth/2fa/setup", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Verify verifies the TOTP code and enables 2FA.
func (s *TwoFactorService) Verify(ctx context.Context, code string) (*models.MessageResponse, error) {
	req := &models.VerifyTwoFARequest{Code: code}

	var resp models.MessageResponse
	if err := s.client.post(ctx, "/api/auth/2fa/verify", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Disable disables 2FA for the current user.
func (s *TwoFactorService) Disable(ctx context.Context, password string) (*models.MessageResponse, error) {
	req := &models.DisableTwoFARequest{Password: password}

	var resp models.MessageResponse
	if err := s.client.post(ctx, "/api/auth/2fa/disable", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Status retrieves the current 2FA status.
func (s *TwoFactorService) Status(ctx context.Context) (*models.TwoFAStatusResponse, error) {
	var resp models.TwoFAStatusResponse
	if err := s.client.get(ctx, "/api/auth/2fa/status", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RegenerateBackupCodes generates new backup codes.
func (s *TwoFactorService) RegenerateBackupCodes(ctx context.Context) (*models.BackupCodesResponse, error) {
	var resp models.BackupCodesResponse
	if err := s.client.post(ctx, "/api/auth/2fa/backup-codes/regenerate", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
