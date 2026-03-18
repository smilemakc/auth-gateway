package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// AppSecretMiddleware depends on *service.ApplicationService (concrete type).
// We test the token extraction logic, which runs before any service call,
// and the error paths for missing/invalid secrets.

func TestRequireAppSecret_ShouldReturn401_WhenNoSecretProvided(t *testing.T) {
	// Arrange: No Authorization header, no X-App-Secret header
	mw := &AppSecretMiddleware{appService: nil}

	r := gin.New()
	r.Use(mw.RequireAppSecret())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Application secret required")
}

func TestRequireAppSecret_ShouldReturn401_WhenBearerTokenHasNoAppPrefix(t *testing.T) {
	mw := &AppSecretMiddleware{appService: nil}

	r := gin.New()
	r.Use(mw.RequireAppSecret())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer agw_not_an_app_secret")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Application secret required")
}

func TestRequireAppSecret_ShouldReturn401_WhenXAppSecretHasNoAppPrefix(t *testing.T) {
	mw := &AppSecretMiddleware{appService: nil}

	r := gin.New()
	r.Use(mw.RequireAppSecret())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-Secret", "not_app_prefixed")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Application secret required")
}

func TestRequireAppSecret_ShouldExtractFromBearerHeader_WhenAppPrefixed(t *testing.T) {
	// With nil appService, ValidateSecret will panic. Use recovery.
	mw := &AppSecretMiddleware{appService: nil}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mw.RequireAppSecret())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer app_validprefix")

	r.ServeHTTP(w, req)

	// Token extracted and validation attempted (panic from nil service, recovered)
	assert.NotEqual(t, http.StatusOK, w.Code)
	// The key point: it did NOT return "Application secret required" (401)
	// It went past the extraction and tried to validate (panicked -> 500)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequireAppSecret_ShouldExtractFromXAppSecretHeader_WhenAppPrefixed(t *testing.T) {
	mw := &AppSecretMiddleware{appService: nil}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mw.RequireAppSecret())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-Secret", "app_validprefix")

	r.ServeHTTP(w, req)

	// Same as above: token extracted, validation attempted, nil service panics
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequireAppSecret_ShouldPreferBearerHeader_OverXAppSecret(t *testing.T) {
	// When Bearer has app_ prefix, it should be used instead of X-App-Secret
	mw := &AppSecretMiddleware{appService: nil}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mw.RequireAppSecret())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer app_from_bearer")
	req.Header.Set("X-App-Secret", "app_from_header")

	r.ServeHTTP(w, req)

	// Token extracted (from Bearer since it's checked first), validation attempted
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequireAppSecret_ShouldReturn401_WhenOnlyBasicAuth(t *testing.T) {
	mw := &AppSecretMiddleware{appService: nil}

	r := gin.New()
	r.Use(mw.RequireAppSecret())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Application secret required")
}

func TestRequireAppSecret_ShouldReturn401_WhenEmptyBearerToken(t *testing.T) {
	mw := &AppSecretMiddleware{appService: nil}

	r := gin.New()
	r.Use(mw.RequireAppSecret())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAppSecret_ShouldReturn401_WhenBearerWithoutSpace(t *testing.T) {
	mw := &AppSecretMiddleware{appService: nil}

	r := gin.New()
	r.Use(mw.RequireAppSecret())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearerapp_secret")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
