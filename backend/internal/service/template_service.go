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

// ListEmailTemplates lists all global email templates
func (s *TemplateService) ListEmailTemplates(ctx context.Context) ([]models.EmailTemplate, error) {
	return s.repo.ListEmailTemplates(ctx, nil)
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
		models.EmailTemplateTypeOTPLogin,
		models.EmailTemplateTypeOTPRegistration,
		models.EmailTemplateTypePasswordChanged,
		models.EmailTemplateTypeLoginAlert,
		models.EmailTemplateType2FAEnabled,
		models.EmailTemplateType2FADisabled,
		models.EmailTemplateTypeCustom,
	}
}

// GetDefaultVariables returns default variables for a template type
func (s *TemplateService) GetDefaultVariables(templateType string) []string {
	return models.GetDefaultTemplateVariables(templateType)
}

// CreateEmailTemplateForApp creates a template for a specific application
func (s *TemplateService) CreateEmailTemplateForApp(ctx context.Context, applicationID uuid.UUID, req *models.CreateEmailTemplateRequest, createdBy uuid.UUID) (*models.EmailTemplate, error) {
	if err := s.validateTemplateSyntax(req.HTMLBody); err != nil {
		return nil, fmt.Errorf("invalid HTML template: %w", err)
	}
	if req.TextBody != "" {
		if err := s.validateTemplateSyntax(req.TextBody); err != nil {
			return nil, fmt.Errorf("invalid text template: %w", err)
		}
	}

	variables := req.Variables
	if len(variables) == 0 {
		variables = models.GetDefaultTemplateVariables(req.Type)
	}
	variablesJSON, _ := json.Marshal(variables)

	emailTemplate := &models.EmailTemplate{
		Type:          req.Type,
		Name:          req.Name,
		Subject:       req.Subject,
		HTMLBody:      req.HTMLBody,
		TextBody:      req.TextBody,
		Variables:     variablesJSON,
		IsActive:      true,
		ApplicationID: &applicationID,
	}

	if err := s.repo.CreateEmailTemplate(ctx, emailTemplate); err != nil {
		return nil, err
	}

	s.auditService.Log(AuditLogParams{
		UserID: &createdBy,
		Action: models.ActionCreate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type":  models.ResourceEmailTemplate,
			"template_id":    emailTemplate.ID.String(),
			"template_type":  emailTemplate.Type,
			"template_name":  emailTemplate.Name,
			"application_id": applicationID.String(),
		},
	})

	return emailTemplate, nil
}

// ListEmailTemplatesForApp lists templates for a specific application
func (s *TemplateService) ListEmailTemplatesForApp(ctx context.Context, applicationID uuid.UUID) ([]models.EmailTemplate, error) {
	return s.repo.ListEmailTemplatesByApplication(ctx, applicationID)
}

// GetEmailTemplateByTypeAndApp retrieves a template by type for a specific application
func (s *TemplateService) GetEmailTemplateByTypeAndApp(ctx context.Context, templateType string, applicationID uuid.UUID) (*models.EmailTemplate, error) {
	return s.repo.GetEmailTemplateByTypeAndApp(ctx, templateType, applicationID)
}

// UpdateEmailTemplateForApp updates an application-scoped template
func (s *TemplateService) UpdateEmailTemplateForApp(ctx context.Context, applicationID, templateID uuid.UUID, req *models.UpdateEmailTemplateRequest, updatedBy uuid.UUID) error {
	existingTemplate, err := s.repo.GetEmailTemplateByID(ctx, templateID)
	if err != nil {
		return err
	}

	if existingTemplate.ApplicationID == nil || *existingTemplate.ApplicationID != applicationID {
		return fmt.Errorf("template does not belong to application")
	}

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

	if err := s.repo.CreateTemplateVersionWithApp(ctx, templateID, &applicationID, existingTemplate.Subject, existingTemplate.HTMLBody, existingTemplate.TextBody, &updatedBy); err != nil {
		fmt.Printf("Failed to create template version: %v\n", err)
	}

	if err := s.repo.UpdateEmailTemplate(ctx, templateID, name, subject, htmlBody, textBody, variablesJSON, isActive); err != nil {
		return err
	}

	s.auditService.Log(AuditLogParams{
		UserID: &updatedBy,
		Action: models.ActionUpdate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type":  models.ResourceEmailTemplate,
			"template_id":    templateID.String(),
			"template_name":  name,
			"application_id": applicationID.String(),
		},
	})

	return nil
}

