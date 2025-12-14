package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"

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
	_, err := r.db.NewInsert().
		Model(webhook).
		Returning("*").
		Exec(ctx)

	return err
}

// GetWebhookByID retrieves a webhook by ID
func (r *WebhookRepository) GetWebhookByID(ctx context.Context, id uuid.UUID) (*models.Webhook, error) {
	webhook := new(models.Webhook)

	err := r.db.NewSelect().
		Model(webhook).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("webhook not found")
	}

	return webhook, err
}

// ListWebhooks retrieves all webhooks with pagination
func (r *WebhookRepository) ListWebhooks(ctx context.Context, page, perPage int) ([]models.WebhookWithCreator, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	total, err := r.db.NewSelect().
		Model((*models.Webhook)(nil)).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get webhooks with creator info
	webhooks := make([]models.WebhookWithCreator, 0)

	err = r.db.NewSelect().
		Model((*models.Webhook)(nil)).
		ColumnExpr("w.*").
		ColumnExpr("u.username as creator_username").
		ColumnExpr("u.email as creator_email").
		TableExpr("webhooks AS w").
		Join("LEFT JOIN users AS u ON w.created_by = u.id").
		Order("w.created_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(ctx, &webhooks)

	return webhooks, total, err
}

// GetActiveWebhooksByEvent retrieves active webhooks subscribed to an event
func (r *WebhookRepository) GetActiveWebhooksByEvent(ctx context.Context, eventType string) ([]models.Webhook, error) {
	webhooks := make([]models.Webhook, 0)

	eventJSON, _ := json.Marshal([]string{eventType})

	err := r.db.NewSelect().
		Model(&webhooks).
		Where("is_active = ?", true).
		Where("events @> ?::jsonb", eventJSON).
		Order("created_at").
		Scan(ctx)

	return webhooks, err
}

// UpdateWebhook updates a webhook
func (r *WebhookRepository) UpdateWebhook(ctx context.Context, id uuid.UUID, name, url string, events, headers json.RawMessage, isActive bool) error {
	result, err := r.db.NewUpdate().
		Model((*models.Webhook)(nil)).
		Set("name = ?", name).
		Set("url = ?", url).
		Set("events = ?", events).
		Set("headers = ?", headers).
		Set("is_active = ?", isActive).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", id).
		Exec(ctx)

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
	_, err := r.db.NewUpdate().
		Model((*models.Webhook)(nil)).
		Set("last_triggered_at = ?", bun.Safe("NOW()")).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

// DeleteWebhook deletes a webhook
func (r *WebhookRepository) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.Webhook)(nil)).
		Where("id = ?", id).
		Exec(ctx)

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
	_, err := r.db.NewInsert().
		Model(delivery).
		Returning("*").
		Exec(ctx)

	return err
}

// GetWebhookDeliveryByID retrieves a delivery by ID
func (r *WebhookRepository) GetWebhookDeliveryByID(ctx context.Context, id uuid.UUID) (*models.WebhookDelivery, error) {
	delivery := new(models.WebhookDelivery)

	err := r.db.NewSelect().
		Model(delivery).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("webhook delivery not found")
	}

	return delivery, err
}

// ListWebhookDeliveries retrieves deliveries for a webhook with pagination
func (r *WebhookRepository) ListWebhookDeliveries(ctx context.Context, webhookID uuid.UUID, page, perPage int) ([]models.WebhookDelivery, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	total, err := r.db.NewSelect().
		Model((*models.WebhookDelivery)(nil)).
		Where("webhook_id = ?", webhookID).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get deliveries
	deliveries := make([]models.WebhookDelivery, 0)

	err = r.db.NewSelect().
		Model(&deliveries).
		Where("webhook_id = ?", webhookID).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(ctx)

	return deliveries, total, err
}

// GetPendingDeliveries retrieves pending/failed deliveries ready for retry
func (r *WebhookRepository) GetPendingDeliveries(ctx context.Context, limit int) ([]models.WebhookDelivery, error) {
	deliveries := make([]models.WebhookDelivery, 0)

	err := r.db.NewSelect().
		Model(&deliveries).
		Where("status IN (?)", bun.In([]string{"pending", "failed"})).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("next_retry_at IS NULL").
				WhereOr("next_retry_at <= ?", bun.Safe("NOW()"))
		}).
		Order("created_at").
		Limit(limit).
		Scan(ctx)

	return deliveries, err
}

// UpdateDeliveryStatus updates a delivery's status
func (r *WebhookRepository) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status string, httpStatus *int, responseBody string, nextRetry *interface{}) error {
	query := r.db.NewUpdate().
		Model((*models.WebhookDelivery)(nil)).
		Set("status = ?", status).
		Set("http_status_code = ?", httpStatus).
		Set("response_body = ?", responseBody).
		SetColumn("attempts", "attempts + 1").
		Where("id = ?", id)

	if nextRetry != nil {
		query = query.Set("next_retry_at = ?", nextRetry)
	} else {
		query = query.Set("completed_at = ?", bun.Safe("NOW()"))
	}

	_, err := query.Exec(ctx)
	return err
}

// DeleteOldDeliveries deletes old completed deliveries
func (r *WebhookRepository) DeleteOldDeliveries(ctx context.Context, olderThanDays int) error {
	_, err := r.db.NewDelete().
		Model((*models.WebhookDelivery)(nil)).
		Where("completed_at IS NOT NULL").
		Where("completed_at < ?", bun.Safe("NOW() - INTERVAL '1 day' * ?"), olderThanDays).
		Exec(ctx)

	return err
}
