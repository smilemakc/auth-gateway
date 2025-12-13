package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

const (
	OTPLength     = 6
	OTPExpiration = 10 * time.Minute
	OTPRateLimit  = 3
)

type OTPService struct {
	otpRepo      *repository.OTPRepository
	userRepo     *repository.UserRepository
	emailService *EmailService
	auditService *AuditService
}

func NewOTPService(
	otpRepo *repository.OTPRepository,
	userRepo *repository.UserRepository,
	emailService *EmailService,
	auditService *AuditService,
) *OTPService {
	return &OTPService{
		otpRepo:      otpRepo,
		userRepo:     userRepo,
		emailService: emailService,
		auditService: auditService,
	}
}

// GenerateOTPCode generates a random 6-digit OTP code
func (s *OTPService) GenerateOTPCode() (string, error) {
	// Generate 6-digit code (000000-999999)
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	// Format as 6 digits with leading zeros
	code := fmt.Sprintf("%06d", n.Int64())
	return code, nil
}

// SendOTP generates and sends an OTP code
func (s *OTPService) SendOTP(ctx context.Context, req *models.SendOTPRequest) error {
	// Validate that either email or phone is provided
	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		return models.NewAppError(400, "Either email or phone is required")
	}

	// Only email OTP is supported in this service
	// Phone OTP should use SMSService
	if req.Email == nil || *req.Email == "" {
		return models.NewAppError(400, "Email is required for email OTP")
	}

	email := *req.Email

	// Check rate limiting
	count, err := s.otpRepo.CountRecentByEmail(ctx, email, req.Type, 1*time.Hour)
	if err != nil {
		return err
	}

	if count >= OTPRateLimit {
		return models.NewAppError(429, "Too many OTP requests. Please try again later.")
	}

	// Generate OTP code
	plainCode, err := s.GenerateOTPCode()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Hash the code
	codeHash, err := utils.HashPassword(plainCode, 10)
	if err != nil {
		return fmt.Errorf("failed to hash OTP code: %w", err)
	}

	// Invalidate all previous OTPs for this email and type
	if err := s.otpRepo.InvalidateAllForEmail(ctx, email, req.Type); err != nil {
		return err
	}

	// Create OTP record
	otp := &models.OTP{
		ID:        uuid.New(),
		Email:     &email,
		Code:      codeHash,
		Type:      req.Type,
		Used:      false,
		ExpiresAt: time.Now().Add(OTPExpiration),
	}

	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return err
	}

	// Send email
	if err := s.emailService.SendOTP(email, plainCode, string(req.Type)); err != nil {
		// Log error but don't fail - OTP is already created
		fmt.Printf("Failed to send OTP email: %v\n", err)
	}

	// Audit log
	s.logAudit(nil, "otp_sent", "success", "", "", map[string]interface{}{
		"email": email,
		"type":  req.Type,
	})

	return nil
}

// VerifyOTP verifies an OTP code
func (s *OTPService) VerifyOTP(ctx context.Context, req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
	// Validate that either email or phone is provided
	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		return nil, models.NewAppError(400, "Either email or phone is required")
	}

	// Only email OTP is supported in this service
	if req.Email == nil || *req.Email == "" {
		return nil, models.NewAppError(400, "Email is required for email OTP verification")
	}

	email := *req.Email

	// Get the latest valid OTP
	otp, err := s.otpRepo.GetByEmailAndType(ctx, email, req.Type)
	if err != nil {
		s.logAudit(nil, "otp_verify", "failed", "", "", map[string]interface{}{
			"email":  email,
			"type":   req.Type,
			"reason": "otp_not_found",
		})
		return &models.VerifyOTPResponse{Valid: false}, nil
	}

	// Check if expired
	if otp.IsExpired() {
		s.logAudit(nil, "otp_verify", "failed", "", "", map[string]interface{}{
			"email":  email,
			"type":   req.Type,
			"reason": "expired",
		})
		return &models.VerifyOTPResponse{Valid: false}, nil
	}

	// Verify code
	if err := utils.CheckPassword(otp.Code, req.Code); err != nil {
		s.logAudit(nil, "otp_verify", "failed", "", "", map[string]interface{}{
			"email":  email,
			"type":   req.Type,
			"reason": "invalid_code",
		})
		return &models.VerifyOTPResponse{Valid: false}, nil
	}

	// Mark as used
	if err := s.otpRepo.MarkAsUsed(ctx, otp.ID); err != nil {
		return nil, err
	}

	// Handle verification type
	response := &models.VerifyOTPResponse{Valid: true}

	switch req.Type {
	case models.OTPTypeVerification:
		// Mark email as verified
		user, err := s.userRepo.GetByEmail(ctx, email, utils.Ptr(true))
		if err == nil && user != nil {
			if err := s.userRepo.MarkEmailVerified(ctx, user.ID); err != nil {
				return nil, err
			}
			response.User = user
		}

	case models.OTPTypeLogin:
		// For passwordless login, create session
		user, err := s.userRepo.GetByEmail(ctx, email, utils.Ptr(true))
		if err != nil {
			return nil, err
		}
		response.User = user
		// Tokens will be generated by the handler

	case models.OTPTypePasswordReset:
		// Password reset will be handled separately
		// Just verify the OTP is valid
		break

	case models.OTPType2FA:
		// 2FA verification
		break
	}

	s.logAudit(nil, "otp_verify", "success", "", "", map[string]interface{}{
		"email": email,
		"type":  req.Type,
	})

	return response, nil
}

// CleanupExpiredOTPs removes expired OTPs
func (s *OTPService) CleanupExpiredOTPs() error {
	// Delete OTPs older than 7 days
	return s.otpRepo.DeleteExpired(context.Background(), 7*24*time.Hour)
}

func (s *OTPService) logAudit(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
	s.auditService.LogWithAction(userID, action, status, ip, userAgent, details)
}
