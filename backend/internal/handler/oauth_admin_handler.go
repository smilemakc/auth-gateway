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

type OAuthAdminHandler struct {
	service *service.OAuthProviderService
	logger  *logger.Logger
}

func NewOAuthAdminHandler(service *service.OAuthProviderService, logger *logger.Logger) *OAuthAdminHandler {
	return &OAuthAdminHandler{
		service: service,
		logger:  logger,
	}
}

// CreateClient creates a new OAuth client
// @Summary Create OAuth client
// @Description Create a new OAuth 2.0 client application (admin only)
// @Tags Admin - OAuth Clients
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateOAuthClientRequest true "Client creation data"
// @Success 201 {object} models.CreateOAuthClientResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/clients [post]
func (h *OAuthAdminHandler) CreateClient(c *gin.Context) {
	var req models.CreateOAuthClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	response, err := h.service.CreateClient(c.Request.Context(), &req, nil)
	if err != nil {
		h.logger.Error("Failed to create OAuth client", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListClients lists OAuth clients with pagination
// @Summary List OAuth clients
// @Description Get paginated list of OAuth 2.0 clients (admin only)
// @Tags Admin - OAuth Clients
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param owner_id query string false "Filter by owner ID"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} models.OAuthClientListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/clients [get]
func (h *OAuthAdminHandler) ListClients(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)

	var opts []service.OAuthClientListOption

	if ownerIDStr := c.Query("owner_id"); ownerIDStr != "" {
		id, err := uuid.Parse(ownerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Invalid owner ID"),
			))
			return
		}
		opts = append(opts, service.OAuthClientListOwner(id))
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		val := isActiveStr == "true"
		opts = append(opts, service.OAuthClientListActive(&val))
	}

	clients, total, err := h.service.ListClients(c.Request.Context(), page, pageSize, opts...)
	if err != nil {
		h.logger.Error("Failed to list OAuth clients", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	totalPages := (total + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, models.OAuthClientListResponse{
		Clients:    clients,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetClient gets a single OAuth client by ID
// @Summary Get OAuth client
// @Description Get detailed information about a specific OAuth client (admin only)
// @Tags Admin - OAuth Clients
// @Security BearerAuth
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} models.OAuthClient
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/clients/{id} [get]
func (h *OAuthAdminHandler) GetClient(c *gin.Context) {
	clientID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	client, err := h.service.GetClient(c.Request.Context(), clientID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.NewAppError(http.StatusNotFound, "Client not found"),
		))
		return
	}

	c.JSON(http.StatusOK, client)
}

// UpdateClient updates an OAuth client
// @Summary Update OAuth client
// @Description Update OAuth 2.0 client information (admin only)
// @Tags Admin - OAuth Clients
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body models.UpdateOAuthClientRequest true "Client update data"
// @Success 200 {object} models.OAuthClient
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/clients/{id} [put]
func (h *OAuthAdminHandler) UpdateClient(c *gin.Context) {
	clientID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.UpdateOAuthClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	client, err := h.service.UpdateClient(c.Request.Context(), clientID, &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to update OAuth client", map[string]interface{}{
			"error":     err.Error(),
			"client_id": clientID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, client)
}

// DeleteClient soft-deletes an OAuth client
// @Summary Delete OAuth client
// @Description Soft-delete an OAuth 2.0 client (admin only)
// @Tags Admin - OAuth Clients
// @Security BearerAuth
// @Param id path string true "Client ID"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/clients/{id} [delete]
func (h *OAuthAdminHandler) DeleteClient(c *gin.Context) {
	clientID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteClient(c.Request.Context(), clientID); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to delete OAuth client", map[string]interface{}{
			"error":     err.Error(),
			"client_id": clientID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.Status(http.StatusNoContent)
}

// RotateSecret generates a new client secret
// @Summary Rotate client secret
// @Description Generate a new client secret for an OAuth 2.0 client (admin only)
// @Tags Admin - OAuth Clients
// @Security BearerAuth
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} map[string]string "Response with client_secret"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/clients/{id}/rotate-secret [post]
func (h *OAuthAdminHandler) RotateSecret(c *gin.Context) {
	clientID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	clientSecret, err := h.service.RotateClientSecret(c.Request.Context(), clientID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to rotate client secret", map[string]interface{}{
			"error":     err.Error(),
			"client_id": clientID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_secret": clientSecret,
	})
}

