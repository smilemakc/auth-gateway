package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EmailTemplate represents a customizable email template
type EmailTemplate struct {
	ID        uuid.UUID       `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Type      string          `json:"type" bun:"type" binding:"required,oneof=verification password_reset welcome 2fa"`
	Name      string          `json:"name" bun:"name" binding:"required,max=100"`
	Subject   string          `json:"subject" bun:"subject" binding:"required,max=200"`
	HTMLBody  string          `json:"html_body" bun:"html_body" binding:"required"`
	TextBody  string          `json:"text_body,omitempty" bun:"text_body"`
	Variables json.RawMessage `json:"variables" bun:"variables,type:jsonb"` // Available variables as JSON array
	IsActive  bool            `json:"is_active" bun:"is_active"`
	CreatedAt time.Time       `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time       `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

// EmailTemplateVersion represents a historical version of a template
type EmailTemplateVersion struct {
	ID         uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	TemplateID uuid.UUID  `json:"template_id" bun:"template_id,type:uuid"`
	Subject    string     `json:"subject" bun:"subject"`
	HTMLBody   string     `json:"html_body" bun:"html_body"`
	TextBody   string     `json:"text_body,omitempty" bun:"text_body"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty" bun:"created_by,type:uuid"`
	CreatedAt  time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

// CreateEmailTemplateRequest is the request to create an email template
type CreateEmailTemplateRequest struct {
	// Template type: verification, password_reset, welcome, 2fa, or custom
	Type string `json:"type" binding:"required,oneof=verification password_reset welcome 2fa custom" example:"verification"`
	// Template name (max 100 characters)
	Name string `json:"name" binding:"required,max=100" example:"Email Verification Template"`
	// Email subject line (max 200 characters)
	Subject string `json:"subject" binding:"required,max=200" example:"Verify your email address"`
	// HTML email body
	HTMLBody string `json:"html_body" binding:"required" example:"<p>Hello {{username}}, your verification code is {{code}}</p>"`
	// Plain text email body
	TextBody string `json:"text_body" example:"Hello {{username}}, your verification code is {{code}}"`
	// Available variable names for template
	Variables []string `json:"variables" example:"username,email,code,expiry_minutes"`
}

// UpdateEmailTemplateRequest is the request to update an email template
type UpdateEmailTemplateRequest struct {
	// Template name (max 100 characters)
	Name string `json:"name" binding:"max=100" example:"Updated Template Name"`
	// Email subject line (max 200 characters)
	Subject string `json:"subject" binding:"max=200" example:"Updated subject"`
	// HTML email body
	HTMLBody string `json:"html_body" example:"<p>Updated HTML content</p>"`
	// Plain text email body
	TextBody string `json:"text_body" example:"Updated text content"`
	// Available variable names for template
	Variables []string `json:"variables" example:"username,email"`
	// Whether the template is active
	IsActive *bool `json:"is_active" example:"true"`
}

// PreviewEmailTemplateRequest is used to preview a template with sample data
type PreviewEmailTemplateRequest struct {
	// HTML template to preview
	HTMLBody string `json:"html_body" binding:"required" example:"<p>Hello {{username}}</p>"`
	// Text template to preview
	TextBody string `json:"text_body" example:"Hello {{username}}"`
	// Sample variable values for preview
	Variables map[string]interface{} `json:"variables" example:"username:John Doe,email:john@example.com"`
}

// PreviewEmailTemplateResponse returns rendered template preview
type PreviewEmailTemplateResponse struct {
	// Rendered HTML content
	RenderedHTML string `json:"rendered_html" example:"<p>Hello John Doe</p>"`
	// Rendered text content
	RenderedText string `json:"rendered_text" example:"Hello John Doe"`
}

// EmailTemplateListResponse contains paginated template list
type EmailTemplateListResponse struct {
	// List of email templates
	Templates []EmailTemplate `json:"templates"`
	// Total number of templates
	Total int `json:"total" example:"10"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PerPage int `json:"per_page" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"1"`
}

// EmailTemplateVersionListResponse contains version history
type EmailTemplateVersionListResponse struct {
	// List of template versions
	Versions []EmailTemplateVersion `json:"versions"`
	// Total number of versions
	Total int `json:"total" example:"5"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PerPage int `json:"per_page" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"1"`
}

// Email template types
const (
	EmailTemplateTypeVerification  = "verification"
	EmailTemplateTypePasswordReset = "password_reset"
	EmailTemplateTypeWelcome       = "welcome"
	EmailTemplateType2FA           = "2fa"
	EmailTemplateTypeCustom        = "custom"
)

// GetDefaultTemplateVariables returns default variables for each template type
func GetDefaultTemplateVariables(templateType string) []string {
	switch templateType {
	case EmailTemplateTypeVerification:
		return []string{"username", "email", "code", "expiry_minutes"}
	case EmailTemplateTypePasswordReset:
		return []string{"username", "email", "code", "expiry_minutes"}
	case EmailTemplateTypeWelcome:
		return []string{"username", "email", "full_name"}
	case EmailTemplateType2FA:
		return []string{"username", "email", "code", "expiry_minutes"}
	default:
		return []string{}
	}
}
