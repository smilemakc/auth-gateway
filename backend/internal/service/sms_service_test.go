package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/config"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
)

type mockSMSProvider struct {
	SendSMSFunc func(ctx context.Context, to, message string) (string, error)
}

func (m *mockSMSProvider) SendSMS(ctx context.Context, to, message string) (string, error) {
	if m.SendSMSFunc != nil {
		return m.SendSMSFunc(ctx, to, message)
	}
	return "msg-id", nil
}
func (m *mockSMSProvider) GetProviderName() string { return "mock" }
func (m *mockSMSProvider) ValidateConfig() error   { return nil }

func TestSMSService_SendOTP(t *testing.T) {
	mockProvider := &mockSMSProvider{}
	mockOTP := &mockOTPStore{}
	mockSMSLog := &mockSMSLogStore{}
	mockUser := &mockUserStore{}
	mockRedis := &mockCacheService{}
	cfg := &config.Config{}
	cfg.SMS.SMSMaxPerNumber = 10
	cfg.SMS.SMSMaxPerDay = 50
	cfg.SMS.SMSMaxPerHour = 100

	svc := NewSMSService(mockProvider, mockOTP, mockSMSLog, mockUser, cfg, mockRedis)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &models.SendSMSRequest{
			Phone: "+1234567890",
			Type:  models.OTPTypeVerification,
		}

		mockOTP.CountRecentByPhoneFunc = func(ctx context.Context, phone string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 0, nil
		}

		mockRedis.IncrementRateLimitFunc = func(ctx context.Context, key string, window time.Duration) (int64, error) {
			return 1, nil
		}

		mockOTP.InvalidateAllForPhoneFunc = func(ctx context.Context, phone string, otpType models.OTPType) error {
			return nil
		}

		mockOTP.CreateFunc = func(ctx context.Context, otp *models.OTP) error {
			return nil
		}

		mockUser.GetByPhoneFunc = func(ctx context.Context, phone string, isActive *bool) (*models.User, error) {
			return nil, assert.AnError // Simulate user not found, which is fine
		}

		mockSMSLog.CreateFunc = func(ctx context.Context, log *models.SMSLog) error {
			return nil
		}

		mockProvider.SendSMSFunc = func(ctx context.Context, to, message string) (string, error) {
			return "msg-id", nil
		}

		mockSMSLog.UpdateStatusFunc = func(ctx context.Context, id uuid.UUID, status models.SMSStatus, errorMsg *string) error {
			return nil
		}

		resp, err := svc.SendOTP(ctx, req, "127.0.0.1")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.NotNil(t, resp.MessageID)
	})

	t.Run("RateLimitExceeded", func(t *testing.T) {
		req := &models.SendSMSRequest{
			Phone: "+1234567890",
			Type:  models.OTPTypeVerification,
		}

		mockOTP.CountRecentByPhoneFunc = func(ctx context.Context, phone string, otpType models.OTPType, duration time.Duration) (int, error) {
			return 3, nil // Limit is 3
		}

		resp, err := svc.SendOTP(ctx, req, "127.0.0.1")
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "Too many SMS requests")
	})
}

func TestSMSService_VerifyOTP(t *testing.T) {
	mockProvider := &mockSMSProvider{}
	mockOTP := &mockOTPStore{}
	mockSMSLog := &mockSMSLogStore{}
	mockUser := &mockUserStore{}
	mockRedis := &mockCacheService{}
	cfg := &config.Config{}

	svc := NewSMSService(mockProvider, mockOTP, mockSMSLog, mockUser, cfg, mockRedis)
	ctx := context.Background()

	t.Run("Success_Verification", func(t *testing.T) {
		phone := "+1234567890"
		code := "123456"
		codeHash, _ := utils.HashPassword(code, 10)

		otp := &models.OTP{
			ID:        uuid.New(),
			Phone:     &phone,
			Code:      codeHash,
			Type:      models.OTPTypeVerification,
			ExpiresAt: time.Now().Add(time.Minute),
			Used:      false,
		}

		req := &models.VerifySMSOTPRequest{
			Phone: phone,
			Code:  code,
			Type:  models.OTPTypeVerification,
		}

		mockOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, phone string, otpType models.OTPType) (*models.OTP, error) {
			return otp, nil
		}

		mockOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		mockUser.GetByPhoneFunc = func(ctx context.Context, phone string, isActive *bool) (*models.User, error) {
			return &models.User{ID: uuid.New()}, nil
		}

		mockUser.MarkPhoneVerifiedFunc = func(ctx context.Context, userID uuid.UUID) error {
			return nil
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.Valid)
		assert.NotNil(t, resp.User)
	})

	t.Run("InvalidCode", func(t *testing.T) {
		phone := "+1234567890"
		code := "123456"
		wrongCode := "654321"
		codeHash, _ := utils.HashPassword(code, 10)

		otp := &models.OTP{
			ID:        uuid.New(),
			Phone:     &phone,
			Code:      codeHash,
			Type:      models.OTPTypeVerification,
			ExpiresAt: time.Now().Add(time.Minute),
			Used:      false,
		}

		req := &models.VerifySMSOTPRequest{
			Phone: phone,
			Code:  wrongCode,
			Type:  models.OTPTypeVerification,
		}

		mockOTP.GetByPhoneAndTypeFunc = func(ctx context.Context, phone string, otpType models.OTPType) (*models.OTP, error) {
			return otp, nil
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err) // It returns valid=false, not error
		assert.NotNil(t, resp)
		assert.False(t, resp.Valid)
		assert.Equal(t, "Invalid OTP code", resp.Message)
	})
}