// RenderTemplateForApp renders a template by type for a specific application
func (s *TemplateService) RenderTemplateForApp(ctx context.Context, templateType string, applicationID uuid.UUID, variables map[string]interface{}) (subject, htmlBody, textBody string, err error) {
	emailTemplate, err := s.repo.GetEmailTemplateByTypeAndApp(ctx, templateType, applicationID)
	if err != nil {
		emailTemplate, err = s.repo.GetEmailTemplateByType(ctx, templateType)
		if err != nil {
			return "", "", "", err
		}
	}

	subject, err = s.renderTemplate(emailTemplate.Subject, variables)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render subject: %w", err)
	}

	htmlBody, err = s.renderTemplate(emailTemplate.HTMLBody, variables)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render HTML body: %w", err)
	}

	if emailTemplate.TextBody != "" {
		textBody, err = s.renderTemplate(emailTemplate.TextBody, variables)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to render text body: %w", err)
		}
	}

	return subject, htmlBody, textBody, nil
}

// InitializeTemplatesForApp creates default templates for a new application
func (s *TemplateService) InitializeTemplatesForApp(ctx context.Context, applicationID uuid.UUID, createdBy uuid.UUID) error {
	templateTypes := []string{
		models.EmailTemplateTypeVerification,
		models.EmailTemplateTypePasswordReset,
		models.EmailTemplateTypeWelcome,
		models.EmailTemplateType2FA,
		models.EmailTemplateTypeOTPLogin,
		models.EmailTemplateTypeOTPRegistration,
		models.EmailTemplateTypePasswordChanged,
		models.EmailTemplateTypeLoginAlert,
		models.EmailTemplateType2FAEnabled,
		models.EmailTemplateType2FADisabled,
	}

	for _, templateType := range templateTypes {
		variables := models.GetDefaultTemplateVariables(templateType)
		variablesJSON, _ := json.Marshal(variables)

		subject, htmlBody, textBody := s.getDefaultTemplateContent(templateType)

		emailTemplate := &models.EmailTemplate{
			Type:          templateType,
			Name:          s.getDefaultTemplateName(templateType),
			Subject:       subject,
			HTMLBody:      htmlBody,
			TextBody:      textBody,
			Variables:     variablesJSON,
			IsActive:      true,
			ApplicationID: &applicationID,
		}

		if err := s.repo.CreateEmailTemplate(ctx, emailTemplate); err != nil {
			return fmt.Errorf("failed to create %s template: %w", templateType, err)
		}
	}

	s.auditService.Log(AuditLogParams{
		UserID: &createdBy,
		Action: models.ActionCreate,
		Status: models.StatusSuccess,
		Details: map[string]interface{}{
			"resource_type":  models.ResourceEmailTemplate,
			"action":         "initialize_templates",
			"application_id": applicationID.String(),
			"template_count": len(templateTypes),
		},
	})

	return nil
}

func (s *TemplateService) getDefaultTemplateName(templateType string) string {
	switch templateType {
	case models.EmailTemplateTypeVerification:
		return "Email Verification"
	case models.EmailTemplateTypePasswordReset:
		return "Password Reset"
	case models.EmailTemplateTypeWelcome:
		return "Welcome Email"
	case models.EmailTemplateType2FA:
		return "Two-Factor Authentication"
	case models.EmailTemplateTypeOTPLogin:
		return "OTP Login"
	case models.EmailTemplateTypeOTPRegistration:
		return "OTP Registration"
	case models.EmailTemplateTypePasswordChanged:
		return "Password Changed"
	case models.EmailTemplateTypeLoginAlert:
		return "Login Alert"
	case models.EmailTemplateType2FAEnabled:
		return "2FA Enabled"
	case models.EmailTemplateType2FADisabled:
		return "2FA Disabled"
	default:
		return "Custom Template"
	}
}

