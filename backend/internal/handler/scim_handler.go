package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/service"
	"github.com/smilemakc/auth-gateway/pkg/logger"
)

// SCIMHandler handles SCIM 2.0 API requests
type SCIMHandler struct {
	scimService *service.SCIMService
	logger      *logger.Logger
}

// NewSCIMHandler creates a new SCIM handler
func NewSCIMHandler(scimService *service.SCIMService, logger *logger.Logger) *SCIMHandler {
	return &SCIMHandler{
		scimService: scimService,
		logger:      logger,
	}
}

// GetUsers handles GET /scim/v2/Users
func (h *SCIMHandler) GetUsers(c *gin.Context) {
	// Parse query parameters
	filter := c.Query("filter")
	startIndex, _ := strconv.Atoi(c.DefaultQuery("startIndex", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "100"))

	if startIndex < 1 {
		startIndex = 1
	}
	if count < 1 {
		count = 100
	}
	if count > 100 {
		count = 100
	}

	response, err := h.scimService.GetUsers(c.Request.Context(), filter, startIndex, count)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUser handles GET /scim/v2/Users/{id}
func (h *SCIMHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.scimService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /scim/v2/Users
func (h *SCIMHandler) CreateUser(c *gin.Context) {
	var scimUser models.SCIMUser
	if err := c.ShouldBindJSON(&scimUser); err != nil {
		h.handleSCIMError(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	user, err := h.scimService.CreateUser(c.Request.Context(), &scimUser)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser handles PUT /scim/v2/Users/{id}
func (h *SCIMHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var scimUser models.SCIMUser
	if err := c.ShouldBindJSON(&scimUser); err != nil {
		h.handleSCIMError(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	user, err := h.scimService.UpdateUser(c.Request.Context(), id, &scimUser)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// PatchUser handles PATCH /scim/v2/Users/{id}
func (h *SCIMHandler) PatchUser(c *gin.Context) {
	id := c.Param("id")

	var patchReq models.SCIMPatchRequest
	if err := c.ShouldBindJSON(&patchReq); err != nil {
		h.handleSCIMError(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	user, err := h.scimService.PatchUser(c.Request.Context(), id, &patchReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /scim/v2/Users/{id}
func (h *SCIMHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.scimService.DeleteUser(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetGroups handles GET /scim/v2/Groups
func (h *SCIMHandler) GetGroups(c *gin.Context) {
	filter := c.Query("filter")
	startIndex, _ := strconv.Atoi(c.DefaultQuery("startIndex", "1"))
	count, _ := strconv.Atoi(c.DefaultQuery("count", "100"))

	if startIndex < 1 {
		startIndex = 1
	}
	if count < 1 {
		count = 100
	}
	if count > 100 {
		count = 100
	}

	response, err := h.scimService.GetGroups(c.Request.Context(), filter, startIndex, count)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetGroup handles GET /scim/v2/Groups/{id}
func (h *SCIMHandler) GetGroup(c *gin.Context) {
	id := c.Param("id")

	group, err := h.scimService.GetGroup(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, group)
}

// GetServiceProviderConfig handles GET /scim/v2/ServiceProviderConfig
func (h *SCIMHandler) GetServiceProviderConfig(c *gin.Context) {
	config := h.scimService.GetServiceProviderConfig(c.Request.Context())
	c.JSON(http.StatusOK, config)
}

// GetSchemas handles GET /scim/v2/Schemas
func (h *SCIMHandler) GetSchemas(c *gin.Context) {
	schemas := h.scimService.GetSchemas(c.Request.Context())
	c.JSON(http.StatusOK, schemas)
}

// Helper methods

func (h *SCIMHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*models.AppError); ok {
		statusCode := appErr.Code
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}
		h.handleSCIMError(c, statusCode, "invalid_value", appErr.Message)
		return
	}

	h.handleSCIMError(c, http.StatusInternalServerError, "internal_error", err.Error())
}

func (h *SCIMHandler) handleSCIMError(c *gin.Context, status int, scimType, detail string) {
	errorResponse := models.SCIMError{
		Schemas:  []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		Status:   strconv.Itoa(status),
		ScimType: scimType,
		Detail:   detail,
	}
	c.JSON(status, errorResponse)
}
