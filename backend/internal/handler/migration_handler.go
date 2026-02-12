package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

type MigrationHandler struct {
	migrationService *service.MigrationService
	logger           *logger.Logger
}

func NewMigrationHandler(migrationService *service.MigrationService, logger *logger.Logger) *MigrationHandler {
	return &MigrationHandler{
		migrationService: migrationService,
		logger:           logger,
	}
}

func (h *MigrationHandler) ImportUsers(c *gin.Context) {
	var req models.ImportUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid request body", err.Error())))
		return
	}

	appID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		h.logger.Error("Invalid application ID", map[string]interface{}{
			"application_id": req.ApplicationID,
			"error":          err.Error(),
		})
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid application ID")))
		return
	}

	result, err := h.migrationService.ImportUsers(c.Request.Context(), appID, req.Users)
	if err != nil {
		h.logger.Error("Failed to import users", map[string]interface{}{
			"application_id": appID.String(),
			"error":          err.Error(),
		})
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		}
		return
	}

	h.logger.Info("Users imported successfully", map[string]interface{}{
		"application_id": appID.String(),
		"total":          result.Total,
		"created":        result.Created,
		"skipped":        result.Skipped,
	})

	c.JSON(http.StatusOK, result)
}

func (h *MigrationHandler) ImportOAuthAccounts(c *gin.Context) {
	var req models.ImportOAuthAccountsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid request body", err.Error())))
		return
	}

	result, err := h.migrationService.ImportOAuthAccounts(c.Request.Context(), req.Accounts)
	if err != nil {
		h.logger.Error("Failed to import OAuth accounts", map[string]interface{}{
			"error": err.Error(),
		})
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		}
		return
	}

	h.logger.Info("OAuth accounts imported successfully", map[string]interface{}{
		"total":   result.Total,
		"created": result.Created,
		"skipped": result.Skipped,
	})

	c.JSON(http.StatusOK, result)
}

func (h *MigrationHandler) ImportRoles(c *gin.Context) {
	var req models.ImportRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid request body", err.Error())))
		return
	}

	appID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		h.logger.Error("Invalid application ID", map[string]interface{}{
			"application_id": req.ApplicationID,
			"error":          err.Error(),
		})
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.NewAppError(http.StatusBadRequest, "Invalid application ID")))
		return
	}

	result, err := h.migrationService.ImportRoles(c.Request.Context(), appID, req.Roles)
	if err != nil {
		h.logger.Error("Failed to import roles", map[string]interface{}{
			"application_id": appID.String(),
			"error":          err.Error(),
		})
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		}
		return
	}

	h.logger.Info("Roles imported successfully", map[string]interface{}{
		"application_id": appID.String(),
		"total":          result.Total,
		"created":        result.Created,
		"skipped":        result.Skipped,
	})

	c.JSON(http.StatusOK, result)
}
