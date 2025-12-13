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

// OAuthRepository handles OAuth-related database operations
type OAuthRepository struct {
	db *Database
}

// NewOAuthRepository creates a new OAuth repository
func NewOAuthRepository(db *Database) *OAuthRepository {
	return &OAuthRepository{db: db}
}

// CreateOAuthAccount creates a new OAuth account
func (r *OAuthRepository) CreateOAuthAccount(ctx context.Context, account *models.OAuthAccount) error {
	_, err := r.db.NewInsert().
		Model(account).
		On("CONFLICT (provider, provider_user_id) DO UPDATE").
		Set("access_token = EXCLUDED.access_token").
		Set("refresh_token = EXCLUDED.refresh_token").
		Set("token_expires_at = EXCLUDED.token_expires_at").
		Set("profile_data = EXCLUDED.profile_data").
		Set("updated_at = ?", bun.Ident("CURRENT_TIMESTAMP")).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create oauth account: %w", err)
	}

	return nil
}

// GetOAuthAccount retrieves an OAuth account by provider and provider user ID
func (r *OAuthRepository) GetOAuthAccount(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
	account := new(models.OAuthAccount)

	err := r.db.NewSelect().
		Model(account).
		Where("provider = ?", provider).
		Where("provider_user_id = ?", providerUserID).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Not found, but not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth account: %w", err)
	}

	return account, nil
}

// GetOAuthAccountsByUserID retrieves all OAuth accounts for a user
func (r *OAuthRepository) GetOAuthAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
	accounts := make([]*models.OAuthAccount, 0)

	err := r.db.NewSelect().
		Model(&accounts).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get oauth accounts by user id: %w", err)
	}

	return accounts, nil
}

// UpdateOAuthAccount updates an OAuth account
func (r *OAuthRepository) UpdateOAuthAccount(ctx context.Context, account *models.OAuthAccount) error {
	result, err := r.db.NewUpdate().
		Model(account).
		Column("access_token", "refresh_token", "token_expires_at", "profile_data").
		Set("updated_at = ?", bun.Ident("CURRENT_TIMESTAMP")).
		WherePK().
		Returning("updated_at").
		Exec(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("oauth account not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update oauth account: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("oauth account not found")
	}

	return nil
}

// DeleteOAuthAccount deletes an OAuth account
func (r *OAuthRepository) DeleteOAuthAccount(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.OAuthAccount)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete oauth account: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("oauth account not found")
	}

	return nil
}

// DeleteOAuthAccountsByProvider deletes all OAuth accounts for a user by provider
func (r *OAuthRepository) DeleteOAuthAccountsByProvider(ctx context.Context, userID uuid.UUID, provider string) error {
	_, err := r.db.NewDelete().
		Model((*models.OAuthAccount)(nil)).
		Where("user_id = ?", userID).
		Where("provider = ?", provider).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete oauth accounts by provider: %w", err)
	}

	return nil
}

// GetByUserID returns all OAuth accounts for a user
func (r *OAuthRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
	accounts := make([]*models.OAuthAccount, 0)

	err := r.db.NewSelect().
		Model(&accounts).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get oauth accounts by user ID: %w", err)
	}

	return accounts, nil
}

// ListAll returns all OAuth accounts (admin only)
func (r *OAuthRepository) ListAll(ctx context.Context) ([]*models.OAuthAccount, error) {
	accounts := make([]*models.OAuthAccount, 0)

	err := r.db.NewSelect().
		Model(&accounts).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list all oauth accounts: %w", err)
	}

	return accounts, nil
}
