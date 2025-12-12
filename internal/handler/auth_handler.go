package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
	otpService  *service.OTPService
	logger      *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, userService *service.UserService, otpService *service.OTPService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		otpService:  otpService,
		logger:      log,
	}
}

// SignUp handles user registration
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "User registration data"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/signup [post]
func (h *AuthHandler) SignUp(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	authResp, err := h.authService.SignUp(c.Request.Context(), &req, ip, userAgent)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	// Send verification email (non-blocking)
	go func() {
		otpReq := &models.SendOTPRequest{
			Email: req.Email,
			Type:  models.OTPTypeVerification,
		}
		if err := h.otpService.SendOTP(c.Request.Context(), otpReq); err != nil {
			h.logger.Error("Failed to send verification email", map[string]interface{}{
				"error": err.Error(),
				"email": req.Email,
			})
		}
	}()

	c.JSON(http.StatusCreated, authResp)
}

// SignIn handles user login
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.SignInRequest true "User login data"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/signin [post]
func (h *AuthHandler) SignIn(c *gin.Context) {
	var req models.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	authResp, err := h.authService.SignIn(c.Request.Context(), &req, ip, userAgent)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	authResp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, ip, userAgent)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// Logout handles user logout
// @Summary Logout user
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	// Extract token
	parts := len(authHeader)
	if parts < 7 {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrInvalidToken))
		return
	}
	token := authHeader[7:] // Remove "Bearer " prefix

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.authService.Logout(c.Request.Context(), token, ip, userAgent); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// GetProfile retrieves the authenticated user's profile
// @Summary Get user profile
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), *userID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the authenticated user's profile
// @Summary Update user profile
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.UpdateUserRequest true "Profile update data"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	user, err := h.userService.UpdateProfile(c.Request.Context(), *userID, &req, ip, userAgent)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// ChangePassword changes the authenticated user's password
// @Summary Change password
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.ChangePasswordRequest true "Password change data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.authService.ChangePassword(c.Request.Context(), *userID, req.OldPassword, req.NewPassword, ip, userAgent); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// ResendVerificationEmail resends the email verification code
// @Summary Resend verification email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Email"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 429 {object} models.ErrorResponse
// @Router /auth/verify/resend [post]
func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	otpReq := &models.SendOTPRequest{
		Email: req.Email,
		Type:  models.OTPTypeVerification,
	}

	if err := h.otpService.SendOTP(c.Request.Context(), otpReq); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to resend verification email", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification code sent to your email",
		"email":   req.Email,
	})
}

// VerifyEmail verifies the user's email with OTP code
// @Summary Verify email address
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Email and Code"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/verify/email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	verifyReq := &models.VerifyOTPRequest{
		Email: req.Email,
		Code:  req.Code,
		Type:  models.OTPTypeVerification,
	}

	response, err := h.otpService.VerifyOTP(c.Request.Context(), verifyReq)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to verify email", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
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
		"message": "Email verified successfully",
	})
}

// RequestPasswordReset sends a password reset OTP code
// @Summary Request password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Email"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 429 {object} models.ErrorResponse
// @Router /auth/password/reset/request [post]
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// Check if user exists
	email := utils.NormalizeEmail(req.Email)
	user, err := h.userService.GetByEmail(c.Request.Context(), email)
	if err != nil || user == nil {
		// Don't reveal if user exists or not for security
		c.JSON(http.StatusOK, gin.H{
			"message": "If an account with that email exists, a password reset code has been sent",
			"email":   req.Email,
		})
		return
	}

	otpReq := &models.SendOTPRequest{
		Email: email,
		Type:  models.OTPTypePasswordReset,
	}

	if err := h.otpService.SendOTP(c.Request.Context(), otpReq); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to send password reset email", map[string]interface{}{
			"error": err.Error(),
			"email": email,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset code sent to your email",
		"email":   req.Email,
	})
}

// ResetPassword resets the password using OTP verification
// @Summary Reset password with OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Password reset data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/password/reset/complete [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// Verify OTP
	verifyReq := &models.VerifyOTPRequest{
		Email: req.Email,
		Code:  req.Code,
		Type:  models.OTPTypePasswordReset,
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
			"message": "Invalid or expired reset code",
		})
		return
	}

	// Get user
	email := utils.NormalizeEmail(req.Email)
	user, err := h.userService.GetByEmail(c.Request.Context(), email)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrUserNotFound))
		return
	}

	// Validate new password
	if !utils.IsPasswordValid(req.NewPassword) {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Password must be at least 8 characters"),
		))
		return
	}

	// Update password
	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.authService.ResetPassword(c.Request.Context(), user.ID, req.NewPassword, ip, userAgent); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}
