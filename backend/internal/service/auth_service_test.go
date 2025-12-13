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

	svc := NewAuthService(mUser, mToken, mRBAC, mAudit, mJWT, mCache, 10)
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
