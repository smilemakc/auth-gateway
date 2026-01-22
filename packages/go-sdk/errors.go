// Package authgateway provides a Go SDK for the Auth Gateway API.
package authgateway

import (
	"fmt"
)

// Error codes
const (
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeBadRequest         = "BAD_REQUEST"
	ErrCodeConflict           = "CONFLICT"
	ErrCodeTooManyRequests    = "TOO_MANY_REQUESTS"
	ErrCodeInternalServer     = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeTwoFactorRequired  = "TWO_FACTOR_REQUIRED"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeEmailNotVerified   = "EMAIL_NOT_VERIFIED"
	ErrCodeAccountDisabled    = "ACCOUNT_DISABLED"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
)

// APIError represents an error returned by the Auth Gateway API.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s (status: %d)", e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("%s (status: %d)", e.Code, e.StatusCode)
}

// IsCode checks if the error matches a specific error code.
func (e *APIError) IsCode(code string) bool {
	return e.Code == code
}

// IsUnauthorized returns true if the error is an authentication error.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// IsForbidden returns true if the error is an authorization error.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == 403
}

// IsNotFound returns true if the resource was not found.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsTooManyRequests returns true if rate limited.
func (e *APIError) IsTooManyRequests() bool {
	return e.StatusCode == 429
}

// TwoFactorRequiredError is returned when 2FA verification is needed.
type TwoFactorRequiredError struct {
	TwoFactorToken string
}

func (e *TwoFactorRequiredError) Error() string {
	return "two-factor authentication required"
}

// AuthenticationError is returned for authentication failures.
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "authentication failed"
}

// ValidationError is returned for validation failures.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NetworkError is returned for network-related errors.
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}
