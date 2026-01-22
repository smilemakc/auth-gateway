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

// EmailProviderRepository handles email provider database operations
type EmailProviderRepository struct {
	db *Database
}

// NewEmailProviderRepository creates a new email provider repository
func NewEmailProviderRepository(db *Database) *EmailProviderRepository {
	return &EmailProviderRepository{db: db}
}

// Create creates a new email provider
func (r *EmailProviderRepository) Create(ctx context.Context, provider *models.EmailProvider) error {
	_, err := r.db.NewInsert().
		Model(provider).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create email provider: %w", err)
	}

	return nil
}

// GetByID retrieves an email provider by ID
func (r *EmailProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.EmailProvider, error) {
	provider := new(models.EmailProvider)

	err := r.db.NewSelect().
		Model(provider).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get email provider: %w", err)
	}

	return provider, nil
}

// GetAll retrieves all email providers
func (r *EmailProviderRepository) GetAll(ctx context.Context) ([]*models.EmailProvider, error) {
	providers := make([]*models.EmailProvider, 0)

	err := r.db.NewSelect().
		Model(&providers).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get all email providers: %w", err)
	}

	return providers, nil
}

// GetActive retrieves all active email providers
func (r *EmailProviderRepository) GetActive(ctx context.Context) ([]*models.EmailProvider, error) {
	providers := make([]*models.EmailProvider, 0)

	err := r.db.NewSelect().
		Model(&providers).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get active email providers: %w", err)
	}

	return providers, nil
}

// Update updates an email provider
func (r *EmailProviderRepository) Update(ctx context.Context, id uuid.UUID, provider *models.EmailProvider) error {
	result, err := r.db.NewUpdate().
		Model((*models.EmailProvider)(nil)).
		Set("name = ?", provider.Name).
		Set("type = ?", provider.Type).
		Set("is_active = ?", provider.IsActive).
		Set("smtp_host = ?", provider.SMTPHost).
		Set("smtp_port = ?", provider.SMTPPort).
		Set("smtp_username = ?", provider.SMTPUsername).
		Set("smtp_password = ?", provider.SMTPPassword).
		Set("smtp_use_tls = ?", provider.SMTPUseTLS).
		Set("sendgrid_api_key = ?", provider.SendGridAPIKey).
		Set("ses_region = ?", provider.SESRegion).
		Set("ses_access_key_id = ?", provider.SESAccessKeyID).
		Set("ses_secret_access_key = ?", provider.SESSecretAccessKey).
		Set("mailgun_domain = ?", provider.MailgunDomain).
		Set("mailgun_api_key = ?", provider.MailgunAPIKey).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update email provider: %w", err)
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

// Delete deletes an email provider
func (r *EmailProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.EmailProvider)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete email provider: %w", err)
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

// DisableAll disables all email providers
func (r *EmailProviderRepository) DisableAll(ctx context.Context) error {
	_, err := r.db.NewUpdate().
		Model((*models.EmailProvider)(nil)).
		Set("is_active = ?", false).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to disable all email providers: %w", err)
	}

	return nil
}
