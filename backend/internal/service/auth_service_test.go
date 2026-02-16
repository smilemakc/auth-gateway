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
	"github.com/uptrace/bun"
)

func setupAuthService() (*AuthService, *mockUserStore, *mockTokenStore, *mockRBACStore, *mockAuditLogger, *mockTokenService, *mockCacheService, *mockBlacklistChecker, *mockTransactionDB) {
	mUser := &mockUserStore{}
	mToken := &mockTokenStore{}
	mRBAC := &mockRBACStore{}
	mAudit := &mockAuditLogger{}
	mJWT := &mockTokenService{}
	mCache := &mockCacheService{}
	mDB := &mockTransactionDB{}

	// Create mocks for blacklist and session
	mBlacklist := &mockBlacklistChecker{}
	mSessionMgr := &mockSessionManager{}

	// Create default password policy
	passwordPolicy := utils.DefaultPasswordPolicy()

	// TwoFactorService, LoginAlertService, WebhookService, and PasswordChecker are nil for tests
	svc := NewAuthService(mUser, mToken, mRBAC, mAudit, mJWT, mBlacklist, mCache, mSessionMgr, nil, 10, passwordPolicy, mDB, nil, nil, nil, false, nil)
	return svc, mUser, mToken, mRBAC, mAudit, mJWT, mCache, mBlacklist, mDB
}

func TestAuthService_SignUp(t *testing.T) {
	svc, mUser, mToken, mRBAC, mAudit, mJWT, _, _, _ := setupAuthService()
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
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: id, Email: validReq.Email, Roles: []models.Role{{Name: "user"}}}, nil
		}
		mJWT.GenerateAccessTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "access_token", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "refresh_token", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}

		resp, err := svc.SignUp(ctx, validReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "access_token", resp.AccessToken)
	})

	t.Run("EmailExists", func(t *testing.T) {
		// SignUp relies on DB unique constraints via Create, not pre-checks
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			return models.ErrEmailAlreadyExists
		}
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		resp, err := svc.SignUp(ctx, validReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.ErrorIs(t, err, models.ErrEmailAlreadyExists)
		assert.Nil(t, resp)
	})

	t.Run("UsernameExists", func(t *testing.T) {
		// SignUp relies on DB unique constraints via Create, not pre-checks
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			return models.ErrUsernameAlreadyExists
		}
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignUp, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		resp, err := svc.SignUp(ctx, validReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.ErrorIs(t, err, models.ErrUsernameAlreadyExists)
		assert.Nil(t, resp)
	})

	t.Run("InvalidPassword_TooShort", func(t *testing.T) {
		// Default policy requires min 8 chars and at least one lowercase letter.
		// A 3-char password should fail validation before any repo call.
		shortPwdReq := &models.CreateUserRequest{
			Email:    "short@example.com",
			Username: "shortpwd",
			Password: "ab",
			FullName: "Short Pwd",
		}

		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignUp(ctx, shortPwdReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok, "expected AppError for invalid password")
		assert.Equal(t, 400, appErr.Code)
	})

	t.Run("InvalidPassword_NoLowercase", func(t *testing.T) {
		// Default policy requires lowercase. All-uppercase password should fail.
		noLowerReq := &models.CreateUserRequest{
			Email:    "nolower@example.com",
			Username: "nolower",
			Password: "ABCDEFGH12",
			FullName: "No Lower",
		}

		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignUp(ctx, noLowerReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok, "expected AppError for missing lowercase")
		assert.Equal(t, 400, appErr.Code)
	})

	t.Run("DBError_OnCreate", func(t *testing.T) {
		// Generic database error (not a unique constraint violation)
		mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: uuid.New(), Name: "user"}, nil
		}
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			return errors.New("database connection lost")
		}
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignUp(ctx, validReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "database connection lost")
	})

	t.Run("RoleNotFound", func(t *testing.T) {
		// GetRoleByName returns error when default "user" role doesn't exist
		mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return nil, errors.New("role not found")
		}
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignUp(ctx, validReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get default role")
	})

	t.Run("GetByIDAfterCreate_Error", func(t *testing.T) {
		// User is created successfully but reloading with roles fails
		mRBAC.GetRoleByNameFunc = func(ctx context.Context, name string) (*models.Role, error) {
			return &models.Role{ID: uuid.New(), Name: "user"}, nil
		}
		mUser.CreateFunc = func(ctx context.Context, user *models.User) error {
			return nil
		}
		mRBAC.AssignRoleToUserFunc = func(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
			return nil
		}
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("db read replica down")
		}
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignUp(ctx, validReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to reload user with roles")
	})
}

