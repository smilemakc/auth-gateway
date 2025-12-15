package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

func setupAuthService() (*AuthService, *mockUserStore, *mockTokenStore, *mockRBACStore, *mockAuditLogger, *mockTokenService, *mockCacheService) {
	mUser := &mockUserStore{}
	mToken := &mockTokenStore{}
	mRBAC := &mockRBACStore{}
	mAudit := &mockAuditLogger{}
	mJWT := &mockTokenService{}
	mCache := &mockCacheService{}

	// SessionCreationService is nil for tests (non-fatal session creation)
	svc := NewAuthService(mUser, mToken, mRBAC, mAudit, mJWT, mCache, nil, 10)
	return svc, mUser, mToken, mRBAC, mAudit, mJWT, mCache
}

func TestAuthService_SignUp(t *testing.T) {
	svc, mUser, mToken, mRBAC, mAudit, mJWT, _ := setupAuthService()
	ctx := context.Background()

	validReq := &models.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		FullName: "Test User",
	}

	t.Run("Success", func(t *testing.T) {
		mUser.EmailExistsFunc = func(ctx context.Context, email string) (bool, error) { return false, nil }
		mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
		mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: uuid.New(), Name: "user"}, nil
		}
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			assert.NotEmpty(t, user.PasswordHash)
			return nil
		}
		mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }
		mUser.GetByIDWithRolesFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: id, Email: validReq.Email, Roles: []models.Role{{Name: "user"}}}, nil
		}
		mJWT.GenerateAccessTokenFunc = func(user *models.User) (string, error) { return "access_token", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User) (string, error) { return "refresh_token", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}

		resp, err := svc.SignUp(ctx, validReq, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "access_token", resp.AccessToken)
	})

	t.Run("EmailExists", func(t *testing.T) {
		mUser.EmailExistsFunc = func(ctx context.Context, email string) (bool, error) { return true, nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		resp, err := svc.SignUp(ctx, validReq, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.ErrorIs(t, err, models.ErrEmailAlreadyExists)
		assert.Nil(t, resp)
	})

	t.Run("UsernameExists", func(t *testing.T) {
		mUser.EmailExistsFunc = func(ctx context.Context, email string) (bool, error) { return false, nil }
		mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return true, nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		resp, err := svc.SignUp(ctx, validReq, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.ErrorIs(t, err, models.ErrUsernameAlreadyExists)
		assert.Nil(t, resp)
	})
}

func TestAuthService_SignIn(t *testing.T) {
	svc, mUser, mToken, _, mAudit, mJWT, _ := setupAuthService()
	ctx := context.Background()

	password := "password123"
	hash, _ := utils.HashPassword(password, 10)
	userID := uuid.New()

	req := &models.SignInRequest{
		Email:    "test@example.com",
		Password: password,
	}

	t.Run("Success", func(t *testing.T) {
		mUser.GetByEmailWithRolesFunc = func(ctx context.Context, email string, isActive *bool) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Email:        req.Email,
				PasswordHash: hash,
				IsActive:     true,
			}, nil
		}

		mJWT.GenerateAccessTokenFunc = func(user *models.User) (string, error) { return "access_token", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User) (string, error) { return "refresh_token", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignIn, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}

		resp, err := svc.SignIn(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "access_token", resp.AccessToken)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		mUser.GetByEmailWithRolesFunc = func(ctx context.Context, email string, isActive *bool) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Email:        req.Email,
				PasswordHash: hash,
			}, nil
		}

		// Wrong password
		reqWrong := &models.SignInRequest{Email: req.Email, Password: "wrong"}

		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignInFailed, params.Action)
		}

		resp, err := svc.SignIn(ctx, reqWrong, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.ErrorIs(t, err, models.ErrInvalidCredentials)
		assert.Nil(t, resp)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mUser.GetByEmailWithRolesFunc = func(ctx context.Context, email string, isActive *bool) (*models.User, error) {
			return nil, errors.New("not found")
		}

		req := &models.SignInRequest{Email: "unknown@example.com", Password: "pwd"}

		resp, err := svc.SignIn(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.ErrorIs(t, err, models.ErrInvalidCredentials)
		assert.Nil(t, resp)
	})

	t.Run("TOTPRequired", func(t *testing.T) {
		mUser.GetByEmailWithRolesFunc = func(ctx context.Context, email string, isActive *bool) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Email:        req.Email,
				PasswordHash: hash,
				TOTPEnabled:  true,
			}, nil
		}

		mJWT.GenerateTwoFactorTokenFunc = func(user *models.User) (string, error) { return "2fa_token", nil }

		resp, err := svc.SignIn(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.NoError(t, err)
		assert.True(t, resp.Requires2FA)
		assert.Equal(t, "2fa_token", resp.TwoFactorToken)
		assert.Empty(t, resp.AccessToken)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	svc, mUser, mToken, _, mAudit, mJWT, mCache := setupAuthService()
	ctx := context.Background()
	refreshToken := "valid_refresh_token"
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mJWT.ValidateRefreshTokenFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		mCache.IsBlacklistedFunc = func(ctx context.Context, tokenHash string) (bool, error) { return false, nil }
		mToken.GetRefreshTokenFunc = func(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
			return &models.RefreshToken{
				UserID:    userID,
				TokenHash: "hash",
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		}
		mUser.GetByIDWithRolesFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: userID}, nil
		}
		mToken.RevokeRefreshTokenFunc = func(ctx context.Context, tokenHash string) error { return nil }
		mJWT.GenerateAccessTokenFunc = func(user *models.User) (string, error) { return "at", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User) (string, error) { return "rt", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionRefreshToken, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}

		resp, err := svc.RefreshToken(ctx, refreshToken, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "at", resp.AccessToken)
	})

	t.Run("RevokedToken", func(t *testing.T) {
		mJWT.ValidateRefreshTokenFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		mCache.IsBlacklistedFunc = func(ctx context.Context, tokenHash string) (bool, error) { return false, nil }
		mToken.GetRefreshTokenFunc = func(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
			return &models.RefreshToken{
				UserID:    userID,
				TokenHash: "hash",
				ExpiresAt: time.Now().Add(time.Hour),
				RevokedAt: utils.Ptr(time.Now()), // Revoked
			}, nil
		}
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionRefreshToken, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		resp, err := svc.RefreshToken(ctx, refreshToken, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.ErrorIs(t, err, models.ErrTokenRevoked)
		assert.Nil(t, resp)
	})
}

