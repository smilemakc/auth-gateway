package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/jwt"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// OTPHandler handles OTP-related requests
type OTPHandler struct {
	otpService  *service.OTPService
	authService *service.AuthService
	jwtService  *jwt.Service
	logger      *logger.Logger
}

// NewOTPHandler creates a new OTP handler
func NewOTPHandler(
	otpService *service.OTPService,
	authService *service.AuthService,
	jwtService *jwt.Service,
	logger *logger.Logger,
) *OTPHandler {
	return &OTPHandler{
		otpService:  otpService,
		authService: authService,
		jwtService:  jwtService,
		logger:      logger,
	}
}

// SendOTP handles sending an OTP code
// @Summary Send OTP code
// @Description Send an OTP code to email for verification, password reset, or login
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body models.SendOTPRequest true "Send OTP request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 429 {object} models.ErrorResponse
// @Router /otp/send [post]
func (h *OTPHandler) SendOTP(c *gin.Context) {
	var req models.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	if err := h.otpService.SendOTP(c.Request.Context(), &req); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to send OTP", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP code sent successfully",
		"email":   req.Email,
	})
}

// VerifyOTP handles verifying an OTP code
// @Summary Verify OTP code
// @Description Verify an OTP code for email verification, password reset, or login
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body models.VerifyOTPRequest true "Verify OTP request"
// @Success 200 {object} models.VerifyOTPResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /otp/verify [post]
func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	response, err := h.otpService.VerifyOTP(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to verify OTP", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	if !response.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid":   false,
			"message": "Invalid or expired OTP code",
		})
		return
	}

	// For login type, generate JWT tokens
	if req.Type == models.OTPTypeLogin && response.User != nil {
		accessToken, err := h.jwtService.GenerateAccessToken(response.User)
		if err != nil {
			h.logger.Error("Failed to generate access token", map[string]interface{}{
				"error": err.Error(),
			})
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}

		refreshToken, err := h.jwtService.GenerateRefreshToken(response.User)
		if err != nil {
			h.logger.Error("Failed to generate refresh token", map[string]interface{}{
				"error": err.Error(),
			})
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}

		response.AccessToken = accessToken
		response.RefreshToken = refreshToken
	}

	c.JSON(http.StatusOK, response)
}

// RequestPasswordlessLogin handles passwordless login request
// @Summary Request passwordless login
// @Description Send OTP code for passwordless login
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body map[string]string true "Email"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/passwordless/request [post]
func (h *OTPHandler) RequestPasswordlessLogin(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	otpReq := &models.SendOTPRequest{
		Email: &req.Email,
		Type:  models.OTPTypeLogin,
	}

	if err := h.otpService.SendOTP(c.Request.Context(), otpReq); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login code sent to your email",
		"email":   req.Email,
	})
}

// VerifyPasswordlessLogin handles passwordless login verification
// @Summary Verify passwordless login
// @Description Verify OTP code and login
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body map[string]string true "Email and Code"
// @Success 200 {object} models.VerifyOTPResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/passwordless/verify [post]
func (h *OTPHandler) VerifyPasswordlessLogin(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	verifyReq := &models.VerifyOTPRequest{
		Email: &req.Email,
		Code:  req.Code,
		Type:  models.OTPTypeLogin,
	}

	response, err := h.otpService.VerifyOTP(c.Request.Context(), verifyReq)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	if !response.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid":   false,
			"message": "Invalid or expired login code",
		})
		return
	}

	// Generate tokens
	if response.User != nil {
		accessToken, err := h.jwtService.GenerateAccessToken(response.User)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}

		refreshToken, err := h.jwtService.GenerateRefreshToken(response.User)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user":          response.User,
		})
		return
	}

	c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrInvalidCredentials))
}
