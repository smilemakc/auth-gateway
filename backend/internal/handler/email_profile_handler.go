package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

type EmailProfileHandler struct {
	emailProfileService *service.EmailProfileService
	logger              *logger.Logger
}

func NewEmailProfileHandler(emailProfileService *service.EmailProfileService, logger *logger.Logger) *EmailProfileHandler {
	return &EmailProfileHandler{
		emailProfileService: emailProfileService,
		logger:              logger,
	}
}

// ListProviders handles listing all email providers
// @Summary List email providers
// @Description Get a list of all configured email providers
// @Tags Email Providers
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.EmailProviderResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-providers [get]
func (h *EmailProfileHandler) ListProviders(c *gin.Context) {
	appID, _ := utils.GetApplicationIDFromContext(c)
	providers, err := h.emailProfileService.ListProviders(c.Request.Context(), appID)
	if err != nil {
		h.logger.Error("Failed to list email providers", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, providers)
}

// GetProvider handles retrieving a specific email provider
// @Summary Get email provider
// @Description Get details of a specific email provider by ID
// @Tags Email Providers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Provider ID (UUID)"
// @Success 200 {object} models.EmailProviderResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-providers/{id} [get]
func (h *EmailProfileHandler) GetProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	provider, err := h.emailProfileService.GetProvider(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to get email provider", map[string]interface{}{
			"error":       err.Error(),
			"provider_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, provider)
}

// CreateProvider handles creating a new email provider
// @Summary Create email provider
// @Description Create a new email provider configuration
// @Tags Email Providers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateEmailProviderRequest true "Create provider request"
// @Success 201 {object} models.EmailProvider
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-providers [post]
func (h *EmailProfileHandler) CreateProvider(c *gin.Context) {
	var req models.CreateEmailProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	provider, err := h.emailProfileService.CreateProvider(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to create email provider", map[string]interface{}{
			"error": err.Error(),
			"name":  req.Name,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// UpdateProvider handles updating an existing email provider
// @Summary Update email provider
// @Description Update an existing email provider configuration
// @Tags Email Providers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Provider ID (UUID)"
// @Param request body models.UpdateEmailProviderRequest true "Update provider request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-providers/{id} [put]
func (h *EmailProfileHandler) UpdateProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	var req models.UpdateEmailProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	if err := h.emailProfileService.UpdateProvider(c.Request.Context(), id, &req); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to update email provider", map[string]interface{}{
			"error":       err.Error(),
			"provider_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email provider updated successfully"})
}

// DeleteProvider handles deleting an email provider
// @Summary Delete email provider
// @Description Delete an email provider configuration
// @Tags Email Providers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Provider ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-providers/{id} [delete]
func (h *EmailProfileHandler) DeleteProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	if err := h.emailProfileService.DeleteProvider(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to delete email provider", map[string]interface{}{
			"error":       err.Error(),
			"provider_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email provider deleted successfully"})
}

// TestProvider handles testing an email provider
// @Summary Test email provider
// @Description Test an email provider by sending a test email
// @Tags Email Providers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Provider ID (UUID)"
// @Param request body map[string]string true "Test request with email field"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-providers/{id}/test [post]
func (h *EmailProfileHandler) TestProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	if err := h.emailProfileService.TestProvider(c.Request.Context(), id, req.Email); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to test email provider", map[string]interface{}{
			"error":       err.Error(),
			"provider_id": id.String(),
			"email":       req.Email,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test email sent successfully"})
}

// ListProfiles handles listing all email profiles
// @Summary List email profiles
// @Description Get a list of all configured email profiles
// @Tags Email Profiles
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.EmailProfile
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles [get]
func (h *EmailProfileHandler) ListProfiles(c *gin.Context) {
	appID, _ := utils.GetApplicationIDFromContext(c)
	profiles, err := h.emailProfileService.ListProfiles(c.Request.Context(), appID)
	if err != nil {
		h.logger.Error("Failed to list email profiles", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, profiles)
}

// GetProfile handles retrieving a specific email profile
// @Summary Get email profile
// @Description Get details of a specific email profile by ID
// @Tags Email Profiles
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 200 {object} models.EmailProfile
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles/{id} [get]
func (h *EmailProfileHandler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	profile, err := h.emailProfileService.GetProfile(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to get email profile", map[string]interface{}{
			"error":      err.Error(),
			"profile_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, profile)
}

// CreateProfile handles creating a new email profile
// @Summary Create email profile
// @Description Create a new email profile configuration
// @Tags Email Profiles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateEmailProfileRequest true "Create profile request"
// @Success 201 {object} models.EmailProfile
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles [post]
func (h *EmailProfileHandler) CreateProfile(c *gin.Context) {
	var req models.CreateEmailProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	profile, err := h.emailProfileService.CreateProfile(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to create email profile", map[string]interface{}{
			"error": err.Error(),
			"name":  req.Name,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusCreated, profile)
}

// UpdateProfile handles updating an existing email profile
// @Summary Update email profile
// @Description Update an existing email profile configuration
// @Tags Email Profiles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Param request body models.UpdateEmailProfileRequest true "Update profile request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles/{id} [put]
func (h *EmailProfileHandler) UpdateProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	var req models.UpdateEmailProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	if err := h.emailProfileService.UpdateProfile(c.Request.Context(), id, &req); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to update email profile", map[string]interface{}{
			"error":      err.Error(),
			"profile_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email profile updated successfully"})
}

