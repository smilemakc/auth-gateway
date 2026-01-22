package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdvancedAdminHandler handles advanced admin endpoints
type AdvancedAdminHandler struct {
	rbacService     *service.RBACService
	sessionService  *service.SessionService
	ipFilterService *service.IPFilterService
	brandingRepo    *repository.BrandingRepository
	systemRepo      *repository.SystemRepository
	geoRepo         *repository.GeoRepository
	log             *logger.Logger
}

// NewAdvancedAdminHandler creates a new advanced admin handler
func NewAdvancedAdminHandler(
	rbacService *service.RBACService,
	sessionService *service.SessionService,
	ipFilterService *service.IPFilterService,
	brandingRepo *repository.BrandingRepository,
	systemRepo *repository.SystemRepository,
	geoRepo *repository.GeoRepository,
	log *logger.Logger,
) *AdvancedAdminHandler {
	return &AdvancedAdminHandler{
		rbacService:     rbacService,
		sessionService:  sessionService,
		ipFilterService: ipFilterService,
		brandingRepo:    brandingRepo,
		systemRepo:      systemRepo,
		geoRepo:         geoRepo,
		log:             log,
	}
}

// ============================================================
// RBAC Endpoints
// ============================================================

// ListPermissions godoc
// @Summary List all permissions
// @Description Get a list of all available permissions in the system
// @Tags Admin - RBAC
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Permission
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/rbac/permissions [get]
func (h *AdvancedAdminHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.rbacService.ListPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// CreatePermission godoc
// @Summary Create a new permission
// @Description Create a new permission in the RBAC system
// @Tags Admin - RBAC
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param permission body models.CreatePermissionRequest true "Permission data"
// @Success 201 {object} models.Permission
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/rbac/permissions [post]
func (h *AdvancedAdminHandler) CreatePermission(c *gin.Context) {
	var req models.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	permission, err := h.rbacService.CreatePermission(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, permission)
}

// UpdatePermission godoc
// @Summary Update a permission
// @Description Update an existing permission's description
// @Tags Admin - RBAC
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Permission ID (UUID)"
// @Param permission body models.UpdatePermissionRequest true "Permission data"
// @Success 200 {object} models.Permission
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/rbac/permissions/{id} [put]
func (h *AdvancedAdminHandler) UpdatePermission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid permission ID"})
		return
	}

	var req models.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	err = h.rbacService.UpdatePermission(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Fetch updated permission to return
	permissions, err := h.rbacService.ListPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	for _, p := range permissions {
		if p.ID == id {
			c.JSON(http.StatusOK, p)
			return
		}
	}

	c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Permission not found"})
}

