package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/smilemakc/auth-gateway/internal/repository"
)

// TemplateService handles email template operations
type TemplateService struct {
	repo         *repository.TemplateRepository
	auditService *AuditService
}

// NewTemplateService creates a new template service
func NewTemplateService(repo *repository.TemplateRepository, auditService *AuditService) *TemplateService {
	return &TemplateService{
		repo:         repo,
		auditService: auditService,
	}
}

// CreateEmailTemplate creates a new email template
func (s *TemplateService) CreateEmailTemplate(ctx context.Context, req *models.CreateEmailTemplateRequest, createdBy uuid.UUID) (*models.EmailTemplate, error) {
	// Validate template syntax
	if err := s.validateTemplateSyntax(req.HTMLBody); err != nil {
		return nil, fmt.Errorf("invalid HTML template: %w", err)
	}
	if req.TextBody != "" {
		if err := s.validateTemplateSyntax(req.TextBody); err != nil {
			return nil, fmt.Errorf("invalid text template: %w", err)
		}
	}

	// Set default variables if not provided
	variables := req.Variables
	if len(variables) == 0 {
		variables = models.GetDefaultTemplateVariables(req.Type)
	}
	variablesJSON, _ := json.Marshal(variables)

	emailTemplate := &models.EmailTemplate{
		Type:      req.Type,
		Name:      req.Name,
		Subject:   req.Subject,
		HTMLBody:  req.HTMLBody,
		TextBody:  req.TextBody,
		Variables: variablesJSON,
		IsActive:  true,
	}

	if err := s.repo.CreateEmailTemplate(ctx, emailTemplate); err != nil {
		return nil, err
	}

	// Log audit
	s.auditService.Log(AuditLogParams{
		UserID: &createdBy,
		Action: models.ActionCreate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailTemplate,
			"template_id":   emailTemplate.ID.String(),
			"template_type": emailTemplate.Type,
			"template_name": emailTemplate.Name,
		},
	})

	return emailTemplate, nil
}

// GetEmailTemplate retrieves a template by ID
func (s *TemplateService) GetEmailTemplate(ctx context.Context, id uuid.UUID) (*models.EmailTemplate, error) {
	return s.repo.GetEmailTemplateByID(ctx, id)
}

// GetEmailTemplateByType retrieves a template by type
func (s *TemplateService) GetEmailTemplateByType(ctx context.Context, templateType string) (*models.EmailTemplate, error) {
	return s.repo.GetEmailTemplateByType(ctx, templateType)
}

// ListEmailTemplates lists all email templates
func (s *TemplateService) ListEmailTemplates(ctx context.Context) ([]models.EmailTemplate, error) {
	return s.repo.ListEmailTemplates(ctx)
}

// UpdateEmailTemplate updates an email template
func (s *TemplateService) UpdateEmailTemplate(ctx context.Context, id uuid.UUID, req *models.UpdateEmailTemplateRequest, updatedBy uuid.UUID) error {
	// Get existing template
	existingTemplate, err := s.repo.GetEmailTemplateByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates
	name := existingTemplate.Name
	if req.Name != "" {
		name = req.Name
	}

	subject := existingTemplate.Subject
	if req.Subject != "" {
		subject = req.Subject
	}

	htmlBody := existingTemplate.HTMLBody
	if req.HTMLBody != "" {
		if err := s.validateTemplateSyntax(req.HTMLBody); err != nil {
			return fmt.Errorf("invalid HTML template: %w", err)
		}
		htmlBody = req.HTMLBody
	}

	textBody := existingTemplate.TextBody
	if req.TextBody != "" {
		if err := s.validateTemplateSyntax(req.TextBody); err != nil {
			return fmt.Errorf("invalid text template: %w", err)
		}
		textBody = req.TextBody
	}

	var variablesJSON interface{} = existingTemplate.Variables
	if len(req.Variables) > 0 {
		variablesJSON, _ = json.Marshal(req.Variables)
	}

	isActive := existingTemplate.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create version history before update
	if err := s.repo.CreateTemplateVersion(ctx, id, existingTemplate.Subject, existingTemplate.HTMLBody, existingTemplate.TextBody, &updatedBy); err != nil {
		// Log but don't fail
		fmt.Printf("Failed to create template version: %v\n", err)
	}

	if err := s.repo.UpdateEmailTemplate(ctx, id, name, subject, htmlBody, textBody, variablesJSON, isActive); err != nil {
		return err
	}

	// Log audit
	s.auditService.Log(AuditLogParams{
		UserID: &updatedBy,
		Action: models.ActionUpdate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailTemplate,
			"template_id":   id.String(),
			"template_name": name,
		},
	})

	return nil
}

