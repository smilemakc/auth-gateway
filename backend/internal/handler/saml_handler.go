package handler

import (
	"encoding/base64"
	"encoding/xml"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/internal/utils"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// SAMLHandler handles SAML-related HTTP requests
type SAMLHandler struct {
	samlService *service.SAMLService
	logger      *logger.Logger
}

// NewSAMLHandler creates a new SAML handler
func NewSAMLHandler(samlService *service.SAMLService, logger *logger.Logger) *SAMLHandler {
	return &SAMLHandler{
		samlService: samlService,
		logger:      logger,
	}
}

// GetMetadata handles SAML metadata requests
// @Summary Get SAML IdP Metadata
// @Description Get SAML 2.0 Identity Provider metadata XML
// @Tags SAML
// @Produce application/xml
// @Success 200 {string} string "SAML Metadata XML"
// @Router /saml/metadata [get]
func (h *SAMLHandler) GetMetadata(c *gin.Context) {
	metadata, err := h.samlService.GetMetadata()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	c.Header("Content-Type", "application/xml")
	c.XML(http.StatusOK, metadata)
}

// SSO handles SAML SSO requests
// @Summary SAML SSO Endpoint
// @Description Handle SAML Single Sign-On requests
// @Tags SAML
// @Accept application/x-www-form-urlencoded
// @Produce text/html
// @Param SAMLRequest formData string false "SAML Request (base64 encoded)"
// @Param RelayState formData string false "Relay State"
// @Success 200 {string} string "HTML form for POST to SP"
// @Failure 400 {object} models.ErrorResponse
// @Router /saml/sso [post]
func (h *SAMLHandler) SSO(c *gin.Context) {
	// Get user from context (should be set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.NewAppError(http.StatusUnauthorized, "User not authenticated"),
		))
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.NewAppError(http.StatusInternalServerError, "Invalid user ID"),
		))
		return
	}

	// Get SAMLRequest and RelayState
	samlRequest := c.PostForm("SAMLRequest")
	relayState := c.PostForm("RelayState")

	// Decode SAMLRequest (base64)
	if samlRequest == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "SAMLRequest is required"),
		))
		return
	}

	// Parse SAMLRequest to get EntityID
	// For now, we'll get it from query parameter or form
	entityID := c.Query("entityID")
	if entityID == "" {
		entityID = c.PostForm("entityID")
	}

	if entityID == "" {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "EntityID is required"),
		))
		return
	}

	// Get SP configuration
	sp, err := h.samlService.GetSPByEntityID(c.Request.Context(), entityID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.NewAppError(http.StatusNotFound, "SAML Service Provider not found"),
		))
		return
	}

	// Create SAML assertion
	response, err := h.samlService.CreateAssertion(c.Request.Context(), userID, sp)
	if err != nil {
		h.logger.Error("Failed to create SAML assertion", map[string]interface{}{
			"user_id":   userID,
			"entity_id": entityID,
			"error":     err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	// Serialize response to XML
	responseXML, err := xml.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal SAML response", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	// Encode response as base64
	responseBase64 := base64.StdEncoding.EncodeToString(responseXML)

	// Create HTML form for POST to SP
	html := h.createPOSTForm(sp.ACSURL, responseBase64, relayState)
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// SLO handles SAML Single Logout requests
// @Summary SAML SLO Endpoint
// @Description Handle SAML Single Logout requests
// @Tags SAML
// @Accept application/x-www-form-urlencoded
// @Produce text/html
// @Param SAMLRequest formData string false "SAML Logout Request"
// @Param RelayState formData string false "Relay State"
// @Success 200 {string} string "HTML form for POST to SP"
// @Router /saml/slo [post]
func (h *SAMLHandler) SLO(c *gin.Context) {
	// TODO: Implement Single Logout
	c.JSON(http.StatusNotImplemented, models.NewErrorResponse(
		models.NewAppError(http.StatusNotImplemented, "Single Logout not yet implemented"),
	))
}

// CreateSP handles SAML SP creation
// @Summary Create SAML Service Provider
// @Description Create a new SAML Service Provider configuration
// @Tags Admin - SAML
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateSAMLSPRequest true "SAML SP configuration"
// @Success 201 {object} models.SAMLServiceProvider
// @Failure 400 {object} models.ErrorResponse
// @Router /api/admin/saml/sp [post]
func (h *SAMLHandler) CreateSP(c *gin.Context) {
	var req models.CreateSAMLSPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	sp, err := h.samlService.CreateSP(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusCreated, sp)
}

// GetSP handles retrieving SAML SP
// @Summary Get SAML Service Provider
// @Description Get SAML SP by ID
// @Tags Admin - SAML
// @Security BearerAuth
// @Produce json
// @Param id path string true "SP ID (UUID)"
// @Success 200 {object} models.SAMLServiceProvider
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/saml/sp/{id} [get]
func (h *SAMLHandler) GetSP(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	sp, err := h.samlService.GetSP(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, sp)
}

// ListSPs handles listing all SAML SPs
// @Summary List SAML Service Providers
// @Description Get list of all SAML SPs
// @Tags Admin - SAML
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.ListSAMLSPsResponse
// @Router /api/admin/saml/sp [get]
func (h *SAMLHandler) ListSPs(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)

	sps, total, err := h.samlService.ListSPs(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		return
	}

	responses := make([]models.SAMLSPResponse, len(sps))
	for i, sp := range sps {
		responses[i] = models.SAMLSPResponse{
			ID:        sp.ID,
			Name:      sp.Name,
			EntityID:  sp.EntityID,
			ACSURL:    sp.ACSURL,
			SLOURL:    sp.SLOURL,
			IsActive:  sp.IsActive,
			CreatedAt: sp.CreatedAt,
			UpdatedAt: sp.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, models.ListSAMLSPsResponse{
		SPs:      responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// UpdateSP handles updating SAML SP
// @Summary Update SAML Service Provider
// @Description Update an existing SAML SP
// @Tags Admin - SAML
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "SP ID (UUID)"
// @Param request body models.UpdateSAMLSPRequest true "SAML SP update data"
// @Success 200 {object} models.SAMLServiceProvider
// @Failure 400 {object} models.ErrorResponse
// @Router /api/admin/saml/sp/{id} [put]
func (h *SAMLHandler) UpdateSP(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.UpdateSAMLSPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.NewAppError(http.StatusBadRequest, "Invalid request", err.Error()),
		))
		return
	}

	sp, err := h.samlService.UpdateSP(c.Request.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, sp)
}

// DeleteSP handles deleting SAML SP
// @Summary Delete SAML Service Provider
// @Description Delete a SAML SP
// @Tags Admin - SAML
// @Security BearerAuth
// @Param id path string true "SP ID (UUID)"
// @Success 204 "No Content"
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/saml/sp/{id} [delete]
func (h *SAMLHandler) DeleteSP(c *gin.Context) {
	id, ok := utils.ParseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.samlService.DeleteSP(c.Request.Context(), id); err != nil {
		if appErr, ok := err.(*models.AppError); ok {
			c.JSON(appErr.Code, models.NewErrorResponse(appErr))
		} else {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(err))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper methods

func (h *SAMLHandler) createPOSTForm(acsURL, samlResponse, relayState string) string {
	relayStateEscaped := url.QueryEscape(relayState)
	return `<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <title>SAML SSO</title>
</head>
<body onload="document.forms[0].submit()">
    <form method="post" action="` + acsURL + `">
        <input type="hidden" name="SAMLResponse" value="` + samlResponse + `">
        <input type="hidden" name="RelayState" value="` + relayStateEscaped + `">
        <noscript>
            <input type="submit" value="Continue">
        </noscript>
    </form>
</body>
</html>`
}
