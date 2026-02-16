package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	jwtpkg "github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- Mock CacheService for BlacklistService ---

type mockCacheService struct {
	isBlacklistedFn func(ctx context.Context, tokenHash string) (bool, error)
}

func (m *mockCacheService) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	if m.isBlacklistedFn != nil {
		return m.isBlacklistedFn(ctx, tokenHash)
	}
	return false, nil
}

func (m *mockCacheService) AddToBlacklist(ctx context.Context, tokenHash string, expiration time.Duration) error {
	return nil
}

func (m *mockCacheService) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	return 0, nil
}

func (m *mockCacheService) StorePendingRegistration(ctx context.Context, identifier string, data *models.PendingRegistration, expiration time.Duration) error {
	return nil
}

func (m *mockCacheService) GetPendingRegistration(ctx context.Context, identifier string) (*models.PendingRegistration, error) {
	return nil, nil
}

func (m *mockCacheService) DeletePendingRegistration(ctx context.Context, identifier string) error {
	return nil
}

// --- Helper functions ---

func newTestJWTService() *jwtpkg.Service {
	return jwtpkg.NewService("test-access-secret", "test-refresh-secret", 15*time.Minute, 7*24*time.Hour)
}

func newTestUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		IsActive: true,
		Roles: []models.Role{
			{Name: "user"},
		},
	}
}

func generateValidAccessToken(t *testing.T, jwtSvc *jwtpkg.Service, user *models.User) string {
	t.Helper()
	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)
	return token
}

// newTestBlacklistService creates a BlacklistService with a mock cache that
// returns "not blacklisted" for all tokens (default happy path).
func newTestBlacklistService(jwtSvc *jwtpkg.Service) *service.BlacklistService {
	cache := &mockCacheService{
		isBlacklistedFn: func(ctx context.Context, tokenHash string) (bool, error) {
			return false, nil
		},
	}
	log := logger.New("test", logger.DebugLevel, false)
	return service.NewBlacklistService(cache, nil, nil, jwtSvc, log, &mockAuditLoggerAuth{})
}

// newBlacklistedBlacklistService creates a BlacklistService where all tokens are blacklisted.
func newBlacklistedBlacklistService(jwtSvc *jwtpkg.Service) *service.BlacklistService {
	cache := &mockCacheService{
		isBlacklistedFn: func(ctx context.Context, tokenHash string) (bool, error) {
			return true, nil
		},
	}
	log := logger.New("test", logger.DebugLevel, false)
	return service.NewBlacklistService(cache, nil, nil, jwtSvc, log, &mockAuditLoggerAuth{})
}

type mockAuditLoggerAuth struct{}

func (m *mockAuditLoggerAuth) LogWithAction(userID *uuid.UUID, action, status, ip, userAgent string, details map[string]interface{}) {
}
func (m *mockAuditLoggerAuth) Log(params service.AuditLogParams) {}

// newTestAuthMiddleware creates an AuthMiddleware with non-nil blacklist service.
func newTestAuthMiddleware(jwtSvc *jwtpkg.Service) *AuthMiddleware {
	return NewAuthMiddleware(jwtSvc, newTestBlacklistService(jwtSvc))
}

// --- Authenticate middleware tests ---

func TestAuthenticate_ShouldSetUserContext_WhenValidBearerToken(t *testing.T) {
	// Arrange
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)
	user := newTestUser()
	token := generateValidAccessToken(t, jwtSvc, user)

	r := gin.New()
	r.Use(authMw.Authenticate())
	var capturedUserID uuid.UUID
	var capturedEmail string
	r.GET("/test", func(c *gin.Context) {
		uid, _ := utils.GetUserIDFromContext(c)
		capturedUserID = *uid
		capturedEmail, _ = utils.GetUserEmailFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, user.ID, capturedUserID)
	assert.Equal(t, user.Email, capturedEmail)
}

