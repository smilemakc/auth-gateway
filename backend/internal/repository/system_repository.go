package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
)

// SystemRepository handles system settings database operations
type SystemRepository struct {
	db *Database
}

// NewSystemRepository creates a new system repository
func NewSystemRepository(db *Database) *SystemRepository {
	return &SystemRepository{db: db}
}

// GetSetting retrieves a system setting by key
func (r *SystemRepository) GetSetting(ctx context.Context, key string) (*models.SystemSetting, error) {
	setting := new(models.SystemSetting)

	err := r.db.NewSelect().
		Model(setting).
		Where("key = ?", key).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("setting not found: %s", key)
	}

	return setting, err
}

// GetAllSettings retrieves all system settings
func (r *SystemRepository) GetAllSettings(ctx context.Context) ([]models.SystemSetting, error) {
	settings := make([]models.SystemSetting, 0)

	err := r.db.NewSelect().
		Model(&settings).
		Order("key").
		Scan(ctx)

	return settings, err
}

// GetPublicSettings retrieves public system settings
func (r *SystemRepository) GetPublicSettings(ctx context.Context) ([]models.SystemSetting, error) {
	settings := make([]models.SystemSetting, 0)

	err := r.db.NewSelect().
		Model(&settings).
		Where("is_public = ?", true).
		Order("key").
		Scan(ctx)

	return settings, err
}

// UpdateSetting updates a system setting
func (r *SystemRepository) UpdateSetting(ctx context.Context, key, value string, updatedBy *uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.SystemSetting)(nil)).
		Set("value = ?", value).
		Set("updated_at = ?", bun.Ident("CURRENT_TIMESTAMP")).
		Set("updated_by = ?", updatedBy).
		Where("key = ?", key).
		Exec(ctx)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("setting not found: %s", key)
	}

	return nil
}

// CreateSetting creates a new system setting
func (r *SystemRepository) CreateSetting(ctx context.Context, setting *models.SystemSetting) error {
	_, err := r.db.NewInsert().
		Model(setting).
		Returning("updated_at").
		Exec(ctx)

	return err
}

// DeleteSetting deletes a system setting
func (r *SystemRepository) DeleteSetting(ctx context.Context, key string) error {
	result, err := r.db.NewDelete().
		Model((*models.SystemSetting)(nil)).
		Where("key = ?", key).
		Exec(ctx)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("setting not found: %s", key)
	}

	return nil
}

// ============================================================
// Health Metrics
// ============================================================

// RecordHealthMetric records a health metric
func (r *SystemRepository) RecordHealthMetric(ctx context.Context, metric *models.HealthMetric) error {
	_, err := r.db.NewInsert().
		Model(metric).
		Returning("*").
		Exec(ctx)

	return err
}

// GetRecentMetrics retrieves recent metrics for a specific metric name
func (r *SystemRepository) GetRecentMetrics(ctx context.Context, metricName string, limit int) ([]models.HealthMetric, error) {
	metrics := make([]models.HealthMetric, 0)

	err := r.db.NewSelect().
		Model(&metrics).
		Where("metric_name = ?", metricName).
		Order("recorded_at DESC").
		Limit(limit).
		Scan(ctx)

	return metrics, err
}

// DeleteOldMetrics deletes old health metrics
func (r *SystemRepository) DeleteOldMetrics(ctx context.Context, olderThanDays int) error {
	_, err := r.db.NewDelete().
		Model((*models.HealthMetric)(nil)).
		Where("recorded_at < ?", bun.Safe("NOW() - INTERVAL '1 day' * ?"), olderThanDays).
		Exec(ctx)

	return err
}
