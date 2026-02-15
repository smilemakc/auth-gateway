package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

type ApplicationHandler struct {
	appService *service.ApplicationService
	logger     *logger.Logger
}

func NewApplicationHandler(appService *service.ApplicationService, logger *logger.Logger) *ApplicationHandler {
	return &ApplicationHandler{
		appService: appService,
		logger:     logger,
	}
}

// CreateApplication creates a new application
// @Summary Create application
// @Description Create a new application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateApplicationRequest true "Application creation data"
// @Success 201 {object} models.Application
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications [post]
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	var req models.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	app, secret, err := h.appService.CreateApplication(c.Request.Context(), &req, userID)
	if err != nil {
		if err == service.ErrInvalidApplicationName {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Invalid application name: must be lowercase alphanumeric with hyphens only"),
			))
			return
		}
		if err == service.ErrApplicationNameExists {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				models.NewAppError(http.StatusConflict, "Application with this name already exists"),
			))
			return
		}
		h.logger.Error("Failed to create application", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"application": app,
		"secret":      secret,
		"warning":     "Store this secret securely. It will not be shown again.",
	})
}

// GetApplication retrieves an application by ID
// @Summary Get application
// @Description Get application details by ID (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Success 200 {object} models.Application
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id} [get]
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	id, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	app, err := h.appService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrApplicationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Application not found"),
			))
			return
		}
		h.logger.Error("Failed to get application", map[string]interface{}{
			"error":          err.Error(),
			"application_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, app)
}

// ListApplications returns a paginated list of applications
// @Summary List applications
// @Description Get paginated list of applications (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} models.ApplicationListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications [get]
func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	page, perPage := utils.ParsePagination(c)

	var isActive *bool
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		val := isActiveStr == "true"
		isActive = &val
	}

	response, err := h.appService.ListApplications(c.Request.Context(), page, perPage, isActive)
	if err != nil {
		h.logger.Error("Failed to list applications", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateApplication updates an application
// @Summary Update application
// @Description Update application details (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param request body models.UpdateApplicationRequest true "Application update data"
// @Success 200 {object} models.Application
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id} [put]
func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	id, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	var req models.UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	app, err := h.appService.UpdateApplication(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrApplicationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Application not found"),
			))
			return
		}
		h.logger.Error("Failed to update application", map[string]interface{}{
			"error":          err.Error(),
			"application_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, app)
}

// DeleteApplication soft-deletes an application
// @Summary Delete application
// @Description Soft delete an application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id} [delete]
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	id, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	if err := h.appService.DeleteApplication(c.Request.Context(), id); err != nil {
		if err == service.ErrApplicationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Application not found"),
			))
			return
		}
		if err == service.ErrCannotDeleteSystemApp {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(
				models.NewAppError(http.StatusForbidden, "Cannot delete system application"),
			))
			return
		}
		h.logger.Error("Failed to delete application", map[string]interface{}{
			"error":          err.Error(),
			"application_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetBranding retrieves application branding
// @Summary Get application branding
// @Description Get branding configuration for an application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Success 200 {object} models.ApplicationBranding
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/branding [get]
func (h *ApplicationHandler) GetBranding(c *gin.Context) {
	id, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	branding, err := h.appService.GetBranding(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get branding", map[string]interface{}{
			"error":          err.Error(),
			"application_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, branding)
}

// UpdateBranding updates application branding
// @Summary Update application branding
// @Description Update branding configuration for an application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param request body models.UpdateApplicationBrandingRequest true "Branding update data"
// @Success 200 {object} models.ApplicationBranding
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/branding [put]
func (h *ApplicationHandler) UpdateBranding(c *gin.Context) {
	id, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	var req models.UpdateApplicationBrandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	branding, err := h.appService.UpdateBranding(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("Failed to update branding", map[string]interface{}{
			"error":          err.Error(),
			"application_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, branding)
}

// ListApplicationUsers returns users of an application
// @Summary List application users
// @Description Get paginated list of users for an application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.UserAppProfileListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/users [get]
func (h *ApplicationHandler) ListApplicationUsers(c *gin.Context) {
	id, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	page, perPage := utils.ParsePagination(c)

	response, err := h.appService.ListApplicationUsers(c.Request.Context(), id, page, perPage)
	if err != nil {
		h.logger.Error("Failed to list application users", map[string]interface{}{
			"error":          err.Error(),
			"application_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, response)
}

// BanUser bans a user from an application
// @Summary Ban user from application
// @Description Ban a user from accessing an application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param user_id path string true "User ID (UUID)"
// @Param request body object{reason=string} true "Ban reason"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/users/{user_id}/ban [post]
func (h *ApplicationHandler) BanUser(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	userID, err := h.parseIDParam(c, "user_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	adminID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	if err := h.appService.BanUser(c.Request.Context(), userID, appID, *adminID, req.Reason); err != nil {
		if err == service.ErrUserProfileNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "User profile not found in this application"),
			))
			return
		}
		h.logger.Error("Failed to ban user", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
			"user_id":        userID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "User banned successfully"})
}

// UnbanUser unbans a user from an application
// @Summary Unban user from application
// @Description Unban a user from an application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param user_id path string true "User ID (UUID)"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/users/{user_id}/unban [post]
func (h *ApplicationHandler) UnbanUser(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	userID, err := h.parseIDParam(c, "user_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	if err := h.appService.UnbanUser(c.Request.Context(), userID, appID); err != nil {
		if err == service.ErrUserProfileNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "User profile not found in this application"),
			))
			return
		}
		h.logger.Error("Failed to unban user", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
			"user_id":        userID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "User unbanned successfully"})
}

// GetApplicationUserProfile returns a specific user's profile in an application
// @Summary Get user profile in application
// @Description Get a specific user's profile within an application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} models.UserApplicationProfile
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/users/{user_id} [get]
func (h *ApplicationHandler) GetApplicationUserProfile(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		return
	}
	userID, err := h.parseIDParam(c, "user_id")
	if err != nil {
		return
	}

	profile, err := h.appService.GetUserProfile(c.Request.Context(), userID, appID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.NewAppError(http.StatusNotFound, "PROFILE_NOT_FOUND", "User profile not found in this application"),
		))
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateApplicationUserProfile updates a specific user's profile in an application
// @Summary Update user profile in application
// @Description Update a specific user's profile within an application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param user_id path string true "User ID"
// @Param request body models.UpdateUserAppProfileRequest true "Profile update data"
// @Success 200 {object} models.UserApplicationProfile
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/users/{user_id} [put]
func (h *ApplicationHandler) UpdateApplicationUserProfile(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		return
	}
	userID, err := h.parseIDParam(c, "user_id")
	if err != nil {
		return
	}

	var req models.UpdateUserAppProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "INVALID_REQUEST", err.Error()),
		))
		return
	}

	profile, err := h.appService.UpdateUserProfile(c.Request.Context(), userID, appID, &req)
	if err != nil {
		if err.Error() == "profile not found" {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "PROFILE_NOT_FOUND", "User profile not found"),
			))
			return
		}
		h.logger.Error("Failed to update user profile", map[string]interface{}{
			"error":          err.Error(),
			"user_id":        userID.String(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, profile)
}

// DeleteApplicationUserProfile removes a user's profile from an application
// @Summary Delete user profile from application
// @Description Remove a user's profile from an application (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Param id path string true "Application ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/users/{user_id} [delete]
func (h *ApplicationHandler) DeleteApplicationUserProfile(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		return
	}
	userID, err := h.parseIDParam(c, "user_id")
	if err != nil {
		return
	}

	if err := h.appService.DeleteUserProfile(c.Request.Context(), userID, appID); err != nil {
		h.logger.Error("Failed to delete user profile", map[string]interface{}{
			"error":          err.Error(),
			"user_id":        userID.String(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "User profile removed from application"})
}

// GetPublicBranding retrieves public branding for an application
// @Summary Get public application branding
// @Description Get public branding configuration for an application (no auth required)
// @Tags Applications - Public
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Success 200 {object} models.PublicApplicationBrandingResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/applications/{id}/branding [get]
func (h *ApplicationHandler) GetPublicBranding(c *gin.Context) {
	id, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	branding, err := h.appService.GetBranding(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get public branding", map[string]interface{}{
			"error":          err.Error(),
			"application_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	if branding == nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.NewAppError(http.StatusNotFound, "Branding not found"),
		))
		return
	}

	c.JSON(http.StatusOK, branding.ToPublicResponse())
}

