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

// TemplateHandler handles email template endpoints
type TemplateHandler struct {
	templateService *service.TemplateService
	logger          *logger.Logger
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(templateService *service.TemplateService, log *logger.Logger) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
		logger:          log,
	}
}

// ListEmailTemplates godoc
// @Summary List all email templates
// @Description Get a list of all email templates (admin only)
// @Tags Admin - Email Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.EmailTemplateListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/templates [get]
func (h *TemplateHandler) ListEmailTemplates(c *gin.Context) {
	appID, _ := utils.GetApplicationIDFromContext(c)

	var templates []models.EmailTemplate
	var err error
	if appID != nil {
		templates, err = h.templateService.ListEmailTemplatesForApp(c.Request.Context(), *appID)
	} else {
		templates, err = h.templateService.ListEmailTemplates(c.Request.Context())
	}
	if err != nil {
		h.logger.Error("Failed to list templates", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, models.EmailTemplateListResponse{
		Templates:  templates,
		Total:      len(templates),
		Page:       1,
		PageSize:   len(templates),
		TotalPages: 1,
	})
}

// GetEmailTemplate godoc
// @Summary Get email template by ID
// @Description Get a specific email template by its ID (admin only)
// @Tags Admin - Email Templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} models.EmailTemplate
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/templates/{id} [get]
func (h *TemplateHandler) GetEmailTemplate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid template ID"})
		return
	}

	template, err := h.templateService.GetEmailTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// CreateEmailTemplate godoc
