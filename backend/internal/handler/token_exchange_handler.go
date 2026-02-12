package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
)

type TokenExchangeHandler struct {
	tokenExchangeService *service.TokenExchangeService
}

func NewTokenExchangeHandler(tokenExchangeService *service.TokenExchangeService) *TokenExchangeHandler {
	return &TokenExchangeHandler{
		tokenExchangeService: tokenExchangeService,
	}
}

func (h *TokenExchangeHandler) CreateExchange(c *gin.Context) {
	var req models.CreateTokenExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	appID, _ := utils.GetApplicationIDFromContext(c)

	resp, err := h.tokenExchangeService.CreateExchange(c.Request.Context(), &req, appID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TokenExchangeHandler) RedeemExchange(c *gin.Context) {
	var req models.RedeemTokenExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	appID, _ := utils.GetApplicationIDFromContext(c)

	resp, err := h.tokenExchangeService.RedeemExchange(c.Request.Context(), &req, appID)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