func TestAuthService_Logout(t *testing.T) {
	svc, _, mToken, _, mAudit, mJWT, mCache := setupAuthService()
	ctx := context.Background()
	accessToken := "access_token"
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mJWT.ExtractClaimsFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mCache.AddToBlacklistFunc = func(ctx context.Context, tokenHash string, expiration time.Duration) error { return nil }
		mToken.AddToBlacklistFunc = func(ctx context.Context, token *models.TokenBlacklist) error { return nil }
		mToken.RevokeAllUserTokensFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignOut, params.Action)
		}

		err := svc.Logout(ctx, accessToken, "1.1.1.1", "ua")
		assert.NoError(t, err)
	})
}

func TestAuthService_ChangePassword(t *testing.T) {
	svc, mUser, mToken, _, mAudit, _, _ := setupAuthService()
	ctx := context.Background()
	userID := uuid.New()
	oldPwd := "oldpassword"
	newPwd := "newpassword"
	hash, _ := utils.HashPassword(oldPwd, 10)

	t.Run("Success", func(t *testing.T) {
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: userID, PasswordHash: hash, IsActive: true}, nil
		}
		mUser.UpdatePasswordFunc = func(ctx context.Context, id uuid.UUID, hash string) error { return nil }
		mToken.RevokeAllUserTokensFunc = func(ctx context.Context, id uuid.UUID) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionChangePassword, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}

		err := svc.ChangePassword(ctx, userID, oldPwd, newPwd, "1.1.1.1", "ua")
		assert.NoError(t, err)
	})

	t.Run("WrongPassword", func(t *testing.T) {
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: userID, PasswordHash: hash, IsActive: true}, nil
		}

		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionChangePassword, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		err := svc.ChangePassword(ctx, userID, "wrong", newPwd, "1.1.1.1", "ua")
		assert.ErrorIs(t, err, models.ErrInvalidCredentials)
	})
}

