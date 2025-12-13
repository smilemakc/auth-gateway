package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/smilemakc/auth-gateway/internal/models"

	"github.com/google/uuid"
)

// WebhookRepository handles webhook database operations
type WebhookRepository struct {
	db *Database
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *Database) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// CreateWebhook creates a new webhook
func (r *WebhookRepository) CreateWebhook(ctx context.Context, webhook *models.Webhook) error {
	query := `
		INSERT INTO webhooks (name, url, secret_key, events, headers, is_active, retry_config, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(
		ctx, query,
		webhook.Name, webhook.URL, webhook.SecretKey, webhook.Events,
		webhook.Headers, webhook.IsActive, webhook.RetryConfig, webhook.CreatedBy,
	).Scan(&webhook.ID, &webhook.CreatedAt, &webhook.UpdatedAt)
}

// GetWebhookByID retrieves a webhook by ID
func (r *WebhookRepository) GetWebhookByID(ctx context.Context, id uuid.UUID) (*models.Webhook, error) {
	var webhook models.Webhook
	query := `SELECT * FROM webhooks WHERE id = $1`
	err := r.db.GetContext(ctx, &webhook, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("webhook not found")
	}
	return &webhook, err
}

// ListWebhooks retrieves all webhooks with pagination
func (r *WebhookRepository) ListWebhooks(ctx context.Context, page, perPage int) ([]models.WebhookWithCreator, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM webhooks`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// Get webhooks
	var webhooks []models.WebhookWithCreator
	query := `
		SELECT
			w.*,
			u.username as creator_username,
			u.email as creator_email
		FROM webhooks w
		LEFT JOIN users u ON w.created_by = u.id
		ORDER BY w.created_at DESC
		LIMIT $1 OFFSET $2
	`
	err = r.db.SelectContext(ctx, &webhooks, query, perPage, offset)
	return webhooks, total, err
}

// GetActiveWebhooksByEvent retrieves active webhooks subscribed to an event
func (r *WebhookRepository) GetActiveWebhooksByEvent(ctx context.Context, eventType string) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	query := `
		SELECT * FROM webhooks
		WHERE is_active = true
		  AND events @> $1::jsonb
		ORDER BY created_at
	`
	eventJSON, _ := json.Marshal([]string{eventType})
	err := r.db.SelectContext(ctx, &webhooks, query, eventJSON)
	return webhooks, err
}

// UpdateWebhook updates a webhook
func (r *WebhookRepository) UpdateWebhook(ctx context.Context, id uuid.UUID, name, url string, events, headers json.RawMessage, isActive bool) error {
	query := `
		UPDATE webhooks
		SET name = $1, url = $2, events = $3, headers = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`
	result, err := r.db.ExecContext(ctx, query, name, url, events, headers, isActive, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("webhook not found")
	}
	return nil
}

// UpdateWebhookLastTriggered updates the last triggered timestamp
func (r *WebhookRepository) UpdateWebhookLastTriggered(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE webhooks SET last_triggered_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DeleteWebhook deletes a webhook
func (r *WebhookRepository) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("webhook not found")
	}
	return nil
}

// ============================================================
// Webhook Delivery Methods
// ============================================================

// CreateWebhookDelivery creates a new webhook delivery record
func (r *WebhookRepository) CreateWebhookDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	query := `
		INSERT INTO webhook_deliveries (webhook_id, event_type, payload, status, attempts, next_retry_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(
		ctx, query,
		delivery.WebhookID, delivery.EventType, delivery.Payload,
		delivery.Status, delivery.Attempts, delivery.NextRetryAt,
	).Scan(&delivery.ID, &delivery.CreatedAt)
}

// GetWebhookDeliveryByID retrieves a delivery by ID
func (r *WebhookRepository) GetWebhookDeliveryByID(ctx context.Context, id uuid.UUID) (*models.WebhookDelivery, error) {
	var delivery models.WebhookDelivery
	query := `SELECT * FROM webhook_deliveries WHERE id = $1`
	err := r.db.GetContext(ctx, &delivery, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("webhook delivery not found")
	}
	return &delivery, err
}

// ListWebhookDeliveries retrieves deliveries for a webhook with pagination
func (r *WebhookRepository) ListWebhookDeliveries(ctx context.Context, webhookID uuid.UUID, page, perPage int) ([]models.WebhookDelivery, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id = $1`
	err := r.db.GetContext(ctx, &total, countQuery, webhookID)
	if err != nil {
		return nil, 0, err
	}

	// Get deliveries
	var deliveries []models.WebhookDelivery
	query := `
		SELECT * FROM webhook_deliveries
		WHERE webhook_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err = r.db.SelectContext(ctx, &deliveries, query, webhookID, perPage, offset)
	return deliveries, total, err
}

// GetPendingDeliveries retrieves pending/failed deliveries ready for retry
func (r *WebhookRepository) GetPendingDeliveries(ctx context.Context, limit int) ([]models.WebhookDelivery, error) {
	var deliveries []models.WebhookDelivery
	query := `
		SELECT * FROM webhook_deliveries
		WHERE status IN ('pending', 'failed')
		  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at
		LIMIT $1
	`
	err := r.db.SelectContext(ctx, &deliveries, query, limit)
	return deliveries, err
}

// UpdateDeliveryStatus updates a delivery's status
func (r *WebhookRepository) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status string, httpStatus *int, responseBody string, nextRetry *interface{}) error {
	var query string
	var args []interface{}

	if nextRetry != nil {
		query = `
			UPDATE webhook_deliveries
			SET status = $1, http_status_code = $2, response_body = $3,
			    attempts = attempts + 1, next_retry_at = $4
			WHERE id = $5
		`
		args = []interface{}{status, httpStatus, responseBody, nextRetry, id}
	} else {
		query = `
			UPDATE webhook_deliveries
			SET status = $1, http_status_code = $2, response_body = $3,
			    attempts = attempts + 1, completed_at = NOW()
			WHERE id = $4
		`
		args = []interface{}{status, httpStatus, responseBody, id}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteOldDeliveries deletes old completed deliveries
func (r *WebhookRepository) DeleteOldDeliveries(ctx context.Context, olderThanDays int) error {
	query := `
		DELETE FROM webhook_deliveries
		WHERE completed_at IS NOT NULL
		  AND completed_at < NOW() - INTERVAL '1 day' * $1
	`
	_, err := r.db.ExecContext(ctx, query, olderThanDays)
	return err
}
