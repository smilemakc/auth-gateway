package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

// APIKeyMiddleware validates API keys
type APIKeyMiddleware struct {
	apiKeyService *service.APIKeyService
}

// NewAPIKeyMiddleware creates a new API key middleware
func NewAPIKeyMiddleware(apiKeyService *service.APIKeyService) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		apiKeyService: apiKeyService,
	}
}

// Authenticate validates the API key and sets user context
func (m *APIKeyMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		// Support both "X-API-Key" and "Authorization: Bearer <key>" headers
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer agw_") {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 {
					apiKey = parts[1]
				}
			}
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		// Validate API key
		key, user, err := m.apiKeyService.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			if appErr, ok := err.(*models.AppError); ok {
				c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			} else {
				c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrInvalidToken))
			}
			c.Abort()
			return
		}

		// Set user context
		c.Set(utils.UserIDKey, user.ID)
		c.Set(utils.UserEmailKey, user.Email)
		c.Set(utils.UserRoleKey, user.Role)
		c.Set("api_key_id", key.ID)
		c.Set("api_key", key)

		c.Next()
	}
}

// RequireScope checks if the API key has the required scope
func (m *APIKeyMiddleware) RequireScope(scope models.APIKeyScope) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from context
		apiKeyVal, exists := c.Get("api_key")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		apiKey, ok := apiKeyVal.(*models.APIKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		// Check scope
		if !m.apiKeyService.HasScope(apiKey, scope) {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.NewAppError(http.StatusForbidden, "API key does not have required scope: "+string(scope)),
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyScope checks if the API key has any of the required scopes
func (m *APIKeyMiddleware) RequireAnyScope(scopes ...models.APIKeyScope) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from context
		apiKeyVal, exists := c.Get("api_key")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		apiKey, ok := apiKeyVal.(*models.APIKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		// Check if has any of the required scopes
		hasScope := false
		for _, scope := range scopes {
			if m.apiKeyService.HasScope(apiKey, scope) {
				hasScope = true
				break
			}
		}

		if !hasScope {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.NewAppError(http.StatusForbidden, "API key does not have any of the required scopes"),
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
