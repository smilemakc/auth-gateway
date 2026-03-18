package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// ApplicationServiceInterface defines the interface for application operations used by middleware
type ApplicationServiceInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Application, error)
}

// ApplicationAccessChecker defines the interface for checking user access to applications
type ApplicationAccessChecker interface {
	CheckUserAccess(ctx context.Context, userID, applicationID uuid.UUID) error
}

// ApplicationMiddleware handles application context extraction and validation
type ApplicationMiddleware struct {
	appService    ApplicationServiceInterface
	accessChecker ApplicationAccessChecker
	logger        *logger.Logger
}

// NewApplicationMiddleware creates a new ApplicationMiddleware
func NewApplicationMiddleware(appService ApplicationServiceInterface, accessChecker ApplicationAccessChecker, log *logger.Logger) *ApplicationMiddleware {
	return &ApplicationMiddleware{
		appService:    appService,
		accessChecker: accessChecker,
		logger:        log,
	}
}

// ExtractApplicationID extracts application ID from X-Application-ID header
// and sets it in the context. Does not require the header to be present.
func (m *ApplicationMiddleware) ExtractApplicationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.GetHeader("X-Application-ID")
		if appIDStr == "" {
			appIDStr = c.Query("app_id")
		}

		if appIDStr != "" {
			appID, err := uuid.Parse(appIDStr)
			if err != nil {
				if m.logger != nil {
					m.logger.Warn("Invalid application ID format", map[string]interface{}{
						"app_id": appIDStr,
						"error":  err.Error(),
					})
				}
				c.Next()
				return
			}

			c.Set(utils.ApplicationIDKey, appID)
		}

		c.Next()
	}
}

// RequireApplicationID requires a valid X-Application-ID header
func (m *ApplicationMiddleware) RequireApplicationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.GetHeader("X-Application-ID")
		if appIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Application-ID header is required",
			})
			c.Abort()
			return
		}

		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid application ID format",
			})
			c.Abort()
			return
		}

		c.Set(utils.ApplicationIDKey, appID)
		c.Next()
	}
}

// ValidateApplicationAccess validates that the user has access to the application
func (m *ApplicationMiddleware) ValidateApplicationAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, hasUser := utils.GetUserIDFromContext(c)
		appID, hasApp := utils.GetApplicationIDFromContext(c)

		if !hasUser || !hasApp {
			c.Next()
			return
		}

		if m.accessChecker == nil {
			c.Next()
			return
		}

		if err := m.accessChecker.CheckUserAccess(c.Request.Context(), *userID, *appID); err != nil {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.NewAppError(http.StatusForbidden, "User does not have access to this application"),
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateApplicationExists validates that the application exists and is active
func (m *ApplicationMiddleware) ValidateApplicationExists() gin.HandlerFunc {
	return func(c *gin.Context) {
		appID, hasApp := utils.GetApplicationIDFromContext(c)
		if !hasApp {
			c.Next()
			return
		}

		if m.appService != nil {
			app, err := m.appService.GetByID(c.Request.Context(), *appID)
			if err != nil {
				c.JSON(http.StatusNotFound, models.NewErrorResponse(
					models.NewAppError(http.StatusNotFound, "Application not found"),
				))
				c.Abort()
				return
			}

			if app == nil || !app.IsActive {
				c.JSON(http.StatusForbidden, models.NewErrorResponse(
					models.NewAppError(http.StatusForbidden, "Application is inactive"),
				))
				c.Abort()
				return
			}

			c.Set("application", app)
		}

		c.Next()
	}
}
