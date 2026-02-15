package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid UserID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		userID := uuid.New()
		c.Set(UserIDKey, userID)

		result, ok := GetUserIDFromContext(c)
		assert.True(t, ok)
		assert.NotNil(t, result)
		assert.Equal(t, userID, *result)
	})

	t.Run("No UserID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result, ok := GetUserIDFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Invalid UserID Type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set(UserIDKey, "not-a-uuid")

		result, ok := GetUserIDFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})
}

func TestGetUserEmailFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid Email", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		email := "test@example.com"
		c.Set(UserEmailKey, email)

		result, ok := GetUserEmailFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, email, result)
	})

	t.Run("No Email", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result, ok := GetUserEmailFromContext(c)
		assert.False(t, ok)
		assert.Equal(t, "", result)
	})
}

func TestGetUserRoleFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid Role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		role := "admin"
		c.Set(UserRoleKey, role)

		result, ok := GetUserRoleFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, role, result)
	})
}

func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("DelegatesToGinClientIP", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Request.RemoteAddr = "192.168.1.1:12345"

		ip := GetClientIP(c)
		assert.Equal(t, "192.168.1.1", ip)
	})

	t.Run("IgnoresProxyHeadersWithoutTrust", func(t *testing.T) {
		// SEC-07: X-Forwarded-For is NOT trusted when SetTrustedProxies(nil) is called
		r := gin.New()
		r.SetTrustedProxies(nil) // mirrors cmd/server.go setup

		var capturedIP string
		r.GET("/test", func(c *gin.Context) {
			capturedIP = GetClientIP(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		r.ServeHTTP(w, req)

		assert.Equal(t, "192.168.1.1", capturedIP)
	})
}

func TestHasRole(t *testing.T) {
	roles := []string{"admin", "user"}
	assert.True(t, HasRole(roles, "admin"))
	assert.False(t, HasRole(roles, "guest"))
}

func TestHasAnyRole(t *testing.T) {
	userRoles := []string{"editor", "viewer"}
	required := []string{"admin", "editor"}

	assert.True(t, HasAnyRole(userRoles, required))
	assert.False(t, HasAnyRole([]string{"viewer"}, []string{"admin"}))
}

func TestGetDeviceInfoFromContext(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("User-Agent", "Mozilla/5.0")

	info := GetDeviceInfoFromContext(c)
	assert.NotNil(t, info)
	assert.Equal(t, "unknown", info.Browser) // Basic check
}

func TestMustGetUserID_Present(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	uid := uuid.New()
	c.Set(UserIDKey, uid)

	result, ok := MustGetUserID(c)
	assert.True(t, ok)
	assert.Equal(t, uid, result)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, c.IsAborted())
}

func TestMustGetUserID_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	result, ok := MustGetUserID(c)
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, result)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, c.IsAborted())
}
