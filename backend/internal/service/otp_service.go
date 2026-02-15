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
	"github.com/smilemakc/auth-gateway/internal/sms"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

const (
	OTPLength     = 6
	OTPExpiration = 10 * time.Minute
	OTPRateLimit  = 3
)

// OTPSendChannel identifies how the OTP is delivered.
type OTPSendChannel string

const (
	OTPChannelEmail OTPSendChannel = "email"
	OTPChannelSMS   OTPSendChannel = "sms"
)

// OTPServiceOptions contains optional dependencies for OTP delivery.
type OTPServiceOptions struct {
	EmailSender         EmailSender
	EmailProfileService EmailProfileSender
	SMSProvider         sms.SMSProvider
	SMSLogRepo          SMSLogStore
	Cache               CacheService
	Config              *config.Config
}

// EmailProfileSender defines the interface for profile-based email sending
type EmailProfileSender interface {
	SendOTPEmail(ctx context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, otpType models.OTPType, code string) error
}

type OTPService struct {
	otpRepo             OTPStore
	userRepo            UserStore
	emailService        EmailSender
	emailProfileService EmailProfileSender
	smsProvider         sms.SMSProvider
	smsLogRepo          SMSLogStore
	auditService        AuditLogger
	cache               CacheService
	cfg                 *config.Config
}

func NewOTPService(
	otpRepo OTPStore,
	userRepo UserStore,
	auditService AuditLogger,
	opts OTPServiceOptions,
) *OTPService {
	return &OTPService{
		otpRepo:             otpRepo,
		userRepo:            userRepo,
		emailService:        opts.EmailSender,
		emailProfileService: opts.EmailProfileService,
		smsProvider:         opts.SMSProvider,
		smsLogRepo:          opts.SMSLogRepo,
		auditService:        auditService,
		cache:               opts.Cache,
		cfg:                 opts.Config,
	}
}

// GenerateOTPCode generates a random 6-digit OTP code
func (s *OTPService) GenerateOTPCode() (string, error) {
	// Generate 6-digit code (000000-999999)
	m := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, m)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	// Format as 6 digits with leading zeros
	code := fmt.Sprintf("%06d", n.Int64())
	return code, nil
}

// SendOTP generates and sends an OTP code
func (s *OTPService) SendOTP(ctx context.Context, req *models.SendOTPRequest) error {
	if err := validateOTPType(req.Type); err != nil {
		return err
	}

	channel, destination, err := s.resolveDestination(req)
	if err != nil {
		return err
	}

	// Check rate limiting
	if err := s.checkRateLimit(ctx, channel, destination, req.Type); err != nil {
		return err
	}

	// Generate OTP code
	plainCode, err := s.GenerateOTPCode()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Hash the code using HMAC (secure against brute-force on 6-digit codes)
	codeHash := utils.HMACHash(plainCode, s.cfg.Security.OTPHMACSecret)

	otp := &models.OTP{
		ID:        uuid.New(),
		Code:      codeHash,
		Type:      req.Type,
		Used:      false,
		ExpiresAt: time.Now().Add(OTPExpiration),
	}

	switch channel {
	case OTPChannelEmail:
		otp.Email = utils.Ptr(destination)
		if err := s.otpRepo.InvalidateAllForEmail(ctx, destination, req.Type); err != nil {
			return err
		}
	case OTPChannelSMS:
		otp.Phone = utils.Ptr(destination)
		if err := s.otpRepo.InvalidateAllForPhone(ctx, destination, req.Type); err != nil {
			return err
		}
	default:
		return models.NewAppError(400, "Unsupported OTP delivery channel")
	}

	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return err
	}

	if err := s.dispatchOTP(ctx, channel, destination, plainCode, req.Type, req.ProfileID, req.ApplicationID); err != nil {
		return err
	}

	s.logAudit(nil, "otp_sent", "success", "", "", map[string]interface{}{
		"email": otp.Email,
		"phone": otp.Phone,
		"type":  req.Type,
	})

	return nil
}

