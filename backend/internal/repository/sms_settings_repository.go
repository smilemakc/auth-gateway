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

// SMSSettingsRepository handles SMS settings database operations
type SMSSettingsRepository struct {
	db *Database
}

// NewSMSSettingsRepository creates a new SMS settings repository
func NewSMSSettingsRepository(db *Database) *SMSSettingsRepository {
	return &SMSSettingsRepository{db: db}
}

// Create creates new SMS settings
func (r *SMSSettingsRepository) Create(ctx context.Context, settings *models.SMSSettings) error {
	_, err := r.db.NewInsert().
		Model(settings).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create SMS settings: %w", err)
	}

	return nil
}

// GetByID retrieves SMS settings by ID
func (r *SMSSettingsRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SMSSettings, error) {
	settings := new(models.SMSSettings)

	err := r.db.NewSelect().
		Model(settings).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get SMS settings: %w", err)
	}

	return settings, nil
}

// GetActive retrieves the active SMS settings
func (r *SMSSettingsRepository) GetActive(ctx context.Context) (*models.SMSSettings, error) {
	settings := new(models.SMSSettings)

	err := r.db.NewSelect().
		Model(settings).
		Where("enabled = ?", true).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get active SMS settings: %w", err)
	}

	return settings, nil
}

// GetAll retrieves all SMS settings
func (r *SMSSettingsRepository) GetAll(ctx context.Context) ([]*models.SMSSettings, error) {
	settings := make([]*models.SMSSettings, 0)

	err := r.db.NewSelect().
		Model(&settings).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get all SMS settings: %w", err)
	}

	return settings, nil
}

// Update updates SMS settings
func (r *SMSSettingsRepository) Update(ctx context.Context, id uuid.UUID, settings *models.SMSSettings) error {
	result, err := r.db.NewUpdate().
		Model((*models.SMSSettings)(nil)).
		Set("provider = ?", settings.Provider).
		Set("enabled = ?", settings.Enabled).
		Set("account_sid = ?", settings.AccountSID).
		Set("auth_token = ?", settings.AuthToken).
		Set("from_number = ?", settings.FromNumber).
		Set("aws_region = ?", settings.AWSRegion).
		Set("aws_access_key_id = ?", settings.AWSAccessKeyID).
		Set("aws_secret_access_key = ?", settings.AWSSecretAccessKey).
		Set("aws_sender_id = ?", settings.AWSSenderID).
		Set("max_per_hour = ?", settings.MaxPerHour).
		Set("max_per_day = ?", settings.MaxPerDay).
		Set("max_per_number = ?", settings.MaxPerNumber).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", id).
		Exec(ctx)

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
	result, err := r.db.NewDelete().
		Model((*models.SMSSettings)(nil)).
		Where("id = ?", id).
		Exec(ctx)

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
	_, err := r.db.NewUpdate().
		Model((*models.SMSSettings)(nil)).
		Set("enabled = ?", false).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to disable all SMS settings: %w", err)
	}

	return nil
}
