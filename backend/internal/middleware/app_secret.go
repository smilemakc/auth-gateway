package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

type AppSecretMiddleware struct {
	appService *service.ApplicationService
}

func NewAppSecretMiddleware(appService *service.ApplicationService) *AppSecretMiddleware {
	return &AppSecretMiddleware{appService: appService}
}

func (m *AppSecretMiddleware) RequireAppSecret() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := ""

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" && strings.HasPrefix(parts[1], "app_") {
				secret = parts[1]
			}
		}

		if secret == "" {
			secret = c.GetHeader("X-App-Secret")
		}

		if secret == "" || !strings.HasPrefix(secret, "app_") {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
				models.NewAppError(http.StatusUnauthorized, "Application secret required"),
			))
			c.Abort()
			return
		}

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
}
