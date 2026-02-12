package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
)

// BrandingRepository handles branding settings database operations
type BrandingRepository struct {
	db *Database
}

// NewBrandingRepository creates a new branding repository
func NewBrandingRepository(db *Database) *BrandingRepository {
	return &BrandingRepository{db: db}
}

// GetBrandingSettings retrieves the branding settings (single row table)
func (r *BrandingRepository) GetBrandingSettings(ctx context.Context) (*models.BrandingSettings, error) {
	settings := new(models.BrandingSettings)

	err := r.db.NewSelect().
		Model(settings).
		Limit(1).
		Scan(ctx)

	return settings, err
}

// UpdateBrandingSettings updates the branding settings (single-row table)
func (r *BrandingRepository) UpdateBrandingSettings(ctx context.Context, settings *models.BrandingSettings, updatedBy uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model(settings).
		Column("logo_url", "favicon_url", "primary_color", "secondary_color").
		Column("background_color", "custom_css", "company_name", "support_email").
		Column("terms_url", "privacy_url").
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Set("updated_by = ?", updatedBy).
		Where("TRUE").
		Exec(ctx)

	return err
}
