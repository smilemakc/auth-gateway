package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// SMSSettingsRepository handles SMS settings database operations
type SMSSettingsRepository struct {
	db *sqlx.DB
}

// NewSMSSettingsRepository creates a new SMS settings repository
func NewSMSSettingsRepository(db *sqlx.DB) *SMSSettingsRepository {
	return &SMSSettingsRepository{db: db}
}

// Create creates new SMS settings
func (r *SMSSettingsRepository) Create(ctx context.Context, settings *models.SMSSettings) error {
	query := `
		INSERT INTO sms_settings (
			id, provider, enabled, account_sid, auth_token, from_number,
			aws_region, aws_access_key_id, aws_secret_access_key, aws_sender_id,
			max_per_hour, max_per_day, max_per_number, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		settings.ID,
		settings.Provider,
		settings.Enabled,
		settings.AccountSID,
		settings.AuthToken,
		settings.FromNumber,
		settings.AWSRegion,
		settings.AWSAccessKeyID,
		settings.AWSSecretAccessKey,
		settings.AWSSenderID,
		settings.MaxPerHour,
		settings.MaxPerDay,
		settings.MaxPerNumber,
		settings.CreatedBy,
		settings.CreatedAt,
		settings.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create SMS settings: %w", err)
	}

	return nil
}

// GetByID retrieves SMS settings by ID
func (r *SMSSettingsRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SMSSettings, error) {
	var settings models.SMSSettings
	query := `SELECT * FROM sms_settings WHERE id = $1`

	err := r.db.GetContext(ctx, &settings, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get SMS settings: %w", err)
	}

	return &settings, nil
}

// GetActive retrieves the active SMS settings
func (r *SMSSettingsRepository) GetActive(ctx context.Context) (*models.SMSSettings, error) {
	var settings models.SMSSettings
	query := `
		SELECT * FROM sms_settings
		WHERE enabled = true
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &settings, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get active SMS settings: %w", err)
	}

	return &settings, nil
}

// GetAll retrieves all SMS settings
func (r *SMSSettingsRepository) GetAll(ctx context.Context) ([]*models.SMSSettings, error) {
	var settings []*models.SMSSettings
	query := `SELECT * FROM sms_settings ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &settings, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all SMS settings: %w", err)
	}

	return settings, nil
}

// Update updates SMS settings
func (r *SMSSettingsRepository) Update(ctx context.Context, id uuid.UUID, settings *models.SMSSettings) error {
	query := `
		UPDATE sms_settings SET
			provider = $2,
			enabled = $3,
			account_sid = $4,
			auth_token = $5,
			from_number = $6,
			aws_region = $7,
			aws_access_key_id = $8,
			aws_secret_access_key = $9,
			aws_sender_id = $10,
			max_per_hour = $11,
			max_per_day = $12,
			max_per_number = $13,
			updated_at = $14
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		id,
		settings.Provider,
		settings.Enabled,
		settings.AccountSID,
		settings.AuthToken,
		settings.FromNumber,
		settings.AWSRegion,
		settings.AWSAccessKeyID,
		settings.AWSSecretAccessKey,
		settings.AWSSenderID,
		settings.MaxPerHour,
		settings.MaxPerDay,
		settings.MaxPerNumber,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update SMS settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrNotFound
	}

	return nil
}

// Delete deletes SMS settings
func (r *SMSSettingsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sms_settings WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete SMS settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrNotFound
	}

	return nil
}

// DisableAll disables all SMS settings
func (r *SMSSettingsRepository) DisableAll(ctx context.Context) error {
	query := `UPDATE sms_settings SET enabled = false, updated_at = $1`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to disable all SMS settings: %w", err)
	}

	return nil
}