// GetMyProfile retrieves the current user's profile for an application
// @Summary Get my application profile
// @Description Get current user's profile for an application (creates if not exists)
// @Tags Applications - User
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Success 200 {object} models.UserApplicationProfile
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/applications/{id}/profile [get]
func (h *ApplicationHandler) GetMyProfile(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	profile, err := h.appService.GetOrCreateUserProfile(c.Request.Context(), *userID, appID)
	if err != nil {
		h.logger.Error("Failed to get user profile", map[string]interface{}{
			"error":          err.Error(),
			"user_id":        userID.String(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateMyProfile updates the current user's profile for an application
// @Summary Update my application profile
// @Description Update current user's profile for an application
// @Tags Applications - User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param request body models.UpdateUserAppProfileRequest true "Profile update data"
// @Success 200 {object} models.UserApplicationProfile
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/applications/{id}/profile [put]
func (h *ApplicationHandler) UpdateMyProfile(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	var req models.UpdateUserAppProfileRequest
	req.IsActive = nil
	req.IsBanned = nil
	req.BanReason = nil

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	profile, err := h.appService.UpdateUserProfile(c.Request.Context(), *userID, appID, &req)
	if err != nil {
		if err == service.ErrUserProfileNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "User profile not found"),
			))
			return
		}
		h.logger.Error("Failed to update user profile", map[string]interface{}{
			"error":          err.Error(),
			"user_id":        userID.String(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, profile)
}

// RotateSecret rotates the application secret
// @Summary Rotate application secret
// @Description Rotate the application secret, invalidating the old one (admin only)
// @Tags Admin - Applications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Success 200 {object} object{secret=string,prefix=string,rotated_at=string,warning=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/rotate-secret [post]
func (h *ApplicationHandler) RotateSecret(c *gin.Context) {
	id, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	secret, err := h.appService.RotateSecret(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else if err == service.ErrApplicationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Application not found"),
			))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.NewAppError(http.StatusInternalServerError, "Failed to rotate secret"),
			))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":     secret,
		"prefix":     secret[:12],
		"rotated_at": time.Now(),
		"warning":    "Store this secret securely. It will not be shown again.",
	})
}

func (h *ApplicationHandler) parseIDParam(c *gin.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// GetAuthConfig retrieves auth configuration for the calling application
// @Summary Get application auth config
// @Description Get authentication configuration (requires app_ secret)
// @Tags Applications - Config
// @Produce json
// @Success 200 {object} models.AuthConfigResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/applications/config [get]
func (h *ApplicationHandler) GetAuthConfig(c *gin.Context) {
	appInterface, exists := c.Get("application")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.NewAppError(http.StatusUnauthorized, "Application secret required"),
		))
		return
	}

	app, ok := appInterface.(*models.Application)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	config, err := h.appService.GetAuthConfig(c.Request.Context(), app)
	if err != nil {
		h.logger.Error("Failed to get auth config", map[string]interface{}{
			"error":          err.Error(),
			"application_id": app.ID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, config)
}

