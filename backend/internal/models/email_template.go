package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EmailTemplate represents a customizable email template
type EmailTemplate struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	Type      string          `json:"type" db:"type" binding:"required,oneof=verification password_reset welcome 2fa"`
	Name      string          `json:"name" db:"name" binding:"required,max=100"`
	Subject   string          `json:"subject" db:"subject" binding:"required,max=200"`
	HTMLBody  string          `json:"html_body" db:"html_body" binding:"required"`
	TextBody  string          `json:"text_body,omitempty" db:"text_body"`
	Variables json.RawMessage `json:"variables" db:"variables"` // Available variables as JSON array
	IsActive  bool            `json:"is_active" db:"is_active"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// EmailTemplateVersion represents a historical version of a template
type EmailTemplateVersion struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TemplateID uuid.UUID  `json:"template_id" db:"template_id"`
	Subject    string     `json:"subject" db:"subject"`
	HTMLBody   string     `json:"html_body" db:"html_body"`
	TextBody   string     `json:"text_body,omitempty" db:"text_body"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// CreateEmailTemplateRequest is the request to create an email template
type CreateEmailTemplateRequest struct {
	Type      string   `json:"type" binding:"required,oneof=verification password_reset welcome 2fa custom"`
	Name      string   `json:"name" binding:"required,max=100"`
	Subject   string   `json:"subject" binding:"required,max=200"`
	HTMLBody  string   `json:"html_body" binding:"required"`
	TextBody  string   `json:"text_body"`
	Variables []string `json:"variables"` // Available variable names
}

// UpdateEmailTemplateRequest is the request to update an email template
type UpdateEmailTemplateRequest struct {
	Name      string   `json:"name" binding:"max=100"`
	Subject   string   `json:"subject" binding:"max=200"`
	HTMLBody  string   `json:"html_body"`
	TextBody  string   `json:"text_body"`
	Variables []string `json:"variables"`
	IsActive  *bool    `json:"is_active"`
}

// PreviewEmailTemplateRequest is used to preview a template with sample data
type PreviewEmailTemplateRequest struct {
	HTMLBody  string                 `json:"html_body" binding:"required"`
	TextBody  string                 `json:"text_body"`
	Variables map[string]interface{} `json:"variables"` // Sample variable values
}

// PreviewEmailTemplateResponse returns rendered template preview
type PreviewEmailTemplateResponse struct {
	RenderedHTML string `json:"rendered_html"`
	RenderedText string `json:"rendered_text"`
}

// EmailTemplateListResponse contains paginated template list
type EmailTemplateListResponse struct {
	Templates  []EmailTemplate `json:"templates"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

// EmailTemplateVersionListResponse contains version history
type EmailTemplateVersionListResponse struct {
	Versions   []EmailTemplateVersion `json:"versions"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	PerPage    int                    `json:"per_page"`
	TotalPages int                    `json:"total_pages"`
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
