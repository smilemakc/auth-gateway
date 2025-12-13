package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	Name            string          `json:"name" db:"name" binding:"required,max=100"`
	URL             string          `json:"url" db:"url" binding:"required,url,max=500"`
	SecretKey       string          `json:"secret_key,omitempty" db:"secret_key"` // Only sent on creation
	Events          json.RawMessage `json:"events" db:"events"`                   // JSON array of event types
	Headers         json.RawMessage `json:"headers,omitempty" db:"headers"`       // Custom headers as JSON object
	IsActive        bool            `json:"is_active" db:"is_active"`
	RetryConfig     json.RawMessage `json:"retry_config" db:"retry_config"`
	CreatedBy       *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	LastTriggeredAt *time.Time      `json:"last_triggered_at,omitempty" db:"last_triggered_at"`
}

// WebhookWithCreator includes creator information
type WebhookWithCreator struct {
	Webhook
	CreatorUsername string `json:"creator_username,omitempty" db:"creator_username"`
	CreatorEmail    string `json:"creator_email,omitempty" db:"creator_email"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	WebhookID      uuid.UUID       `json:"webhook_id" db:"webhook_id"`
	EventType      string          `json:"event_type" db:"event_type"`
	Payload        json.RawMessage `json:"payload" db:"payload"`
	Status         string          `json:"status" db:"status"` // "pending", "success", "failed"
	HTTPStatusCode *int            `json:"http_status_code,omitempty" db:"http_status_code"`
	ResponseBody   string          `json:"response_body,omitempty" db:"response_body"`
	Attempts       int             `json:"attempts" db:"attempts"`
	NextRetryAt    *time.Time      `json:"next_retry_at,omitempty" db:"next_retry_at"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
}

// CreateWebhookRequest is the request to create a webhook
type CreateWebhookRequest struct {
	Name        string            `json:"name" binding:"required,max=100"`
	URL         string            `json:"url" binding:"required,url,max=500"`
	Events      []string          `json:"events" binding:"required,min=1"`
	Headers     map[string]string `json:"headers"`
	RetryConfig *RetryConfig      `json:"retry_config"`
}

// UpdateWebhookRequest is the request to update a webhook
type UpdateWebhookRequest struct {
	Name        string            `json:"name" binding:"max=100"`
	URL         string            `json:"url" binding:"url,max=500"`
	Events      []string          `json:"events"`
	Headers     map[string]string `json:"headers"`
	IsActive    *bool             `json:"is_active"`
	RetryConfig *RetryConfig      `json:"retry_config"`
}

// RetryConfig defines webhook retry behavior
type RetryConfig struct {
	MaxAttempts    int   `json:"max_attempts"`
	BackoffSeconds []int `json:"backoff_seconds"` // Delay in seconds for each retry attempt
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:    3,
		BackoffSeconds: []int{60, 300, 900}, // 1 min, 5 min, 15 min
	}
}

// WebhookListResponse contains paginated webhook list
type WebhookListResponse struct {
	Webhooks   []WebhookWithCreator `json:"webhooks"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	TotalPages int                  `json:"total_pages"`
}

// WebhookDeliveryListResponse contains paginated delivery list
type WebhookDeliveryListResponse struct {
	Deliveries []WebhookDelivery `json:"deliveries"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}

// TestWebhookRequest is used to test a webhook
type TestWebhookRequest struct {
	EventType string                 `json:"event_type" binding:"required"`
	Payload   map[string]interface{} `json:"payload"`
}

// WebhookEvent represents a webhook event payload
type WebhookEvent struct {
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Available webhook events
const (
	WebhookEventUserCreated       = "user.created"
	WebhookEventUserUpdated       = "user.updated"
	WebhookEventUserDeleted       = "user.deleted"
	WebhookEventUserBlocked       = "user.blocked"
	WebhookEventUserUnblocked     = "user.unblocked"
	WebhookEventUserLogin         = "user.login"
	WebhookEventUserLogout        = "user.logout"
	WebhookEventUserPasswordReset = "user.password_reset"
	WebhookEventAPIKeyCreated     = "api_key.created"
	WebhookEventAPIKeyRevoked     = "api_key.revoked"
	WebhookEventRoleCreated       = "role.created"
	WebhookEventRoleUpdated       = "role.updated"
	WebhookEventRoleDeleted       = "role.deleted"
)

// GetAvailableEvents returns all available webhook events
func GetAvailableEvents() []string {
	return []string{
		WebhookEventUserCreated,
		WebhookEventUserUpdated,
		WebhookEventUserDeleted,
		WebhookEventUserBlocked,
		WebhookEventUserUnblocked,
		WebhookEventUserLogin,
		WebhookEventUserLogout,
		WebhookEventUserPasswordReset,
		WebhookEventAPIKeyCreated,
		WebhookEventAPIKeyRevoked,
		WebhookEventRoleCreated,
		WebhookEventRoleUpdated,
		WebhookEventRoleDeleted,
	}
}
