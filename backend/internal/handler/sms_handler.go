package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// SMSHandler handles SMS-related requests
type SMSHandler struct {
	smsService *service.SMSService
	logger     *logger.Logger
}

// NewSMSHandler creates a new SMS handler
func NewSMSHandler(smsService *service.SMSService, logger *logger.Logger) *SMSHandler {
	return &SMSHandler{
		smsService: smsService,
		logger:     logger,
	}
}

// SendSMS handles sending an SMS OTP code
// @Summary Send SMS OTP code
// @Description Send an OTP code via SMS for verification, password reset, or login
// @Tags SMS
// @Accept json
// @Produce json
// @Param request body models.SendSMSRequest true "Send SMS request"
// @Success 200 {object} models.SendSMSResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 429 {object} models.ErrorResponse
// @Router /sms/send [post]
func (h *SMSHandler) SendSMS(c *gin.Context) {
	var req models.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	// Get client IP address
	ipAddress := utils.GetClientIP(c)

	response, err := h.smsService.SendOTP(c.Request.Context(), &req, ipAddress)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// VerifySMS handles verifying an SMS OTP code
// @Summary Verify SMS OTP code
// @Description Verify an OTP code sent via SMS for verification, password reset, or login
// @Tags SMS
// @Accept json
// @Produce json
// @Param request body models.VerifySMSOTPRequest true "Verify SMS OTP request"
// @Success 200 {object} models.VerifySMSOTPResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /sms/verify [post]
func (h *SMSHandler) VerifySMS(c *gin.Context) {
	var req models.VerifySMSOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	response, err := h.smsService.VerifyOTP(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStats handles retrieving SMS statistics (admin only)
// @Summary Get SMS statistics
// @Description Get statistics about sent SMS messages
// @Tags SMS
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.SMSStatsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /sms/stats [get]
func (h *SMSHandler) GetStats(c *gin.Context) {
	stats, err := h.smsService.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get SMS stats", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, stats)
}
