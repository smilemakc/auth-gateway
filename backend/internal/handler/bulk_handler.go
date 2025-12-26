package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// BulkHandler handles bulk operations HTTP requests
type BulkHandler struct {
	bulkService *service.BulkService
	logger      *logger.Logger
}

// NewBulkHandler creates a new bulk handler
func NewBulkHandler(bulkService *service.BulkService, logger *logger.Logger) *BulkHandler {
	return &BulkHandler{
		bulkService: bulkService,
		logger:      logger,
	}
}

// BulkCreateUsers handles bulk user creation
// @Summary Bulk create users
// @Description Create multiple users in a single operation
// @Tags Admin - Bulk Operations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.BulkCreateUsersRequest true "Users to create"
// @Success 200 {object} models.BulkOperationResult
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/bulk-create [post]
func (h *BulkHandler) BulkCreateUsers(c *gin.Context) {
	var req models.BulkCreateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// Limit batch size
	if len(req.Users) > 100 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Batch size cannot exceed 100 users"),
		))
		return
	}

	result, err := h.bulkService.BulkCreateUsers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

// BulkUpdateUsers handles bulk user updates
// @Summary Bulk update users
// @Description Update multiple users in a single operation
// @Tags Admin - Bulk Operations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.BulkUpdateUsersRequest true "Users to update"
// @Success 200 {object} models.BulkOperationResult
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/bulk-update [put]
func (h *BulkHandler) BulkUpdateUsers(c *gin.Context) {
	var req models.BulkUpdateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// Limit batch size
	if len(req.Users) > 100 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Batch size cannot exceed 100 users"),
		))
		return
	}

	result, err := h.bulkService.BulkUpdateUsers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

// BulkDeleteUsers handles bulk user deletion
// @Summary Bulk delete users
// @Description Delete multiple users in a single operation (soft delete)
// @Tags Admin - Bulk Operations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.BulkDeleteUsersRequest true "User IDs to delete"
// @Success 200 {object} models.BulkOperationResult
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/bulk-delete [post]
func (h *BulkHandler) BulkDeleteUsers(c *gin.Context) {
	var req models.BulkDeleteUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// Limit batch size
	if len(req.UserIDs) > 100 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Batch size cannot exceed 100 users"),
		))
		return
	}

	result, err := h.bulkService.BulkDeleteUsers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

// BulkAssignRoles handles bulk role assignment
// @Summary Bulk assign roles
// @Description Assign roles to multiple users in a single operation
// @Tags Admin - Bulk Operations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.BulkAssignRolesRequest true "User IDs and role IDs"
// @Success 200 {object} models.BulkOperationResult
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/bulk-assign-roles [post]
func (h *BulkHandler) BulkAssignRoles(c *gin.Context) {
	var req models.BulkAssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	// Limit batch size
	if len(req.UserIDs) > 100 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Batch size cannot exceed 100 users"),
		))
		return
	}

	// Get current user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		// Try alternative key
		userIDInterface, exists = c.Get("userID")
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.NewAppError(http.StatusUnauthorized, "User not authenticated"),
		))
		return
	}

	var assignedBy uuid.UUID
	switch v := userIDInterface.(type) {
	case uuid.UUID:
		assignedBy = v
	case string:
		var err error
		assignedBy, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.NewAppError(http.StatusInternalServerError, "Invalid user ID in context"),
			))
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.NewAppError(http.StatusInternalServerError, "Invalid user ID type in context"),
		))
		return
	}

	result, err := h.bulkService.BulkAssignRoles(c.Request.Context(), &req, assignedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}
