package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// LDAPHandler handles LDAP-related HTTP requests
type LDAPHandler struct {
	ldapService *service.LDAPService
	logger      *logger.Logger
}

// NewLDAPHandler creates a new LDAP handler
func NewLDAPHandler(ldapService *service.LDAPService, logger *logger.Logger) *LDAPHandler {
	return &LDAPHandler{
		ldapService: ldapService,
		logger:      logger,
	}
}

// CreateConfig handles LDAP configuration creation
// @Summary Create LDAP configuration
// @Description Create a new LDAP/Active Directory configuration
// @Tags Admin - LDAP
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateLDAPConfigRequest true "LDAP configuration data"
// @Success 201 {object} models.LDAPConfig
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ldap/config [post]
func (h *LDAPHandler) CreateConfig(c *gin.Context) {
	var req models.CreateLDAPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	config, err := h.ldapService.CreateConfig(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusCreated, config)
}

// GetConfig handles retrieving LDAP configuration
// @Summary Get LDAP configuration
// @Description Get LDAP configuration by ID
// @Tags Admin - LDAP
// @Security BearerAuth
// @Produce json
// @Param id path string true "Configuration ID (UUID)"
// @Success 200 {object} models.LDAPConfig
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ldap/config/{id} [get]
func (h *LDAPHandler) GetConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid configuration ID"),
		))
		return
	}

	config, err := h.ldapService.GetConfig(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, config)
}

// GetActiveConfig handles retrieving active LDAP configuration
// @Summary Get active LDAP configuration
// @Description Get the currently active LDAP configuration
// @Tags Admin - LDAP
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.LDAPConfig
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ldap/config [get]
func (h *LDAPHandler) GetActiveConfig(c *gin.Context) {
	config, err := h.ldapService.GetActiveConfig(c.Request.Context())
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, config)
}

// ListConfigs handles listing all LDAP configurations
// @Summary List LDAP configurations
// @Description Get list of all LDAP configurations
// @Tags Admin - LDAP
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.LDAPConfigListResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ldap/configs [get]
func (h *LDAPHandler) ListConfigs(c *gin.Context) {
	configs, err := h.ldapService.ListConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, models.LDAPConfigListResponse{
		Configs: configs,
		Total:   len(configs),
	})
}

// UpdateConfig handles updating LDAP configuration
// @Summary Update LDAP configuration
// @Description Update an existing LDAP configuration
// @Tags Admin - LDAP
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Configuration ID (UUID)"
// @Param request body models.UpdateLDAPConfigRequest true "LDAP configuration update data"
// @Success 200 {object} models.LDAPConfig
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ldap/config/{id} [put]
func (h *LDAPHandler) UpdateConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid configuration ID"),
		))
		return
	}

	var req models.UpdateLDAPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	config, err := h.ldapService.UpdateConfig(c.Request.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, config)
}

// DeleteConfig handles deleting LDAP configuration
// @Summary Delete LDAP configuration
// @Description Delete an LDAP configuration
// @Tags Admin - LDAP
// @Security BearerAuth
// @Param id path string true "Configuration ID (UUID)"
// @Success 204 "No Content"
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ldap/config/{id} [delete]
func (h *LDAPHandler) DeleteConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid configuration ID"),
		))
		return
	}

	if err := h.ldapService.DeleteConfig(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// TestConnection handles testing LDAP connection
// @Summary Test LDAP connection
// @Description Test connection to LDAP server with provided credentials
// @Tags Admin - LDAP
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.LDAPTestConnectionRequest true "LDAP connection test data"
// @Success 200 {object} models.LDAPTestConnectionResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ldap/test-connection [post]
func (h *LDAPHandler) TestConnection(c *gin.Context) {
	var req models.LDAPTestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	response, err := h.ldapService.TestConnection(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, response)
}

// Sync handles triggering LDAP synchronization
// @Summary Trigger LDAP synchronization
// @Description Manually trigger synchronization of users and groups from LDAP
// @Tags Admin - LDAP
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Configuration ID (UUID)"
// @Param request body models.LDAPSyncRequest true "Sync options"
// @Success 200 {object} models.LDAPSyncResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ldap/config/{id}/sync [post]
func (h *LDAPHandler) Sync(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid configuration ID"),
		))
		return
	}

	var req models.LDAPSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	response, err := h.ldapService.Sync(c.Request.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSyncLogs handles retrieving LDAP sync logs
// @Summary Get LDAP sync logs
// @Description Get paginated list of LDAP synchronization logs
// @Tags Admin - LDAP
// @Security BearerAuth
// @Produce json
// @Param id path string true "Configuration ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.LDAPSyncLogListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ldap/config/{id}/sync-logs [get]
func (h *LDAPHandler) GetSyncLogs(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid configuration ID"),
		))
		return
	}

	page, pageSize := utils.ParsePagination(c)

	logs, total, err := h.ldapService.GetSyncLogs(c.Request.Context(), id, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	totalPages := (total + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, models.LDAPSyncLogListResponse{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}