func TestAuthenticate_ShouldReturn401_WhenNoAuthorizationHeader(t *testing.T) {
	// Arrange
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(authMw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Unauthorized", body.Message)
}

func TestAuthenticate_ShouldReturn401_WhenBadAuthorizationFormat(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{name: "missing Bearer prefix", header: "Token abc123"},
		{name: "only Bearer keyword", header: "Bearer"},
		{name: "Basic auth", header: "Basic abc123"},
		{name: "three parts", header: "Bearer abc 123"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jwtSvc := newTestJWTService()
			authMw := newTestAuthMiddleware(jwtSvc)

			r := gin.New()
			r.Use(authMw.Authenticate())
			r.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)

			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthenticate_ShouldReturn401_WhenTokenIsInvalid(t *testing.T) {
	// Arrange
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(authMw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Invalid token", body.Message)
}

func TestAuthenticate_ShouldReturn401_WhenTokenIsExpired(t *testing.T) {
	// Arrange: create a JWT service with extremely short-lived access tokens
	shortJWTSvc := jwtpkg.NewService("test-secret", "test-refresh-secret", 1*time.Millisecond, 7*24*time.Hour)
	user := newTestUser()
	token := generateValidAccessToken(t, shortJWTSvc, user)

	// Wait for token to expire
	time.Sleep(5 * time.Millisecond)

	authMw := NewAuthMiddleware(shortJWTSvc, newTestBlacklistService(shortJWTSvc))

	r := gin.New()
	r.Use(authMw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Token expired", body.Message)
}

func TestAuthenticate_ShouldReturn401_WhenTokenSignedWithDifferentSecret(t *testing.T) {
	// Arrange
	otherJWTSvc := jwtpkg.NewService("other-secret", "other-refresh", 15*time.Minute, 7*24*time.Hour)
	user := newTestUser()
	token := generateValidAccessToken(t, otherJWTSvc, user)

	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(authMw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthenticate_ShouldReturn401_WhenTokenIsBlacklisted(t *testing.T) {
	// Arrange
	jwtSvc := newTestJWTService()
	blacklistSvc := newBlacklistedBlacklistService(jwtSvc)
	authMw := NewAuthMiddleware(jwtSvc, blacklistSvc)
	user := newTestUser()
	token := generateValidAccessToken(t, jwtSvc, user)

	r := gin.New()
	r.Use(authMw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Token revoked", body.Message)
}

func TestAuthenticate_ShouldSetApplicationIDFromClaims_WhenPresent(t *testing.T) {
	// Arrange
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)
	user := newTestUser()
	appID := uuid.New()
	token, err := jwtSvc.GenerateAccessTokenWithApp(user, appID)
	require.NoError(t, err)

	r := gin.New()
	r.Use(authMw.Authenticate())
	var capturedAppID *uuid.UUID
	r.GET("/test", func(c *gin.Context) {
		capturedAppID, _ = utils.GetApplicationIDFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedAppID)
	assert.Equal(t, appID, *capturedAppID)
}

func TestAuthenticate_ShouldSetRolesInContext_WhenValidToken(t *testing.T) {
	// Arrange
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)
	user := &models.User{
		ID:       uuid.New(),
		Email:    "admin@example.com",
		Username: "admin",
		IsActive: true,
		Roles: []models.Role{
			{Name: "admin"},
			{Name: "user"},
		},
	}
	token := generateValidAccessToken(t, jwtSvc, user)

	r := gin.New()
	r.Use(authMw.Authenticate())
	var capturedRoles []string
	r.GET("/test", func(c *gin.Context) {
		capturedRoles, _ = utils.GetUserRolesFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, capturedRoles, "admin")
	assert.Contains(t, capturedRoles, "user")
}

func TestAuthenticate_ShouldStoreTokenInContext_WhenValid(t *testing.T) {
	// Arrange
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)
	user := newTestUser()
	token := generateValidAccessToken(t, jwtSvc, user)

	r := gin.New()
	r.Use(authMw.Authenticate())
	var capturedToken string
	r.GET("/test", func(c *gin.Context) {
		capturedToken, _ = utils.GetTokenFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, token, capturedToken)
}

func TestAuthenticate_ShouldNotOverrideExistingAppID_WhenAlreadySet(t *testing.T) {
	// Arrange: AppID already set in context (by ExtractApplicationID middleware)
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)
	user := newTestUser()
	tokenAppID := uuid.New()
	presetAppID := uuid.New()
	token, err := jwtSvc.GenerateAccessTokenWithApp(user, tokenAppID)
	require.NoError(t, err)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.ApplicationIDKey, presetAppID)
		c.Next()
	})
	r.Use(authMw.Authenticate())
	var capturedAppID *uuid.UUID
	r.GET("/test", func(c *gin.Context) {
		capturedAppID, _ = utils.GetApplicationIDFromContext(c)
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedAppID)
	assert.Equal(t, presetAppID, *capturedAppID, "should NOT override existing app ID")
}

func TestAuthenticate_ShouldDelegateToAPIKeyMiddleware_WhenXAPIKeyHeader(t *testing.T) {
	// Arrange: Without apiKeyMiddleware set, falls through to JWT which fails
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(authMw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "agw_test_key_12345")

	// Act: apiKeyMiddleware is nil, so isAPIKeyOrAppSecret check fails,
	// falls through to JWT path with the garbage header value
	r.ServeHTTP(w, req)

	// Assert: 401 because X-API-Key was set but no Authorization header
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- isAPIKeyOrAppSecret tests ---

func TestIsAPIKeyOrAppSecret_ShouldReturnTrue_WhenXAPIKeyHeaderSet(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-API-Key", "agw_some_key")

	assert.True(t, authMw.isAPIKeyOrAppSecret(c))
}

func TestIsAPIKeyOrAppSecret_ShouldReturnTrue_WhenXAppSecretHeaderSet(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("X-App-Secret", "app_some_secret")

	assert.True(t, authMw.isAPIKeyOrAppSecret(c))
}

func TestIsAPIKeyOrAppSecret_ShouldReturnTrue_WhenBearerAgwPrefix(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer agw_somekey123")

	assert.True(t, authMw.isAPIKeyOrAppSecret(c))
}

func TestIsAPIKeyOrAppSecret_ShouldReturnTrue_WhenBearerAppPrefix(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer app_somekey123")

	assert.True(t, authMw.isAPIKeyOrAppSecret(c))
}

func TestIsAPIKeyOrAppSecret_ShouldReturnFalse_WhenRegularBearerToken(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.payload.signature")

	assert.False(t, authMw.isAPIKeyOrAppSecret(c))
}

func TestIsAPIKeyOrAppSecret_ShouldReturnFalse_WhenNoHeaders(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	assert.False(t, authMw.isAPIKeyOrAppSecret(c))
}

// --- RequireRole tests ---

func TestRequireRole_ShouldAllow_WhenUserHasRequiredRole(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"user", "moderator"})
		c.Next()
	})
	r.Use(authMw.RequireRole("moderator"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_ShouldAllow_WhenUserIsAdmin(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"admin"})
		c.Next()
	})
	r.Use(authMw.RequireRole("superspecial"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_ShouldReturn403_WhenUserLacksRole(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"user"})
		c.Next()
	})
	r.Use(authMw.RequireRole("admin"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_ShouldReturn401_WhenNoRolesInContext(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(authMw.RequireRole("admin"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- RequireAnyRole tests ---

func TestRequireAnyRole_ShouldAllow_WhenUserHasOneOfRequiredRoles(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"moderator"})
		c.Next()
	})
	r.Use(authMw.RequireAnyRole("admin", "moderator", "editor"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAnyRole_ShouldAllow_WhenUserIsAdmin(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"admin"})
		c.Next()
	})
	r.Use(authMw.RequireAnyRole("editor", "reviewer"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAnyRole_ShouldReturn403_WhenUserHasNoneOfRequiredRoles(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"user"})
		c.Next()
	})
	r.Use(authMw.RequireAnyRole("admin", "moderator", "editor"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAnyRole_ShouldReturn401_WhenNoRolesInContext(t *testing.T) {
	jwtSvc := newTestJWTService()
	authMw := newTestAuthMiddleware(jwtSvc)

	r := gin.New()
	r.Use(authMw.RequireAnyRole("admin", "moderator"))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- contains helper tests ---

func TestContains_ShouldReturnTrue_WhenItemExists(t *testing.T) {
	assert.True(t, contains([]string{"a", "b", "c"}, "b"))
}

func TestContains_ShouldReturnFalse_WhenItemMissing(t *testing.T) {
	assert.False(t, contains([]string{"a", "b", "c"}, "d"))
}

func TestContains_ShouldReturnFalse_WhenSliceIsEmpty(t *testing.T) {
	assert.False(t, contains([]string{}, "a"))
}

func TestContains_ShouldReturnFalse_WhenSliceIsNil(t *testing.T) {
	assert.False(t, contains(nil, "a"))
}