// ListScopes lists all OAuth scopes
// @Summary List OAuth scopes
// @Description Get list of all OAuth 2.0 scopes (admin only)
// @Tags Admin - OAuth Scopes
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.OAuthScopeListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/scopes [get]
func (h *OAuthAdminHandler) ListScopes(c *gin.Context) {
	scopes, err := h.service.ListScopes(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list OAuth scopes", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, models.OAuthScopeListResponse{Scopes: scopes, Total: len(scopes)})
}

// CreateScope creates a custom OAuth scope
// @Summary Create OAuth scope
// @Description Create a custom OAuth 2.0 scope (admin only, is_system=false)
// @Tags Admin - OAuth Scopes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Scope data with name, display_name, description"
// @Success 201 {object} models.OAuthScope
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/scopes [post]
func (h *OAuthAdminHandler) CreateScope(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=50"`
		DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	scope := &models.OAuthScope{
		ID:          uuid.New(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsDefault:   false,
		IsSystem:    false,
	}

	if err := h.service.CreateScope(c.Request.Context(), scope); err != nil {
		h.logger.Error("Failed to create OAuth scope", map[string]interface{}{
			"error": err.Error(),
			"name":  req.Name,
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusCreated, scope)
}

// DeleteScope deletes a non-system OAuth scope
// @Summary Delete OAuth scope
// @Description Delete a custom OAuth 2.0 scope (admin only, cannot delete system scopes)
// @Tags Admin - OAuth Scopes
// @Security BearerAuth
// @Param id path string true "Scope ID"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/scopes/{id} [delete]
func (h *OAuthAdminHandler) DeleteScope(c *gin.Context) {
	scopeID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteScope(c.Request.Context(), scopeID); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to delete OAuth scope", map[string]interface{}{
			"error":    err.Error(),
			"scope_id": scopeID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.Status(http.StatusNoContent)
}

// ListClientConsents lists all user consents for a client
// @Summary List client consents
// @Description Get all user consents for a specific OAuth client (admin only)
// @Tags Admin - OAuth Consents
// @Security BearerAuth
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} models.OAuthConsentListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/clients/{id}/consents [get]
func (h *OAuthAdminHandler) ListClientConsents(c *gin.Context) {
	clientID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	consents, err := h.service.ListClientConsents(c.Request.Context(), clientID)
	if err != nil {
		h.logger.Error("Failed to list client consents", map[string]interface{}{
			"error":     err.Error(),
			"client_id": clientID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, models.OAuthConsentListResponse{Consents: consents, Total: len(consents)})
}

// RevokeUserConsent revokes user consent for a client
// @Summary Revoke user consent
// @Description Revoke a user's consent for a specific OAuth client (admin action)
// @Tags Admin - OAuth Consents
// @Security BearerAuth
// @Param id path string true "Client ID"
// @Param user_id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/oauth/clients/{id}/consents/{user_id} [delete]
func (h *OAuthAdminHandler) RevokeUserConsent(c *gin.Context) {
	clientID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	userID, ok := utils.ParseUUIDParam(c, "user_id")
	if !ok {
		return
	}

	if err := h.service.RevokeConsent(c.Request.Context(), userID, clientID); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
			return
		}
		h.logger.Error("Failed to revoke user consent", map[string]interface{}{
			"error":     err.Error(),
			"client_id": clientID.String(),
			"user_id":   userID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.Status(http.StatusNoContent)
}
