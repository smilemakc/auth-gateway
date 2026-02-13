package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

// RequireAdmin checks if the user has admin role.
// Application secrets and API keys bypass the role check (service-level access).
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authType, exists := c.Get("auth_type"); exists {
			if authType == "application" || authType == "api_key" {
				c.Next()
				return
			}
		}

		roles, exists := utils.GetUserRolesFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		if !utils.HasRole(roles, "admin") {
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
		roles, exists := utils.GetUserRolesFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
			c.Abort()
			return
		}

		if !utils.HasAnyRole(roles, []string{"admin", "moderator"}) {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.NewAppError(http.StatusForbidden, "Admin or moderator access required"),
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
