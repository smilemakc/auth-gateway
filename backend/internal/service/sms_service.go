package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/sms"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

const (
	SMSOTPLength     = 6
	SMSOTPExpiration = 10 * time.Minute
	SMSRateLimit     = 3 // Max 3 SMS per hour for same phone/type
)

// SMSService handles SMS-related operations
type SMSService struct {
	provider        sms.SMSProvider
	otpRepo         *repository.OTPRepository
	smsLogRepo      *repository.SMSLogRepository
	smsSettingsRepo *repository.SMSSettingsRepository
	userRepo        *repository.UserRepository
	config          *config.Config
	redis           *RedisService
}

// NewSMSService creates a new SMS service
func NewSMSService(
	provider sms.SMSProvider,
	otpRepo *repository.OTPRepository,
	smsLogRepo *repository.SMSLogRepository,
	smsSettingsRepo *repository.SMSSettingsRepository,
	userRepo *repository.UserRepository,
	cfg *config.Config,
	redis *RedisService,
) *SMSService {
	return &SMSService{
		provider:        provider,
		otpRepo:         otpRepo,
		smsLogRepo:      smsLogRepo,
		smsSettingsRepo: smsSettingsRepo,
		userRepo:        userRepo,
		config:          cfg,
		redis:           redis,
	}
}

// GenerateOTPCode generates a 6-digit OTP code
func (s *SMSService) GenerateOTPCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	code := fmt.Sprintf("%06d", n.Int64())
	return code, nil
}

