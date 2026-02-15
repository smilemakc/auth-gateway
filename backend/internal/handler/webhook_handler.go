package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// WebhookHandler handles webhook endpoints
type WebhookHandler struct {
	webhookService *service.WebhookService
	logger         *logger.Logger
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService *service.WebhookService, log *logger.Logger) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		logger:         log,
	}
}

// ListWebhooks godoc
// @Summary List all webhooks
// @Description Get a paginated list of all webhooks (admin only)
// @Tags Admin - Webhooks
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Security BearerAuth
// @Success 200 {object} models.WebhookListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/webhooks [get]
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	appID, _ := utils.GetApplicationIDFromContext(c)
	if appID != nil {
		webhooks, err := h.webhookService.ListWebhooksByApp(c.Request.Context(), *appID)
		if err != nil {
			h.logger.Error("Failed to list webhooks", map[string]interface{}{"error": err.Error()})
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to list webhooks"})
			return
		}
		wrapped := make([]models.WebhookWithCreator, len(webhooks))
		for i, w := range webhooks {
			wrapped[i] = models.WebhookWithCreator{Webhook: *w}
		}
		c.JSON(http.StatusOK, models.WebhookListResponse{
			Webhooks:   wrapped,
			Total:      len(wrapped),
			Page:       1,
			PageSize:   len(wrapped),
			TotalPages: 1,
		})
		return
	}

	page, pageSize := utils.ParsePagination(c)

	resp, err := h.webhookService.ListWebhooks(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list webhooks", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to list webhooks"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetWebhook godoc
// @Summary Get webhook by ID
// @Description Get a specific webhook by its ID (admin only)
// @Tags Admin - Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} models.Webhook
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/webhooks/{id} [get]
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	webhook, err := h.webhookService.GetWebhook(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Webhook not found"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// CreateWebhook godoc
// @Summary Create a new webhook
// @Description Create a new webhook for receiving event notifications (admin only)
// @Tags Admin - Webhooks
// @Accept json
// @Produce json
// @Param request body models.CreateWebhookRequest true "Webhook data"
// @Security BearerAuth
// @Success 201 {object} models.CreateWebhookResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/webhooks [post]
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	var req models.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	webhook, secretKey, err := h.webhookService.CreateWebhook(c.Request.Context(), &req, userID)
	if err != nil {
		h.logger.Error("Failed to create webhook", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"webhook":    webhook,
		"secret_key": secretKey,
	})
}

// UpdateWebhook godoc
// @Summary Update a webhook
// @Description Update an existing webhook configuration (admin only)
// @Tags Admin - Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID (UUID)"
// @Param request body models.UpdateWebhookRequest true "Webhook update data"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/webhooks/{id} [put]
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.webhookService.UpdateWebhook(c.Request.Context(), id, &req, userID); err != nil {
		h.logger.Error("Failed to update webhook", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Webhook updated successfully"})
}

// DeleteWebhook godoc
// @Summary Delete a webhook
// @Description Delete a webhook by ID (admin only)
// @Tags Admin - Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/webhooks/{id} [delete]
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	userID, ok := utils.MustGetUserID(c)
	if !ok {
		return
	}

	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.webhookService.DeleteWebhook(c.Request.Context(), id, userID); err != nil {
		h.logger.Error("Failed to delete webhook", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Webhook not found"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Webhook deleted successfully"})
}

// TestWebhook godoc
// @Summary Test a webhook
// @Description Send a test event to a webhook (admin only)
// @Tags Admin - Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID (UUID)"
// @Param request body models.TestWebhookRequest true "Test data"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/webhooks/{id}/test [post]
func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.TestWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.webhookService.TestWebhook(c.Request.Context(), id, &req); err != nil {
		h.logger.Error("Failed to test webhook", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Webhook not found"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Test webhook sent"})
}

// ListWebhookDeliveries godoc
// @Summary List webhook deliveries
// @Description Get a paginated list of webhook delivery attempts (admin only)
// @Tags Admin - Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Security BearerAuth
// @Success 200 {object} models.WebhookDeliveryListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/webhooks/{id}/deliveries [get]
func (h *WebhookHandler) ListWebhookDeliveries(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	page, pageSize := utils.ParsePagination(c)

	resp, err := h.webhookService.ListWebhookDeliveries(c.Request.Context(), id, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list deliveries", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to list deliveries"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetAvailableEvents godoc
// @Summary Get available webhook events
// @Description Get a list of all available webhook event types (admin only)
// @Tags Admin - Webhooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.WebhookEventsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/webhooks/events [get]
func (h *WebhookHandler) GetAvailableEvents(c *gin.Context) {
	events := h.webhookService.GetAvailableEvents()
	c.JSON(http.StatusOK, gin.H{"events": events})
}
