package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
)

func testConfig() *config.Config {
	return &config.Config{
		Security: config.SecurityConfig{
			OTPHMACSecret: "test-otp-hmac-secret-for-testing-32-chars-minimum",
		},
		SMS: config.SMSConfig{
			SMSMaxPerNumber: 5,
			SMSMaxPerDay:    50,
			SMSMaxPerHour:   10,
		},
	}
}

func setupOTPService() (*OTPService, *mockOTPStore, *mockUserStore, *mockEmailSender, *mockAuditLogger) {
	mOTP := &mockOTPStore{}
	mUser := &mockUserStore{}
	mEmail := &mockEmailSender{}
	mAudit := &mockAuditLogger{}

	cfg := testConfig()
	svc := NewOTPService(mOTP, mUser, mAudit, OTPServiceOptions{
		EmailSender: mEmail,
		Config:      cfg,
	})
	return svc, mOTP, mUser, mEmail, mAudit
}

func setupOTPServiceWithSMS() (*OTPService, *mockOTPStore, *mockUserStore, *mockEmailSender, *mockSMSProvider, *mockSMSLogStore, *mockCacheService, *mockAuditLogger) {
	mOTP := &mockOTPStore{}
	mUser := &mockUserStore{}
	mEmail := &mockEmailSender{}
	mSMS := &mockSMSProvider{}
	mSMSLog := &mockSMSLogStore{}
	mCache := &mockCacheService{}
	mAudit := &mockAuditLogger{}

	cfg := testConfig()
	svc := NewOTPService(mOTP, mUser, mAudit, OTPServiceOptions{
		EmailSender: mEmail,
		SMSProvider: mSMS,
		SMSLogRepo:  mSMSLog,
		Cache:       mCache,
		Config:      cfg,
	})
	return svc, mOTP, mUser, mEmail, mSMS, mSMSLog, mCache, mAudit
}

