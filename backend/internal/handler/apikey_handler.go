package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// APIKeyHandler handles API key-related HTTP requests
type APIKeyHandler struct {
	apiKeyService *service.APIKeyService
	logger        *logger.Logger
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService *service.APIKeyService, log *logger.Logger) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
		logger:        log,
	}
}

// Create creates a new API key
// @Summary Create API key
// @Description Create a new API key for service-to-service authentication
// @Tags API Keys
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateAPIKeyRequest true "API key data"
// @Success 201 {object} models.CreateAPIKeyResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/api-keys [post]
func (h *APIKeyHandler) Create(c *gin.Context) {
	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	var req models.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	resp, err := h.apiKeyService.Create(c.Request.Context(), userID, &req, ip, userAgent)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// List lists all API keys for the authenticated user
// @Summary List API keys
// @Description Get all API keys belonging to the authenticated user
// @Tags API Keys
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.ListAPIKeysResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/api-keys [get]
func (h *APIKeyHandler) List(c *gin.Context) {
	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	apiKeys, err := h.apiKeyService.List(c.Request.Context(), userID)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.ListAPIKeysResponse{
		APIKeys: apiKeys,
		Total:   len(apiKeys),
	})
}

// Get retrieves a specific API key
// @Summary Get API key
// @Description Get details of a specific API key owned by the authenticated user
// @Tags API Keys
// @Security BearerAuth
// @Produce json
// @Param id path string true "API Key ID (UUID)"
// @Success 200 {object} models.APIKey
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/api-keys/{id} [get]
func (h *APIKeyHandler) Get(c *gin.Context) {
	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	apiKeyID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), userID, apiKeyID)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// Update updates an API key
// @Summary Update API key
// @Description Update API key metadata (name, scopes, etc.)
// @Tags API Keys
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "API Key ID (UUID)"
// @Param request body models.UpdateAPIKeyRequest true "Update data"
// @Success 200 {object} models.APIKey
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/api-keys/{id} [put]
func (h *APIKeyHandler) Update(c *gin.Context) {
	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	apiKeyID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	apiKey, err := h.apiKeyService.Update(c.Request.Context(), userID, apiKeyID, &req, ip, userAgent)
	if err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// Revoke revokes an API key
// @Summary Revoke API key
// @Description Revoke an API key to prevent further use
// @Tags API Keys
// @Security BearerAuth
// @Produce json
// @Param id path string true "API Key ID (UUID)"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/api-keys/{id}/revoke [post]
func (h *APIKeyHandler) Revoke(c *gin.Context) {
	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	apiKeyID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.apiKeyService.Revoke(c.Request.Context(), userID, apiKeyID, ip, userAgent); err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "API key revoked successfully"})
}

// Delete deletes an API key
// @Summary Delete API key
// @Description Permanently delete an API key
// @Tags API Keys
// @Security BearerAuth
// @Produce json
// @Param id path string true "API Key ID (UUID)"
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/api-keys/{id} [delete]
func (h *APIKeyHandler) Delete(c *gin.Context) {
	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	apiKeyID, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	ip := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.apiKeyService.Delete(c.Request.Context(), userID, apiKeyID, ip, userAgent); err != nil {
		utils.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "API key deleted successfully"})
}
