package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// GroupHandler handles group-related HTTP requests
type GroupHandler struct {
	groupService *service.GroupService
	logger       *logger.Logger
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(groupService *service.GroupService, logger *logger.Logger) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		logger:       logger,
	}
}

// CreateGroup handles group creation
// @Summary Create a new group
// @Description Create a new organizational group/department
// @Tags Admin - Groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateGroupRequest true "Group creation data"
// @Success 201 {object} models.GroupResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	group, err := h.groupService.CreateGroup(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	memberCount, _ := h.groupService.GetGroupMemberCount(c.Request.Context(), group.ID)
	response := models.GroupResponse{
		ID:            group.ID,
		Name:          group.Name,
		DisplayName:   group.DisplayName,
		Description:   group.Description,
		ParentGroupID: group.ParentGroupID,
		IsSystemGroup: group.IsSystemGroup,
		MemberCount:   memberCount,
		CreatedAt:     group.CreatedAt,
		UpdatedAt:     group.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetGroup handles retrieving a group by ID
// @Summary Get group details
// @Description Get detailed information about a specific group
// @Tags Admin - Groups
// @Security BearerAuth
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Success 200 {object} models.GroupResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/groups/{id} [get]
func (h *GroupHandler) GetGroup(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	group, err := h.groupService.GetGroup(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	memberCount, _ := h.groupService.GetGroupMemberCount(c.Request.Context(), group.ID)
	response := models.GroupResponse{
		ID:            group.ID,
		Name:          group.Name,
		DisplayName:   group.DisplayName,
		Description:   group.Description,
		ParentGroupID: group.ParentGroupID,
		IsSystemGroup: group.IsSystemGroup,
		MemberCount:   memberCount,
		CreatedAt:     group.CreatedAt,
		UpdatedAt:     group.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// ListGroups handles listing all groups
// @Summary List all groups
// @Description Get paginated list of all groups
// @Tags Admin - Groups
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.GroupListResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/groups [get]
func (h *GroupHandler) ListGroups(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)

	response, err := h.groupService.ListGroups(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateGroup handles updating a group
// @Summary Update group
// @Description Update group information
// @Tags Admin - Groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Param request body models.UpdateGroupRequest true "Group update data"
// @Success 200 {object} models.GroupResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/groups/{id} [put]
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	group, err := h.groupService.UpdateGroup(c.Request.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	memberCount, _ := h.groupService.GetGroupMemberCount(c.Request.Context(), group.ID)
	response := models.GroupResponse{
		ID:            group.ID,
		Name:          group.Name,
		DisplayName:   group.DisplayName,
		Description:   group.Description,
		ParentGroupID: group.ParentGroupID,
		IsSystemGroup: group.IsSystemGroup,
		MemberCount:   memberCount,
		CreatedAt:     group.CreatedAt,
		UpdatedAt:     group.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteGroup handles deleting a group
// @Summary Delete group
// @Description Delete a group (only if it has no members and no child groups)
// @Tags Admin - Groups
// @Security BearerAuth
// @Param id path string true "Group ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/groups/{id} [delete]
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.groupService.DeleteGroup(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// AddGroupMembers handles adding users to a group
// @Summary Add members to group
// @Description Add one or more users to a group
// @Tags Admin - Groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Param request body models.AddGroupMembersRequest true "User IDs to add"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/groups/{id}/members [post]
func (h *GroupHandler) AddGroupMembers(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.AddGroupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	if err := h.groupService.AddGroupMembers(c.Request.Context(), id, req.UserIDs); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Users added to group successfully",
	})
}

// RemoveGroupMember handles removing a user from a group
// @Summary Remove member from group
// @Description Remove a user from a group
// @Tags Admin - Groups
// @Security BearerAuth
// @Param id path string true "Group ID (UUID)"
// @Param user_id path string true "User ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/groups/{id}/members/{user_id} [delete]
func (h *GroupHandler) RemoveGroupMember(c *gin.Context) {
	groupID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	userID, ok := utils.ParseUUIDParam(c, "user_id")
	if !ok {
		return
	}

	if err := h.groupService.RemoveGroupMember(c.Request.Context(), groupID, userID); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetGroupMembers handles retrieving group members
// @Summary Get group members
// @Description Get paginated list of users in a group
// @Tags Admin - Groups
// @Security BearerAuth
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.AdminUserListResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/groups/{id}/members [get]
func (h *GroupHandler) GetGroupMembers(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	page, pageSize := utils.ParsePagination(c)

	users, total, err := h.groupService.GetGroupMembers(c.Request.Context(), id, page, pageSize)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	// Convert to AdminUserResponse format
	userResponses := make([]models.AdminUserResponse, len(users))
	for i, user := range users {
		userResponses[i] = models.AdminUserResponse{
			ID:            user.ID,
			Email:         user.Email,
			Username:      user.Username,
			FullName:      user.FullName,
			IsActive:      user.IsActive,
			EmailVerified: user.EmailVerified,
			TOTPEnabled:   user.TOTPEnabled,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		}
	}

	// Convert []models.AdminUserResponse to []*models.AdminUserResponse
	userResponsePtrs := make([]*models.AdminUserResponse, len(userResponses))
	for i := range userResponses {
		userResponsePtrs[i] = &userResponses[i]
	}

	c.JSON(http.StatusOK, models.AdminUserListResponse{
		Users:    userResponsePtrs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
