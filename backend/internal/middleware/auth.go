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
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *jwt.Service, blacklistService *service.BlacklistService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:       jwtService,
		blacklistService: blacklistService,
	}
}

// Authenticate validates the JWT token and sets user context
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		// Check if token starts with "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrInvalidToken))
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
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

		// Check if token is blacklisted using unified blacklist service
		tokenHash := utils.HashToken(token)
		if m.blacklistService.IsBlacklisted(c.Request.Context(), tokenHash) {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrTokenRevoked))
			c.Abort()
			return
		}

		// Set user context
		c.Set(utils.UserIDKey, claims.UserID)
		c.Set(utils.UserEmailKey, claims.Email)
		c.Set(utils.UserRolesKey, claims.Roles)

		c.Next()
	}
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
