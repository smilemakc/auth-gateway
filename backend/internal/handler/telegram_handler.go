package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

type TelegramHandler struct {
	telegramService *service.TelegramService
	logger          *logger.Logger
}

func NewTelegramHandler(telegramService *service.TelegramService, logger *logger.Logger) *TelegramHandler {
	return &TelegramHandler{
		telegramService: telegramService,
		logger:          logger,
	}
}

// CreateBot creates a new Telegram bot
// @Summary Create Telegram bot
// @Description Create a new Telegram bot for an application (admin only)
// @Tags Admin - Telegram Bots
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param request body models.CreateTelegramBotRequest true "Bot creation data"
// @Success 201 {object} models.TelegramBot
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/telegram-bots [post]
func (h *TelegramHandler) CreateBot(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	var req models.CreateTelegramBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	bot, err := h.telegramService.CreateBot(c.Request.Context(), appID, &req)
	if err != nil {
		if err == service.ErrApplicationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Application not found"),
			))
			return
		}
		h.logger.Error("Failed to create Telegram bot", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusCreated, bot)
}

// GetBot retrieves a Telegram bot by ID
// @Summary Get Telegram bot
// @Description Get Telegram bot details by ID (admin only)
// @Tags Admin - Telegram Bots
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param botId path string true "Bot ID (UUID)"
// @Success 200 {object} models.TelegramBot
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/telegram-bots/{botId} [get]
func (h *TelegramHandler) GetBot(c *gin.Context) {
	botID, err := h.parseIDParam(c, "botId")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid bot ID"),
		))
		return
	}

	bot, err := h.telegramService.GetBot(c.Request.Context(), botID)
	if err != nil {
		if err == service.ErrBotNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Telegram bot not found"),
			))
			return
		}
		h.logger.Error("Failed to get Telegram bot", map[string]interface{}{
			"error":  err.Error(),
			"bot_id": botID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, bot)
}

// ListBots returns all Telegram bots for an application
// @Summary List Telegram bots
// @Description Get all Telegram bots for an application (admin only)
// @Tags Admin - Telegram Bots
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Success 200 {object} models.TelegramBotListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/telegram-bots [get]
func (h *TelegramHandler) ListBots(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	bots, err := h.telegramService.ListBotsByApp(c.Request.Context(), appID)
	if err != nil {
		h.logger.Error("Failed to list Telegram bots", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, models.TelegramBotListResponse{
		Bots:  bots,
		Total: len(bots),
	})
}

// UpdateBot updates a Telegram bot
// @Summary Update Telegram bot
// @Description Update Telegram bot details (admin only)
// @Tags Admin - Telegram Bots
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param botId path string true "Bot ID (UUID)"
// @Param request body models.UpdateTelegramBotRequest true "Bot update data"
// @Success 200 {object} models.TelegramBot
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/telegram-bots/{botId} [put]
func (h *TelegramHandler) UpdateBot(c *gin.Context) {
	botID, err := h.parseIDParam(c, "botId")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid bot ID"),
		))
		return
	}

	var req models.UpdateTelegramBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	bot, err := h.telegramService.UpdateBot(c.Request.Context(), botID, &req)
	if err != nil {
		if err == service.ErrBotNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Telegram bot not found"),
			))
			return
		}
		h.logger.Error("Failed to update Telegram bot", map[string]interface{}{
			"error":  err.Error(),
			"bot_id": botID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, bot)
}

// DeleteBot deletes a Telegram bot
// @Summary Delete Telegram bot
// @Description Delete a Telegram bot (admin only)
// @Tags Admin - Telegram Bots
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param botId path string true "Bot ID (UUID)"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/telegram-bots/{botId} [delete]
func (h *TelegramHandler) DeleteBot(c *gin.Context) {
	botID, err := h.parseIDParam(c, "botId")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid bot ID"),
		))
		return
	}

	if err := h.telegramService.DeleteBot(c.Request.Context(), botID); err != nil {
		if err == service.ErrBotNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Telegram bot not found"),
			))
			return
		}
		h.logger.Error("Failed to delete Telegram bot", map[string]interface{}{
			"error":  err.Error(),
			"bot_id": botID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.Status(http.StatusNoContent)
}

// ListUserTelegramAccounts returns user's Telegram accounts
// @Summary List user Telegram accounts
// @Description Get all Telegram accounts linked to a user (admin only)
// @Tags Admin - User Telegram
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} models.UserTelegramAccountListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/telegram-accounts [get]
func (h *TelegramHandler) ListUserTelegramAccounts(c *gin.Context) {
	userID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	accounts, err := h.telegramService.ListAccountsByUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list user Telegram accounts", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, models.UserTelegramAccountListResponse{
		Accounts: accounts,
		Total:    len(accounts),
	})
}

// ListUserTelegramBotAccess returns user's Telegram bot access records
// @Summary List user Telegram bot access
// @Description Get all Telegram bot access records for a user, optionally filtered by application (admin only)
// @Tags Admin - User Telegram
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param app_id query string false "Application ID (UUID) to filter by"
// @Success 200 {object} models.UserTelegramBotAccessListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users/{id}/telegram-bot-access [get]
func (h *TelegramHandler) ListUserTelegramBotAccess(c *gin.Context) {
	userID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid user ID"),
		))
		return
	}

	appIDStr := c.Query("app_id")
	var botAccess []*models.UserTelegramBotAccess

	if appIDStr != "" {
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
			))
			return
		}

		botAccess, err = h.telegramService.ListBotAccessByUserAndApp(c.Request.Context(), userID, appID)
		if err != nil {
			h.logger.Error("Failed to list user Telegram bot access by app", map[string]interface{}{
				"error":          err.Error(),
				"user_id":        userID.String(),
				"application_id": appID.String(),
			})
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}
	} else {
		botAccess, err = h.telegramService.ListBotAccessByUser(c.Request.Context(), userID)
		if err != nil {
			h.logger.Error("Failed to list user Telegram bot access", map[string]interface{}{
				"error":   err.Error(),
				"user_id": userID.String(),
			})
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}
	}

	c.JSON(http.StatusOK, models.UserTelegramBotAccessListResponse{
		Access: botAccess,
		Total:  len(botAccess),
	})
}

func (h *TelegramHandler) parseIDParam(c *gin.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
