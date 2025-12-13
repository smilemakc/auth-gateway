package middleware

import (
	"net/http"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"

	"github.com/gin-gonic/gin"
)

// MaintenanceMiddleware checks if the system is in maintenance mode
type MaintenanceMiddleware struct {
	systemRepo *repository.SystemRepository
}

// NewMaintenanceMiddleware creates a new maintenance middleware
func NewMaintenanceMiddleware(systemRepo *repository.SystemRepository) *MaintenanceMiddleware {
	return &MaintenanceMiddleware{
		systemRepo: systemRepo,
	}
}

// CheckMaintenance checks if the system is in maintenance mode
func (m *MaintenanceMiddleware) CheckMaintenance() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip check for health endpoints
		if c.Request.URL.Path == "/auth/health" ||
			c.Request.URL.Path == "/auth/ready" ||
			c.Request.URL.Path == "/auth/live" {
			c.Next()
			return
		}

		// Get maintenance mode setting
		setting, err := m.systemRepo.GetSetting(c.Request.Context(), models.SettingMaintenanceMode)
		if err != nil {
			// On error, allow (fail-open)
			c.Next()
			return
		}

		if setting.Value == "true" {
			// Get maintenance message
			messageSetting, _ := m.systemRepo.GetSetting(c.Request.Context(), models.SettingMaintenanceMessage)
			message := "System is under maintenance. Please try again later."
			if messageSetting != nil {
				message = messageSetting.Value
			}

			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error: message,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