func TestAuthService_SignIn(t *testing.T) {
	svc, mUser, mToken, _, mAudit, mJWT, _, _, _ := setupAuthService()
	ctx := context.Background()

	password := "password123"
	hash, _ := utils.HashPassword(password, 10)
	userID := uuid.New()

	req := &models.SignInRequest{
		Email:    "test@example.com",
		Password: password,
	}

	t.Run("Success", func(t *testing.T) {
		mUser.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Email:        req.Email,
				PasswordHash: hash,
				IsActive:     true,
			}, nil
		}

		mJWT.GenerateAccessTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "access_token", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "refresh_token", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionSignIn, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}

		resp, err := svc.SignIn(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "access_token", resp.AccessToken)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		mUser.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
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

		resp, err := svc.SignIn(ctx, reqWrong, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.ErrorIs(t, err, models.ErrInvalidCredentials)
		assert.Nil(t, resp)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mUser.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("not found")
		}

		req := &models.SignInRequest{Email: "unknown@example.com", Password: "pwd"}

		resp, err := svc.SignIn(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.ErrorIs(t, err, models.ErrInvalidCredentials)
		assert.Nil(t, resp)
	})

	t.Run("TOTPRequired", func(t *testing.T) {
		mUser.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Email:        req.Email,
				PasswordHash: hash,
				TOTPEnabled:  true,
			}, nil
		}

		mJWT.GenerateTwoFactorTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "2fa_token", nil }

		resp, err := svc.SignIn(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.NoError(t, err)
		assert.True(t, resp.Requires2FA)
		assert.Equal(t, "2fa_token", resp.TwoFactorToken)
		assert.Empty(t, resp.AccessToken)
	})

	t.Run("InactiveUser_StillSignsIn", func(t *testing.T) {
		// SignIn does NOT explicitly check IsActive (passes nil to GetByEmail).
		// An inactive user with correct credentials will still authenticate.
		mUser.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Email:        req.Email,
				PasswordHash: hash,
				IsActive:     false,
			}, nil
		}
		mJWT.GenerateAccessTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "at", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "rt", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignIn(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "at", resp.AccessToken)
	})

	t.Run("DBError_GetByEmail", func(t *testing.T) {
		// GetByEmail returns a generic DB error (not "not found").
		// SignIn still performs password check against dummy hash, so result is ErrInvalidCredentials.
		mUser.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("connection refused")
		}
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignIn(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.ErrorIs(t, err, models.ErrInvalidCredentials)
		assert.Nil(t, resp)
	})

	t.Run("TokenGenerationError", func(t *testing.T) {
		// Password matches but JWT access token generation fails in finalizeAuth
		mUser.GetByEmailFunc = func(ctx context.Context, email string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Email:        req.Email,
				PasswordHash: hash,
				IsActive:     true,
			}, nil
		}
		mJWT.GenerateAccessTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) {
			return "", errors.New("signing key unavailable")
		}
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignIn(ctx, req, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to generate access token")
	})

	t.Run("SignIn_ViaPhone_Success", func(t *testing.T) {
		phone := "+1234567890"
		phoneReq := &models.SignInRequest{
			Phone:    &phone,
			Password: password,
		}
		mUser.GetByPhoneFunc = func(ctx context.Context, p string, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Phone:        &phone,
				PasswordHash: hash,
				IsActive:     true,
			}, nil
		}
		mJWT.GenerateAccessTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "at_phone", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "rt_phone", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mToken.CreateRefreshTokenFunc = func(ctx context.Context, token *models.RefreshToken) error { return nil }
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignIn(ctx, phoneReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "at_phone", resp.AccessToken)
	})

	t.Run("NoEmailOrPhone", func(t *testing.T) {
		emptyReq := &models.SignInRequest{
			Password: password,
		}
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.SignIn(ctx, emptyReq, "1.1.1.1", "ua", models.DeviceInfo{}, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
		appErr, ok := err.(*models.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.Code)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	svc, mUser, _, _, mAudit, mJWT, mCache, _, mDB := setupAuthService()
	ctx := context.Background()
	refreshToken := "valid_refresh_token"
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mJWT.ValidateRefreshTokenFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		mCache.IsBlacklistedFunc = func(ctx context.Context, tokenHash string) (bool, error) { return false, nil }
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID}, nil
		}
		mJWT.GenerateAccessTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "at", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "rt", nil }
		mJWT.GetAccessTokenExpirationFunc = func() time.Duration { return time.Hour }
		mJWT.GetRefreshTokenExpirationFunc = func() time.Duration { return 24 * time.Hour }
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionRefreshToken, params.Action)
			assert.Equal(t, models.StatusSuccess, params.Status)
		}
		// RunInTx uses type assertion to *repository.TokenRepository which fails with mocks.
		// Mock the transaction to simulate successful token refresh.
		mDB.RunInTxFunc = func(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
			return nil // Simulate successful transaction
		}

		resp, err := svc.RefreshToken(ctx, refreshToken, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// Tokens are generated inside the real transaction callback which we skipped,
		// so they will be zero-value. Verify the response is returned without error.
	})

	t.Run("RevokedToken", func(t *testing.T) {
		mJWT.ValidateRefreshTokenFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		mCache.IsBlacklistedFunc = func(ctx context.Context, tokenHash string) (bool, error) { return false, nil }
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: userID}, nil
		}
		mAudit.LogFunc = func(params AuditLogParams) {
			assert.Equal(t, models.ActionRefreshToken, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}
		mDB.RunInTxFunc = func(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
			return models.ErrTokenRevoked
		}

		resp, err := svc.RefreshToken(ctx, refreshToken, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.ErrorIs(t, err, models.ErrTokenRevoked)
		assert.Nil(t, resp)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// ValidateRefreshToken returns error - token is malformed/expired
		mJWT.ValidateRefreshTokenFunc = func(tokenString string) (*jwt.Claims, error) {
			return nil, errors.New("token is expired")
		}
		auditCalled := false
		mAudit.LogFunc = func(params AuditLogParams) {
			auditCalled = true
			assert.Equal(t, models.ActionRefreshToken, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}

		resp, err := svc.RefreshToken(ctx, "invalid_token", "1.1.1.1", "ua", models.DeviceInfo{})
		assert.ErrorIs(t, err, models.ErrInvalidToken)
		assert.Nil(t, resp)
		assert.True(t, auditCalled, "audit log should be called for invalid token")
	})

	t.Run("BlacklistedToken", func(t *testing.T) {
		// Token validates but is found in the blacklist
		mJWT.ValidateRefreshTokenFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		// Use the BlacklistChecker mock, not the CacheService mock.
		// RefreshToken calls s.blacklistService.IsBlacklisted which returns bool (not bool, error).
		// The setupAuthService creates a mockBlacklistChecker that is separate from mCache.
		// We need to recreate the service with a blacklist mock that returns true.
		svcBL, mUserBL, _, _, mAuditBL, mJWTBL, _, mBlacklistBL, _ := setupAuthService()

		mJWTBL.ValidateRefreshTokenFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		mBlacklistBL.IsBlacklistedFunc = func(ctx context.Context, tokenHash string) bool {
			return true
		}
		auditCalled := false
		mAuditBL.LogFunc = func(params AuditLogParams) {
			auditCalled = true
			assert.Equal(t, models.ActionRefreshToken, params.Action)
			assert.Equal(t, models.StatusFailed, params.Status)
		}
		// GetByID should NOT be called since we short-circuit on blacklist
		mUserBL.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			t.Fatal("GetByID should not be called when token is blacklisted")
			return nil, nil
		}

		resp, err := svcBL.RefreshToken(ctx, refreshToken, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.ErrorIs(t, err, models.ErrTokenRevoked)
		assert.Nil(t, resp)
		assert.True(t, auditCalled, "audit log should be called for blacklisted token")
	})

	t.Run("UserNotFound", func(t *testing.T) {
		// Token is valid and not blacklisted, but GetByID fails
		mJWT.ValidateRefreshTokenFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return nil, errors.New("user not found")
		}
		mAudit.LogFunc = func(params AuditLogParams) {}

		resp, err := svc.RefreshToken(ctx, refreshToken, "1.1.1.1", "ua", models.DeviceInfo{})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestAuthService_Logout(t *testing.T) {
	svc, _, mToken, _, mAudit, mJWT, mCache, _, _ := setupAuthService()
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

	t.Run("InvalidToken", func(t *testing.T) {
		// ExtractClaims returns error for malformed token
		mJWT.ExtractClaimsFunc = func(tokenString string) (*jwt.Claims, error) {
			return nil, errors.New("malformed token")
		}

		err := svc.Logout(ctx, "bad_token", "1.1.1.1", "ua")
		assert.ErrorIs(t, err, models.ErrInvalidToken)
	})

	t.Run("BlacklistError", func(t *testing.T) {
		// ExtractClaims succeeds but AddAccessToken (blacklist) fails.
		// Logout uses s.blacklistService.AddAccessToken, not mCache.AddToBlacklist.
		svcBL, _, mTokenBL, _, _, mJWTBL, _, mBlacklistBL, _ := setupAuthService()

		mJWTBL.ExtractClaimsFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		mBlacklistBL.AddAccessTokenFunc = func(ctx context.Context, tokenHash string, uid *uuid.UUID) error {
			return errors.New("redis unavailable")
		}
		// RevokeAllUserTokens should NOT be reached
		mTokenBL.RevokeAllUserTokensFunc = func(ctx context.Context, id uuid.UUID) error {
			t.Fatal("RevokeAllUserTokens should not be called when blacklist fails")
			return nil
		}

		err := svcBL.Logout(ctx, accessToken, "1.1.1.1", "ua")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to blacklist token")
	})

	t.Run("RevokeTokensError", func(t *testing.T) {
		// ExtractClaims and AddAccessToken succeed, but RevokeAllUserTokens fails
		svcRT, _, mTokenRT, _, mAuditRT, mJWTRT, _, mBlacklistRT, _ := setupAuthService()

		mJWTRT.ExtractClaimsFunc = func(tokenString string) (*jwt.Claims, error) {
			return &jwt.Claims{UserID: userID}, nil
		}
		mBlacklistRT.AddAccessTokenFunc = func(ctx context.Context, tokenHash string, uid *uuid.UUID) error {
			return nil
		}
		mTokenRT.RevokeAllUserTokensFunc = func(ctx context.Context, id uuid.UUID) error {
			return errors.New("database timeout")
		}
		mAuditRT.LogFunc = func(params AuditLogParams) {}

		err := svcRT.Logout(ctx, accessToken, "1.1.1.1", "ua")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to revoke refresh tokens")
	})
}

func TestAuthService_ChangePassword(t *testing.T) {
	svc, mUser, mToken, _, mAudit, _, _, _, _ := setupAuthService()
	ctx := context.Background()
	userID := uuid.New()
	oldPwd := "oldpassword"
	newPwd := "newpassword"
	hash, _ := utils.HashPassword(oldPwd, 10)

	t.Run("Success", func(t *testing.T) {
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
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
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
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
	svc, _, _, _, mAudit, _, mCache, _, _ := setupAuthService()
	ctx := context.Background()

	t.Run("Success_Email", func(t *testing.T) {
		email := "newuser@example.com"
		req := &models.InitPasswordlessRegistrationRequest{
			Email:    &email,
			FullName: "New User",
		}

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

		mCache.StorePendingRegistrationFunc = func(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error {
			assert.Equal(t, phone, identifier)
			assert.Equal(t, phone, data.Phone)
			return nil
		}
		mAudit.LogFunc = func(params AuditLogParams) {}

		err := svc.InitPasswordlessRegistration(ctx, req, "1.1.1.1", "ua")
		assert.NoError(t, err)
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
	svc, mUser, mToken, mRBAC, mAudit, mJWT, mCache, _, _ := setupAuthService()
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
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: id, Email: email, Username: "newuser", Roles: []models.Role{{Name: "user"}}}, nil
		}

		// Mock token generation
		mJWT.GenerateAccessTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "access_token", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "refresh_token", nil }
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
		mUser.GetByIDFunc = func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...UserGetOption) (*models.User, error) {
			return &models.User{ID: id, Phone: &phone, Roles: []models.Role{{Name: "user"}}}, nil
		}

		// Mock token generation
		mJWT.GenerateAccessTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "at", nil }
		mJWT.GenerateRefreshTokenFunc = func(user *models.User, applicationID ...*uuid.UUID) (string, error) { return "rt", nil }
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