// @Summary Create a new email template
// @Description Create a new email template for customized emails (admin only)
// @Tags Admin - Email Templates
// @Accept json
// @Produce json
// @Param request body models.CreateEmailTemplateRequest true "Template data"
// @Security BearerAuth
// @Success 201 {object} models.EmailTemplate
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/templates [post]
func (h *TemplateHandler) CreateEmailTemplate(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req models.CreateEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	template, err := h.templateService.CreateEmailTemplate(c.Request.Context(), &req, *userID)
	if err != nil {
		h.logger.Error("Failed to create template", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// UpdateEmailTemplate godoc
// @Summary Update an email template
// @Description Update an existing email template (admin only)
// @Tags Admin - Email Templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID (UUID)"
// @Param request body models.UpdateEmailTemplateRequest true "Template update data"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/templates/{id} [put]
func (h *TemplateHandler) UpdateEmailTemplate(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid template ID"})
		return
	}

	var req models.UpdateEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.templateService.UpdateEmailTemplate(c.Request.Context(), id, &req, *userID); err != nil {
		h.logger.Error("Failed to update template", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Template updated successfully"})
}

// DeleteEmailTemplate godoc
// @Summary Delete an email template
// @Description Delete an email template by ID (admin only)
// @Tags Admin - Email Templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/templates/{id} [delete]
func (h *TemplateHandler) DeleteEmailTemplate(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid template ID"})
		return
	}

	if err := h.templateService.DeleteEmailTemplate(c.Request.Context(), id, *userID); err != nil {
		h.logger.Error("Failed to delete template", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Template not found"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Template deleted successfully"})
}

// PreviewEmailTemplate godoc
// @Summary Preview an email template
// @Description Render a template with sample data for preview (admin only)
// @Tags Admin - Email Templates
// @Accept json
// @Produce json
// @Param request body models.PreviewEmailTemplateRequest true "Preview data"
// @Security BearerAuth
// @Success 200 {object} models.PreviewEmailTemplateResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/templates/preview [post]
func (h *TemplateHandler) PreviewEmailTemplate(c *gin.Context) {
	var req models.PreviewEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := h.templateService.PreviewEmailTemplate(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to preview template", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetAvailableTemplateTypes godoc
// @Summary Get available template types
// @Description Get a list of all available email template types (admin only)
// @Tags Admin - Email Templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.TemplateTypesResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/templates/types [get]
func (h *TemplateHandler) GetAvailableTemplateTypes(c *gin.Context) {
	types := h.templateService.GetAvailableTemplateTypes()
	c.JSON(http.StatusOK, gin.H{"types": types})
}

// GetDefaultVariables godoc
// @Summary Get default variables for a template type
// @Description Get the list of default variables available for a specific template type (admin only)
// @Tags Admin - Email Templates
// @Accept json
// @Produce json
// @Param type path string true "Template type (e.g., welcome, password_reset, verification)"
// @Security BearerAuth
// @Success 200 {object} models.TemplateVariablesResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/admin/templates/variables/{type} [get]
func (h *TemplateHandler) GetDefaultVariables(c *gin.Context) {
	templateType := c.Param("type")
	variables := h.templateService.GetDefaultVariables(templateType)
	c.JSON(http.StatusOK, gin.H{"variables": variables})
}

// ListApplicationEmailTemplates godoc
// @Summary List email templates for an application
// @Description Get all email templates for a specific application (admin only)
// @Tags Admin - Application Email Templates
// @Accept json
// @Produce json
// @Param appId path string true "Application ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} models.EmailTemplateListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{appId}/email-templates [get]
func (h *TemplateHandler) ListApplicationEmailTemplates(c *gin.Context) {
	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid application ID"})
		return
	}

	templates, err := h.templateService.ListEmailTemplatesForApp(c.Request.Context(), appID)
	if err != nil {
		h.logger.Error("Failed to list application templates", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, models.EmailTemplateListResponse{
		Templates:  templates,
		Total:      len(templates),
		Page:       1,
		PageSize:   len(templates),
		TotalPages: 1,
	})
}

// CreateApplicationEmailTemplate godoc
// @Summary Create email template for application
// @Description Create a new email template for a specific application (admin only)
// @Tags Admin - Application Email Templates
// @Accept json
// @Produce json
// @Param appId path string true "Application ID (UUID)"
// @Param request body models.CreateEmailTemplateRequest true "Template data"
// @Security BearerAuth
// @Success 201 {object} models.EmailTemplate
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{appId}/email-templates [post]
func (h *TemplateHandler) CreateApplicationEmailTemplate(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid application ID"})
		return
	}

	var req models.CreateEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	template, err := h.templateService.CreateEmailTemplateForApp(c.Request.Context(), appID, &req, *userID)
	if err != nil {
		h.logger.Error("Failed to create application template", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetApplicationEmailTemplate godoc
// @Summary Get application email template by ID
// @Description Get a specific email template for an application by its ID (admin only)
// @Tags Admin - Application Email Templates
// @Accept json
// @Produce json
// @Param appId path string true "Application ID (UUID)"
// @Param id path string true "Template ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} models.EmailTemplate
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/applications/{appId}/email-templates/{id} [get]
func (h *TemplateHandler) GetApplicationEmailTemplate(c *gin.Context) {
	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid application ID"})
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid template ID"})
		return
	}

	template, err := h.templateService.GetEmailTemplate(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Template not found"})
		return
	}

	if template.ApplicationID == nil || *template.ApplicationID != appID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Template not found for this application"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// UpdateApplicationEmailTemplate godoc
// @Summary Update application email template
// @Description Update an existing email template for a specific application (admin only)
// @Tags Admin - Application Email Templates
// @Accept json
// @Produce json
// @Param appId path string true "Application ID (UUID)"
// @Param id path string true "Template ID (UUID)"
// @Param request body models.UpdateEmailTemplateRequest true "Template update data"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/applications/{appId}/email-templates/{id} [put]
func (h *TemplateHandler) UpdateApplicationEmailTemplate(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid application ID"})
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid template ID"})
		return
	}

	var req models.UpdateEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.templateService.UpdateEmailTemplateForApp(c.Request.Context(), appID, templateID, &req, *userID); err != nil {
		h.logger.Error("Failed to update application template", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
			"template_id":    templateID.String(),
		})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Template updated successfully"})
}

// DeleteApplicationEmailTemplate godoc
// @Summary Delete application email template
// @Description Delete an email template for a specific application by ID (admin only)
// @Tags Admin - Application Email Templates
// @Accept json
// @Produce json
// @Param appId path string true "Application ID (UUID)"
// @Param id path string true "Template ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/applications/{appId}/email-templates/{id} [delete]
func (h *TemplateHandler) DeleteApplicationEmailTemplate(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid application ID"})
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid template ID"})
		return
	}

	template, err := h.templateService.GetEmailTemplate(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Template not found"})
		return
	}

	if template.ApplicationID == nil || *template.ApplicationID != appID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Template not found for this application"})
		return
	}

	if err := h.templateService.DeleteEmailTemplate(c.Request.Context(), templateID, *userID); err != nil {
		h.logger.Error("Failed to delete application template", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
			"template_id":    templateID.String(),
		})
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Template not found"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Template deleted successfully"})
}

// InitializeApplicationTemplates godoc
// @Summary Initialize default templates for application
// @Description Create default email templates for a specific application (admin only)
// @Tags Admin - Application Email Templates
// @Accept json
// @Produce json
// @Param appId path string true "Application ID (UUID)"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/applications/{appId}/email-templates/initialize [post]
func (h *TemplateHandler) InitializeApplicationTemplates(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid application ID"})
		return
	}

	if err := h.templateService.InitializeTemplatesForApp(c.Request.Context(), appID, *userID); err != nil {
		h.logger.Error("Failed to initialize application templates", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
		})
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Default templates initialized successfully"})
}

// PreviewApplicationEmailTemplate godoc
// @Summary Preview application email template
// @Description Render a template with sample data for preview for a specific application (admin only)
// @Tags Admin - Application Email Templates
// @Accept json
// @Produce json
// @Param appId path string true "Application ID (UUID)"
// @Param id path string true "Template ID (UUID)"
// @Param request body models.PreviewEmailTemplateRequest true "Preview data"
// @Security BearerAuth
// @Success 200 {object} models.PreviewEmailTemplateResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/admin/applications/{appId}/email-templates/{id}/preview [post]
func (h *TemplateHandler) PreviewApplicationEmailTemplate(c *gin.Context) {
	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid application ID"})
		return
	}

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid template ID"})
		return
	}

	template, err := h.templateService.GetEmailTemplate(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Template not found"})
		return
	}

	if template.ApplicationID == nil || *template.ApplicationID != appID {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Template not found for this application"})
		return
	}

	var req models.PreviewEmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	req.HTMLBody = template.HTMLBody
	req.TextBody = template.TextBody

	resp, err := h.templateService.PreviewEmailTemplate(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to preview application template", map[string]interface{}{
			"error":          err.Error(),
			"application_id": appID.String(),
			"template_id":    templateID.String(),
		})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
