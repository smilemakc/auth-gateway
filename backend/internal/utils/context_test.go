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

	t.Run("Wrong type integer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(UserIDKey, 12345)

		result, ok := GetUserIDFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Nil value", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(UserIDKey, nil)

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

	t.Run("Wrong type integer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(UserEmailKey, 12345)

		result, ok := GetUserEmailFromContext(c)
		assert.False(t, ok)
		assert.Equal(t, "", result)
	})

	t.Run("Empty string value", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(UserEmailKey, "")

		result, ok := GetUserEmailFromContext(c)
		assert.True(t, ok)
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

	t.Run("No Role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result, ok := GetUserRoleFromContext(c)
		assert.False(t, ok)
		assert.Equal(t, "", result)
	})

	t.Run("Wrong type integer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(UserRoleKey, 42)

		result, ok := GetUserRoleFromContext(c)
		assert.False(t, ok)
		assert.Equal(t, "", result)
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
	t.Run("Role present", func(t *testing.T) {
		roles := []string{"admin", "user"}
		assert.True(t, HasRole(roles, "admin"))
	})

	t.Run("Role not present", func(t *testing.T) {
		roles := []string{"admin", "user"}
		assert.False(t, HasRole(roles, "guest"))
	})

	t.Run("Empty roles slice", func(t *testing.T) {
		assert.False(t, HasRole([]string{}, "admin"))
	})

	t.Run("Nil roles slice", func(t *testing.T) {
		assert.False(t, HasRole(nil, "admin"))
	})

	t.Run("Empty role to find", func(t *testing.T) {
		roles := []string{"admin", "user"}
		assert.False(t, HasRole(roles, ""))
	})

	t.Run("Case sensitive", func(t *testing.T) {
		roles := []string{"Admin"}
		assert.False(t, HasRole(roles, "admin"))
		assert.True(t, HasRole(roles, "Admin"))
	})
}

func TestHasAnyRole(t *testing.T) {
	t.Run("Has matching role", func(t *testing.T) {
		userRoles := []string{"editor", "viewer"}
		required := []string{"admin", "editor"}
		assert.True(t, HasAnyRole(userRoles, required))
	})

	t.Run("No matching role", func(t *testing.T) {
		assert.False(t, HasAnyRole([]string{"viewer"}, []string{"admin"}))
	})

	t.Run("Empty user roles", func(t *testing.T) {
		assert.False(t, HasAnyRole([]string{}, []string{"admin"}))
	})

	t.Run("Empty required roles", func(t *testing.T) {
		assert.False(t, HasAnyRole([]string{"admin"}, []string{}))
	})

	t.Run("Both empty", func(t *testing.T) {
		assert.False(t, HasAnyRole([]string{}, []string{}))
	})

	t.Run("Nil user roles", func(t *testing.T) {
		assert.False(t, HasAnyRole(nil, []string{"admin"}))
	})

	t.Run("Nil required roles", func(t *testing.T) {
		assert.False(t, HasAnyRole([]string{"admin"}, nil))
	})

	t.Run("Both nil", func(t *testing.T) {
		assert.False(t, HasAnyRole(nil, nil))
	})

	t.Run("Multiple matches returns true", func(t *testing.T) {
		userRoles := []string{"admin", "editor"}
		required := []string{"admin", "editor"}
		assert.True(t, HasAnyRole(userRoles, required))
	})
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

func TestGetApplicationIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid UUID value", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		appID := uuid.New()
		c.Set(ApplicationIDKey, appID)

		result, ok := GetApplicationIDFromContext(c)
		assert.True(t, ok)
		assert.NotNil(t, result)
		assert.Equal(t, appID, *result)
	})

	t.Run("Valid UUID pointer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		appID := uuid.New()
		c.Set(ApplicationIDKey, &appID)

		result, ok := GetApplicationIDFromContext(c)
		assert.True(t, ok)
		assert.NotNil(t, result)
		assert.Equal(t, appID, *result)
	})

	t.Run("Nil UUID pointer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ApplicationIDKey, (*uuid.UUID)(nil))

		result, ok := GetApplicationIDFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Valid string UUID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		appID := uuid.New()
		c.Set(ApplicationIDKey, appID.String())

		result, ok := GetApplicationIDFromContext(c)
		assert.True(t, ok)
		assert.NotNil(t, result)
		assert.Equal(t, appID, *result)
	})

	t.Run("Empty string", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ApplicationIDKey, "")

		result, ok := GetApplicationIDFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Invalid string UUID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ApplicationIDKey, "not-a-valid-uuid")

		result, ok := GetApplicationIDFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Not set in context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result, ok := GetApplicationIDFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Wrong type integer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ApplicationIDKey, 12345)

		result, ok := GetApplicationIDFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Wrong type bool", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ApplicationIDKey, true)

		result, ok := GetApplicationIDFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})
}

func TestSetApplicationIDInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Set and retrieve", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		appID := uuid.New()

		SetApplicationIDInContext(c, appID)

		result, ok := GetApplicationIDFromContext(c)
		assert.True(t, ok)
		assert.NotNil(t, result)
		assert.Equal(t, appID, *result)
	})

	t.Run("Overwrite existing value", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		firstID := uuid.New()
		secondID := uuid.New()

		SetApplicationIDInContext(c, firstID)
		SetApplicationIDInContext(c, secondID)

		result, ok := GetApplicationIDFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, secondID, *result)
	})
}

func TestGetUserRolesFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid roles", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		roles := []string{"admin", "editor"}
		c.Set(UserRolesKey, roles)

		result, ok := GetUserRolesFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, roles, result)
	})

	t.Run("No roles set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result, ok := GetUserRolesFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Wrong type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(UserRolesKey, "admin")

		result, ok := GetUserRolesFromContext(c)
		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("Empty roles slice", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(UserRolesKey, []string{})

		result, ok := GetUserRolesFromContext(c)
		assert.True(t, ok)
		assert.Empty(t, result)
	})

	t.Run("Nil slice stored", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(UserRolesKey, ([]string)(nil))

		result, ok := GetUserRolesFromContext(c)
		assert.True(t, ok)
		assert.Nil(t, result)
	})
}

func TestGetTokenFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(TokenKey, "eyJhbGciOiJIUzI1NiJ9.test.signature")

		result, ok := GetTokenFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, "eyJhbGciOiJIUzI1NiJ9.test.signature", result)
	})

	t.Run("No token set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result, ok := GetTokenFromContext(c)
		assert.False(t, ok)
		assert.Equal(t, "", result)
	})

	t.Run("Empty string token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(TokenKey, "")

		result, ok := GetTokenFromContext(c)
		assert.False(t, ok, "empty token should return false")
		assert.Equal(t, "", result)
	})

	t.Run("Wrong type integer", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(TokenKey, 12345)

		result, ok := GetTokenFromContext(c)
		assert.False(t, ok)
		assert.Equal(t, "", result)
	})
}

func TestGetUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("User-Agent present", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

		ua := GetUserAgent(c)
		assert.Equal(t, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", ua)
	})

	t.Run("No User-Agent header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", nil)

		ua := GetUserAgent(c)
		assert.Equal(t, "", ua)
	})
}

func TestContextKeys_ShouldHaveExpectedValues(t *testing.T) {
	assert.Equal(t, "user_id", UserIDKey)
	assert.Equal(t, "user_email", UserEmailKey)
	assert.Equal(t, "user_role", UserRoleKey)
	assert.Equal(t, "user_roles", UserRolesKey)
	assert.Equal(t, "access_token", TokenKey)
	assert.Equal(t, "application_id", ApplicationIDKey)
}
