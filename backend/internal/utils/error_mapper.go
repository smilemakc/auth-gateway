package utils

import (
	"fmt"
	"strings"

	"github.com/smilemakc/auth-gateway/internal/models"
)

// ErrorMapper provides centralized error handling and mapping
// Hides internal error details in production environment
type ErrorMapper struct {
	isProduction bool
}

// NewErrorMapper creates a new error mapper
func NewErrorMapper(env string) *ErrorMapper {
	return &ErrorMapper{
		isProduction: env == "production" || env == "prod",
	}
}

// MapError maps internal errors to safe user-facing errors
// In production, hides internal details; in development, shows more information
func (m *ErrorMapper) MapError(err error) error {
	if err == nil {
		return nil
	}

	// If it's already an AppError, return as-is (it's already safe)
	if appErr, ok := err.(*models.AppError); ok {
		return appErr
	}

	// Check for known model errors
	switch err {
	case models.ErrNotFound:
		return models.NewAppError(404, "Resource not found")
	case models.ErrEmailAlreadyExists:
		return models.NewAppError(409, "Email already exists")
	case models.ErrUsernameAlreadyExists:
		return models.NewAppError(409, "Username already exists")
	case models.ErrPhoneAlreadyExists:
		return models.NewAppError(409, "Phone number already exists")
	case models.ErrInvalidCredentials:
		return models.NewAppError(401, "Invalid credentials")
	case models.ErrInvalidToken:
		return models.NewAppError(401, "Invalid or expired token")
	case models.ErrTokenRevoked:
		return models.NewAppError(401, "Token has been revoked")
	case models.ErrTokenExpired:
		return models.NewAppError(401, "Token has expired")
	case models.ErrRateLimitExceeded:
		return models.NewAppError(429, "Rate limit exceeded")
	case models.ErrAlreadyExists:
		return models.NewAppError(409, "Resource already exists")
	case models.ErrForeignKeyViolation:
		if m.isProduction {
			return models.NewAppError(400, "Invalid reference")
		}
		return models.NewAppError(400, fmt.Sprintf("Foreign key violation: %v", err))
	case models.ErrRequiredField:
		if m.isProduction {
			return models.NewAppError(400, "Required field is missing")
		}
		return models.NewAppError(400, fmt.Sprintf("Required field error: %v", err))
	}

	// For unknown errors, hide details in production
	errStr := err.Error()
	if m.isProduction {
		// Hide database connection errors, SQL errors, etc.
		if strings.Contains(errStr, "database") ||
			strings.Contains(errStr, "connection") ||
			strings.Contains(errStr, "SQL") ||
			strings.Contains(errStr, "pq:") ||
			strings.Contains(errStr, "driver") {
			return models.NewAppError(500, "Internal server error")
		}

		// Hide file system errors
		if strings.Contains(errStr, "no such file") ||
			strings.Contains(errStr, "permission denied") {
			return models.NewAppError(500, "Internal server error")
		}

		// For other errors, return generic message
		return models.NewAppError(500, "An error occurred")
	}

	// In development, return more details
	return models.NewAppError(500, fmt.Sprintf("Internal error: %v", err))
}

// LogError logs an error with appropriate detail level
// In production, logs full details; in responses, returns safe messages
func (m *ErrorMapper) LogError(err error) string {
	if err == nil {
		return ""
	}

	// In production, still log full details for debugging
	// but return safe message to user
	if m.isProduction {
		return err.Error() // Full error for logging
	}

	return err.Error() // Full error in development
}

// ShouldLogError determines if an error should be logged
// Some errors (like validation errors) don't need logging
func (m *ErrorMapper) ShouldLogError(err error) bool {
	if err == nil {
		return false
	}

	// Don't log client errors (4xx)
	if appErr, ok := err.(*models.AppError); ok {
		return appErr.Code >= 500
	}

	// Log all non-AppError errors
	return true
}
