package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// TokenRepository handles token-related database operations
type TokenRepository struct {
	db *Database
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *Database) *TokenRepository {
	return &TokenRepository{db: db}
}

// CreateRefreshToken creates a new refresh token
func (r *TokenRepository) CreateRefreshToken(token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	err := r.db.QueryRow(
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves a refresh token by token hash
func (r *TokenRepository) GetRefreshToken(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	query := `SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked_at IS NULL`

	err := r.db.Get(&token, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

// RevokeRefreshToken revokes a refresh token
func (r *TokenRepository) RevokeRefreshToken(tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE token_hash = $1`

	result, err := r.db.Exec(query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.ErrInvalidToken
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *TokenRepository) RevokeAllUserTokens(userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND revoked_at IS NULL`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	return nil
}

// DeleteExpiredRefreshTokens deletes expired refresh tokens
func (r *TokenRepository) DeleteExpiredRefreshTokens() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}

// AddToBlacklist adds a token to the blacklist
func (r *TokenRepository) AddToBlacklist(token *models.TokenBlacklist) error {
	query := `
		INSERT INTO token_blacklist (id, token_hash, user_id, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (token_hash) DO NOTHING
		RETURNING created_at
	`

	err := r.db.QueryRow(
		query,
		token.ID,
		token.TokenHash,
		token.UserID,
		token.ExpiresAt,
	).Scan(&token.CreatedAt)

	if err != nil {
		// If error is due to ON CONFLICT, it's okay
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (r *TokenRepository) IsBlacklisted(tokenHash string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP)`

	err := r.db.QueryRow(query, tokenHash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists, nil
}

// DeleteExpiredBlacklistedTokens deletes expired tokens from the blacklist
func (r *TokenRepository) DeleteExpiredBlacklistedTokens() error {
	query := `DELETE FROM token_blacklist WHERE expires_at < CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete expired blacklisted tokens: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes all expired tokens (both refresh and blacklist)
func (r *TokenRepository) CleanupExpiredTokens() error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete expired refresh tokens
	if _, err := tx.Exec(`DELETE FROM refresh_tokens WHERE expires_at < $1`, time.Now()); err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	// Delete expired blacklisted tokens
	if _, err := tx.Exec(`DELETE FROM token_blacklist WHERE expires_at < $1`, time.Now()); err != nil {
		return fmt.Errorf("failed to delete expired blacklisted tokens: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
