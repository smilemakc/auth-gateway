package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

type AppOAuthProviderRepository struct {
	db *Database
}

func NewAppOAuthProviderRepository(db *Database) *AppOAuthProviderRepository {
	return &AppOAuthProviderRepository{db: db}
}

func (r *AppOAuthProviderRepository) Create(ctx context.Context, provider *models.ApplicationOAuthProvider) error {
	provider.CreatedAt = time.Now()
	provider.UpdatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(provider).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

func (r *AppOAuthProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ApplicationOAuthProvider, error) {
	provider := new(models.ApplicationOAuthProvider)

	err := r.db.NewSelect().
		Model(provider).
		Where("aop.id = ?", id).
		Relation("Application").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("app oauth provider not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app oauth provider by id: %w", err)
	}

	return provider, nil
}

func (r *AppOAuthProviderRepository) GetByAppAndProvider(ctx context.Context, appID uuid.UUID, provider string) (*models.ApplicationOAuthProvider, error) {
	p := new(models.ApplicationOAuthProvider)

	err := r.db.NewSelect().
		Model(p).
		Where("aop.application_id = ?", appID).
		Where("aop.provider = ?", provider).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("app oauth provider not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app oauth provider: %w", err)
	}

	return p, nil
}

func (r *AppOAuthProviderRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]*models.ApplicationOAuthProvider, error) {
	providers := make([]*models.ApplicationOAuthProvider, 0)

	err := r.db.NewSelect().
		Model(&providers).
		Where("aop.application_id = ?", appID).
		Order("aop.provider ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list app oauth providers: %w", err)
	}

	return providers, nil
}

func (r *AppOAuthProviderRepository) ListAll(ctx context.Context) ([]*models.ApplicationOAuthProvider, error) {
	providers := make([]*models.ApplicationOAuthProvider, 0)

	err := r.db.NewSelect().
		Model(&providers).
		Order("aop.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list all oauth providers: %w", err)
	}

	return providers, nil
}

func (r *AppOAuthProviderRepository) Update(ctx context.Context, provider *models.ApplicationOAuthProvider) error {
	provider.UpdatedAt = time.Now()

	result, err := r.db.NewUpdate().
		Model(provider).
		Column("provider", "client_id", "client_secret", "callback_url",
			"scopes", "auth_url", "token_url", "user_info_url", "is_active", "updated_at").
		WherePK().
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update app oauth provider: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("app oauth provider not found")
	}

	return nil
}

func (r *AppOAuthProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.ApplicationOAuthProvider)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete app oauth provider: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("app oauth provider not found")
	}

	return nil
}
