package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

// APIKeyMiddleware validates API keys and application secrets
type APIKeyMiddleware struct {
	apiKeyService *service.APIKeyService
	appService    *service.ApplicationService
	rbacRepo      *repository.RBACRepository
}

// NewAPIKeyMiddleware creates a new API key middleware
func NewAPIKeyMiddleware(apiKeyService *service.APIKeyService, appService *service.ApplicationService, rbacRepo *repository.RBACRepository) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		apiKeyService: apiKeyService,
		appService:    appService,
		rbacRepo:      rbacRepo,
	}
}

// Authenticate validates the API key or application secret and sets context.
// Supports: X-API-Key header, X-App-Secret header, Authorization: Bearer agw_/app_
func (m *APIKeyMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-API-Key")
		if token == "" {
			token = c.GetHeader("X-App-Secret")
		}
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		// Dispatch based on token prefix
		if strings.HasPrefix(token, "app_") {
			m.authenticateAppSecret(c, token)
		} else {
			m.authenticateAPIKey(c, token)
		}
	}
}

func (m *APIKeyMiddleware) authenticateAPIKey(c *gin.Context, apiKey string) {
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

	c.Set(utils.UserIDKey, user.ID)
	c.Set(utils.UserEmailKey, user.Email)
	c.Set("api_key_id", key.ID)
	c.Set("api_key", key)
	c.Set("auth_type", "api_key")

	ctx := c.Request.Context()
	roles, err := m.rbacRepo.GetUserRoles(ctx, user.ID)
	if err == nil {
		roleNames := make([]string, len(roles))
		for i, role := range roles {
			roleNames[i] = role.Name
		}
		c.Set(utils.UserRolesKey, roleNames)
	} else {
		c.Set(utils.UserRolesKey, []string{})
	}

	c.Next()
}

func (m *APIKeyMiddleware) authenticateAppSecret(c *gin.Context, secret string) {
	app, err := m.appService.ValidateSecret(c.Request.Context(), secret)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.NewAppError(http.StatusUnauthorized, "Invalid application secret"),
			))
		}
		c.Abort()
		return
	}

	c.Set("application", app)
	c.Set(utils.ApplicationIDKey, app.ID)
	c.Set("auth_type", "application")

	c.Next()
}

// RequireScope checks if the API key has the required scope.
// Application secrets bypass scope checks (full access).
func (m *APIKeyMiddleware) RequireScope(scope models.APIKeyScope) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Application secrets have full access
		if authType, _ := c.Get("auth_type"); authType == "application" {
			c.Next()
			return
		}

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
