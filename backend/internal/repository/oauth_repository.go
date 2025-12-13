package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/smilemakc/auth-gateway/internal/models"
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
func (r *OAuthRepository) CreateOAuthAccount(account *models.OAuthAccount) error {
	query := `
		INSERT INTO oauth_accounts (id, user_id, provider, provider_user_id, access_token, refresh_token, token_expires_at, profile_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (provider, provider_user_id)
		DO UPDATE SET
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_expires_at = EXCLUDED.token_expires_at,
			profile_data = EXCLUDED.profile_data,
			updated_at = CURRENT_TIMESTAMP
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		account.ID,
		account.UserID,
		account.Provider,
		account.ProviderUserID,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
		account.ProfileData,
	).Scan(&account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				// Unique violation - account already exists and was updated
				// Query again to get the timestamps
				return r.db.Get(account,
					`SELECT * FROM oauth_accounts WHERE provider = $1 AND provider_user_id = $2`,
					account.Provider, account.ProviderUserID)
			}
		}
		return fmt.Errorf("failed to create oauth account: %w", err)
	}

	return nil
}

// GetOAuthAccount retrieves an OAuth account by provider and provider user ID
func (r *OAuthRepository) GetOAuthAccount(provider, providerUserID string) (*models.OAuthAccount, error) {
	var account models.OAuthAccount
	query := `SELECT * FROM oauth_accounts WHERE provider = $1 AND provider_user_id = $2`

	err := r.db.Get(&account, query, provider, providerUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found, but not an error
		}
		return nil, fmt.Errorf("failed to get oauth account: %w", err)
	}

	return &account, nil
}

// GetOAuthAccountsByUserID retrieves all OAuth accounts for a user
func (r *OAuthRepository) GetOAuthAccountsByUserID(userID uuid.UUID) ([]*models.OAuthAccount, error) {
	var accounts []*models.OAuthAccount
	query := `SELECT * FROM oauth_accounts WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.Select(&accounts, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth accounts by user id: %w", err)
	}

	return accounts, nil
}

// UpdateOAuthAccount updates an OAuth account
func (r *OAuthRepository) UpdateOAuthAccount(account *models.OAuthAccount) error {
	query := `
		UPDATE oauth_accounts
		SET access_token = $1, refresh_token = $2, token_expires_at = $3, profile_data = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
		account.ProfileData,
		account.ID,
	).Scan(&account.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("oauth account not found")
		}
		return fmt.Errorf("failed to update oauth account: %w", err)
	}

	return nil
}

// DeleteOAuthAccount deletes an OAuth account
func (r *OAuthRepository) DeleteOAuthAccount(id uuid.UUID) error {
	query := `DELETE FROM oauth_accounts WHERE id = $1`

	result, err := r.db.Exec(query, id)
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
func (r *OAuthRepository) DeleteOAuthAccountsByProvider(userID uuid.UUID, provider string) error {
	query := `DELETE FROM oauth_accounts WHERE user_id = $1 AND provider = $2`

	_, err := r.db.Exec(query, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to delete oauth accounts by provider: %w", err)
	}

	return nil
}

// GetByUserID returns all OAuth accounts for a user
func (r *OAuthRepository) GetByUserID(userID uuid.UUID) ([]*models.OAuthAccount, error) {
	var accounts []*models.OAuthAccount
	query := `SELECT * FROM oauth_accounts WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.Select(&accounts, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth accounts by user ID: %w", err)
	}

	return accounts, nil
}

// ListAll returns all OAuth accounts (admin only)
func (r *OAuthRepository) ListAll() ([]*models.OAuthAccount, error) {
	var accounts []*models.OAuthAccount
	query := `SELECT * FROM oauth_accounts ORDER BY created_at DESC`

	err := r.db.Select(&accounts, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all oauth accounts: %w", err)
	}

	return accounts, nil
}