func TestOTPService_SendOTP(t *testing.T) {
	svc, mOTP, _, mEmail, mAudit := setupOTPService()
	ctx := context.Background()
	email := "test@example.com"
	req := &models.SendOTPRequest{
		Email: &email,
		Type:  models.OTPTypeVerification,
	}

	t.Run("Success", func(t *testing.T) {
		mOTP.CountRecentByEmailFunc = func(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mOTP.InvalidateAllForEmailFunc = func(ctx context.Context, email string, otpType models.OTPType) error {
			return nil
		}
		mOTP.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			assert.NotEmpty(t, otp.Code)
			return nil
		}
		mEmail.SendOTPFunc = func(to, code, otpType string) error {
			assert.Equal(t, email, to)
			return nil
		}
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, "otp_sent", params.Action)
			assert.Equal(t, "success", params.Status)
		}

		err := svc.SendOTP(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("RateLimitExceeded", func(t *testing.T) {
		mOTP.CountRecentByEmailFunc = func(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
			return OTPRateLimit, nil
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, 429, err.(*models.AppError).Code)
	})

	t.Run("InvalidType", func(t *testing.T) {
		invalidReq := &models.SendOTPRequest{
			Email: &email,
			Type:  "invalid_type",
		}

		err := svc.SendOTP(ctx, invalidReq)
		assert.Error(t, err)
		assert.Equal(t, 400, err.(*models.AppError).Code)
	})

	t.Run("NoEmailOrPhone", func(t *testing.T) {
		emptyReq := &models.SendOTPRequest{
			Type: models.OTPTypeVerification,
		}

		err := svc.SendOTP(ctx, emptyReq)
		assert.Error(t, err)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.Code)
		assert.Contains(t, appErr.Message, "email or phone")
	})

	t.Run("NoEmailOrPhone_EmptyStrings", func(t *testing.T) {
		emptyEmail := ""
		emptyPhone := ""
		emptyReq := &models.SendOTPRequest{
			Email: &emptyEmail,
			Phone: &emptyPhone,
			Type:  models.OTPTypeVerification,
		}

		err := svc.SendOTP(ctx, emptyReq)
		assert.Error(t, err)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.Code)
	})

	t.Run("EmailSendError", func(t *testing.T) {
		mOTP.CountRecentByEmailFunc = func(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mOTP.InvalidateAllForEmailFunc = func(ctx context.Context, email string, otpType models.OTPType) error {
			return nil
		}
		mOTP.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			return nil
		}
		mEmail.SendOTPFunc = func(to, code, otpType string) error {
			return fmt.Errorf("SMTP connection refused")
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send OTP email")
	})

	t.Run("CreateOTPError", func(t *testing.T) {
		mOTP.CountRecentByEmailFunc = func(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mOTP.InvalidateAllForEmailFunc = func(ctx context.Context, email string, otpType models.OTPType) error {
			return nil
		}
		mOTP.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			return fmt.Errorf("db write error")
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db write error")
	})

	t.Run("InvalidateAllForEmailError", func(t *testing.T) {
		mOTP.CountRecentByEmailFunc = func(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mOTP.InvalidateAllForEmailFunc = func(ctx context.Context, email string, otpType models.OTPType) error {
			return fmt.Errorf("db error during invalidation")
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error during invalidation")
	})

	t.Run("CountRecentByEmailError", func(t *testing.T) {
		mOTP.CountRecentByEmailFunc = func(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, fmt.Errorf("redis unavailable")
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis unavailable")
	})

	t.Run("EmailServiceNotConfigured", func(t *testing.T) {
		// Build a service with no email sender and no email profile service
		mOTPLocal := &mockOTPStore{}
		mUserLocal := &mockUserStore{}
		mAuditLocal := &mockAuditLogger{}
		cfg := testConfig()
		svcNoEmail := NewOTPService(mOTPLocal, mUserLocal, mAuditLocal, OTPServiceOptions{
			Config: cfg,
		})

		mOTPLocal.CountRecentByEmailFunc = func(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mOTPLocal.InvalidateAllForEmailFunc = func(ctx context.Context, email string, otpType models.OTPType) error {
			return nil
		}
		mOTPLocal.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			return nil
		}

		err := svcNoEmail.SendOTP(ctx, req)
		assert.Error(t, err)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 503, appErr.Code)
		assert.Contains(t, appErr.Message, "Email provider is not configured")
	})
}

func TestOTPService_SendOTP_SMS(t *testing.T) {
	svc, mOTP, mUser, _, mSMS, mSMSLog, mCache, _ := setupOTPServiceWithSMS()
	ctx := context.Background()
	phone := "+79991234567"

	t.Run("SMSSuccess", func(t *testing.T) {
		req := &models.SendOTPRequest{
			Phone: &phone,
			Type:  models.OTPTypeVerification,
		}

		mOTP.CountRecentByPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mCache.IncrementRateLimitFunc = func(ctx context.Context, key string, window time.Duration) (int64, error) {
			return 1, nil
		}
		mOTP.InvalidateAllForPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType) error {
			return nil
		}
		mOTP.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			assert.NotNil(t, otp.Phone)
			assert.Equal(t, phone, *otp.Phone)
			assert.Nil(t, otp.Email)
			assert.NotEmpty(t, otp.Code)
			return nil
		}
		mUser.GetByPhoneFunc = func(ctx context.Context, ph string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		}
		mSMSLog.CreateFunc = func(ctx context.Context, log *models.SMSLog) error {
			return nil
		}
		mSMS.SendSMSFunc = func(ctx context.Context, to, message string) (string, error) {
			assert.Equal(t, phone, to)
			assert.Contains(t, message, "verification code")
			return "msg-123", nil
		}
		mSMSLog.UpdateStatusFunc = func(ctx context.Context, id uuid.UUID, status models.SMSStatus, errorMsg *string) error {
			assert.Equal(t, models.SMSStatusSent, status)
			return nil
		}

		err := svc.SendOTP(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("SMSRateLimitExceeded", func(t *testing.T) {
		req := &models.SendOTPRequest{
			Phone: &phone,
			Type:  models.OTPTypeVerification,
		}

		mOTP.CountRecentByPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType, duration time.Duration) (int, error) {
			return SMSRateLimit, nil
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 429, appErr.Code)
	})

	t.Run("SMSProviderError", func(t *testing.T) {
		req := &models.SendOTPRequest{
			Phone: &phone,
			Type:  models.OTPTypeVerification,
		}

		mOTP.CountRecentByPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mCache.IncrementRateLimitFunc = func(ctx context.Context, key string, window time.Duration) (int64, error) {
			return 1, nil
		}
		mOTP.InvalidateAllForPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType) error {
			return nil
		}
		mOTP.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			return nil
		}
		mUser.GetByPhoneFunc = func(ctx context.Context, ph string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		}
		mSMSLog.CreateFunc = func(ctx context.Context, log *models.SMSLog) error {
			return nil
		}
		mSMS.SendSMSFunc = func(ctx context.Context, to, message string) (string, error) {
			return "", fmt.Errorf("twilio unavailable")
		}
		mSMSLog.UpdateStatusFunc = func(ctx context.Context, id uuid.UUID, status models.SMSStatus, errorMsg *string) error {
			assert.Equal(t, models.SMSStatusFailed, status)
			assert.NotNil(t, errorMsg)
			return nil
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send SMS")
	})

	t.Run("SMSProviderNotConfigured", func(t *testing.T) {
		// Build a service without SMS provider
		mOTPLocal := &mockOTPStore{}
		mUserLocal := &mockUserStore{}
		mAuditLocal := &mockAuditLogger{}
		cfg := testConfig()
		svcNoSMS := NewOTPService(mOTPLocal, mUserLocal, mAuditLocal, OTPServiceOptions{
			Config: cfg,
		})

		req := &models.SendOTPRequest{
			Phone: &phone,
			Type:  models.OTPTypeVerification,
		}

		mOTPLocal.CountRecentByPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mOTPLocal.InvalidateAllForPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType) error {
			return nil
		}
		mOTPLocal.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			return nil
		}

		err := svcNoSMS.SendOTP(ctx, req)
		assert.Error(t, err)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 503, appErr.Code)
		assert.Contains(t, appErr.Message, "SMS provider is not configured")
	})

	t.Run("SMSInvalidateAllForPhoneError", func(t *testing.T) {
		req := &models.SendOTPRequest{
			Phone: &phone,
			Type:  models.OTPTypeVerification,
		}

		mOTP.CountRecentByPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mCache.IncrementRateLimitFunc = func(ctx context.Context, key string, window time.Duration) (int64, error) {
			return 1, nil
		}
		mOTP.InvalidateAllForPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType) error {
			return fmt.Errorf("db phone invalidation error")
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db phone invalidation error")
	})

	t.Run("SMSCreateOTPError", func(t *testing.T) {
		req := &models.SendOTPRequest{
			Phone: &phone,
			Type:  models.OTPTypeVerification,
		}

		mOTP.CountRecentByPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mCache.IncrementRateLimitFunc = func(ctx context.Context, key string, window time.Duration) (int64, error) {
			return 1, nil
		}
		mOTP.InvalidateAllForPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType) error {
			return nil
		}
		mOTP.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			return fmt.Errorf("db create error")
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db create error")
	})

	t.Run("SMSCacheRateLimitError", func(t *testing.T) {
		req := &models.SendOTPRequest{
			Phone: &phone,
			Type:  models.OTPTypeVerification,
		}

		mOTP.CountRecentByPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mCache.IncrementRateLimitFunc = func(ctx context.Context, key string, window time.Duration) (int64, error) {
			return 0, fmt.Errorf("cache unavailable")
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check phone rate limit")
	})

	t.Run("SMSPerNumberLimitExceeded", func(t *testing.T) {
		req := &models.SendOTPRequest{
			Phone: &phone,
			Type:  models.OTPTypeVerification,
		}

		mOTP.CountRecentByPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mCache.IncrementRateLimitFunc = func(ctx context.Context, key string, window time.Duration) (int64, error) {
			// Exceed per-number limit (config has SMSMaxPerNumber=5)
			return 6, nil
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 429, appErr.Code)
		assert.Contains(t, appErr.Message, "Too many SMS sent to this number")
	})

	t.Run("InvalidPhoneFormat", func(t *testing.T) {
		badPhone := "not-a-phone"
		req := &models.SendOTPRequest{
			Phone: &badPhone,
			Type:  models.OTPTypeVerification,
		}

		err := svc.SendOTP(ctx, req)
		assert.Error(t, err)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.Code)
		assert.Contains(t, appErr.Message, "Invalid phone number format")
	})

	t.Run("SMSWithExistingUser", func(t *testing.T) {
		req := &models.SendOTPRequest{
			Phone: &phone,
			Type:  models.OTPTypeLogin,
		}

		userID := uuid.New()
		mOTP.CountRecentByPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}
		mCache.IncrementRateLimitFunc = func(ctx context.Context, key string, window time.Duration) (int64, error) {
			return 1, nil
		}
		mOTP.InvalidateAllForPhoneFunc = func(ctx context.Context, ph string, otpType models.OTPType) error {
			return nil
		}
		mOTP.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			return nil
		}
		mUser.GetByPhoneFunc = func(ctx context.Context, ph string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID, Phone: &phone}, nil
		}
		mSMSLog.CreateFunc = func(ctx context.Context, log *models.SMSLog) error {
			assert.Equal(t, &userID, log.UserID)
			return nil
		}
		mSMS.SendSMSFunc = func(ctx context.Context, to, message string) (string, error) {
			assert.Contains(t, message, "login code")
			return "msg-456", nil
		}
		mSMSLog.UpdateStatusFunc = func(ctx context.Context, id uuid.UUID, status models.SMSStatus, errorMsg *string) error {
			return nil
		}

		err := svc.SendOTP(ctx, req)
		assert.NoError(t, err)
	})
}

