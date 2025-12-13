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

// BackupCodeRepository handles backup code database operations
type BackupCodeRepository struct {
	db *Database
}

// NewBackupCodeRepository creates a new backup code repository
func NewBackupCodeRepository(db *Database) *BackupCodeRepository {
	return &BackupCodeRepository{db: db}
}

// CreateBatch creates multiple backup codes
func (r *BackupCodeRepository) CreateBatch(ctx context.Context, codes []*models.BackupCode) error {
	if len(codes) == 0 {
		return nil
	}

	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().
			Model(&codes).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to create backup codes: %w", err)
		}

		return nil
	})
}

// GetUnusedByUserID retrieves all unused backup codes for a user
func (r *BackupCodeRepository) GetUnusedByUserID(ctx context.Context, userID uuid.UUID) ([]*models.BackupCode, error) {
	codes := make([]*models.BackupCode, 0)

	err := r.db.NewSelect().
		Model(&codes).
		Where("user_id = ?", userID).
		Where("used = ?", false).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get backup codes: %w", err)
	}

	return codes, nil
}

// CountUnusedByUserID counts unused backup codes for a user
func (r *BackupCodeRepository) CountUnusedByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.BackupCode)(nil)).
		Where("user_id = ?", userID).
		Where("used = ?", false).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count backup codes: %w", err)
	}

	return count, nil
}

// MarkAsUsed marks a backup code as used
func (r *BackupCodeRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.BackupCode)(nil)).
		Set("used = ?", true).
		Set("used_at = ?", time.Now()).
		Where("id = ?", id).
		Where("used = ?", false).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to mark backup code as used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return errors.New("backup code not found or already used")
	}

	return nil
}

// DeleteAllByUserID deletes all backup codes for a user
func (r *BackupCodeRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.NewDelete().
		Model((*models.BackupCode)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete backup codes: %w", err)
	}

	return nil
}

// GetByCodeHash retrieves a backup code by its hash
func (r *BackupCodeRepository) GetByCodeHash(ctx context.Context, userID uuid.UUID, codeHash string) (*models.BackupCode, error) {
	code := new(models.BackupCode)

	err := r.db.NewSelect().
		Model(code).
		Where("user_id = ?", userID).
		Where("code_hash = ?", codeHash).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get backup code: %w", err)
	}

	return code, nil
}
