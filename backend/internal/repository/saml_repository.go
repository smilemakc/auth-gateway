package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// SAMLRepository handles SAML Service Provider database operations
type SAMLRepository struct {
	db *Database
}

// NewSAMLRepository creates a new SAML repository
func NewSAMLRepository(db *Database) *SAMLRepository {
	return &SAMLRepository{db: db}
}

// Create creates a new SAML Service Provider
func (r *SAMLRepository) Create(ctx context.Context, sp *models.SAMLServiceProvider) error {
	_, err := r.db.NewInsert().
		Model(sp).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// GetByID retrieves a SAML SP by ID
func (r *SAMLRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.SAMLServiceProvider, error) {
	sp := new(models.SAMLServiceProvider)
	err := r.db.NewSelect().
		Model(sp).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SAML SP: %w", err)
	}

	return sp, nil
}

// GetByEntityID retrieves a SAML SP by EntityID
func (r *SAMLRepository) GetByEntityID(ctx context.Context, entityID string) (*models.SAMLServiceProvider, error) {
	sp := new(models.SAMLServiceProvider)
	err := r.db.NewSelect().
		Model(sp).
		Where("entity_id = ?", entityID).
		Where("is_active = ?", true).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SAML SP by entity ID: %w", err)
	}

	return sp, nil
}

// List retrieves all SAML SPs with pagination
func (r *SAMLRepository) List(ctx context.Context, page, pageSize int) ([]*models.SAMLServiceProvider, int, error) {
	var sps []*models.SAMLServiceProvider

	// Get total count
	count, err := r.db.NewSelect().
		Model((*models.SAMLServiceProvider)(nil)).
		Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count SAML SPs: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err = r.db.NewSelect().
		Model(&sps).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list SAML SPs: %w", err)
	}

	return sps, count, nil
}

// Update updates a SAML SP
func (r *SAMLRepository) Update(ctx context.Context, sp *models.SAMLServiceProvider) error {
	_, err := r.db.NewUpdate().
		Model(sp).
		Where("id = ?", sp.ID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update SAML SP: %w", err)
	}

	return nil
}

// Delete deletes a SAML SP
func (r *SAMLRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().
		Model((*models.SAMLServiceProvider)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return handlePgError(err)
}