func TestOTPService_VerifyOTP(t *testing.T) {
	svc, mOTP, mUser, _, mAudit := setupOTPService()
	ctx := context.Background()
	email := "test@example.com"
	otpCode := "123456"
	hashedCode := utils.HMACHash(otpCode, svc.cfg.Security.OTPHMACSecret)

	req := &models.VerifyOTPRequest{
		Email: &email,
		Code:  otpCode,
		Type:  models.OTPTypeVerification,
	}

	t.Run("SuccessVerification", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, em string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Email:     &em,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mUser.GetByEmailFunc = func(ctx context.Context, em string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: uuid.New(), Email: em}, nil
		}
		mUser.MarkEmailVerifiedFunc = func(ctx context.Context, userID uuid.UUID) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, "otp_verify", params.Action)
			assert.Equal(t, "success", params.Status)
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.NotNil(t, resp.User)
	})

	t.Run("SuccessVerification_UserNotFound", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Email:     &email,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mUser.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, fmt.Errorf("user not found")
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Nil(t, resp.User)
	})

	t.Run("SuccessLogin", func(t *testing.T) {
		loginReq := &models.VerifyOTPRequest{
			Email: &email,
			Code:  otpCode,
			Type:  models.OTPTypeLogin,
		}

		userID := uuid.New()
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Email:     &email,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mUser.GetByEmailFunc = func(ctx context.Context, em string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID, Email: em}, nil
		}

		resp, err := svc.VerifyOTP(ctx, loginReq)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.NotNil(t, resp.User)
		assert.Equal(t, userID, resp.User.ID)
	})

	t.Run("SuccessPasswordReset", func(t *testing.T) {
		resetReq := &models.VerifyOTPRequest{
			Email: &email,
			Code:  otpCode,
			Type:  models.OTPTypePasswordReset,
		}

		userID := uuid.New()
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Email:     &email,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mUser.GetByEmailFunc = func(ctx context.Context, em string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID, Email: em}, nil
		}

		resp, err := svc.VerifyOTP(ctx, resetReq)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.NotNil(t, resp.User)
		assert.Equal(t, userID, resp.User.ID)
	})

	t.Run("InvalidCode", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		}

		reqInvalid := &models.VerifyOTPRequest{Email: &email, Code: "000000", Type: models.OTPTypeVerification}

		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, "otp_verify", params.Action)
			assert.Equal(t, "failed", params.Status)
		}

		resp, err := svc.VerifyOTP(ctx, reqInvalid)
		assert.NoError(t, err)
		assert.False(t, resp.Valid)
	})

	t.Run("ExpiredOTP", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(-time.Hour), // Expired
			}, nil
		}

		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, "otp_verify", params.Action)
			assert.Equal(t, "failed", params.Status)
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.False(t, resp.Valid)
	})

	t.Run("AlreadyUsed", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Email:     &email,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      true,
			}, nil
		}

		auditCalled := false
		mAudit.LogWithActionFunc = func(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
			if action == "otp_verify" && status == "failed" {
				if reason, ok := details["reason"]; ok && reason == "already_used" {
					auditCalled = true
				}
			}
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.False(t, resp.Valid)
		assert.True(t, auditCalled, "audit should log already_used reason")
	})

	t.Run("NotFound_Returns404AppError", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return nil, models.NewAppError(404, "OTP not found")
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Valid)
	})

	t.Run("RepositoryErrorIsReturned", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return nil, fmt.Errorf("db unavailable")
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("MarkAsUsedError", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Email:     &email,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error {
			return fmt.Errorf("db error marking used")
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "db error marking used")
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		badReq := &models.VerifyOTPRequest{
			Email: &email,
			Code:  otpCode,
			Type:  "not_supported",
		}

		resp, err := svc.VerifyOTP(ctx, badReq)
		assert.Error(t, err)
		assert.Nil(t, resp)
		if appErr, ok := err.(*models.AppError); ok {
			assert.Equal(t, 400, appErr.Code)
		} else {
			t.Fatalf("expected AppError, got %T", err)
		}
	})

	t.Run("NoEmailOrPhone", func(t *testing.T) {
		emptyReq := &models.VerifyOTPRequest{
			Code: otpCode,
			Type: models.OTPTypeVerification,
		}

		resp, err := svc.VerifyOTP(ctx, emptyReq)
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.Code)
		assert.Contains(t, appErr.Message, "email or phone")
	})
}

