package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
)

// BackupCodeRepository handles backup code database operations
type BackupCodeRepository struct {
	db *Database
}

// NewBackupCodeRepository creates a new backup code repository
func NewBackupCodeRepository(db *Database) *BackupCodeRepository {
	return &BackupCodeRepository{db: db}
}

// Create creates multiple backup codes
func (r *BackupCodeRepository) CreateBatch(codes []*models.BackupCode) error {
	query := `
		INSERT INTO backup_codes (id, user_id, code_hash, used, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, code := range codes {
		_, err := tx.Exec(query, code.ID, code.UserID, code.CodeHash, code.Used, code.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to create backup code: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUnusedByUserID retrieves all unused backup codes for a user
func (r *BackupCodeRepository) GetUnusedByUserID(userID uuid.UUID) ([]*models.BackupCode, error) {
	var codes []*models.BackupCode
	query := `
		SELECT id, user_id, code_hash, used, used_at, created_at
		FROM backup_codes
		WHERE user_id = $1 AND used = FALSE
		ORDER BY created_at DESC
	`

	err := r.db.Select(&codes, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup codes: %w", err)
	}

	return codes, nil
}

// CountUnusedByUserID counts unused backup codes for a user
func (r *BackupCodeRepository) CountUnusedByUserID(userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM backup_codes WHERE user_id = $1 AND used = FALSE`

	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count backup codes: %w", err)
	}

	return count, nil
}

// MarkAsUsed marks a backup code as used
func (r *BackupCodeRepository) MarkAsUsed(id uuid.UUID) error {
	query := `
		UPDATE backup_codes
		SET used = TRUE, used_at = $1
		WHERE id = $2 AND used = FALSE
	`

	result, err := r.db.Exec(query, time.Now(), id)
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
func (r *BackupCodeRepository) DeleteAllByUserID(userID uuid.UUID) error {
	query := `DELETE FROM backup_codes WHERE user_id = $1`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete backup codes: %w", err)
	}

	return nil
}

// GetByCodeHash retrieves a backup code by its hash
func (r *BackupCodeRepository) GetByCodeHash(userID uuid.UUID, codeHash string) (*models.BackupCode, error) {
	var code models.BackupCode
	query := `
		SELECT id, user_id, code_hash, used, used_at, created_at
		FROM backup_codes
		WHERE user_id = $1 AND code_hash = $2
	`

	err := r.db.Get(&code, query, userID, codeHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get backup code: %w", err)
	}

	return &code, nil
}
