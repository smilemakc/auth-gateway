package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// TwoFactorHandler handles 2FA-related requests
type TwoFactorHandler struct {
	twoFAService        *service.TwoFactorService
	userService         *service.UserService
	emailProfileService *service.EmailProfileService
	logger              *logger.Logger
}

// NewTwoFactorHandler creates a new 2FA handler
func NewTwoFactorHandler(
	twoFAService *service.TwoFactorService,
	userService *service.UserService,
	emailProfileService *service.EmailProfileService,
	logger *logger.Logger,
) *TwoFactorHandler {
	return &TwoFactorHandler{
		twoFAService:        twoFAService,
		userService:         userService,
		emailProfileService: emailProfileService,
		logger:              logger,
	}
}

// Setup initiates 2FA setup
// @Summary Setup 2FA
// @Description Generate TOTP secret and backup codes for two-factor authentication
// @Tags 2FA
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.TwoFactorSetupRequest true "Password for verification"
// @Success 200 {object} models.TwoFactorSetupResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/2fa/setup [post]
func (h *TwoFactorHandler) Setup(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	var req models.TwoFactorSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// Setup 2FA (password verification happens in service)
	response, err := h.twoFAService.SetupTOTP(c.Request.Context(), *userID, req.Password)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to setup 2FA", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, response)
}

// Verify verifies 2FA setup with initial code
// @Summary Verify 2FA setup
// @Description Verify initial TOTP code and enable 2FA for the user account
// @Tags 2FA
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.TwoFactorVerifyRequest true "TOTP code"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/2fa/verify [post]
func (h *TwoFactorHandler) Verify(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	var req models.TwoFactorVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	if err := h.twoFAService.VerifyTOTPSetup(c.Request.Context(), *userID, req.Code); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	// Send 2fa_enabled notification (non-blocking)
	if h.emailProfileService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			user, err := h.userService.GetProfile(ctx, *userID)
			if err != nil {
				return
			}
			appID, _ := utils.GetApplicationIDFromContext(c)
			variables := map[string]interface{}{
				"username":  user.Username,
				"email":     user.Email,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
			if err := h.emailProfileService.SendEmail(ctx, nil, appID, user.Email, models.EmailTemplateType2FAEnabled, variables); err != nil {
				h.logger.Error("Failed to send 2FA enabled notification", map[string]interface{}{
					"error":   err.Error(),
					"user_id": userID.String(),
				})
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA enabled successfully",
	})
}

// Disable disables 2FA
// @Summary Disable 2FA
// @Description Disable TOTP two-factor authentication for the user account
// @Tags 2FA
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.TwoFactorDisableRequest true "Password and TOTP code for verification"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/2fa/disable [post]
func (h *TwoFactorHandler) Disable(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	var req models.TwoFactorDisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	if err := h.twoFAService.DisableTOTP(c.Request.Context(), *userID, req.Password, req.Code); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	// Send 2fa_disabled notification (non-blocking)
	if h.emailProfileService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			user, err := h.userService.GetProfile(ctx, *userID)
			if err != nil {
				return
			}
			appID, _ := utils.GetApplicationIDFromContext(c)
			variables := map[string]interface{}{
				"username":  user.Username,
				"email":     user.Email,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
			if err := h.emailProfileService.SendEmail(ctx, nil, appID, user.Email, models.EmailTemplateType2FADisabled, variables); err != nil {
				h.logger.Error("Failed to send 2FA disabled notification", map[string]interface{}{
					"error":   err.Error(),
					"user_id": userID.String(),
				})
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA disabled successfully",
	})
}

// GetStatus returns 2FA status
// @Summary Get 2FA status
// @Description Get current 2FA status for the authenticated user
// @Tags 2FA
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.TwoFactorStatusResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/2fa/status [get]
func (h *TwoFactorHandler) GetStatus(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	status, err := h.twoFAService.GetStatus(c.Request.Context(), *userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, status)
}

// RegenerateBackupCodes generates new backup codes
// @Summary Regenerate backup codes
// @Description Generate new backup codes for 2FA recovery (invalidates previous codes)
// @Tags 2FA
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.RegenerateBackupCodesRequest true "Password for verification"
// @Success 200 {object} models.BackupCodesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/2fa/backup-codes/regenerate [post]
func (h *TwoFactorHandler) RegenerateBackupCodes(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	codes, err := h.twoFAService.RegenerateBackupCodes(c.Request.Context(), *userID, req.Password)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"backup_codes": codes,
		"message":      "Backup codes regenerated successfully. Save them in a secure location.",
	})
}
