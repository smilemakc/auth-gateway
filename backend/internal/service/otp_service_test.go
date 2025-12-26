package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
)

func setupOTPService() (*OTPService, *mockOTPStore, *mockUserStore, *mockEmailSender, *mockAuditLogger) {
	mOTP := &mockOTPStore{}
	mUser := &mockUserStore{}
	mEmail := &mockEmailSender{}
	mAudit := &mockAuditLogger{}

	svc := NewOTPService(mOTP, mUser, mAudit, OTPServiceOptions{
		EmailSender: mEmail,
	})
	return svc, mOTP, mUser, mEmail, mAudit
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
}

func TestOTPService_VerifyOTP(t *testing.T) {
	svc, mOTP, mUser, _, mAudit := setupOTPService()
	ctx := context.Background()
	email := "test@example.com"
	otpCode := "123456"
	hashedCode, _ := utils.HashPassword(otpCode, 10)

	req := &models.VerifyOTPRequest{
		Email: &email,
		Code:  otpCode,
		Type:  models.OTPTypeVerification,
	}

	t.Run("SuccessVerification", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return &models.OTP{
				ID:        uuid.New(),
				Code:      hashedCode,
				ExpiresAt: time.Now().Add(time.Hour),
				Used:      false,
			}, nil
		}
		mOTP.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mUser.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool) (*models.User, error) {
			return &models.User{ID: uuid.New(), Email: email}, nil
		}
		mUser.MarkEmailVerifiedFunc = func(ctx context.Context, userID uuid.UUID) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, "otp_verify", params.Action)
			assert.Equal(t, "success", params.Status)
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.NoError(t, err)
		assert.True(t, resp.Valid)
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

	t.Run("RepositoryErrorIsReturned", func(t *testing.T) {
		mOTP.GetByEmailAndTypeFunc = func(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
			return nil, fmt.Errorf("db unavailable")
		}

		resp, err := svc.VerifyOTP(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
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
}
