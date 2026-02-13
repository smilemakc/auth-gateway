package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
)

// AuthMiddleware validates JWT tokens and sets user context
type AuthMiddleware struct {
	jwtService       *jwt.Service
	blacklistService *service.BlacklistService
	apiKeyMiddleware *APIKeyMiddleware
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *jwt.Service, blacklistService *service.BlacklistService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:       jwtService,
		blacklistService: blacklistService,
	}
}

// SetAPIKeyMiddleware sets the API key middleware for combined auth
func (m *AuthMiddleware) SetAPIKeyMiddleware(apiKeyMw *APIKeyMiddleware) {
	m.apiKeyMiddleware = apiKeyMw
}

// Authenticate validates JWT token, API key, or application secret.
// Priority: X-API-Key / X-App-Secret / Bearer agw_ / Bearer app_ → delegate to APIKeyMiddleware.
// Otherwise treat as JWT.
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request carries an API key or app secret
		if m.apiKeyMiddleware != nil && m.isAPIKeyOrAppSecret(c) {
			m.apiKeyMiddleware.Authenticate()(c)
			return
		}

		// Fall through to JWT authentication
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrInvalidToken))
			c.Abort()
			return
		}

		token := parts[1]

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			if errors.Is(err, jwt.ErrExpiredToken) {
				c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrTokenExpired))
			} else {
				c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrInvalidToken))
			}
			c.Abort()
			return
		}

		tokenHash := utils.HashToken(token)
		if m.blacklistService.IsBlacklisted(c.Request.Context(), tokenHash) {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrTokenRevoked))
			c.Abort()
			return
		}

		c.Set(utils.UserIDKey, claims.UserID)
		c.Set(utils.UserEmailKey, claims.Email)
		c.Set(utils.UserRolesKey, claims.Roles)
		c.Set(utils.TokenKey, token)

		if claims.ApplicationID != nil {
			if _, exists := utils.GetApplicationIDFromContext(c); !exists {
				c.Set(utils.ApplicationIDKey, *claims.ApplicationID)
			}
		}

		c.Next()
	}
}

// isAPIKeyOrAppSecret checks if the request carries API key or app secret credentials.
func (m *AuthMiddleware) isAPIKeyOrAppSecret(c *gin.Context) bool {
	if c.GetHeader("X-API-Key") != "" || c.GetHeader("X-App-Secret") != "" {
		return true
	}
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			if strings.HasPrefix(token, "agw_") || strings.HasPrefix(token, "app_") {
				return true
			}
		}
	}
	return false
}

// RequireRole checks if user has the required role
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := utils.GetUserRolesFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		// Admin has access to everything
		if contains(roles, string(models.RoleAdmin)) {
			c.Next()
			return
		}

		// Check if user has required role
		if !contains(roles, requiredRole) {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrForbidden))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole checks if user has any of the required roles
func (m *AuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := utils.GetUserRolesFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		// Admin has access to everything
		if contains(userRoles, string(models.RoleAdmin)) {
			c.Next()
			return
		}

		// Check if user has any of the required roles
		for _, requiredRole := range roles {
			if contains(userRoles, requiredRole) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrForbidden))
		c.Abort()
	}
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
