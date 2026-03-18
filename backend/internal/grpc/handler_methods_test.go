package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	pb "github.com/smilemakc/auth-gateway/proto"
)

// newTestJWTService creates a JWT service for testing with long-enough secrets
func newTestJWTService() *jwt.Service {
	return jwt.NewService(
		"test-access-secret-key-min-32-chars!!",
		"test-refresh-secret-key-min-32-chars!!",
		15*time.Minute,
		7*24*time.Hour,
	)
}

// newTestUser creates a test user with default fields and provided roles
func newTestUser(id uuid.UUID, roles ...string) *models.User {
	modelRoles := make([]models.Role, len(roles))
	for i, r := range roles {
		modelRoles[i] = models.Role{Name: r}
	}
	return &models.User{
		ID:            id,
		Email:         "test@example.com",
		Username:      "testuser",
		FullName:      "Test User",
		IsActive:      true,
		EmailVerified: true,
		Roles:         modelRoles,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// ===================== ValidateToken Tests =====================

func TestValidateToken_ShouldReturnInvalid_WhenEmptyTokenWithMocks(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	resp, err := h.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		AccessToken: "",
	})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.Equal(t, "access_token is required", resp.ErrorMessage)
}

func TestValidateToken_ShouldReturnValid_WhenValidJWT(t *testing.T) {
	jwtSvc := newTestJWTService()
	userID := uuid.New()
	user := newTestUser(userID, "admin")

	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)

	redisMock := &mockRedisServicerGRPC{
		IsBlacklistedFunc: func(ctx context.Context, tokenHash string) (bool, error) {
			return false, nil
		},
	}

	h := newTestAuthHandlerV2(jwtSvc, func(h *AuthHandlerV2) {
		h.redis = redisMock
	})

	resp, err := h.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		AccessToken: token,
	})

	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, userID.String(), resp.UserId)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "testuser", resp.Username)
	assert.Contains(t, resp.Roles, "admin")
	assert.True(t, resp.IsActive)
	assert.True(t, resp.ExpiresAt > 0)
}

func TestValidateToken_ShouldReturnInvalid_WhenBlacklistedJWT(t *testing.T) {
	jwtSvc := newTestJWTService()
	user := newTestUser(uuid.New(), "user")

	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)

	redisMock := &mockRedisServicerGRPC{
		IsBlacklistedFunc: func(ctx context.Context, tokenHash string) (bool, error) {
			return true, nil
		},
	}

	h := newTestAuthHandlerV2(jwtSvc, func(h *AuthHandlerV2) {
		h.redis = redisMock
	})

	resp, err := h.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		AccessToken: token,
	})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.Equal(t, "token is blacklisted", resp.ErrorMessage)
}

func TestValidateToken_ShouldReturnValid_WhenValidAPIKey(t *testing.T) {
	userID := uuid.New()
	user := newTestUser(userID, "admin")
	apiKey := &models.APIKey{ID: uuid.New(), UserID: userID, IsActive: true}

	apiKeyMock := &mockAPIKeyServicerGRPC{
		ValidateAPIKeyFunc: func(ctx context.Context, plainKey string) (*models.APIKey, *models.User, error) {
			return apiKey, user, nil
		},
	}
	userRepoMock := &mockUserStoreGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			return user, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.apiKeyService = apiKeyMock
		h.userRepo = userRepoMock
	})

	resp, err := h.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		AccessToken: "agw_test_valid_key_abc123",
	})

	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, userID.String(), resp.UserId)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Contains(t, resp.Roles, "admin")
}

func TestValidateToken_ShouldReturnInvalid_WhenAPIKeyInvalid(t *testing.T) {
	apiKeyMock := &mockAPIKeyServicerGRPC{
		ValidateAPIKeyFunc: func(ctx context.Context, plainKey string) (*models.APIKey, *models.User, error) {
			return nil, nil, errors.New("invalid API key")
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.apiKeyService = apiKeyMock
	})

	resp, err := h.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		AccessToken: "agw_invalid_key",
	})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.Contains(t, resp.ErrorMessage, "invalid API key")
}

func TestValidateToken_ShouldReturnInvalid_WhenMalformedJWT(t *testing.T) {
	redisMock := &mockRedisServicerGRPC{
		IsBlacklistedFunc: func(ctx context.Context, tokenHash string) (bool, error) {
			return false, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.redis = redisMock
	})

	resp, err := h.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		AccessToken: "this.is.not.a.valid.jwt",
	})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
}

// ===================== GetUser Tests =====================

func TestGetUser_ShouldReturnUser_WhenExists(t *testing.T) {
	userID := uuid.New()
	user := newTestUser(userID, "admin", "editor")

	userRepoMock := &mockUserStoreGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			return user, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.userRepo = userRepoMock
	})

	resp, err := h.GetUser(context.Background(), &pb.GetUserRequest{
		UserId: userID.String(),
	})

	require.NoError(t, err)
	require.NotNil(t, resp.User)
	assert.Equal(t, userID.String(), resp.User.Id)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "testuser", resp.User.Username)
	assert.Equal(t, "Test User", resp.User.FullName)
	assert.True(t, resp.User.IsActive)
	assert.True(t, resp.User.EmailVerified)
	assert.Len(t, resp.User.Roles, 2)
	assert.Contains(t, resp.User.Roles, "admin")
	assert.Contains(t, resp.User.Roles, "editor")
}