// DeletePermission godoc
// @Summary Delete a permission
// @Description Delete a permission by ID
// @Tags Admin - RBAC
// @Security BearerAuth
// @Param id path string true "Permission ID (UUID)"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/rbac/permissions/{id} [delete]
func (h *AdvancedAdminHandler) DeletePermission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid permission ID"})
		return
	}

	err = h.rbacService.DeletePermission(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListRoles godoc
// @Summary List all roles
// @Description Get a list of all roles in the RBAC system
// @Tags Admin - RBAC
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Role
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/rbac/roles [get]
func (h *AdvancedAdminHandler) ListRoles(c *gin.Context) {
	roles, err := h.rbacService.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, roles)
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role in the RBAC system
// @Tags Admin - RBAC
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param role body models.CreateRoleRequest true "Role data"
// @Success 201 {object} models.Role
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/rbac/roles [post]
func (h *AdvancedAdminHandler) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	role, err := h.rbacService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// GetRole godoc
// @Summary Get a role by ID
// @Description Get details of a specific role
// @Tags Admin - RBAC
// @Security BearerAuth
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Success 200 {object} models.Role
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/rbac/roles/{id} [get]
func (h *AdvancedAdminHandler) GetRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	role, err := h.rbacService.GetRole(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// UpdateRole godoc
// @Summary Update a role
// @Description Update an existing role in the RBAC system
// @Tags Admin - RBAC
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Param role body models.UpdateRoleRequest true "Role data"
// @Success 200 {object} models.Role
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/rbac/roles/{id} [put]
func (h *AdvancedAdminHandler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	role, err := h.rbacService.UpdateRole(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Delete a role from the RBAC system
// @Tags Admin - RBAC
// @Security BearerAuth
// @Param id path string true "Role ID (UUID)"
// @Success 204 "Role deleted successfully"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/rbac/roles/{id} [delete]
func (h *AdvancedAdminHandler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	err = h.rbacService.DeleteRole(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetPermissionMatrix godoc
// @Summary Get permission matrix for all roles
// @Description Get a matrix showing which roles have which permissions
// @Tags Admin - RBAC
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.PermissionMatrix
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/rbac/permission-matrix [get]
func (h *AdvancedAdminHandler) GetPermissionMatrix(c *gin.Context) {
	matrix, err := h.rbacService.GetPermissionMatrix(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, matrix)
}

// ============================================================
// Session Management Endpoints
// ============================================================

// ListUserSessions godoc
// @Summary List all sessions for the current user
// @Description Get a list of all active sessions for the authenticated user
// @Tags Sessions
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} models.SessionListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/sessions [get]
func (h *AdvancedAdminHandler) ListUserSessions(c *gin.Context) {
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	sessions, err := h.sessionService.GetUserSessions(c.Request.Context(), userID.(uuid.UUID), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// RevokeSession godoc
// @Summary Revoke a specific session
// @Description Terminate a specific session for the authenticated user
// @Tags Sessions
// @Security BearerAuth
// @Param id path string true "Session ID (UUID)"
// @Success 204 "Session revoked successfully"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/sessions/{id} [delete]
func (h *AdvancedAdminHandler) RevokeSession(c *gin.Context) {
	userID, _ := c.Get("userID")
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid session ID"})
		return
	}

	err = h.sessionService.RevokeSession(c.Request.Context(), userID.(uuid.UUID), sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AdminRevokeSession godoc
// @Summary Revoke any session (admin only)
// @Description Terminate any session by ID, regardless of owner
// @Tags Admin - Sessions
// @Security BearerAuth
// @Param id path string true "Session ID (UUID)"
// @Success 204 "Session revoked successfully"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/sessions/{id} [delete]
func (h *AdvancedAdminHandler) AdminRevokeSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid session ID"})
		return
	}

	err = h.sessionService.AdminRevokeSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RevokeAllSessions godoc
// @Summary Revoke all sessions except current
// @Description Terminate all active sessions except the current one
// @Tags Sessions
// @Security BearerAuth
// @Success 204 "All sessions revoked successfully"
// @Param user_id query string false "User ID (UUID)"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/sessions/revoke-all [post]
func (h *AdvancedAdminHandler) RevokeAllSessions(c *gin.Context) {
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
			return
		}
		err = h.sessionService.RevokeAllUserSessions(c.Request.Context(), userID, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			return
		}
	}
	currentPage := 1
	perPage := 100
	for {
		resp, err := h.sessionService.GetAllSessions(c.Request.Context(), currentPage, perPage)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			break
		}
		for _, session := range resp.Sessions {
			if err := h.sessionService.RevokeSession(c.Request.Context(), session.UserID, session.ID); err != nil {
				h.log.Error(err.Error())
				continue
			}
		}
		currentPage++
		if resp.Total > currentPage*perPage {
			break
		}
	}

	c.Status(http.StatusNoContent)
}

// GetSessionStats godoc
// @Summary Get session statistics (admin only)
// @Description Get statistics about all sessions in the system
// @Tags Admin - Sessions
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.SessionStats
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/sessions/stats [get]
func (h *AdvancedAdminHandler) GetSessionStats(c *gin.Context) {
	stats, err := h.sessionService.GetSessionStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ListAllSessions godoc
// @Summary List all sessions (admin only)
// @Description Get a list of all active sessions in the system
// @Tags Admin - Sessions
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Success 200 {array} models.ActiveSessionResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/sessions [get]
func (h *AdvancedAdminHandler) ListAllSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}

	sessions, err := h.sessionService.GetAllSessions(c.Request.Context(), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// ============================================================
// IP Filter Endpoints
// ============================================================

// ListIPFilters godoc
// @Summary List IP filters
// @Description Get a list of all IP filters (whitelist/blacklist)
// @Tags Admin - IP Filters
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param type query string false "Filter type (whitelist/blacklist)"
// @Success 200 {object} models.IPFilterListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/ip-filters [get]
func (h *AdvancedAdminHandler) ListIPFilters(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	filterType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	filters, err := h.ipFilterService.ListIPFilters(c.Request.Context(), page, perPage, filterType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, filters)
}

// CreateIPFilter godoc
// @Summary Create an IP filter
// @Description Create a new IP filter rule (whitelist or blacklist)
// @Tags Admin - IP Filters
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param filter body models.CreateIPFilterRequest true "IP filter data"
// @Success 201 {object} models.IPFilter
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/ip-filters [post]
func (h *AdvancedAdminHandler) CreateIPFilter(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.CreateIPFilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	filter, err := h.ipFilterService.CreateIPFilter(c.Request.Context(), &req, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, filter)
}

// DeleteIPFilter godoc
// @Summary Delete an IP filter
// @Description Delete an existing IP filter rule
// @Tags Admin - IP Filters
// @Security BearerAuth
// @Param id path string true "Filter ID (UUID)"
// @Success 204 "IP filter deleted successfully"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/ip-filters/{id} [delete]
func (h *AdvancedAdminHandler) DeleteIPFilter(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid filter ID"})
		return
	}

	err = h.ipFilterService.DeleteIPFilter(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ============================================================
// Branding Endpoints
// ============================================================

// GetBranding godoc
// @Summary Get branding settings
// @Description Get public branding settings (logo, colors, company info)
// @Tags Branding
// @Produce json
// @Success 200 {object} models.PublicBrandingResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/branding [get]
func (h *AdvancedAdminHandler) GetBranding(c *gin.Context) {
	settings, err := h.brandingRepo.GetBrandingSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Return public branding info
	response := models.PublicBrandingResponse{
		LogoURL:    settings.LogoURL,
		FaviconURL: settings.FaviconURL,
		Theme: models.BrandingTheme{
			PrimaryColor:    settings.PrimaryColor,
			SecondaryColor:  settings.SecondaryColor,
			BackgroundColor: settings.BackgroundColor,
		},
		CompanyName:  settings.CompanyName,
		SupportEmail: settings.SupportEmail,
		TermsURL:     settings.TermsURL,
		PrivacyURL:   settings.PrivacyURL,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateBranding godoc
// @Summary Update branding settings (admin only)
// @Description Update the system's branding and theming settings
// @Tags Admin - Branding
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param branding body models.UpdateBrandingRequest true "Branding data"
// @Success 200 {object} models.BrandingSettings
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/branding [put]
func (h *AdvancedAdminHandler) UpdateBranding(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.UpdateBrandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Get current settings
	settings, err := h.brandingRepo.GetBrandingSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Update fields
	if req.LogoURL != "" {
		settings.LogoURL = req.LogoURL
	}
	if req.FaviconURL != "" {
		settings.FaviconURL = req.FaviconURL
	}
	if req.PrimaryColor != "" {
		settings.PrimaryColor = req.PrimaryColor
	}
	if req.SecondaryColor != "" {
		settings.SecondaryColor = req.SecondaryColor
	}
	if req.BackgroundColor != "" {
		settings.BackgroundColor = req.BackgroundColor
	}
	if req.CustomCSS != "" {
		settings.CustomCSS = req.CustomCSS
	}
	if req.CompanyName != "" {
		settings.CompanyName = req.CompanyName
	}
	if req.SupportEmail != "" {
		settings.SupportEmail = req.SupportEmail
	}
	if req.TermsURL != "" {
		settings.TermsURL = req.TermsURL
	}
	if req.PrivacyURL != "" {
		settings.PrivacyURL = req.PrivacyURL
	}

	err = h.brandingRepo.UpdateBrandingSettings(c.Request.Context(), settings, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// ============================================================
// System Settings & Health Endpoints
// ============================================================

// GetMaintenanceMode godoc
// @Summary Get maintenance mode status
// @Description Check if the system is in maintenance mode
// @Tags System
// @Produce json
// @Success 200 {object} models.MaintenanceModeResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/system/maintenance [get]
func (h *AdvancedAdminHandler) GetMaintenanceMode(c *gin.Context) {
	setting, err := h.systemRepo.GetSetting(c.Request.Context(), models.SettingMaintenanceMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	messageSetting, _ := h.systemRepo.GetSetting(c.Request.Context(), models.SettingMaintenanceMessage)

	response := models.MaintenanceModeResponse{
		Enabled: setting.Value == "true",
		Message: messageSetting.Value,
	}

	c.JSON(http.StatusOK, response)
}

// SetMaintenanceMode godoc
// @Summary Set maintenance mode (admin only)
// @Description Enable or disable system maintenance mode
// @Tags Admin - System
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param mode body models.MaintenanceModeRequest true "Maintenance mode data"
// @Success 200 {object} models.MaintenanceModeResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/system/maintenance [put]
func (h *AdvancedAdminHandler) SetMaintenanceMode(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.MaintenanceModeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	uid := userID.(uuid.UUID)

	// Update maintenance mode
	value := "false"
	if req.Enabled {
		value = "true"
	}
	err := h.systemRepo.UpdateSetting(c.Request.Context(), models.SettingMaintenanceMode, value, &uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Update message if provided
	if req.Message != "" {
		err = h.systemRepo.UpdateSetting(c.Request.Context(), models.SettingMaintenanceMessage, req.Message, &uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			return
		}
	}

	response := models.MaintenanceModeResponse{
		Enabled: req.Enabled,
		Message: req.Message,
	}

	c.JSON(http.StatusOK, response)
}

// GetSystemHealth godoc
// @Summary Get system health metrics
// @Description Get health status of system components (database, redis, etc.)
// @Tags Admin - System
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.SystemHealthResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/system/health [get]
func (h *AdvancedAdminHandler) GetSystemHealth(c *gin.Context) {
	// This is a placeholder - implement actual health checks
	response := models.SystemHealthResponse{
		Status:         "healthy",
		DatabaseStatus: "healthy",
		RedisStatus:    "healthy",
		Uptime:         0,
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================
// Geo-Distribution Endpoints
// ============================================================

// GetGeoDistribution godoc
// @Summary Get login geo-distribution for map
// @Description Get geographical distribution of logins for analytics and mapping
// @Tags Admin - Analytics
// @Security BearerAuth
// @Produce json
// @Param days query int false "Number of days" default(30)
// @Success 200 {object} models.GeoDistributionResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/analytics/geo-distribution [get]
func (h *AdvancedAdminHandler) GetGeoDistribution(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	locations, err := h.geoRepo.GetLoginGeoDistribution(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Calculate totals
	totalLogins := 0
	countryMap := make(map[string]bool)
	cityMap := make(map[string]bool)

	for _, loc := range locations {
		totalLogins += loc.LoginCount
		countryMap[loc.CountryCode] = true
		if loc.City != "" {
			cityMap[loc.City] = true
		}
	}

	response := models.GeoDistributionResponse{
		Locations: locations,
		Total:     totalLogins,
		Countries: len(countryMap),
		Cities:    len(cityMap),
	}

	c.JSON(http.StatusOK, response)
}
