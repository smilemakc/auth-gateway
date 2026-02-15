package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// OTPHandler handles OTP-related requests
type OTPHandler struct {
	otpService  *service.OTPService
	authService *service.AuthService
	logger      *logger.Logger
}

// NewOTPHandler creates a new OTP handler
func NewOTPHandler(
	otpService *service.OTPService,
	authService *service.AuthService,
	logger *logger.Logger,
) *OTPHandler {
	return &OTPHandler{
		otpService:  otpService,
		authService: authService,
		logger:      logger,
	}
}

// ResendVerification handles sending verification OTP to email or phone
// @Summary Resend verification code
// @Description Resend the verification code to email or phone
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body models.SendOTPRequest true "Send OTP request (type must be verification)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 429 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/verify/resend [post]
func (h *OTPHandler) ResendVerification(c *gin.Context) {
	var req models.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	req.Type = models.OTPTypeVerification

	if req.ApplicationID == nil {
		appID, _ := utils.GetApplicationIDFromContext(c)
		req.ApplicationID = appID
	}

	if err := h.otpService.SendOTP(c.Request.Context(), &req); err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification code sent",
		"email":   req.Email,
		"phone":   req.Phone,
	})
}

// VerifyEmailOTP verifies email/phone verification code
// @Summary Verify email or phone
// @Description Verify a verification code for email or phone
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body models.VerifyOTPRequest true "Verify OTP request (type must be verification)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/verify/email [post]
func (h *OTPHandler) VerifyEmailOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	req.Type = models.OTPTypeVerification

	response, err := h.otpService.VerifyOTP(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	if !response.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid":   false,
			"message": "Invalid or expired verification code",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Verification successful",
		"user":    response.User,
	})
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
// @Failure 500 {object} models.ErrorResponse
// @Router /api/otp/send [post]
func (h *OTPHandler) SendOTP(c *gin.Context) {
	var req models.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	if req.ApplicationID == nil {
		appID, _ := utils.GetApplicationIDFromContext(c)
		req.ApplicationID = appID
	}

	if err := h.otpService.SendOTP(c.Request.Context(), &req); err != nil {
		utils.RespondWithError(c, err)
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
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/otp/verify [post]
func (h *OTPHandler) VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	response, err := h.otpService.VerifyOTP(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	if !response.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid":   false,
			"message": "Invalid or expired OTP code",
		})
		return
	}

	// For login type, generate JWT tokens with proper session tracking
	if req.Type == models.OTPTypeLogin && response.User != nil {
		authResp, err := h.authService.GenerateTokensForUser(
			c.Request.Context(),
			response.User,
			utils.GetClientIP(c),
			c.Request.UserAgent(),
		)
		if err != nil {
			h.logger.Error("Failed to generate tokens for OTP login", map[string]interface{}{
				"error":   err.Error(),
				"user_id": response.User.ID,
			})
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}

		response.AccessToken = authResp.AccessToken
		response.RefreshToken = authResp.RefreshToken
	}

	c.JSON(http.StatusOK, response)
}

// RequestPasswordlessLogin handles passwordless login request
// @Summary Request passwordless login
// @Description Send OTP code for passwordless login
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.PasswordlessLoginRequest true "Email for passwordless login"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/passwordless/request [post]
func (h *OTPHandler) RequestPasswordlessLogin(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err))
		return
	}

	appID, _ := utils.GetApplicationIDFromContext(c)
	otpReq := &models.SendOTPRequest{
		Email:         &req.Email,
		Type:          models.OTPTypeLogin,
		ApplicationID: appID,
	}

	if err := h.otpService.SendOTP(c.Request.Context(), otpReq); err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login code sent to your email",
		"email":   req.Email,
	})
}

// VerifyPasswordlessLogin handles passwordless login verification
// @Summary Verify passwordless login
// @Description Verify OTP code and complete passwordless login, returns access and refresh tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.PasswordlessLoginVerifyRequest true "Email and OTP code"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/passwordless/verify [post]
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
		utils.RespondWithError(c, err)
		return
	}

	if !response.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid":   false,
			"message": "Invalid or expired login code",
		})
		return
	}

	// Generate tokens with proper session tracking
	if response.User != nil {
		authResp, err := h.authService.GenerateTokensForUser(
			c.Request.Context(),
			response.User,
			utils.GetClientIP(c),
			c.Request.UserAgent(),
		)
		if err != nil {
			h.logger.Error("Failed to generate tokens for passwordless login", map[string]interface{}{
				"error":   err.Error(),
				"user_id": response.User.ID,
			})
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  authResp.AccessToken,
			"refresh_token": authResp.RefreshToken,
			"user":          response.User,
		})
		return
	}

	c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrInvalidCredentials))
}
