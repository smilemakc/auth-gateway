package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

// RequireAdmin checks if the user has admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by auth middleware)
		userRole, exists := utils.GetUserRoleFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		// Check if user is admin
		if userRole != string(models.RoleAdmin) {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.NewAppError(http.StatusForbidden, "Admin access required"),
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminOrModerator checks if the user has admin or moderator role
func RequireAdminOrModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := utils.GetUserRoleFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		if userRole != string(models.RoleAdmin) && userRole != string(models.RoleModerator) {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.NewAppError(http.StatusForbidden, "Admin or moderator access required"),
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