func TestOTPService_VerifyOTP_SMS(t *testing.T) {
	svc, mOTP, mUser, _, _, _, _, _ := setupOTPServiceWithSMS()
	ctx := context.Background()
	phone := "+79991234567"
	otpCode := "654321"
	hashedCode := utils.HMACHash(otpCode, svc.cfg.Security.OTPHMACSecret)

	t.Run("SuccessVerification_Phone", func(t *testing.T) {
		req := &models.VerifyOTPRequest{
			Phone: &phone,
			Code:  otpCode,
			Type:  models.OTPTypeVerification,
		}

		userID := uuid.New()
		mOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, ph string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Phone:     &phone,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mUser.GetByPhoneFunc = func(ctx context.Context, ph string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID, Phone: &phone}, nil
		}
		mUser.MarkPhoneVerifiedFunc = func(ctx context.Context, uid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			return nil
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.NotNil(t, resp.User)
		assert.Equal(t, userID, resp.User.ID)
	})

	t.Run("SuccessLogin_Phone", func(t *testing.T) {
		req := &models.VerifyOTPRequest{
			Phone: &phone,
			Code:  otpCode,
			Type:  models.OTPTypeLogin,
		}

		userID := uuid.New()
		mOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, ph string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Phone:     &phone,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mUser.GetByPhoneFunc = func(ctx context.Context, ph string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID, Phone: &phone}, nil
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.NotNil(t, resp.User)
	})

	t.Run("Success2FA_Phone", func(t *testing.T) {
		req := &models.VerifyOTPRequest{
			Phone: &phone,
			Code:  otpCode,
			Type:  models.OTPType2FA,
		}

		userID := uuid.New()
		mOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, ph string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Phone:     &phone,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mUser.GetByPhoneFunc = func(ctx context.Context, ph string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID, Phone: &phone}, nil
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.NotNil(t, resp.User)
	})

	t.Run("PhoneNotFound", func(t *testing.T) {
		req := &models.VerifyOTPRequest{
			Phone: &phone,
			Code:  otpCode,
			Type:  models.OTPTypeVerification,
		}

		mOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, ph string, otpType models.OTPType) (*models.OTP, error) {
			return nil, models.NewAppError(404, "OTP not found")
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Valid)
	})

	t.Run("PhoneExpiredOTP", func(t *testing.T) {
		req := &models.VerifyOTPRequest{
			Phone: &phone,
			Code:  otpCode,
			Type:  models.OTPTypeVerification,
		}

		mOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, ph string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Phone:     &phone,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(-time.Hour),
				Used:      false,
			}, nil
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.False(t, resp.Valid)
	})

	t.Run("PhoneAlreadyUsed", func(t *testing.T) {
		req := &models.VerifyOTPRequest{
			Phone: &phone,
			Code:  otpCode,
			Type:  models.OTPTypeVerification,
		}

		mOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, ph string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Phone:     &phone,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      true,
			}, nil
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.False(t, resp.Valid)
	})

	t.Run("PhoneInvalidCode", func(t *testing.T) {
		req := &models.VerifyOTPRequest{
			Phone: &phone,
			Code:  "999999",
			Type:  models.OTPTypeVerification,
		}

		mOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, ph string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Phone:     &phone,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.False(t, resp.Valid)
	})

	t.Run("PhoneVerification_MarkPhoneVerifiedError", func(t *testing.T) {
		req := &models.VerifyOTPRequest{
			Phone: &phone,
			Code:  otpCode,
			Type:  models.OTPTypeVerification,
		}

		mOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, ph string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Phone:     &phone,
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mUser.GetByPhoneFunc = func(ctx context.Context, ph string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: uuid.New(), Phone: &phone}, nil
		}
		mUser.MarkPhoneVerifiedFunc = func(ctx context.Context, uid uuid.UUID) error {
			return fmt.Errorf("db error marking phone verified")
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "db error marking phone verified")
	})

	t.Run("InvalidPhoneFormat", func(t *testing.T) {
		badPhone := "not-a-phone"
		req := &models.VerifyOTPRequest{
			Phone: &badPhone,
			Code:  otpCode,
			Type:  models.OTPTypeVerification,
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.Code)
		assert.Contains(t, appErr.Message, "Invalid phone number format")
	})
}

