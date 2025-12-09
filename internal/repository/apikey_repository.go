package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/smilemakc/auth-gateway/internal/models"
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
func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	query := `
		INSERT INTO api_keys (id, user_id, name, description, key_hash, key_prefix, scopes, is_active, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		apiKey.ID,
		apiKey.UserID,
		apiKey.Name,
		apiKey.Description,
		apiKey.KeyHash,
		apiKey.KeyPrefix,
		apiKey.Scopes,
		apiKey.IsActive,
		apiKey.ExpiresAt,
	).Scan(&apiKey.CreatedAt, &apiKey.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("API key already exists")
			}
		}
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

// GetByID retrieves an API key by ID
func (r *APIKeyRepository) GetByID(id uuid.UUID) (*models.APIKey, error) {
	var apiKey models.APIKey
	query := `SELECT * FROM api_keys WHERE id = $1`

	err := r.db.Get(&apiKey, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &apiKey, nil
}

// GetByKeyHash retrieves an API key by its hash
func (r *APIKeyRepository) GetByKeyHash(keyHash string) (*models.APIKey, error) {
	var apiKey models.APIKey
	query := `SELECT * FROM api_keys WHERE key_hash = $1`

	err := r.db.Get(&apiKey, query, keyHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &apiKey, nil
}

// GetByUserID retrieves all API keys for a user
func (r *APIKeyRepository) GetByUserID(userID uuid.UUID) ([]*models.APIKey, error) {
	var apiKeys []*models.APIKey
	query := `
		SELECT * FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.Select(&apiKeys, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	return apiKeys, nil
}

// GetActiveByUserID retrieves active API keys for a user
func (r *APIKeyRepository) GetActiveByUserID(userID uuid.UUID) ([]*models.APIKey, error) {
	var apiKeys []*models.APIKey
	query := `
		SELECT * FROM api_keys
		WHERE user_id = $1 AND is_active = true
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC
	`

	err := r.db.Select(&apiKeys, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active API keys: %w", err)
	}

	return apiKeys, nil
}

// Update updates an API key
func (r *APIKeyRepository) Update(apiKey *models.APIKey) error {
	query := `
		UPDATE api_keys
		SET name = $1, description = $2, scopes = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		apiKey.Name,
		apiKey.Description,
		apiKey.Scopes,
		apiKey.IsActive,
		apiKey.ID,
	).Scan(&apiKey.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("API key not found")
		}
		return fmt.Errorf("failed to update API key: %w", err)
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *APIKeyRepository) UpdateLastUsed(id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.Exec(query, id)
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
func (r *APIKeyRepository) Revoke(id uuid.UUID) error {
	query := `UPDATE api_keys SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.Exec(query, id)
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
func (r *APIKeyRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`

	result, err := r.db.Exec(query, id)
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
func (r *APIKeyRepository) DeleteExpired() error {
	query := `DELETE FROM api_keys WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete expired API keys: %w", err)
	}

	return nil
}

// Count returns the total number of API keys for a user
func (r *APIKeyRepository) Count(userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM api_keys WHERE user_id = $1`

	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count API keys: %w", err)
	}

	return count, nil
}

// CountActive returns the number of active API keys for a user
func (r *APIKeyRepository) CountActive(userID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM api_keys
		WHERE user_id = $1 AND is_active = true
		AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`

	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active API keys: %w", err)
	}

	return count, nil
}
