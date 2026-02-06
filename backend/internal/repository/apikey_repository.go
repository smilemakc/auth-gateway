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

// APIKeyRepository handles API key-related database operations
type APIKeyRepository struct {
	db *Database
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *Database) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(ctx context.Context, apiKey *models.APIKey) error {
	_, err := r.db.NewInsert().
		Model(apiKey).
		Returning("*").
		Exec(ctx)

	return handlePgError(err)
}

// GetByID retrieves an API key by ID
func (r *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	apiKey := new(models.APIKey)

	err := r.db.NewSelect().
		Model(apiKey).
		Where("id = ?", id).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return apiKey, nil
}

// GetByKeyHash retrieves an API key by its hash
func (r *APIKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	apiKey := new(models.APIKey)

	err := r.db.NewSelect().
		Model(apiKey).
		Where("key_hash = ?", keyHash).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return apiKey, nil
}

// GetByUserID retrieves all API keys for a user
func (r *APIKeyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	apiKeys := make([]*models.APIKey, 0)

	err := r.db.NewSelect().
		Model(&apiKeys).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	return apiKeys, nil
}

// GetActiveByUserID retrieves active API keys for a user
func (r *APIKeyRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	apiKeys := make([]*models.APIKey, 0)

	err := r.db.NewSelect().
		Model(&apiKeys).
		Where("user_id = ?", userID).
		Where("is_active = ?", true).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("expires_at IS NULL").
				WhereOr("expires_at > ?", bun.Safe("NOW()"))
		}).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get active API keys: %w", err)
	}

	return apiKeys, nil
}

// Update updates an API key
func (r *APIKeyRepository) Update(ctx context.Context, apiKey *models.APIKey) error {
	result, err := r.db.NewUpdate().
		Model(apiKey).
		Column("name", "description", "scopes", "is_active").
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		WherePK().
		Returning("updated_at").
		Exec(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("API key not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.APIKey)(nil)).
		Set("last_used_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// Revoke revokes an API key (sets is_active to false)
func (r *APIKeyRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.APIKey)(nil)).
		Set("is_active = ?", false).
		Set("updated_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// Delete permanently deletes an API key
func (r *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*models.APIKey)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// DeleteExpired deletes expired API keys
func (r *APIKeyRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.NewDelete().
		Model((*models.APIKey)(nil)).
		Where("expires_at IS NOT NULL").
		Where("expires_at < ?", bun.Safe("CURRENT_TIMESTAMP")).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete expired API keys: %w", err)
	}

	return nil
}

// Count returns the total number of API keys for a user
func (r *APIKeyRepository) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.APIKey)(nil)).
		Where("user_id = ?", userID).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count API keys: %w", err)
	}

	return count, nil
}

// CountActive returns the number of active API keys for a user
func (r *APIKeyRepository) CountActive(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.APIKey)(nil)).
		Where("user_id = ?", userID).
		Where("is_active = ?", true).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("expires_at IS NULL").
				WhereOr("expires_at > ?", bun.Safe("CURRENT_TIMESTAMP"))
		}).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count active API keys: %w", err)
	}

	return count, nil
}

// ListAll returns all API keys (admin only)
func (r *APIKeyRepository) ListAll(ctx context.Context) ([]*models.APIKey, error) {
	keys := make([]*models.APIKey, 0)

	err := r.db.NewSelect().
		Model(&keys).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list all API keys: %w", err)
	}

	return keys, nil
}

// GetByUserIDAndApp retrieves API keys for a user in a specific application
func (r *APIKeyRepository) GetByUserIDAndApp(ctx context.Context, userID, appID uuid.UUID) ([]*models.APIKey, error) {
	apiKeys := make([]*models.APIKey, 0)

	err := r.db.NewSelect().
		Model(&apiKeys).
		Where("user_id = ?", userID).
		Where("application_id = ?", appID).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get API keys by user and app: %w", err)
	}

	return apiKeys, nil
}
