package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers – construct real services wired with mock repositories
// ---------------------------------------------------------------------------

func testLogger() *logger.Logger {
	return logger.New("test", logger.ErrorLevel, false)
}

func newTestAuthService(
	userRepo service.UserStore,
	tokenRepo service.TransactionalTokenStore,
	rbacRepo service.RBACStore,
	auditLogger service.AuditLogger,
	jwtSvc service.TokenService,
	blacklist service.BlacklistChecker,
	cache service.CacheService,
	sessionMgr service.SessionManager,
	db service.TransactionDB,
) *service.AuthService {
	return service.NewAuthService(
		userRepo,
		tokenRepo,
		rbacRepo,
		auditLogger,
		jwtSvc,
		blacklist,
		cache,
		sessionMgr,
		nil, // twoFAService
		10,  // bcryptCost
		utils.DefaultPasswordPolicy(),
		db,
		nil, // appRepo
		nil, // loginAlertService
		nil, // webhookService
		false,
		nil, // passwordChecker
	)
}

func newTestUserService(userRepo service.UserStore, auditLogger service.AuditLogger) *service.UserService {
	return service.NewUserService(userRepo, auditLogger)
}

func newTestOTPService(otpRepo service.OTPStore, userRepo service.UserStore, auditLogger service.AuditLogger) *service.OTPService {
	return service.NewOTPService(otpRepo, userRepo, auditLogger, service.OTPServiceOptions{})
}

// authTestFixture wraps handler and mock dependencies for auth handler tests.
type authTestFixture struct {
	handler   *AuthHandler
	userStore *mockUserStoreHandler
	tokenSvc  *mockTokenServiceHandler
	auditLog  *mockAuditLoggerHandler
	rbacStore *mockRBACStoreHandler
	otpStore  *mockOTPStoreHandler
}

func setupAuthTestFixture() *authTestFixture {
	userStore := &mockUserStoreHandler{}
	tokenSvc := &mockTokenServiceHandler{}
	auditLog := &mockAuditLoggerHandler{}
	tokenRepo := &mockTokenStoreHandler{}
	rbacStore := &mockRBACStoreHandler{}
	blacklist := &mockBlacklistCheckerHandler{}
	cache := &mockCacheServiceHandler{}
	sessionMgr := &mockSessionManagerHandler{}
	db := &mockTransactionDBHandler{}
	otpStore := &mockOTPStoreHandler{}

	authSvc := newTestAuthService(
		userStore, tokenRepo, rbacStore, auditLog, tokenSvc, blacklist, cache, sessionMgr, db,
	)
	userSvc := newTestUserService(userStore, auditLog)
	otpSvc := newTestOTPService(otpStore, userStore, auditLog)

	h := NewAuthHandler(authSvc, userSvc, otpSvc, nil, testLogger())
	return &authTestFixture{
		handler:   h,
		userStore: userStore,
		tokenSvc:  tokenSvc,
		auditLog:  auditLog,
		rbacStore: rbacStore,
		otpStore:  otpStore,
	}
}

