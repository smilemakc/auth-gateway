package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// AdminHandler handles admin operations
type AdminHandler struct {
	adminService *service.AdminService
	userService  *service.UserService
	otpService   *service.OTPService
	auditService *service.AuditService
	logger       *logger.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService *service.AdminService, userService *service.UserService, otpService *service.OTPService, auditService *service.AuditService, logger *logger.Logger) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		userService:  userService,
		otpService:   otpService,
		auditService: auditService,
		logger:       logger,
	}
}

// GetStats returns system statistics
// @Summary Get system statistics
// @Description Get system-wide statistics including user counts, sessions, and API keys (admin only)
// @Tags Admin - Dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.AdminStatsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/stats [get]
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get stats", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ListUsers returns paginated list of users
// @Summary List all users
// @Description Get paginated list of all users (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.AdminUserListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	appID, _ := utils.GetApplicationIDFromContext(c)

	response, err := h.adminService.ListUsers(c.Request.Context(), appID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list users", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUser returns detailed user information
// @Summary Get user details
// @Description Get detailed information about a specific user (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} models.AdminUserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id} [get]
func (h *AdminHandler) GetUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	user, err := h.adminService.GetUser(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserOAuthAccounts returns OAuth accounts linked to a user
// @Summary Get user OAuth accounts
// @Description Get OAuth accounts linked to a specific user (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {array} models.OAuthAccount
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/oauth-accounts [get]
func (h *AdminHandler) GetUserOAuthAccounts(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	accounts, err := h.adminService.GetUserOAuthAccounts(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// UpdateUser updates user information
// @Summary Update user
// @Description Update user information (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param request body models.AdminUpdateUserRequest true "User update data"
// @Success 200 {object} models.AdminUserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id} [put]
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	var req models.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// Get admin ID from context
	adminID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	user, err := h.adminService.UpdateUser(c.Request.Context(), userID, &req, *adminID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser creates a new user
// @Summary Create user
// @Description Create a new user (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.AdminCreateUserRequest true "User creation data"
// @Success 201 {object} models.AdminUserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users [post]
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req models.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// Get admin ID from context
	adminID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	user, err := h.adminService.CreateUser(c.Request.Context(), &req, *adminID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user
// @Summary Delete user
// @Description Soft delete a user (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	if err := h.adminService.DeleteUser(c.Request.Context(), userID); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListAPIKeys returns all API keys
// @Summary List all API keys
// @Description Get list of all API keys (admin only)
// @Tags Admin - API Keys
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Success 200 {array} models.AdminAPIKeyResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/api-keys [get]
func (h *AdminHandler) ListAPIKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	appID, _ := utils.GetApplicationIDFromContext(c)

	result, err := h.adminService.ListAPIKeys(c.Request.Context(), appID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list API keys", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, result)
}

// RevokeAPIKey revokes an API key
// @Summary Revoke API key
// @Description Revoke an API key (admin only)
// @Tags Admin - API Keys
// @Security BearerAuth
// @Produce json
// @Param id path string true "API Key ID (UUID)"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/api-keys/{id}/revoke [post]
func (h *AdminHandler) RevokeAPIKey(c *gin.Context) {
	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid API key ID"),
		))
		return
	}

	if err := h.adminService.RevokeAPIKey(c.Request.Context(), keyID); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}

// ListAuditLogs returns audit logs
// @Summary List audit logs
// @Description Get paginated audit logs (admin only)
// @Tags Admin - Audit Logs
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Param user_id query string false "Filter by user ID (UUID)"
// @Success 200 {array} models.AdminAuditLogResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/audit-logs [get]
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
			))
			return
		}
		userID = &id
	}

	appID, _ := utils.GetApplicationIDFromContext(c)
	if appID != nil {
		offset := (page - 1) * pageSize
		logs, total, err := h.auditService.ListByApp(c.Request.Context(), *appID, pageSize, offset)
		if err != nil {
			h.logger.Error("Failed to list audit logs", map[string]interface{}{
				"error": err.Error(),
			})
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"logs":      logs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
		return
	}

	logs, err := h.adminService.ListAuditLogs(c.Request.Context(), page, pageSize, userID)
	if err != nil {
		h.logger.Error("Failed to list audit logs", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, logs)
}

// AssignRole assigns a role to a user
// @Summary Assign role to user
// @Description Assign a role to a user (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param request body models.AssignRoleRequest true "Role assignment data"
// @Success 200 {object} models.AdminUserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/roles [post]
func (h *AdminHandler) AssignRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	var req models.AssignRoleRequest
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

	user, err := h.adminService.AssignRole(c.Request.Context(), userID, req.RoleID, *adminID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, user)
}