func TestOTPService_CleanupExpiredOTPs(t *testing.T) {
	svc, mOTP, _, _, _ := setupOTPService()

	t.Run("Success", func(t *testing.T) {
		mOTP.DeleteExpiredFunc = func(ctx context.Context, olderThan time.Duration) error {
			assert.Equal(t, 7*24*time.Hour, olderThan)
			return nil
		}

		err := svc.CleanupExpiredOTPs()
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mOTP.DeleteExpiredFunc = func(ctx context.Context, olderThan time.Duration) error {
			return fmt.Errorf("cleanup failed")
		}

		err := svc.CleanupExpiredOTPs()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cleanup failed")
	})
}

func TestOTPService_GenerateOTPCode(t *testing.T) {
	svc, _, _, _, _ := setupOTPService()

	t.Run("GeneratesValidCode", func(t *testing.T) {
		code, err := svc.GenerateOTPCode()
		assert.NoError(t, err)
		assert.Len(t, code, 6)
		// Verify all characters are digits
		for _, c := range code {
			assert.True(t, c >= '0' && c <= '9', "expected digit, got %c", c)
		}
	})

	t.Run("GeneratesUniqueCodesOverMultipleCalls", func(t *testing.T) {
		codes := make(map[string]struct{})
		for i := 0; i < 100; i++ {
			code, err := svc.GenerateOTPCode()
			assert.NoError(t, err)
			codes[code] = struct{}{}
		}
		// With 100 random 6-digit codes, we should have significant diversity
		// (probability of all same is astronomically low)
		assert.Greater(t, len(codes), 1)
	})
}
