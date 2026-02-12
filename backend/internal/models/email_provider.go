package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailProvider represents an email provider configuration in the database
type EmailProvider struct {
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	ApplicationID *uuid.UUID `bun:"application_id,type:uuid" json:"application_id,omitempty"`
	Name          string     `json:"name" bun:"name"`
	Type      string     `json:"type" bun:"type"`
	IsActive  bool       `json:"is_active" bun:"is_active"`
	CreatedAt time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty" bun:"created_by,type:uuid"`

	// SMTP fields
	SMTPHost     *string `json:"smtp_host,omitempty" bun:"smtp_host"`
	SMTPPort     *int    `json:"smtp_port,omitempty" bun:"smtp_port"`
	SMTPUsername *string `json:"smtp_username,omitempty" bun:"smtp_username"`
	SMTPPassword *string `json:"-" bun:"smtp_password"`
	SMTPUseTLS   *bool   `json:"smtp_use_tls,omitempty" bun:"smtp_use_tls"`

	// SendGrid fields
	SendGridAPIKey *string `json:"-" bun:"sendgrid_api_key"`

	// AWS SES fields
	SESRegion          *string `json:"ses_region,omitempty" bun:"ses_region"`
	SESAccessKeyID     *string `json:"ses_access_key_id,omitempty" bun:"ses_access_key_id"`
	SESSecretAccessKey *string `json:"-" bun:"ses_secret_access_key"`

	// Mailgun fields
	MailgunDomain *string `json:"mailgun_domain,omitempty" bun:"mailgun_domain"`
	MailgunAPIKey *string `json:"-" bun:"mailgun_api_key"`
}

// EmailProfile represents an email sending profile with from address configuration
type EmailProfile struct {
	ID            uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	ApplicationID *uuid.UUID `bun:"application_id,type:uuid" json:"application_id,omitempty"`
	Name          string     `json:"name" bun:"name"`
	ProviderID uuid.UUID  `json:"provider_id" bun:"provider_id,type:uuid"`
	FromEmail  string     `json:"from_email" bun:"from_email"`
	FromName   string     `json:"from_name" bun:"from_name"`
	ReplyTo    *string    `json:"reply_to,omitempty" bun:"reply_to"`
	IsDefault  bool       `json:"is_default" bun:"is_default"`
	IsActive   bool       `json:"is_active" bun:"is_active"`
	CreatedAt  time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt  time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty" bun:"created_by,type:uuid"`

	// Relations
	Provider *EmailProvider `json:"provider,omitempty" bun:"rel:belongs-to,join:provider_id=id"`
}

// EmailProfileTemplate represents the relationship between email profiles and templates
type EmailProfileTemplate struct {
	ID         uuid.UUID  `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	ProfileID  uuid.UUID  `json:"profile_id" bun:"profile_id,type:uuid"`
	OTPType    OTPType    `json:"otp_type" bun:"otp_type"`
	TemplateID uuid.UUID  `json:"template_id" bun:"template_id,type:uuid"`
	CreatedAt  time.Time  `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt  time.Time  `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty" bun:"created_by,type:uuid"`

	// Relations
	Profile  *EmailProfile  `json:"profile,omitempty" bun:"rel:belongs-to,join:profile_id=id"`
	Template *EmailTemplate `json:"template,omitempty" bun:"rel:belongs-to,join:template_id=id"`
}

// EmailLog represents a log of sent email messages
type EmailLog struct {
	ID             uuid.UUID   `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	ProfileID      uuid.UUID   `json:"profile_id" bun:"profile_id,type:uuid"`
	RecipientEmail string      `json:"recipient_email" bun:"recipient_email"`
	Subject        string      `json:"subject" bun:"subject"`
	TemplateType   string      `json:"template_type" bun:"template_type"`
	ProviderType   string      `json:"provider_type" bun:"provider_type"`
	MessageID      *string     `json:"message_id,omitempty" bun:"message_id"`
	Status         EmailStatus `json:"status" bun:"status"`
	ErrorMessage   *string     `json:"error_message,omitempty" bun:"error_message"`
	SentAt         *time.Time  `json:"sent_at,omitempty" bun:"sent_at"`
	CreatedAt      time.Time   `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UserID         *uuid.UUID  `json:"user_id,omitempty" bun:"user_id,type:uuid"`
	IPAddress      *string     `json:"ip_address,omitempty" bun:"ip_address"`

	// Relations
	Profile *EmailProfile `json:"profile,omitempty" bun:"rel:belongs-to,join:profile_id=id"`
}

// EmailStatus represents the status of an email message
type EmailStatus string

const (
	EmailStatusPending   EmailStatus = "pending"
	EmailStatusSent      EmailStatus = "sent"
	EmailStatusFailed    EmailStatus = "failed"
	EmailStatusDelivered EmailStatus = "delivered"
	EmailStatusBounced   EmailStatus = "bounced"
)

