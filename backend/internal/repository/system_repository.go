package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/smilemakc/auth-gateway/internal/models"

	"github.com/google/uuid"
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
	var setting models.SystemSetting
	query := `SELECT * FROM system_settings WHERE key = $1`
	err := r.db.GetContext(ctx, &setting, query, key)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("setting not found: %s", key)
	}
	return &setting, err
}

// GetAllSettings retrieves all system settings
func (r *SystemRepository) GetAllSettings(ctx context.Context) ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	query := `SELECT * FROM system_settings ORDER BY key`
	err := r.db.SelectContext(ctx, &settings, query)
	return settings, err
}

// GetPublicSettings retrieves public system settings
func (r *SystemRepository) GetPublicSettings(ctx context.Context) ([]models.SystemSetting, error) {
	var settings []models.SystemSetting
	query := `SELECT * FROM system_settings WHERE is_public = true ORDER BY key`
	err := r.db.SelectContext(ctx, &settings, query)
	return settings, err
}

// UpdateSetting updates a system setting
func (r *SystemRepository) UpdateSetting(ctx context.Context, key, value string, updatedBy *uuid.UUID) error {
	query := `
		UPDATE system_settings
		SET value = $1, updated_at = CURRENT_TIMESTAMP, updated_by = $2
		WHERE key = $3
	`
	result, err := r.db.ExecContext(ctx, query, value, updatedBy, key)
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
	query := `
		INSERT INTO system_settings (key, value, description, setting_type, is_public, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING updated_at
	`
	return r.db.QueryRowContext(
		ctx, query,
		setting.Key, setting.Value, setting.Description,
		setting.SettingType, setting.IsPublic, setting.UpdatedBy,
	).Scan(&setting.UpdatedAt)
}

// DeleteSetting deletes a system setting
func (r *SystemRepository) DeleteSetting(ctx context.Context, key string) error {
	query := `DELETE FROM system_settings WHERE key = $1`
	result, err := r.db.ExecContext(ctx, query, key)
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
	query := `
		INSERT INTO health_metrics (metric_name, metric_value, metric_unit, metadata)
		VALUES ($1, $2, $3, $4)
		RETURNING id, recorded_at
	`
	return r.db.QueryRowContext(
		ctx, query,
		metric.MetricName, metric.MetricValue, metric.MetricUnit, metric.Metadata,
	).Scan(&metric.ID, &metric.RecordedAt)
}

// GetRecentMetrics retrieves recent metrics for a specific metric name
func (r *SystemRepository) GetRecentMetrics(ctx context.Context, metricName string, limit int) ([]models.HealthMetric, error) {
	var metrics []models.HealthMetric
	query := `
		SELECT * FROM health_metrics
		WHERE metric_name = $1
		ORDER BY recorded_at DESC
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &metrics, query, metricName, limit)
	return metrics, err
}

// DeleteOldMetrics deletes old health metrics
func (r *SystemRepository) DeleteOldMetrics(ctx context.Context, olderThanDays int) error {
	query := `
		DELETE FROM health_metrics
		WHERE recorded_at < NOW() - INTERVAL '1 day' * $1
	`
	_, err := r.db.ExecContext(ctx, query, olderThanDays)
	return err
}