// RemoveRole removes a role from a user
// @Summary Remove role from user
// @Description Remove a role from a user (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param roleId path string true "Role ID (UUID)"
// @Success 200 {object} models.AdminUserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/roles/{roleId} [delete]
func (h *AdminHandler) RemoveRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid role ID"),
		))
		return
	}

	user, err := h.adminService.RemoveRole(c.Request.Context(), userID, roleID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, user)
}

// SendPasswordReset sends a password reset email to a user
// @Summary Send password reset email
// @Description Initiates password reset for a user by sending OTP to their email (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} object{message=string,email=string} "Password reset email sent"
// @Failure 400 {object} models.ErrorResponse "User email is not verified"
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/send-password-reset [post]
func (h *AdminHandler) SendPasswordReset(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	if user.Email == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "User does not have an email address"),
		))
		return
	}

	if !user.EmailVerified {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "User email is not verified"),
		))
		return
	}

	email := utils.NormalizeEmail(user.Email)
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
			"error":   err.Error(),
			"user_id": userID,
			"email":   email,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	adminID, exists := utils.GetUserIDFromContext(c)
	if exists {
		h.auditService.Log(service.AuditLogParams{
			UserID:    adminID,
			Action:    models.ActionAdminPasswordResetInitiate,
			Status:    models.StatusSuccess,
			IP:        c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Details: map[string]interface{}{
				"target_user_id": userID,
				"email":          email,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email has been sent",
		"email":   user.Email,
	})
}

// Reset2FA administratively disables 2FA for a user
// @Summary Reset user 2FA
// @Description Administratively disable 2FA for a user who lost access to their authenticator (admin only)
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} object{message=string,user_id=string}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/reset-2fa [post]
func (h *AdminHandler) Reset2FA(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	adminID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrUnauthorized))
		return
	}

	if err := h.adminService.AdminReset2FA(c.Request.Context(), userID, *adminID); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA has been disabled for user",
		"user_id": userID.String(),
	})
}

// SyncUsers returns users updated after a timestamp for periodic sync
// @Summary Sync users
// @Description Get users updated after a given timestamp (requires users:sync scope)
// @Tags Admin - Users
// @Security ApiKeyAuth
// @Produce json
// @Param updated_after query string true "RFC3339 timestamp" example("2024-01-15T10:30:00Z")
// @Param application_id query string false "Application ID filter"
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} models.SyncUsersResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/sync [get]
func (h *AdminHandler) SyncUsers(c *gin.Context) {
	updatedAfterStr := c.Query("updated_after")
	if updatedAfterStr == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "updated_after parameter is required"),
		))
		return
	}

	updatedAfter, err := time.Parse(time.RFC3339, updatedAfterStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid updated_after format, expected RFC3339"),
		))
		return
	}

	var appID *uuid.UUID
	if appIDStr := c.Query("application_id"); appIDStr != "" {
		parsed, err := uuid.Parse(appIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Invalid application_id format"),
			))
			return
		}
		appID = &parsed
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response, err := h.adminService.SyncUsers(c.Request.Context(), updatedAfter, appID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, response)
}

// ImportUsers bulk imports users
// @Summary Import users
// @Description Bulk import users with optional UUID preservation (requires users:import scope)
// @Tags Admin - Users
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param X-Application-ID header string false "Application ID for auto-creating profiles"
// @Param request body models.BulkImportUsersRequest true "Import data"
// @Success 200 {object} models.ImportUsersResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/import [post]
func (h *AdminHandler) ImportUsers(c *gin.Context) {
	var req models.BulkImportUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	appID, _ := utils.GetApplicationIDFromContext(c)

	response, err := h.adminService.ImportUsers(c.Request.Context(), &req, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, response)
}
