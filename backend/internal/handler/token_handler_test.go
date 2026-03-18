package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TokenHandler test helpers
// ---------------------------------------------------------------------------

const (
	testAccessSecret  = "test-access-secret-key-min-32-chars!!"
	testRefreshSecret = "test-refresh-secret-key-min-32-chars!!"
)

func setupTokenHandler() (*TokenHandler, *jwt.Service, *mockAPIKeyServicer, *mockRedisServicer) {
	jwtSvc := jwt.NewService(testAccessSecret, testRefreshSecret, 15*time.Minute, 7*24*time.Hour)
	apiKeySvc := &mockAPIKeyServicer{}
	redisSvc := &mockRedisServicer{}
	h := NewTokenHandler(jwtSvc, apiKeySvc, redisSvc, testLogger())
	return h, jwtSvc, apiKeySvc, redisSvc
}

func generateTestUser() *models.User {
	return &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		IsActive: true,
		Roles:    []models.Role{{Name: "user"}},
	}
}

// ---------------------------------------------------------------------------
// ValidateToken Tests
// ---------------------------------------------------------------------------

func TestTokenHandler_ValidateToken_ShouldReturn200_WhenValidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, jwtSvc, _, _ := setupTokenHandler()

	user := generateTestUser()
	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	body := `{"access_token":"` + token + `"}`
	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ValidateTokenResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, user.ID.String(), resp.UserID)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "testuser", resp.Username)
	assert.Contains(t, resp.Roles, "user")
	assert.True(t, resp.IsActive)
}

func TestTokenHandler_ValidateToken_ShouldReturn400_WhenInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _ := setupTokenHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTokenHandler_ValidateToken_ShouldReturn400_WhenMissingAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _ := setupTokenHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTokenHandler_ValidateToken_ShouldReturn401_WhenInvalidJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _ := setupTokenHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	body := `{"access_token":"totally-not-a-valid-jwt-token"}`
	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp ValidateTokenErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestTokenHandler_ValidateToken_ShouldReturn401_WhenExpiredJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a JWT service with 0 expiration to produce immediately expired tokens
	jwtSvc := jwt.NewService(testAccessSecret, testRefreshSecret, -1*time.Second, 7*24*time.Hour)
	apiKeySvc := &mockAPIKeyServicer{}
	redisSvc := &mockRedisServicer{}
	h := NewTokenHandler(jwtSvc, apiKeySvc, redisSvc, testLogger())

	user := generateTestUser()
	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	body := `{"access_token":"` + token + `"}`
	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp ValidateTokenErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestTokenHandler_ValidateToken_ShouldReturn401_WhenBlacklisted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, jwtSvc, _, redisSvc := setupTokenHandler()

	user := generateTestUser()
	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)

	redisSvc.IsBlacklistedFunc = func(tokenHash string) (bool, error) {
		return true, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	body := `{"access_token":"` + token + `"}`
	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp ValidateTokenErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Valid)
	assert.Equal(t, "token is blacklisted", resp.ErrorMessage)
}

func TestTokenHandler_ValidateToken_ShouldReturn200_WhenAPIKeyValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, apiKeySvc, _ := setupTokenHandler()

	userID := uuid.New()
	apiKeySvc.ValidateAPIKeyFunc = func(plainKey string) (*models.APIKey, *models.User, error) {
		return &models.APIKey{ID: uuid.New()}, &models.User{
			ID:       userID,
			Email:    "apikey@example.com",
			Username: "apikeyuser",
			IsActive: true,
			Roles:    []models.Role{{Name: "admin"}},
		}, nil
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	body := `{"access_token":"agw_test_key_12345678901234567890"}`
	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ValidateTokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, userID.String(), resp.UserID)
	assert.Equal(t, "apikey@example.com", resp.Email)
	assert.Contains(t, resp.Roles, "admin")
}

func TestTokenHandler_ValidateToken_ShouldReturn401_WhenAPIKeyInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, apiKeySvc, _ := setupTokenHandler()

	apiKeySvc.ValidateAPIKeyFunc = func(plainKey string) (*models.APIKey, *models.User, error) {
		return nil, nil, errors.New("invalid API key")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	body := `{"access_token":"agw_invalid_key_12345678901234567890"}`
	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp ValidateTokenErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestTokenHandler_ValidateToken_ShouldReturn200_WhenRedisErrorButNotBlacklisted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, jwtSvc, _, redisSvc := setupTokenHandler()

	user := generateTestUser()
	token, err := jwtSvc.GenerateAccessToken(user)
	require.NoError(t, err)

	// Redis returns an error but blacklisted=false - handler should still succeed
	redisSvc.IsBlacklistedFunc = func(tokenHash string) (bool, error) {
		return false, errors.New("redis connection error")
	}

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	body := `{"access_token":"` + token + `"}`
	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ValidateTokenResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, user.ID.String(), resp.UserID)
}

func TestTokenHandler_ValidateToken_ShouldReturn400_WhenEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _ := setupTokenHandler()

	w := httptest.NewRecorder()
	r := gin.New()
	r.POST("/token/validate", h.ValidateToken)

	req := httptest.NewRequest(http.MethodPost, "/token/validate", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
