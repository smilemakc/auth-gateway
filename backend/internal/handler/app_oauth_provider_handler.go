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

type AppOAuthProviderHandler struct {
	service *service.AppOAuthProviderService
	logger  *logger.Logger
}

func NewAppOAuthProviderHandler(service *service.AppOAuthProviderService, logger *logger.Logger) *AppOAuthProviderHandler {
	return &AppOAuthProviderHandler{
		service: service,
		logger:  logger,
	}
}

// CreateProvider creates a new OAuth provider for an application
// @Summary Create OAuth provider
// @Description Create a new OAuth provider for an application (admin only)
// @Tags Admin - Application OAuth Providers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param request body models.CreateAppOAuthProviderRequest true "OAuth provider creation data"
// @Success 201 {object} models.ApplicationOAuthProvider
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/oauth-providers [post]
func (h *AppOAuthProviderHandler) CreateProvider(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	var req models.CreateAppOAuthProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	provider, err := h.service.Create(c.Request.Context(), appID, &req)
	if err != nil {
		if err == service.ErrApplicationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Application not found"),
			))
			return
		}
		if err == service.ErrOAuthProviderNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "OAuth provider not found"),
			))
			return
		}
		if err == service.ErrOAuthProviderExists {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				models.NewAppError(http.StatusConflict, "OAuth provider already configured for this application"),
			))
			return
		}
		if err == service.ErrInvalidOAuthProvider {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Invalid OAuth provider"),
			))
			return
		}
		if err == service.ErrMissingRequiredCredentials {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Missing required credentials"),
			))
			return
		}
		h.logger.Error("Failed to create OAuth provider", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// GetProvider retrieves an OAuth provider by ID
// @Summary Get OAuth provider
// @Description Get OAuth provider details by ID (admin only)
// @Tags Admin - Application OAuth Providers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param providerId path string true "OAuth Provider ID (UUID)"
// @Success 200 {object} models.ApplicationOAuthProvider
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/oauth-providers/{providerId} [get]
func (h *AppOAuthProviderHandler) GetProvider(c *gin.Context) {
	providerID, err := h.parseIDParam(c, "providerId")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid provider ID"),
		))
		return
	}

	provider, err := h.service.GetByID(c.Request.Context(), providerID)
	if err != nil {
		if err == service.ErrOAuthProviderNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "OAuth provider not found"),
			))
			return
		}
		h.logger.Error("Failed to get OAuth provider", map[string]interface{}{
			"error":       err.Error(),
			"provider_id": providerID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, provider)
}

// ListProviders returns all OAuth providers for an application
// @Summary List OAuth providers
// @Description Get list of OAuth providers for an application (admin only)
// @Tags Admin - Application OAuth Providers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Success 200 {array} models.ApplicationOAuthProvider
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/oauth-providers [get]
func (h *AppOAuthProviderHandler) ListProviders(c *gin.Context) {
	appID, err := h.parseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid application ID"),
		))
		return
	}

	providers, err := h.service.ListByApp(c.Request.Context(), appID)
	if err != nil {
		h.logger.Error("Failed to list OAuth providers", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, providers)
}

// UpdateProvider updates an OAuth provider
// @Summary Update OAuth provider
// @Description Update OAuth provider details (admin only)
// @Tags Admin - Application OAuth Providers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param providerId path string true "OAuth Provider ID (UUID)"
// @Param request body models.UpdateAppOAuthProviderRequest true "OAuth provider update data"
// @Success 200 {object} models.ApplicationOAuthProvider
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/oauth-providers/{providerId} [put]
func (h *AppOAuthProviderHandler) UpdateProvider(c *gin.Context) {
	providerID, err := h.parseIDParam(c, "providerId")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid provider ID"),
		))
		return
	}

	var req models.UpdateAppOAuthProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	provider, err := h.service.Update(c.Request.Context(), providerID, &req)
	if err != nil {
		if err == service.ErrOAuthProviderNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "OAuth provider not found"),
			))
			return
		}
		if err == service.ErrMissingRequiredCredentials {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Missing required credentials"),
			))
			return
		}
		h.logger.Error("Failed to update OAuth provider", map[string]interface{}{
			"error":       err.Error(),
			"provider_id": providerID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, provider)
}

