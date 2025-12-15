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
	// Webhook name (max 100 characters)
	Name string `json:"name" binding:"required,max=100" example:"User Events Webhook"`
	// Webhook URL endpoint (max 500 characters)
	URL string `json:"url" binding:"required,url,max=500" example:"https://api.example.com/webhooks/auth"`
	// List of events to subscribe to
	Events []string `json:"events" binding:"required,min=1" example:"user.created,user.updated,user.login"`
	// Custom HTTP headers to send with webhook requests
	Headers map[string]string `json:"headers" example:"Authorization:Bearer token123,X-Custom-Header:value"`
	// Retry configuration for failed deliveries
	RetryConfig *RetryConfig `json:"retry_config"`
}

// UpdateWebhookRequest is the request to update a webhook
type UpdateWebhookRequest struct {
	// Webhook name (max 100 characters)
	Name string `json:"name" binding:"max=100" example:"Updated Webhook Name"`
	// Webhook URL endpoint (max 500 characters)
	URL string `json:"url" binding:"url,max=500" example:"https://api.example.com/webhooks/updated"`
	// List of events to subscribe to
	Events []string `json:"events" example:"user.created,user.deleted"`
	// Custom HTTP headers to send with webhook requests
	Headers map[string]string `json:"headers" example:"Authorization:Bearer newtoken,X-Custom:value"`
	// Whether the webhook is active
	IsActive *bool `json:"is_active" example:"true"`
	// Retry configuration for failed deliveries
	RetryConfig *RetryConfig `json:"retry_config"`
}

// RetryConfig defines webhook retry behavior
type RetryConfig struct {
	// Maximum number of retry attempts
	MaxAttempts int `json:"max_attempts" example:"3"`
	// Delay in seconds for each retry attempt
	BackoffSeconds []int `json:"backoff_seconds" example:"60,300,900"`
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
	// List of webhooks
	Webhooks []WebhookWithCreator `json:"webhooks"`
	// Total number of webhooks
	Total int `json:"total" example:"15"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PerPage int `json:"per_page" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"1"`
}

// WebhookDeliveryListResponse contains paginated delivery list
type WebhookDeliveryListResponse struct {
	// List of webhook deliveries
	Deliveries []WebhookDelivery `json:"deliveries"`
	// Total number of deliveries
	Total int `json:"total" example:"150"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Number of items per page
	PerPage int `json:"per_page" example:"20"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"8"`
}

// TestWebhookRequest is used to test a webhook
type TestWebhookRequest struct {
	// Event type to simulate
	EventType string `json:"event_type" binding:"required" example:"user.created"`
	// Test payload data
	Payload map[string]interface{} `json:"payload" swaggertype:"object,string" example:"user_id:123,action:created"`
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
