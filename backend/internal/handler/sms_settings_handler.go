package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// SMSSettingsHandler handles SMS settings management (admin only)
type SMSSettingsHandler struct {
	smsSettingsRepo *repository.SMSSettingsRepository
	logger          *logger.Logger
}

// NewSMSSettingsHandler creates a new SMS settings handler
func NewSMSSettingsHandler(smsSettingsRepo *repository.SMSSettingsRepository, logger *logger.Logger) *SMSSettingsHandler {
	return &SMSSettingsHandler{
		smsSettingsRepo: smsSettingsRepo,
		logger:          logger,
	}
}

// CreateSettings handles creating new SMS settings
// @Summary Create SMS settings
// @Description Create new SMS provider settings (admin only)
// @Tags SMS Settings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateSMSSettingsRequest true "Create SMS settings request"
// @Success 201 {object} models.SMSSettings
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/sms/settings [post]
func (h *SMSSettingsHandler) CreateSettings(c *gin.Context) {
	var req models.CreateSMSSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	// Set defaults if not provided
	maxPerHour := 10
	if req.MaxPerHour != nil {
		maxPerHour = *req.MaxPerHour
	}
	maxPerDay := 50
	if req.MaxPerDay != nil {
		maxPerDay = *req.MaxPerDay
	}
	maxPerNumber := 5
	if req.MaxPerNumber != nil {
		maxPerNumber = *req.MaxPerNumber
	}

	settings := &models.SMSSettings{
		ID:                 uuid.New(),
		Provider:           req.Provider,
		Enabled:            req.Enabled,
		AccountSID:         req.AccountSID,
		AuthToken:          req.AuthToken,
		FromNumber:         req.FromNumber,
		AWSRegion:          req.AWSRegion,
		AWSAccessKeyID:     req.AWSAccessKeyID,
		AWSSecretAccessKey: req.AWSSecretAccessKey,
		AWSSenderID:        req.AWSSenderID,
		MaxPerHour:         maxPerHour,
		MaxPerDay:          maxPerDay,
		MaxPerNumber:       maxPerNumber,
		CreatedBy:          &userID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.smsSettingsRepo.Create(c.Request.Context(), settings); err != nil {
		h.logger.Error("Failed to create SMS settings", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusCreated, settings)
}

// GetSettings handles retrieving SMS settings by ID
// @Summary Get SMS settings
// @Description Get SMS settings by ID (admin only)
// @Tags SMS Settings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Settings ID"
// @Success 200 {object} models.SMSSettings
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/sms/settings/{id} [get]
func (h *SMSSettingsHandler) GetSettings(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	settings, err := h.smsSettingsRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(err))
			return
		}
		h.logger.Error("Failed to get SMS settings", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetActiveSettings handles retrieving active SMS settings
// @Summary Get active SMS settings
// @Description Get the currently active SMS settings (admin only)
// @Tags SMS Settings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.SMSSettings
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/sms/settings/active [get]
func (h *SMSSettingsHandler) GetActiveSettings(c *gin.Context) {
	settings, err := h.smsSettingsRepo.GetActive(c.Request.Context())
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(http.StatusNotFound, models.MessageResponse{Message: "No active SMS settings found"})
			return
		}
		h.logger.Error("Failed to get active SMS settings", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetAllSettings handles retrieving all SMS settings
// @Summary Get all SMS settings
// @Description Get all SMS settings (admin only)
// @Tags SMS Settings
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.SMSSettings
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/sms/settings [get]
func (h *SMSSettingsHandler) GetAllSettings(c *gin.Context) {
	settings, err := h.smsSettingsRepo.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get all SMS settings", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings handles updating SMS settings
// @Summary Update SMS settings
// @Description Update SMS settings by ID (admin only)
// @Tags SMS Settings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Settings ID"
// @Param request body models.UpdateSMSSettingsRequest true "Update SMS settings request"
// @Success 200 {object} models.SMSSettings
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/sms/settings/{id} [put]
func (h *SMSSettingsHandler) UpdateSettings(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	var req models.UpdateSMSSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	// Get existing settings
	existing, err := h.smsSettingsRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == models.ErrNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	// Update fields if provided
	if req.Provider != nil {
		existing.Provider = *req.Provider
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if req.AccountSID != nil {
		existing.AccountSID = req.AccountSID
	}
	if req.AuthToken != nil {
		existing.AuthToken = req.AuthToken
	}
	if req.FromNumber != nil {
		existing.FromNumber = req.FromNumber
	}
	if req.AWSRegion != nil {
		existing.AWSRegion = req.AWSRegion
	}
	if req.AWSAccessKeyID != nil {
		existing.AWSAccessKeyID = req.AWSAccessKeyID
	}
	if req.AWSSecretAccessKey != nil {
		existing.AWSSecretAccessKey = req.AWSSecretAccessKey
	}
	if req.AWSSenderID != nil {
		existing.AWSSenderID = req.AWSSenderID
	}
	if req.MaxPerHour != nil {
		existing.MaxPerHour = *req.MaxPerHour
	}
	if req.MaxPerDay != nil {
		existing.MaxPerDay = *req.MaxPerDay
	}
	if req.MaxPerNumber != nil {
		existing.MaxPerNumber = *req.MaxPerNumber
	}

	if err := h.smsSettingsRepo.Update(c.Request.Context(), id, existing); err != nil {
		h.logger.Error("Failed to update SMS settings", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteSettings handles deleting SMS settings
// @Summary Delete SMS settings
// @Description Delete SMS settings by ID (admin only)
// @Tags SMS Settings
// @Security BearerAuth
// @Param id path string true "Settings ID"
// @Success 204
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/sms/settings/{id} [delete]
func (h *SMSSettingsHandler) DeleteSettings(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	if err := h.smsSettingsRepo.Delete(c.Request.Context(), id); err != nil {
		if err == models.ErrNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(err))
			return
		}
		h.logger.Error("Failed to delete SMS settings", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.Status(http.StatusNoContent)
}
