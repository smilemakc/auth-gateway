package middleware

import (
	"net/http"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RBACMiddleware provides permission-based access control
type RBACMiddleware struct {
	rbacService *service.RBACService
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(rbacService *service.RBACService) *RBACMiddleware {
	return &RBACMiddleware{
		rbacService: rbacService,
	}
}

// RequirePermission checks if the user has a specific permission
func (m *RBACMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
			c.Abort()
			return
		}

		hasPermission, err := m.rbacService.CheckUserPermission(c.Request.Context(), userID.(uuid.UUID), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission checks if the user has any of the specified permissions
func (m *RBACMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
			c.Abort()
			return
		}

		hasPermission, err := m.rbacService.CheckUserAnyPermission(c.Request.Context(), userID.(uuid.UUID), permissions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions checks if the user has all of the specified permissions
func (m *RBACMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
			c.Abort()
			return
		}

		hasPermission, err := m.rbacService.CheckUserAllPermissions(c.Request.Context(), userID.(uuid.UUID), permissions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to check permissions"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
