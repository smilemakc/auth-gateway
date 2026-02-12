package authgateway

import (
	"context"

	"github.com/smilemakc/auth-gateway/packages/go-sdk/models"
)

// OTPService handles OTP (One-Time Password) operations.
type OTPService struct {
	client *Client
}

// Send sends an OTP to the specified email or phone.
func (s *OTPService) Send(ctx context.Context, req *models.SendOTPRequest) (*models.OTPResponse, error) {
	var resp models.OTPResponse
	if err := s.client.post(ctx, "/api/otp/send", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendToEmail is a convenience method to send OTP via email.
func (s *OTPService) SendToEmail(ctx context.Context, email, otpType string) (*models.OTPResponse, error) {
	return s.Send(ctx, &models.SendOTPRequest{
		Email: &email,
		Type:  otpType,
	})
}

// SendToPhone is a convenience method to send OTP via SMS.
func (s *OTPService) SendToPhone(ctx context.Context, phone, otpType string) (*models.OTPResponse, error) {
	return s.Send(ctx, &models.SendOTPRequest{
		Phone: &phone,
		Type:  otpType,
	})
}

// Verify verifies an OTP code.
func (s *OTPService) Verify(ctx context.Context, req *models.VerifyOTPRequest) (*models.MessageResponse, error) {
	var resp models.MessageResponse
	if err := s.client.post(ctx, "/api/otp/verify", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// VerifyEmail is a convenience method to verify email OTP.
func (s *OTPService) VerifyEmail(ctx context.Context, email, code string) (*models.MessageResponse, error) {
	return s.Verify(ctx, &models.VerifyOTPRequest{
		Email: &email,
		Code:  code,
	})
}

// VerifyPhone is a convenience method to verify phone OTP.
func (s *OTPService) VerifyPhone(ctx context.Context, phone, code string) (*models.MessageResponse, error) {
	return s.Verify(ctx, &models.VerifyOTPRequest{
		Phone: &phone,
		Code:  code,
	})
}

// OTP types
const (
	OTPTypeVerification  = "verification"
	OTPTypePasswordReset = "password_reset"
	OTPTypePasswordless  = "passwordless"
	OTPType2FA           = "2fa"
)
