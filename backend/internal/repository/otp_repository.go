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

// OTPRepository handles OTP-related database operations
type OTPRepository struct {
	db *Database
}

// NewOTPRepository creates a new OTP repository
func NewOTPRepository(db *Database) *OTPRepository {
	return &OTPRepository{db: db}
}

// Create creates a new OTP
func (r *OTPRepository) Create(ctx context.Context, otp *models.OTP) error {
	_, err := r.db.NewInsert().
		Model(otp).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create OTP: %w", err)
	}

	return nil
}

// GetByEmailAndType retrieves the latest valid OTP for email and type
func (r *OTPRepository) GetByEmailAndType(ctx context.Context, email string, otpType models.OTPType) (*models.OTP, error) {
	otp := new(models.OTP)

	err := r.db.NewSelect().
		Model(otp).
		Where("email = ?", email).
		Where("type = ?", otpType).
		Where("used = ?", false).
		Where("expires_at > ?", bun.Safe("CURRENT_TIMESTAMP")).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.NewAppError(404, "OTP not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	return otp, nil
}

// GetByPhoneAndType retrieves the latest valid OTP for phone and type
func (r *OTPRepository) GetByPhoneAndType(ctx context.Context, phone string, otpType models.OTPType) (*models.OTP, error) {
	otp := new(models.OTP)

	err := r.db.NewSelect().
		Model(otp).
		Where("phone = ?", phone).
		Where("type = ?", otpType).
		Where("used = ?", false).
		Where("expires_at > ?", bun.Safe("CURRENT_TIMESTAMP")).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.NewAppError(404, "OTP not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	return otp, nil
}

// MarkAsUsed marks an OTP as used
func (r *OTPRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewUpdate().
		Model((*models.OTP)(nil)).
		Set("used = ?", true).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return models.NewAppError(404, "OTP not found")
	}

	return nil
}

// InvalidateAllForEmail invalidates all OTPs for an email and type
func (r *OTPRepository) InvalidateAllForEmail(ctx context.Context, email string, otpType models.OTPType) error {
	_, err := r.db.NewUpdate().
		Model((*models.OTP)(nil)).
		Set("used = ?", true).
		Where("email = ?", email).
		Where("type = ?", otpType).
		Where("used = ?", false).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to invalidate OTPs: %w", err)
	}

	return nil
}

// InvalidateAllForPhone invalidates all OTPs for a phone number and type
func (r *OTPRepository) InvalidateAllForPhone(ctx context.Context, phone string, otpType models.OTPType) error {
	_, err := r.db.NewUpdate().
		Model((*models.OTP)(nil)).
		Set("used = ?", true).
		Where("phone = ?", phone).
		Where("type = ?", otpType).
		Where("used = ?", false).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to invalidate OTPs: %w", err)
	}

	return nil
}

// DeleteExpired deletes expired OTPs older than the specified duration
func (r *OTPRepository) DeleteExpired(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	_, err := r.db.NewDelete().
		Model((*models.OTP)(nil)).
		Where("expires_at < ?", cutoff).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete expired OTPs: %w", err)
	}

	return nil
}

// CountRecentByEmail counts OTPs created for an email in the last duration
func (r *OTPRepository) CountRecentByEmail(ctx context.Context, email string, otpType models.OTPType, duration time.Duration) (int, error) {
	cutoff := time.Now().Add(-duration)

	count, err := r.db.NewSelect().
		Model((*models.OTP)(nil)).
		Where("email = ?", email).
		Where("type = ?", otpType).
		Where("created_at > ?", cutoff).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count recent OTPs: %w", err)
	}

	return count, nil
}

// CountRecentByPhone counts OTPs created for a phone number in the last duration
func (r *OTPRepository) CountRecentByPhone(ctx context.Context, phone string, otpType models.OTPType, duration time.Duration) (int, error) {
	cutoff := time.Now().Add(-duration)

	count, err := r.db.NewSelect().
		Model((*models.OTP)(nil)).
		Where("phone = ?", phone).
		Where("type = ?", otpType).
		Where("created_at > ?", cutoff).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count recent OTPs: %w", err)
	}

	return count, nil
}
