package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
)

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *Database
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *Database) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	_, err := r.db.NewInsert().
		Model(log).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByUserID retrieves audit logs for a specific user
func (r *AuditRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	logs := make([]*models.AuditLog, 0)

	err := r.db.NewSelect().
		Model(&logs).
		Relation("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by user id: %w", err)
	}

	return logs, nil
}

// GetByAction retrieves audit logs for a specific action
func (r *AuditRepository) GetByAction(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error) {
	logs := make([]*models.AuditLog, 0)

	err := r.db.NewSelect().
		Model(&logs).
		Where("action = ?", action).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by action: %w", err)
	}

	return logs, nil
}

// GetFailedLoginAttempts retrieves failed login attempts for an IP or user
func (r *AuditRepository) GetFailedLoginAttempts(ctx context.Context, ipAddress string, limit int) ([]*models.AuditLog, error) {
	logs := make([]*models.AuditLog, 0)

	err := r.db.NewSelect().
		Model(&logs).
		Where("action = ?", "signin_failed").
		Where("ip_address = ?", ipAddress).
		Where("created_at > ?", bun.Safe("NOW() - INTERVAL '15 minutes'")).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get failed login attempts: %w", err)
	}

	return logs, nil
}

// List retrieves all audit logs with pagination
func (r *AuditRepository) List(ctx context.Context, limit, offset int) ([]*models.AuditLog, error) {
	logs := make([]*models.AuditLog, 0)

	err := r.db.NewSelect().
		Model(&logs).
		Relation("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return logs, nil
}

// Count returns the total number of audit logs
func (r *AuditRepository) Count(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.AuditLog)(nil)).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// DeleteOlderThan deletes audit logs older than a specified duration (for cleanup)
func (r *AuditRepository) DeleteOlderThan(ctx context.Context, days int) error {
	_, err := r.db.NewDelete().
		Model((*models.AuditLog)(nil)).
		Where("created_at < ?", bun.Safe(fmt.Sprintf("NOW() - INTERVAL '%d days'", days))).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	return nil
}

// CountByActionSince counts audit log entries for a specific action since a time
func (r *AuditRepository) CountByActionSince(ctx context.Context, action models.AuditAction, since time.Time) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.AuditLog)(nil)).
		Where("action = ?", action).
		Where("created_at >= ?", since).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs by action: %w", err)
	}

	return count, nil
}

// ListByApp retrieves audit logs for a specific application with pagination
func (r *AuditRepository) ListByApp(ctx context.Context, appID uuid.UUID, limit, offset int) ([]*models.AuditLog, int, error) {
	logs := make([]*models.AuditLog, 0)

	query := r.db.NewSelect().
		Model(&logs).
		Where("application_id = ?", appID)

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs by app: %w", err)
	}

	err = query.
		Relation("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs by app: %w", err)
	}

	return logs, total, nil
}
