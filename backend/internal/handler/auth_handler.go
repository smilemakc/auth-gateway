package handler

import (
	"net/http"

	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService         *service.AuthService
	userService         *service.UserService
	otpService          *service.OTPService
	emailProfileService *service.EmailProfileService
	logger              *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, userService *service.UserService, otpService *service.OTPService, emailProfileService *service.EmailProfileService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService:         authService,
		userService:         userService,
		otpService:          otpService,
		emailProfileService: emailProfileService,
		logger:              log,
	}
}

// SignUp handles user registration
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "User registration data"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/signup [post]
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
	deviceInfo := utils.GetDeviceInfoFromContext(c)
	appID, _ := utils.GetApplicationIDFromContext(c)

	authResp, err := h.authService.SignUp(c.Request.Context(), &req, ip, userAgent, deviceInfo, appID)
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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		otpReq := &models.SendOTPRequest{
			Email:         &req.Email,
			Type:          models.OTPTypeVerification,
			ApplicationID: appID,
		}
		if err := h.otpService.SendOTP(ctx, otpReq); err != nil {
			h.logger.Error("Failed to send verification email", map[string]interface{}{
				"error": err.Error(),
				"email": req.Email,
			})
		}
	}()

	// Send welcome email (non-blocking)
	if h.emailProfileService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			variables := map[string]interface{}{
				"username":  req.Username,
				"email":     req.Email,
				"full_name": req.FullName,
			}
			if err := h.emailProfileService.SendEmail(ctx, nil, appID, req.Email, models.EmailTemplateTypeWelcome, variables); err != nil {
				h.logger.Error("Failed to send welcome email", map[string]interface{}{
					"error": err.Error(),
					"email": req.Email,
				})
			}
		}()
	}

	c.JSON(http.StatusCreated, authResp)
}

// SignIn handles user login
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.SignInRequest true "User login data"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/signin [post]
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
	deviceInfo := utils.GetDeviceInfoFromContext(c)
	appID, _ := utils.GetApplicationIDFromContext(c)

	authResp, err := h.authService.SignIn(c.Request.Context(), &req, ip, userAgent, deviceInfo, appID)
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
// @Description Generate new access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/refresh [post]
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
	deviceInfo := utils.GetDeviceInfoFromContext(c)

	authResp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, ip, userAgent, deviceInfo)
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
// @Description Invalidate the current access token
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.MessageResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Token already validated by middleware; fetch raw token from context
	token, ok := utils.GetTokenFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

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

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Successfully logged out"})
}

// GetProfile retrieves the authenticated user's profile
// @Summary Get user profile
// @Description Get the profile information of the authenticated user
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/profile [get]
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
// @Description Update the profile information of the authenticated user
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.UpdateUserRequest true "Profile update data"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/profile [put]
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
// @Description Change the password for the authenticated user
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.ChangePasswordRequest true "Password change data"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/change-password [post]
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

	// Send password_changed notification (non-blocking)
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
				"username":   user.Username,
				"email":      user.Email,
				"ip_address": ip,
				"user_agent": userAgent,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}
			if err := h.emailProfileService.SendEmail(ctx, nil, appID, user.Email, models.EmailTemplateTypePasswordChanged, variables); err != nil {
				h.logger.Error("Failed to send password changed notification", map[string]interface{}{
					"error":   err.Error(),
					"user_id": userID.String(),
				})
			}
		}()
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Password changed successfully"})
}

