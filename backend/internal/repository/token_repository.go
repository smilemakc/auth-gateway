package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun"
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
func (r *TokenRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	_, err := r.db.NewInsert().
		Model(token).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves a refresh token by token hash
func (r *TokenRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	token := new(models.RefreshToken)

	err := r.db.NewSelect().
		Model(token).
		Where("token_hash = ?", tokenHash).
		Where("revoked_at IS NULL").
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return token, nil
}

// RevokeRefreshToken revokes a refresh token
func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	result, err := r.db.NewUpdate().
		Model((*models.RefreshToken)(nil)).
		Set("revoked_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("token_hash = ?", tokenHash).
		Exec(ctx)

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
func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*models.RefreshToken)(nil)).
		Set("revoked_at = ?", bun.Safe("CURRENT_TIMESTAMP")).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	return nil
}

// DeleteExpiredRefreshTokens deletes expired refresh tokens
func (r *TokenRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	_, err := r.db.NewDelete().
		Model((*models.RefreshToken)(nil)).
		Where("expires_at < ?", bun.Safe("CURRENT_TIMESTAMP")).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}

// AddToBlacklist adds a token to the blacklist
func (r *TokenRepository) AddToBlacklist(ctx context.Context, token *models.TokenBlacklist) error {
	// Use INSERT ... ON CONFLICT DO UPDATE to ensure the row is inserted or updated
	// This avoids issues with DO NOTHING not returning anything
	_, err := r.db.NewInsert().
		Model(token).
		On("CONFLICT (token_hash) DO UPDATE").
		Set("expires_at = EXCLUDED.expires_at").
		Set("user_id = EXCLUDED.user_id").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (r *TokenRepository) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*models.TokenBlacklist)(nil)).
		Where("token_hash = ?", tokenHash).
		Where("expires_at > ?", bun.Safe("CURRENT_TIMESTAMP")).
		Exists(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists, nil
}

// DeleteExpiredBlacklistedTokens deletes expired tokens from the blacklist
func (r *TokenRepository) DeleteExpiredBlacklistedTokens(ctx context.Context) error {
	_, err := r.db.NewDelete().
		Model((*models.TokenBlacklist)(nil)).
		Where("expires_at < ?", bun.Safe("CURRENT_TIMESTAMP")).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete expired blacklisted tokens: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes all expired tokens (both refresh and blacklist)
func (r *TokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete expired refresh tokens
		_, err := tx.NewDelete().
			Model((*models.RefreshToken)(nil)).
			Where("expires_at < ?", time.Now()).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
		}

		// Delete expired blacklisted tokens
		_, err = tx.NewDelete().
			Model((*models.TokenBlacklist)(nil)).
			Where("expires_at < ?", time.Now()).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete expired blacklisted tokens: %w", err)
		}

		return nil
	})
}