func (s *TemplateService) getDefaultTemplateContent(templateType string) (subject, htmlBody, textBody string) {
	switch templateType {
	case models.EmailTemplateTypeVerification:
		subject = "Verify Your Email Address"
		htmlBody = `<html><body><h2>Email Verification</h2><p>Hello {{.username}},</p><p>Your verification code is: <strong>{{.code}}</strong></p><p>This code will expire in {{.expiry_minutes}} minutes.</p></body></html>`
		textBody = `Email Verification\n\nHello {{.username}},\n\nYour verification code is: {{.code}}\n\nThis code will expire in {{.expiry_minutes}} minutes.`
	case models.EmailTemplateTypePasswordReset:
		subject = "Reset Your Password"
		htmlBody = `<html><body><h2>Password Reset</h2><p>Hello {{.username}},</p><p>Your password reset code is: <strong>{{.code}}</strong></p><p>This code will expire in {{.expiry_minutes}} minutes.</p></body></html>`
		textBody = `Password Reset\n\nHello {{.username}},\n\nYour password reset code is: {{.code}}\n\nThis code will expire in {{.expiry_minutes}} minutes.`
	case models.EmailTemplateTypeWelcome:
		subject = "Welcome to Our Platform"
		htmlBody = `<html><body><h2>Welcome!</h2><p>Hello {{.full_name}},</p><p>Welcome to our platform. We're excited to have you on board!</p></body></html>`
		textBody = `Welcome!\n\nHello {{.full_name}},\n\nWelcome to our platform. We're excited to have you on board!`
	case models.EmailTemplateType2FA:
		subject = "Your Two-Factor Authentication Code"
		htmlBody = `<html><body><h2>Two-Factor Authentication</h2><p>Hello {{.username}},</p><p>Your 2FA code is: <strong>{{.code}}</strong></p><p>This code will expire in {{.expiry_minutes}} minutes.</p></body></html>`
		textBody = `Two-Factor Authentication\n\nHello {{.username}},\n\nYour 2FA code is: {{.code}}\n\nThis code will expire in {{.expiry_minutes}} minutes.`
	case models.EmailTemplateTypeOTPLogin:
		subject = "Your Login Code"
		htmlBody = `<html><body><h2>Login Code</h2><p>Hello {{.username}},</p><p>Your one-time login code is: <strong>{{.code}}</strong></p><p>This code will expire in {{.expiry_minutes}} minutes.</p></body></html>`
		textBody = `Login Code\n\nHello {{.username}},\n\nYour one-time login code is: {{.code}}\n\nThis code will expire in {{.expiry_minutes}} minutes.`
	case models.EmailTemplateTypeOTPRegistration:
		subject = "Complete Your Registration"
		htmlBody = `<html><body><h2>Registration Code</h2><p>Hello {{.username}},</p><p>Your registration code is: <strong>{{.code}}</strong></p><p>This code will expire in {{.expiry_minutes}} minutes.</p></body></html>`
		textBody = `Registration Code\n\nHello {{.username}},\n\nYour registration code is: {{.code}}\n\nThis code will expire in {{.expiry_minutes}} minutes.`
	case models.EmailTemplateTypePasswordChanged:
		subject = "Your Password Has Been Changed"
		htmlBody = `<html><body><h2>Password Changed</h2><p>Hello {{.username}},</p><p>Your password was successfully changed.</p><p><strong>IP Address:</strong> {{.ip_address}}</p><p><strong>Time:</strong> {{.timestamp}}</p><p>If you did not make this change, please contact support immediately.</p></body></html>`
		textBody = `Password Changed\n\nHello {{.username}},\n\nYour password was successfully changed.\n\nIP Address: {{.ip_address}}\nTime: {{.timestamp}}\n\nIf you did not make this change, please contact support immediately.`
	case models.EmailTemplateTypeLoginAlert:
		subject = "New Login to Your Account"
		htmlBody = `<html><body><h2>New Login Detected</h2><p>Hello {{.username}},</p><p>A new login to your account was detected.</p><p><strong>IP Address:</strong> {{.ip_address}}</p><p><strong>Device:</strong> {{.device_type}}</p><p><strong>Time:</strong> {{.timestamp}}</p><p>If this wasn't you, please change your password immediately.</p></body></html>`
		textBody = `New Login Detected\n\nHello {{.username}},\n\nA new login to your account was detected.\n\nIP Address: {{.ip_address}}\nDevice: {{.device_type}}\nTime: {{.timestamp}}\n\nIf this wasn't you, please change your password immediately.`
	case models.EmailTemplateType2FAEnabled:
		subject = "Two-Factor Authentication Enabled"
		htmlBody = `<html><body><h2>2FA Enabled</h2><p>Hello {{.username}},</p><p>Two-factor authentication has been successfully enabled on your account.</p><p><strong>Time:</strong> {{.timestamp}}</p><p>Your account is now more secure. You will need your authenticator app to sign in.</p></body></html>`
		textBody = `2FA Enabled\n\nHello {{.username}},\n\nTwo-factor authentication has been successfully enabled on your account.\n\nTime: {{.timestamp}}\n\nYour account is now more secure.`
	case models.EmailTemplateType2FADisabled:
		subject = "Two-Factor Authentication Disabled"
		htmlBody = `<html><body><h2>2FA Disabled</h2><p>Hello {{.username}},</p><p>Two-factor authentication has been disabled on your account.</p><p><strong>Time:</strong> {{.timestamp}}</p><p>Your account is now less secure. We recommend re-enabling 2FA as soon as possible.</p></body></html>`
		textBody = `2FA Disabled\n\nHello {{.username}},\n\nTwo-factor authentication has been disabled on your account.\n\nTime: {{.timestamp}}\n\nWe recommend re-enabling 2FA as soon as possible.`
	default:
		subject = "Notification"
		htmlBody = `<html><body><p>Default template content</p></body></html>`
		textBody = `Default template content`
	}
	return
}