func TestAuthService_InitPasswordlessRegistration(t *testing.T) {
	svc, mUser, _, _, mAudit, _, mCache := setupAuthService()
	ctx := context.Background()

	t.Run("Success_Email", func(t *testing.T) {
		email := "newuser@example.com"
		req := &models.InitPasswordlessRegistrationRequest{
			Email:    &email,
			FullName: "New User",
		}

		mUser.EmailExistsFunc = func(ctx context.Context, e string) (bool, error) { return false, nil }
		mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
		mCache.StorePendingRegistrationFunc = func(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error {
			assert.Equal(t, email, identifier)
			assert.Equal(t, email, data.Email)
			assert.Equal(t, "newuser", data.Username) // Auto-generated from email
			assert.Equal(t, "New User", data.FullName)
			return nil
		}
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
		}

		err := svc.InitPasswordlessRegistration(ctx, req, "1.1.1.1", "ua")
		assert.NoError(t, err)
	})

	t.Run("Success_Phone", func(t *testing.T) {
		phone := "+1234567890"
		req := &models.InitPasswordlessRegistrationRequest{
			Phone:    &phone,
			FullName: "Phone User",
		}

		mUser.PhoneExistsFunc = func(ctx context.Context, p string) (bool, error) { return false, nil }
		mUser.UsernameExistsFunc = func(ctx context.Context, username string) (bool, error) { return false, nil }
		mCache.StorePendingRegistrationFunc = func(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error {
			assert.Equal(t, "+1234567890", identifier)
			assert.Equal(t, "+1234567890", data.Phone)
			return nil
		}
		mAudit.LogFunc = func(params AuditLogParams) {}

		err := svc.InitPasswordlessRegistration(ctx, req, "1.1.1.1", "ua")
		assert.NoError(t, err)
	})

	t.Run("EmailAlreadyExists", func(t *testing.T) {
		email := "existing@example.com"
		req := &models.InitPasswordlessRegistrationRequest{
			Email: &email,
		}

		mUser.EmailExistsFunc = func(ctx context.Context, e string) (bool, error) { return true, nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		err := svc.InitPasswordlessRegistration(ctx, req, "1.1.1.1", "ua")
		assert.ErrorIs(t, err, models.ErrEmailAlreadyExists)
	})

	t.Run("PhoneAlreadyExists", func(t *testing.T) {
		phone := "+1234567890"
		req := &models.InitPasswordlessRegistrationRequest{
			Phone: &phone,
		}

		mUser.PhoneExistsFunc = func(ctx context.Context, p string) (bool, error) { return true, nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		err := svc.InitPasswordlessRegistration(ctx, req, "1.1.1.1", "ua")
		assert.ErrorIs(t, err, models.ErrPhoneAlreadyExists)
	})

	t.Run("NoEmailOrPhone", func(t *testing.T) {
		req := &models.InitPasswordlessRegistrationRequest{
			FullName: "No Contact",
		}

		err := svc.InitPasswordlessRegistration(ctx, req, "1.1.1.1", "ua")
		assert.Error(t, err)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.Code)
	})
}

func TestAuthService_CompletePasswordlessRegistration(t *testing.T) {
	svc, mUser, mToken, mRBAC, mAudit, mJWT, mCache := setupAuthService()
	ctx := context.Background()

	t.Run("Success_Email", func(t *testing.T) {
		email := "newuser@example.com"
		req := &models.CompletePasswordlessRegistrationRequest{
			Email: &email,
			Code:  "123456",
		}

		// Mock pending registration data
		mCache.GetPendingRegistrationFunc = func(ctx context.Context, identifier string) (*models.PendingRegistration, error) {
			return &models.PendingRegistration{
				Email:     email,
				Username:  "newuser",
				FullName:  "New User",
				CreatedAt: time.Now().Unix(),
			}, nil
		}
		mCache.DeletePendingRegistrationFunc = func(ctx context.Context, identifier string) error { return nil }

		// Mock user creation
		mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: uuid.New(), Name: "user"}, nil
		}
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			assert.Equal(t, email, user.Email)
			assert.Equal(t, "newuser", user.Username)
			assert.Empty(t, user.PasswordHash) // No password for passwordless
			assert.True(t, user.EmailVerified)
			return nil
		}
		mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }
		mUser.GetByIDWithRolesFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: id, Email: email, Username: "newuser", Roles: []models.Role{{Name: "user"}}}, nil
		}

		// Mock token generation
		mJWT.GenerateAccessTokenFunc = func(user *models.User) (string, error) { return "access_token", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User) (string, error) { return "refresh_token", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.CompletePasswordlessRegistration(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "access_token", resp.AccessToken)
		assert.Equal(t, "refresh_token", resp.RefreshToken)
	})

	t.Run("Success_Phone", func(t *testing.T) {
		phone := "+1234567890"
		req := &models.CompletePasswordlessRegistrationRequest{
			Phone: &phone,
			Code:  "123456",
		}

		// Mock pending registration data
		mCache.GetPendingRegistrationFunc = func(ctx context.Context, identifier string) (*models.PendingRegistration, error) {
			return &models.PendingRegistration{
				Phone:     "+1234567890",
				Username:  "1234567890",
				FullName:  "Phone User",
				CreatedAt: time.Now().Unix(),
			}, nil
		}
		mCache.DeletePendingRegistrationFunc = func(ctx context.Context, identifier string) error { return nil }

		// Mock user creation
		mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: uuid.New(), Name: "user"}, nil
		}
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			assert.Equal(t, "+1234567890", *user.Phone)
			assert.True(t, user.PhoneVerified)
			return nil
		}
		mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error { return nil }
		mUser.GetByIDWithRolesFunc = func(ctx context.Context, id uuid.UUID, isActive *bool) (*models.User, error) {
			return &models.User{ID: id, Phone: &phone, Roles: []models.Role{{Name: "user"}}}, nil
		}

		// Mock token generation
		mJWT.GenerateAccessTokenFunc = func(user *models.User) (string, error) { return "at", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User) (string, error) { return "rt", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.CompletePasswordlessRegistration(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("PendingNotFound", func(t *testing.T) {
		email := "unknown@example.com"
		req := &models.CompletePasswordlessRegistrationRequest{
			Email: &email,
			Code:  "123456",
		}

		mCache.GetPendingRegistrationFunc = func(ctx context.Context, identifier string) (*models.PendingRegistration, error) {
			return nil, errors.New("not found")
		}
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		resp, err := svc.CompletePasswordlessRegistration(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.Code)
	})

	t.Run("NoEmailOrPhone", func(t *testing.T) {
		req := &models.CompletePasswordlessRegistrationRequest{
			Code: "123456",
		}

		resp, err := svc.CompletePasswordlessRegistration(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.Code)
	})
}
