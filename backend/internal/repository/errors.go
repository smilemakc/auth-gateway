package repository

import (
	"database/sql"
	"errors"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun/driver/pgdriver"
)

// handlePgError converts PostgreSQL errors to application-specific errors
func handlePgError(err error) error {
	if err == nil {
		return nil
	}

	// Check for no rows error
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrNotFound
	}

	// Check for PostgreSQL-specific errors
	if pgErr, ok := err.(pgdriver.Error); ok {
		code := pgErr.Field('C')

		switch code {
		case "23505": // unique_violation
			constraint := pgErr.Field('n')

			// User-specific constraints
			if constraint == "users_email_key" {
				return models.ErrEmailAlreadyExists
			}
			if constraint == "users_username_key" {
				return models.ErrUsernameAlreadyExists
			}
			if constraint == "idx_users_phone_unique" {
				return models.ErrPhoneAlreadyExists
			}

			// Generic unique violation
			return models.ErrAlreadyExists

		case "23503": // foreign_key_violation
			return models.ErrForeignKeyViolation

		case "23502": // not_null_violation
			return models.ErrRequiredField
		}
	}

	// Return original error if not a known PostgreSQL error
	return err
}
