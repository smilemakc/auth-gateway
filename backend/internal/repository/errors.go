package repository

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	"github.com/smilemakc/auth-gateway/internal/models"
	"github.com/uptrace/bun/driver/pgdriver"
)

// handlePgError converts PostgreSQL errors to application-specific errors
// In production, hides internal database details
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

		case "40001": // serialization_failure
			return &RetryableError{Err: err, Code: code}

		case "40P01": // deadlock_detected
			return &RetryableError{Err: err, Code: code}
		}
	}

	// In production, hide database error details
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("SERVER_ENV")
	}
	if env == "production" || env == "prod" {
		// Hide database connection errors, SQL syntax errors, etc.
		errStr := err.Error()
		if strings.Contains(errStr, "database") ||
			strings.Contains(errStr, "connection") ||
			strings.Contains(errStr, "SQL") ||
			strings.Contains(errStr, "pq:") ||
			strings.Contains(errStr, "driver") {
			return models.ErrInternalServer
		}
	}

	// Check for retryable errors in error message (fallback)
	errStr := err.Error()
	if strings.Contains(strings.ToLower(errStr), "deadlock") ||
		strings.Contains(strings.ToLower(errStr), "serialization") {
		return &RetryableError{Err: err, Code: ""}
	}

	// Return original error if not a known PostgreSQL error (or in development)
	return err
}

// RetryableError represents an error that can be retried (deadlock, serialization failure)
type RetryableError struct {
	Err  error
	Code string // PostgreSQL error code
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	_, ok := err.(*RetryableError)
	return ok
}
