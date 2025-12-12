package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smilemakc/auth-gateway/internal/models"
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
func (r *OTPRepository) Create(otp *models.OTP) error {
	query := `
		INSERT INTO otps (id, email, phone, code, type, used, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`

	err := r.db.QueryRow(
		query,
		otp.ID,
		otp.Email,
		otp.Phone,
		otp.Code,
		otp.Type,
		otp.Used,
		otp.ExpiresAt,
	).Scan(&otp.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create OTP: %w", err)
	}

	return nil
}

// GetByEmailAndType retrieves the latest valid OTP for email and type
func (r *OTPRepository) GetByEmailAndType(email string, otpType models.OTPType) (*models.OTP, error) {
	query := `
		SELECT id, email, phone, code, type, used, expires_at, created_at
		FROM otps
		WHERE email = $1 AND type = $2 AND used = FALSE AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC
		LIMIT 1
	`

	otp := &models.OTP{}
	err := r.db.QueryRow(query, email, otpType).Scan(
		&otp.ID,
		&otp.Email,
		&otp.Phone,
		&otp.Code,
		&otp.Type,
		&otp.Used,
		&otp.ExpiresAt,
		&otp.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.NewAppError(404, "OTP not found or expired")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	return otp, nil
}

// GetByPhoneAndType retrieves the latest valid OTP for phone and type
func (r *OTPRepository) GetByPhoneAndType(phone string, otpType models.OTPType) (*models.OTP, error) {
	query := `
		SELECT id, email, phone, code, type, used, expires_at, created_at
		FROM otps
		WHERE phone = $1 AND type = $2 AND used = FALSE AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC
		LIMIT 1
	`

	otp := &models.OTP{}
	err := r.db.QueryRow(query, phone, otpType).Scan(
		&otp.ID,
		&otp.Email,
		&otp.Phone,
		&otp.Code,
		&otp.Type,
		&otp.Used,
		&otp.ExpiresAt,
		&otp.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, models.NewAppError(404, "OTP not found or expired")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	return otp, nil
}

// MarkAsUsed marks an OTP as used
func (r *OTPRepository) MarkAsUsed(id uuid.UUID) error {
	query := `UPDATE otps SET used = TRUE WHERE id = $1`

	result, err := r.db.Exec(query, id)
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
func (r *OTPRepository) InvalidateAllForEmail(email string, otpType models.OTPType) error {
	query := `UPDATE otps SET used = TRUE WHERE email = $1 AND type = $2 AND used = FALSE`

	_, err := r.db.Exec(query, email, otpType)
	if err != nil {
		return fmt.Errorf("failed to invalidate OTPs: %w", err)
	}

	return nil
}

// InvalidateAllForPhone invalidates all OTPs for a phone number and type
func (r *OTPRepository) InvalidateAllForPhone(phone string, otpType models.OTPType) error {
	query := `UPDATE otps SET used = TRUE WHERE phone = $1 AND type = $2 AND used = FALSE`

	_, err := r.db.Exec(query, phone, otpType)
	if err != nil {
		return fmt.Errorf("failed to invalidate OTPs: %w", err)
	}

	return nil
}

// DeleteExpired deletes expired OTPs older than the specified duration
func (r *OTPRepository) DeleteExpired(olderThan time.Duration) error {
	query := `DELETE FROM otps WHERE expires_at < $1`

	cutoff := time.Now().Add(-olderThan)
	_, err := r.db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete expired OTPs: %w", err)
	}

	return nil
}

// CountRecentByEmail counts OTPs created for an email in the last duration
func (r *OTPRepository) CountRecentByEmail(email string, otpType models.OTPType, duration time.Duration) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM otps
		WHERE email = $1 AND type = $2 AND created_at > $3
	`

	cutoff := time.Now().Add(-duration)
	var count int

	err := r.db.QueryRow(query, email, otpType, cutoff).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count recent OTPs: %w", err)
	}

	return count, nil
}

// CountRecentByPhone counts OTPs created for a phone number in the last duration
func (r *OTPRepository) CountRecentByPhone(phone string, otpType models.OTPType, duration time.Duration) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM otps
		WHERE phone = $1 AND type = $2 AND created_at > $3
	`

	cutoff := time.Now().Add(-duration)
	var count int

	err := r.db.QueryRow(query, phone, otpType, cutoff).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count recent OTPs: %w", err)
	}

	return count, nil
}