// ---------------------------------------------------------------------------
// SignUp Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_SignUp_ShouldReturn400_WhenBodyIsInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing password", `{"email":"test@example.com","username":"testuser"}`},
		{"missing username", `{"email":"test@example.com","password":"SecurePass1!"}`},
		{"password too short", `{"email":"test@example.com","username":"testuser","password":"abc"}`},
		{"invalid JSON", `{invalid}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fix := setupAuthTestFixture()
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/auth/signup", fix.handler.SignUp)

			req := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAuthHandler_SignUp_ShouldReturnServiceError_WhenEmailMissing(t *testing.T) {
	// The service validates that at least email or phone is provided
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/signup", fix.handler.SignUp)

	// Valid JSON but no email and no phone -> service returns 400
	body := `{"username":"testuser","password":"SecurePass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_SignUp_ShouldReturn409_WhenEmailExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	// The service calls rbacRepo.GetRoleByName("user") before creating the user
	fix.rbacStore.GetRoleByNameFunc = func(name string) (*models.Role, error) {
		return &models.Role{ID: uuid.New(), Name: "user"}, nil
	}

	// Simulate unique constraint violation when creating user
	fix.userStore.CreateFunc = func(user *models.User) error {
		return models.ErrEmailAlreadyExists
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/signup", fix.handler.SignUp)

	body := `{"email":"existing@example.com","username":"newuser","password":"SecurePass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// ---------------------------------------------------------------------------
// SignIn Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_SignIn_ShouldReturn400_WhenBodyIsInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing password", `{"email":"test@example.com"}`},
		{"invalid JSON", `{bad json}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fix := setupAuthTestFixture()
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/auth/signin", fix.handler.SignIn)

			req := httptest.NewRequest(http.MethodPost, "/auth/signin", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAuthHandler_SignIn_ShouldReturn401_WhenCredentialsInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	// User not found -> invalid credentials
	fix.userStore.GetByEmailFunc = func(email string) (*models.User, error) {
		return nil, models.ErrInvalidCredentials
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/signin", fix.handler.SignIn)

	body := `{"email":"wrong@example.com","password":"WrongPass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/signin", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// RefreshToken Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_RefreshToken_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing refresh_token", `{}`},
		{"invalid JSON", `{invalid}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fix := setupAuthTestFixture()
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/auth/refresh", fix.handler.RefreshToken)

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestAuthHandler_RefreshToken_ShouldReturn401_WhenTokenInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	fix.tokenSvc.ValidateRefreshTokenFunc = func(token string) error {
		return models.ErrInvalidToken
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/refresh", fix.handler.RefreshToken)

	body := `{"refresh_token":"invalid-token"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// Logout Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_Logout_ShouldReturn401_WhenNoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/logout", fix.handler.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_ShouldReturn200_WhenTokenPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	userID := uuid.New()

	fix.tokenSvc.ExtractClaimsFunc = func(token string) (uuid.UUID, error) {
		return userID, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/logout", func(c *gin.Context) {
		c.Set(utils.TokenKey, "test-access-token")
		c.Set(utils.UserIDKey, userID)
		fix.handler.Logout(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.MessageResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Successfully logged out", resp.Message)
}

// ---------------------------------------------------------------------------
// GetProfile Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_GetProfile_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/auth/profile", fix.handler.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetProfile_ShouldReturn200_WhenAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	userID := uuid.New()
	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return &models.User{
			ID:       id,
			Email:    "profile@example.com",
			Username: "profileuser",
			IsActive: true,
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/auth/profile", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.GetProfile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user models.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	require.NoError(t, err)
	assert.Equal(t, "profile@example.com", user.Email)
	assert.Equal(t, "profileuser", user.Username)
}

func TestAuthHandler_GetProfile_ShouldReturn500_WhenServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	userID := uuid.New()
	fix.userStore.GetByIDFunc = func(id uuid.UUID) (*models.User, error) {
		return nil, fmt.Errorf("database error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.GET("/auth/profile", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.GetProfile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// UpdateProfile Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_UpdateProfile_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/auth/profile", fix.handler.UpdateProfile)

	body := `{"full_name":"John Doe"}`
	req := httptest.NewRequest(http.MethodPut, "/auth/profile", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_UpdateProfile_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()
	userID := uuid.New()

	w := httptest.NewRecorder()
	r := gin.New()
	r.PUT("/auth/profile", func(c *gin.Context) {
		c.Set(utils.UserIDKey, userID)
		fix.handler.UpdateProfile(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/auth/profile", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// ChangePassword Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_ChangePassword_ShouldReturn401_WhenNoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/change-password", fix.handler.ChangePassword)

	body := `{"old_password":"OldPass1!","new_password":"NewPass1!!"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/change-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_ChangePassword_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()
	userID := uuid.New()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing old_password", `{"new_password":"NewPass1!!"}`},
		{"missing new_password", `{"old_password":"OldPass1!"}`},
		{"new_password too short", `{"old_password":"OldPass1!","new_password":"abc"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/auth/change-password", func(c *gin.Context) {
				c.Set(utils.UserIDKey, userID)
				fix.handler.ChangePassword(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/auth/change-password", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// RequestPasswordReset Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_RequestPasswordReset_ShouldReturn400_WhenEmailMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/password/reset/request", fix.handler.RequestPasswordReset)

	req := httptest.NewRequest(http.MethodPost, "/auth/password/reset/request", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RequestPasswordReset_ShouldReturn200_WhenUserNotFound(t *testing.T) {
	// Should always return 200 to prevent user enumeration
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	fix.userStore.GetByEmailFunc = func(email string) (*models.User, error) {
		return nil, fmt.Errorf("not found")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/password/reset/request", fix.handler.RequestPasswordReset)

	body := `{"email":"nonexistent@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/password/reset/request", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// ResetPassword Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_ResetPassword_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing code", `{"email":"test@example.com","new_password":"NewPass1!!"}`},
		{"missing email", `{"code":"123456","new_password":"NewPass1!!"}`},
		{"missing new_password", `{"email":"test@example.com","code":"123456"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/auth/password/reset/complete", fix.handler.ResetPassword)

			req := httptest.NewRequest(http.MethodPost, "/auth/password/reset/complete", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// Verify2FA Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_Verify2FA_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing two_factor_token", `{"code":"123456"}`},
		{"missing code", `{"two_factor_token":"token-abc"}`},
		{"code wrong length", `{"two_factor_token":"token-abc","code":"12345"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/auth/2fa/login/verify", fix.handler.Verify2FA)

			req := httptest.NewRequest(http.MethodPost, "/auth/2fa/login/verify", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// InitPasswordlessRegistration Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_InitPasswordlessRegistration_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/auth/signup/phone", fix.handler.InitPasswordlessRegistration)

	req := httptest.NewRequest(http.MethodPost, "/auth/signup/phone", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// CompletePasswordlessRegistration Tests
// ---------------------------------------------------------------------------

func TestAuthHandler_CompletePasswordlessRegistration_ShouldReturn400_WhenBodyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fix := setupAuthTestFixture()

	tests := []struct {
		name string
		body string
	}{
		{"empty body", ""},
		{"missing code", `{"email":"test@example.com"}`},
		{"code wrong length", `{"email":"test@example.com","code":"12345"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := gin.New()
			r.POST("/auth/signup/phone/verify", fix.handler.CompletePasswordlessRegistration)

			req := httptest.NewRequest(http.MethodPost, "/auth/signup/phone/verify", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