// CreateEmailProviderRequest represents the request to create an email provider
type CreateEmailProviderRequest struct {
	// Provider name (max 100 characters)
	Name string `json:"name" binding:"required,max=100" example:"Primary SMTP Server"`
	// Provider type: smtp, sendgrid, ses, or mailgun
	Type string `json:"type" binding:"required,oneof=smtp sendgrid ses mailgun" example:"smtp"`
	// Whether the provider is active
	IsActive bool `json:"is_active" example:"true"`

	// SMTP configuration (required for type=smtp)
	SMTPHost     *string `json:"smtp_host,omitempty" example:"smtp.gmail.com"`
	SMTPPort     *int    `json:"smtp_port,omitempty" example:"587"`
	SMTPUsername *string `json:"smtp_username,omitempty" example:"noreply@example.com"`
	SMTPPassword *string `json:"smtp_password,omitempty" example:"password123"`
	SMTPUseTLS   *bool   `json:"smtp_use_tls,omitempty" example:"true"`

	// SendGrid configuration (required for type=sendgrid)
	SendGridAPIKey *string `json:"sendgrid_api_key,omitempty" example:"SG.xxxxxxxxxxxxx"`

	// AWS SES configuration (required for type=ses)
	SESRegion          *string `json:"ses_region,omitempty" example:"us-east-1"`
	SESAccessKeyID     *string `json:"ses_access_key_id,omitempty" example:"AKIAIOSFODNN7EXAMPLE"`
	SESSecretAccessKey *string `json:"ses_secret_access_key,omitempty" example:"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"`

	// Mailgun configuration (required for type=mailgun)
	MailgunDomain *string `json:"mailgun_domain,omitempty" example:"mg.example.com"`
	MailgunAPIKey *string `json:"mailgun_api_key,omitempty" example:"key-xxxxxxxxxxxxxxxxx"`
}

// UpdateEmailProviderRequest represents the request to update an email provider
type UpdateEmailProviderRequest struct {
	// Provider name (max 100 characters)
	Name *string `json:"name,omitempty" binding:"omitempty,max=100" example:"Updated SMTP Server"`
	// Whether the provider is active
	IsActive *bool `json:"is_active,omitempty" example:"false"`

	// SMTP configuration
	SMTPHost     *string `json:"smtp_host,omitempty" example:"smtp.sendgrid.net"`
	SMTPPort     *int    `json:"smtp_port,omitempty" example:"465"`
	SMTPUsername *string `json:"smtp_username,omitempty" example:"apikey"`
	SMTPPassword *string `json:"smtp_password,omitempty" example:"new_password"`
	SMTPUseTLS   *bool   `json:"smtp_use_tls,omitempty" example:"false"`

	// SendGrid configuration
	SendGridAPIKey *string `json:"sendgrid_api_key,omitempty" example:"SG.new_api_key"`

	// AWS SES configuration
	SESRegion          *string `json:"ses_region,omitempty" example:"eu-west-1"`
	SESAccessKeyID     *string `json:"ses_access_key_id,omitempty" example:"AKIAIOSFODNN7NEWKEY"`
	SESSecretAccessKey *string `json:"ses_secret_access_key,omitempty" example:"new_secret_key"`

	// Mailgun configuration
	MailgunDomain *string `json:"mailgun_domain,omitempty" example:"mg.newdomain.com"`
	MailgunAPIKey *string `json:"mailgun_api_key,omitempty" example:"key-new_api_key"`
}

// CreateEmailProfileRequest represents the request to create an email profile
type CreateEmailProfileRequest struct {
	// Profile name (max 100 characters)
	Name string `json:"name" binding:"required,max=100" example:"Marketing Emails"`
	// Email provider ID
	ProviderID uuid.UUID `json:"provider_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// From email address
	FromEmail string `json:"from_email" binding:"required,email" example:"noreply@example.com"`
	// From name (sender name)
	FromName string `json:"from_name" binding:"required,max=100" example:"Example Support"`
	// Reply-to email address
	ReplyTo *string `json:"reply_to,omitempty" binding:"omitempty,email" example:"support@example.com"`
	// Whether this is the default profile
	IsDefault bool `json:"is_default" example:"false"`
	// Whether the profile is active
	IsActive bool `json:"is_active" example:"true"`
}

// UpdateEmailProfileRequest represents the request to update an email profile
type UpdateEmailProfileRequest struct {
	// Profile name (max 100 characters)
	Name *string `json:"name,omitempty" binding:"omitempty,max=100" example:"Updated Profile"`
	// Email provider ID
	ProviderID *uuid.UUID `json:"provider_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
	// From email address
	FromEmail *string `json:"from_email,omitempty" binding:"omitempty,email" example:"updated@example.com"`
	// From name (sender name)
	FromName *string `json:"from_name,omitempty" binding:"omitempty,max=100" example:"Updated Support"`
	// Reply-to email address
	ReplyTo *string `json:"reply_to,omitempty" binding:"omitempty,email" example:"newreply@example.com"`
	// Whether this is the default profile
	IsDefault *bool `json:"is_default,omitempty" example:"true"`
	// Whether the profile is active
	IsActive *bool `json:"is_active,omitempty" example:"false"`
}

