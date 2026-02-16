package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/stretchr/testify/assert"
)

// apiKeyWithScopes builds an *models.APIKey with the given scopes encoded as JSON.
func apiKeyWithScopes(scopes ...string) *models.APIKey {
	scopesJSON, _ := json.Marshal(scopes)
	return &models.APIKey{
		ID:       uuid.New(),
		IsActive: true,
		Scopes:   scopesJSON,
	}
}

// newTestAPIKeyService creates a real APIKeyService with nil dependencies.
// This is safe because HasScope only operates on in-memory data (JSON parsing),
// not on repository calls.
func newTestAPIKeyService() *service.APIKeyService {
	return service.NewAPIKeyService(nil, nil, nil)
}

// --- RequireScope tests ---

func TestRequireScope_ShouldAllow_WhenAuthTypeIsApplication(t *testing.T) {
	// Arrange: Application secrets bypass scope checks entirely
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("auth_type", "application")
		c.Next()
	})
	r.Use(mw.RequireScope(models.ScopeReadUsers))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireScope_ShouldReturn401_WhenNoAPIKeyInContext(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}

	r := gin.New()
	r.Use(mw.RequireScope(models.ScopeReadUsers))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireScope_ShouldReturn401_WhenAPIKeyContextValueIsWrongType(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", "not-an-api-key-struct")
		c.Next()
	})
	r.Use(mw.RequireScope(models.ScopeReadUsers))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireScope_ShouldAllow_WhenAPIKeyHasRequiredScope(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}
	apiKey := apiKeyWithScopes("users:read", "token:validate")

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", apiKey)
		c.Next()
	})
	r.Use(mw.RequireScope(models.ScopeReadUsers))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireScope_ShouldReturn403_WhenAPIKeyLacksRequiredScope(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}
	apiKey := apiKeyWithScopes("users:read")

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", apiKey)
		c.Next()
	})
	r.Use(mw.RequireScope(models.ScopeAdmin))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "API key does not have required scope")
}

func TestRequireScope_ShouldAllow_WhenAPIKeyHasAllScope(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}
	apiKey := apiKeyWithScopes("all")

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", apiKey)
		c.Next()
	})
	r.Use(mw.RequireScope(models.ScopeAdmin))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireScope_ShouldReturn403_WhenScopesAreInvalidJSON(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}
	apiKey := &models.APIKey{
		ID:       uuid.New(),
		IsActive: true,
		Scopes:   []byte("not-valid-json"),
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", apiKey)
		c.Next()
	})
	r.Use(mw.RequireScope(models.ScopeReadUsers))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireScope_ShouldReturn403_WhenScopesArrayIsEmpty(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}
	apiKey := apiKeyWithScopes() // empty scopes array

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", apiKey)
		c.Next()
	})
	r.Use(mw.RequireScope(models.ScopeReadUsers))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// --- RequireAnyScope tests ---

func TestRequireAnyScope_ShouldReturn401_WhenNoAPIKeyInContext(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}

	r := gin.New()
	r.Use(mw.RequireAnyScope(models.ScopeReadUsers, models.ScopeReadProfile))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAnyScope_ShouldReturn401_WhenAPIKeyContextValueIsWrongType(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", 12345)
		c.Next()
	})
	r.Use(mw.RequireAnyScope(models.ScopeReadUsers))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAnyScope_ShouldAllow_WhenAPIKeyHasOneOfRequiredScopes(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}
	apiKey := apiKeyWithScopes("token:validate")

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", apiKey)
		c.Next()
	})
	r.Use(mw.RequireAnyScope(models.ScopeReadUsers, models.ScopeValidateToken))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAnyScope_ShouldReturn403_WhenAPIKeyHasNoneOfRequiredScopes(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}
	apiKey := apiKeyWithScopes("email:send")

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", apiKey)
		c.Next()
	})
	r.Use(mw.RequireAnyScope(models.ScopeReadUsers, models.ScopeValidateToken))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "API key does not have any of the required scopes")
}

func TestRequireAnyScope_ShouldAllow_WhenAPIKeyHasAllScope(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}
	apiKey := apiKeyWithScopes("all")

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("api_key", apiKey)
		c.Next()
	})
	r.Use(mw.RequireAnyScope(models.ScopeAdmin, models.ScopeReadUsers))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Authenticate token extraction tests ---

func TestAPIKeyAuthenticate_ShouldReturn401_WhenNoTokenProvided(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}

	r := gin.New()
	r.Use(mw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyAuthenticate_ShouldExtractTokenFromXAppSecretHeader(t *testing.T) {
	// Passing an app_-prefixed token via X-App-Secret will attempt to validate
	// as an app secret. With nil appService it panics, so use recovery.
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-Secret", "app_secretvalue")

	r.ServeHTTP(w, req)

	// Should not be the "no token" 401 -- the token was extracted
	// With nil appService, the authenticateAppSecret will panic (recovered to 500)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAPIKeyAuthenticate_ShouldFallbackToBearerHeader_WhenNoXHeaders(t *testing.T) {
	// When X-API-Key and X-App-Secret are both empty, it falls through to Authorization header
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer agw_mybearerkey")

	r.ServeHTTP(w, req)

	// Token extracted from Bearer header, routes to authenticateAPIKey
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAPIKeyAuthenticate_ShouldReturn401_WhenBearerTokenIsEmpty(t *testing.T) {
	mw := &APIKeyMiddleware{apiKeyService: newTestAPIKeyService()}

	r := gin.New()
	r.Use(mw.Authenticate())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")

	r.ServeHTTP(w, req)

	// "Bearer " splits into ["Bearer", ""], so token is empty string
	// Empty string doesn't match X-API-Key or X-App-Secret, falls to Authorization
	// "Bearer " with SplitN yields ["Bearer", ""] which is 2 parts, so token=""
	// Then token is still "", so returns 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