// DeleteEmailTemplate deletes an email template
func (s *TemplateService) DeleteEmailTemplate(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	emailTemplate, err := s.repo.GetEmailTemplateByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteEmailTemplate(ctx, id); err != nil {
		return err
	}

	// Log audit
	s.auditService.Log(AuditLogParams{
		UserID: &deletedBy,
		Action: models.ActionDelete,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type": models.ResourceEmailTemplate,
			"template_id":   id.String(),
			"template_type": emailTemplate.Type,
			"template_name": emailTemplate.Name,
		},
	})

	return nil
}

// PreviewEmailTemplate renders a template with sample data
func (s *TemplateService) PreviewEmailTemplate(ctx context.Context, req *models.PreviewEmailTemplateRequest) (*models.PreviewEmailTemplateResponse, error) {
	// Validate template syntax
	if err := s.validateTemplateSyntax(req.HTMLBody); err != nil {
		return nil, fmt.Errorf("invalid HTML template: %w", err)
	}

	// Render HTML
	renderedHTML, err := s.renderTemplate(req.HTMLBody, req.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to render HTML template: %w", err)
	}

	// Render text if provided
	var renderedText string
	if req.TextBody != "" {
		if err := s.validateTemplateSyntax(req.TextBody); err != nil {
			return nil, fmt.Errorf("invalid text template: %w", err)
		}
		renderedText, err = s.renderTemplate(req.TextBody, req.Variables)
		if err != nil {
			return nil, fmt.Errorf("failed to render text template: %w", err)
		}
	}

	return &models.PreviewEmailTemplateResponse{
		RenderedHTML: renderedHTML,
		RenderedText: renderedText,
	}, nil
}

// RenderTemplate renders a template by type with given variables
func (s *TemplateService) RenderTemplate(ctx context.Context, templateType string, variables map[string]interface{}) (subject, htmlBody, textBody string, err error) {
	emailTemplate, err := s.repo.GetEmailTemplateByType(ctx, templateType)
	if err != nil {
		return "", "", "", err
	}

	// Render subject
	subject, err = s.renderTemplate(emailTemplate.Subject, variables)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render subject: %w", err)
	}

	// Render HTML body
	htmlBody, err = s.renderTemplate(emailTemplate.HTMLBody, variables)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render HTML body: %w", err)
	}

	// Render text body if available
	if emailTemplate.TextBody != "" {
		textBody, err = s.renderTemplate(emailTemplate.TextBody, variables)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render text body: %w", err)
		}
	}

	return subject, htmlBody, textBody, nil
}

// validateTemplateSyntax validates Go template syntax
func (s *TemplateService) validateTemplateSyntax(templateStr string) error {
	_, err := template.New("validation").Parse(templateStr)
	return err
}

// renderTemplate renders a Go template with given variables
func (s *TemplateService) renderTemplate(templateStr string, variables map[string]interface{}) (string, error) {
	// Convert {{variable}} to {{.variable}} for Go templates
	converted := s.convertMustacheToGo(templateStr)

	tmpl, err := template.New("render").Parse(converted)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// convertMustacheToGo converts {{variable}} to {{.variable}}
func (s *TemplateService) convertMustacheToGo(templateStr string) string {
	// Simple conversion for {{variable}} to {{.variable}}
	// This handles common cases but more complex templates may need adjustments
	result := templateStr

	// Replace {{variable}} with {{.variable}} where variable doesn't start with .
	for {
		start := strings.Index(result, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}}")
		if end == -1 {
			break
		}
		end += start

		inner := strings.TrimSpace(result[start+2 : end])
		if inner != "" && !strings.HasPrefix(inner, ".") && !strings.HasPrefix(inner, "if") && !strings.HasPrefix(inner, "range") && !strings.HasPrefix(inner, "end") && !strings.HasPrefix(inner, "else") {
			result = result[:start+2] + " ." + inner + " " + result[end:]
		}
		// Move past this occurrence to avoid infinite loop
		if start+end+2 < len(result) {
			result = result[:end+2] + result[end+2:]
		}
		break // Process one at a time to avoid index issues
	}

	return result
}

// GetAvailableTemplateTypes returns all available template types
func (s *TemplateService) GetAvailableTemplateTypes() []string {
	return []string{
		models.EmailTemplateTypeVerification,
		models.EmailTemplateTypePasswordReset,
		models.EmailTemplateTypeWelcome,
		models.EmailTemplateType2FA,
		models.EmailTemplateTypeCustom,
	}
}

// GetDefaultVariables returns default variables for a template type
func (s *TemplateService) GetDefaultVariables(templateType string) []string {
	return models.GetDefaultTemplateVariables(templateType)
}
