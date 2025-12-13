package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID              uuid.UUID       `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Name            string          `json:"name" bun:"name" binding:"required,max=100"`
	URL             string          `json:"url" bun:"url" binding:"required,url,max=500"`
	SecretKey       string          `json:"secret_key,omitempty" bun:"secret_key"`      // Only sent on creation
	Events          json.RawMessage `json:"events" bun:"events,type:jsonb"`             // JSON array of event types
	Headers         json.RawMessage `json:"headers,omitempty" bun:"headers,type:jsonb"` // Custom headers as JSON object
	IsActive        bool            `json:"is_active" bun:"is_active"`
	RetryConfig     json.RawMessage `json:"retry_config" bun:"retry_config,type:jsonb"`
	CreatedBy       *uuid.UUID      `json:"created_by,omitempty" bun:"created_by,type:uuid"`
	CreatedAt       time.Time       `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt       time.Time       `json:"updated_at" bun:"updated_at,nullzero,notnull,default:current_timestamp"`
	LastTriggeredAt *time.Time      `json:"last_triggered_at,omitempty" bun:"last_triggered_at"`
}

// WebhookWithCreator includes creator information
type WebhookWithCreator struct {
	Webhook
	CreatorUsername string `json:"creator_username,omitempty" bun:"creator_username"`
	CreatorEmail    string `json:"creator_email,omitempty" bun:"creator_email"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID             uuid.UUID       `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	WebhookID      uuid.UUID       `json:"webhook_id" bun:"webhook_id,type:uuid"`
	EventType      string          `json:"event_type" bun:"event_type"`
	Payload        json.RawMessage `json:"payload" bun:"payload,type:jsonb"`
	Status         string          `json:"status" bun:"status"` // "pending", "success", "failed"
	HTTPStatusCode *int            `json:"http_status_code,omitempty" bun:"http_status_code"`
	ResponseBody   string          `json:"response_body,omitempty" bun:"response_body"`
	Attempts       int             `json:"attempts" bun:"attempts"`
	NextRetryAt    *time.Time      `json:"next_retry_at,omitempty" bun:"next_retry_at"`
	CreatedAt      time.Time       `json:"created_at" bun:"created_at,nullzero,notnull,default:current_timestamp"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty" bun:"completed_at"`
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
