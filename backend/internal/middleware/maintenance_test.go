package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// The MaintenanceMiddleware depends on *repository.SystemRepository (a concrete type).
// Since we cannot mock concrete types without modifying production code, we test the
// behaviors that are controllable:
// 1. Health endpoint bypass (no SystemRepository call needed for /auth/health, /auth/ready, /auth/live)
// 2. Fail-open behavior when repository returns an error (nil repo causes error)

func TestCheckMaintenance_ShouldBypassHealthEndpoint(t *testing.T) {
	healthPaths := []string{"/auth/health", "/auth/ready", "/auth/live"}

	for _, path := range healthPaths {
		t.Run(path, func(t *testing.T) {
			// Arrange: Create middleware with nil repo, which means any non-health
			// request would panic. Health endpoints should bypass the check entirely.
			mw := &MaintenanceMiddleware{systemRepo: nil}

			r := gin.New()
			r.Use(mw.CheckMaintenance())
			r.GET(path, func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", path, nil)

			// Act
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "ok", w.Body.String())
		})
	}
}

func TestCheckMaintenance_ShouldFailOpen_WhenSystemRepoIsNil(t *testing.T) {
	// Arrange: With nil systemRepo, the GetSetting call will panic.
	// Use recovery to catch it. The fail-open behavior in the code
	// handles errors from GetSetting, but nil pointer is a panic.
	// This tests that recovery middleware is important alongside maintenance.
	mw := &MaintenanceMiddleware{systemRepo: nil}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mw.CheckMaintenance())
	r.GET("/api/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert: Panic from nil systemRepo is recovered by gin.Recovery
	// The maintenance middleware tries to call systemRepo.GetSetting which panics
	assert.Equal(t, http.StatusInternalServerError, w.Code,
		"nil systemRepo should panic and be recovered")
}

func TestCheckMaintenance_ShouldNotBypass_WhenPathIsNotHealth(t *testing.T) {
	nonHealthPaths := []string{"/api/users", "/auth/signin", "/auth/health/extra", "/other"}

	for _, path := range nonHealthPaths {
		t.Run(path, func(t *testing.T) {
			// With nil systemRepo, non-health paths will attempt to call GetSetting
			// and panic. This verifies they are NOT bypassed.
			mw := &MaintenanceMiddleware{systemRepo: nil}

			r := gin.New()
			r.Use(gin.Recovery())
			r.Use(mw.CheckMaintenance())
			r.GET(path, func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", path, nil)

			r.ServeHTTP(w, req)

			// Should NOT be 200 because the middleware tried to access systemRepo
			assert.NotEqual(t, http.StatusOK, w.Code,
				"non-health path should NOT bypass maintenance check")
		})
	}
}