func TestGetUser_ShouldReturnError_WhenEmptyUserIDWithMocks(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.GetUser(context.Background(), &pb.GetUserRequest{UserId: ""})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestGetUser_ShouldReturnNotFound_WhenUserDoesNotExist(t *testing.T) {
	userRepoMock := &mockUserStoreGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			return nil, models.ErrUserNotFound
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.userRepo = userRepoMock
	})

	_, err := h.GetUser(context.Background(), &pb.GetUserRequest{
		UserId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// ===================== CheckPermission Tests =====================

func TestCheckPermission_ShouldReturnAllowed_WhenUserHasPermission(t *testing.T) {
	userID := uuid.New()

	rbacMock := &mockRBACStoreGRPC{
		GetUserRolesFunc: func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return []models.Role{
				{
					Name: "admin",
					Permissions: []models.Permission{
						{Resource: "users", Action: "read"},
						{Resource: "users", Action: "write"},
					},
				},
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.rbacRepo = rbacMock
	})

	resp, err := h.CheckPermission(context.Background(), &pb.CheckPermissionRequest{
		UserId:   userID.String(),
		Resource: "users",
		Action:   "read",
	})

	require.NoError(t, err)
	assert.True(t, resp.Allowed)
	assert.Equal(t, "admin", resp.Role)
}

func TestCheckPermission_ShouldReturnDenied_WhenUserLacksPermission(t *testing.T) {
	userID := uuid.New()

	rbacMock := &mockRBACStoreGRPC{
		GetUserRolesFunc: func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return []models.Role{
				{
					Name: "viewer",
					Permissions: []models.Permission{
						{Resource: "users", Action: "read"},
					},
				},
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.rbacRepo = rbacMock
	})

	resp, err := h.CheckPermission(context.Background(), &pb.CheckPermissionRequest{
		UserId:   userID.String(),
		Resource: "users",
		Action:   "delete",
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed)
}

func TestCheckPermission_ShouldReturnError_WhenEmptyFields(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.CheckPermission(context.Background(), &pb.CheckPermissionRequest{
		UserId:   "",
		Resource: "users",
		Action:   "read",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestCheckPermission_ShouldReturnNotAllowed_WhenNoRoles(t *testing.T) {
	rbacMock := &mockRBACStoreGRPC{
		GetUserRolesFunc: func(ctx context.Context, uid uuid.UUID) ([]models.Role, error) {
			return []models.Role{}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.rbacRepo = rbacMock
	})

	resp, err := h.CheckPermission(context.Background(), &pb.CheckPermissionRequest{
		UserId:   uuid.New().String(),
		Resource: "users",
		Action:   "read",
	})

	require.NoError(t, err)
	assert.False(t, resp.Allowed)
	assert.Equal(t, "user has no roles", resp.ErrorMessage)
}

// ===================== Login Tests =====================

func TestLogin_ShouldReturnTokens_WhenValidCredentials(t *testing.T) {
	userID := uuid.New()
	user := newTestUser(userID, "user")

	authMock := &mockAuthServicerGRPC{
		SignInFunc: func(ctx context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
			return &models.AuthResponse{
				AccessToken:  "access-token-abc",
				RefreshToken: "refresh-token-xyz",
				User:         user,
				ExpiresIn:    900,
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.authService = authMock
	})

	resp, err := h.Login(context.Background(), &pb.LoginRequest{
		Email:    "test@example.com",
		Password: "SecurePass123",
	})

	require.NoError(t, err)
	assert.Equal(t, "access-token-abc", resp.AccessToken)
	assert.Equal(t, "refresh-token-xyz", resp.RefreshToken)
	assert.Equal(t, int64(900), resp.ExpiresIn)
	require.NotNil(t, resp.User)
	assert.Equal(t, userID.String(), resp.User.Id)
	assert.Equal(t, "test@example.com", resp.User.Email)
}

func TestLogin_ShouldReturnError_WhenInvalidCredentials(t *testing.T) {
	authMock := &mockAuthServicerGRPC{
		SignInFunc: func(ctx context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
			return nil, models.ErrInvalidCredentials
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.authService = authMock
	})

	_, err := h.Login(context.Background(), &pb.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong-password",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid credentials")
}

func TestLogin_ShouldReturnError_WhenEmptyFields(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.Login(context.Background(), &pb.LoginRequest{
		Email:    "",
		Phone:    "",
		Password: "SecurePass123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "either email or phone is required")
}

func TestLogin_ShouldReturn2FARequired_WhenTwoFactorEnabled(t *testing.T) {
	authMock := &mockAuthServicerGRPC{
		SignInFunc: func(ctx context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
			return &models.AuthResponse{
				Requires2FA: true,
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.authService = authMock
	})

	resp, err := h.Login(context.Background(), &pb.LoginRequest{
		Email:    "test@example.com",
		Password: "SecurePass123",
	})

	require.NoError(t, err)
	assert.Contains(t, resp.ErrorMessage, "2FA required")
}

// ===================== CreateUser Tests =====================

func TestCreateUser_ShouldReturnUser_WhenValidData(t *testing.T) {
	userID := uuid.New()
	user := newTestUser(userID, "user")

	authMock := &mockAuthServicerGRPC{
		SignUpFunc: func(ctx context.Context, req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
			return &models.AuthResponse{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				User:         user,
				ExpiresIn:    900,
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.authService = authMock
	})

	resp, err := h.CreateUser(context.Background(), &pb.CreateUserRequest{
		Email:    "new@example.com",
		Password: "SecurePass123",
		Username: "newuser",
	})

	require.NoError(t, err)
	assert.Equal(t, "new-access-token", resp.AccessToken)
	assert.Equal(t, "new-refresh-token", resp.RefreshToken)
	require.NotNil(t, resp.User)
	assert.Equal(t, userID.String(), resp.User.Id)
}

func TestCreateUser_ShouldReturnError_WhenEmptyEmail(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.CreateUser(context.Background(), &pb.CreateUserRequest{
		Email:    "",
		Phone:    "",
		Password: "SecurePass123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "either email or phone is required")
}

func TestCreateUser_ShouldReturnError_WhenServiceError(t *testing.T) {
	authMock := &mockAuthServicerGRPC{
		SignUpFunc: func(ctx context.Context, req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
			return nil, models.ErrEmailAlreadyExists
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.authService = authMock
	})

	_, err := h.CreateUser(context.Background(), &pb.CreateUserRequest{
		Email:    "dup@example.com",
		Password: "SecurePass123",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Email already exists")
}

// ===================== SendOTP / VerifyOTP Tests =====================

func TestSendOTP_ShouldReturnSuccess_WhenOTPSent(t *testing.T) {
	otpMock := &mockOTPServicerGRPC{
		SendOTPFunc: func(ctx context.Context, req *models.SendOTPRequest) error {
			return nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.otpService = otpMock
	})

	resp, err := h.SendOTP(context.Background(), &pb.SendOTPRequest{
		Email:   "user@example.com",
		OtpType: pb.OTPType_OTP_TYPE_VERIFICATION,
	})

	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "OTP sent successfully", resp.Message)
	assert.Equal(t, int32(600), resp.ExpiresIn)
}

func TestSendOTP_ShouldReturnError_WhenEmailEmptyWithMocks(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.SendOTP(context.Background(), &pb.SendOTPRequest{
		Email: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestSendOTP_ShouldReturnFailed_WhenServiceError(t *testing.T) {
	otpMock := &mockOTPServicerGRPC{
		SendOTPFunc: func(ctx context.Context, req *models.SendOTPRequest) error {
			return &models.AppError{Code: 429, Message: "rate limit exceeded"}
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.otpService = otpMock
	})

	resp, err := h.SendOTP(context.Background(), &pb.SendOTPRequest{
		Email:   "user@example.com",
		OtpType: pb.OTPType_OTP_TYPE_VERIFICATION,
	})

	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "rate limit exceeded", resp.ErrorMessage)
}

func TestVerifyOTP_ShouldReturnValid_WhenCodeCorrect(t *testing.T) {
	otpMock := &mockOTPServicerGRPC{
		VerifyOTPFunc: func(ctx context.Context, req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
			return &models.VerifyOTPResponse{Valid: true}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.otpService = otpMock
	})

	resp, err := h.VerifyOTP(context.Background(), &pb.VerifyOTPRequest{
		Email:   "user@example.com",
		Code:    "123456",
		OtpType: pb.OTPType_OTP_TYPE_VERIFICATION,
	})

	require.NoError(t, err)
	assert.True(t, resp.Valid)
}

func TestVerifyOTP_ShouldReturnInvalid_WhenCodeWrong(t *testing.T) {
	otpMock := &mockOTPServicerGRPC{
		VerifyOTPFunc: func(ctx context.Context, req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
			return nil, &models.AppError{Code: 400, Message: "invalid OTP code"}
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.otpService = otpMock
	})

	resp, err := h.VerifyOTP(context.Background(), &pb.VerifyOTPRequest{
		Email:   "user@example.com",
		Code:    "000000",
		OtpType: pb.OTPType_OTP_TYPE_VERIFICATION,
	})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.Equal(t, "invalid OTP code", resp.ErrorMessage)
}

// ===================== SyncUsers Tests =====================

func TestSyncUsers_ShouldReturnUsers_WhenValidTimestamp(t *testing.T) {
	now := time.Now()

	adminMock := &mockAdminServicerGRPC{
		SyncUsersFunc: func(ctx context.Context, updatedAfter time.Time, appID *uuid.UUID, limit, offset int) (*models.SyncUsersResponse, error) {
			return &models.SyncUsersResponse{
				Users: []models.SyncUserResponse{
					{
						ID:       uuid.New(),
						Email:    "user1@example.com",
						Username: "user1",
						IsActive: true,
						UpdatedAt: now,
					},
				},
				Total:         1,
				HasMore:       false,
				SyncTimestamp: now.Format(time.RFC3339),
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.adminService = adminMock
	})

	resp, err := h.SyncUsers(context.Background(), &pb.SyncUsersRequest{
		UpdatedAfter: now.Add(-1 * time.Hour).Format(time.RFC3339),
		Limit:        100,
	})

	require.NoError(t, err)
	assert.Empty(t, resp.ErrorMessage)
	assert.Len(t, resp.Users, 1)
	assert.Equal(t, "user1@example.com", resp.Users[0].Email)
	assert.Equal(t, int32(1), resp.Total)
	assert.False(t, resp.HasMore)
}

func TestSyncUsers_ShouldReturnError_WhenServiceFails(t *testing.T) {
	adminMock := &mockAdminServicerGRPC{
		SyncUsersFunc: func(ctx context.Context, updatedAfter time.Time, appID *uuid.UUID, limit, offset int) (*models.SyncUsersResponse, error) {
			return nil, errors.New("database error")
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.adminService = adminMock
	})

	resp, err := h.SyncUsers(context.Background(), &pb.SyncUsersRequest{
		UpdatedAfter: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
	})

	require.NoError(t, err)
	assert.Equal(t, "database error", resp.ErrorMessage)
}

// ===================== CreateTokenExchange / RedeemTokenExchange Tests =====================

func TestCreateTokenExchange_ShouldReturnCode_WhenValid(t *testing.T) {
	exchangeMock := &mockTokenExchangeServicerGRPC{
		CreateExchangeFunc: func(ctx context.Context, req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error) {
			return &models.CreateTokenExchangeResponse{
				ExchangeCode: "exchange-code-123",
				ExpiresAt:    time.Now().Add(5 * time.Minute),
				RedirectURL:  "https://target-app.com/auth",
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.tokenExchangeService = exchangeMock
	})

	resp, err := h.CreateTokenExchange(context.Background(), &pb.CreateTokenExchangeGrpcRequest{
		AccessToken:         "valid-access-token",
		TargetApplicationId: uuid.New().String(),
	})

	require.NoError(t, err)
	assert.Empty(t, resp.ErrorMessage)
	assert.Equal(t, "exchange-code-123", resp.ExchangeCode)
	assert.NotEmpty(t, resp.ExpiresAt)
	assert.Equal(t, "https://target-app.com/auth", resp.RedirectUrl)
}

func TestCreateTokenExchange_ShouldReturnError_WhenEmptyToken(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	resp, err := h.CreateTokenExchange(context.Background(), &pb.CreateTokenExchangeGrpcRequest{
		AccessToken:         "",
		TargetApplicationId: uuid.New().String(),
	})

	require.NoError(t, err)
	assert.Equal(t, "access_token is required", resp.ErrorMessage)
}

func TestCreateTokenExchange_ShouldReturnError_WhenServiceFails(t *testing.T) {
	exchangeMock := &mockTokenExchangeServicerGRPC{
		CreateExchangeFunc: func(ctx context.Context, req *models.CreateTokenExchangeRequest, sourceAppID *uuid.UUID) (*models.CreateTokenExchangeResponse, error) {
			return nil, &models.AppError{Code: 400, Message: "invalid token"}
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.tokenExchangeService = exchangeMock
	})

	resp, err := h.CreateTokenExchange(context.Background(), &pb.CreateTokenExchangeGrpcRequest{
		AccessToken:         "bad-token",
		TargetApplicationId: uuid.New().String(),
	})

	require.NoError(t, err)
	assert.Equal(t, "invalid token", resp.ErrorMessage)
}

func TestRedeemTokenExchange_ShouldReturnTokens_WhenValid(t *testing.T) {
	userID := uuid.New()

	exchangeMock := &mockTokenExchangeServicerGRPC{
		RedeemExchangeFunc: func(ctx context.Context, req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error) {
			return &models.RedeemTokenExchangeResponse{
				AccessToken:   "new-access-token",
				RefreshToken:  "new-refresh-token",
				User:          &models.User{ID: userID, Email: "user@example.com"},
				ApplicationID: "target-app-id",
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.tokenExchangeService = exchangeMock
	})

	resp, err := h.RedeemTokenExchange(context.Background(), &pb.RedeemTokenExchangeGrpcRequest{
		ExchangeCode: "valid-code",
	})

	require.NoError(t, err)
	assert.Empty(t, resp.ErrorMessage)
	assert.Equal(t, "new-access-token", resp.AccessToken)
	assert.Equal(t, "new-refresh-token", resp.RefreshToken)
	assert.Equal(t, userID.String(), resp.UserId)
	assert.Equal(t, "user@example.com", resp.Email)
	assert.Equal(t, "target-app-id", resp.ApplicationId)
}

func TestRedeemTokenExchange_ShouldReturnError_WhenInvalidCode(t *testing.T) {
	exchangeMock := &mockTokenExchangeServicerGRPC{
		RedeemExchangeFunc: func(ctx context.Context, req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error) {
			return nil, &models.AppError{Code: 404, Message: "exchange code not found"}
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.tokenExchangeService = exchangeMock
	})

	resp, err := h.RedeemTokenExchange(context.Background(), &pb.RedeemTokenExchangeGrpcRequest{
		ExchangeCode: "invalid-code",
	})

	require.NoError(t, err)
	assert.Equal(t, "exchange code not found", resp.ErrorMessage)
}

// ===================== BanUser / UnbanUser Tests =====================

func TestBanUser_ShouldReturnSuccess_WhenValidRequest(t *testing.T) {
	appMock := &mockApplicationServicerGRPC{
		BanUserFunc: func(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error {
			return nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.BanUser(context.Background(), &pb.BanUserRequest{
		UserId:        uuid.New().String(),
		ApplicationId: uuid.New().String(),
		BannedBy:      uuid.New().String(),
		Reason:        "spamming",
	})

	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "User banned", resp.Message)
}

func TestBanUser_ShouldReturnError_WhenInvalidUserID(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.BanUser(context.Background(), &pb.BanUserRequest{
		UserId:        "not-a-uuid",
		ApplicationId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user_id format")
}

func TestBanUser_ShouldReturnError_WhenMissingIDsWithMocks(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.BanUser(context.Background(), &pb.BanUserRequest{
		UserId:        "",
		ApplicationId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id and application_id are required")
}

func TestUnbanUser_ShouldReturnSuccess_WhenValidRequest(t *testing.T) {
	appMock := &mockApplicationServicerGRPC{
		UnbanUserFunc: func(ctx context.Context, userID, applicationID uuid.UUID) error {
			return nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.UnbanUser(context.Background(), &pb.UnbanUserRequest{
		UserId:        uuid.New().String(),
		ApplicationId: uuid.New().String(),
	})

	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "User unbanned", resp.Message)
}

func TestUnbanUser_ShouldReturnError_WhenInvalidUserID(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.UnbanUser(context.Background(), &pb.UnbanUserRequest{
		UserId:        "bad",
		ApplicationId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user_id format")
}

// ===================== IntrospectToken Tests =====================

func TestIntrospectToken_ShouldReturnActive_WhenValidJWT(t *testing.T) {
	jwtSvc := newTestJWTService()
	userID := uuid.New()
	user := newTestUser(userID, "admin")

	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)

	redisMock := &mockRedisServicerGRPC{
		IsBlacklistedFunc: func(ctx context.Context, tokenHash string) (bool, error) {
			return false, nil
		},
	}

	h := newTestAuthHandlerV2(jwtSvc, func(h *AuthHandlerV2) {
		h.redis = redisMock
	})

	resp, err := h.IntrospectToken(context.Background(), &pb.IntrospectTokenRequest{
		AccessToken: token,
	})

	require.NoError(t, err)
	assert.True(t, resp.Active)
	assert.Equal(t, userID.String(), resp.UserId)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, "admin", resp.Role)
	assert.False(t, resp.Blacklisted)
	assert.True(t, resp.ExpiresAt > 0)
	assert.True(t, resp.IssuedAt > 0)
}

func TestIntrospectToken_ShouldReturnError_WhenEmpty(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.IntrospectToken(context.Background(), &pb.IntrospectTokenRequest{
		AccessToken: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "access_token is required")
}

func TestIntrospectToken_ShouldReturnInactive_WhenBlacklisted(t *testing.T) {
	jwtSvc := newTestJWTService()
	user := newTestUser(uuid.New(), "user")

	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)

	redisMock := &mockRedisServicerGRPC{
		IsBlacklistedFunc: func(ctx context.Context, tokenHash string) (bool, error) {
			return true, nil
		},
	}

	h := newTestAuthHandlerV2(jwtSvc, func(h *AuthHandlerV2) {
		h.redis = redisMock
	})

	resp, err := h.IntrospectToken(context.Background(), &pb.IntrospectTokenRequest{
		AccessToken: token,
	})

	require.NoError(t, err)
	assert.False(t, resp.Active)
	assert.True(t, resp.Blacklisted)
}

// ===================== ListApplicationUsers Tests =====================

func TestListApplicationUsers_ShouldReturnProfiles_WhenValid(t *testing.T) {
	appID := uuid.New()
	userID := uuid.New()
	displayName := "John"

	appMock := &mockApplicationServicerGRPC{
		ListApplicationUsersFunc: func(ctx context.Context, applicationID uuid.UUID, page, perPage int) (*models.UserAppProfileListResponse, error) {
			return &models.UserAppProfileListResponse{
				Profiles: []models.UserApplicationProfile{
					{
						UserID:        userID,
						ApplicationID: appID,
						DisplayName:   &displayName,
						IsActive:      true,
						AppRoles:      []string{"member"},
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					},
				},
				Total:      1,
				Page:       1,
				PageSize:   20,
				TotalPages: 1,
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.ListApplicationUsers(context.Background(), &pb.ListApplicationUsersRequest{
		ApplicationId: appID.String(),
		Page:          1,
		PageSize:      20,
	})

	require.NoError(t, err)
	assert.Len(t, resp.Profiles, 1)
	assert.Equal(t, int32(1), resp.Total)
	assert.Equal(t, userID.String(), resp.Profiles[0].UserId)
	assert.Equal(t, "John", resp.Profiles[0].DisplayName)
}

func TestListApplicationUsers_ShouldReturnError_WhenInvalidAppID(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.ListApplicationUsers(context.Background(), &pb.ListApplicationUsersRequest{
		ApplicationId: "not-a-uuid",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid application_id format")
}

// ===================== SendEmail Tests =====================

func TestSendEmail_ShouldReturnSuccess_WhenEmailSent(t *testing.T) {
	emailMock := &mockEmailProfileServicerGRPC{
		SendEmailFunc: func(ctx context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, templateType string, variables map[string]interface{}) error {
			return nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.emailProfileService = emailMock
	})

	resp, err := h.SendEmail(context.Background(), &pb.SendEmailRequest{
		ToEmail:      "user@example.com",
		TemplateType: "welcome",
		Variables:    map[string]string{"name": "John"},
	})

	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Email sent successfully", resp.Message)
}

func TestSendEmail_ShouldReturnError_WhenToEmailEmptyWithMocks(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.SendEmail(context.Background(), &pb.SendEmailRequest{
		ToEmail:      "",
		TemplateType: "welcome",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "to_email is required")
}

func TestSendEmail_ShouldReturnFailed_WhenServiceError(t *testing.T) {
	emailMock := &mockEmailProfileServicerGRPC{
		SendEmailFunc: func(ctx context.Context, profileID *uuid.UUID, applicationID *uuid.UUID, toEmail string, templateType string, variables map[string]interface{}) error {
			return errors.New("SMTP connection failed")
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.emailProfileService = emailMock
	})

	resp, err := h.SendEmail(context.Background(), &pb.SendEmailRequest{
		ToEmail:      "user@example.com",
		TemplateType: "welcome",
	})

	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "SMTP connection failed", resp.ErrorMessage)
}

// ===================== IntrospectOAuthToken Tests =====================

func TestIntrospectOAuthToken_ShouldReturnActive_WhenTokenValid(t *testing.T) {
	oauthMock := &mockOAuthProviderServicerGRPC{
		IntrospectTokenFunc: func(ctx context.Context, token, tokenTypeHint string, clientID *string) (*models.IntrospectionResponse, error) {
			return &models.IntrospectionResponse{
				Active:    true,
				Scope:     "openid profile",
				ClientID:  "client-123",
				Username:  "johndoe",
				TokenType: "Bearer",
				ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.oauthProviderService = oauthMock
	})

	resp, err := h.IntrospectOAuthToken(context.Background(), &pb.IntrospectOAuthTokenRequest{
		Token: "oauth-token-abc",
	})

	require.NoError(t, err)
	assert.True(t, resp.Active)
	assert.Equal(t, "openid profile", resp.Scope)
	assert.Equal(t, "client-123", resp.ClientId)
	assert.Equal(t, "johndoe", resp.Username)
}

func TestIntrospectOAuthToken_ShouldReturnInactive_WhenEmpty(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	resp, err := h.IntrospectOAuthToken(context.Background(), &pb.IntrospectOAuthTokenRequest{
		Token: "",
	})

	require.NoError(t, err)
	assert.False(t, resp.Active)
	assert.Equal(t, "token is required", resp.ErrorMessage)
}

// ===================== ValidateOAuthClient Tests =====================

func TestValidateOAuthClient_ShouldReturnValid_WhenCredentialsCorrect(t *testing.T) {
	oauthMock := &mockOAuthProviderServicerGRPC{
		ValidateClientCredentialsFunc: func(ctx context.Context, clientID, clientSecret string) (*models.OAuthClient, error) {
			return &models.OAuthClient{
				ClientID:          "my-client",
				Name:              "My Client App",
				ClientType:        "confidential",
				AllowedScopes:     []string{"openid", "profile"},
				RedirectURIs:      []string{"https://app.example.com/callback"},
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.oauthProviderService = oauthMock
	})

	resp, err := h.ValidateOAuthClient(context.Background(), &pb.ValidateOAuthClientRequest{
		ClientId:     "my-client",
		ClientSecret: "secret123",
	})

	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, "my-client", resp.ClientId)
	assert.Equal(t, "My Client App", resp.ClientName)
	assert.Contains(t, resp.Scopes, "openid")
}

func TestValidateOAuthClient_ShouldReturnInvalid_WhenCredentialsWrong(t *testing.T) {
	oauthMock := &mockOAuthProviderServicerGRPC{
		ValidateClientCredentialsFunc: func(ctx context.Context, clientID, clientSecret string) (*models.OAuthClient, error) {
			return nil, errors.New("invalid client credentials")
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.oauthProviderService = oauthMock
	})

	resp, err := h.ValidateOAuthClient(context.Background(), &pb.ValidateOAuthClientRequest{
		ClientId:     "my-client",
		ClientSecret: "wrong-secret",
	})

	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.Equal(t, "invalid client credentials", resp.ErrorMessage)
}

// ===================== GetOAuthClient Tests =====================

func TestGetOAuthClient_ShouldReturnClient_WhenExists(t *testing.T) {
	clientID := uuid.New()

	oauthMock := &mockOAuthProviderServicerGRPC{
		GetClientByClientIDFunc: func(ctx context.Context, cid string) (*models.OAuthClient, error) {
			return &models.OAuthClient{
				ID:                clientID,
				ClientID:          "my-client",
				Name:              "My Client",
				ClientType:        "public",
				RedirectURIs:      []string{"https://app.com/callback"},
				AllowedScopes:     []string{"openid"},
				AllowedGrantTypes: []string{"authorization_code"},
				IsActive:          true,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.oauthProviderService = oauthMock
	})

	resp, err := h.GetOAuthClient(context.Background(), &pb.GetOAuthClientRequest{
		ClientId: "my-client",
	})

	require.NoError(t, err)
	require.NotNil(t, resp.Client)
	assert.Equal(t, "my-client", resp.Client.ClientId)
	assert.Equal(t, "My Client", resp.Client.ClientName)
	assert.True(t, resp.Client.IsActive)
}

func TestGetOAuthClient_ShouldReturnError_WhenNotFound(t *testing.T) {
	oauthMock := &mockOAuthProviderServicerGRPC{
		GetClientByClientIDFunc: func(ctx context.Context, cid string) (*models.OAuthClient, error) {
			return nil, errors.New("not found")
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.oauthProviderService = oauthMock
	})

	_, err := h.GetOAuthClient(context.Background(), &pb.GetOAuthClientRequest{
		ClientId: "unknown-client",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth client not found")
}

// ===================== GetApplicationAuthConfig Tests =====================

func TestGetApplicationAuthConfig_ShouldReturnConfig_WhenValid(t *testing.T) {
	appID := uuid.New()

	appMock := &mockApplicationServicerGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return &models.Application{ID: appID, Name: "my-app", DisplayName: "My App"}, nil
		},
		GetAuthConfigFunc: func(ctx context.Context, app *models.Application) (*models.AuthConfigResponse, error) {
			return &models.AuthConfigResponse{
				ApplicationID:      appID,
				Name:               "my-app",
				DisplayName:        "My App",
				AllowedAuthMethods: []string{"password", "oauth_google"},
				OAuthProviders:     []string{"google"},
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.GetApplicationAuthConfig(context.Background(), &pb.GetApplicationAuthConfigRequest{
		ApplicationId: appID.String(),
	})

	require.NoError(t, err)
	assert.Empty(t, resp.ErrorMessage)
	assert.Equal(t, appID.String(), resp.ApplicationId)
	assert.Equal(t, "my-app", resp.Name)
	assert.Equal(t, "My App", resp.DisplayName)
	assert.Contains(t, resp.AllowedAuthMethods, "password")
	assert.Contains(t, resp.OauthProviders, "google")
}

func TestGetApplicationAuthConfig_ShouldReturnError_WhenAppNotFound(t *testing.T) {
	appMock := &mockApplicationServicerGRPC{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Application, error) {
			return nil, errors.New("not found")
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.GetApplicationAuthConfig(context.Background(), &pb.GetApplicationAuthConfigRequest{
		ApplicationId: uuid.New().String(),
	})

	require.NoError(t, err)
	assert.Equal(t, "application not found", resp.ErrorMessage)
}

// ===================== DeleteUserProfile Tests =====================

func TestDeleteUserProfile_ShouldReturnSuccess_WhenValid(t *testing.T) {
	appMock := &mockApplicationServicerGRPC{
		DeleteUserProfileFunc: func(ctx context.Context, userID, applicationID uuid.UUID) error {
			return nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.DeleteUserProfile(context.Background(), &pb.DeleteUserProfileRequest{
		UserId:        uuid.New().String(),
		ApplicationId: uuid.New().String(),
	})

	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "User profile deleted", resp.Message)
}

func TestDeleteUserProfile_ShouldReturnError_WhenMissingIDsWithMocks(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.DeleteUserProfile(context.Background(), &pb.DeleteUserProfileRequest{
		UserId:        "",
		ApplicationId: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id and application_id are required")
}

// ===================== UpdateUserProfile Tests =====================

func TestUpdateUserProfile_ShouldReturnProfile_WhenValid(t *testing.T) {
	userID := uuid.New()
	appID := uuid.New()
	displayName := "Updated Name"

	appMock := &mockApplicationServicerGRPC{
		UpdateUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID, req *models.UpdateUserAppProfileRequest) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				UserID:        uid,
				ApplicationID: aid,
				DisplayName:   &displayName,
				IsActive:      true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.UpdateUserProfile(context.Background(), &pb.UpdateUserProfileRequest{
		UserId:        userID.String(),
		ApplicationId: appID.String(),
		DisplayName:   "Updated Name",
	})

	require.NoError(t, err)
	assert.Equal(t, userID.String(), resp.UserId)
	assert.Equal(t, appID.String(), resp.ApplicationId)
	assert.Equal(t, "Updated Name", resp.DisplayName)
}

// ===================== CreateUserProfile Tests =====================

func TestCreateUserProfile_ShouldReturnProfile_WhenValid(t *testing.T) {
	userID := uuid.New()
	appID := uuid.New()

	appMock := &mockApplicationServicerGRPC{
		GetOrCreateUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				UserID:        uid,
				ApplicationID: aid,
				IsActive:      true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.CreateUserProfile(context.Background(), &pb.CreateUserProfileRequest{
		UserId:        userID.String(),
		ApplicationId: appID.String(),
	})

	require.NoError(t, err)
	assert.Equal(t, userID.String(), resp.UserId)
	assert.Equal(t, appID.String(), resp.ApplicationId)
	assert.True(t, resp.IsActive)
}

// ===================== GetUserApplicationProfile Tests =====================

func TestGetUserApplicationProfile_ShouldReturnProfile_WhenValid(t *testing.T) {
	userID := uuid.New()
	appID := uuid.New()
	displayName := "Profile Name"

	appMock := &mockApplicationServicerGRPC{
		GetOrCreateUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return &models.UserApplicationProfile{
				UserID:        uid,
				ApplicationID: aid,
				DisplayName:   &displayName,
				IsActive:      true,
				AppRoles:      []string{"viewer"},
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.GetUserApplicationProfile(context.Background(), &pb.GetUserAppProfileRequest{
		UserId:        userID.String(),
		ApplicationId: appID.String(),
	})

	require.NoError(t, err)
	assert.Equal(t, userID.String(), resp.UserId)
	assert.Equal(t, "Profile Name", resp.DisplayName)
	assert.Contains(t, resp.AppRoles, "viewer")
}

func TestGetUserApplicationProfile_ShouldReturnError_WhenServiceFails(t *testing.T) {
	appMock := &mockApplicationServicerGRPC{
		GetOrCreateUserProfileFunc: func(ctx context.Context, uid, aid uuid.UUID) (*models.UserApplicationProfile, error) {
			return nil, errors.New("database error")
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	_, err := h.GetUserApplicationProfile(context.Background(), &pb.GetUserAppProfileRequest{
		UserId:        uuid.New().String(),
		ApplicationId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user profile")
}

// ===================== LoginWithOTP Tests =====================

func TestLoginWithOTP_ShouldReturnSuccess_WhenUserExistsAndOTPSent(t *testing.T) {
	userID := uuid.New()
	user := newTestUser(userID, "user")

	userRepoMock := &mockUserStoreGRPC{
		GetByEmailFunc: func(ctx context.Context, email string, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			return user, nil
		},
	}
	otpMock := &mockOTPServicerGRPC{
		SendOTPFunc: func(ctx context.Context, req *models.SendOTPRequest) error {
			return nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.userRepo = userRepoMock
		h.otpService = otpMock
	})

	resp, err := h.LoginWithOTP(context.Background(), &pb.LoginWithOTPRequest{
		Email: "user@example.com",
	})

	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "OTP sent to your email", resp.Message)
}

func TestLoginWithOTP_ShouldReturnFailed_WhenUserNotFound(t *testing.T) {
	userRepoMock := &mockUserStoreGRPC{
		GetByEmailFunc: func(ctx context.Context, email string, isActive *bool, opts ...service.UserGetOption) (*models.User, error) {
			return nil, models.ErrUserNotFound
		},
	}
	otpMock := &mockOTPServicerGRPC{}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.userRepo = userRepoMock
		h.otpService = otpMock
	})

	resp, err := h.LoginWithOTP(context.Background(), &pb.LoginWithOTPRequest{
		Email: "nonexistent@example.com",
	})

	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "user not found", resp.ErrorMessage)
}

// ===================== InitPasswordlessRegistration Tests =====================

func TestInitPasswordlessRegistration_ShouldReturnSuccess_WhenValid(t *testing.T) {
	authMock := &mockAuthServicerGRPC{
		InitPasswordlessRegistrationFunc: func(ctx context.Context, req *models.InitPasswordlessRegistrationRequest, ip, userAgent string) error {
			return nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.authService = authMock
	})

	resp, err := h.InitPasswordlessRegistration(context.Background(), &pb.InitPasswordlessRegistrationRequest{
		Email:    "new@example.com",
		Username: "newuser",
	})

	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Contains(t, resp.Message, "OTP sent successfully")
}

func TestInitPasswordlessRegistration_ShouldReturnError_WhenNoEmailOrPhoneWithMocks(t *testing.T) {
	h := newTestAuthHandlerV2(newTestJWTService())

	_, err := h.InitPasswordlessRegistration(context.Background(), &pb.InitPasswordlessRegistrationRequest{
		Email: "",
		Phone: "",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "either email or phone is required")
}

// ===================== ValidateToken with Redis fallback Tests =====================

func TestValidateToken_ShouldFallbackToTokenRepo_WhenRedisFails(t *testing.T) {
	jwtSvc := newTestJWTService()
	user := newTestUser(uuid.New(), "user")

	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)

	redisMock := &mockRedisServicerGRPC{
		IsBlacklistedFunc: func(ctx context.Context, tokenHash string) (bool, error) {
			return false, errors.New("redis unavailable")
		},
	}
	tokenRepoMock := &mockTokenStoreGRPC{
		IsBlacklistedFunc: func(ctx context.Context, tokenHash string) (bool, error) {
			return false, nil
		},
	}

	h := newTestAuthHandlerV2(jwtSvc, func(h *AuthHandlerV2) {
		h.redis = redisMock
		h.tokenRepo = tokenRepoMock
	})

	resp, err := h.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
		AccessToken: token,
	})

	require.NoError(t, err)
	assert.True(t, resp.Valid)
}

// ===================== Login with Phone Tests =====================

func TestLogin_ShouldAcceptPhone_WhenPhoneProvided(t *testing.T) {
	userID := uuid.New()
	user := newTestUser(userID, "user")

	authMock := &mockAuthServicerGRPC{
		SignInFunc: func(ctx context.Context, req *models.SignInRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
			return &models.AuthResponse{
				AccessToken:  "phone-access-token",
				RefreshToken: "phone-refresh-token",
				User:         user,
				ExpiresIn:    900,
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.authService = authMock
	})

	resp, err := h.Login(context.Background(), &pb.LoginRequest{
		Phone:    "+1234567890",
		Password: "SecurePass123",
	})

	require.NoError(t, err)
	assert.Equal(t, "phone-access-token", resp.AccessToken)
	require.NotNil(t, resp.User)
}

// ===================== BanUser with service error Tests =====================

func TestBanUser_ShouldReturnError_WhenServiceFails(t *testing.T) {
	appMock := &mockApplicationServicerGRPC{
		BanUserFunc: func(ctx context.Context, userID, applicationID, bannedBy uuid.UUID, reason string) error {
			return errors.New("user already banned")
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	_, err := h.BanUser(context.Background(), &pb.BanUserRequest{
		UserId:        uuid.New().String(),
		ApplicationId: uuid.New().String(),
		Reason:        "test",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ban user")
}

// ===================== UnbanUser with service error Tests =====================

func TestUnbanUser_ShouldReturnError_WhenServiceFails(t *testing.T) {
	appMock := &mockApplicationServicerGRPC{
		UnbanUserFunc: func(ctx context.Context, userID, applicationID uuid.UUID) error {
			return errors.New("user not banned")
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	_, err := h.UnbanUser(context.Background(), &pb.UnbanUserRequest{
		UserId:        uuid.New().String(),
		ApplicationId: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unban user")
}

// ===================== CreateUser with phone =====================

func TestCreateUser_ShouldAcceptPhone_WhenPhoneProvided(t *testing.T) {
	userID := uuid.New()
	user := newTestUser(userID, "user")

	authMock := &mockAuthServicerGRPC{
		SignUpFunc: func(ctx context.Context, req *models.CreateUserRequest, ip, userAgent string, deviceInfo models.DeviceInfo, appID *uuid.UUID) (*models.AuthResponse, error) {
			return &models.AuthResponse{
				AccessToken:  "phone-token",
				RefreshToken: "phone-refresh",
				User:         user,
				ExpiresIn:    900,
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.authService = authMock
	})

	resp, err := h.CreateUser(context.Background(), &pb.CreateUserRequest{
		Phone:    "+1234567890",
		Password: "SecurePass123",
	})

	require.NoError(t, err)
	assert.Equal(t, "phone-token", resp.AccessToken)
	require.NotNil(t, resp.User)
}

// ===================== RedeemTokenExchange with nil user =====================

func TestRedeemTokenExchange_ShouldHandleNilUser(t *testing.T) {
	exchangeMock := &mockTokenExchangeServicerGRPC{
		RedeemExchangeFunc: func(ctx context.Context, req *models.RedeemTokenExchangeRequest, redeemingAppID *uuid.UUID) (*models.RedeemTokenExchangeResponse, error) {
			return &models.RedeemTokenExchangeResponse{
				AccessToken:   "token-abc",
				RefreshToken:  "refresh-xyz",
				User:          nil,
				ApplicationID: "app-id",
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.tokenExchangeService = exchangeMock
	})

	resp, err := h.RedeemTokenExchange(context.Background(), &pb.RedeemTokenExchangeGrpcRequest{
		ExchangeCode: "valid-code",
	})

	require.NoError(t, err)
	assert.Equal(t, "token-abc", resp.AccessToken)
	assert.Empty(t, resp.UserId)
	assert.Empty(t, resp.Email)
}

// ===================== SyncUsers with app profile =====================

func TestSyncUsers_ShouldIncludeAppProfile_WhenPresent(t *testing.T) {
	now := time.Now()

	adminMock := &mockAdminServicerGRPC{
		SyncUsersFunc: func(ctx context.Context, updatedAfter time.Time, appID *uuid.UUID, limit, offset int) (*models.SyncUsersResponse, error) {
			return &models.SyncUsersResponse{
				Users: []models.SyncUserResponse{
					{
						ID:        uuid.New(),
						Email:     "user@example.com",
						Username:  "user1",
						IsActive:  true,
						UpdatedAt: now,
						AppProfile: &models.SyncUserAppProfile{
							DisplayName: "User Display",
							AppRoles:    []string{"editor"},
							IsActive:    true,
							IsBanned:    false,
						},
					},
				},
				Total:         1,
				HasMore:       false,
				SyncTimestamp: now.Format(time.RFC3339),
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.adminService = adminMock
	})

	resp, err := h.SyncUsers(context.Background(), &pb.SyncUsersRequest{
		UpdatedAfter: now.Add(-1 * time.Hour).Format(time.RFC3339),
	})

	require.NoError(t, err)
	require.Len(t, resp.Users, 1)
	require.NotNil(t, resp.Users[0].AppProfile)
	assert.Equal(t, "User Display", resp.Users[0].AppProfile.DisplayName)
	assert.Contains(t, resp.Users[0].AppProfile.AppRoles, "editor")
	assert.True(t, resp.Users[0].AppProfile.IsActive)
	assert.False(t, resp.Users[0].AppProfile.IsBanned)
}

// ===================== ListApplicationUsers pagination defaults =====================

func TestListApplicationUsers_ShouldUseDefaultPagination_WhenZeroValues(t *testing.T) {
	appID := uuid.New()
	var capturedPage, capturedPageSize int

	appMock := &mockApplicationServicerGRPC{
		ListApplicationUsersFunc: func(ctx context.Context, applicationID uuid.UUID, page, perPage int) (*models.UserAppProfileListResponse, error) {
			capturedPage = page
			capturedPageSize = perPage
			return &models.UserAppProfileListResponse{
				Profiles:   []models.UserApplicationProfile{},
				Total:      0,
				Page:       page,
				PageSize:   perPage,
				TotalPages: 0,
			}, nil
		},
	}

	h := newTestAuthHandlerV2(newTestJWTService(), func(h *AuthHandlerV2) {
		h.appService = appMock
	})

	resp, err := h.ListApplicationUsers(context.Background(), &pb.ListApplicationUsersRequest{
		ApplicationId: appID.String(),
		Page:          0,
		PageSize:      0,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, capturedPage)
	assert.Equal(t, 20, capturedPageSize)
	assert.Equal(t, int32(0), resp.Total)
}