// DeleteProvider deletes an OAuth provider
// @Summary Delete OAuth provider
// @Description Delete an OAuth provider (admin only)
// @Tags Admin - Application OAuth Providers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Application ID (UUID)"
// @Param providerId path string true "OAuth Provider ID (UUID)"
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{id}/oauth-providers/{providerId} [delete]
func (h *AppOAuthProviderHandler) DeleteProvider(c *gin.Context) {
	providerID, err := h.parseIDParam(c, "providerId")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid provider ID"),
		))
		return
	}

	if err := h.service.Delete(c.Request.Context(), providerID); err != nil {
		if err == service.ErrOAuthProviderNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "OAuth provider not found"),
			))
			return
		}
		h.logger.Error("Failed to delete OAuth provider", map[string]interface{}{
			"error":       err.Error(),
			"provider_id": providerID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AppOAuthProviderHandler) parseIDParam(c *gin.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// ListProvidersAdmin lists OAuth providers, filtered by X-Application-ID header if present
func (h *AppOAuthProviderHandler) ListProvidersAdmin(c *gin.Context) {
	appID, _ := utils.GetApplicationIDFromContext(c)

	if appID != nil {
		providers, err := h.service.ListByApp(c.Request.Context(), *appID)
		if err != nil {
			h.logger.Error("Failed to list OAuth providers", map[string]interface{}{"error": err.Error()})
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
			return
		}
		c.JSON(http.StatusOK, gin.H{"providers": providers})
		return
	}

	providers, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list OAuth providers", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// CreateProviderAdmin creates an OAuth provider using X-Application-ID from context
func (h *AppOAuthProviderHandler) CreateProviderAdmin(c *gin.Context) {
	appID, ok := utils.GetApplicationIDFromContext(c)
	if !ok || appID == nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "X-Application-ID header is required to create OAuth provider"),
		))
		return
	}

	var req models.CreateAppOAuthProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	provider, err := h.service.Create(c.Request.Context(), *appID, &req)
	if err != nil {
		if err == service.ErrApplicationNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "Application not found"),
			))
			return
		}
		if err == service.ErrOAuthProviderExists {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				models.NewAppError(http.StatusConflict, "OAuth provider already configured for this application"),
			))
			return
		}
		if err == service.ErrInvalidOAuthProvider {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Invalid OAuth provider"),
			))
			return
		}
		if err == service.ErrMissingRequiredCredentials {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Missing required credentials"),
			))
			return
		}
		h.logger.Error("Failed to create OAuth provider", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusCreated, provider)
}

// GetProviderAdmin gets an OAuth provider by ID from path param
func (h *AppOAuthProviderHandler) GetProviderAdmin(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("providerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid provider ID"),
		))
		return
	}

	provider, err := h.service.GetByID(c.Request.Context(), providerID)
	if err != nil {
		if err == service.ErrOAuthProviderNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "OAuth provider not found"),
			))
			return
		}
		h.logger.Error("Failed to get OAuth provider", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, provider)
}

// UpdateProviderAdmin updates an OAuth provider by ID
func (h *AppOAuthProviderHandler) UpdateProviderAdmin(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("providerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid provider ID"),
		))
		return
	}

	var req models.UpdateAppOAuthProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	provider, err := h.service.Update(c.Request.Context(), providerID, &req)
	if err != nil {
		if err == service.ErrOAuthProviderNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "OAuth provider not found"),
			))
			return
		}
		if err == service.ErrMissingRequiredCredentials {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse(
				models.NewAppError(http.StatusBadRequest, "Missing required credentials"),
			))
			return
		}
		h.logger.Error("Failed to update OAuth provider", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, provider)
}

// DeleteProviderAdmin deletes an OAuth provider by ID
func (h *AppOAuthProviderHandler) DeleteProviderAdmin(c *gin.Context) {
	providerID, err := uuid.Parse(c.Param("providerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid provider ID"),
		))
		return
	}

	if err := h.service.Delete(c.Request.Context(), providerID); err != nil {
		if err == service.ErrOAuthProviderNotFound {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				models.NewAppError(http.StatusNotFound, "OAuth provider not found"),
			))
			return
		}
		h.logger.Error("Failed to delete OAuth provider", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "OAuth provider deleted successfully"})
}