// VerifyOTP verifies an OTP code
func (s *OTPService) VerifyOTP(ctx context.Context, req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
	if err := validateOTPType(req.Type); err != nil {
		return nil, err
	}

	channel, destination, err := s.resolveVerifyDestination(req)
	if err != nil {
		return nil, err
	}

	otp, err := s.fetchOTP(ctx, channel, destination, req.Type)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok && appErr.Code == 404 {
			return &models.VerifyOTPResponse{Valid: false}, nil
		}
		return nil, err
	}

	// Check if expired
	if otp.IsExpired() {
		s.logAudit(nil, "otp_verify", "failed", "", "", map[string]interface{}{
			"email":  otp.Email,
			"phone":  otp.Phone,
			"type":   req.Type,
			"reason": "expired",
		})
		return &models.VerifyOTPResponse{Valid: false}, nil
	}

	if otp.Used {
		s.logAudit(nil, "otp_verify", "failed", "", "", map[string]interface{}{
			"email":  otp.Email,
			"phone":  otp.Phone,
			"type":   req.Type,
			"reason": "already_used",
		})
		return &models.VerifyOTPResponse{Valid: false}, nil
	}

	// Verify code using HMAC comparison
	if !utils.HMACVerify(req.Code, s.cfg.Security.OTPHMACSecret, otp.Code) {
		s.logAudit(nil, "otp_verify", "failed", "", "", map[string]interface{}{
			"email":  otp.Email,
			"phone":  otp.Phone,
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
	switch channel {
	case OTPChannelEmail:
		if otp.Email != nil {
			s.enrichVerificationResponseEmail(ctx, response, *otp.Email, req.Type)
		}
	case OTPChannelSMS:
		if otp.Phone != nil {
			if err := s.enrichVerificationResponsePhone(ctx, response, *otp.Phone, req.Type); err != nil {
				return nil, err
			}
		}
	}

	s.logAudit(nil, "otp_verify", "success", "", "", map[string]interface{}{
		"email": otp.Email,
		"phone": otp.Phone,
		"type":  req.Type,
	})

	return response, nil
}

// CleanupExpiredOTPs removes expired OTPs
func (s *OTPService) CleanupExpiredOTPs() error {
	// Delete OTPs older than 7 days
	// Use context with timeout for cleanup operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return s.otpRepo.DeleteExpired(ctx, 7*24*time.Hour)
}

func (s *OTPService) logAudit(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
	s.auditService.LogWithAction(userID, action, status, ip, userAgent, details)
}

func (s *OTPService) resolveDestination(req *models.SendOTPRequest) (OTPSendChannel, string, error) {
	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		return "", "", models.NewAppError(400, "Either email or phone is required")
	}

	if req.Email != nil && *req.Email != "" {
		email := utils.NormalizeEmail(*req.Email)
		return OTPChannelEmail, email, nil
	}

	phone := utils.NormalizePhone(*req.Phone)
	if !utils.IsValidPhone(phone) {
		return "", "", models.NewAppError(400, "Invalid phone number format")
	}

	return OTPChannelSMS, phone, nil
}

func (s *OTPService) resolveVerifyDestination(req *models.VerifyOTPRequest) (OTPSendChannel, string, error) {
	if (req.Email == nil || *req.Email == "") && (req.Phone == nil || *req.Phone == "") {
		return "", "", models.NewAppError(400, "Either email or phone is required")
	}

	if req.Email != nil && *req.Email != "" {
		email := utils.NormalizeEmail(*req.Email)
		return OTPChannelEmail, email, nil
	}

	phone := utils.NormalizePhone(*req.Phone)
	if !utils.IsValidPhone(phone) {
		return "", "", models.NewAppError(400, "Invalid phone number format")
	}

	return OTPChannelSMS, phone, nil
}

func (s *OTPService) checkRateLimit(ctx context.Context, channel OTPSendChannel, destination string, otpType models.OTPType) error {
	switch channel {
	case OTPChannelEmail:
		count, err := s.otpRepo.CountRecentByEmail(ctx, destination, otpType, time.Hour)
		if err != nil {
			return err
		}
		if count >= OTPRateLimit {
			return models.NewAppError(429, "Too many OTP requests. Please try again later.")
		}
	case OTPChannelSMS:
		count, err := s.otpRepo.CountRecentByPhone(ctx, destination, otpType, time.Hour)
		if err != nil {
			return err
		}
		if count >= SMSRateLimit {
			return models.NewAppError(429, "Too many OTP requests. Please try again later.")
		}

		// Additional SMS-specific limits if configured
		if s.cfg != nil && s.cache != nil {
			if err := s.checkSMSRateLimits(ctx, destination); err != nil {
				return err
			}
		}
	default:
		return models.NewAppError(400, "Unsupported OTP delivery channel")
	}

	return nil
}

func (s *OTPService) dispatchOTP(ctx context.Context, channel OTPSendChannel, destination, code string, otpType models.OTPType, profileID *uuid.UUID, applicationID *uuid.UUID) error {
	switch channel {
	case OTPChannelEmail:
		// Try email profile service first if available
		if s.emailProfileService != nil {
			if err := s.emailProfileService.SendOTPEmail(ctx, profileID, applicationID, destination, otpType, code); err != nil {
				return fmt.Errorf("failed to send OTP email via profile: %w", err)
			}
			return nil
		}
		// Fallback to legacy email service
		if s.emailService == nil {
			return models.NewAppError(503, "Email provider is not configured")
		}
		if err := s.emailService.SendOTP(destination, code, string(otpType)); err != nil {
			return fmt.Errorf("failed to send OTP email: %w", err)
		}
	case OTPChannelSMS:
		if err := s.sendSMS(ctx, destination, code, otpType); err != nil {
			return err
		}
	default:
		return models.NewAppError(400, "Unsupported OTP delivery channel")
	}

	return nil
}

func (s *OTPService) sendSMS(ctx context.Context, phone, code string, otpType models.OTPType) error {
	if s.smsProvider == nil {
		return models.NewAppError(503, "SMS provider is not configured")
	}

	message := s.formatOTPMessage(code, otpType)

	var logID *uuid.UUID
	if s.smsLogRepo != nil {
		smsLog := &models.SMSLog{
			ID:        uuid.New(),
			Phone:     phone,
			Message:   message,
			Type:      otpType,
			Provider:  s.smsProvider.GetProviderName(),
			Status:    models.SMSStatusPending,
			CreatedAt: time.Now(),
		}

		if user, err := s.userRepo.GetByPhone(ctx, phone, utils.Ptr(true)); err == nil {
			smsLog.UserID = &user.ID
		}

		if err := s.smsLogRepo.Create(ctx, smsLog); err == nil {
			logID = &smsLog.ID
		}
	}

	messageID, err := s.smsProvider.SendSMS(ctx, phone, message)
	if err != nil {
		if logID != nil {
			errMsg := err.Error()
			_ = s.smsLogRepo.UpdateStatus(ctx, *logID, models.SMSStatusFailed, &errMsg)
		}
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	if logID != nil {
		_ = s.smsLogRepo.UpdateStatus(ctx, *logID, models.SMSStatusSent, nil)
		// store provider message id if repository supports it via UpdateStatus? For now ignored
		_ = messageID
	}

	return nil
}

func (s *OTPService) fetchOTP(ctx context.Context, channel OTPSendChannel, destination string, otpType models.OTPType) (*models.OTP, error) {
	switch channel {
	case OTPChannelEmail:
		otp, err := s.otpRepo.GetByEmailAndType(ctx, destination, otpType)
		if err != nil {
			s.logOTPError(otpType, map[string]interface{}{"email": destination}, err)
			return nil, err
		}
		return otp, nil
	case OTPChannelSMS:
		otp, err := s.otpRepo.GetByPhoneAndType(ctx, destination, otpType)
		if err != nil {
			s.logOTPError(otpType, map[string]interface{}{"phone": destination}, err)
			return nil, err
		}
		return otp, nil
	default:
		return nil, models.NewAppError(400, "Unsupported OTP delivery channel")
	}
}

func (s *OTPService) logOTPError(otpType models.OTPType, fields map[string]interface{}, err error) {
	details := map[string]interface{}{
		"type": otpType,
	}
	for k, v := range fields {
		details[k] = v
	}

	reason := "otp_lookup_error"
	if appErr, ok := err.(*models.AppError); ok && appErr.Code == 404 {
		reason = "otp_not_found"
	}
	details["reason"] = reason

	s.logAudit(nil, "otp_verify", "failed", "", "", details)
}

func validateOTPType(otpType models.OTPType) error {
	switch otpType {
	case models.OTPTypeVerification, models.OTPTypePasswordReset, models.OTPType2FA, models.OTPTypeLogin, models.OTPTypeRegistration:
		return nil
	default:
		return models.NewAppError(400, "Unsupported OTP type")
	}
}

func (s *OTPService) enrichVerificationResponseEmail(ctx context.Context, resp *models.VerifyOTPResponse, email string, otpType models.OTPType) {
	switch otpType {
	case models.OTPTypeVerification:
		user, err := s.userRepo.GetByEmail(ctx, email, utils.Ptr(true))
		if err == nil && user != nil {
			_ = s.userRepo.MarkEmailVerified(ctx, user.ID)
			resp.User = user
		}
	case models.OTPTypeLogin:
		user, err := s.userRepo.GetByEmail(ctx, email, utils.Ptr(true))
		if err == nil {
			resp.User = user
		}
	case models.OTPTypePasswordReset:
		user, err := s.userRepo.GetByEmail(ctx, email, utils.Ptr(true))
		if err == nil {
			resp.User = user
		}
	}
}

func (s *OTPService) enrichVerificationResponsePhone(ctx context.Context, resp *models.VerifyOTPResponse, phone string, otpType models.OTPType) error {
	switch otpType {
	case models.OTPTypeVerification:
		user, err := s.userRepo.GetByPhone(ctx, phone, utils.Ptr(true))
		if err == nil && user != nil {
			if err := s.userRepo.MarkPhoneVerified(ctx, user.ID); err != nil {
				return err
			}
			resp.User = user
		}
	case models.OTPTypeLogin, models.OTPTypePasswordReset, models.OTPType2FA:
		if user, err := s.userRepo.GetByPhone(ctx, phone, utils.Ptr(true)); err == nil {
			resp.User = user
		}
	}

	return nil
}

func (s *OTPService) formatOTPMessage(code string, otpType models.OTPType) string {
	switch otpType {
	case models.OTPTypeVerification:
		return fmt.Sprintf("Your verification code is: %s\n\nThis code will expire in 10 minutes.\n\nAuth Gateway", code)
	case models.OTPTypePasswordReset:
		return fmt.Sprintf("Your password reset code is: %s\n\nThis code will expire in 10 minutes.\n\nAuth Gateway", code)
	case models.OTPType2FA:
		return fmt.Sprintf("Your 2FA code is: %s\n\nThis code will expire in 10 minutes.\n\nAuth Gateway", code)
	case models.OTPTypeLogin, models.OTPTypeRegistration:
		return fmt.Sprintf("Your login code is: %s\n\nThis code will expire in 10 minutes.\n\nAuth Gateway", code)
	default:
		return fmt.Sprintf("Your verification code is: %s\n\nThis code will expire in 10 minutes.", code)
	}
}

func (s *OTPService) checkSMSRateLimits(ctx context.Context, phone string) error {
	if s.cfg == nil || s.cache == nil {
		return nil
	}

	key := fmt.Sprintf("sms:limit:phone:%s:hour", phone)
	count, err := s.cache.IncrementRateLimit(ctx, key, time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check phone rate limit: %w", err)
	}

	if count > int64(s.cfg.SMS.SMSMaxPerNumber) {
		return models.NewAppError(429, "Too many SMS sent to this number. Please try again later.")
	}

	dailyKey := fmt.Sprintf("sms:limit:phone:%s:day", phone)
	dailyCount, err := s.cache.IncrementRateLimit(ctx, dailyKey, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check daily rate limit: %w", err)
	}

	if dailyCount > int64(s.cfg.SMS.SMSMaxPerDay) {
		return models.NewAppError(429, "Daily SMS limit reached for this number.")
	}

	globalKey := "sms:limit:global:hour"
	globalCount, err := s.cache.IncrementRateLimit(ctx, globalKey, time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check global rate limit: %w", err)
	}

	if globalCount > int64(s.cfg.SMS.SMSMaxPerHour) {
		return models.NewAppError(429, "System SMS limit reached. Please try again later.")
	}

	return nil
}
