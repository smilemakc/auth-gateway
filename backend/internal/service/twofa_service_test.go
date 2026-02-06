package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestTwoFactorService_SetupTOTP(t *testing.T) {
	mockUser := &mockUserStore{}
	mockBackup := &mockBackupCodeStore{}
	issuer := "TestApp"

	svc := NewTwoFactorService(mockUser, mockBackup, issuer)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		password := "password123"
		hashedPassword, _ := utils.HashPassword(password, 10)

		mockUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Email:        "test@example.com",
				PasswordHash: hashedPassword,
			}, nil
		}

		mockUser.UpdateTOTPSecretFunc = func(ctx context.Context, id uuid.UUID, secret string) error {
			assert.NotEmpty(t, secret)
			return nil
		}

		mockBackup.CreateBatchFunc = func(ctx context.Context, codes []*models.BackupCode) error {
			assert.Len(t, codes, 10)
			return nil
		}

		resp, err := svc.SetupTOTP(ctx, userID, password)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Secret)
		assert.NotEmpty(t, resp.QRCodeURL)
		assert.Len(t, resp.BackupCodes, 10)
	})
}

func TestTwoFactorService_VerifyTOTPSetup(t *testing.T) {
	mockUser := &mockUserStore{}
	mockBackup := &mockBackupCodeStore{}
	issuer := "TestApp"

	svc := NewTwoFactorService(mockUser, mockBackup, issuer)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		key, _ := totp.Generate(totp.GenerateOpts{Issuer: issuer, AccountName: "test@example.com"})
		secret := key.Secret()

		mockUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:         userID,
				TOTPSecret: &secret,
			}, nil
		}

		mockUser.EnableTOTPFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		code, _ := totp.GenerateCode(secret, time.Now())
		err := svc.VerifyTOTPSetup(ctx, userID, code)
		assert.NoError(t, err)
	})
}

func TestTwoFactorService_VerifyTOTP(t *testing.T) {
	mockUser := &mockUserStore{}
	mockBackup := &mockBackupCodeStore{}
	issuer := "TestApp"

	svc := NewTwoFactorService(mockUser, mockBackup, issuer)
	ctx := context.Background()

	t.Run("Success_TOTP", func(t *testing.T) {
		userID := uuid.New()
		key, _ := totp.Generate(totp.GenerateOpts{Issuer: issuer, AccountName: "test@example.com"})
		secret := key.Secret()

		mockUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:          userID,
				TOTPEnabled: true,
				TOTPSecret:  &secret,
			}, nil
		}

		code, _ := totp.GenerateCode(secret, time.Now())
		valid, err := svc.VerifyTOTP(ctx, userID, code)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("Success_BackupCode", func(t *testing.T) {
		userID := uuid.New()
		// Invalid TOTP code to force check backup codes
		backupCodeRaw := "12345678"
		backupCodeHash, _ := utils.HashPassword(backupCodeRaw, 10)

		key, _ := totp.Generate(totp.GenerateOpts{Issuer: issuer, AccountName: "test@example.com"})
		secret := key.Secret()

		mockUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:          userID,
				TOTPEnabled: true,
				TOTPSecret:  &secret,
			}, nil
		}

		mockBackup.GetUnusedByUserIDFunc = func(ctx context.Context, id uuid.UUID) ([]*models.BackupCode, error) {
			return []*models.BackupCode{
				{ID: uuid.New(), CodeHash: backupCodeHash, Used: false},
			}, nil
		}

		mockBackup.MarkAsUsedFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		valid, err := svc.VerifyTOTP(ctx, userID, backupCodeRaw)
		assert.NoError(t, err)
		assert.True(t, valid)
	})
}

func TestTwoFactorService_DisableTOTP(t *testing.T) {
	mockUser := &mockUserStore{}
	mockBackup := &mockBackupCodeStore{}
	issuer := "TestApp"

	svc := NewTwoFactorService(mockUser, mockBackup, issuer)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		password := "password123"
		hashedPassword, _ := utils.HashPassword(password, 10)

		key, _ := totp.Generate(totp.GenerateOpts{Issuer: issuer, AccountName: "test@example.com"})
		secret := key.Secret()

		mockUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:           userID,
				PasswordHash: hashedPassword,
				TOTPEnabled:  true,
				TOTPSecret:   &secret,
			}, nil
		}

		mockUser.DisableTOTPFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		mockBackup.DeleteAllByUserIDFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		code, _ := totp.GenerateCode(secret, time.Now())
		err := svc.DisableTOTP(ctx, userID, password, code)
		assert.NoError(t, err)
	})
}

func TestTwoFactorService_GetStatus(t *testing.T) {
	mockUser := &mockUserStore{}
	mockBackup := &mockBackupCodeStore{}
	issuer := "TestApp"

	svc := NewTwoFactorService(mockUser, mockBackup, issuer)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		now := time.Now()

		mockUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:            userID,
				TOTPEnabled:   true,
				TOTPEnabledAt: &now,
			}, nil
		}

		mockBackup.CountUnusedByUserIDFunc = func(ctx context.Context, id uuid.UUID) (int, error) {
			return 8, nil
		}

		status, err := svc.GetStatus(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.True(t, status.Enabled)
		assert.Equal(t, 8, status.BackupCodes)
	})
}

func TestTwoFactorService_RegenerateBackupCodes(t *testing.T) {
	mockUser := &mockUserStore{}
	mockBackup := &mockBackupCodeStore{}
	issuer := "TestApp"

	svc := NewTwoFactorService(mockUser, mockBackup, issuer)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		password := "password123"
		hashedPassword, _ := utils.HashPassword(password, 10)

		mockUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:           userID,
				PasswordHash: hashedPassword,
				TOTPEnabled:  true,
			}, nil
		}

		mockBackup.DeleteAllByUserIDFunc = func(ctx context.Context, id uuid.UUID) error {
			return nil
		}

		mockBackup.CreateBatchFunc = func(ctx context.Context, codes []*models.BackupCode) error {
			assert.Len(t, codes, 10)
			return nil
		}

		codes, err := svc.RegenerateBackupCodes(ctx, userID, password)
		assert.NoError(t, err)
		assert.Len(t, codes, 10)
	})
}
