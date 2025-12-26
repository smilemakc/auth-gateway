package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// LDAPRepository handles LDAP configuration database operations
type LDAPRepository struct {
	db *Database
}

// NewLDAPRepository creates a new LDAP repository
func NewLDAPRepository(db *Database) *LDAPRepository {
	return &LDAPRepository{db: db}
}

// Create creates a new LDAP configuration
func (r *LDAPRepository) Create(ctx context.Context, config *models.LDAPConfig) error {
	_, err := r.db.NewInsert().
		Model(config).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// GetByID retrieves an LDAP configuration by ID
func (r *LDAPRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.LDAPConfig, error) {
	config := new(models.LDAPConfig)
	err := r.db.NewSelect().
		Model(config).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get LDAP config: %w", err)
	}

	return config, nil
}

// GetActive retrieves the active LDAP configuration
func (r *LDAPRepository) GetActive(ctx context.Context) (*models.LDAPConfig, error) {
	config := new(models.LDAPConfig)
	err := r.db.NewSelect().
		Model(config).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active LDAP config: %w", err)
	}

	return config, nil
}

// List retrieves all LDAP configurations
func (r *LDAPRepository) List(ctx context.Context) ([]*models.LDAPConfig, error) {
	var configs []*models.LDAPConfig
	err := r.db.NewSelect().
		Model(&configs).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list LDAP configs: %w", err)
	}

	return configs, nil
}

// Update updates an LDAP configuration
func (r *LDAPRepository) Update(ctx context.Context, config *models.LDAPConfig) error {
	_, err := r.db.NewUpdate().
		Model(config).
		Where("id = ?", config.ID).
		Exec(ctx)

	return handlePgError(err)
}

// Delete deletes an LDAP configuration
func (r *LDAPRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().
		Model((*models.LDAPConfig)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return handlePgError(err)
}

// CreateSyncLog creates a new LDAP sync log entry
func (r *LDAPRepository) CreateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error {
	_, err := r.db.NewInsert().
		Model(log).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// GetSyncLogs retrieves sync logs for an LDAP configuration
func (r *LDAPRepository) GetSyncLogs(ctx context.Context, configID uuid.UUID, limit, offset int) ([]*models.LDAPSyncLog, int, error) {
	var logs []*models.LDAPSyncLog

	// Get total count
	count, err := r.db.NewSelect().
		Model((*models.LDAPSyncLog)(nil)).
		Where("ldap_config_id = ?", configID).
		Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sync logs: %w", err)
	}

	// Get paginated results
	err = r.db.NewSelect().
		Model(&logs).
		Where("ldap_config_id = ?", configID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get sync logs: %w", err)
	}

	return logs, count, nil
}

// UpdateSyncLog updates a sync log entry
func (r *LDAPRepository) UpdateSyncLog(ctx context.Context, log *models.LDAPSyncLog) error {
	_, err := r.db.NewUpdate().
		Model(log).
		Where("id = ?", log.ID).
		Exec(ctx)

	return handlePgError(err)
}