// DeleteProfile handles deleting an email profile
// @Summary Delete email profile
// @Description Delete an email profile configuration
// @Tags Email Profiles
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles/{id} [delete]
func (h *EmailProfileHandler) DeleteProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	if err := h.emailProfileService.DeleteProfile(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to delete email profile", map[string]interface{}{
			"error":      err.Error(),
			"profile_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email profile deleted successfully"})
}

// SetDefaultProfile handles setting a profile as default
// @Summary Set default email profile
// @Description Set an email profile as the default profile for sending emails
// @Tags Email Profiles
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles/{id}/default [put]
func (h *EmailProfileHandler) SetDefaultProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	if err := h.emailProfileService.SetDefaultProfile(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to set default email profile", map[string]interface{}{
			"error":      err.Error(),
			"profile_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default email profile set successfully"})
}

// GetProfileTemplates handles retrieving templates assigned to a profile
// @Summary Get profile templates
// @Description Get all templates assigned to an email profile
// @Tags Email Profile Templates
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 200 {array} models.EmailProfileTemplate
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles/{id}/templates [get]
func (h *EmailProfileHandler) GetProfileTemplates(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	templates, err := h.emailProfileService.GetProfileTemplates(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to get profile templates", map[string]interface{}{
			"error":      err.Error(),
			"profile_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, templates)
}

// SetProfileTemplate handles assigning a template to a profile for a specific OTP type
// @Summary Set profile template
// @Description Assign a template to an email profile for a specific OTP type
// @Tags Email Profile Templates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Param otp_type path string true "OTP Type (verification, password_reset, 2fa, login, registration)"
// @Param request body map[string]string true "Request with template_id field"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles/{id}/templates/{otp_type} [put]
func (h *EmailProfileHandler) SetProfileTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	otpType := c.Param("otp_type")

	var req struct {
		TemplateID string `json:"template_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	if err := h.emailProfileService.SetProfileTemplate(c.Request.Context(), id, otpType, templateID); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to set profile template", map[string]interface{}{
			"error":       err.Error(),
			"profile_id":  id.String(),
			"otp_type":    otpType,
			"template_id": templateID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile template set successfully"})
}

// RemoveProfileTemplate handles removing a template assignment from a profile
// @Summary Remove profile template
// @Description Remove a template assignment from an email profile for a specific OTP type
// @Tags Email Profile Templates
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Param otp_type path string true "OTP Type (verification, password_reset, 2fa, login, registration)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles/{id}/templates/{otp_type} [delete]
func (h *EmailProfileHandler) RemoveProfileTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	otpType := c.Param("otp_type")

	if err := h.emailProfileService.RemoveProfileTemplate(c.Request.Context(), id, otpType); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to remove profile template", map[string]interface{}{
			"error":      err.Error(),
			"profile_id": id.String(),
			"otp_type":   otpType,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile template removed successfully"})
}

// SendEmail handles sending an email using a template
// @Summary Send email
// @Description Send an email using a specified template type
// @Tags Email
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.AdminSendEmailRequest true "Send email request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email/send [post]
func (h *EmailProfileHandler) SendEmail(c *gin.Context) {
	var req models.AdminSendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	// Use application ID from body, or fallback to context
	applicationID := req.ApplicationID
	if applicationID == nil {
		appID, _ := utils.GetApplicationIDFromContext(c)
		applicationID = appID
	}

	if err := h.emailProfileService.SendEmail(c.Request.Context(), req.ProfileID, applicationID, req.ToEmail, req.TemplateType, req.Variables); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to send email", map[string]interface{}{
			"error":         err.Error(),
			"template_type": req.TemplateType,
			"to_email":      req.ToEmail,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully"})
}

// GetProfileStats handles retrieving statistics for an email profile
// @Summary Get profile statistics
// @Description Get email sending statistics for a specific profile
// @Tags Email Profiles
// @Security BearerAuth
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 200 {object} models.EmailStatsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles/{id}/stats [get]
func (h *EmailProfileHandler) GetProfileStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	stats, err := h.emailProfileService.GetProfileStats(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to get profile stats", map[string]interface{}{
			"error":      err.Error(),
			"profile_id": id.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, stats)
}

// TestProfile handles testing an email profile by sending a test email
// @Summary Test email profile
// @Description Test an email profile by sending a test email
// @Tags Email Profiles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Param request body map[string]string true "Test request with email field"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/email-profiles/{id}/test [post]
func (h *EmailProfileHandler) TestProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid UUID format")))
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	if err := h.emailProfileService.TestProfile(c.Request.Context(), id, req.Email); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to test email profile", map[string]interface{}{
			"error":      err.Error(),
			"profile_id": id.String(),
			"email":      req.Email,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test email sent successfully"})
}
