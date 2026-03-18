package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/stretchr/testify/assert"
)

// --- RequireAdmin tests ---

func TestRequireAdmin_ShouldAllow_WhenUserIsAdmin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"admin"})
		c.Next()
	})
	r.Use(RequireAdmin())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdmin_ShouldAllow_WhenAuthTypeIsApplication(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("auth_type", "application")
		c.Next()
	})
	r.Use(RequireAdmin())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "application secrets should bypass admin check")
}

func TestRequireAdmin_ShouldAllow_WhenAuthTypeIsAPIKey(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("auth_type", "api_key")
		c.Next()
	})
	r.Use(RequireAdmin())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "API keys should bypass admin check")
}

func TestRequireAdmin_ShouldReturn403_WhenUserIsNotAdmin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"user", "moderator"})
		c.Next()
	})
	r.Use(RequireAdmin())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Admin access required")
}

func TestRequireAdmin_ShouldReturn401_WhenNoRolesInContext(t *testing.T) {
	r := gin.New()
	r.Use(RequireAdmin())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAdmin_ShouldReturn401_WhenEmptyRoles(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{})
		c.Next()
	})
	r.Use(RequireAdmin())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- RequireAdminOrModerator tests ---

func TestRequireAdminOrModerator_ShouldAllow_WhenUserIsAdmin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"admin"})
		c.Next()
	})
	r.Use(RequireAdminOrModerator())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdminOrModerator_ShouldAllow_WhenUserIsModerator(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"moderator"})
		c.Next()
	})
	r.Use(RequireAdminOrModerator())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdminOrModerator_ShouldReturn403_WhenUserIsNeitherAdminNorModerator(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"user", "editor"})
		c.Next()
	})
	r.Use(RequireAdminOrModerator())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Admin or moderator access required")
}

func TestRequireAdminOrModerator_ShouldReturn401_WhenNoRolesInContext(t *testing.T) {
	r := gin.New()
	r.Use(RequireAdminOrModerator())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAdminOrModerator_ShouldAllow_WhenUserHasBothRoles(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(utils.UserRolesKey, []string{"admin", "moderator"})
		c.Next()
	})
	r.Use(RequireAdminOrModerator())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