// SendOTP sends an OTP code via SMS
func (s *SMSService) SendOTP(ctx context.Context, req *models.SendSMSRequest, ipAddress string) (*models.SendSMSResponse, error) {
	// Validate phone number
	if req.Phone == "" {
		return nil, models.NewAppError(400, "Phone number is required")
	}

	// Normalize phone number
	phone := utils.NormalizePhone(req.Phone)
	if !utils.IsValidPhone(phone) {
		return nil, models.NewAppError(400, "Invalid phone number format")
	}

	// Check rate limiting - max 3 SMS per hour for same phone/type
	count, err := s.otpRepo.CountRecentByPhone(phone, req.Type, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if count >= SMSRateLimit {
		return nil, models.NewAppError(429, "Too many SMS requests. Please try again later.")
	}

	// Check SMS-specific rate limits from settings
	if err := s.checkSMSRateLimits(ctx, phone); err != nil {
		return nil, err
	}

	// Generate OTP code
	plainCode, err := s.GenerateOTPCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP code: %w", err)
	}

	// Hash the code before storing
	codeHash, err := utils.HashPassword(plainCode, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to hash OTP code: %w", err)
	}

	// Invalidate all previous OTPs for this phone/type
	if err := s.otpRepo.InvalidateAllForPhone(phone, req.Type); err != nil {
		return nil, fmt.Errorf("failed to invalidate previous OTPs: %w", err)
	}

	// Create OTP record
	otpID := uuid.New()
	expiresAt := time.Now().Add(SMSOTPExpiration)
	otp := &models.OTP{
		ID:        otpID,
		Phone:     &phone,
		Code:      codeHash,
		Type:      req.Type,
		Used:      false,
		ExpiresAt: expiresAt,
	}

	if err := s.otpRepo.Create(otp); err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Format SMS message
	message := s.formatOTPMessage(plainCode, req.Type)

	// Create SMS log entry
	smsLog := &models.SMSLog{
		ID:        uuid.New(),
		Phone:     phone,
		Message:   message,
		Type:      req.Type,
		Provider:  s.provider.GetProviderName(),
		Status:    models.SMSStatusPending,
		CreatedAt: time.Now(),
		IPAddress: &ipAddress,
	}

	// Try to get user ID if exists
	if user, err := s.userRepo.GetByPhone(phone); err == nil {
		smsLog.UserID = &user.ID
	}

	if err := s.smsLogRepo.Create(ctx, smsLog); err != nil {
		return nil, fmt.Errorf("failed to create SMS log: %w", err)
	}

	// Send SMS via provider
	messageID, err := s.provider.SendSMS(ctx, phone, message)
	if err != nil {
		// Update log with failure
		errMsg := err.Error()
		_ = s.smsLogRepo.UpdateStatus(ctx, smsLog.ID, models.SMSStatusFailed, &errMsg)
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	// Update log with success
	_ = s.smsLogRepo.UpdateStatus(ctx, smsLog.ID, models.SMSStatusSent, nil)

	return &models.SendSMSResponse{
		Success:   true,
		MessageID: &messageID,
		ExpiresAt: expiresAt,
	}, nil
}

// VerifyOTP verifies an OTP code sent via SMS
func (s *SMSService) VerifyOTP(ctx context.Context, req *models.VerifySMSOTPRequest) (*models.VerifySMSOTPResponse, error) {
	// Validate phone number
	if req.Phone == "" {
		return nil, models.NewAppError(400, "Phone number is required")
	}

	// Normalize phone number
	phone := utils.NormalizePhone(req.Phone)
	if !utils.IsValidPhone(phone) {
		return nil, models.NewAppError(400, "Invalid phone number format")
	}

	// Get latest valid OTP
	otp, err := s.otpRepo.GetByPhoneAndType(phone, req.Type)
	if err != nil {
		return &models.VerifySMSOTPResponse{
			Valid:   false,
			Message: "Invalid or expired OTP code",
		}, nil
	}

	// Check expiration
	if otp.IsExpired() {
		return &models.VerifySMSOTPResponse{
			Valid:   false,
			Message: "OTP code has expired",
		}, nil
	}

	// Verify code using bcrypt
	if err := utils.CheckPassword(otp.Code, req.Code); err != nil {
		return &models.VerifySMSOTPResponse{
			Valid:   false,
			Message: "Invalid OTP code",
		}, nil
	}

	// Mark OTP as used
	if err := s.otpRepo.MarkAsUsed(otp.ID); err != nil {
		return nil, fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	response := &models.VerifySMSOTPResponse{
		Valid:   true,
		Message: "OTP verified successfully",
	}

	// Handle by type
	switch req.Type {
	case models.OTPTypeVerification:
		// Mark phone as verified
		user, err := s.userRepo.GetByPhone(phone)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		if err := s.userRepo.MarkPhoneVerified(user.ID); err != nil {
			return nil, fmt.Errorf("failed to mark phone verified: %w", err)
		}
		response.User = user

	case models.OTPType2FA:
		// 2FA verification handled in auth flow
		user, err := s.userRepo.GetByPhone(phone)
		if err == nil {
			response.User = user
		}

	case models.OTPTypePasswordReset:
		// Password reset flow handled separately
		user, err := s.userRepo.GetByPhone(phone)
		if err == nil {
			response.User = user
		}

	case models.OTPTypeLogin:
		// Passwordless login
		user, err := s.userRepo.GetByPhone(phone)
		if err == nil {
			response.User = user
		}
	}

	return response, nil
}

// formatOTPMessage formats the SMS message for different OTP types
func (s *SMSService) formatOTPMessage(code string, otpType models.OTPType) string {
	switch otpType {
	case models.OTPTypeVerification:
		return fmt.Sprintf("Your verification code is: %s\n\nThis code will expire in 10 minutes.\n\nAuth Gateway", code)
	case models.OTPTypePasswordReset:
		return fmt.Sprintf("Your password reset code is: %s\n\nThis code will expire in 10 minutes.\n\nAuth Gateway", code)
	case models.OTPType2FA:
		return fmt.Sprintf("Your 2FA code is: %s\n\nThis code will expire in 10 minutes.\n\nAuth Gateway", code)
	case models.OTPTypeLogin:
		return fmt.Sprintf("Your login code is: %s\n\nThis code will expire in 10 minutes.\n\nAuth Gateway", code)
	default:
		return fmt.Sprintf("Your verification code is: %s\n\nThis code will expire in 10 minutes.", code)
	}
}

// checkSMSRateLimits checks SMS-specific rate limits from settings
func (s *SMSService) checkSMSRateLimits(ctx context.Context, phone string) error {
	// Check per-number limit (hourly)
	key := fmt.Sprintf("sms:limit:phone:%s:hour", phone)
	count, err := s.redis.IncrementRateLimit(ctx, key, 1*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check phone rate limit: %w", err)
	}

	if count > int64(s.config.SMS.SMSMaxPerNumber) {
		return models.NewAppError(429, "Too many SMS sent to this number. Please try again later.")
	}

	// Check daily limit for the phone number
	dailyKey := fmt.Sprintf("sms:limit:phone:%s:day", phone)
	dailyCount, err := s.redis.IncrementRateLimit(ctx, dailyKey, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check daily rate limit: %w", err)
	}

	if dailyCount > int64(s.config.SMS.SMSMaxPerDay) {
		return models.NewAppError(429, "Daily SMS limit reached for this number.")
	}

	// Check global hourly limit
	globalKey := "sms:limit:global:hour"
	globalCount, err := s.redis.IncrementRateLimit(ctx, globalKey, 1*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check global rate limit: %w", err)
	}

	if globalCount > int64(s.config.SMS.SMSMaxPerHour) {
		return models.NewAppError(429, "System SMS limit reached. Please try again later.")
	}

	return nil
}

// GetStats retrieves SMS statistics
func (s *SMSService) GetStats(ctx context.Context) (*models.SMSStatsResponse, error) {
	return s.smsLogRepo.GetStats(ctx)
}

// CleanupOldLogs deletes SMS logs older than the specified duration
func (s *SMSService) CleanupOldLogs(ctx context.Context, duration time.Duration) (int64, error) {
	return s.smsLogRepo.DeleteOlderThan(ctx, duration)
}