// EmailProviderResponse represents a single email provider response (with secrets masked)
type EmailProviderResponse struct {
	ID        uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string     `json:"name" example:"Primary SMTP"`
	Type      string     `json:"type" example:"smtp"`
	IsActive  bool       `json:"is_active" example:"true"`
	CreatedAt time.Time  `json:"created_at" example:"2024-01-15T10:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2024-01-15T10:00:00Z"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`

	// Non-secret configuration (secrets are masked)
	SMTPHost       *string `json:"smtp_host,omitempty" example:"smtp.gmail.com"`
	SMTPPort       *int    `json:"smtp_port,omitempty" example:"587"`
	SMTPUsername   *string `json:"smtp_username,omitempty" example:"noreply@example.com"`
	SMTPUseTLS     *bool   `json:"smtp_use_tls,omitempty" example:"true"`
	SESRegion      *string `json:"ses_region,omitempty" example:"us-east-1"`
	SESAccessKeyID *string `json:"ses_access_key_id,omitempty" example:"AKIAIOSFODNN7EXAMPLE"`
	MailgunDomain  *string `json:"mailgun_domain,omitempty" example:"mg.example.com"`

	// Indicate if secrets are configured (without exposing them)
	HasSMTPPassword       bool `json:"has_smtp_password" example:"true"`
	HasSendGridAPIKey     bool `json:"has_sendgrid_api_key" example:"false"`
	HasSESSecretAccessKey bool `json:"has_ses_secret_access_key" example:"false"`
	HasMailgunAPIKey      bool `json:"has_mailgun_api_key" example:"false"`
}

// EmailProviderListResponse contains paginated provider list
type EmailProviderListResponse struct {
	// List of email providers
	Providers []EmailProviderResponse `json:"providers"`
	// Total number of providers
	Total int `json:"total" example:"5"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PerPage int `json:"per_page" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"1"`
}

// EmailProfileListResponse contains paginated profile list
type EmailProfileListResponse struct {
	// List of email profiles
	Profiles []EmailProfile `json:"profiles"`
	// Total number of profiles
	Total int `json:"total" example:"3"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PerPage int `json:"per_page" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"1"`
}

// EmailLogListResponse contains paginated email log list
type EmailLogListResponse struct {
	// List of email logs
	Logs []EmailLog `json:"logs"`
	// Total number of logs
	Total int `json:"total" example:"150"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PerPage int `json:"per_page" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"8"`
}

// EmailStatsResponse represents email statistics
type EmailStatsResponse struct {
	// Total emails sent
	TotalSent int64 `json:"total_sent" example:"1250"`
	// Total emails failed
	TotalFailed int64 `json:"total_failed" example:"25"`
	// Total emails delivered
	TotalDelivered int64 `json:"total_delivered" example:"1200"`
	// Total emails bounced
	TotalBounced int64 `json:"total_bounced" example:"25"`
	// Emails sent today
	SentToday int64 `json:"sent_today" example:"45"`
	// Emails sent this hour
	SentThisHour int64 `json:"sent_this_hour" example:"8"`
	// Statistics by template type
	ByTemplateType map[string]int64 `json:"by_template_type" example:"verification:500,password_reset:300,welcome:450"`
	// Statistics by status
	ByStatus map[string]int64 `json:"by_status" example:"sent:1225,failed:25"`
	// Statistics by provider
	ByProvider map[string]int64 `json:"by_provider" example:"smtp:800,sendgrid:450"`
	// Recent email logs (last 10)
	RecentMessages []EmailLog `json:"recent_messages"`
}

// SendEmailRequest represents the request to send a test email
type SendEmailRequest struct {
	// Profile ID to use for sending
	ProfileID uuid.UUID `json:"profile_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Recipient email address
	RecipientEmail string `json:"recipient_email" binding:"required,email" example:"test@example.com"`
	// Email subject
	Subject string `json:"subject" binding:"required,max=200" example:"Test Email"`
	// HTML body
	HTMLBody string `json:"html_body" binding:"required" example:"<p>This is a test email</p>"`
	// Text body
	TextBody *string `json:"text_body,omitempty" example:"This is a test email"`
}

// SendEmailResponse represents the response after sending an email
type SendEmailResponse struct {
	// Whether email was sent successfully
	Success bool `json:"success" example:"true"`
	// Message ID from email provider
	MessageID *string `json:"message_id,omitempty" example:"<1234567890.abcdef@example.com>"`
	// Log ID for tracking
	LogID uuid.UUID `json:"log_id" example:"550e8400-e29b-41d4-a716-446655440003"`
}

// AssignTemplateToProfileRequest assigns a template to a profile for specific OTP type
type AssignTemplateToProfileRequest struct {
	// Email profile ID
	ProfileID uuid.UUID `json:"profile_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	// OTP type: verification, password_reset, 2fa, login, or registration
	OTPType OTPType `json:"otp_type" binding:"required,oneof=verification password_reset 2fa login registration" example:"verification"`
	// Email template ID
	TemplateID uuid.UUID `json:"template_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
}

// Email provider types
const (
	EmailProviderTypeSMTP     = "smtp"
	EmailProviderTypeSendGrid = "sendgrid"
	EmailProviderTypeSES      = "ses"
	EmailProviderTypeMailgun  = "mailgun"
)
