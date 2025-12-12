package repository

import (
	"context"

	"github.com/smilemakc/auth-gateway/internal/models"

	"github.com/google/uuid"
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
	var settings models.BrandingSettings
	query := `SELECT * FROM branding_settings LIMIT 1`
	err := r.db.GetContext(ctx, &settings, query)
	return &settings, err
}

// UpdateBrandingSettings updates the branding settings
func (r *BrandingRepository) UpdateBrandingSettings(ctx context.Context, settings *models.BrandingSettings, updatedBy uuid.UUID) error {
	query := `
		UPDATE branding_settings
		SET logo_url = $1, favicon_url = $2, primary_color = $3, secondary_color = $4,
		    background_color = $5, custom_css = $6, company_name = $7, support_email = $8,
		    terms_url = $9, privacy_url = $10, updated_at = CURRENT_TIMESTAMP, updated_by = $11
		WHERE id = $12
	`
	_, err := r.db.ExecContext(
		ctx, query,
		settings.LogoURL, settings.FaviconURL, settings.PrimaryColor, settings.SecondaryColor,
		settings.BackgroundColor, settings.CustomCSS, settings.CompanyName, settings.SupportEmail,
		settings.TermsURL, settings.PrivacyURL, updatedBy, settings.ID,
	)
	return err
}