// RequestPasswordReset sends a password reset OTP code
// @Summary Request password reset
// @Description Send a password reset code to the user's email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.ForgotPasswordRequest true "Email address"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 429 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/password/reset/request [post]
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

	appID, _ := utils.GetApplicationIDFromContext(c)
	otpReq := &models.SendOTPRequest{
		Email:         &email,
		Type:          models.OTPTypePasswordReset,
		ApplicationID: appID,
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
// @Description Complete password reset using OTP verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Password reset data with OTP code"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/password/reset/complete [post]
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
		Email: &req.Email,
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

	// Send password_changed notification (non-blocking)
	if h.emailProfileService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			appID, _ := utils.GetApplicationIDFromContext(c)
			variables := map[string]interface{}{
				"username":   user.Username,
				"email":      user.Email,
				"ip_address": ip,
				"user_agent": userAgent,
				"timestamp":  time.Now().UTC().Format(time.RFC3339),
			}
			if err := h.emailProfileService.SendEmail(ctx, nil, appID, user.Email, models.EmailTemplateTypePasswordChanged, variables); err != nil {
				h.logger.Error("Failed to send password changed notification", map[string]interface{}{
					"error":   err.Error(),
					"user_id": user.ID.String(),
				})
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// Verify2FA verifies 2FA code during login
// @Summary Verify 2FA during login
// @Description Complete two-factor authentication during login using TOTP code
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.TwoFactorLoginVerifyRequest true "2FA token and TOTP code"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/2fa/login/verify [post]
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req models.TwoFactorLoginVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)
	deviceInfo := utils.GetDeviceInfoFromContext(c)

	authResp, err := h.authService.Verify2FALogin(c.Request.Context(), req.TwoFactorToken, req.Code, ip, userAgent, deviceInfo)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// InitPasswordlessRegistration initiates passwordless registration
// @Summary Initiate passwordless registration
// @Description Start two-step registration without password. Sends OTP to email or phone.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.InitPasswordlessRegistrationRequest true "Registration data (email or phone)"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/signup/phone [post]
func (h *AuthHandler) InitPasswordlessRegistration(c *gin.Context) {
	var req models.InitPasswordlessRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.authService.InitPasswordlessRegistration(c.Request.Context(), &req, ip, userAgent); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	// Send OTP via email or SMS
	if req.Email != nil && *req.Email != "" {
		appID, _ := utils.GetApplicationIDFromContext(c)
		otpReq := &models.SendOTPRequest{
			Email:         req.Email,
			Type:          models.OTPTypeRegistration,
			ApplicationID: appID,
		}
		if err := h.otpService.SendOTP(c.Request.Context(), otpReq); err != nil {
			h.logger.Error("Failed to send registration OTP email", map[string]interface{}{
				"error": err.Error(),
				"email": *req.Email,
			})
			// Don't fail - registration is initiated, OTP might be resent
		}
	}

	// Determine message based on delivery method
	message := "Registration initiated. Please check your email for the verification code."
	if req.Phone != nil && *req.Phone != "" && (req.Email == nil || *req.Email == "") {
		message = "Registration initiated. Please check your phone for the verification code."
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// CompletePasswordlessRegistration completes passwordless registration after OTP verification
// @Summary Complete passwordless registration
// @Description Complete registration by verifying OTP code. Creates user account and returns tokens.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.CompletePasswordlessRegistrationRequest true "OTP verification"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/signup/phone/verify [post]
func (h *AuthHandler) CompletePasswordlessRegistration(c *gin.Context) {
	var req models.CompletePasswordlessRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// First verify the OTP
	verifyReq := &models.VerifyOTPRequest{
		Email: req.Email,
		Phone: req.Phone,
		Code:  req.Code,
		Type:  models.OTPTypeRegistration,
	}

	response, err := h.otpService.VerifyOTP(c.Request.Context(), verifyReq)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	if !response.Valid {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.NewAppError(http.StatusUnauthorized, "Invalid or expired verification code"),
		))
		return
	}

	// OTP verified, complete registration
	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)
	deviceInfo := utils.GetDeviceInfoFromContext(c)

	authResp, err := h.authService.CompletePasswordlessRegistration(c.Request.Context(), &req, ip, userAgent, deviceInfo)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	// Send welcome email (non-blocking)
	if h.emailProfileService != nil && authResp.User != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			appID, _ := utils.GetApplicationIDFromContext(c)
			variables := map[string]interface{}{
				"username":  authResp.User.Username,
				"email":     authResp.User.Email,
				"full_name": authResp.User.FullName,
			}
			if err := h.emailProfileService.SendEmail(ctx, nil, appID, authResp.User.Email, models.EmailTemplateTypeWelcome, variables); err != nil {
				h.logger.Error("Failed to send welcome email", map[string]interface{}{
					"error": err.Error(),
					"email": authResp.User.Email,
				})
			}
		}()
	}

	c.JSON(http.StatusCreated, authResp)
}
